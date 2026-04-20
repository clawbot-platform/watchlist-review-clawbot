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

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

type QdrantSearcher struct {
	baseURL    string
	apiKey     string
	collection string
	embedder   Embedder
	httpClient *http.Client
}

func NewQdrantSearcher(baseURL, apiKey, collection string, embedder Embedder, timeout time.Duration) *QdrantSearcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &QdrantSearcher{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		collection: strings.TrimSpace(collection),
		embedder:   embedder,
		httpClient: &http.Client{Timeout: timeout},
	}
}

type qdrantQueryRequest struct {
	Query       []float64      `json:"query"`
	Limit       int            `json:"limit"`
	WithPayload bool           `json:"with_payload"`
	Filter      map[string]any `json:"filter,omitempty"`
}

type qdrantQueryResponse struct {
	Result struct {
		Points []struct {
			ID      any            `json:"id"`
			Score   float64        `json:"score"`
			Payload map[string]any `json:"payload"`
		} `json:"points"`
	} `json:"result"`
}

func (s *QdrantSearcher) Search(ctx context.Context, query retrieval.Query) (retrieval.SearchResponse, error) {
	if s == nil || s.baseURL == "" {
		return retrieval.SearchResponse{}, fmt.Errorf("qdrant base url is required")
	}
	if s.collection == "" {
		return retrieval.SearchResponse{}, fmt.Errorf("qdrant collection is required")
	}
	if s.embedder == nil {
		return retrieval.SearchResponse{}, fmt.Errorf("qdrant embedder is not configured")
	}
	vector, err := s.embedder.Embed(ctx, query.Text)
	if err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("embed query: %w", err)
	}
	limit := query.TopK
	if limit <= 0 {
		limit = 4
	}
	reqBody := qdrantQueryRequest{
		Query:       vector,
		Limit:       limit,
		WithPayload: true,
	}
	if strings.TrimSpace(query.TenantID) != "" {
		reqBody.Filter = map[string]any{
			"must": []map[string]any{
				{
					"key": "tenant_id",
					"match": map[string]any{"value": strings.TrimSpace(query.TenantID)},
				},
			},
		}
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("marshal qdrant request: %w", err)
	}
	url := s.baseURL + "/collections/" + s.collection + "/points/query"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("build qdrant request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("api-key", s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("execute qdrant request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return retrieval.SearchResponse{}, fmt.Errorf("qdrant status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded qdrantQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("decode qdrant response: %w", err)
	}

	var snippets []retrieval.Snippet
	for _, point := range decoded.Result.Points {
		tags := anyStrings(point.Payload["tags"])
		snippets = append(snippets, retrieval.Snippet{
			SnippetID: fmt.Sprint(point.ID),
			Source:    anyString(point.Payload["source"]),
			Title:     anyString(point.Payload["title"]),
			Text:      anyString(point.Payload["text"]),
			Score:     point.Score,
			Tags:      tags,
		})
	}
	return retrieval.SearchResponse{Snippets: snippets}, nil
}

func anyString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func anyStrings(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
