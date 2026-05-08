package fixtures

import (
	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// SampleUsers returns a slice of sample users for testing
func SampleUsers() []*domain.User {
	return []*domain.User{
		{
			ID:           "user-1",
			Email:        "test1@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuv", // Mock hash
		},
		{
			ID:           "user-2",
			Email:        "test2@example.com",
			PasswordHash: "$2a$10$zyxwvutsrqponmlkjihgfed", // Mock hash
		},
	}
}

// SampleDocuments returns a slice of sample documents for testing
func SampleDocuments() []*domain.Document {
	return []*domain.Document{
		{
			ID:       "doc-1",
			UserID:   "user-1",
			Title:    "Research Paper on Machine Learning",
			Authors:  stringPtr("John Doe, Jane Smith"),
			Year:     intPtr(2023),
			Category: stringPtr("Research"),
			FilePath: "/uploads/doc1.pdf",
			FileSize: 1024 * 1024, // 1MB
			Status:   domain.DocumentStatusDone,
		},
		{
			ID:       "doc-2",
			UserID:   "user-1",
			Title:    "Meeting Notes Q4 2023",
			Authors:  nil,
			Year:     intPtr(2023),
			Category: stringPtr("Business"),
			FilePath: "/uploads/doc2.pdf",
			FileSize: 512 * 1024, // 512KB
			Status:   domain.DocumentStatusProcessing,
		},
		{
			ID:       "doc-3",
			UserID:   "user-2",
			Title:    "Technical Documentation",
			Authors:  stringPtr("Alice Johnson"),
			Year:     intPtr(2024),
			Category: stringPtr("Documentation"),
			FilePath: "/uploads/doc3.pdf",
			FileSize: 2 * 1024 * 1024, // 2MB
			Status:   domain.DocumentStatusDone,
		},
	}
}

// SampleChunks returns a slice of sample chunks for testing
func SampleChunks() []*domain.Chunk {
	return []*domain.Chunk{
		{
			ID:         "chunk-1",
			DocumentID: "doc-1",
			Content:    "Machine learning is a subset of artificial intelligence that enables systems to learn and improve from experience without being explicitly programmed.",
			Index:      0,
		},
		{
			ID:         "chunk-2",
			DocumentID: "doc-1",
			Content:    "There are three main types of machine learning: supervised learning, unsupervised learning, and reinforcement learning.",
			Index:      1,
		},
		{
			ID:         "chunk-3",
			DocumentID: "doc-1",
			Content:    "Supervised learning involves training a model on labeled data, where the correct output is provided for each input.",
			Index:      2,
		},
		{
			ID:         "chunk-4",
			DocumentID: "doc-3",
			Content:    "This documentation covers the installation and usage of the software package.",
			Index:      0,
		},
		{
			ID:         "chunk-5",
			DocumentID: "doc-3",
			Content:    "To install the package, run the following command: `npm install package-name`.",
			Index:      1,
		},
	}
}

// SampleEmbedding returns a sample embedding vector for testing
func SampleEmbedding() []float32 {
	// Return a simple 384-dimensional vector (common embedding size)
	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = 0.1 // Simple constant value for testing
	}
	return embedding
}

// Helper functions for pointer types
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
