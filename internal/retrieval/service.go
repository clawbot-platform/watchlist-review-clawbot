package retrieval

import (
	"context"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
}

func (s *Service) BuildPromptContext(
	ctx context.Context,
	tenantID string,
	alert *alerts.CanonicalAlert,
	score *scoring.Result,
) *PromptContext {
	if s == nil || s.client == nil || strings.TrimSpace(s.client.BaseURL()) == "" {
		return &PromptContext{Warnings: []string{"retrieval_not_configured"}}
	}
	if alert == nil || score == nil {
		return &PromptContext{Warnings: []string{"retrieval_input_unavailable"}}
	}

	queryText := buildQueryText(alert, score)
	resp, err := s.client.Search(ctx, Query{
		TenantID: tenantID,
		CaseID:   alert.Metadata.CaseID,
		AlertID:  alert.Metadata.AlertID,
		Text:     queryText,
		TopK:     4,
		Tags: []string{
			string(alert.Kind),
			strings.ToLower(string(alert.MatchedParty.EntityType)),
			strings.ToLower(alert.MatchedParty.Program),
		},
	})
	if err != nil {
		return &PromptContext{
			QueryText: queryText,
			Warnings:  []string{fmt.Sprintf("retrieval_failed: %v", err)},
		}
	}

	return &PromptContext{
		QueryText: queryText,
		Snippets:  limitSnippets(resp.Snippets, 4),
	}
}

func buildQueryText(alert *alerts.CanonicalAlert, score *scoring.Result) string {
	var parts []string
	parts = append(parts, string(alert.Kind))
	if name := strings.TrimSpace(alert.ScreenedParty.Name.FullName); name != "" {
		parts = append(parts, "screened="+name)
	}
	if name := strings.TrimSpace(alert.MatchedParty.Name.FullName); name != "" {
		parts = append(parts, "matched="+name)
	}
	if program := strings.TrimSpace(alert.MatchedParty.Program); program != "" {
		parts = append(parts, "program="+program)
	}
	if score.DecisionLabel != "" {
		parts = append(parts, "decision="+score.DecisionLabel)
	}
	for _, item := range score.EvidenceFor {
		if item = strings.TrimSpace(item); item != "" {
			parts = append(parts, item)
		}
	}
	return strings.Join(parts, " | ")
}

func limitSnippets(in []Snippet, limit int) []Snippet {
	if limit <= 0 || len(in) <= limit {
		return in
	}
	return append([]Snippet(nil), in[:limit]...)
}
