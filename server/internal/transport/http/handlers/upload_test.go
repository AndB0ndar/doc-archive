package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/transport/http/middleware"
	"github.com/AndB0ndar/doc-archive/test/helpers"
)

func setupTestUploadHandler(t *testing.T) (*UploadHandler, *mockDocumentService) {
	t.Helper()
	mockSvc := &mockDocumentService{}
	log := logger.New("test")
	handler := NewUploadHandler(mockSvc, log)
	return handler, mockSvc
}

// buildUploadRequest creates an HTTP request with multipart/form-data for testing uploads.
func buildUploadRequest(t *testing.T, userID, title, authors, year, category string, filename string, fileContent []byte) *http.Request {
	t.Helper()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Write form fields
	if title != "" {
		_ = w.WriteField("title", title)
	}
	if authors != "" {
		_ = w.WriteField("authors", authors)
	}
	if year != "" {
		_ = w.WriteField("year", year)
	}
	if category != "" {
		_ = w.WriteField("category", category)
	}

	// Write file
	if filename != "" && fileContent != nil {
		part, err := w.CreateFormFile("file", filename)
		helpers.AssertNoError(t, err)
		_, err = part.Write(fileContent)
		helpers.AssertNoError(t, err)
	}

	w.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", &b)
	r.Header.Set("Content-Type", w.FormDataContentType())

	if userID != "" {
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
		r = r.WithContext(ctx)
	}

	return r
}

func TestUploadHandler_ValidUpload(t *testing.T) {
	handler, mockSvc := setupTestUploadHandler(t)

	// Valid PDF content header
	pdfContent := []byte("%PDF-1.4 some pdf content\x00\x00")

	mockSvc.uploadFn = func(ctx context.Context, file multipart.File, title, authors, year, category string, userID string) (string, error) {
		helpers.AssertEqual(t, title, "Test Document")
		helpers.AssertEqual(t, authors, "John Doe")
		helpers.AssertEqual(t, year, "2024")
		helpers.AssertEqual(t, category, "Research")
		helpers.AssertEqual(t, userID, "test-user")
		return "doc-123", nil
	}

	r := buildUploadRequest(t, "test-user", "Test Document", "John Doe", "2024", "Research", "document.pdf", pdfContent)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusCreated)

	var resp models.UploadResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, resp.DocumentID, "doc-123")
	helpers.AssertEqual(t, resp.Status, "pending")
}

func TestUploadHandler_MissingUserID(t *testing.T) {
	handler, _ := setupTestUploadHandler(t)

	r := buildUploadRequest(t, "", "Test Document", "", "", "", "document.pdf", []byte("%PDF-1.4 content"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusUnauthorized)
}

func TestUploadHandler_MissingFile(t *testing.T) {
	handler, _ := setupTestUploadHandler(t)

	r := buildUploadRequest(t, "test-user", "Test Document", "", "", "", "", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
}

func TestUploadHandler_InvalidExtension(t *testing.T) {
	handler, _ := setupTestUploadHandler(t)

	r := buildUploadRequest(t, "test-user", "Test Document", "", "", "", "document.txt", []byte("not a pdf"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
}

func TestUploadHandler_InvalidMimeType(t *testing.T) {
	handler, _ := setupTestUploadHandler(t)

	// Create a .pdf file but with non-PDF content that has a different MIME type
	r := buildUploadRequest(t, "test-user", "Test Document", "", "", "", "document.pdf", []byte("This is not a PDF file at all"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
}

func TestUploadHandler_MissingTitle(t *testing.T) {
	handler, _ := setupTestUploadHandler(t)

	r := buildUploadRequest(t, "test-user", "", "", "", "", "document.pdf", []byte("%PDF-1.4 content"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusBadRequest)
}

func TestUploadHandler_ServiceError(t *testing.T) {
	handler, mockSvc := setupTestUploadHandler(t)

	mockSvc.uploadFn = func(ctx context.Context, file multipart.File, title, authors, year, category string, userID string) (string, error) {
		return "", errors.New("storage unavailable")
	}

	r := buildUploadRequest(t, "test-user", "Test Document", "", "", "", "document.pdf", []byte("%PDF-1.4 content"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	helpers.AssertEqual(t, w.Code, http.StatusInternalServerError)
}
