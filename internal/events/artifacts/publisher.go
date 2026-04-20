package artifactevents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
	"github.com/nats-io/nats.go"
)

const DefaultSubject = "clawbot.watchlist.review.artifact.created.v1"

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

func (p *NATSPublisher) PublishArtifactCreated(_ context.Context, event artifacts.ArtifactCreatedEvent) error {
	if p == nil || p.conn == nil {
		return fmt.Errorf("nats publisher is not configured")
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal artifact event: %w", err)
	}
	if err := p.conn.Publish(p.subject, payload); err != nil {
		return fmt.Errorf("publish artifact event: %w", err)
	}
	return p.conn.Flush()
}

func (p *NATSPublisher) Close() {
	if p != nil && p.conn != nil {
		p.conn.Close()
	}
}
