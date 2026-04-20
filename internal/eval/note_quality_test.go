package eval

import (
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

func TestEvaluateCaseHappyPath(t *testing.T) {
	resp := &api.ReviewResponse{
		DecisionLabel: "escalate",
		AnalystNote: &notes.AnalystNote{
			Status:                    "generated",
			Note:                      "Escalate for analyst review based on corroborated evidence.",
			EvidenceSummary:           []string{"Exact normalized name match", "Exact identifier match on passport"},
			MissingInformationSummary: []string{"Additional source documentation could further support the case."},
			NextStepRationale:         "Escalate for analyst review.",
		},
		ReviewContext: map[string]any{
			"retrieval_context": map[string]any{
				"snippets": []map[string]any{{"snippet_id": "snip-1"}},
			},
		},
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

func TestEvaluateCaseCatchesPoorNoteQuality(t *testing.T) {
	resp := &api.ReviewResponse{
		DecisionLabel: "escalate",
		AnalystNote: &notes.AnalystNote{
			Status:                    "generated",
			Note:                      "Escalate.",
			EvidenceSummary:           []string{"* Strong corroborated match"},
			MissingInformationSummary: []string{"*No explicit missing information noted in payload."},
			NextStepRationale:         "Escalate for analyst review due to exact date",
		},
		ReviewContext: map[string]any{
			"retrieval_context": map[string]any{
				"warnings": []string{"retrieval_failed: 404"},
			},
		},
	}
	errors, _ := EvaluateCase(CaseSpec{
		ExpectDecisionLabel:  "escalate",
		RequireGeneratedNote: true,
		RequireRetrieval:     true,
	}, resp)
	if len(errors) < 4 {
		t.Fatalf("expected multiple errors, got %+v", errors)
	}
}
