package domain

import (
	"context"
	"mime/multipart"
	"time"
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
	CreatedAt time.Time
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) (string, error)
	GetByID(ctx context.Context, id string, userID string) (*Document, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Document, error)
	Delete(ctx context.Context, id string, userID string) error
}

type DocumentService interface {
	Upload(ctx context.Context, file multipart.File, title, authors, year, category string, userID string) (string, error)
	GetDocumentByID(ctx context.Context, id string, userID string) (*Document, error)
	GetUserDocuments(ctx context.Context, userID string, limit, offset int) ([]*Document, error)
	DeleteDocument(ctx context.Context, id string, userID string) error
	GetDocumentDownloadURL(ctx context.Context, docID, userID string) (string, time.Duration, error)
}
