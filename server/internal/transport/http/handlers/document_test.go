package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/transport/http/middleware"
	"github.com/AndB0ndar/doc-archive/test/helpers"
)

// mockDocumentService implements domain.DocumentService for testing
type mockDocumentService struct {
	getDocumentByIDFn        func(ctx context.Context, id, userID string) (*domain.Document, error)
	getUserDocumentsFn       func(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error)
	deleteDocumentFn         func(ctx context.Context, id, userID string) error
	getDocumentStatusFn      func(ctx context.Context, docID, userID string) (string, error)
	getDocumentDownloadURLFn func(ctx context.Context, docID, userID string) (string, time.Duration, error)
	uploadFn                 func(ctx context.Context, file multipart.File, title, authors, year, category string, userID string) (string, error)
	processDocumentFn        func(ctx context.Context, docID, objectKey string) error
}

func (m *mockDocumentService) GetDocumentByID(ctx context.Context, id, userID string) (*domain.Document, error) {
	if m.getDocumentByIDFn != nil {
		return m.getDocumentByIDFn(ctx, id, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDocumentService) GetUserDocuments(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
	if m.getUserDocumentsFn != nil {
		return m.getUserDocumentsFn(ctx, userID, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDocumentService) DeleteDocument(ctx context.Context, id, userID string) error {
	if m.deleteDocumentFn != nil {
		return m.deleteDocumentFn(ctx, id, userID)
	}
	return errors.New("not implemented")
}

func (m *mockDocumentService) GetDocumentStatus(ctx context.Context, docID, userID string) (string, error) {
	if m.getDocumentStatusFn != nil {
		return m.getDocumentStatusFn(ctx, docID, userID)
	}
	return "", errors.New("not implemented")
}

func (m *mockDocumentService) GetDocumentDownloadURL(ctx context.Context, docID, userID string) (string, time.Duration, error) {
	if m.getDocumentDownloadURLFn != nil {
		return m.getDocumentDownloadURLFn(ctx, docID, userID)
	}
	return "", 0, errors.New("not implemented")
}

func (m *mockDocumentService) Upload(ctx context.Context, file multipart.File, title, authors, year, category string, userID string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, file, title, authors, year, category, userID)
	}
	return "", errors.New("not implemented")
}

func (m *mockDocumentService) ProcessDocument(ctx context.Context, docID, objectKey string) error {
	if m.processDocumentFn != nil {
		return m.processDocumentFn(ctx, docID, objectKey)
	}
	return errors.New("not implemented")
}

func setupTestDocumentHandler(t *testing.T) (*DocumentHandler, *mockDocumentService) {
	t.Helper()
	mockSvc := &mockDocumentService{}
	log := logger.New("test")
	handler := NewDocumentHandler(mockSvc, log)
	return handler, mockSvc
}

func addUserIDToRequest(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

func addChiURLParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx)
	return r.WithContext(ctx)
}

func TestDocumentHandler_GetDocument(t *testing.T) {
	t.Run("valid request returns 200 with document", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		now := time.Now()
		mockSvc.getDocumentByIDFn = func(ctx context.Context, id, userID string) (*domain.Document, error) {
			helpers.AssertEqual(t, id, "doc-123")
			helpers.AssertEqual(t, userID, "test-user")
			return &domain.Document{
				ID:        "doc-123",
				UserID:    "test-user",
				Title:     "Test Document",
				FilePath:  "/path/to/doc.pdf",
				FileSize:  1024,
				Status:    domain.DocumentStatusDone,
				CreatedAt: now,
			}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)

		var resp models.DocumentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, resp.ID, "doc-123")
		helpers.AssertEqual(t, resp.Title, "Test Document")
	})

	t.Run("missing userID returns 401", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		w := httptest.NewRecorder()

		handler.GetDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	})

	t.Run("empty document ID returns 400", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/documents/", nil)
		r = addChiURLParam(r, "id", "")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
	})

	t.Run("document not found returns 404", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentByIDFn = func(ctx context.Context, id, userID string) (*domain.Document, error) {
			return nil, errors.New("document not found")
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusNotFound)
	})

	t.Run("document with nil optional fields returns 200", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentByIDFn = func(ctx context.Context, id, userID string) (*domain.Document, error) {
			return &domain.Document{
				ID:        "doc-123",
				UserID:    "test-user",
				Title:     "Test Document",
				FilePath:  "/path/to/doc.pdf",
				FileSize:  1024,
				Status:    domain.DocumentStatusDone,
				CreatedAt: time.Now(),
				Authors:   nil,
				Year:      nil,
				Category:  nil,
			}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)

		var resp models.DocumentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, resp.ID, "doc-123")
	})
}

