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
