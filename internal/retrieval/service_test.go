package retrieval

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

func TestBuildPromptContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"snippets":[{"snippet_id":"snip-1","source":"clawmem","title":"Prior review","text":"Prior case involved passport corroboration.","score":0.91}]}`))
	}))
	defer server.Close()

	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{
			AlertID:      "alert-1",
			SourceSystem: "screening_json",
			CaseID:       "case-1",
			CreatedAt:    time.Now().UTC(),
		},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "watchlist_candidates",
			Program:    "SDN",
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
	}
	fx, err := features.Extract(alert)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	score := scoring.Evaluate(alert, fx)

	service := NewService(NewClient(server.URL, 5*time.Second))
	ctx := service.BuildPromptContext(context.Background(), "tenant-1", alert, score)
	if ctx == nil {
		t.Fatal("expected prompt context")
	}
	if len(ctx.Snippets) != 1 {
		t.Fatalf("snippets len = %d, want 1", len(ctx.Snippets))
	}
	if ctx.QueryText == "" {
		t.Fatal("expected query text")
	}
}
