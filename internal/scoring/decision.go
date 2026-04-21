package scoring

import (
	
	"sort"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
)

const (
	DecisionClose               = "close"
	DecisionInvestigateNextStep = "investigate_next_step"
	DecisionEscalate            = "escalate"
	DecisionReviewPending       = "review_pending"
)

func Evaluate(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures) *Result {
	result := &Result{}
	if alert == nil || fx == nil {
		result.DecisionLabel = DecisionReviewPending
		result.DecisionReason = "alert or extracted features unavailable"
		result.NextStep = "Review worker inputs and retry"
		return result
	}

	result.Contradictions = uniqueStrings(append([]string(nil), fx.Contradictions...))

	computeMatchStrength(fx, result)
	computeDataSufficiency(fx, result)
	result.MissingInformation = deriveMissingInformation(fx)

	applyDeterministicAdjustments(result)
	decide(alert, fx, result)

	return result
}

func applyDeterministicAdjustments(result *Result) {
	if result == nil {
		return
	}

	// 1) Previous reviewed disposition signal.
	if result.PreviousDispositionScore != 0 {
		result.MatchStrengthScore = clampScore(result.MatchStrengthScore + result.PreviousDispositionScore)

		if result.PreviousDispositionScore > 0 {
			for _, reason := range result.PreviousDispositionReasons {
				if strings.TrimSpace(reason) != "" {
					result.EvidenceFor = append(result.EvidenceFor, reason)
				}
			}
		} else {
			for _, reason := range result.PreviousDispositionReasons {
				if strings.TrimSpace(reason) != "" {
					result.EvidenceAgainst = append(result.EvidenceAgainst, reason)
				}
			}
		}
	}

	// 2) Relationship support and linked-doc support reinforce corroboration.
	positiveRelationshipBoost := result.RelationshipSupportScore + result.OfficialDocLinkScore + result.ProgramContextScore
	if positiveRelationshipBoost > 0 {
		result.MatchStrengthScore = clampScore(result.MatchStrengthScore + positiveRelationshipBoost)

		// Relationship/doc support also increases confidence that the case is reviewable.
		result.DataSufficiencyScore = clampScore(result.DataSufficiencyScore + minInt(positiveRelationshipBoost, 12))

		if len(result.RelationshipReasons) > 0 {
			for _, reason := range result.RelationshipReasons {
				if strings.TrimSpace(reason) != "" {
					result.EvidenceFor = append(result.EvidenceFor, reason)
				}
			}
		} else {
			if result.RelationshipSupportScore > 0 {
				result.EvidenceFor = append(result.EvidenceFor, "ofac relationship support")
			}
			if result.OfficialDocLinkScore > 0 {
				result.EvidenceFor = append(result.EvidenceFor, "linked official document support")
			}
			if result.ProgramContextScore > 0 {
				result.EvidenceFor = append(result.EvidenceFor, "program context support")
			}
		}
	}

	// 3) Relationship conflict should pull the case away from auto-escalate.
	if result.RelationshipConflictPenalty > 0 {
		result.MatchStrengthScore = clampScore(result.MatchStrengthScore - result.RelationshipConflictPenalty)

		// Penalize sufficiency modestly so relationship conflict keeps gray-zone cases conservative.
		result.DataSufficiencyScore = clampScore(result.DataSufficiencyScore - minInt(result.RelationshipConflictPenalty, 10))

		result.EvidenceAgainst = append(result.EvidenceAgainst, "relationship conflict")
		if !hasAny(result.Contradictions, "relationship_conflict") {
			result.Contradictions = append(result.Contradictions, "relationship_conflict")
		}
	}
}

func decide(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures, result *Result) {
	hasHardConflict := hasAny(result.Contradictions, "entity_type_conflict", "identifier_conflict", "relationship_conflict")
	hasAnyConflict := len(result.Contradictions) > 0

	switch {
	case result.MatchStrengthScore >= 70 &&
		result.DataSufficiencyScore >= 50 &&
		!hasAnyConflict &&
		hasNonNameCorroboration(result):
		result.DecisionLabel = DecisionEscalate
		result.DecisionReason = "Strong corroborated match with sufficient data and no contradictions."
		result.NextStep = "Escalate for analyst review."

	case hasHardConflict:
		if result.MatchStrengthScore <= 30 {
			result.DecisionLabel = DecisionClose
			result.DecisionReason = "Hard contradiction present and overall match strength is weak."
			result.NextStep = "Close as likely false positive and document contradiction."
		} else {
			result.DecisionLabel = DecisionInvestigateNextStep
			result.DecisionReason = "Hard contradiction present; do not escalate without further verification."
			result.NextStep = "Investigate contradiction and gather additional corroboration."
		}

	case organizationGrayZone(alert, fx, result):
		result.DecisionLabel = DecisionInvestigateNextStep
		result.DecisionReason = "Organization name overlap exists, but corroboration is too weak for escalation and too meaningful for immediate closure."
		result.NextStep = "Investigate organization identifiers, registration details, and geography before closing."

	case result.MatchStrengthScore <= 30:
		result.DecisionLabel = DecisionClose
		result.DecisionReason = "Low match strength with insufficient evidence for escalation."
		result.NextStep = "Close as likely false positive."

	case result.DataSufficiencyScore < 50 || hasAnyConflict:
		result.DecisionLabel = DecisionInvestigateNextStep
		result.DecisionReason = "Case is plausible but under-evidenced or contradictory."
		result.NextStep = "Investigate next step and gather missing data."

	default:
		result.DecisionLabel = DecisionInvestigateNextStep
		result.DecisionReason = "Case falls in the gray zone and requires conservative follow-up."
		result.NextStep = "Investigate next step and gather corroborating evidence."
	}

	// Calibrate weak but contradiction-heavy sparse cases away from auto-close.
	ApplySparseContradictionPolicy(result)

	result.EvidenceFor = uniqueStrings(result.EvidenceFor)
	result.EvidenceAgainst = uniqueStrings(result.EvidenceAgainst)
	result.MissingInformation = uniqueStrings(result.MissingInformation)
	result.PreviousDispositionReasons = uniqueStrings(result.PreviousDispositionReasons)
	result.RelationshipReasons = uniqueStrings(result.RelationshipReasons)

	sort.Strings(result.Contradictions)
	sort.Strings(result.EvidenceFor)
	sort.Strings(result.EvidenceAgainst)
	sort.Strings(result.MissingInformation)
	sort.Strings(result.PreviousDispositionReasons)
	sort.Strings(result.RelationshipReasons)
}

func organizationGrayZone(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures, result *Result) bool {
	if alert == nil || fx == nil || result == nil {
		return false
	}
	if alert.Kind != alerts.AlertKindOrganizationOnboarding {
		return false
	}
	if hasAny(result.Contradictions, "entity_type_conflict", "identifier_conflict", "relationship_conflict") {
		return false
	}
	if hasNonNameCorroboration(result) {
		return false
	}

	return result.NameMatchScore >= 4 &&
		result.MatchStrengthScore <= 30 &&
		result.DataSufficiencyScore >= 30
}

func clampScore(v int) int {
	switch {
	case v < 0:
		return 0
	case v > 100:
		return 100
	default:
		return v
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
