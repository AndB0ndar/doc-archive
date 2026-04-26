package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/domain"

	api "github.com/AndB0ndar/doc-archive/internal/transport/http"
	api_handlers "github.com/AndB0ndar/doc-archive/internal/transport/http/handlers"
	indexer "github.com/AndB0ndar/doc-archive/internal/transport/queue"
	indexer_handlers "github.com/AndB0ndar/doc-archive/internal/transport/queue/handlers"

	"github.com/AndB0ndar/doc-archive/internal/repository"

	"github.com/AndB0ndar/doc-archive/internal/infrastructure/cache"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/minio"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/postgres"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/queue"

	"github.com/AndB0ndar/doc-archive/internal/infrastructure/clients/embedder"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/clients/reader"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/clients/reranker"

	"github.com/AndB0ndar/doc-archive/internal/service/auth"
	"github.com/AndB0ndar/doc-archive/internal/service/chunker"
	"github.com/AndB0ndar/doc-archive/internal/service/document"
	"github.com/AndB0ndar/doc-archive/internal/service/search"
)

type CommonDependencies struct {
	Log            *logger.Logger
	DB             *postgres.Pool // *pgxpool.Pool
	MinIO          *minio.MinIOStorage
	DocRepo        domain.DocumentRepository
	ChunkRepo      domain.ChunkRepository
	UserRepo       domain.UserRepository
	EmbedderClient domain.EmbedderClient
	RerankerClient domain.RerankerClient
	ReaderClient   domain.ReaderClient
	Chunker        domain.ChunkerService
}

type App struct {
	config *config.Config
}

func New(cfg *config.Config) *App {
	return &App{config: cfg}
}

func (a *App) initCommonDependencies() (*CommonDependencies, error) {
	log := logger.New(a.config.Env)
	log.Info("initializing common dependencies")

	// DB PostgreSQL
	pool, err := postgres.NewPool(a.config.Database, log)
	if err != nil {
		return nil, err
	}

	// MinIO
	minioStorage, err := minio.NewMinIOStorage(
		a.config.MinIO.URL,
		a.config.MinIO.AccessKey,
		a.config.MinIO.SecretKey,
		a.config.MinIO.Bucket,
		a.config.MinIO.UseSSL,
	)
	if err != nil {
		pool.Close()
		return nil, err
	}

	// Repositories
	oDocRepo := repository.NewDocumentRepository(pool)
	chunkRepo := repository.NewChunkRepository(pool)
	oUserRepo := repository.NewUserRepository(pool)

	// Caches (Redis)
	var docRepo domain.DocumentRepository
	var userRepo domain.UserRepository
	redisCache, err := cache.NewRedisCache(a.config.RedisURL)
	if err != nil {
		log.Warn("Redis unavailable, running without cache", "error", err)
		docRepo = oDocRepo
		userRepo = oUserRepo
	} else {
		log.Info("connected to Redis")
		docRepo = repository.NewCachedDocumentRepository(
			oDocRepo, redisCache, 5*time.Minute, log,
		)
		userRepo = repository.NewCachedUserRepository(
			oUserRepo, redisCache, 1*time.Hour,
		)
	}

	// Clients
	embedderClient := embedder.New(a.config.EmbedderURL)
	rerankerClient := reranker.New(a.config.RerankerURL)
	readerClient := reader.New(a.config.ReaderURL)

	// Chunker
	chunkerSvc := chunker.New(
		a.config.Chunk.Size,
		a.config.Chunk.Overlap,
		a.config.Chunk.Separators,
		a.config.Chunk.SplitBySentences,
	)

	deps := &CommonDependencies{
		Log:            log,
		DB:             pool,
		MinIO:          minioStorage,
		DocRepo:        docRepo,
		ChunkRepo:      chunkRepo,
		UserRepo:       userRepo,
		EmbedderClient: embedderClient,
		RerankerClient: rerankerClient,
		ReaderClient:   readerClient,
		Chunker:        chunkerSvc,
	}
	return deps, nil
}

func (a *App) RunAPI() error {
	deps, err := a.initCommonDependencies()
	if err != nil {
		return err
	}
	defer deps.DB.Close()

	log := deps.Log

	// Asynq Client
	asynqClient, err := queue.NewClient(
		a.config.RedisURL, 0,
		a.config.Queue.MaxRetry,
		time.Duration(a.config.Queue.TimeoutSec)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("failed to create asynq server: %w", err)
	}

	// Services
	authService := auth.New(
		deps.UserRepo,
		a.config.JWTSecret, time.Duration(a.config.JWTExpiry)*time.Hour,
	)
	docService := document.New(
		deps.DocRepo, deps.ChunkRepo,
		deps.EmbedderClient, deps.Chunker,
		deps.MinIO,
		asynqClient,
		log,
	)
	searchService := search.New(
		deps.ChunkRepo,
		deps.EmbedderClient, deps.RerankerClient, deps.ReaderClient,
		a.config.Search,
		log,
	)

	// Handlers
	authHandler := api_handlers.NewAuthHandler(authService, log)
	uploadHandler := api_handlers.NewUploadHandler(docService, log)
	searchHandler := api_handlers.NewSearchHandler(searchService, log)
	docHandler := api_handlers.NewDocumentHandler(docService, log)

	// Router
	router := api.NewRouter(
		authHandler, uploadHandler, searchHandler, docHandler,
		a.config.JWTSecret,
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("received signal", "signal", sig)
	case err := <-serverErr:
		log.Error("server error", "error", err)
		return err
	}

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	log.Info("server stopped gracefully")
	return nil
}

func (a *App) RunIndexer() error {
	deps, err := a.initCommonDependencies()
	if err != nil {
		return err
	}
	defer deps.DB.Close()

	log := deps.Log
	log.Info(
		"starting indexer",
		"redis", a.config.RedisURL, "concurrency", a.config.Queue.Concurrency,
	)

	// Asynq Server
	srv := queue.NewServer(
		a.config.RedisURL, 0, a.config.Queue.Concurrency,
	)

	// Services
	docService := document.New(
		deps.DocRepo, deps.ChunkRepo,
		deps.EmbedderClient, deps.Chunker,
		deps.MinIO,
		nil,
		log,
	)

	// Handlers
	docHandler := indexer_handlers.NewDocumentHandler(docService, log)

	// Router
	mux := indexer.NewRouter(docHandler)

	// Launching a indexer (blocks it)
	if err := srv.Run(mux); err != nil {
		return fmt.Errorf("asynq server failed: %w", err)
	}
	return nil
}
