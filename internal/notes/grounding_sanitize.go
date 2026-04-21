package notes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

func sanitizeGroundedNames(note *AnalystNote, alert *alerts.CanonicalAlert, screening any, score *scoring.Result) {
	if note == nil || alert == nil {
		return
	}

	allowed := allowedEntityNameSet(alert)
	if len(allowed) == 0 {
		return
	}

	forbidden := screeningOnlyNames(screening, allowed)
	if len(forbidden) == 0 {
		return
	}

	if containsAnyForbiddenName(note.Note, forbidden) {
		note.Note = deterministicGroundedNote(alert, score)
		appendWarning(note, "granite_analyst_note_rewritten_for_unapproved_name")
	}

	filteredEvidence, droppedEvidence := filterForbiddenNameItems(note.EvidenceSummary, forbidden)
	if droppedEvidence {
		note.EvidenceSummary = filteredEvidence
		appendWarning(note, "granite_analyst_note_evidence_items_dropped_for_unapproved_name")
	}
	if len(note.EvidenceSummary) == 0 {
		note.EvidenceSummary = fallbackEvidenceSummary(score)
	}

	filteredMissing, droppedMissing := filterForbiddenNameItems(note.MissingInformationSummary, forbidden)
	if droppedMissing {
		note.MissingInformationSummary = filteredMissing
		appendWarning(note, "granite_analyst_note_missing_items_dropped_for_unapproved_name")
	}

	if containsAnyForbiddenName(note.NextStepRationale, forbidden) {
		note.NextStepRationale = deterministicNextStepRationale(score)
		appendWarning(note, "granite_analyst_note_next_step_rationale_rewritten_for_unapproved_name")
	}
}

func allowedEntityNameSet(alert *alerts.CanonicalAlert) map[string]struct{} {
	out := map[string]struct{}{}
	for _, name := range allowedEntityNames(alert) {
		name = normalizeNameKey(name)
		if name == "" {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}

func screeningOnlyNames(screening any, allowed map[string]struct{}) []string {
	if screening == nil {
		return nil
	}

	raw, err := json.Marshal(screening)
	if err != nil {
		return nil
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}

	seen := map[string]struct{}{}
	var out []string

	var visit func(v any)
	visit = func(v any) {
		switch tv := v.(type) {
		case map[string]any:
			if candidates, ok := tv["candidates"]; ok {
				if arr, ok := candidates.([]any); ok {
					for _, item := range arr {
						obj, ok := item.(map[string]any)
						if !ok {
							continue
						}
						name, ok := obj["name"].(string)
						if !ok {
							continue
						}
						key := normalizeNameKey(name)
						if key == "" {
							continue
						}
						if _, ok := allowed[key]; ok {
							continue
						}
						if _, ok := seen[key]; ok {
							continue
						}
						seen[key] = struct{}{}
						out = append(out, strings.TrimSpace(name))
					}
				}
			}
			for _, child := range tv {
				visit(child)
			}
		case []any:
			for _, child := range tv {
				visit(child)
			}
		}
	}

	visit(decoded)
	return out
}

func filterForbiddenNameItems(items []string, forbidden []string) ([]string, bool) {
	if len(items) == 0 || len(forbidden) == 0 {
		return items, false
	}

	var out []string
	dropped := false
	for _, item := range items {
		if containsAnyForbiddenName(item, forbidden) {
			dropped = true
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil, dropped
	}
	return out, dropped
}

func containsAnyForbiddenName(text string, forbidden []string) bool {
	lower := strings.ToLower(text)
	for _, name := range forbidden {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			continue
		}
		if strings.Contains(lower, name) {
			return true
		}
	}
	return false
}

func deterministicGroundedNote(alert *alerts.CanonicalAlert, score *scoring.Result) string {
	screened := strings.TrimSpace(screenedName(alert))
	matched := strings.TrimSpace(matchedName(alert))

	subject := screened
	if subject == "" {
		subject = matched
	}
	if subject == "" {
		subject = "the screened entity"
	}

	switch {
	case score == nil:
		return fmt.Sprintf("The review for %s requires analyst attention based on the grounded alert details and deterministic assessment.", subject)
	case strings.EqualFold(score.DecisionLabel, scoring.DecisionEscalate):
		if screened != "" && matched != "" && !strings.EqualFold(screened, matched) {
			return fmt.Sprintf("The review for %s against %s is escalated because the deterministic assessment found strong corroborated match evidence and no contradictions.", screened, matched)
		}
		return fmt.Sprintf("The review for %s is escalated because the deterministic assessment found strong corroborated match evidence and no contradictions.", subject)
	case strings.EqualFold(score.DecisionLabel, scoring.DecisionInvestigateNextStep):
		if screened != "" && matched != "" && !strings.EqualFold(screened, matched) {
			return fmt.Sprintf("The review for %s against %s requires investigation because contradictions or sparse corroboration remain unresolved.", screened, matched)
		}
		return fmt.Sprintf("The review for %s requires investigation because contradictions or sparse corroboration remain unresolved.", subject)
	case strings.EqualFold(score.DecisionLabel, scoring.DecisionClose):
		if screened != "" && matched != "" && !strings.EqualFold(screened, matched) {
			return fmt.Sprintf("The review for %s against %s is closed as likely false positive because the available evidence does not support escalation.", screened, matched)
		}
		return fmt.Sprintf("The review for %s is closed as likely false positive because the available evidence does not support escalation.", subject)
	default:
		return fmt.Sprintf("The review for %s follows the deterministic decision based on the grounded alert details.", subject)
	}
}

func fallbackEvidenceSummary(score *scoring.Result) []string {
	if score == nil {
		return []string{"Grounded alert details reviewed for deterministic decision."}
	}

	var out []string
	for _, item := range score.EvidenceFor {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= 4 {
			return out
		}
	}
	for _, item := range score.EvidenceAgainst {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= 4 {
			return out
		}
	}
	if len(out) == 0 {
		out = append(out, "Grounded deterministic evidence reviewed.")
	}
	return out
}

func normalizeNameKey(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}
