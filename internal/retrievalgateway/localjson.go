package retrievalgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

type LocalJSONSearcher struct {
	snippets []retrieval.Snippet
}

func NewLocalJSONSearcher(path string) (*LocalJSONSearcher, error) {
	raw, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return nil, fmt.Errorf("read local retrieval snippets: %w", err)
	}
	var snippets []retrieval.Snippet
	if err := json.Unmarshal(raw, &snippets); err != nil {
		return nil, fmt.Errorf("decode local retrieval snippets: %w", err)
	}
	return &LocalJSONSearcher{snippets: snippets}, nil
}

func (s *LocalJSONSearcher) Search(_ context.Context, query retrieval.Query) (retrieval.SearchResponse, error) {
	if s == nil {
		return retrieval.SearchResponse{}, fmt.Errorf("local json searcher is not configured")
	}
	limit := query.TopK
	if limit <= 0 {
		limit = 4
	}

	terms := tokenize(query.Text)
	tagTerms := make([]string, 0, len(query.Tags))
	for _, tag := range query.Tags {
		tagTerms = append(tagTerms, strings.ToLower(strings.TrimSpace(tag)))
	}

	type ranked struct {
		snippet retrieval.Snippet
		score   float64
	}
	var rankedSnippets []ranked
	for _, snip := range s.snippets {
		score := scoreSnippet(snip, terms, tagTerms)
		if score <= 0 {
			continue
		}
		copySnip := snip
		copySnip.Score = score
		rankedSnippets = append(rankedSnippets, ranked{snippet: copySnip, score: score})
	}

	sort.Slice(rankedSnippets, func(i, j int) bool {
		if rankedSnippets[i].score == rankedSnippets[j].score {
			return rankedSnippets[i].snippet.SnippetID < rankedSnippets[j].snippet.SnippetID
		}
		return rankedSnippets[i].score > rankedSnippets[j].score
	})

	out := make([]retrieval.Snippet, 0, limit)
	for _, item := range rankedSnippets {
		out = append(out, item.snippet)
		if len(out) >= limit {
			break
		}
	}
	return retrieval.SearchResponse{Snippets: out}, nil
}

func scoreSnippet(s retrieval.Snippet, terms []string, tags []string) float64 {
	text := strings.ToLower(strings.Join([]string{s.Title, s.Text, s.Source}, " "))
	score := 0.0
	for _, term := range terms {
		if term == "" {
			continue
		}
		if strings.Contains(text, term) {
			score += 1.0
		}
	}
	for _, needle := range tags {
		if needle == "" {
			continue
		}
		for _, tag := range s.Tags {
			if strings.EqualFold(strings.TrimSpace(tag), needle) {
				score += 1.5
				break
			}
		}
	}
	return score
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	replacer := strings.NewReplacer("|", " ", "=", " ", ",", " ", ".", " ", ";", " ", ":", " ", "/", " ", "-", " ")
	text = replacer.Replace(text)
	fields := strings.Fields(text)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 2 {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}
