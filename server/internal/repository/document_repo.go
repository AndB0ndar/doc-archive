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
        INSERT INTO documents (title, authors, year, category, file_path, file_size, user_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at
    `
	err := r.db.QueryRow(r.ctx, query,
		doc.Title, doc.Authors, doc.Year, doc.Category, doc.FilePath, doc.FileSize, doc.UserID,
	).Scan(&doc.ID, &doc.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert document: %w", err)
	}
	return doc.ID, nil
}

func (r *DocumentRepository) GetByID(id, userID int) (*models.Document, error) {
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
		FROM documents WHERE id = $1 AND user_id = $2
	`
	var doc models.Document
	err := r.db.QueryRow(r.ctx, query, id, userID).Scan(
		&doc.ID, &doc.Title, &doc.Authors, &doc.Year, &doc.Category,
		&doc.FilePath, &doc.FileSize, &doc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepository) GetAll(
	userID, limit, offset int,
) ([]models.Document, error) {
	if limit <= 0 {
		limit = 20
	}
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
		FROM documents WHERE user_id = $1 ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(r.ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all documents: %w", err)
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(
			&d.ID, &d.Title, &d.Authors, &d.Year, &d.Category,
			&d.FilePath, &d.FileSize, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *DocumentRepository) Delete(id, userID int) error {
	query := `DELETE FROM documents WHERE id = $1 AND user_id = $2`
	cmdTag, err := r.db.Exec(r.ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("document with id %d not found", id)
	}
	return nil
}
