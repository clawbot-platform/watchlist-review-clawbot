package feedbackevents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/evalstore"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	conn    *nats.Conn
	subject string
	logger  *slog.Logger
	store   *evalstore.PostgresStore
}

func NewConsumer(natsURL, subject string, logger *slog.Logger, store *evalstore.PostgresStore) (*Consumer, error) {
	conn, err := nats.Connect(strings.TrimSpace(natsURL))
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	if strings.TrimSpace(subject) == "" {
		subject = DefaultSubject
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Consumer{
		conn:    conn,
		subject: strings.TrimSpace(subject),
		logger:  logger,
		store:   store,
	}, nil
}

func (c *Consumer) Close() {
	if c != nil && c.conn != nil {
		c.conn.Close()
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("consumer is not configured")
	}
	if c.store == nil {
		return fmt.Errorf("feedback eval store is not configured")
	}

	_, err := c.conn.Subscribe(c.subject, func(msg *nats.Msg) {
		if err := c.handleMessage(ctx, msg); err != nil {
			c.logger.Error("feedback.consumer.handle_failed", slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe feedback events: %w", err)
	}
	c.logger.Info("feedback.consumer.subscribed", slog.String("subject", c.subject))
	<-ctx.Done()
	return nil
}

func (c *Consumer) handleMessage(ctx context.Context, msg *nats.Msg) error {
	var event feedback.FeedbackCreatedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return fmt.Errorf("decode feedback event: %w", err)
	}
	row := evalstore.FeedbackRow{
		FeedbackID:        event.Feedback.FeedbackID,
		EventID:           event.EventID,
		EventType:         event.EventType,
		OccurredAt:        firstTime(event.OccurredAt, event.Feedback.CreatedAt),
		TenantID:          event.TenantID,
		CaseID:            event.CaseID,
		AlertID:           event.AlertID,
		CorrelationID:     event.CorrelationID,
		AnalystID:         event.Feedback.AnalystID,
		SystemDecision:    firstString(event.SystemDecision, event.Feedback.SystemDecision),
		DecisionAgreement: string(event.Feedback.DecisionAgreement),
		CorrectedLabel:    event.Feedback.CorrectedLabel,
		NoteRating:        event.Feedback.NoteRating,
		OutcomeRating:     event.Feedback.OutcomeRating,
		Comment:           event.Feedback.Comment,
		Tags:              append([]string(nil), event.Feedback.Tags...),
		DerivedSignals:    append([]string(nil), event.Feedback.DerivedSignals...),
	}
	if event.ArtifactRef != nil {
		row.ArtifactID = event.ArtifactRef.ArtifactID
		row.ArtifactKind = string(event.ArtifactRef.Kind)
		row.ArtifactPath = event.ArtifactRef.RelativePath
	}
	return c.store.UpsertFeedback(ctx, row)
}

func firstString(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func firstTime(values ...time.Time) time.Time {
	for _, v := range values {
		if !v.IsZero() {
			return v
		}
	}
	return time.Now().UTC()
}
