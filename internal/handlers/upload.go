package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/pdfextractor"
	"github.com/AndB0ndar/doc-archive/internal/repository"
	"github.com/AndB0ndar/doc-archive/internal/vectorizer"
)

type UploadHandler struct {
	cfg       *config.Config
	repo      *repository.DocumentRepository
	embedderClient *vectorizer.Client
	uploadDir string
}

func NewUploadHandler(cfg *config.Config, repo *repository.DocumentRepository) *UploadHandler {
	return &UploadHandler{
		cfg:       cfg,
		repo:      repo,
		embedderClient: vectorizer.NewClient(cfg.EmbedderURL),
		uploadDir: cfg.UploadDir,
	}
}

// ServeHTTP process POST /upload
func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Required fields
	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Optional fields
	var authors *string
	if v := r.FormValue("authors"); v != "" {
		authors = &v
	}
	var year *int
	if v := r.FormValue("year"); v != "" {
		var y int
		if _, err := fmt.Sscanf(v, "%d", &y); err == nil && y > 0 && y <= time.Now().Year()+1 {
			year = &y
		}
	}
	var category *string
	if v := r.FormValue("category"); v != "" {
		category = &v
	}

	// Generate uniqe name of file
	uniqueID := uuid.New().String()
	filename := uniqueID + ".pdf"
	fullPath := filepath.Join(h.uploadDir, filename)

	// Create directory (if not exist)
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		slog.Error("failed to create upload directory", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Save file
	dst, err := os.Create(fullPath)
	if err != nil {
		slog.Error("failed to create destination file", "error", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		slog.Error("failed to copy file", "error", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create metadata in DB
	doc := &models.Document{
		Title:    title,
		Authors:  authors,
		Year:     year,
		Category: category,
		FilePath: fullPath,
		FileSize: written,
	}

	id, err := h.repo.Create(doc)
	if err != nil {
		slog.Error("failed to save document metadata", "error", err)
		os.Remove(fullPath) // remove file, if not save in DB
		http.Error(
			w,
			"Failed to save document metadata",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Document uploaded successfully. Processing started.",
		"status":  "pending",
	})

	slog.Info("document uploaded", "id", id, "title", title, "size", written)

	// Background processing (does not block the response)
	go h.processDocument(id, fullPath)
}

func (h *UploadHandler) processDocument(docID int, filePath string) {
	slog.Info("starting document processing", "id", docID, "path", filePath)

	text, err := pdfextractor.ExtractText(filePath)
	if err != nil {
		slog.Error("failed to extract text from PDF", "id", docID, "error", err)
		// TODO: can write error in database, in status field
		return
	}

	if err := h.repo.UpdateFullText(docID, text); err != nil {
		slog.Error("failed to update full_text", "id", docID, "error", err)
		return
	}

	slog.Info("document text extracted and saved", "id", docID, "text_length", len(text))

	maxTextLen := h.cfg.MaxTextLen
    truncated := text
    if len(truncated) > maxTextLen {
        truncated = truncated[:maxTextLen]
    }

    embedding, err := h.embedderClient.Embed(truncated)
    if err != nil {
        slog.Error("failed to get embedding", "id", docID, "error", err)
        return
    }

    if err := h.repo.UpdateEmbedding(docID, embedding); err != nil {
        slog.Error("failed to save embedding", "id", docID, "error", err)
        return
    }

    slog.Info("document embedding saved", "id", docID)
}
