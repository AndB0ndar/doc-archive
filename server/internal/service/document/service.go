package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/parser"
	"github.com/AndB0ndar/doc-archive/internal/tasks"
)

type service struct {
	docRepo     domain.DocumentRepository
	chunkRepo   domain.ChunkRepository
	embedder    domain.EmbedderClient
	chunker     domain.ChunkerService
	fileStorage domain.FileStorage
	taskQueue   domain.TaskQueue
	logger      *logger.Logger
}

func New(
	docRepo domain.DocumentRepository,
	chunkRepo domain.ChunkRepository,
	embedder domain.EmbedderClient,
	chunker domain.ChunkerService,
	fileStorage domain.FileStorage,
	taskQueue domain.TaskQueue,
	logger *logger.Logger,
) domain.DocumentService {
	return &service{
		docRepo:     docRepo,
		chunkRepo:   chunkRepo,
		embedder:    embedder,
		chunker:     chunker,
		fileStorage: fileStorage,
		taskQueue:   taskQueue,
		logger:      logger,
	}
}

func (s *service) enqueueProcessDocument(
	ctx context.Context, docID, objectKey string,
) error {
	task, err := tasks.NewProcessDocumentTask(docID, objectKey)
	if err != nil {
		return err
	}
	_, err = s.taskQueue.EnqueueAny(task)
	return err
}

func (s *service) Upload(
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
	err = s.fileStorage.Upload(
		ctx,
		objectKey,
		reader,
		int64(len(fileBytes)),
		"application/pdf",
	)
	if err != nil {
		s.logger.Error("failed to upload to MinIO", "error", err)
		return "", fmt.Errorf("upload to minio: %w", err)
	}

	doc := &domain.Document{
		Title:    title,
		Authors:  authorsPtr,
		Year:     yearPtr,
		Category: categoryPtr,
		FilePath: objectKey,
		FileSize: int64(len(fileBytes)),
		Status:   domain.DocumentStatusPending,
		UserID:   userID,
	}

	id, err := s.docRepo.Create(ctx, doc)
	if err != nil {
		s.logger.Error("failed to save document metadata", "error", err)
		_ = s.fileStorage.Delete(ctx, objectKey) // remove file, if not save in DB
		return "", fmt.Errorf("save metadata: %w", err)
	}

	s.logger.Info(
		"document uploaded", "id", id, "title", title, "size", len(fileBytes),
	)

	if err := s.enqueueProcessDocument(ctx, id, objectKey); err != nil {
		s.logger.Error("failed to enqueue task", "error", err)
		_ = s.docRepo.UpdateStatus(ctx, id, domain.DocumentStatusError)
	} else {
		s.logger.Info("task enqueued", "doc_id", id)
	}

	return id, nil
}

func (s *service) ProcessDocument(
	ctx context.Context, docID, objectKey string,
) error {
	return s.executeWithStatus(ctx, docID, func() error {
		s.logger.Info(
			"starting document processing",
			"doc_id", docID, "object_key", objectKey,
		)

		reader, err := s.fileStorage.Download(ctx, objectKey)
		if err != nil {
			return fmt.Errorf("download from minio: %w", err)
		}
		defer reader.Close()

		tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, reader); err != nil {
			return fmt.Errorf("copy to temp: %w", err)
		}

		text, err := parser.ExtractFromPDF(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("extract text: %w", err)
		}
		s.logger.Debug("text extracted", "doc_id", docID, "length", len(text))

		chunkTexts, err := s.chunker.Split(text)
		if err != nil {
			return fmt.Errorf("split chunks: %w", err)
		}
		s.logger.Info(
			"text chunked", "doc_id", docID, "chunks", len(chunkTexts),
		)

		chunks := make([]*domain.Chunk, 0, len(chunkTexts))
		for idx, ct := range chunkTexts {
			chunks = append(chunks, &domain.Chunk{
				DocumentID: docID,
				Index:      idx,
				Content:    ct,
			})
		}

		if err := s.chunkRepo.CreateBatch(ctx, chunks); err != nil {
			return fmt.Errorf("save chunks: %w", err)
		}

		for _, chunk := range chunks {
			embResp, err := s.embedder.Embed(chunk.Content)
			if err != nil {
				s.logger.Error(
					"embedding failed",
					"doc_id", docID,
					"chunk_idx", chunk.Index,
					"error", err,
				)
				continue
			}
			if err := s.chunkRepo.UpdateEmbedding(
				ctx, chunk.ID, embResp.Embeddings,
			); err != nil {
				s.logger.Error(
					"update embedding failed",
					"doc_id", docID,
					"chunk_idx", chunk.Index,
					"error", err,
				)
			}
		}

		s.logger.Info("document processing completed", "doc_id", docID)
		return nil
	})
}

func (s *service) executeWithStatus(
	ctx context.Context, docID string, fn func() error,
) (err error) {
	if err := s.docRepo.UpdateStatus(
		ctx, docID, domain.DocumentStatusProcessing,
	); err != nil {
		s.logger.Error(
			"failed to set status processing", "doc_id", docID, "error", err,
		)
		return err
	}

	err = fn()
	if err != nil {
		s.logger.Error("processing failed", "doc_id", docID, "error", err)
		if updateErr := s.docRepo.UpdateStatus(
			ctx, docID, domain.DocumentStatusError,
		); updateErr != nil {
			s.logger.Error(
				"failed to set status error",
				"doc_id", docID, "error", updateErr,
			)
		}
		return err
	}

	if err := s.docRepo.UpdateStatus(
		ctx, docID, domain.DocumentStatusDone,
	); err != nil {
		s.logger.Error(
			"failed to set status done", "doc_id", docID, "error", err,
		)
	}

	return nil
}

func (s *service) GetDocumentByID(
	ctx context.Context, id string, userID string,
) (*domain.Document, error) {
	doc, err := s.docRepo.GetByID(ctx, id, userID)
	if err != nil {
		s.logger.Error(
			"failed to get document",
			"id", id, "user_id", userID, "error", err,
		)
		return nil, err
	}
	return doc, nil
}

func (s *service) GetUserDocuments(
	ctx context.Context, userID string, limit, offset int,
) ([]*domain.Document, error) {
	docs, err := s.docRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error(
			"failed to get user documents",
			"user_id", userID, "error", err,
		)
		return nil, err
	}
	return docs, nil
}

func (s *service) GetDocumentStatus(
	ctx context.Context, docID, userID string,
) (string, error) {
	doc, err := s.docRepo.GetByID(ctx, docID, userID)
	if err != nil {
		return "", err
	}
	return string(doc.Status), nil
}

func (s *service) DeleteDocument(
	ctx context.Context, id string, userID string,
) error {
	doc, err := s.docRepo.GetByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("document not found or not owned: %w", err)
	}

	if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error(
			"failed to delete file", "path", doc.FilePath, "error", err,
		)
		return fmt.Errorf("delete file: %w", err)
	}

	if err := s.docRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete document record: %w", err)
	}

	s.logger.Info("document deleted", "id", id, "user_id", userID)
	return nil
}

func (s *service) GetDocumentDownloadURL(
	ctx context.Context, docID, userID string,
) (string, time.Duration, error) {
	doc, err := s.docRepo.GetByID(ctx, docID, userID)
	if err != nil {
		return "", 0, fmt.Errorf("document not found: %w", err)
	}

	expiresIn := 15 * time.Minute // TODO: move timeout to config
	presignedURL, err := s.fileStorage.GetPresignedURL(
		ctx, doc.FilePath, expiresIn,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL, expiresIn, nil
}
