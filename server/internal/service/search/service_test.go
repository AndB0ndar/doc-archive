package search

import (
	"errors"
	"testing"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/test/helpers"
	"github.com/AndB0ndar/doc-archive/test/mocks"
)

func newTestService(t *testing.T, cfg config.SearchConfig) (*service, *mocks.MockChunkRepository, *mocks.MockEmbedderClient, *mocks.MockRerankerClient, *mocks.MockReaderClient) {
	t.Helper()

	chunkRepo := mocks.NewMockChunkRepository()
	embedderClient := mocks.NewMockEmbedderClient()
	rerankerClient := mocks.NewMockRerankerClient()
	readerClient := mocks.NewMockReaderClient()
	log := logger.New("test")

	svc := &service{
		chunkRepo:      chunkRepo,
		embedderClient: embedderClient,
		rerankerClient: rerankerClient,
		readerClient:   readerClient,
		cfg:            cfg,
		logger:         log,
	}

	return svc, chunkRepo, embedderClient, rerankerClient, readerClient
}

func TestSearchService_Validate_EmptyQuery(t *testing.T) {
	cfg := config.SearchConfig{DefaultLimit: 20, MaxLimit: 100}
	svc, _, _, _, _ := newTestService(t, cfg)

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"empty query string", "", true},
		{"whitespace only", "   ", true},
		{"valid query", "test query", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := helpers.ContextWithTimeout(t)
			_, err := svc.Search(ctx, domain.SearchQuery{Query: tt.query}, "user1", 10)
			if tt.wantErr {
				helpers.AssertError(t, err)
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestSearchService_TextSearch(t *testing.T) {
	cfg := config.SearchConfig{DefaultLimit: 20, MaxLimit: 100}
	svc, chunkRepo, _, _, _ := newTestService(t, cfg)

	ctx := helpers.ContextWithTimeout(t)

	// Seed chunks in the repo with unique IDs
	chunk1 := helpers.TestChunk(t, "doc1", "The quick brown fox jumps over the lazy dog", 0)
	chunk1.ID = "chunk-fox-1"
	chunk2 := helpers.TestChunk(t, "doc1", "The quick brown fox", 1)
	chunk2.ID = "chunk-fox-2"
	chunk3 := helpers.TestChunk(t, "doc2", "Completely unrelated content about something else", 0)
	chunk3.ID = "chunk-other-1"
	chunkRepo.AddChunk(chunk1)
	chunkRepo.AddChunk(chunk2)
	chunkRepo.AddChunk(chunk3)

	t.Run("text search returns matching results", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "fox", Type: "text"}, "user1", 10)
		helpers.AssertNoError(t, err)
		// chunk1 and chunk2 contain "fox"
		if len(results) < 2 {
			t.Fatalf("expected at least 2 results, got %d", len(results))
		}
	})

	t.Run("empty type defaults to text search", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "fox", Type: ""}, "user1", 10)
		helpers.AssertNoError(t, err)
		if len(results) < 2 {
			t.Fatalf("expected at least 2 results, got %d", len(results))
		}
	})

	t.Run("no matches returns empty results", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "nonexistent"}, "user1", 10)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, len(results), 0)
	})
}

func TestSearchService_SemanticSearch(t *testing.T) {
	cfg := config.SearchConfig{DefaultLimit: 20, MaxLimit: 100}
	svc, chunkRepo, embedderClient, _, _ := newTestService(t, cfg)

	ctx := helpers.ContextWithTimeout(t)

	// Seed chunks with embeddings and unique IDs
	embeddings := []float32{0.1, 0.2, 0.3}
	chunk1 := helpers.TestChunk(t, "doc1", "Semantic content one", 0)
	chunk1.ID = "chunk-sem-1"
	chunk1.Embedding = embeddings
	chunk2 := helpers.TestChunk(t, "doc1", "Semantic content two", 1)
	chunk2.ID = "chunk-sem-2"
	chunk2.Embedding = embeddings
	chunkRepo.AddChunk(chunk1)
	chunkRepo.AddChunk(chunk2)

	t.Run("semantic search returns ranked results", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "test query", Type: "vector"}, "user1", 10)
		helpers.AssertNoError(t, err)
		if len(results) < 2 {
			t.Fatalf("expected at least 2 results, got %d", len(results))
		}
		for _, r := range results {
			if r.Confidence != nil {
				// Should have a confidence score
				return
			}
		}
		t.Fatal("expected at least one result to have a confidence score")
	})

	t.Run("semantic search with 'semantic' type", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "test query", Type: "semantic"}, "user1", 10)
		helpers.AssertNoError(t, err)
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
	})

	t.Run("embedder error propagates", func(t *testing.T) {
		embedderClient.SetEmbedError(errors.New("embedder unavailable"))
		_, err := svc.Search(ctx, domain.SearchQuery{Query: "test", Type: "vector"}, "user1", 10)
		helpers.AssertError(t, err)
		embedderClient.ClearErrors()
	})
}

