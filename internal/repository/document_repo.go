package repository

import (
	"context"
	"fmt"

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
		RETURNING id, created_at
	`
	err := r.db.QueryRow(r.ctx, query,
		doc.Title, doc.Authors, doc.Year, doc.Category, doc.FilePath, doc.FileSize,
	).Scan(&doc.ID, &doc.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert document: %w", err)
	}

	return doc.ID, nil
}

func (r *DocumentRepository) GetByID(id int) (*models.Document, error) {
	query := `
		SELECT
			id,
			title,
			authors,
			year,
			category,
			file_path,
			file_size,
			created_at
		FROM documents WHERE id = $1
	`
	var doc models.Document
	err := r.db.QueryRow(r.ctx, query, id).Scan(
		&doc.ID, &doc.Title, &doc.Authors, &doc.Year, &doc.Category,
		&doc.FilePath, &doc.FileSize, &doc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}
