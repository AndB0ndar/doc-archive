package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AndB0ndar/doc-archive/pkg/jwt"
	"github.com/AndB0ndar/doc-archive/test/helpers"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	middleware := AuthMiddleware("test-secret")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware(next).ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	helpers.AssertEqual(t, nextCalled, false)
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	middleware := AuthMiddleware("test-secret")

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "Token mytoken"},
		{"wrong casing", "bearer mytoken"},
		{"only Bearer", "Bearer "},
		{"empty header value", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			middleware(next).ServeHTTP(w, r)

			helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
			helpers.AssertEqual(t, nextCalled, false)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	middleware := AuthMiddleware("test-secret")

	t.Run("garbage token string", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer this-is-not-a-valid-jwt-token")
		w := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		middleware(next).ServeHTTP(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
		helpers.AssertEqual(t, nextCalled, false)
	})

	t.Run("token signed with different secret", func(t *testing.T) {
		// Generate token with a different secret
		token, err := jwt.GenerateToken("test-user", "different-secret", time.Hour)
		helpers.AssertNoError(t, err)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		middleware(next).ServeHTTP(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
		helpers.AssertEqual(t, nextCalled, false)
	})

	t.Run("expired token", func(t *testing.T) {
		// Generate token that's already expired
		token, err := jwt.GenerateToken("test-user", "test-secret", -time.Hour)
		helpers.AssertNoError(t, err)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		middleware(next).ServeHTTP(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
		helpers.AssertEqual(t, nextCalled, false)
	})
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	middleware := AuthMiddleware(secret)

	t.Run("valid token passes and sets userID", func(t *testing.T) {
		token, err := jwt.GenerateToken("test-user", secret, time.Hour)
		helpers.AssertNoError(t, err)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		var capturedUserID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(string)
			if ok {
				capturedUserID = userID
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware(next).ServeHTTP(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
		helpers.AssertEqual(t, capturedUserID, "test-user")
	})

	t.Run("multiple valid tokens with different user IDs", func(t *testing.T) {
		for _, uid := range []string{"user1", "user2", "user-123"} {
			token, err := jwt.GenerateToken(uid, secret, time.Hour)
			helpers.AssertNoError(t, err)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			var capturedUserID string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := r.Context().Value(UserIDKey).(string)
				if ok {
					capturedUserID = userID
				}
				w.WriteHeader(http.StatusOK)
			})

			middleware(next).ServeHTTP(w, r)

			helpers.AssertEqual(t, w.Code, http.StatusOK)
			helpers.AssertEqual(t, capturedUserID, uid)
		}
	})
}
