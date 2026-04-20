package retrievalgateway

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

type Embedder interface {
	Embed(context.Context, string) ([]float64, error)
}

type OllamaEmbedder struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOllamaEmbedder(baseURL, model string, timeout time.Duration) *OllamaEmbedder {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &OllamaEmbedder{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		model:   strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	if e == nil || e.baseURL == "" {
		return nil, fmt.Errorf("ollama embedder base url is required")
	}
	payload, err := json.Marshal(embedRequest{Model: e.model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute embed request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("embed status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var decoded embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(decoded.Embeddings) == 0 {
		return nil, fmt.Errorf("embed response missing embeddings")
	}
	return decoded.Embeddings[0], nil
}
