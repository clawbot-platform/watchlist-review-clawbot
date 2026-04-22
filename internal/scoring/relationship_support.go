package scoring

import "github.com/clawbot-platform/watchlist-review-clawbot/internal/ofacgraph"

func ApplyRelationshipEvidence(result *Result, ev ofacgraph.RelationshipEvidence) {
    if result == nil {
        return
    }

    result.RelationshipSupportScore = ev.RelationshipSupportScore
    result.RelationshipConflictPenalty = ev.RelationshipConflictPenalty
    result.OfficialDocLinkScore = ev.OfficialDocLinkScore
    result.ProgramContextScore = ev.ProgramContextScore

    if ev.RelationshipSupportScore > 0 {
        result.MatchStrengthScore += ev.RelationshipSupportScore
        result.EvidenceFor = append(result.EvidenceFor, "relationship graph support")
    }
    if ev.OfficialDocLinkScore > 0 {
        result.MatchStrengthScore += ev.OfficialDocLinkScore
        result.EvidenceFor = append(result.EvidenceFor, "linked official document support")
    }
    if ev.ProgramContextScore > 0 {
        result.SupportingContextScore += ev.ProgramContextScore
        result.EvidenceFor = append(result.EvidenceFor, "program context support")
    }
    if ev.RelationshipConflictPenalty > 0 {
        result.MatchStrengthScore -= ev.RelationshipConflictPenalty
        result.Contradictions = append(result.Contradictions, "relationship_context_conflict")
        result.EvidenceAgainst = append(result.EvidenceAgainst, "relationship neighborhood conflict")
    }

    result.RelationshipReasons = append(result.RelationshipReasons, ev.Reasons...)
    if len(ev.Conflicts) > 0 {
        result.RelationshipReasons = append(result.RelationshipReasons, ev.Conflicts...)
    }
}
