package retrievalgateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

type fakeSearcher struct{}

func (f *fakeSearcher) Search(_ context.Context, query retrieval.Query) (retrieval.SearchResponse, error) {
	return retrieval.SearchResponse{
		Snippets: []retrieval.Snippet{{SnippetID: "snip-1", Text: query.Text}},
	}, nil
}

func TestSearchHandler(t *testing.T) {
	handler := NewHandler(&fakeSearcher{})
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body, _ := json.Marshal(retrieval.Query{Text: "hello", TopK: 1})
	req := httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}
