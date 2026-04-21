package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
)

type trainingExample struct {
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
	Labels map[string]any `json:"labels,omitempty"`
}

func main() {
	specPath := envOr("TUNING_EXPORT_BATCH_SPEC", "eval/goldset/manifest.calibrated.v1.json")
	reportPath := envOr("TUNING_EXPORT_REPORT_JSON", "eval/reports/latest/report.json")
	outPath := envOr("TUNING_EXPORT_OUT", "eval/training/note_tuning_v1.jsonl")

	spec, err := eval.LoadBatchSpec(specPath)
	if err != nil {
		log.Fatalf("load batch spec: %v", err)
	}

	report, err := loadReport(reportPath)
	if err != nil {
		log.Fatalf("load report: %v", err)
	}

	specByName := map[string]eval.CaseSpec{}
	for _, c := range spec.Cases {
		specByName[c.Name] = c
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		log.Fatalf("mkdir output dir: %v", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	count := 0
	for _, cr := range report.CaseResults {
		if !cr.Passed || cr.Response == nil || cr.Response.AnalystNote == nil {
			continue
		}
		specCase, ok := specByName[cr.Name]
		if !ok {
			continue
		}

		ex := trainingExample{
			Input: map[string]any{
				"case_name":          cr.Name,
				"request_path":       specCase.RequestPath,
				"expected_case_id":   cr.ExpectedCaseID,
				"expected_alert_id":  cr.ExpectedAlertID,
				"review_context":     cr.Response.ReviewContext,
				"decision_label":     cr.Response.DecisionLabel,
				"decision_reason":    cr.Response.DecisionReason,
				"match_strength":     cr.Response.MatchStrengthScore,
				"data_sufficiency":   cr.Response.DataSufficiencyScore,
				"evidence_for":       cr.Response.EvidenceFor,
				"evidence_against":   cr.Response.EvidenceAgainst,
				"next_step":          cr.Response.NextStep,
				"require_retrieval":  specCase.RequireRetrieval,
				"require_note":       specCase.RequireGeneratedNote,
				"exported_at":        time.Now().UTC().Format(time.RFC3339),
				"goldset_generation": "v1",
			},
			Output: map[string]any{
				"status":                      cr.Response.AnalystNote.Status,
				"note":                        cr.Response.AnalystNote.Note,
				"evidence_summary":            cr.Response.AnalystNote.EvidenceSummary,
				"missing_information_summary": cr.Response.AnalystNote.MissingInformationSummary,
				"next_step_rationale":         cr.Response.AnalystNote.NextStepRationale,
			},
			Labels: map[string]any{
				"accepted":          true,
				"decision_label":    cr.Response.DecisionLabel,
				"prompt_version":    cr.Response.AnalystNote.PromptVersion,
				"model":             cr.Response.AnalystNote.Model,
				"case_name":         cr.Name,
				"retrieval_required": specCase.RequireRetrieval,
				"quality_flags":     qualityFlags(cr.Response.AnalystNote),
			},
		}

		line, err := json.Marshal(ex)
		if err != nil {
			log.Fatalf("marshal jsonl line for %s: %v", cr.Name, err)
		}
		if _, err := w.Write(append(line, '\n')); err != nil {
			log.Fatalf("write jsonl line: %v", err)
		}
		count++
	}

	log.Printf("exported %d training examples to %s", count, outPath)
}

func loadReport(path string) (*eval.BatchReport, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var report eval.BatchReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func qualityFlags(note any) []string {
	b, _ := json.Marshal(note)
	text := string(b)

	var flags []string
	if containsAny(text,
		"3-5 concise grounded bullets",
		"0-4 concise grounded items",
		"without markdown bullet prefixes",
		"no placeholder text",
	) {
		flags = append(flags, "instruction_leakage")
	}
	if containsAny(stringsToLower(text),
		"no missing information",
		"all required evidence for decision is present",
		"all necessary details for decision-making are present",
	) {
		flags = append(flags, "placeholder_missing_info")
	}
	if len(flags) == 0 {
		return nil
	}
	return flags
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if n != "" && contains(s, n) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func stringsToLower(s string) string {
	return stringLower(s)
}

// intentionally tiny helpers to avoid extra imports in the first exporter patch

func stringLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func indexOf(s, sub string) int {
	n := len(sub)
	if n == 0 {
		return 0
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sub {
			return i
		}
	}
	return -1
}