package domain

import (
	"context"
)

type Chunk struct {
	ID         string
	DocumentID string
	Content    string
	Embedding  []float32
	Index      int
}

type ChunkSearchResult struct {
	Chunk
	Similarity float64
}

type ChunkRepository interface {
	Create(ctx context.Context, chunk *Chunk) error // TODO remove it
	CreateBatch(ctx context.Context, chunks []*Chunk) error
	UpdateEmbedding(ctx context.Context, chunkID string, embedding []float32) error
	FullTextSearch(ctx context.Context, query string, userID string, limit int) ([]*ChunkSearchResult, error)
	SemanticSearch(ctx context.Context, embedding []float32, userID string, limit int) ([]*ChunkSearchResult, error)
}

type ChunkerService interface {
	Split(text string) ([]string, error)
}
