package reranker

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

func New(RerankerURL string) domain.RerankerClient {
	return &client{
		URL: RerankerURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type rerankResponse struct {
	Scores []float32 `json:"embeddings"`
}

func (c *client) Rerank(
	query string, passages []string,
) (*domain.RerankResult, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"query": query,
		"texts": passages,
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

	var response rerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &domain.RerankResult{Scores: response.Scores}, nil
}
