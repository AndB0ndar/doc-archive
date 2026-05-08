package domain_test

import (
	"testing"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

func TestDocument_ValidStatusTransitions(t *testing.T) {
	tests := []struct {
		name  string
		from  domain.DocumentStatus
		to    domain.DocumentStatus
		valid bool
	}{
		// Valid transitions
		{
			name:  "pending to processing",
			from:  domain.DocumentStatusPending,
			to:    domain.DocumentStatusProcessing,
			valid: true,
		},
		{
			name:  "pending to error",
			from:  domain.DocumentStatusPending,
			to:    domain.DocumentStatusError,
			valid: true,
		},
		{
			name:  "processing to done",
			from:  domain.DocumentStatusProcessing,
			to:    domain.DocumentStatusDone,
			valid: true,
		},
		{
			name:  "processing to error",
			from:  domain.DocumentStatusProcessing,
			to:    domain.DocumentStatusError,
			valid: true,
		},
		// Invalid transitions (some examples)
		{
			name:  "done to processing",
			from:  domain.DocumentStatusDone,
			to:    domain.DocumentStatusProcessing,
			valid: false,
		},
		{
			name:  "error to done",
			from:  domain.DocumentStatusError,
			to:    domain.DocumentStatusDone,
			valid: false,
		},
		// Same status is valid
		{
			name:  "same status",
			from:  domain.DocumentStatusPending,
			to:    domain.DocumentStatusPending,
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &domain.Document{Status: tt.from}
			got := doc.CanTransitionTo(tt.to)
			if got != tt.valid {
				t.Errorf("CanTransitionTo(%s) from %s = %v, want %v", tt.to, tt.from, got, tt.valid)
			}
		})
	}
}

func TestDocument_StructInitialization(t *testing.T) {
	doc := &domain.Document{
		ID:        "doc-123",
		UserID:    "user-456",
		Title:     "Test Document",
		FilePath:  "/uploads/test.pdf",
		FileSize:  1024,
		Status:    domain.DocumentStatusPending,
		CreatedAt: time.Now(),
	}

	if doc.ID != "doc-123" {
		t.Errorf("ID = %q, want %q", doc.ID, "doc-123")
	}
	if doc.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", doc.UserID, "user-456")
	}
	if doc.Title != "Test Document" {
		t.Errorf("Title = %q, want %q", doc.Title, "Test Document")
	}
	if doc.FilePath != "/uploads/test.pdf" {
		t.Errorf("FilePath = %q, want %q", doc.FilePath, "/uploads/test.pdf")
	}
	if doc.FileSize != 1024 {
		t.Errorf("FileSize = %d, want %d", doc.FileSize, 1024)
	}
	if doc.Status != domain.DocumentStatusPending {
		t.Errorf("Status = %q, want %q", doc.Status, domain.DocumentStatusPending)
	}
	if doc.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestDocument_OptionalFields(t *testing.T) {
	// Test that optional fields can be nil
	doc := &domain.Document{
		ID:       "doc-123",
		UserID:   "user-456",
		Title:    "Test Document",
		FilePath: "/uploads/test.pdf",
		FileSize: 1024,
		Status:   domain.DocumentStatusPending,
		// Authors, Year, Category are nil
	}

	if doc.Authors != nil {
		t.Error("Authors should be nil")
	}
	if doc.Year != nil {
		t.Error("Year should be nil")
	}
	if doc.Category != nil {
		t.Error("Category should be nil")
	}

	// Test with optional fields set
	author := "John Doe"
	year := 2023
	category := "Research"

	doc2 := &domain.Document{
		ID:       "doc-456",
		UserID:   "user-789",
		Title:    "Another Document",
		Authors:  &author,
		Year:     &year,
		Category: &category,
		FilePath: "/uploads/another.pdf",
		FileSize: 2048,
		Status:   domain.DocumentStatusDone,
	}

	if doc2.Authors == nil || *doc2.Authors != author {
		t.Errorf("Authors = %v, want %q", doc2.Authors, author)
	}
	if doc2.Year == nil || *doc2.Year != year {
		t.Errorf("Year = %v, want %d", doc2.Year, year)
	}
	if doc2.Category == nil || *doc2.Category != category {
		t.Errorf("Category = %v, want %q", doc2.Category, category)
	}
}

func TestDocumentStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.DocumentStatus
		expected string
	}{
		{
			name:     "Pending",
			status:   domain.DocumentStatusPending,
			expected: "pending",
		},
		{
			name:     "Processing",
			status:   domain.DocumentStatusProcessing,
			expected: "processing",
		},
		{
			name:     "Done",
			status:   domain.DocumentStatusDone,
			expected: "done",
		},
		{
			name:     "Error",
			status:   domain.DocumentStatusError,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("status = %q, want %q", tt.status, tt.expected)
			}
		})
	}
}

func TestDocument_DefaultValues(t *testing.T) {
	// Test that a document can be created with minimal fields
	doc := &domain.Document{
		ID:       "doc-123",
		UserID:   "user-456",
		Title:    "Test Document",
		FilePath: "/uploads/test.pdf",
		Status:   domain.DocumentStatusPending,
	}

	// FileSize should be zero if not set
	if doc.FileSize != 0 {
		t.Errorf("FileSize = %d, want 0", doc.FileSize)
	}

	// CreatedAt should be zero if not set
	if !doc.CreatedAt.IsZero() {
		t.Error("CreatedAt should be zero when not set")
	}
}
