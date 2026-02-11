package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AndB0ndar/doc-archive/internal/handlers"
)

func New() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Logger)
	//  TODO
	// import "log/slog"
	// import "github.com/AndB0ndar/doc-archive/internal/middleware"
	// r.Use(customMW.Logger(slog.Default()))

	// Public routes
	r.Get("/", handlers.Home)
	r.Get("/health", handlers.Health)

	// TODO: add
	// r.Post("/upload", handlers.Upload)
	// r.Get("/search", handlers.Search)
	// r.Get("/documents/{id}", handlers.Document)

	return r
}
