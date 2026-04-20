package retrievalgateway

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

func TestLocalJSONSearcher(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snippets.json")
	payload := `[
  {"snippet_id":"snip-1","source":"clawmem","title":"Passport corroboration","text":"Jane Citizen passport corroboration on prior SDN review.","tags":["sdn","passport","individual_onboarding"]},
  {"snippet_id":"snip-2","source":"clawmem","title":"Low value","text":"Unrelated note.","tags":["other"]}
]`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	searcher, err := NewLocalJSONSearcher(path)
	if err != nil {
		t.Fatalf("NewLocalJSONSearcher() error = %v", err)
	}
	resp, err := searcher.Search(context.Background(), retrieval.Query{
		Text: "individual_onboarding | screened=Jane Citizen | program=SDN | exact identifier match on passport",
		TopK: 1,
		Tags: []string{"individual_onboarding", "sdn"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Snippets) != 1 {
		t.Fatalf("snippets len = %d, want 1", len(resp.Snippets))
	}
	if resp.Snippets[0].SnippetID != "snip-1" {
		t.Fatalf("top snippet = %q", resp.Snippets[0].SnippetID)
	}
}
