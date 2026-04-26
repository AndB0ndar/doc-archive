package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/models"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) domain.DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(
	ctx context.Context, doc *domain.Document,
) (string, error) {
	query := `
        INSERT INTO documents (
			user_id,
			title,
			authors,
			year,
			category,
			file_path,
			file_size,
			status
		)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `
	if doc.Status == "" {
		doc.Status = domain.DocumentStatusPending
	}
	var id string
	err := r.db.QueryRow(ctx, query,
		doc.UserID,
		doc.Title,
		doc.Authors,
		doc.Year,
		doc.Category,
		doc.FilePath,
		doc.FileSize,
		doc.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert document: %w", err)
	}
	doc.ID = id
	return id, nil
}

func (r *DocumentRepository) GetByID(
	ctx context.Context, id string, userID string,
) (*domain.Document, error) {
	query := `
        SELECT id, user_id, title, authors, year, category, file_path, file_size, status, created_at
        FROM documents
        WHERE id = $1 AND user_id = $2
    `
	var dbDoc models.DocumentDB
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&dbDoc.ID, &dbDoc.UserID,
		&dbDoc.Title, &dbDoc.Authors, &dbDoc.Year, &dbDoc.Category,
		&dbDoc.FilePath, &dbDoc.FileSize,
		&dbDoc.Status,
		&dbDoc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &domain.Document{
		ID:        dbDoc.ID,
		UserID:    dbDoc.UserID,
		Title:     dbDoc.Title,
		Authors:   dbDoc.Authors,
		Year:      dbDoc.Year,
		Category:  dbDoc.Category,
		FilePath:  dbDoc.FilePath,
		FileSize:  dbDoc.FileSize,
		Status:    domain.DocumentStatus(dbDoc.Status),
		CreatedAt: dbDoc.CreatedAt,
	}, nil
}

func (r *DocumentRepository) GetByUserID(
	ctx context.Context, userID string, limit, offset int,
) ([]*domain.Document, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
        SELECT id, user_id, title, authors, year, category, file_path, file_size, status, created_at
        FROM documents
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get documents by user: %w", err)
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var dbDoc models.DocumentDB
		if err := rows.Scan(
			&dbDoc.ID, &dbDoc.UserID,
			&dbDoc.Title, &dbDoc.Authors, &dbDoc.Year, &dbDoc.Category,
			&dbDoc.FilePath, &dbDoc.FileSize,
			&dbDoc.Status,
			&dbDoc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, &domain.Document{
			ID:        dbDoc.ID,
			UserID:    dbDoc.UserID,
			Title:     dbDoc.Title,
			Authors:   dbDoc.Authors,
			Year:      dbDoc.Year,
			Category:  dbDoc.Category,
			FilePath:  dbDoc.FilePath,
			FileSize:  dbDoc.FileSize,
			Status:    domain.DocumentStatus(dbDoc.Status),
			CreatedAt: dbDoc.CreatedAt,
		})
	}
	return docs, nil
}

func (r *DocumentRepository) UpdateStatus(
	ctx context.Context, docID string, status domain.DocumentStatus,
) error {
	query := `UPDATE documents SET status = $1 WHERE id = $2`
	cmdTag, err := r.db.Exec(ctx, query, status, docID)
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("document %s not found", docID)
	}
	return nil
}

func (r *DocumentRepository) Delete(
	ctx context.Context, id string, userID string,
) error {
	query := `DELETE FROM documents WHERE id = $1 AND user_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf(
			"document with id %s and user_id %s not found", id, userID,
		)
	}
	return nil
}
