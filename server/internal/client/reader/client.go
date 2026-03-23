package reader

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

func New(ReaderURL string) domain.ReaderClient {
	return &client{
		URL: ReaderURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type answerResponse struct {
	Answer     string  `json:"answer"`
	Confidence float64 `json:"confidence"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
}

func (c *client) Answer(
	question string, context string,
) (*domain.Answer, error) {
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

	var response answerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &domain.Answer{
		Answer:     response.Answer,
		Confidence: response.Confidence,
		Start:      response.Start,
		End:        response.End,
	}, nil
}
