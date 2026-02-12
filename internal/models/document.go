package models

import (
	"time"
)

type Document struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Authors   *string   `json:"authors,omitempty"`
	Year      *int      `json:"year,omitempty"`
	Category  *string   `json:"category,omitempty"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	FullText  *string   `json:"full_text,omitempty"`
	Embedding []float32 `json:"-"` // not view in JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Document) TableName() string {
	return "documents"
}
