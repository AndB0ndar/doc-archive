package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/AndB0ndar/doc-archive/internal/repository"
)

type DocumentHandler struct {
	repo *repository.DocumentRepository
}

func NewDocumentHandler(repo *repository.DocumentRepository) *DocumentHandler {
	return &DocumentHandler{repo: repo}
}

func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	doc, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
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

	docs, err := h.repo.GetAll(limit, offset)
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

func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	doc, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
		slog.Error("failed to delete file", "path", doc.FilePath, "error", err)
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		slog.Error("failed to delete document from DB", "id", id, "error", err)
		http.Error(w, "Failed to delete document from database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
