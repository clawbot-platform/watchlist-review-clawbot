package reports

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/promotion"
)

func TestWriteAll(t *testing.T) {
	dir := t.TempDir()
	report := eval.BatchReport{
		GeneratedAt: time.Date(2026, 4, 20, 20, 26, 0, 0, time.UTC),
		Passed:      true,
		CaseResults: []eval.CaseResult{{Name: "case-1", Passed: true}},
	}
	decision := promotion.Decision{
		Promoted: true,
		Metrics: promotion.Metrics{
			TotalCases:                 1,
			PassedCases:                1,
			NoteQualityPassRate:        1.0,
			DeterministicRegressions:   0,
			RetrievalRequiredFailures:  0,
		},
	}

	paths, err := WriteAll(filepath.Join(dir, "reports"), report, decision)
	if err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	for _, path := range []string{paths.ReportJSONPath, paths.SummaryMarkdownPath, paths.PromotionJSONPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
	}
}
