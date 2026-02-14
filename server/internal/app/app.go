package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/auth"
	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/db"
	"github.com/AndB0ndar/doc-archive/internal/logger"
	"github.com/AndB0ndar/doc-archive/internal/repository"
	"github.com/AndB0ndar/doc-archive/internal/server"
	"github.com/AndB0ndar/doc-archive/internal/service"
)

type App struct {
	config *config.Config
}

func New(cfg *config.Config) *App {
	return &App{config: cfg}
}

func (a *App) Run() error {
	logger.Setup(a.config.Env)
	slog.Info("config loaded", "port", a.config.Port, "env", a.config.Env)

	auth.SetJWTSecret(a.config.JWTSecret)
	if a.config.JWTSecret == "default-secret-change-me" && a.config.Env == "production" {
		slog.Warn("JWT_SECRET is set to default value, please change it in production")
	}

	// DB
	pool, err := db.NewPool(a.config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()
	if err := db.RunMigrations(pool, a.config.Database); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Repositories
	docRepo := repository.NewDocumentRepository(pool)
	chunkRepo := repository.NewChunkRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	// Service
	embedderService := service.NewEmbedder(a.config)
	docService := service.NewDocumentService(
		a.config, docRepo, chunkRepo, embedderService,
	)
	searchService := service.NewSearchService(
		a.config, chunkRepo, embedderService,
	)

	handler := server.NewRouter(userRepo, docRepo, docService, searchService)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("received signal", "signal", sig)
	case err := <-serverErr:
		slog.Error("server error", "error", err)
		return err
	}

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	os.Exit(0)
	return nil
}
