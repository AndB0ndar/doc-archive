package models

import (
	"time"
)

type Chunk struct {
	ID         int64     `json:"id"`
	DocumentID int       `json:"document_id"`
	ChunkIndex int       `json:"chunk_index"`
	Content    string    `json:"content"`
	Embedding  []float32 `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChunkSearchResponse struct {
	ChunkID    int64     `json:"chunk_id"`
	DocumentID int       `json:"document_id"`
	ChunkIndex int       `json:"chunk_index"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Similarity float64   `json:"similarity"` // from 0 to 1
	Title      string    `json:"title"`
	Authors    *string   `json:"authors,omitempty"`
	Year       *int      `json:"year,omitempty"`
	Category   *string   `json:"category,omitempty"`
}

