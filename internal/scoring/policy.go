package scoring

// ApplySparseContradictionPolicy should be called after the base deterministic
// result is computed and before it is returned to callers.
//
// Intent:
//   - preserve strong corroborated escalations
//   - avoid auto-closing sparse contradiction-heavy cases
//   - send weak but contradiction-heavy cases to investigate_next_step
func ApplySparseContradictionPolicy(result *Result) {
	if result == nil {
		return
	}
	if result.DecisionLabel == DecisionEscalate {
		return
	}
	if !shouldInvestigateSparseContradictory(result) {
		return
	}

	result.DecisionLabel = DecisionInvestigateNextStep
	result.DecisionReason = "Sparse or contradictory case requires analyst investigation before closure."
	result.NextStep = "Investigate contradiction and gather additional corroboration."

	appendUnique(&result.MissingInformation, "additional corroboration")
	appendUnique(&result.EvidenceAgainst, "hard contradiction requires analyst confirmation")
}

func shouldInvestigateSparseContradictory(result *Result) bool {
	if result == nil {
		return false
	}

	contradictions := len(result.Contradictions)
	missing := len(result.MissingInformation)
	against := len(result.EvidenceAgainst)

	lowStrength := result.MatchStrengthScore <= 25
	moderateStrength := result.MatchStrengthScore <= 45
	lowSufficiency := result.DataSufficiencyScore <= 70

	// Very weak plus contradiction plus sparse evidence should investigate.
	if lowStrength && contradictions >= 1 && missing >= 2 {
		return true
	}

	// Multiple contradictions with sparse support should investigate.
	if contradictions >= 2 && missing >= 2 {
		return true
	}

	// Weak / moderate match plus contradictions and adverse evidence should investigate.
	if contradictions >= 1 && against >= 2 && lowSufficiency {
		return true
	}
	if moderateStrength && contradictions >= 2 {
		return true
	}

	return false
}

func appendUnique(dst *[]string, value string) {
	for _, current := range *dst {
		if current == value {
			return
		}
	}
	*dst = append(*dst, value)
}
