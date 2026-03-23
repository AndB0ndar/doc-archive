package models

import (
	"time"
)

type DocumentDB struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Authors   *string   `json:"authors,omitempty"`
	Year      *int      `json:"year,omitempty"`
	Category  *string   `json:"category,omitempty"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

func (DocumentDB) TableName() string {
	return "documents"
}

type ChunkDB struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Content    string    `json:"content"`
	Embedding  []float32 `json:"-"`
	Index      int       `json:"index"`
}

func (ChunkDB) TableName() string {
	return "chunks"
}
