package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
)

type fakeFeedbackStore struct{}

func (f *fakeFeedbackStore) WriteJSON(input artifacts.WriteInput) (artifacts.ArtifactRef, error) {
	return artifacts.ArtifactRef{
		ArtifactID:   "art-feedback-1",
		Kind:         input.Kind,
		ContentType:  "application/json",
		RelativePath: "feedback/art-feedback-1.json",
		CreatedAt:    time.Unix(0, 0).UTC(),
	}, nil
}

type fakeFeedbackManifest struct{}

func (f *fakeFeedbackManifest) UpsertCaseManifest(input artifacts.ManifestInput) (artifacts.ArtifactRef, error) {
	return artifacts.ArtifactRef{
		ArtifactID:   "manifest-case-1",
		Kind:         artifacts.KindCaseManifest,
		ContentType:  "application/json",
		RelativePath: "cases/tenant/case-1/case_manifest.json",
		CreatedAt:    time.Unix(1, 0).UTC(),
	}, nil
}

func TestFeedbackEndpointCapturesFeedback(t *testing.T) {
	feedbackService := feedback.NewService(&fakeFeedbackStore{}, &fakeFeedbackManifest{})
	server, err := NewServer(nil, feedbackService)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, err := json.Marshal(FeedbackRequest{
		TenantID:          "tenant-1",
		CaseID:            "case-1",
		AlertID:           "alert-1",
		AnalystID:         "analyst-1",
		SystemDecision:    "escalate",
		DecisionAgreement: feedback.DecisionAgreementDisagree,
		CorrectedLabel:    "investigate_next_step",
		NoteRating:        2,
		OutcomeRating:     3,
		Comment:           "Needs better explanation.",
		Tags:              []string{"prompt_issue"},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "feedback-test-001")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	var resp FeedbackResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if resp.Status != "feedback_captured" {
		t.Fatalf("status = %q", resp.Status)
	}
	if resp.Feedback.FeedbackID == "" {
		t.Fatal("expected feedback id")
	}
	if len(resp.ArtifactRefs) == 0 {
		t.Fatal("expected artifact refs")
	}
}
