package artifacts

import (
	"context"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

type Kind string

const (
	KindReviewOutput   Kind = "review_output"
	KindAnalystNote    Kind = "analyst_note"
	KindAnalystFeedback Kind = "analyst_feedback"
	KindCaseManifest   Kind = "case_manifest"
)

type ArtifactRef struct {
	ArtifactID   string    `json:"artifact_id"`
	Kind         Kind      `json:"kind"`
	ContentType  string    `json:"content_type"`
	RelativePath string    `json:"relative_path"`
	CreatedAt    time.Time `json:"created_at"`
}

type WriteInput struct {
	TenantID      string
	CaseID        string
	AlertID       string
	CorrelationID string
	Kind          Kind
	Payload       any
}

type Store interface {
	WriteJSON(input WriteInput) (ArtifactRef, error)
}

type ManifestWriter interface {
	UpsertCaseManifest(input ManifestInput) (ArtifactRef, error)
}

type EventPublisher interface {
	PublishArtifactCreated(ctx context.Context, event ArtifactCreatedEvent) error
}

type PersistInput struct {
	TenantID      string
	CaseID        string
	AlertID       string
	CorrelationID string
	DecisionLabel string
	ReviewContext any
	AnalystNote   *notes.AnalystNote
}

type ManifestInput struct {
	TenantID      string
	CaseID        string
	CorrelationID string
	NewArtifacts  []ArtifactRef
}

type CaseManifest struct {
	ManifestID string         `json:"manifest_id"`
	TenantID   string         `json:"tenant_id"`
	CaseID     string         `json:"case_id"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Artifacts  []ManifestItem `json:"artifacts"`
}

type ManifestItem struct {
	ArtifactID   string    `json:"artifact_id"`
	Kind         Kind      `json:"kind"`
	ContentType  string    `json:"content_type"`
	RelativePath string    `json:"relative_path"`
	CreatedAt    time.Time `json:"created_at"`
}

type ArtifactCreatedEvent struct {
	EventID        string      `json:"event_id"`
	EventType      string      `json:"event_type"`
	OccurredAt     time.Time   `json:"occurred_at"`
	TenantID       string      `json:"tenant_id"`
	CaseID         string      `json:"case_id"`
	AlertID        string      `json:"alert_id,omitempty"`
	CorrelationID  string      `json:"correlation_id,omitempty"`
	DecisionLabel  string      `json:"decision_label,omitempty"`
	Artifact       ArtifactRef `json:"artifact"`
}
