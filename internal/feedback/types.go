package feedback

import (
	"context"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
)

type DecisionAgreement string

const (
	DecisionAgreementAgree         DecisionAgreement = "agree"
	DecisionAgreementDisagree      DecisionAgreement = "disagree"
	DecisionAgreementPartial       DecisionAgreement = "partial"
)

type AnalystFeedback struct {
	FeedbackID         string            `json:"feedback_id"`
	TenantID           string            `json:"tenant_id"`
	CaseID             string            `json:"case_id"`
	AlertID            string            `json:"alert_id,omitempty"`
	CorrelationID      string            `json:"correlation_id,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	AnalystID          string            `json:"analyst_id,omitempty"`
	SystemDecision     string            `json:"system_decision,omitempty"`
	DecisionAgreement  DecisionAgreement `json:"decision_agreement"`
	CorrectedLabel     string            `json:"corrected_label,omitempty"`
	NoteRating         int               `json:"note_rating,omitempty"`
	OutcomeRating      int               `json:"outcome_rating,omitempty"`
	Comment            string            `json:"comment,omitempty"`
	Tags               []string          `json:"tags,omitempty"`
	DerivedSignals     []string          `json:"derived_signals,omitempty"`
}

type CreateInput struct {
	TenantID          string
	CaseID            string
	AlertID           string
	CorrelationID     string
	AnalystID         string
	SystemDecision    string
	DecisionAgreement DecisionAgreement
	CorrectedLabel    string
	NoteRating        int
	OutcomeRating     int
	Comment           string
	Tags              []string
}

type CreateResult struct {
	Feedback    AnalystFeedback          `json:"feedback"`
	ArtifactRefs []artifacts.ArtifactRef `json:"artifact_refs,omitempty"`
	Warnings    []string                 `json:"warnings,omitempty"`
}

type EventPublisher interface {
	PublishFeedbackCreated(ctx context.Context, event FeedbackCreatedEvent) error
}

type FeedbackCreatedEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	OccurredAt    time.Time              `json:"occurred_at"`
	TenantID      string                 `json:"tenant_id"`
	CaseID        string                 `json:"case_id"`
	AlertID       string                 `json:"alert_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	SystemDecision string                `json:"system_decision,omitempty"`
	Feedback      AnalystFeedback        `json:"feedback"`
	ArtifactRef   *artifacts.ArtifactRef `json:"artifact_ref,omitempty"`
}
