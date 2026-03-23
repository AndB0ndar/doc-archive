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
	"github.com/AndB0ndar/doc-archive/internal/db"
	"github.com/AndB0ndar/doc-archive/internal/handlers"
	"github.com/AndB0ndar/doc-archive/internal/logger"
	"github.com/AndB0ndar/doc-archive/internal/repository"
	"github.com/AndB0ndar/doc-archive/internal/server"

	"github.com/AndB0ndar/doc-archive/internal/client/embedder"
	"github.com/AndB0ndar/doc-archive/internal/client/reader"
	"github.com/AndB0ndar/doc-archive/internal/client/reranker"

	"github.com/AndB0ndar/doc-archive/internal/service/auth"
	"github.com/AndB0ndar/doc-archive/internal/service/chunker"
	"github.com/AndB0ndar/doc-archive/internal/service/document"
	"github.com/AndB0ndar/doc-archive/internal/service/search"
)

type App struct {
	config *config.Config
}

func New(cfg *config.Config) *App {
	return &App{config: cfg}
}

func (a *App) Run() error {
	log := logger.New(a.config.Env)
	log.Info("config loaded", "port", a.config.Port, "env", a.config.Env)

	if a.config.JWTSecret == "default-secret-change-me" && a.config.Env == "production" {
		log.Warn("JWT_SECRET is set to default value, please change it in production")
	}

	// DB
	pool, err := db.NewPool(a.config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Repositories
	docRepo := repository.NewDocumentRepository(pool)
	chunkRepo := repository.NewChunkRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	// Clients
	embedderClient := embedder.New(a.config.EmbedderURL)
	rerankerClient := reranker.New(a.config.RerankerURL)
	readerClient := reader.New(a.config.ReaderURL)

	// Service
	authService := auth.New(
		userRepo,
		a.config.JWTSecret, time.Duration(a.config.JWTExpiry)*time.Hour,
	)
	chunkerSvc := chunker.New(
		a.config.Chunk.Size,
		a.config.Chunk.Overlap,
		a.config.Chunk.Separators,
		a.config.Chunk.SplitBySentences,
	)
	docService := document.New(
		docRepo, chunkRepo, embedderClient, chunkerSvc,
		a.config.UploadDir, log,
	)
	searchService := search.New(
		chunkRepo, embedderClient, rerankerClient, readerClient,
		a.config.Search,
	)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	uploadHandler := handlers.NewUploadHandler(docService, log)
	searchHandler := handlers.NewSearchHandler(searchService, log)
	docHandler := handlers.NewDocumentHandler(docService, log)

	// Router
	handler := server.NewRouter(
		authHandler, uploadHandler, searchHandler, docHandler,
		a.config.JWTSecret,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	log.Info("server stopped gracefully")
	return nil
}
