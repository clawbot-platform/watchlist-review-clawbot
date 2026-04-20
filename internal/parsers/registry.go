package parsers

import (
	"context"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

type Parser interface {
	SourceSystem() string
	Parse(context.Context, []byte) (*alerts.CanonicalAlert, error)
}

type Registry struct {
	parsers map[string]Parser
}

func NewRegistry(parsers ...Parser) (*Registry, error) {
	r := &Registry{
		parsers: map[string]Parser{},
	}
	for _, p := range parsers {
		if err := r.Register(p); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Registry) Register(p Parser) error {
	if p == nil {
		return fmt.Errorf("parser is required")
	}

	key := strings.TrimSpace(strings.ToLower(p.SourceSystem()))
	if key == "" {
		return fmt.Errorf("parser source system is required")
	}
	if _, exists := r.parsers[key]; exists {
		return fmt.Errorf("parser already registered for %q", key)
	}

	r.parsers[key] = p
	return nil
}

func (r *Registry) Parse(ctx context.Context, sourceSystem string, raw []byte) (*alerts.CanonicalAlert, error) {
	if r == nil {
		return nil, fmt.Errorf("parser registry is nil")
	}

	key := strings.TrimSpace(strings.ToLower(sourceSystem))
	p, ok := r.parsers[key]
	if !ok {
		return nil, fmt.Errorf("no parser registered for %q", sourceSystem)
	}

	return p.Parse(ctx, raw)
}