package scoring

import (
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/ofacgraph"
)

func TestApplyRelationshipEvidence(t *testing.T) {
	result := &Result{
		MatchStrengthScore:    40,
		SupportingContextScore: 5,
	}
	ev := ofacgraph.RelationshipEvidence{
		RelationshipSupportScore:    6,
		OfficialDocLinkScore:        4,
		ProgramContextScore:         2,
		RelationshipConflictPenalty: 3,
		Reasons:                     []string{"direct OFAC relationship support"},
	}

	ApplyRelationshipEvidence(result, ev)

	if result.RelationshipSupportScore != 6 {
		t.Fatalf("RelationshipSupportScore = %d, want 6", result.RelationshipSupportScore)
	}
	if result.OfficialDocLinkScore != 4 {
		t.Fatalf("OfficialDocLinkScore = %d, want 4", result.OfficialDocLinkScore)
	}
	if result.ProgramContextScore != 2 {
		t.Fatalf("ProgramContextScore = %d, want 2", result.ProgramContextScore)
	}
	if result.RelationshipConflictPenalty != 3 {
		t.Fatalf("RelationshipConflictPenalty = %d, want 3", result.RelationshipConflictPenalty)
	}
}
