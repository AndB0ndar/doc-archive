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
	CreatedAt time.Time `json:"created_at"`
	UserID    int       `json:"user_id"`
}
