package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndB0ndar/doc-archive/internal/models"
)

type DocumentRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{
		ctx: context.Background(),
		db:  db,
	}
}

func (r *DocumentRepository) Create(doc *models.Document) (int, error) {
	query := `
		INSERT INTO documents (title, authors, year, category, file_path, file_size)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	var id int
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(r.ctx, query,
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

func (r *DocumentRepository) UpdateFullText(id int, text string) error {
	query := `UPDATE documents SET full_text = $1, updated_at = NOW() WHERE id = $2`
	cmdTag, err := r.db.Exec(r.ctx, query, text, id)
	if err != nil {
		return fmt.Errorf("update full_text: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("document with id %d not found", id)
	}
	return nil
}
