package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/AndB0ndar/doc-archive/internal/middleware"
	"github.com/AndB0ndar/doc-archive/internal/repository"
)

type DocumentHandler struct {
	repo *repository.DocumentRepository
}

func NewDocumentHandler(repo *repository.DocumentRepository) *DocumentHandler {
	return &DocumentHandler{repo: repo}
}

// GetDocument возвращает информацию о конкретном документе.
// @Summary      Получить документ
// @Description  Возвращает метаданные документа по ID.
// @Tags         documents
// @Produce      json
// @Param        id path int true "ID документа"
// @Success      200  {object}  models.Document
// @Failure      400  {string}  string "Invalid document ID"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      404  {string}  string "Document not found"
// @Security     BearerAuth
// @Router       /documents/{id} [get]
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	doc, err := h.repo.GetByID(id, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// ListDocuments возвращает список всех документов (с пагинацией).
// @Summary      Список документов
// @Description  Возвращает метаданные всех загруженных документов.
// @Tags         documents
// @Produce      json
// @Param        limit query int false "Максимальное количество документов на странице (по умолчанию 20, макс 100)"
// @Param        offset query int false "Смещение от начала списка (по умолчанию 0)"
// @Success      200  {array}   models.Document
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Failed to fetch documents"
// @Security     BearerAuth
// @Router       /documents [get]
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(offsetStr)
	if offset < 0 {
		offset = 0
	}

	docs, err := h.repo.GetAll(userID, limit, offset)
	if err != nil {
		slog.Error("failed to list documents", "error", err)
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(docs); err != nil {
		slog.Error("failed to encode documents", "error", err)
	}
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
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	doc, err := h.repo.GetByID(id, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
		slog.Error("failed to delete file", "path", doc.FilePath, "error", err)
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	if err := h.repo.Delete(id, userID); err != nil {
		slog.Error("failed to delete document from DB", "id", id, "error", err)
		http.Error(w, "Failed to delete document from database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
