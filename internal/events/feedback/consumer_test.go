package feedbackevents

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
)

func TestFeedbackEventShape(t *testing.T) {
	now := time.Now().UTC()
	event := feedback.FeedbackCreatedEvent{
		EventID:        "evt-1",
		EventType:      DefaultSubject,
		OccurredAt:     now,
		TenantID:       "tenant-1",
		CaseID:         "case-1",
		AlertID:        "alert-1",
		CorrelationID:  "corr-1",
		SystemDecision: "escalate",
		Feedback: feedback.AnalystFeedback{
			FeedbackID:        "fb-1",
			TenantID:          "tenant-1",
			CaseID:            "case-1",
			DecisionAgreement: feedback.DecisionAgreementDisagree,
			CreatedAt:         now,
			DerivedSignals:    []string{"prompt_tuning_candidate"},
		},
		ArtifactRef: &artifacts.ArtifactRef{
			ArtifactID:   "art-1",
			Kind:         artifacts.KindAnalystFeedback,
			RelativePath: "feedback/art-1.json",
		},
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded feedback.FeedbackCreatedEvent
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Feedback.FeedbackID != "fb-1" {
		t.Fatalf("FeedbackID = %q", decoded.Feedback.FeedbackID)
	}
}
