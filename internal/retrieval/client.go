package retrieval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}

func (c *Client) Search(ctx context.Context, query Query) (SearchResponse, error) {
	if c == nil || strings.TrimSpace(c.baseURL) == "" {
		return SearchResponse{}, fmt.Errorf("retrieval base url is required")
	}
	payload, err := json.Marshal(query)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("marshal retrieval query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/search", bytes.NewReader(payload))
	if err != nil {
		return SearchResponse{}, fmt.Errorf("build retrieval request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("execute retrieval request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return SearchResponse{}, fmt.Errorf("retrieval status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return SearchResponse{}, fmt.Errorf("decode retrieval response: %w", err)
	}
	return decoded, nil
}
