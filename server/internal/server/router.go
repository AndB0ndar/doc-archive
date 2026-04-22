package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

    "github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/AndB0ndar/doc-archive/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/AndB0ndar/doc-archive/internal/handlers"
	mdwr "github.com/AndB0ndar/doc-archive/internal/middleware"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	uploadHandler *handlers.UploadHandler,
	searchHandler *handlers.SearchHandler,
	docHandler *handlers.DocumentHandler,
	JWTSecret string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Logger)

	//r.Use(mdwr.Logger(slog.Default()))

	r.Get("/health", handlers.Health)

	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(mdwr.AuthMiddleware(JWTSecret))
		r.Use(mdwr.MetricsMiddleware())

		r.Post("/upload", uploadHandler.ServeHTTP)

		r.Get("/search", searchHandler.Search)

		r.Route("/documents", func(r chi.Router) {
			r.Get("/", docHandler.ListDocuments)
			r.Get("/{id}", docHandler.GetDocument)
			r.Delete("/{id}", docHandler.DeleteDocument)
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Handle("/metrics", promhttp.Handler())

	return r
}
