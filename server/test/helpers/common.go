package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// TestUser creates a test user with predictable data
func TestUser(t *testing.T, email, password string) *domain.User {
	t.Helper()
	user, err := domain.NewUser(email, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	// Use a predictable ID for testing if not set
	if user.ID == "" {
		user.ID = "test-user-id"
	}
	return user
}

// TestDocument creates a test document with predictable data
func TestDocument(t *testing.T, userID, title string) *domain.Document {
	t.Helper()
	return &domain.Document{
		ID:        "test-doc-id",
		UserID:    userID,
		Title:     title,
		FilePath:  "/test/path/document.pdf",
		FileSize:  1024,
		Status:    domain.DocumentStatusPending,
		CreatedAt: time.Now(),
	}
}

// TestChunk creates a test chunk with predictable data
func TestChunk(t *testing.T, documentID, content string, index int) *domain.Chunk {
	t.Helper()
	return &domain.Chunk{
		ID:         "test-chunk-id",
		DocumentID: documentID,
		Content:    content,
		Index:      index,
	}
}

// ContextWithTimeout returns a context with a test timeout
func ContextWithTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

// AssertEqual fails the test if values are not equal
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// AssertNotEqual fails the test if values are equal
func AssertNotEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got == want {
		t.Fatalf("got %v, which equals %v", got, want)
	}
}
