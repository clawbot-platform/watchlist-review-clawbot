package identity

type OFACSubject struct {
	Name        string            `json:"name"`
	DOB         string            `json:"dob,omitempty"`
	Country     string            `json:"country,omitempty"`
	Identifiers map[string]string `json:"identifiers,omitempty"`
	Aliases     []string          `json:"aliases,omitempty"`
	Address     string            `json:"address,omitempty"`
}

type ScreenOFACRequest struct {
	TenantID string      `json:"tenant_id"`
	CaseID   string      `json:"case_id"`
	Subject  OFACSubject `json:"subject"`
}

type OFACCandidate struct {
	DatasetRunID string `json:"dataset_run_id,omitempty"`
	ListKind     string `json:"list_kind,omitempty"`
	ListUID      string `json:"list_uid,omitempty"`
	Name         string `json:"name"`
	MatchedOn    string `json:"matched_on,omitempty"`
	Score        int    `json:"score"`
	NeedsReview  bool   `json:"needs_review"`
}

type ScreenOFACResponse struct {
	ScreeningID     string          `json:"screening_id"`
	Decision        string          `json:"decision"`
	DecisionTraceID string          `json:"decision_trace_id"`
	Candidates      []OFACCandidate `json:"candidates,omitempty"`
}