func TestDocumentHandler_ListDocuments(t *testing.T) {
	t.Run("valid request returns 200 with documents", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		now := time.Now()
		mockSvc.getUserDocumentsFn = func(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
			helpers.AssertEqual(t, userID, "test-user")
			helpers.AssertEqual(t, limit, 20)
			helpers.AssertEqual(t, offset, 0)
			return []*domain.Document{
				{ID: "doc-1", Title: "Doc 1", FilePath: "/p1.pdf", FileSize: 100, Status: domain.DocumentStatusDone, CreatedAt: now},
				{ID: "doc-2", Title: "Doc 2", FilePath: "/p2.pdf", FileSize: 200, Status: domain.DocumentStatusPending, CreatedAt: now},
			}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents", nil)
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.ListDocuments(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)

		var resp []models.DocumentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, len(resp), 2)
	})

	t.Run("missing userID returns 401", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/documents", nil)
		w := httptest.NewRecorder()

		handler.ListDocuments(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getUserDocumentsFn = func(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
			return nil, errors.New("database error")
		}

		r := httptest.NewRequest(http.MethodGet, "/documents", nil)
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.ListDocuments(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("valid limit and offset query parameters", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getUserDocumentsFn = func(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
			helpers.AssertEqual(t, limit, 10)
			helpers.AssertEqual(t, offset, 5)
			return []*domain.Document{}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents?limit=10&offset=5", nil)
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.ListDocuments(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
	})

	t.Run("invalid limit defaults to 20", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getUserDocumentsFn = func(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
			helpers.AssertEqual(t, limit, 20)
			return []*domain.Document{}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents?limit=-1", nil)
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.ListDocuments(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
	})

	t.Run("negative offset defaults to 0", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getUserDocumentsFn = func(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
			helpers.AssertEqual(t, offset, 0)
			return []*domain.Document{}, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents?offset=-5", nil)
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.ListDocuments(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)
	})
}

func TestDocumentHandler_DeleteDocument(t *testing.T) {
	t.Run("valid delete returns 204", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.deleteDocumentFn = func(ctx context.Context, id, userID string) error {
			helpers.AssertEqual(t, id, "doc-123")
			helpers.AssertEqual(t, userID, "test-user")
			return nil
		}

		r := httptest.NewRequest(http.MethodDelete, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.DeleteDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusNoContent)
	})

	t.Run("missing userID returns 401", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodDelete, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		w := httptest.NewRecorder()

		handler.DeleteDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	})

	t.Run("empty document ID returns 400", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodDelete, "/documents/", nil)
		r = addChiURLParam(r, "id", "")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.DeleteDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
	})

	t.Run("document not found returns 404", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.deleteDocumentFn = func(ctx context.Context, id, userID string) error {
			return errors.New("document not found")
		}

		r := httptest.NewRequest(http.MethodDelete, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.DeleteDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusNotFound)
	})

	t.Run("other service error returns 500", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.deleteDocumentFn = func(ctx context.Context, id, userID string) error {
			return errors.New("internal error")
		}

		r := httptest.NewRequest(http.MethodDelete, "/documents/doc-123", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.DeleteDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusInternalServerError)
	})
}

func TestDocumentHandler_GetDocumentStatus(t *testing.T) {
	t.Run("valid request returns 200 with status", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentStatusFn = func(ctx context.Context, docID, userID string) (string, error) {
			return "done", nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/status", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocumentStatus(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)

		var resp models.DocumentStatusResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, resp.Status, "done")
	})

	t.Run("missing userID returns 401", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/status", nil)
		r = addChiURLParam(r, "id", "doc-123")
		w := httptest.NewRecorder()

		handler.GetDocumentStatus(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	})

	t.Run("document not found returns 404", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentStatusFn = func(ctx context.Context, docID, userID string) (string, error) {
			return "", errors.New("document not found")
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/status", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocumentStatus(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusNotFound)
	})
}

func TestDocumentHandler_DownloadDocument(t *testing.T) {
	t.Run("valid request redirects to presigned URL", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentDownloadURLFn = func(ctx context.Context, docID, userID string) (string, time.Duration, error) {
			return "http://example.com/presigned-url", time.Hour, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/download", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.DownloadDocument(w, r)

		// Should redirect (302)
		helpers.AssertEqual(t, w.Code, http.StatusFound)
		helpers.AssertEqual(t, w.Header().Get("Location"), "http://example.com/presigned-url")
	})

	t.Run("document not found returns 404", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentDownloadURLFn = func(ctx context.Context, docID, userID string) (string, time.Duration, error) {
			return "", 0, errors.New("document not found")
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/download", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.DownloadDocument(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusNotFound)
	})
}

func TestDocumentHandler_GetDocumentDownloadURL(t *testing.T) {
	t.Run("valid request returns 200 with URL", func(t *testing.T) {
		handler, mockSvc := setupTestDocumentHandler(t)

		mockSvc.getDocumentDownloadURLFn = func(ctx context.Context, docID, userID string) (string, time.Duration, error) {
			return "http://example.com/presigned-url", time.Hour, nil
		}

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/download-url", nil)
		r = addChiURLParam(r, "id", "doc-123")
		r = addUserIDToRequest(r, "test-user")
		w := httptest.NewRecorder()

		handler.GetDocumentDownloadURL(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusOK)

		var resp models.DocumentDownloadURLResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, resp.URL, "http://example.com/presigned-url")
		helpers.AssertEqual(t, resp.ExpiresIn, int64(3600))
	})

	t.Run("missing userID returns 401", func(t *testing.T) {
		handler, _ := setupTestDocumentHandler(t)

		r := httptest.NewRequest(http.MethodGet, "/documents/doc-123/download-url", nil)
		r = addChiURLParam(r, "id", "doc-123")
		w := httptest.NewRecorder()

		handler.GetDocumentDownloadURL(w, r)

		helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
	})
}
