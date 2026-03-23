package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/middleware"
	"github.com/AndB0ndar/doc-archive/internal/models"
)

type DocumentHandler struct {
	docService domain.DocumentService
	logger     *slog.Logger
}

func NewDocumentHandler(
	docService domain.DocumentService, logger *slog.Logger,
) *DocumentHandler {
	return &DocumentHandler{
		docService: docService,
		logger:     logger,
	}
}

// GetDocument возвращает информацию о конкретном документе.
// @Summary      Получить документ
// @Description  Возвращает метаданные документа по ID.
// @Tags         documents
// @Produce      json
// @Param        id path int true "ID документа"
// @Success      200  {object}  models.DocumentResponse
// @Failure      400  {string}  string "Invalid document ID"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      404  {string}  string "Document not found"
// @Security     BearerAuth
// @Router       /documents/{id} [get]
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("GetDocument request started")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		logger.Warn("GetDocument: unauthorized - userID missing in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logger = logger.With(slog.String("user_id", userID))

	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Warn("GetDocument: invalid document ID (empty)")
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}
	logger = logger.With(slog.String("document_id", id))
	logger.Debug("GetDocument: parameters validated")

	logger.Debug("GetDocument: calling docService.GetDocumentByID")
	doc, err := h.docService.GetDocumentByID(r.Context(), id, userID)
	if err != nil {
		logger.Error("GetDocument: document not found", slog.Any("error", err))
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	response := mapDocumentToResponse(doc)
	logger.Debug(
		"GetDocument: successfully retrieved document",
		slog.String("title", response.Title),
	)

	w.Header().Set("Content-Type", "application/json")
	logger.Debug("GetDocument: encoding response")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(
			"GetDocument: failed to encode response", slog.Any("error", err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info(
		"GetDocument: completed", slog.Duration("duration", time.Since(start)),
	)
}

// ListDocuments возвращает список всех документов (с пагинацией).
// @Summary      Список документов
// @Description  Возвращает метаданные всех загруженных документов.
// @Tags         documents
// @Produce      json
// @Param        limit query int false "Максимальное количество документов на странице (по умолчанию 20, макс 100)"
// @Param        offset query int false "Смещение от начала списка (по умолчанию 0)"
// @Success      200  {array}   models.DocumentResponse
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Failed to fetch documents"
// @Security     BearerAuth
// @Router       /documents [get]
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("ListDocuments request started")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		logger.Warn("ListDocuments: unauthorized - userID missing in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logger = logger.With(slog.String("user_id", userID))

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		original := limit
		limit = 20
		logger.Debug("ListDocuments: adjusted limit",
			slog.Int("original", original),
			slog.Int("adjusted", limit))
	}
	offset, _ := strconv.Atoi(offsetStr)
	if offset < 0 {
		original := offset
		offset = 0
		logger.Debug("ListDocuments: adjusted offset",
			slog.Int("original", original),
			slog.Int("adjusted", offset))
	}
	logger = logger.With(slog.Int("limit", limit), slog.Int("offset", offset))
	logger.Debug("ListDocuments: parameters validated")

	logger.Debug("ListDocuments: calling docService.GetUserDocuments")
	docs, err := h.docService.GetUserDocuments(
		r.Context(), userID, limit, offset,
	)
	if err != nil {
		logger.Error(
			"ListDocuments: failed to fetch documents", slog.Any("error", err),
		)
		http.Error(
			w, "Failed to fetch documents", http.StatusInternalServerError,
		)
		return
	}

	response := make([]models.DocumentResponse, len(docs))
	for i, doc := range docs {
		response[i] = mapDocumentToResponse(doc)
	}
	logger.Debug(
		"ListDocuments: successfully retrieved documents",
		slog.Int("count", len(docs)),
	)

	w.Header().Set("Content-Type", "application/json")
	logger.Debug("ListDocuments: encoding response")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(
			"ListDocuments: failed to encode response", slog.Any("error", err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info(
		"ListDocuments: completed",
		slog.Duration("duration", time.Since(start)),
	)
}

// DeleteDocument удаляет документ и связанные файлы.
// @Summary      Удалить документ
// @Description  Удаляет документ по ID и его PDF-файл.
// @Tags         documents
// @Param        id path int true "ID документа"
// @Success      204  "No Content"
// @Failure      400  {string}  string "Invalid document ID"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      404  {string}  string "Document not found"
// @Failure      500  {string}  string "Failed to delete file or database record"
// @Security     BearerAuth
// @Router       /documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(
	w http.ResponseWriter, r *http.Request,
) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("DeleteDocument request started")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		logger.Warn("DeleteDocument: unauthorized - userID missing in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logger = logger.With(slog.String("user_id", userID))

	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Warn("DeleteDocument: invalid document ID (empty)")
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}
	logger = logger.With(slog.String("document_id", id))
	logger.Debug("DeleteDocument: parameters validated")

	logger.Debug("DeleteDocument: calling docService.DeleteDocument")
	if err := h.docService.DeleteDocument(
		r.Context(), id, userID,
	); err != nil {
		logger.Error(
			"DeleteDocument: failed to delete document",
			slog.Any("error", err),
		)
		http.Error(
			w, "Failed to delete document", http.StatusInternalServerError,
		)
		return
	}

	logger.Info(
		"DeleteDocument: document deleted successfully",
		slog.Duration("duration", time.Since(start)),
	)
	w.WriteHeader(http.StatusNoContent)
}

func mapDocumentToResponse(doc *domain.Document) models.DocumentResponse {
	return models.DocumentResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		Authors:   doc.Authors,
		Year:      doc.Year,
		FilePath:  doc.FilePath,
		Category:  doc.Category,
		CreatedAt: doc.CreatedAt,
	}
}
