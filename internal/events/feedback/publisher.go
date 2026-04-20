package feedbackevents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
	"github.com/nats-io/nats.go"
)

const DefaultSubject = "clawbot.watchlist.review.feedback.created.v1"

type NATSPublisher struct {
	conn    *nats.Conn
	subject string
}

func NewNATSPublisher(url string, subject string) (*NATSPublisher, error) {
	conn, err := nats.Connect(strings.TrimSpace(url))
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	if strings.TrimSpace(subject) == "" {
		subject = DefaultSubject
	}
	return &NATSPublisher{
		conn:    conn,
		subject: strings.TrimSpace(subject),
	}, nil
}

func (p *NATSPublisher) PublishFeedbackCreated(_ context.Context, event feedback.FeedbackCreatedEvent) error {
	if p == nil || p.conn == nil {
		return fmt.Errorf("nats publisher is not configured")
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal feedback event: %w", err)
	}
	if err := p.conn.Publish(p.subject, payload); err != nil {
		return fmt.Errorf("publish feedback event: %w", err)
	}
	return p.conn.Flush()
}
