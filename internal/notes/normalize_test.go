package notes

import (
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

func TestNormalizeAndValidate_StripsBulletsAndDedupes(t *testing.T) {
	score := &scoring.Result{DecisionLabel: scoring.DecisionEscalate}
	note := NormalizeAndValidate(&AnalystNote{
		Status: StatusGenerated,
		Note:   "  Escalate this case.  ",
		EvidenceSummary: []string{
			"* Exact normalized name match",
			"- Exact normalized name match",
			"• Identifier match on passport",
		},
		MissingInformationSummary: []string{
			"* No contradictions noted",
			"* No contradictions noted",
		},
		NextStepRationale: "  Escalate for analyst review. ",
	}, score)

	if got := len(note.EvidenceSummary); got != 2 {
		t.Fatalf("EvidenceSummary len = %d, want 2; %+v", got, note.EvidenceSummary)
	}
	if got := note.EvidenceSummary[0]; len(got) > 0 && (got[0] == '*' || got[0] == '-') {
		t.Fatalf("expected bullet stripped, got %q", got)
	}
	if got := len(note.MissingInformationSummary); got != 1 {
		t.Fatalf("MissingInformationSummary len = %d, want 1", got)
	}
}

func TestNormalizeAndValidate_FailsWhenGeneratedButEmpty(t *testing.T) {
	score := &scoring.Result{DecisionLabel: scoring.DecisionEscalate}
	note := NormalizeAndValidate(&AnalystNote{
		Status:            StatusGenerated,
		Note:              " ",
		NextStepRationale: "Escalate.",
	}, score)

	if note.Status != StatusFailed {
		t.Fatalf("Status = %q, want %q", note.Status, StatusFailed)
	}
	if len(note.Warnings) == 0 {
		t.Fatal("expected warnings")
	}
}

func TestNormalizeAndValidate_WarnsOnPotentialInconsistency(t *testing.T) {
	score := &scoring.Result{DecisionLabel: scoring.DecisionEscalate}
	note := NormalizeAndValidate(&AnalystNote{
		Status:            StatusGenerated,
		Note:              "This case should be closed.",
		NextStepRationale: "Close the alert.",
	}, score)

	found := false
	for _, w := range note.Warnings {
		if w == "granite_analyst_note_potentially_inconsistent_with_decision" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected inconsistency warning, got %+v", note.Warnings)
	}
}
func TestNormalizeAndValidate_StripsMarkdownWrappedBullets(t *testing.T) {
	score := &scoring.Result{DecisionLabel: scoring.DecisionEscalate}
	note := NormalizeAndValidate(&AnalystNote{
		Status: StatusGenerated,
		Note:   "Escalate this case.",
		EvidenceSummary: []string{
			"*Strong corroborated match with sufficient data and no contradictions.*",
			"*Exact identifier match on passport.*",
		},
		MissingInformationSummary: []string{
			"*No explicit missing information noted in payload.*",
		},
		NextStepRationale: "Escalate for analyst review.",
	}, score)

	if got := note.EvidenceSummary[0]; got != "Strong corroborated match with sufficient data and no contradictions." {
		t.Fatalf("unexpected evidence summary: %q", got)
	}
	if got := note.EvidenceSummary[1]; got != "Exact identifier match on passport." {
		t.Fatalf("unexpected evidence summary: %q", got)
	}
}
