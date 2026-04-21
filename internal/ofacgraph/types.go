package ofacgraph

type RelationshipEvidence struct {
	MatchedPartyID              string   `json:"matched_party_id"`
	DirectRelationshipSupport   bool     `json:"direct_relationship_support"`
	RelationshipSupportScore    int      `json:"relationship_support_score"`
	RelationshipConflictPenalty int      `json:"relationship_conflict_penalty"`
	OfficialDocLinkScore        int      `json:"official_doc_link_score"`
	ProgramContextScore         int      `json:"program_context_score"`
	NeighborCount               int      `json:"neighbor_count"`
	Reasons                     []string `json:"reasons,omitempty"`
	Conflicts                   []string `json:"conflicts,omitempty"`
	LinkedDocuments             []string `json:"linked_documents,omitempty"`
	ProgramContext              []string `json:"program_context,omitempty"`
}
