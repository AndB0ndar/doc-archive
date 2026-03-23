package models

import "time"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type DocumentResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Authors   *string   `json:"authors"`
	Year      *int      `json:"year"`
	Category  *string   `json:"category"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
}

type UploadResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
}

type SearchResultItem struct {
	ChunkID    string   `json:"chunk_id"`
	DocumentID string   `json:"document_id"`
	Content    string   `json:"content"`
	Answer     *string  `json:"answer,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
}

type SearchResponse struct {
	Results []SearchResultItem `json:"results"`
}
