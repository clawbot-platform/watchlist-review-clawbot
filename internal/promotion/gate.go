package promotion

import (
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
)

type Config struct {
	RequireZeroDeterministicRegressions bool    `json:"require_zero_deterministic_regressions"`
	RequireZeroRetrievalFailures        bool    `json:"require_zero_retrieval_failures"`
	MinNoteQualityPassRate              float64 `json:"min_note_quality_pass_rate"`
}

type Metrics struct {
	TotalCases                int     `json:"total_cases"`
	PassedCases               int     `json:"passed_cases"`
	DeterministicRegressions  int     `json:"deterministic_regressions"`
	RetrievalRequiredFailures int     `json:"retrieval_required_failures"`
	NoteQualityPassRate       float64 `json:"note_quality_pass_rate"`
}

type Decision struct {
	Promoted bool     `json:"promoted"`
	Reasons  []string `json:"reasons,omitempty"`
	Metrics  Metrics  `json:"metrics"`
	Config   Config   `json:"config"`
}

func DefaultConfig() Config {
	return Config{
		RequireZeroDeterministicRegressions: true,
		RequireZeroRetrievalFailures:        true,
		MinNoteQualityPassRate:              0.90,
	}
}

func Evaluate(spec eval.BatchSpec, report eval.BatchReport, cfg Config) Decision {
	if cfg.MinNoteQualityPassRate <= 0 {
		cfg = DefaultConfig()
	}

	caseLookup := map[string]eval.CaseSpec{}
	for _, testCase := range spec.Cases {
		caseLookup[testCase.Name] = testCase
	}

	metrics := Metrics{
		TotalCases: len(report.CaseResults),
	}
	noteEligible := 0
	notePassed := 0

	for _, result := range report.CaseResults {
		if result.Passed {
			metrics.PassedCases++
		}
		testCase, ok := caseLookup[result.Name]
		if ok && testCase.RequireGeneratedNote {
			noteEligible++
			if result.Passed {
				notePassed++
			}
		}
		for _, errText := range result.Errors {
			lower := strings.ToLower(errText)
			if strings.Contains(lower, "unexpected decision_label") {
				metrics.DeterministicRegressions++
			}
			if ok && testCase.RequireRetrieval {
				if strings.Contains(lower, "retrieval_context") ||
					strings.Contains(lower, "retrieval failure") ||
					strings.Contains(lower, "retrieval_failed") ||
					strings.Contains(lower, "retrieval_not_configured") {
					metrics.RetrievalRequiredFailures++
					break
				}
			}
		}
	}

	if noteEligible == 0 {
		metrics.NoteQualityPassRate = 1.0
	} else {
		metrics.NoteQualityPassRate = float64(notePassed) / float64(noteEligible)
	}

	var reasons []string
	if !report.Passed {
		reasons = append(reasons, "batch report contains failing cases")
	}
	if cfg.RequireZeroDeterministicRegressions && metrics.DeterministicRegressions > 0 {
		reasons = append(reasons, fmt.Sprintf("deterministic regressions: %d", metrics.DeterministicRegressions))
	}
	if cfg.RequireZeroRetrievalFailures && metrics.RetrievalRequiredFailures > 0 {
		reasons = append(reasons, fmt.Sprintf("retrieval-required failures: %d", metrics.RetrievalRequiredFailures))
	}
	if metrics.NoteQualityPassRate < cfg.MinNoteQualityPassRate {
		reasons = append(reasons, fmt.Sprintf("note-quality pass rate %.2f below threshold %.2f", metrics.NoteQualityPassRate, cfg.MinNoteQualityPassRate))
	}

	return Decision{
		Promoted: len(reasons) == 0,
		Reasons:  reasons,
		Metrics:  metrics,
		Config:   cfg,
	}
}
