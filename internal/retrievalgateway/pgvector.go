package retrievalgateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
	_ "github.com/lib/pq"
)

type PgvectorSearcher struct {
	db       *sql.DB
	embedder Embedder
	table    string
}

func NewPgvectorSearcher(dsn string, table string, embedder Embedder) (*PgvectorSearcher, error) {
	db, err := sql.Open("postgres", strings.TrimSpace(dsn))
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if strings.TrimSpace(table) == "" {
		table = "rag_documents"
	}
	return &PgvectorSearcher{
		db:       db,
		embedder: embedder,
		table:    table,
	}, nil
}

func (s *PgvectorSearcher) Search(ctx context.Context, query retrieval.Query) (retrieval.SearchResponse, error) {
	if s == nil || s.db == nil {
		return retrieval.SearchResponse{}, fmt.Errorf("pgvector searcher is not configured")
	}
	if s.embedder == nil {
		return retrieval.SearchResponse{}, fmt.Errorf("pgvector embedder is not configured")
	}
	vector, err := s.embedder.Embed(ctx, query.Text)
	if err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("embed query: %w", err)
	}
	vectorLiteral := toVectorLiteral(vector)
	topK := query.TopK
	if topK <= 0 {
		topK = 4
	}

	sqlQuery := fmt.Sprintf(`
SELECT id, coalesce(source,''), coalesce(title,''), coalesce(body,''), coalesce(tags, '[]'::jsonb), 1 - (embedding <=> $1::vector) AS score
FROM %s
WHERE ($2 = '' OR tenant_id = $2)
ORDER BY embedding <=> $1::vector
LIMIT $3
`, s.table)

	rows, err := s.db.QueryContext(ctx, sqlQuery, vectorLiteral, strings.TrimSpace(query.TenantID), topK)
	if err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("query pgvector documents: %w", err)
	}
	defer rows.Close()

	var snippets []retrieval.Snippet
	for rows.Next() {
		var id, source, title, text string
		var tagsRaw []byte
		var score float64
		if err := rows.Scan(&id, &source, &title, &text, &tagsRaw, &score); err != nil {
			return retrieval.SearchResponse{}, fmt.Errorf("scan pgvector row: %w", err)
		}
		var tags []string
		_ = json.Unmarshal(tagsRaw, &tags)
		snippets = append(snippets, retrieval.Snippet{
			SnippetID: id,
			Source:    source,
			Title:     title,
			Text:      text,
			Score:     score,
			Tags:      tags,
		})
	}
	if err := rows.Err(); err != nil {
		return retrieval.SearchResponse{}, fmt.Errorf("iterate pgvector rows: %w", err)
	}
	return retrieval.SearchResponse{Snippets: snippets}, nil
}

func toVectorLiteral(vector []float64) string {
	parts := make([]string, 0, len(vector))
	for _, value := range vector {
		parts = append(parts, fmt.Sprintf("%f", value))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
