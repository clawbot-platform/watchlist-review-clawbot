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
			"alert": map[string]any{
				"screened_party": map[string]any{"name": map[string]any{"full_name": "Jane Citizen"}},
				"matched_party":  map[string]any{"name": map[string]any{"full_name": "Jane Citizen"}},
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

func TestEvaluateCaseFlagsCrossEntityContamination(t *testing.T) {
	resp := &api.ReviewResponse{
		DecisionLabel: "escalate",
		AnalystNote: &notes.AnalystNote{
			Status:                    "generated",
			Note:                      "The organization matches Jane Citizen based on passport evidence.",
			EvidenceSummary:           []string{"address support"},
			MissingInformationSummary: []string{},
			NextStepRationale:         "Escalate for analyst review.",
		},
		ReviewContext: map[string]any{
			"retrieval_context": map[string]any{
				"snippets": []map[string]any{{"snippet_id": "snip-1"}},
			},
			"alert": map[string]any{
				"screened_party": map[string]any{"name": map[string]any{"full_name": "North Harbor Trading LLC"}},
				"matched_party":  map[string]any{"name": map[string]any{"full_name": "North Harbor Trading LLC"}},
			},
		},
	}
	errors, _ := EvaluateCase(CaseSpec{
		ExpectDecisionLabel:  "escalate",
		RequireGeneratedNote: true,
		RequireRetrieval:     true,
	}, resp)
	if len(errors) == 0 {
		t.Fatal("expected contamination error")
	}
}
