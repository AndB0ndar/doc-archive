package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/config"
)

type Reranker struct {
	URL        string
	httpClient *http.Client
}

func NewReranker(cfg *config.Config) *Reranker {
	return &Reranker{
		URL: cfg.RerankerURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Reranker) Rerank(query string, texts []string) ([]float32, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"query": query,
		"texts": texts,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.URL+"/rerank", "application/json", bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"reranker returned status %d: %s", resp.StatusCode, string(body),
		)
	}

	var response struct {
		Scores []float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return response.Scores, nil
}
