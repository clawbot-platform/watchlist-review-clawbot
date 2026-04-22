package api

type RelationshipEvidence struct {
    MatchedPartyID              string         `json:"matched_party_id,omitempty"`
    DirectRelationshipSupport   bool           `json:"direct_relationship_support,omitempty"`
    RelationshipSupportScore    int            `json:"relationship_support_score,omitempty"`
    RelationshipConflictPenalty int            `json:"relationship_conflict_penalty,omitempty"`
    OfficialDocLinkScore        int            `json:"official_doc_link_score,omitempty"`
    ProgramContextScore         int            `json:"program_context_score,omitempty"`
    NeighborCount               int            `json:"neighbor_count,omitempty"`
    Reasons                     []string       `json:"reasons,omitempty"`
    Conflicts                   []string       `json:"conflicts,omitempty"`
    LinkedDocuments             []string       `json:"linked_documents,omitempty"`
    ProgramContext              []string       `json:"program_context,omitempty"`
    Explanation                 map[string]any `json:"explanation,omitempty"`
}
