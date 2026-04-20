package notes

import (
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

const (
	maxNoteLen      = 1200
	maxSummaryItems = 6
	maxMissingItems = 5
	maxItemLen      = 220
)

func NormalizeAndValidate(note *AnalystNote, score *scoring.Result) *AnalystNote {
	if note == nil {
		return &AnalystNote{
			Status:   StatusFailed,
			Warnings: []string{"granite_analyst_note_empty"},
		}
	}

	if note.Status == "" {
		note.Status = StatusGenerated
	}
	if note.PromptVersion == "" {
		note.PromptVersion = promptVersion
	}

	note.Note = clampString(cleanSentence(note.Note), maxNoteLen)
	note.EvidenceSummary = normalizeItems(note.EvidenceSummary, maxSummaryItems)
	note.MissingInformationSummary = normalizeItems(note.MissingInformationSummary, maxMissingItems)
	note.NextStepRationale = clampString(cleanSentence(note.NextStepRationale), maxItemLen)
	note.Warnings = dedupeStrings(note.Warnings)

	if note.Status == StatusGenerated {
		if note.Note == "" {
			note.Status = StatusFailed
			note.Warnings = append(note.Warnings, "granite_analyst_note_missing_note")
		}
		if len(note.EvidenceSummary) == 0 {
			note.Warnings = append(note.Warnings, "granite_analyst_note_missing_evidence_summary")
		}
		if note.NextStepRationale == "" {
			note.Warnings = append(note.Warnings, "granite_analyst_note_missing_next_step_rationale")
		}
	}

	if score != nil && inconsistentWithDecision(note, score) {
		note.Warnings = append(note.Warnings, "granite_analyst_note_potentially_inconsistent_with_decision")
	}

	note.Warnings = dedupeStrings(note.Warnings)
	return note
}

func normalizeItems(in []string, limit int) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, item := range in {
		cleaned := clampString(cleanBullet(item), maxItemLen)
		if cleaned == "" {
			continue
		}
		key := strings.ToLower(cleaned)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, cleaned)
		if len(out) >= limit {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cleanBullet(s string) string {
	s = strings.TrimSpace(s)
	for {
		switch {
		case strings.HasPrefix(s, "* "):
			s = strings.TrimSpace(strings.TrimPrefix(s, "* "))
		case strings.HasPrefix(s, "- "):
			s = strings.TrimSpace(strings.TrimPrefix(s, "- "))
		case strings.HasPrefix(s, "• "):
			s = strings.TrimSpace(strings.TrimPrefix(s, "• "))
		default:
			return cleanSentence(s)
		}
	}
}

func cleanSentence(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\t", " "))
	return strings.Join(strings.Fields(s), " ")
}

func clampString(s string, limit int) string {
	if limit <= 0 || len(s) <= limit {
		return s
	}
	return strings.TrimSpace(s[:limit])
}

func dedupeStrings(in []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, item := range in {
		item = cleanSentence(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func inconsistentWithDecision(note *AnalystNote, score *scoring.Result) bool {
	if note == nil || score == nil {
		return false
	}
	text := strings.ToLower(note.Note + " " + note.NextStepRationale)
	switch strings.ToLower(score.DecisionLabel) {
	case "escalate":
		return strings.Contains(text, "close") || strings.Contains(text, "do not escalate")
	case "close":
		return strings.Contains(text, "escalate")
	case "investigate_next_step":
		return strings.Contains(text, "close immediately")
	default:
		return false
	}
}
