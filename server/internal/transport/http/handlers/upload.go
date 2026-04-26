package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/transport/http/middleware"
)

type UploadHandler struct {
	docService domain.DocumentService
	logger     *logger.Logger
}

func NewUploadHandler(
	docService domain.DocumentService, logger *logger.Logger,
) *UploadHandler {
	return &UploadHandler{
		docService: docService,
		logger:     logger,
	}
}

// Upload загружает PDF-файл и запускает обработку.
// @Summary      Загрузка PDF
// @Description  Загружает PDF, сохраняет метаданные и запускает фоновую обработку.
// @Tags         documents
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "PDF-файл"
// @Param        title formData string true "Название документа"
// @Param        authors formData string false "Авторы"
// @Param        year formData string false "Год публикации"
// @Param        category formData string false "Категория"
// @Success      201  {object}  models.UploadResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /upload [post]
func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("Upload request started")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		logger.Warn("Upload: unauthorized - userID missing in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logger = logger.With(slog.String("user_id", userID))

	// Request size limit (50 MB)
	logger.Debug("Upload: applying max bytes reader (50MB)")
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	// Parsing multipart/form-data
	logger.Debug("Upload: parsing multipart form (32MB)")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		logger.Error(
			"Upload: failed to parse multipart form", slog.Any("error", err),
		)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	logger.Debug("Upload: multipart form parsed successfully")

	// Get file
	logger.Debug("Upload: retrieving file from form")
	file, header, err := r.FormFile("file")
	if err != nil {
		logger.Error(
			"Upload: failed to get file from form", slog.Any("error", err),
		)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	logger = logger.With(
		slog.String("filename", header.Filename),
		slog.Int64("size", header.Size),
	)
	logger.Debug("Upload: file retrieved")

	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	logger.Debug(
		"Upload: checking file extension", slog.String("extension", ext),
	)
	if ext != ".pdf" {
		logger.Warn(
			"Upload: invalid file extension", slog.String("extension", ext),
		)
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Check MIME‑type (read first 512 bytes)
	logger.Debug("Upload: reading file header for MIME detection")
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil && err != io.EOF {
		logger.Error(
			"Upload: failed to read file header", slog.Any("error", err),
		)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	file.Seek(0, io.SeekStart)
	mimeType := http.DetectContentType(buf)
	logger.Debug(
		"Upload: detected MIME type", slog.String("mime_type", mimeType),
	)
	if !strings.Contains(mimeType, "application/pdf") && !strings.Contains(mimeType, "application/x-pdf") {
		logger.Warn(
			"Upload: file is not a valid PDF",
			slog.String("mime_type", mimeType),
		)
		http.Error(w, "File is not a valid PDF", http.StatusBadRequest)
		return
	}

	// Extract metadata
	title := strings.TrimSpace(r.FormValue("title"))
	authors := strings.TrimSpace(r.FormValue("authors"))
	year := strings.TrimSpace(r.FormValue("year"))
	category := strings.TrimSpace(r.FormValue("category"))
	logger.Debug("Upload: extracted metadata",
		slog.String("title", title),
		slog.String("authors", authors),
		slog.String("year", year),
		slog.String("category", category),
	)

	if title == "" {
		logger.Warn("Upload: title is empty")
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	logger.Debug("Upload: calling docService.Upload")
	docID, err := h.docService.Upload(
		r.Context(), file, title, authors, year, category, userID,
	)
	if err != nil {
		logger.Error("Upload: upload failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger = logger.With(slog.String("document_id", docID))
	logger.Debug("Upload: document created successfully")

	w.Header().Set("Content-Type", "application/json")
	logger.Debug("Upload: encoding response")
	if err := json.NewEncoder(w).Encode(models.UploadResponse{
		DocumentID: docID,
		Status:     "pending",
	}); err != nil {
		logger.Error(
			"Upload: failed to encode response", slog.Any("error", err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info(
		"Upload: completed", slog.Duration("duration", time.Since(start)),
	)
}
