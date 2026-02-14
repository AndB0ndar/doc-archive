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

type Embedder struct {
	URL        string
	httpClient *http.Client
}

func NewEmbedder(cfg *config.Config) *Embedder {
	return &Embedder{
		URL: cfg.EmbedderURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Embedder) Embed(text string) ([]float32, error) {
	reqBody, err := json.Marshal(map[string][]string{
		"texts": {text},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.URL+"/embed", "application/json", bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"embedder returned status %d: %s", resp.StatusCode, string(body),
		)
	}

	var response struct {
        Embeddings [][]float32 `json:"embeddings"`
    }
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(response.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return response.Embeddings[0], nil
}
