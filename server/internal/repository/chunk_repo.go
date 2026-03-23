package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/models"
)

type ChunkRepository struct {
	db *pgxpool.Pool
}

func NewChunkRepository(db *pgxpool.Pool) domain.ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) Create(
	ctx context.Context, chunk *domain.Chunk,
) error {
	dbChunk := &models.ChunkDB{
		DocumentID: chunk.DocumentID,
		Index:      chunk.Index,
		Content:    chunk.Content,
		Embedding:  chunk.Embedding,
	}
	query := `
		INSERT INTO chunks (document_id, index, content, embedding)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	vec := pgvector.NewVector(dbChunk.Embedding)
	err := r.db.QueryRow(ctx, query,
		dbChunk.DocumentID, dbChunk.Index, dbChunk.Content, vec,
	).Scan(&dbChunk.ID)
	if err != nil {
		return fmt.Errorf("insert chunk: %w", err)
	}
	chunk.ID = dbChunk.ID
	return nil
}

func (r *ChunkRepository) CreateBatch(
	ctx context.Context, chunks []*domain.Chunk,
) error {
	batch := &pgx.Batch{}
	for _, ch := range chunks {
		if ch.Embedding == nil || len(ch.Embedding) == 0 {
			query := `
				INSERT INTO chunks (document_id, index, content)
				VALUES ($1, $2, $3)
				RETURNING id
			`
			batch.Queue(query, ch.DocumentID, ch.Index, ch.Content)
		} else {
			query := `
				INSERT INTO chunks (document_id, index, content, embedding)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`
			batch.Queue(
				query,
				ch.DocumentID, ch.Index, ch.Content,
				pgvector.NewVector(ch.Embedding),
			)
		}
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i, ch := range chunks {
		var id string
		if err := br.QueryRow().Scan(&id); err != nil {
			return fmt.Errorf("insert chunk %d: %w", i, err)
		}
		ch.ID = id
	}
	return nil
}

func (r *ChunkRepository) UpdateEmbedding(
	ctx context.Context, chunkID string, embedding []float32,
) error {
	query := `
		UPDATE chunks
		SET embedding = $1
		WHERE id = $2
	`
	vec := pgvector.NewVector(embedding)
	_, err := r.db.Exec(ctx, query, vec, chunkID)
	if err != nil {
		return fmt.Errorf("update chunk embedding: %w", err)
	}
	return nil
}

func (r *ChunkRepository) FullTextSearch(
	ctx context.Context, query string, userID string, limit int,
) ([]*domain.ChunkSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	sqlQuery := `
        SELECT 
            c.id, c.document_id, c.content,
            similarity(c.content, $1) AS confidence
        FROM chunks c
		JOIN documents d ON c.document_id = d.id
		WHERE d.user_id = $2
		ORDER BY similarity(c.content, $1) DESC
        LIMIT $3
    ` // WHERE c.content % $1
	rows, err := r.db.Query(ctx, sqlQuery, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("full text search chunks: %w", err)
	}
	defer rows.Close()

	var results []*domain.ChunkSearchResult
	for rows.Next() {
		res := &domain.ChunkSearchResult{}
		if err := rows.Scan(
			&res.ID,
			&res.DocumentID,
			&res.Content,
			&res.Similarity,
		); err != nil {
			return nil, fmt.Errorf("scan chunk result: %w", err)
		}
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return results, nil
}

func (r *ChunkRepository) SemanticSearch(
	ctx context.Context, embedding []float32, userID string, limit int,
) ([]*domain.ChunkSearchResult, error) {
	vec := pgvector.NewVector(embedding)
	query := `
        SELECT
            c.id, c.document_id, c.content,
            1 - (c.embedding <=> $1) AS confidence
        FROM chunks c
        JOIN documents d ON c.document_id = d.id
        WHERE c.embedding IS NOT NULL AND d.user_id = $2
        ORDER BY c.embedding <=> $1
        LIMIT $3
    `
	rows, err := r.db.Query(ctx, query, vec, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search chunks: %w", err)
	}
	defer rows.Close()

	var results []*domain.ChunkSearchResult
	for rows.Next() {
		res := &domain.ChunkSearchResult{}
		if err := rows.Scan(
			&res.ID,
			&res.DocumentID,
			&res.Content,
			&res.Similarity,
		); err != nil {
			return nil, fmt.Errorf("scan chunk result: %w", err)
		}
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return results, nil
}
