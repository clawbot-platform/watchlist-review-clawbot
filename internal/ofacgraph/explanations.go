package ofacgraph

func BuildExplanation(ev RelationshipEvidence) map[string]any {
    return map[string]any{
        "matched_party_id":               ev.MatchedPartyID,
        "direct_relationship_support":    ev.DirectRelationshipSupport,
        "relationship_support_score":     ev.RelationshipSupportScore,
        "relationship_conflict_penalty":  ev.RelationshipConflictPenalty,
        "official_doc_link_score":        ev.OfficialDocLinkScore,
        "program_context_score":          ev.ProgramContextScore,
        "neighbor_count":                 ev.NeighborCount,
        "reasons":                        ev.Reasons,
        "conflicts":                      ev.Conflicts,
        "linked_documents":               ev.LinkedDocuments,
        "program_context":                ev.ProgramContext,
    }
}
