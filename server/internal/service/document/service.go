package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/parser"
	"github.com/AndB0ndar/doc-archive/internal/storage"
)

type Service struct {
	docRepo   domain.DocumentRepository
	chunkRepo domain.ChunkRepository
	embedder  domain.EmbedderClient
	chunker   domain.ChunkerService
	storage   *storage.MinIOStorage
	logger    *slog.Logger
}

func New(
	docRepo domain.DocumentRepository,
	chunkRepo domain.ChunkRepository,
	embedder domain.EmbedderClient,
	chunker domain.ChunkerService,
	storage *storage.MinIOStorage,
	logger *slog.Logger,
) domain.DocumentService {
	return &Service{
		docRepo:   docRepo,
		chunkRepo: chunkRepo,
		embedder:  embedder,
		chunker:   chunker,
		storage:   storage,
		logger:    logger,
	}
}

func (s *Service) Upload(
	ctx context.Context,
	file multipart.File,
	title, authors, year, category string,
	userID string,
) (string, error) {
	// Required fields
	if title == "" {
		return "", fmt.Errorf("title is required")
	}

	// Optional fields
	var authorsPtr *string
	if authors != "" {
		authorsPtr = &authors
	}
	var yearPtr *int
	if year != "" {
		var y int
		if _, err := fmt.Sscanf(year, "%d", &y); err == nil && y > 0 && y <= time.Now().Year()+1 {
			yearPtr = &y
		}
	}
	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	// Generate uniqe name of file
	objectKey := uuid.New().String() + ".pdf"

	// Save in MinIO
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("failed to read file", "error", err)
		return "", fmt.Errorf("read file: %w", err)
	}
	reader := bytes.NewReader(fileBytes)
	err = s.storage.Upload(
		ctx,
		objectKey,
		reader,
		int64(len(fileBytes)),
		"application/pdf",
	)
	if err != nil {
		s.logger.Error("failed to upload to MinIO", "error", err)
		return "", fmt.Errorf("upload to storage: %w", err)
	}

	doc := &domain.Document{
		Title:    title,
		Authors:  authorsPtr,
		Year:     yearPtr,
		Category: categoryPtr,
		FilePath: objectKey,
		FileSize: int64(len(fileBytes)),
		UserID:   userID,
	}

	id, err := s.docRepo.Create(ctx, doc)
	if err != nil {
		s.logger.Error("failed to save document metadata", "error", err)
		_ = s.storage.Delete(ctx, objectKey) // remove file, if not save in DB
		return "", fmt.Errorf("save metadata: %w", err)
	}

	s.logger.Info(
		"document uploaded", "id", id, "title", title, "size", len(fileBytes),
	)

	go s.processDocument(id, objectKey)

	return id, nil
}

func (s *Service) processDocument(docID string, objectKey string) {
	ctx := context.Background()
	s.logger.Info(
		"starting document processing", "id", docID, "key", objectKey,
	)

	reader, err := s.storage.Download(ctx, objectKey)
	if err != nil {
		s.logger.Error(
			"failed to download from MinIO", "key", objectKey, "error", err,
		)
		return
	}
	defer reader.Close()

	tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		s.logger.Error("failed to create temp file", "error", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	if _, err := io.Copy(tmpFile, reader); err != nil {
		s.logger.Error("failed to copy to temp file", "error", err)
		return
	}

	text, err := parser.ExtractFromPDF(tmpFile.Name())
	if err != nil {
		s.logger.Error(
			"failed to extract text from PDF", "id", docID, "error", err,
		)
		return
	}

	chunksText, err := s.chunker.Split(text)
	if err != nil {
		s.logger.Error(
			"failed to split text into chunks", "id", docID, "error", err,
		)
		return
	}
	s.logger.Info("text chunked", "id", docID, "chunks", len(chunksText))

	chunks := make([]*domain.Chunk, 0, len(chunksText))
	for idx, chunkText := range chunksText {
		chunks = append(chunks, &domain.Chunk{
			DocumentID: docID,
			Index:      idx,
			Content:    chunkText,
			// Embedding will be filled in after vectorization
		})
	}

	if err := s.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		s.logger.Error("failed to save chunks", "id", docID, "error", err)
		return
	}

	for _, chunk := range chunks {
		embedResult, err := s.embedder.Embed(chunk.Content)
		if err != nil {
			s.logger.Error(
				"failed to get embedding",
				"doc_id", docID,
				"chunk_idx", chunk.Index,
				"error", err,
			)
			continue
		}
		if err := s.chunkRepo.UpdateEmbedding(
			ctx, chunk.ID, embedResult.Embeddings,
		); err != nil {
			s.logger.Error(
				"failed to update chunk with embedding",
				"doc_id", docID,
				"chunk_idx", chunk.Index,
				"error", err,
			)
		}
	}

	s.logger.Info("document processing completed", "id", docID)
}

func (s *Service) GetDocumentByID(ctx context.Context, id string, userID string) (*domain.Document, error) {
	doc, err := s.docRepo.GetByID(ctx, id, userID)
	if err != nil {
		s.logger.Error("failed to get document", "id", id, "user_id", userID, "error", err)
		return nil, err
	}
	return doc, nil
}

func (s *Service) GetUserDocuments(
	ctx context.Context, userID string, limit, offset int,
) ([]*domain.Document, error) {
	docs, err := s.docRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("failed to get user documents", "user_id", userID, "error", err)
		return nil, err
	}
	return docs, nil
}

func (s *Service) DeleteDocument(ctx context.Context, id string, userID string) error {
	doc, err := s.docRepo.GetByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("document not found or not owned: %w", err)
	}

	if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error("failed to delete file", "path", doc.FilePath, "error", err)
		return fmt.Errorf("delete file: %w", err)
	}

	if err := s.docRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete document record: %w", err)
	}

	s.logger.Info("document deleted", "id", id, "user_id", userID)
	return nil
}

func (s *Service) GetDocumentDownloadURL(
	ctx context.Context, docID, userID string,
) (string, time.Duration, error) {
	doc, err := s.docRepo.GetByID(ctx, docID, userID)
	if err != nil {
		return "", 0, fmt.Errorf("document not found: %w", err)
	}

	expiresIn := 15 * time.Minute // TODO: move timeout to config
	presignedURL, err := s.storage.GetPresignedURL(
		ctx, doc.FilePath, expiresIn,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL, expiresIn, nil
}
