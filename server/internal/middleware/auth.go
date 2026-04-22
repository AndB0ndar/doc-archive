package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AndB0ndar/doc-archive/pkg/jwt"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid Authorization header format, expected 'Bearer <token>'", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			userID, err := jwt.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
