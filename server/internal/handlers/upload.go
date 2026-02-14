package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/middleware"
	"github.com/AndB0ndar/doc-archive/internal/service"
)

type UploadHandler struct {
	service *service.DocumentService
}

func NewUploadHandler(service *service.DocumentService) *UploadHandler {
	return &UploadHandler{
		service: service,
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Request size limit (50 MB)
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	// Parsing multipart/form-data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		slog.Error("failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("failed to get file from form", "error", err)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check extension
	if ext := strings.ToLower(filepath.Ext(header.Filename)); ext != ".pdf" {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Check MIME‑type (read first 512 byte)
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil && err != io.EOF {
		slog.Error("failed to read file header", "error", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	file.Seek(0, io.SeekStart)
	mimeType := http.DetectContentType(buf)
	if !strings.Contains(mimeType, "application/pdf") && !strings.Contains(mimeType, "application/x-pdf") {
		http.Error(w, "File is not a valid PDF", http.StatusBadRequest)
		return
	}

	params := &service.UploadParams{
		File:     file,
		Header:   header,
		Title:    strings.TrimSpace(r.FormValue("title")),
		Authors:  strings.TrimSpace(r.FormValue("authors")),
		Year:     strings.TrimSpace(r.FormValue("year")),
		Category: strings.TrimSpace(r.FormValue("category")),
		UserID:   userID,
	}

	id, err := h.service.Upload(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Document uploaded successfully. Processing started.",
		"status":  "pending",
	})
}
