package evalstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type FeedbackRow struct {
	FeedbackID        string
	EventID           string
	EventType         string
	OccurredAt        time.Time
	TenantID          string
	CaseID            string
	AlertID           string
	CorrelationID     string
	AnalystID         string
	SystemDecision    string
	DecisionAgreement string
	CorrectedLabel    string
	NoteRating        int
	OutcomeRating     int
	Comment           string
	Tags              []string
	DerivedSignals    []string
	ArtifactID        string
	ArtifactKind      string
	ArtifactPath      string
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", strings.TrimSpace(dsn))
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("postgres store is not configured")
	}
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) EnsureSchema(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("postgres store is not configured")
	}
	_, err := s.db.ExecContext(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("ensure eval schema: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpsertFeedback(ctx context.Context, row FeedbackRow) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("postgres store is not configured")
	}
	tagsJSON, err := json.Marshal(row.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	signalsJSON, err := json.Marshal(row.DerivedSignals)
	if err != nil {
		return fmt.Errorf("marshal derived signals: %w", err)
	}

	const q = `
INSERT INTO review_feedback_signals (
  feedback_id, event_id, event_type, occurred_at, tenant_id, case_id, alert_id, correlation_id,
  analyst_id, system_decision, decision_agreement, corrected_label, note_rating, outcome_rating,
  comment, tags, derived_signals, artifact_id, artifact_kind, artifact_path
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,
  $9,$10,$11,$12,$13,$14,
  $15,$16::jsonb,$17::jsonb,$18,$19,$20
)
ON CONFLICT (feedback_id) DO UPDATE SET
  event_id = EXCLUDED.event_id,
  event_type = EXCLUDED.event_type,
  occurred_at = EXCLUDED.occurred_at,
  tenant_id = EXCLUDED.tenant_id,
  case_id = EXCLUDED.case_id,
  alert_id = EXCLUDED.alert_id,
  correlation_id = EXCLUDED.correlation_id,
  analyst_id = EXCLUDED.analyst_id,
  system_decision = EXCLUDED.system_decision,
  decision_agreement = EXCLUDED.decision_agreement,
  corrected_label = EXCLUDED.corrected_label,
  note_rating = EXCLUDED.note_rating,
  outcome_rating = EXCLUDED.outcome_rating,
  comment = EXCLUDED.comment,
  tags = EXCLUDED.tags,
  derived_signals = EXCLUDED.derived_signals,
  artifact_id = EXCLUDED.artifact_id,
  artifact_kind = EXCLUDED.artifact_kind,
  artifact_path = EXCLUDED.artifact_path
`
	_, err = s.db.ExecContext(ctx, q,
		row.FeedbackID, row.EventID, row.EventType, row.OccurredAt, row.TenantID, row.CaseID, row.AlertID, row.CorrelationID,
		row.AnalystID, row.SystemDecision, row.DecisionAgreement, row.CorrectedLabel, row.NoteRating, row.OutcomeRating,
		row.Comment, string(tagsJSON), string(signalsJSON), row.ArtifactID, row.ArtifactKind, row.ArtifactPath,
	)
	if err != nil {
		return fmt.Errorf("upsert review_feedback_signals: %w", err)
	}
	return nil
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS review_feedback_signals (
  feedback_id text PRIMARY KEY,
  event_id text NOT NULL,
  event_type text NOT NULL,
  occurred_at timestamptz NOT NULL,
  tenant_id text NOT NULL,
  case_id text NOT NULL,
  alert_id text,
  correlation_id text,
  analyst_id text,
  system_decision text,
  decision_agreement text NOT NULL,
  corrected_label text,
  note_rating integer NOT NULL DEFAULT 0,
  outcome_rating integer NOT NULL DEFAULT 0,
  comment text,
  tags jsonb NOT NULL DEFAULT '[]'::jsonb,
  derived_signals jsonb NOT NULL DEFAULT '[]'::jsonb,
  artifact_id text,
  artifact_kind text,
  artifact_path text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_review_feedback_signals_case
  ON review_feedback_signals (tenant_id, case_id);

CREATE INDEX IF NOT EXISTS idx_review_feedback_signals_signals
  ON review_feedback_signals USING gin (derived_signals);

CREATE INDEX IF NOT EXISTS idx_review_feedback_signals_tags
  ON review_feedback_signals USING gin (tags);
`
