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

			for _, item := range resp.AnalystNote.EvidenceSummary {
				trimmed := strings.TrimSpace(item)
				if strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "•") {
					errors = append(errors, "analyst_note.evidence_summary contains bullet-prefixed item: "+trimmed)
					break
				}
			}

			if rationaleLooksTruncated(resp.AnalystNote.NextStepRationale) {
				errors = append(errors, "analyst_note.next_step_rationale appears truncated")
			}

			for _, item := range resp.AnalystNote.MissingInformationSummary {
				lower := strings.ToLower(strings.TrimSpace(item))
				if strings.Contains(lower, "no explicit missing information") ||
					strings.Contains(lower, "no missing information") ||
					strings.Contains(lower, "none identified") ||
					strings.Contains(lower, "no missing or weak information") ||
					strings.Contains(lower, "no explicitly noted missing information") {
					errors = append(errors, "analyst_note.missing_information_summary contains placeholder text")
					break
				}
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
		text := string(raw)

		if !strings.Contains(text, "retrieval_context") {
			errors = append(errors, "retrieval_context missing from review_context")
		} else {
			if strings.Contains(text, "retrieval_failed") || strings.Contains(text, "retrieval_not_configured") {
				errors = append(errors, "retrieval_context contains retrieval failure or not-configured warning")
			}
			if !strings.Contains(text, "\"snippets\"") {
				errors = append(errors, "retrieval_context does not contain snippets")
			}
		}
	}

	if resp.AnalystNote != nil {
		if strings.Contains(strings.ToLower(resp.AnalystNote.Note), "i think") {
			warnings = append(warnings, "analyst note contains hedging language")
		}
	}

	return dedupe(errors), dedupe(warnings)
}

func rationaleLooksTruncated(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}

	last := s[len(s)-1]
	if last == '.' || last == '!' || last == '?' {
		return false
	}

	return true
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	var out []string

	for _, item := range in {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}