package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AndB0ndar/doc-archive/internal/handlers"
	"github.com/AndB0ndar/doc-archive/internal/service"
)

func NewRouter(
	docService *service.DocumentService,
	searchService *service.SearchService,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", handlers.Health)

	uploadHandler := handlers.NewUploadHandler(docService)
	r.Post("/upload", uploadHandler.ServeHTTP)

	searchAPIHandler := handlers.NewSearchHandler(searchService)
	r.Get("/search", searchAPIHandler.ServeHTTP)

	docHandler := handlers.NewDocumentHandler(docService.docRepo)  // FIXME
	r.Get("/documents/{id}", docHandler.GetDocument)

	return r
}
