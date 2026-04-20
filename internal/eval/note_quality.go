package eval

import (
	"encoding/json"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
)

func EvaluateCase(spec CaseSpec, resp *api.ReviewResponse) ([]string, []string) {
	var errors []string
	var warnings []string
	if resp == nil {
		return []string{"nil response"}, nil
	}
	if spec.ExpectDecisionLabel != "" && resp.DecisionLabel != spec.ExpectDecisionLabel {
		errors = append(errors, "unexpected decision_label: got="+resp.DecisionLabel+" want="+spec.ExpectDecisionLabel)
	}
	if spec.RequireGeneratedNote {
		if resp.AnalystNote == nil {
			errors = append(errors, "missing analyst_note")
		} else {
			if resp.AnalystNote.Status != "generated" {
				errors = append(errors, "analyst_note.status="+resp.AnalystNote.Status+" want=generated")
			}
			if strings.TrimSpace(resp.AnalystNote.Note) == "" {
				errors = append(errors, "analyst_note.note is empty")
			}
			if len(resp.AnalystNote.EvidenceSummary) == 0 {
				errors = append(errors, "analyst_note.evidence_summary is empty")
			}
			if strings.TrimSpace(resp.AnalystNote.NextStepRationale) == "" {
				errors = append(errors, "analyst_note.next_step_rationale is empty")
			}
			for _, warning := range resp.AnalystNote.Warnings {
				if warning == "granite_analyst_note_potentially_inconsistent_with_decision" {
					errors = append(errors, "analyst note inconsistent with deterministic decision")
				}
			}
		}
	}
	if spec.RequireRetrieval {
		raw, _ := json.Marshal(resp.ReviewContext)
		if !strings.Contains(string(raw), "retrieval_context") {
			errors = append(errors, "retrieval_context missing from review_context")
		}
	}
	if resp.AnalystNote != nil {
		if strings.Contains(strings.ToLower(resp.AnalystNote.Note), "i think") {
			warnings = append(warnings, "analyst note contains hedging language")
		}
	}
	return errors, warnings
}
