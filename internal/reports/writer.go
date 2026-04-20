package reports

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/promotion"
)

type ArtifactPaths struct {
	BaseDir             string `json:"base_dir"`
	ReportJSONPath      string `json:"report_json_path"`
	SummaryMarkdownPath string `json:"summary_markdown_path"`
	PromotionJSONPath   string `json:"promotion_json_path"`
}

func WriteAll(baseDir string, report eval.BatchReport, decision promotion.Decision) (ArtifactPaths, error) {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = filepath.Join("eval", "reports")
	}
	ts := report.GeneratedAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	stamp := ts.UTC().Format("20060102-150405")
	reportDir := filepath.Join(baseDir, stamp)
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return ArtifactPaths{}, fmt.Errorf("create report dir: %w", err)
	}

	paths := ArtifactPaths{
		BaseDir:             reportDir,
		ReportJSONPath:      filepath.Join(reportDir, "report.json"),
		SummaryMarkdownPath: filepath.Join(reportDir, "summary.md"),
		PromotionJSONPath:   filepath.Join(reportDir, "promotion.json"),
	}

	if err := writeJSON(paths.ReportJSONPath, report); err != nil {
		return ArtifactPaths{}, err
	}
	if err := writeMarkdown(paths.SummaryMarkdownPath, buildSummaryMarkdown(report, decision)); err != nil {
		return ArtifactPaths{}, err
	}
	if err := writeJSON(paths.PromotionJSONPath, decision); err != nil {
		return ArtifactPaths{}, err
	}

	return paths, nil
}

func writeJSON(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json for %s: %w", path, err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write json %s: %w", path, err)
	}
	return nil
}

func writeMarkdown(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write markdown %s: %w", path, err)
	}
	return nil
}

func buildSummaryMarkdown(report eval.BatchReport, decision promotion.Decision) string {
	var b strings.Builder
	b.WriteString("# Review Regression Summary\n\n")
	b.WriteString(fmt.Sprintf("- Generated at: `%s`\n", report.GeneratedAt.UTC().Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- Batch passed: `%t`\n", report.Passed))
	b.WriteString(fmt.Sprintf("- Promotion decision: `%t`\n", decision.Promoted))
	b.WriteString(fmt.Sprintf("- Total cases: `%d`\n", decision.Metrics.TotalCases))
	b.WriteString(fmt.Sprintf("- Passed cases: `%d`\n", decision.Metrics.PassedCases))
	b.WriteString(fmt.Sprintf("- Deterministic regressions: `%d`\n", decision.Metrics.DeterministicRegressions))
	b.WriteString(fmt.Sprintf("- Retrieval-required failures: `%d`\n", decision.Metrics.RetrievalRequiredFailures))
	b.WriteString(fmt.Sprintf("- Note-quality pass rate: `%.2f`\n", decision.Metrics.NoteQualityPassRate))
	if len(decision.Reasons) > 0 {
		b.WriteString("\n## Promotion gate reasons\n\n")
		for _, reason := range decision.Reasons {
			b.WriteString("- " + reason + "\n")
		}
	}
	if len(report.CaseResults) > 0 {
		b.WriteString("\n## Case results\n\n")
		for _, result := range report.CaseResults {
			status := "PASS"
			if !result.Passed {
				status = "FAIL"
			}
			b.WriteString(fmt.Sprintf("### %s — %s\n\n", result.Name, status))
			if len(result.Errors) > 0 {
				b.WriteString("Errors:\n")
				for _, item := range result.Errors {
					b.WriteString("- " + item + "\n")
				}
			}
			if len(result.Warnings) > 0 {
				b.WriteString("Warnings:\n")
				for _, item := range result.Warnings {
					b.WriteString("- " + item + "\n")
				}
			}
			if len(result.Errors) == 0 && len(result.Warnings) == 0 {
				b.WriteString("- No issues\n")
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
