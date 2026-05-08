package domain

import (
	"context"
	"io"
	"mime/multipart"
	"time"
)

type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusDone       DocumentStatus = "done"
	DocumentStatusError      DocumentStatus = "error"
)

type Document struct {
	ID        string
	UserID    string
	Title     string
	Authors   *string
	Year      *int
	Category  *string
	FilePath  string
	FileSize  int64
	Status    DocumentStatus
	CreatedAt time.Time
}

// validTransitions defines allowed Document status transitions.
var validTransitions = map[DocumentStatus]map[DocumentStatus]bool{
	DocumentStatusPending: {
		DocumentStatusProcessing: true,
		DocumentStatusError:      true,
		DocumentStatusPending:    true,
	},
	DocumentStatusProcessing: {
		DocumentStatusDone:       true,
		DocumentStatusError:      true,
		DocumentStatusProcessing: true,
	},
	DocumentStatusDone: {
		DocumentStatusDone: true,
	},
	DocumentStatusError: {
		DocumentStatusError: true,
	},
}

// CanTransitionTo checks whether the document can transition from its current
// status to the given target status.
func (d *Document) CanTransitionTo(target DocumentStatus) bool {
	transitions, ok := validTransitions[d.Status]
	if !ok {
		return false
	}
	return transitions[target]
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) (string, error)

	GetByID(ctx context.Context, id string, userID string) (*Document, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Document, error)

	UpdateStatus(ctx context.Context, docID string, status DocumentStatus) error

	Delete(ctx context.Context, id string, userID string) error
}

type DocumentService interface {
	Upload(ctx context.Context, file multipart.File, title, authors, year, category string, userID string) (string, error)

	GetDocumentByID(ctx context.Context, id string, userID string) (*Document, error)
	GetUserDocuments(ctx context.Context, userID string, limit, offset int) ([]*Document, error)

	GetDocumentStatus(ctx context.Context, docID, userID string) (string, error)

	ProcessDocument(ctx context.Context, docID, objectKey string) error

	DeleteDocument(ctx context.Context, id string, userID string) error

	GetDocumentDownloadURL(ctx context.Context, docID, userID string) (string, time.Duration, error)
}

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectKey string) error
	GetPresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error)
}

// TaskQueue defines the interface for task queue operations
type TaskQueue interface {
	EnqueueAny(task interface{}) (string, error)
}
