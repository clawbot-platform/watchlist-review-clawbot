package eval

import (
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

func TestEvaluateCase(t *testing.T) {
	resp := &api.ReviewResponse{
		DecisionLabel: "escalate",
		AnalystNote: &notes.AnalystNote{
			Status:            "generated",
			Note:              "Escalate for analyst review.",
			EvidenceSummary:   []string{"Exact normalized name match"},
			NextStepRationale: "Escalate for analyst review.",
		},
		ReviewContext: map[string]any{"retrieval_context": map[string]any{"warnings": []string{}}},
	}
	errors, warnings := EvaluateCase(CaseSpec{
		ExpectDecisionLabel:  "escalate",
		RequireGeneratedNote: true,
		RequireRetrieval:     true,
	}, resp)
	if len(errors) != 0 {
		t.Fatalf("errors = %+v", errors)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %+v", warnings)
	}
}
