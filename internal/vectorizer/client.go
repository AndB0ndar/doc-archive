package vectorizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	URL    string
	httpClient *http.Client
}

func NewClient(URL string) *Client {
	return &Client{
		URL: URL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Embed(text string) ([]float32, error) {
	reqBody, err := json.Marshal(map[string][]string{
		"texts": {text},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.URL+"/embed", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedder returned status %d: %s", resp.StatusCode, string(body))
	}

	var embeddings [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}