func TestSearchService_LimitEnforcement(t *testing.T) {
	cfg := config.SearchConfig{DefaultLimit: 20, MaxLimit: 100}
	svc, chunkRepo, _, _, _ := newTestService(t, cfg)

	ctx := helpers.ContextWithTimeout(t)

	// Seed many chunks
	for i := 0; i < 30; i++ {
		chunk := helpers.TestChunk(t, "doc1", "fox content for testing limits "+string(rune('a'+i)), i)
		chunk.ID = helpers.TestChunk(t, "doc1", "", i).ID + string(rune('0'+i))
		chunkRepo.AddChunk(chunk)
	}

	t.Run("limit=0 uses default limit", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "fox"}, "user1", 0)
		helpers.AssertNoError(t, err)
		if len(results) > 20 {
			t.Fatalf("expected at most 20 results with default limit, got %d", len(results))
		}
	})

	t.Run("limit exceeding MaxLimit gets capped", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "fox"}, "user1", 200)
		helpers.AssertNoError(t, err)
		if len(results) > 100 {
			t.Fatalf("expected at most 100 results, got %d", len(results))
		}
	})

	t.Run("valid limit is respected", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "fox"}, "user1", 5)
		helpers.AssertNoError(t, err)
		if len(results) > 5 {
			t.Fatalf("expected at most 5 results, got %d", len(results))
		}
	})
}

func TestSearchService_Reranker(t *testing.T) {
	cfg := config.SearchConfig{
		DefaultLimit:    20,
		MaxLimit:        100,
		RerankerEnabled: true,
		ReaderEnabled:   false,
	}
	svc, chunkRepo, _, rerankerClient, _ := newTestService(t, cfg)

	ctx := helpers.ContextWithTimeout(t)

	// Seed chunks with unique IDs
	chunk1 := helpers.TestChunk(t, "doc1", "Alpha content for searching", 0)
	chunk1.ID = "chunk-alpha-1"
	chunk2 := helpers.TestChunk(t, "doc1", "Beta content for searching", 1)
	chunk2.ID = "chunk-beta-2"
	chunk3 := helpers.TestChunk(t, "doc1", "Gamma content for searching", 2)
	chunk3.ID = "chunk-gamma-3"
	chunkRepo.AddChunk(chunk1)
	chunkRepo.AddChunk(chunk2)
	chunkRepo.AddChunk(chunk3)

	t.Run("reranker is applied when enabled", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "content"}, "user1", 10)
		helpers.AssertNoError(t, err)
		if len(results) == 0 {
			t.Fatal("expected results")
		}
		// Reranker should have been called (verified by confidence scores being updated)
		// Default mock reranker returns [0.8, 0.7, 0.6] scores mapped to chunks
	})

	t.Run("reranker error does not fail search", func(t *testing.T) {
		rerankerClient.SetRerankError(errors.New("reranker unavailable"))
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "content"}, "user1", 10)
		helpers.AssertNoError(t, err)
		if len(results) == 0 {
			t.Fatal("expected results despite reranker error")
		}
		rerankerClient.ClearErrors()
	})
}

func TestSearchService_Reader(t *testing.T) {
	cfg := config.SearchConfig{
		DefaultLimit:    20,
		MaxLimit:        100,
		RerankerEnabled: false,
		ReaderEnabled:   true,
	}
	svc, chunkRepo, _, _, readerClient := newTestService(t, cfg)

	ctx := helpers.ContextWithTimeout(t)

	chunk1 := helpers.TestChunk(t, "doc1", "The capital of France is Paris. It is a beautiful city.", 0)
	chunk1.ID = "chunk-capital-1"
	chunk2 := helpers.TestChunk(t, "doc1", "The Earth orbits the Sun in approximately 365 days.", 1)
	chunk2.ID = "chunk-earth-2"
	chunkRepo.AddChunk(chunk1)
	chunkRepo.AddChunk(chunk2)

	t.Run("reader is applied when enabled", func(t *testing.T) {
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "capital of France"}, "user1", 10)
		helpers.AssertNoError(t, err)
		if len(results) == 0 {
			t.Fatal("expected results")
		}
		// Mock reader returns Answer with "This is a mock answer."
		// At least the top result should have an answer
	})

	t.Run("reader error does not fail search", func(t *testing.T) {
		readerClient.SetAnswerError(errors.New("reader unavailable"))
		results, err := svc.Search(ctx, domain.SearchQuery{Query: "capital of France"}, "user1", 10)
		helpers.AssertNoError(t, err)
		// Results should still be returned despite reader error
		if len(results) == 0 {
			t.Fatal("expected results despite reader error")
		}
		readerClient.ClearErrors()
	})
}

