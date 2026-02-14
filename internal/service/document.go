package service

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/repository"
)

type DocumentService struct {
	cfg            *config.Config
	docRepo        *repository.DocumentRepository
	chunkRepo      *repository.ChunkRepository
	embedderClient *Embedder
	uploadDir      string
}

func NewDocumentService(
	cfg *config.Config,
	docRepo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	embedderClient *Embedder,
) *DocumentService {
	return &DocumentService{
		cfg:            cfg,
		docRepo:        docRepo,
		chunkRepo:      chunkRepo,
		embedderClient: embedderClient,
		uploadDir:      cfg.UploadDir,
	}
}

type UploadParams struct {
	File     multipart.File
	Header   *multipart.FileHeader
	Title    string
	Authors  string
	Year     string
	Category string
}

func (s *DocumentService) Upload(params *UploadParams) (int, error) {
	// Required fields
	if params.Title == "" {
		return 0, fmt.Errorf("title is required")
	}

	// Optional fields
	var authorsPtr *string
	if params.Authors != "" {
		authorsPtr = &params.Authors
	}
	var yearPtr *int
	if params.Year != "" {
		var y int
		if _, err := fmt.Sscanf(params.Year, "%d", &y); err == nil && y > 0 && y <= time.Now().Year()+1 {
			yearPtr = &y
		}
	}
	var categoryPtr *string
	if params.Category != "" {
		categoryPtr = &params.Category
	}

	// Generate uniqe name of file
	uniqueID := uuid.New().String()
	filename := uniqueID + ".pdf"
	fullPath := filepath.Join(s.uploadDir, filename)

	// Create directory (if not exist)
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		slog.Error("failed to create upload directory", "error", err)
		return 0, fmt.Errorf("create upload dir: %w", err)
	}

	// Save file
	dst, err := os.Create(fullPath)
	if err != nil {
		slog.Error("failed to create destination file", "error", err)
		return 0, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()
	written, err := io.Copy(dst, params.File)
	if err != nil {
		slog.Error("failed to copy file", "error", err)
		return 0, fmt.Errorf("copy file: %w", err)
	}

	doc := &models.Document{
		Title:    params.Title,
		Authors:  authorsPtr,
		Year:     yearPtr,
		Category: categoryPtr,
		FilePath: fullPath,
		FileSize: written,
	}

	id, err := s.docRepo.Create(doc)
	if err != nil {
		slog.Error("failed to save document metadata", "error", err)
		os.Remove(fullPath) // remove file, if not save in DB
		return 0, fmt.Errorf("save metadata: %w", err)
	}
	slog.Info("document uploaded", "id", id, "title", params.Title, "size", written)

	go s.processDocument(id, fullPath)

	return id, nil
}

func (s *DocumentService) processDocument(docID int, filePath string) {
	slog.Info("starting document processing", "id", docID, "path", filePath)

	text, err := ExtractText(filePath)
	if err != nil {
		slog.Error("failed to extract text from PDF", "id", docID, "error", err)
		// TODO: can write error in database, in status field
		return
	}

	chunkSize := s.cfg.ChunkSize
	overlap := s.cfg.ChunkOverlap
	chunks := Chunk(text, chunkSize, overlap)
	slog.Info("text chunked", "id", docID, "chunks", len(chunks))

	for idx, chunkText := range chunks {
		embedding, err := s.embedderClient.Embed(chunkText)
		if err != nil {
			slog.Error("failed to get embedding for chunk", "doc_id", docID, "chunk_idx", idx, "error", err)
			continue
		}
		chunk := &models.Chunk{
			DocumentID: docID,
			ChunkIndex: idx,
			Content:    chunkText,
			Embedding:  embedding,
		}
		if _, err := s.chunkRepo.Create(chunk); err != nil {
			slog.Error("failed to save chunk", "doc_id", docID, "chunk_idx", idx, "error", err)
		}
	}
	slog.Info("document chunks processed", "id", docID)
}
