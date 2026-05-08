package domain_test

import (
	"testing"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

func TestChunk_StructInitialization(t *testing.T) {
	chunk := &domain.Chunk{
		ID:         "chunk-123",
		DocumentID: "doc-456",
		Content:    "This is a test chunk of text.",
		Embedding:  []float32{0.1, 0.2, 0.3},
		Index:      0,
	}

	if chunk.ID != "chunk-123" {
		t.Errorf("ID = %q, want %q", chunk.ID, "chunk-123")
	}
	if chunk.DocumentID != "doc-456" {
		t.Errorf("DocumentID = %q, want %q", chunk.DocumentID, "doc-456")
	}
	if chunk.Content != "This is a test chunk of text." {
		t.Errorf("Content = %q, want %q", chunk.Content, "This is a test chunk of text.")
	}
	if len(chunk.Embedding) != 3 {
		t.Errorf("Embedding length = %d, want 3", len(chunk.Embedding))
	}
	if chunk.Embedding[0] != 0.1 || chunk.Embedding[1] != 0.2 || chunk.Embedding[2] != 0.3 {
		t.Errorf("Embedding = %v, want [0.1, 0.2, 0.3]", chunk.Embedding)
	}
	if chunk.Index != 0 {
		t.Errorf("Index = %d, want 0", chunk.Index)
	}
}

func TestChunk_WithoutEmbedding(t *testing.T) {
	// Chunks can exist without embeddings initially
	chunk := &domain.Chunk{
		ID:         "chunk-123",
		DocumentID: "doc-456",
		Content:    "This is a test chunk.",
		Index:      1,
		// Embedding is nil
	}

	if chunk.Embedding != nil {
		t.Error("Embedding should be nil")
	}
	if chunk.Content == "" {
		t.Error("Content should not be empty")
	}
}

func TestChunk_EmptyContent(t *testing.T) {
	// Chunks with empty content might be valid in some cases
	chunk := &domain.Chunk{
		ID:         "chunk-123",
		DocumentID: "doc-456",
		Content:    "", // Empty content
		Index:      2,
	}

	if chunk.Content != "" {
		t.Errorf("Content = %q, want empty string", chunk.Content)
	}
}

func TestChunkSearchResult_Struct(t *testing.T) {
	chunk := &domain.Chunk{
		ID:         "chunk-123",
		DocumentID: "doc-456",
		Content:    "Test content",
		Index:      0,
	}

	result := &domain.ChunkSearchResult{
		Chunk:      *chunk,
		Similarity: 0.85,
	}

	if result.ID != chunk.ID {
		t.Errorf("result.ID = %q, want %q", result.ID, chunk.ID)
	}
	if result.DocumentID != chunk.DocumentID {
		t.Errorf("result.DocumentID = %q, want %q", result.DocumentID, chunk.DocumentID)
	}
	if result.Content != chunk.Content {
		t.Errorf("result.Content = %q, want %q", result.Content, chunk.Content)
	}
	if result.Index != chunk.Index {
		t.Errorf("result.Index = %d, want %d", result.Index, chunk.Index)
	}
	if result.Similarity != 0.85 {
		t.Errorf("result.Similarity = %f, want 0.85", result.Similarity)
	}
}

func TestChunkSearchResult_Embedding(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3, 0.4}

	chunk := &domain.Chunk{
		ID:         "chunk-123",
		DocumentID: "doc-456",
		Content:    "Test content",
		Embedding:  embedding,
		Index:      0,
	}

	result := &domain.ChunkSearchResult{
		Chunk:      *chunk,
		Similarity: 0.92,
	}

	// Check that embedding is preserved
	if len(result.Embedding) != len(embedding) {
		t.Errorf("result.Embedding length = %d, want %d", len(result.Embedding), len(embedding))
	}
	for i := range embedding {
		if result.Embedding[i] != embedding[i] {
			t.Errorf("result.Embedding[%d] = %f, want %f", i, result.Embedding[i], embedding[i])
		}
	}
}

func TestChunk_IndexValidation(t *testing.T) {
	tests := []struct {
		name  string
		index int
		valid bool
	}{
		{
			name:  "zero index",
			index: 0,
			valid: true,
		},
		{
			name:  "positive index",
			index: 5,
			valid: true,
		},
		{
			name:  "negative index",
			index: -1,
			valid: false, // Negative indices are usually invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &domain.Chunk{
				ID:         "chunk-123",
				DocumentID: "doc-456",
				Content:    "Test",
				Index:      tt.index,
			}

			if tt.valid && chunk.Index != tt.index {
				t.Errorf("Index = %d, want %d", chunk.Index, tt.index)
			}
			if !tt.valid {
				t.Logf("Index %d is typically invalid", tt.index)
			}
		})
	}
}

func TestChunk_MultipleChunksSameDocument(t *testing.T) {
	// Test that multiple chunks can belong to the same document
	docID := "doc-123"

	chunks := []*domain.Chunk{
		{
			ID:         "chunk-1",
			DocumentID: docID,
			Content:    "First chunk",
			Index:      0,
		},
		{
			ID:         "chunk-2",
			DocumentID: docID,
			Content:    "Second chunk",
			Index:      1,
		},
		{
			ID:         "chunk-3",
			DocumentID: docID,
			Content:    "Third chunk",
			Index:      2,
		},
	}

	for i, chunk := range chunks {
		if chunk.DocumentID != docID {
			t.Errorf("chunk[%d].DocumentID = %q, want %q", i, chunk.DocumentID, docID)
		}
		if chunk.Index != i {
			t.Errorf("chunk[%d].Index = %d, want %d", i, chunk.Index, i)
		}
	}
}
