package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/handlers"
	"github.com/AndB0ndar/doc-archive/internal/repository"
)

func NewRouter(cfg *config.Config, docRepo *repository.DocumentRepository) http.Handler {
	r := chi.NewRouter()

	// ---- Global middleware ----
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.StripSlashes)
	//r.Use(middleware.Timeout(60))

	// ---- Public маршруты ----
	r.Get("/", handlers.Home)
	r.Get("/health", handlers.Health)

	uploadHandler := handlers.NewUploadHandler(cfg, docRepo)
	r.Post("/upload", uploadHandler.ServeHTTP)

	return r
}
