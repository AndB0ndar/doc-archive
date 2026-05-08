package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/transport/http/middleware"
	"github.com/AndB0ndar/doc-archive/test/helpers"
)

// mockSearchService implements domain.SearchService for testing
type mockSearchService struct {
	searchFn func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error)
}

func (m *mockSearchService) Search(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, userID, limit)
	}
	return nil, errors.New("not implemented")
}

func setupTestSearchHandler(t *testing.T) (*SearchHandler, *mockSearchService) {
	t.Helper()
	mockSvc := &mockSearchService{}
	log := logger.New("test")
	handler := NewSearchHandler(mockSvc, log)
	return handler, mockSvc
}

func TestSearchHandler_Search(t *testing.T) {
	t.Run("valid search returns 200 with results", func(t *testing.T) {
		handler, mockSvc := setupTestSearchHandler(t)

		mockSvc.searchFn = func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
			helpers.AssertEqual(t, query.Query, "golang")
			helpers.AssertEqual(t, query.Type, "text")
			helpers.AssertEqual(t, userID, "test-user")
			helpers.AssertEqual(t, limit, 10)
			confidence := 0.95
			return []domain.SearchResult{
				{ChunkID: "chunk-1", DocumentID: "doc-1", Content: "Golang is great", Confidence: &confidence},
			}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)

		var resp models.SearchResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, len(resp.Results), 1)
		helpers.AssertEqual(t, resp.Results[0].ChunkID, "chunk-1")
		helpers.AssertEqual(t, resp.Results[0].DocumentID, "doc-1")
		helpers.AssertEqual(t, resp.Results[0].Content, "Golang is great")
	})

	t.Run("missing userID returns 401", func(t *testing.T) {
		handler, _ := setupTestSearchHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang", nil)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	})

	t.Run("empty query returns 400", func(t *testing.T) {
		handler, _ := setupTestSearchHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/search", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
	})

	t.Run("invalid search type returns 400", func(t *testing.T) {
		handler, _ := setupTestSearchHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang&type=hybrid", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		handler, mockSvc := setupTestSearchHandler(t)

		mockSvc.searchFn = func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
			return nil, errors.New("search service unavailable")
		}

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("default search type is text", func(t *testing.T) {
		handler, mockSvc := setupTestSearchHandler(t)

		mockSvc.searchFn = func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
			helpers.AssertEqual(t, query.Type, "text")
			return []domain.SearchResult{}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
	})

	t.Run("valid search types are accepted", func(t *testing.T) {
		handler, mockSvc := setupTestSearchHandler(t)

		for _, searchType := range []string{"text", "vector", "semantic"} {
			mockSvc.searchFn = func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
				return []domain.SearchResult{}, nil
			}

			r := httptest.NewRequest(http.MethodGet, "/search?q=golang&type="+searchType, nil)
			ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.Search(w, r)

			helpers.AssertEqual(t, w.Code, http.StatusOK)
		}
	})

	t.Run("limit parameter is parsed and capped at 100", func(t *testing.T) {
		handler, mockSvc := setupTestSearchHandler(t)

		mockSvc.searchFn = func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
			if limit > 100 {
				t.Fatalf("limit should be capped at 100, got %d", limit)
			}
			return []domain.SearchResult{}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang&limit=200", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
	})

	t.Run("invalid limit parameter uses default", func(t *testing.T) {
		handler, mockSvc := setupTestSearchHandler(t)

		mockSvc.searchFn = func(ctx context.Context, query domain.SearchQuery, userID string, limit int) ([]domain.SearchResult, error) {
			// Default is 10 when limit is invalid
			helpers.AssertEqual(t, limit, 10)
			return []domain.SearchResult{}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/search?q=golang&limit=invalid", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test-user")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Search(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
	})
}
