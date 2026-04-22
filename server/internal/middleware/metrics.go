package middleware

import (
    "net/http"
    "strconv"
    "time"

	"github.com/AndB0ndar/doc-archive/internal/metrics"
)

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)
			duration := time.Since(start).Seconds()

			metrics.HttpRequestsTotal.WithLabelValues(
				r.Method, r.URL.Path, strconv.Itoa(rw.status),
			).Inc()
			metrics.HttpRequestDuration.WithLabelValues(
				r.Method, r.URL.Path,
			).Observe(duration)
		})
	}
}

