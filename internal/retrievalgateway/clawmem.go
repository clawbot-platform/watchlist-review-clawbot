package retrievalgateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

type Searcher interface {
	Search(context.Context, retrieval.Query) (retrieval.SearchResponse, error)
}

type ClawmemHTTPSearcher struct {
	client *retrieval.Client
}

func NewClawmemHTTPSearcher(baseURL string) *ClawmemHTTPSearcher {
	return &ClawmemHTTPSearcher{
		client: retrieval.NewClient(strings.TrimSpace(baseURL), 10*time.Second),
	}
}

func (s *ClawmemHTTPSearcher) Search(ctx context.Context, query retrieval.Query) (retrieval.SearchResponse, error) {
	if s == nil || s.client == nil || s.client.BaseURL() == "" {
		return retrieval.SearchResponse{}, fmt.Errorf("clawmem retrieval client is not configured")
	}
	return s.client.Search(ctx, query)
}
