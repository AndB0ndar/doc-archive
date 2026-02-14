package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/AndB0ndar/doc-archive/internal/models"
)

type ChunkRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewChunkRepository(db *pgxpool.Pool) *ChunkRepository {
	return &ChunkRepository{
		ctx: context.Background(),
		db:  db,
	}
}

func (r *ChunkRepository) Create(chunk *models.Chunk) (int64, error) {
	query := `
		INSERT INTO chunks (document_id, chunk_index, content, embedding)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	vec := pgvector.NewVector(chunk.Embedding)
	err := r.db.QueryRow(r.ctx, query,
		chunk.DocumentID, chunk.ChunkIndex, chunk.Content, vec,
	).Scan(&chunk.ID, &chunk.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert chunk: %w", err)
	}

	return chunk.ID, nil
}

func (r *ChunkRepository) FullTextSearchChunks(
	query string, limit int,
) ([]models.ChunkSearchResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	sqlQuery := `
        SELECT 
            c.id,
			c.document_id,
			c.chunk_index,
			c.content,
			c.created_at,
            similarity(c.content, $1) AS similarity,
            d.title,
			d.authors,
			d.year,
			d.category
        FROM chunks c
			JOIN documents d ON c.document_id = d.id
			ORDER BY similarity(c.content, $1) DESC
        LIMIT $2
    ` // WHERE c.content % $1
	rows, err := r.db.Query(r.ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("full text search chunks: %w", err)
	}
	defer rows.Close()

	var results []models.ChunkSearchResponse
	for rows.Next() {
		var r models.ChunkSearchResponse
		if err := rows.Scan(
			&r.ChunkID, &r.DocumentID, &r.ChunkIndex, &r.Content,
			&r.CreatedAt,
			&r.Similarity,
			&r.Title, &r.Authors, &r.Year, &r.Category,
		); err != nil {
			return nil, fmt.Errorf("scan chunk result: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}

func (r *ChunkRepository) SemanticSearchChunks(
	embedding []float32, limit int,
) ([]models.ChunkSearchResponse, error) {
	vec := pgvector.NewVector(embedding)
	query := `
		SELECT 
			c.id, c.document_id, c.chunk_index, c.content, c.created_at,
			1 - (c.embedding <=> $1) AS similarity,
			d.title, d.authors, d.year, d.category
		FROM chunks c
		JOIN documents d ON c.document_id = d.id
		WHERE c.embedding IS NOT NULL
		ORDER BY c.embedding <=> $1
		LIMIT $2
	`
	rows, err := r.db.Query(r.ctx, query, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search chunks: %w", err)
	}
	defer rows.Close()

	var results []models.ChunkSearchResponse
	for rows.Next() {
		var r models.ChunkSearchResponse
		if err := rows.Scan(
			&r.ChunkID, &r.DocumentID, &r.ChunkIndex, &r.Content, &r.CreatedAt,
			&r.Similarity,
			&r.Title, &r.Authors, &r.Year, &r.Category,
		); err != nil {
			return nil, fmt.Errorf("scan chunk result: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}
