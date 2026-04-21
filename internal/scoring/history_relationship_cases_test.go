package scoring

import "testing"

func TestApplyDeterministicAdjustments_PreviousConfirmedPairBoostsScore(t *testing.T) {
    result := &Result{
        MatchStrengthScore:          62,
        DataSufficiencyScore:        58,
        PreviousDispositionScore:    12,
        PreviousDispositionReasons:  []string{"same_pair previously confirmed true match"},
    }

    applyDeterministicAdjustments(result)

    if result.MatchStrengthScore <= 62 {
        t.Fatalf("MatchStrengthScore = %d, want > 62", result.MatchStrengthScore)
    }
    if len(result.EvidenceFor) == 0 {
        t.Fatalf("expected EvidenceFor to contain previous disposition reason")
    }
}

func TestApplyDeterministicAdjustments_RelationshipConflictPenalizesScore(t *testing.T) {
    result := &Result{
        MatchStrengthScore:          74,
        DataSufficiencyScore:        70,
        RelationshipConflictPenalty: 10,
    }

    applyDeterministicAdjustments(result)

    if result.MatchStrengthScore >= 74 {
        t.Fatalf("MatchStrengthScore = %d, want < 74", result.MatchStrengthScore)
    }
    if len(result.Contradictions) == 0 {
        t.Fatalf("expected relationship_conflict contradiction")
    }
}
