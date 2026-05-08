package mocks

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// MockChunkRepository implements domain.ChunkRepository for testing
type MockChunkRepository struct {
	mu                 sync.RWMutex
	chunks             map[string]*domain.Chunk // key: chunk ID
	chunksByDocument   map[string][]string      // key: document ID, value: list of chunk IDs
	nextID             int
	createErr          error
	createBatchErr     error
	updateEmbeddingErr error
	fullTextSearchErr  error
	semanticSearchErr  error
}

// NewMockChunkRepository creates a new mock chunk repository
func NewMockChunkRepository() *MockChunkRepository {
	return &MockChunkRepository{
		chunks:           make(map[string]*domain.Chunk),
		chunksByDocument: make(map[string][]string),
		nextID:           1,
	}
}

// Create implements domain.ChunkRepository.Create
func (m *MockChunkRepository) Create(ctx context.Context, chunk *domain.Chunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return m.createErr
	}

	// Generate a mock ID if not set
	if chunk.ID == "" {
		chunk.ID = fmt.Sprintf("mock-chunk-%d", m.nextID)
		m.nextID++
	}

	// Store the chunk
	m.chunks[chunk.ID] = chunk
	m.chunksByDocument[chunk.DocumentID] = append(m.chunksByDocument[chunk.DocumentID], chunk.ID)

	return nil
}

// CreateBatch implements domain.ChunkRepository.CreateBatch
func (m *MockChunkRepository) CreateBatch(ctx context.Context, chunks []*domain.Chunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createBatchErr != nil {
		return m.createBatchErr
	}

	for _, chunk := range chunks {
		// Generate a mock ID if not set
		if chunk.ID == "" {
			chunk.ID = fmt.Sprintf("mock-chunk-%d", m.nextID)
			m.nextID++
		}

		// Store the chunk
		m.chunks[chunk.ID] = chunk
		m.chunksByDocument[chunk.DocumentID] = append(m.chunksByDocument[chunk.DocumentID], chunk.ID)
	}

	return nil
}

// UpdateEmbedding implements domain.ChunkRepository.UpdateEmbedding
func (m *MockChunkRepository) UpdateEmbedding(ctx context.Context, chunkID string, embedding []float32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateEmbeddingErr != nil {
		return m.updateEmbeddingErr
	}

	chunk, exists := m.chunks[chunkID]
	if !exists {
		return fmt.Errorf("chunk not found")
	}

	chunk.Embedding = embedding
	return nil
}

// FullTextSearch implements domain.ChunkRepository.FullTextSearch
func (m *MockChunkRepository) FullTextSearch(ctx context.Context, query string, userID string, limit int) ([]*domain.ChunkSearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.fullTextSearchErr != nil {
		return nil, m.fullTextSearchErr
	}

	// Simple mock implementation: find chunks containing the query
	var results []*domain.ChunkSearchResult
	queryLower := strings.ToLower(query)

	for _, chunk := range m.chunks {
		if strings.Contains(strings.ToLower(chunk.Content), queryLower) {
			// Simple similarity calculation based on string contains
			similarity := 0.5
			if strings.Contains(chunk.Content, query) {
				similarity = 0.8
			}

			results = append(results, &domain.ChunkSearchResult{
				Chunk: domain.Chunk{
					ID:         chunk.ID,
					DocumentID: chunk.DocumentID,
					Content:    chunk.Content,
					Embedding:  chunk.Embedding,
					Index:      chunk.Index,
				},
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SemanticSearch implements domain.ChunkRepository.SemanticSearch
func (m *MockChunkRepository) SemanticSearch(ctx context.Context, embedding []float32, userID string, limit int) ([]*domain.ChunkSearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.semanticSearchErr != nil {
		return nil, m.semanticSearchErr
	}

	// Mock implementation: return chunks with embeddings
	var results []*domain.ChunkSearchResult

	for _, chunk := range m.chunks {
		if len(chunk.Embedding) > 0 {
			// Mock similarity calculation
			similarity := 0.7 // Fixed mock value

			results = append(results, &domain.ChunkSearchResult{
				Chunk: domain.Chunk{
					ID:         chunk.ID,
					DocumentID: chunk.DocumentID,
					Content:    chunk.Content,
					Embedding:  chunk.Embedding,
					Index:      chunk.Index,
				},
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SetCreateError sets an error to be returned by Create
func (m *MockChunkRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createErr = err
}

// SetCreateBatchError sets an error to be returned by CreateBatch
func (m *MockChunkRepository) SetCreateBatchError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createBatchErr = err
}

// SetUpdateEmbeddingError sets an error to be returned by UpdateEmbedding
func (m *MockChunkRepository) SetUpdateEmbeddingError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateEmbeddingErr = err
}

// SetFullTextSearchError sets an error to be returned by FullTextSearch
func (m *MockChunkRepository) SetFullTextSearchError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fullTextSearchErr = err
}

// SetSemanticSearchError sets an error to be returned by SemanticSearch
func (m *MockChunkRepository) SetSemanticSearchError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.semanticSearchErr = err
}

// ClearErrors clears all configured errors
func (m *MockChunkRepository) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createErr = nil
	m.createBatchErr = nil
	m.updateEmbeddingErr = nil
	m.fullTextSearchErr = nil
	m.semanticSearchErr = nil
}

// GetChunk returns a chunk by ID (test helper)
func (m *MockChunkRepository) GetChunk(id string) *domain.Chunk {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.chunks[id]
}

// AddChunk adds a chunk directly (test helper)
func (m *MockChunkRepository) AddChunk(chunk *domain.Chunk) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chunks[chunk.ID] = chunk
	m.chunksByDocument[chunk.DocumentID] = append(m.chunksByDocument[chunk.DocumentID], chunk.ID)
}

// GetChunksByDocument returns chunks for a document (test helper)
func (m *MockChunkRepository) GetChunksByDocument(documentID string) []*domain.Chunk {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chunkIDs, exists := m.chunksByDocument[documentID]
	if !exists {
		return nil
	}

	chunks := make([]*domain.Chunk, 0, len(chunkIDs))
	for _, id := range chunkIDs {
		if chunk, ok := m.chunks[id]; ok {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// Clear clears all chunks (test helper)
func (m *MockChunkRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chunks = make(map[string]*domain.Chunk)
	m.chunksByDocument = make(map[string][]string)
	m.nextID = 1
	m.createErr = nil
	m.createBatchErr = nil
	m.updateEmbeddingErr = nil
	m.fullTextSearchErr = nil
	m.semanticSearchErr = nil
}
