package scoring

import "github.com/clawbot-platform/watchlist-review-clawbot/internal/ofacgraph"

func ApplyRelationshipEvidence(result *Result, ev ofacgraph.RelationshipEvidence) {
	if result == nil {
		return
	}

	// Starter-only scoring hook. Tune inside the main repo against eval data.
	result.RelationshipSupportScore = ev.RelationshipSupportScore
	result.RelationshipConflictPenalty = ev.RelationshipConflictPenalty
	result.OfficialDocLinkScore = ev.OfficialDocLinkScore
	result.ProgramContextScore = ev.ProgramContextScore

	if ev.RelationshipSupportScore > 0 {
		result.MatchStrengthScore += ev.RelationshipSupportScore
	}
	if ev.OfficialDocLinkScore > 0 {
		result.MatchStrengthScore += ev.OfficialDocLinkScore
	}
	if ev.ProgramContextScore > 0 {
		result.SupportingContextScore += ev.ProgramContextScore
	}
	if ev.RelationshipConflictPenalty > 0 {
		result.MatchStrengthScore -= ev.RelationshipConflictPenalty
		result.Contradictions = append(result.Contradictions, "relationship_context_conflict")
	}

	result.RelationshipReasons = append(result.RelationshipReasons, ev.Reasons...)
}
