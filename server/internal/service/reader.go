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

type Reader struct {
	URL        string
	httpClient *http.Client
}

func NewReader(cfg *config.Config) *Reader {
	return &Reader{
		URL: cfg.ReaderURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type AnswerResponse struct {
	Answer     string  `json:"answer"`
	Confidence float64 `json:"confidence"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
}

func (c *Reader) ExtractAnswer(
	question, context string,
) (*AnswerResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"question": question,
		"context":  context,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.URL+"/extract_answer", "application/json", bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reader returned status %d: %s", resp.StatusCode, string(body))
	}

	var ans AnswerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ans); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &ans, nil
}
