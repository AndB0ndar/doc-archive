package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

type client struct {
	URL        string
	httpClient *http.Client
}

func New(EmbedderURL string) domain.EmbedderClient {
	return &client{
		URL: EmbedderURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type embedderResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (c *client) Embed(text string) (*domain.EmbedResult, error) {
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

	var response embedderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(response.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return &domain.EmbedResult{Embeddings: response.Embeddings[0]}, nil
}