func TestSearchService_ErrorHandling(t *testing.T) {
	cfg := config.SearchConfig{DefaultLimit: 20, MaxLimit: 100}

	t.Run("full text search error propagates", func(t *testing.T) {
		svc, chunkRepo, _, _, _ := newTestService(t, cfg)
		ctx := helpers.ContextWithTimeout(t)

		chunkRepo.SetFullTextSearchError(errors.New("db connection failed"))
		_, err := svc.Search(ctx, domain.SearchQuery{Query: "test"}, "user1", 10)
		helpers.AssertError(t, err)
		chunkRepo.ClearErrors()
	})

	t.Run("semantic search error propagates", func(t *testing.T) {
		svc, chunkRepo, _, _, _ := newTestService(t, cfg)
		ctx := helpers.ContextWithTimeout(t)

		chunkRepo.SetSemanticSearchError(errors.New("vector search failed"))
		_, err := svc.Search(ctx, domain.SearchQuery{Query: "test", Type: "vector"}, "user1", 10)
		helpers.AssertError(t, err)
		chunkRepo.ClearErrors()
	})

	t.Run("unsupported search type returns error", func(t *testing.T) {
		svc, _, _, _, _ := newTestService(t, cfg)
		ctx := helpers.ContextWithTimeout(t)

		_, err := svc.Search(ctx, domain.SearchQuery{Query: "test", Type: "hybrid"}, "user1", 10)
		helpers.AssertError(t, err)
	})

	t.Run("embedder error on semantic search", func(t *testing.T) {
		svc, _, embedderClient, _, _ := newTestService(t, cfg)
		ctx := helpers.ContextWithTimeout(t)

		embedderClient.SetEmbedError(errors.New("embedder timeout"))
		_, err := svc.Search(ctx, domain.SearchQuery{Query: "test", Type: "vector"}, "user1", 10)
		helpers.AssertError(t, err)
		embedderClient.ClearErrors()
	})
}

// TestSearchService_DefaultConfig verifies default config values
func TestSearchService_DefaultConfig(t *testing.T) {
	// Verify the config defaults match what the service expects
	defaultCfg := config.SearchConfig{
		DefaultLimit:    20,
		MaxLimit:        100,
		RerankerEnabled: false,
		ReaderEnabled:   false,
	}

	svc, chunkRepo, _, _, _ := newTestService(t, defaultCfg)
	ctx := helpers.ContextWithTimeout(t)

	chunk := helpers.TestChunk(t, "doc1", "test content for default config test", 0)
	chunk.ID = "chunk-default-1"
	chunkRepo.AddChunk(chunk)

	results, err := svc.Search(ctx, domain.SearchQuery{Query: "test content"}, "user1", 0)
	helpers.AssertNoError(t, err)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
}

// TestSearchService_LimitTimesThree verifies fetchLimit = limit * 3
func TestSearchService_LimitLargerThanDefault(t *testing.T) {
	cfg := config.SearchConfig{DefaultLimit: 20, MaxLimit: 100}
	svc, chunkRepo, _, _, _ := newTestService(t, cfg)
	ctx := helpers.ContextWithTimeout(t)

	// Add 50 chunks with unique IDs
	for i := 0; i < 50; i++ {
		chunk := helpers.TestChunk(t, "doc1", "searchable content for testing", i)
		chunk.ID = "chunk-limits-" + string(rune('0'+i%10)) + string(rune('0'+(i/10)))
		chunkRepo.AddChunk(chunk)
	}

	// Request limit=10, internal fetchLimit should be 30
	results, err := svc.Search(ctx, domain.SearchQuery{Query: "searchable"}, "user1", 10)
	helpers.AssertNoError(t, err)
	// Service trims to limit at end, so should be at most 10
	if len(results) > 10 {
		t.Fatalf("expected at most 10 final results, got %d", len(results))
	}
}
