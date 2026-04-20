package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

type fakeGenerator struct {
	note *notes.AnalystNote
	err  error
}

func (f *fakeGenerator) Generate(_ context.Context, input notes.PromptInput) (*notes.AnalystNote, error) {
	if input.RetrievalContext == nil {
		return nil, errors.New("missing retrieval context")
	}
	return f.note, f.err
}

type fakeRetrievalService struct{}

func (f *fakeRetrievalService) BuildPromptContext(_ context.Context, _ string, _ any, _ any) *retrieval.PromptContext {
	return &retrieval.PromptContext{
		QueryText: "individual_onboarding | screened=Jane Citizen",
		Snippets: []retrieval.Snippet{
			{SnippetID: "snip-1", Source: "clawmem", Title: "Prior review", Text: "Prior passport corroboration case."},
		},
	}
}

func TestReviewResponseIncludesGeneratedAnalystNote(t *testing.T) {
	server, err := NewServer(nil, notes.NewService(&fakeGenerator{
		note: &notes.AnalystNote{
			Status:            notes.StatusGenerated,
			Model:             "ibm/granite3.3:8b",
			PromptVersion:     "granite-analyst-note-v2-rag",
			Note:              "Escalate this case based on strong corroborated evidence.",
			EvidenceSummary:   []string{"* Exact normalized name match", "* Identifier match on passport"},
			NextStepRationale: "Escalate for analyst review.",
		},
	}), nil, &fakeRetrievalServiceAdapter{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body := buildFixtureBody(t, "alert_with_source_refs.json")
	req := httptest.NewRequest(http.MethodPost, "/v1/reviews", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	var resp ReviewResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if resp.AnalystNote == nil {
		t.Fatal("expected analyst_note")
	}
	if resp.AnalystNote.Status != notes.StatusGenerated {
		t.Fatalf("analyst_note.status = %q", resp.AnalystNote.Status)
	}
	if resp.ReviewContext == nil {
		t.Fatal("expected review_context")
	}
}

type fakeRetrievalServiceAdapter struct{}

func (f *fakeRetrievalServiceAdapter) BuildPromptContext(ctx context.Context, tenant string, alert any, score any) *retrieval.PromptContext {
	return (&fakeRetrievalService{}).BuildPromptContext(ctx, tenant, alert, score)
}

func TestReviewResponseIncludesSkippedAnalystNoteWhenNotConfigured(t *testing.T) {
	server, err := NewServer(nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body := buildFixtureBody(t, "alert_with_source_refs.json")
	req := httptest.NewRequest(http.MethodPost, "/v1/reviews", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var resp ReviewResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if resp.AnalystNote == nil || resp.AnalystNote.Status != notes.StatusSkipped {
		t.Fatalf("expected skipped analyst note, got %+v", resp.AnalystNote)
	}
}

func TestReviewResponseIncludesFailedAnalystNoteWithoutBreakingDeterministicDecision(t *testing.T) {
	server, err := NewServer(nil, notes.NewService(&fakeGenerator{
		err: errors.New("upstream model error"),
	}), nil, &fakeRetrievalServiceAdapter{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body := buildFixtureBody(t, "alert_with_source_refs.json")
	req := httptest.NewRequest(http.MethodPost, "/v1/reviews", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	var resp ReviewResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if resp.AnalystNote == nil || resp.AnalystNote.Status != notes.StatusFailed {
		t.Fatalf("expected failed analyst note, got %+v", resp.AnalystNote)
	}
	if resp.DecisionLabel == "" {
		t.Fatal("expected deterministic decision label to remain present")
	}
}

func buildFixtureBody(t *testing.T, name string) []byte {
	t.Helper()
	rawAlert := mustReadGraniteFixture(t, name)
	body := map[string]any{
		"tenant_id":     "test-tenant",
		"case_id":       "test-case",
		"source_system": "screening_json",
		"raw_alert":     json.RawMessage(rawAlert),
		"options": map[string]any{
			"explain": true,
			"mode":    "deterministic",
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return payload
}

func mustReadGraniteFixture(t *testing.T, name string) []byte {
	t.Helper()
	candidates := []string{
		filepath.Join("test", "fixtures", "screeningjson", name),
		filepath.Join("..", "..", "test", "fixtures", "screeningjson", name),
	}
	for _, path := range candidates {
		if raw, err := os.ReadFile(path); err == nil {
			return raw
		}
	}
	t.Fatalf("fixture %q not found in expected locations", name)
	return nil
}
