package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/handlers"
	"github.com/AndB0ndar/doc-archive/internal/repository"
	"github.com/AndB0ndar/doc-archive/internal/vectorizer"
)

func NewRouter(
	cfg *config.Config,
	docRepo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	embedderClient *vectorizer.Client,
) http.Handler {
	r := chi.NewRouter()

	// ---- Global middleware ----
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(60))

	// ---- Public routes ----
	r.Get("/", handlers.Home)
	r.Get("/health", handlers.Health)

	uploadHandler := handlers.NewUploadHandler(cfg, docRepo, chunkRepo, embedderClient)
	r.Post("/upload", uploadHandler.ServeHTTP)

	searchHandler := handlers.NewSearchHandler(cfg, docRepo, chunkRepo, embedderClient)
	r.Get("/search", searchHandler.ServeHTTP)

	return r
}
