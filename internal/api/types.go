package api

import (
	"encoding/json"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

type ReviewOptions struct {
	Explain bool   `json:"explain,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

type ReviewRequest struct {
	TenantID     string          `json:"tenant_id"`
	CaseID       string          `json:"case_id,omitempty"`
	SourceSystem string          `json:"source_system"`
	RawAlert     json.RawMessage `json:"raw_alert"`
	Options      ReviewOptions   `json:"options,omitempty"`
}

type FeedbackRequest struct {
	TenantID          string                     `json:"tenant_id"`
	CaseID            string                     `json:"case_id"`
	AlertID           string                     `json:"alert_id,omitempty"`
	AnalystID         string                     `json:"analyst_id,omitempty"`
	SystemDecision    string                     `json:"system_decision,omitempty"`
	DecisionAgreement feedback.DecisionAgreement `json:"decision_agreement"`
	CorrectedLabel    string                     `json:"corrected_label,omitempty"`
	NoteRating        int                        `json:"note_rating,omitempty"`
	OutcomeRating     int                        `json:"outcome_rating,omitempty"`
	Comment           string                     `json:"comment,omitempty"`
	Tags              []string                   `json:"tags,omitempty"`
}

type FeedbackResponse struct {
	Status       string                  `json:"status"`
	Feedback     feedback.AnalystFeedback `json:"feedback"`
	ArtifactRefs []artifacts.ArtifactRef `json:"artifact_refs,omitempty"`
	Warnings     []string                `json:"warnings,omitempty"`
}

type IdentityTraceRefs struct {
	DecisionTraceID string `json:"decision_trace_id,omitempty"`
	ExplanationID   string `json:"explanation_id,omitempty"`
	ScreeningID     string `json:"screening_id,omitempty"`
}

type ReviewResponse struct {
	Status               string                  `json:"status"`
	CaseID               string                  `json:"case_id,omitempty"`
	AlertID              string                  `json:"alert_id,omitempty"`
	Warnings             []string                `json:"warnings,omitempty"`
	IdentityTraceRefs    IdentityTraceRefs       `json:"identity_trace_refs,omitempty"`

	MatchStrengthScore   int                     `json:"match_strength_score,omitempty"`
	DataSufficiencyScore int                     `json:"data_sufficiency_score,omitempty"`
	Contradictions       []string                `json:"contradictions,omitempty"`
	DecisionLabel        string                  `json:"decision_label,omitempty"`
	DecisionReason       string                  `json:"decision_reason,omitempty"`
	EvidenceFor          []string                `json:"evidence_for,omitempty"`
	EvidenceAgainst      []string                `json:"evidence_against,omitempty"`
	MissingInformation   []string                `json:"missing_information,omitempty"`
	NextStep             string                  `json:"next_step,omitempty"`

	AnalystNote          *notes.AnalystNote      `json:"analyst_note,omitempty"`
	ArtifactRefs         []artifacts.ArtifactRef `json:"artifact_refs,omitempty"`
	ArtifactWarnings     []string                `json:"artifact_warnings,omitempty"`
	ReviewContext        any                     `json:"review_context,omitempty"`
}
