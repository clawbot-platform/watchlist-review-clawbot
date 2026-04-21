package history

import "time"

type DispositionLabel string

const (
    DispositionEscalate             DispositionLabel = "escalate"
    DispositionInvestigateNextStep  DispositionLabel = "investigate_next_step"
    DispositionClose                DispositionLabel = "close"
    DispositionReviewPending        DispositionLabel = "review_pending"
)

type CaseDisposition struct {
    TenantID                 string           `json:"tenant_id"`
    CaseID                   string           `json:"case_id"`
    AlertID                  string           `json:"alert_id"`
    ScreenedEntityFingerprint string          `json:"screened_entity_fingerprint"`
    MatchedListUID           string           `json:"matched_list_uid"`
    MatchedProgram           string           `json:"matched_program"`
    DecisionLabel            DispositionLabel `json:"decision_label"`
    ContradictionPattern     string           `json:"contradiction_pattern,omitempty"`
    ReviewedAt               time.Time        `json:"reviewed_at"`
    Source                   string           `json:"source,omitempty"`
}

type LookupRequest struct {
    TenantID                  string `json:"tenant_id"`
    ScreenedEntityFingerprint string `json:"screened_entity_fingerprint"`
    MatchedListUID            string `json:"matched_list_uid"`
    MatchedProgram            string `json:"matched_program,omitempty"`
}

type LookupResult struct {
    SamePairCases              []CaseDisposition `json:"same_pair_cases,omitempty"`
    SameMatchedProfileCases    []CaseDisposition `json:"same_matched_profile_cases,omitempty"`
    LatestSamePair             *CaseDisposition  `json:"latest_same_pair,omitempty"`
    EscalateCount              int               `json:"escalate_count"`
    InvestigateCount           int               `json:"investigate_count"`
    CloseCount                 int               `json:"close_count"`
    ReviewPendingCount         int               `json:"review_pending_count"`
    MostRecentReviewedAt       *time.Time        `json:"most_recent_reviewed_at,omitempty"`
    RecurringContradictionKeys []string          `json:"recurring_contradiction_keys,omitempty"`
}

type Evidence struct {
    PreviousDispositionScore   int      `json:"previous_disposition_score"`
    PreviousDispositionReasons []string `json:"previous_disposition_reasons,omitempty"`
    Applied                    bool     `json:"applied"`
    SamePairHistoryFound       bool     `json:"same_pair_history_found"`
}
