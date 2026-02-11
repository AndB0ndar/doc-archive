package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			duration := time.Since(start)
			requestID := middleware.GetReqID(r.Context())
			ip := r.RemoteAddr
			method := r.Method
			path := r.URL.Path
			query := r.URL.RawQuery
			userAgent := r.UserAgent()

			attrs := []slog.Attr{
				slog.Int("status", status),
				slog.String("method", method),
				slog.String("path", path),
				slog.String("query", query),
				slog.Duration("duration", duration),
				slog.String("ip", ip),
				slog.String("user_agent", userAgent),
			}
			if requestID != "" {
				attrs = append(attrs, slog.String("request_id", requestID))
			}

			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			logger.LogAttrs(r.Context(), level, "HTTP request", attrs...)
		})
	}
}
