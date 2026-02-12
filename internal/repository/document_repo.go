package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndB0ndar/doc-archive/internal/models"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) (int, error) {
	query := `
		INSERT INTO documents (title, authors, year, category, file_path, file_size)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	var id int
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		doc.Title, doc.Authors, doc.Year, doc.Category, doc.FilePath, doc.FileSize,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert document: %w", err)
	}

	doc.ID = id
	doc.CreatedAt = createdAt
	doc.UpdatedAt = updatedAt
	return id, nil
}
