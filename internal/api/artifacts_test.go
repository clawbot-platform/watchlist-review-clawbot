package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
)

func TestReviewResponseIncludesArtifactRefs(t *testing.T) {
	store := artifacts.NewFileSystemStore(t.TempDir())
	server, err := NewServer(nil, artifacts.NewService(store))
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
	if len(resp.ArtifactRefs) == 0 {
		t.Fatal("expected artifact_refs")
	}
	for _, ref := range resp.ArtifactRefs {
		if ref.RelativePath == "" {
			t.Fatalf("expected relative path in ref %+v", ref)
		}
		if filepath.Ext(ref.RelativePath) != ".json" {
			t.Fatalf("expected json artifact path, got %q", ref.RelativePath)
		}
	}
}
