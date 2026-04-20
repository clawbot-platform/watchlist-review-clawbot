package retrieval

type Query struct {
	TenantID string   `json:"tenant_id,omitempty"`
	CaseID   string   `json:"case_id,omitempty"`
	AlertID  string   `json:"alert_id,omitempty"`
	Text     string   `json:"text"`
	TopK     int      `json:"top_k,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type Snippet struct {
	SnippetID string   `json:"snippet_id,omitempty"`
	Source    string   `json:"source,omitempty"`
	Title     string   `json:"title,omitempty"`
	Text      string   `json:"text"`
	Score     float64  `json:"score,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type SearchResponse struct {
	Snippets []Snippet `json:"snippets"`
}

type PromptContext struct {
	QueryText string    `json:"query_text,omitempty"`
	Snippets  []Snippet `json:"snippets,omitempty"`
	Warnings  []string  `json:"warnings,omitempty"`
}
