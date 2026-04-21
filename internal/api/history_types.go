package api

type HistoryEvidence struct {
    PreviousDispositionScore   int      `json:"previous_disposition_score"`
    PreviousDispositionReasons []string `json:"previous_disposition_reasons,omitempty"`
    SamePairHistoryFound       bool     `json:"same_pair_history_found"`
}

type PreviousDispositionSummary struct {
    LatestDecisionLabel string `json:"latest_decision_label,omitempty"`
    EscalateCount       int    `json:"escalate_count,omitempty"`
    InvestigateCount    int    `json:"investigate_count,omitempty"`
    CloseCount          int    `json:"close_count,omitempty"`
}
