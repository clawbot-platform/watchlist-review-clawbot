package feedback

import (
	"context"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
)

type fakeStore struct{}

func (f *fakeStore) WriteJSON(input artifacts.WriteInput) (artifacts.ArtifactRef, error) {
	return artifacts.ArtifactRef{
		ArtifactID:   "art-feedback-1",
		Kind:         input.Kind,
		ContentType:  "application/json",
		RelativePath: "feedback/art-feedback-1.json",
		CreatedAt:    time.Unix(0, 0).UTC(),
	}, nil
}

type fakeManifestWriter struct{}

func (f *fakeManifestWriter) UpsertCaseManifest(input artifacts.ManifestInput) (artifacts.ArtifactRef, error) {
	return artifacts.ArtifactRef{
		ArtifactID:   "manifest-case-1",
		Kind:         artifacts.KindCaseManifest,
		ContentType:  "application/json",
		RelativePath: "cases/tenant/case-1/case_manifest.json",
		CreatedAt:    time.Unix(1, 0).UTC(),
	}, nil
}

type fakePublisher struct {
	events []FeedbackCreatedEvent
}

func (f *fakePublisher) PublishFeedbackCreated(_ context.Context, event FeedbackCreatedEvent) error {
	f.events = append(f.events, event)
	return nil
}

func TestCreateFeedback(t *testing.T) {
	publisher := &fakePublisher{}
	service := NewService(&fakeStore{}, &fakeManifestWriter{}, publisher)
	result, err := service.Create(context.Background(), CreateInput{
		TenantID:          "tenant-1",
		CaseID:            "case-1",
		AlertID:           "alert-1",
		SystemDecision:    "escalate",
		DecisionAgreement: DecisionAgreementDisagree,
		CorrectedLabel:    "investigate_next_step",
		NoteRating:        2,
		OutcomeRating:     3,
		Comment:           "Needs better explanation.",
		Tags:              []string{"prompt_issue", "retrieval_gap"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(result.ArtifactRefs) != 2 {
		t.Fatalf("artifact refs len = %d, want 2", len(result.ArtifactRefs))
	}
	if len(publisher.events) != 1 {
		t.Fatalf("publisher events len = %d, want 1", len(publisher.events))
	}
	if len(result.Feedback.DerivedSignals) == 0 {
		t.Fatal("expected derived signals")
	}
}
