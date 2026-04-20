package identity

type SourceRef struct {
	SourceSystem   string `json:"source_system"`
	SourceRecordID string `json:"source_record_id"`
}

type CompareRequest struct {
	TenantID string    `json:"tenant_id"`
	Left     SourceRef `json:"left"`
	Right    SourceRef `json:"right"`
	Explain  bool      `json:"explain"`
}

type ExplanationSourceRef struct {
	SourceSystem   string `json:"source_system"`
	SourceRecordID string `json:"source_record_id"`
}

type Explanation struct {
	ExplanationID string                 `json:"explanation_id"`
	Summary       string                 `json:"summary"`
	Why           []string               `json:"why,omitempty"`
	WhyNot        []string               `json:"why_not,omitempty"`
	How           []string               `json:"how,omitempty"`
	SourceRefs    []ExplanationSourceRef `json:"source_refs,omitempty"`
}

type CompareResponse struct {
	Disposition     string       `json:"disposition"`
	ConfidenceBand  string       `json:"confidence_band,omitempty"`
	Explanation     *Explanation `json:"explanation,omitempty"`
	DecisionTraceID string       `json:"decision_trace_id"`
}
