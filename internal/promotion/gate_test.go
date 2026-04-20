package promotion

import (
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/eval"
)

func TestEvaluatePromoted(t *testing.T) {
	spec := eval.BatchSpec{
		Cases: []eval.CaseSpec{
			{Name: "case-1", RequireGeneratedNote: true, RequireRetrieval: true},
			{Name: "case-2", RequireGeneratedNote: true, RequireRetrieval: false},
		},
	}
	report := eval.BatchReport{
		GeneratedAt: time.Now().UTC(),
		Passed:      true,
		CaseResults: []eval.CaseResult{
			{Name: "case-1", Passed: true},
			{Name: "case-2", Passed: true},
		},
	}
	decision := Evaluate(spec, report, DefaultConfig())
	if !decision.Promoted {
		t.Fatalf("expected promoted, got reasons=%+v", decision.Reasons)
	}
}

func TestEvaluateNotPromoted(t *testing.T) {
	spec := eval.BatchSpec{
		Cases: []eval.CaseSpec{
			{Name: "case-1", RequireGeneratedNote: true, RequireRetrieval: true},
		},
	}
	report := eval.BatchReport{
		GeneratedAt: time.Now().UTC(),
		Passed:      false,
		CaseResults: []eval.CaseResult{
			{Name: "case-1", Passed: false, Errors: []string{"retrieval_context missing from review_context"}},
		},
	}
	decision := Evaluate(spec, report, DefaultConfig())
	if decision.Promoted {
		t.Fatal("expected promotion failure")
	}
	if decision.Metrics.RetrievalRequiredFailures != 1 {
		t.Fatalf("retrieval failures = %d", decision.Metrics.RetrievalRequiredFailures)
	}
}
