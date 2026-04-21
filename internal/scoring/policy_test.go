package scoring

import "testing"

func TestApplySparseContradictionPolicy_SparseIndividualFlipsToInvestigate(t *testing.T) {
	result := Result{
		DecisionLabel:        DecisionClose,
		DecisionReason:       "Low match strength with insufficient evidence for escalation.",
		NextStep:             "Close as likely false positive.",
		MatchStrengthScore:   9,
		DataSufficiencyScore: 42,
		Contradictions:       []string{"geography_conflict"},
		EvidenceAgainst:      []string{"country mismatch"},
		MissingInformation:   []string{"date corroboration", "strong identifier corroboration", "address context"},
	}
	ApplySparseContradictionPolicy(&result)
	if result.DecisionLabel != DecisionInvestigateNextStep {
		t.Fatalf("DecisionLabel = %q, want %q", result.DecisionLabel, DecisionInvestigateNextStep)
	}
	if result.NextStep != "Investigate contradiction and gather additional corroboration." {
		t.Fatalf("NextStep = %q", result.NextStep)
	}
}

func TestApplySparseContradictionPolicy_OrgWrongContextFlipsToInvestigate(t *testing.T) {
	result := Result{
		DecisionLabel:        DecisionClose,
		DecisionReason:       "Hard contradiction present and overall match strength is weak.",
		NextStep:             "Close as likely false positive and document contradiction.",
		MatchStrengthScore:   17,
		DataSufficiencyScore: 78,
		Contradictions:       []string{"dob_conflict", "geography_conflict", "identifier_conflict"},
		EvidenceAgainst:      []string{"country mismatch", "date conflict", "identifier conflict"},
		MissingInformation:   []string{"date corroboration", "geography corroboration", "strong identifier corroboration"},
	}
	ApplySparseContradictionPolicy(&result)
	if result.DecisionLabel != DecisionInvestigateNextStep {
		t.Fatalf("DecisionLabel = %q, want %q", result.DecisionLabel, DecisionInvestigateNextStep)
	}
}

func TestApplySparseContradictionPolicy_KeepEscalateUntouched(t *testing.T) {
	result := Result{
		DecisionLabel:        DecisionEscalate,
		DecisionReason:       "Strong corroborated match with sufficient data and no contradictions.",
		NextStep:             "Escalate for analyst review.",
		MatchStrengthScore:   93,
		DataSufficiencyScore: 92,
	}
	ApplySparseContradictionPolicy(&result)
	if result.DecisionLabel != DecisionEscalate {
		t.Fatalf("DecisionLabel = %q, want %q", result.DecisionLabel, DecisionEscalate)
	}
}
