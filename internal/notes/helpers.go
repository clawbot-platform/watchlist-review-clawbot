package notes

import (
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

func appendWarning(note *AnalystNote, warning string) {
	if note == nil {
		return
	}
	warning = strings.TrimSpace(warning)
	if warning == "" {
		return
	}
	for _, existing := range note.Warnings {
		if strings.EqualFold(strings.TrimSpace(existing), warning) {
			return
		}
	}
	note.Warnings = append(note.Warnings, warning)
}

func deterministicNextStepRationale(score *scoring.Result) string {
	if score == nil {
		return ""
	}

	switch {
	case strings.TrimSpace(score.NextStep) != "":
		return ensureSentence(score.NextStep)
	case strings.TrimSpace(score.DecisionReason) != "":
		return ensureSentence(score.DecisionReason)
	default:
		return ""
	}
}
