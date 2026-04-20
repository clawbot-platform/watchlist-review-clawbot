package notes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OllamaGraniteGenerator struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOllamaGraniteGenerator(baseURL, model string, timeout time.Duration) *OllamaGraniteGenerator {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if strings.TrimSpace(model) == "" {
		model = "granite3.3:8b"
	}
	return &OllamaGraniteGenerator{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		model:   strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format any    `json:"format,omitempty"`
}

type ollamaGenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type generatedNote struct {
	Note                      string   `json:"note"`
	EvidenceSummary           []string `json:"evidence_summary"`
	MissingInformationSummary []string `json:"missing_information_summary"`
	NextStepRationale         string   `json:"next_step_rationale"`
}

func (g *OllamaGraniteGenerator) Generate(ctx context.Context, input PromptInput) (*AnalystNote, error) {
	if g == nil || g.baseURL == "" {
		return nil, fmt.Errorf("ollama base url is required")
	}

	prompt, err := BuildPrompt(input)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(ollamaGenerateRequest{
		Model:  g.model,
		Prompt: prompt,
		Stream: false,
		Format: noteSchema(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute ollama request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("ollama status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var decoded ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode ollama response: %w", err)
	}

	var generated generatedNote
	if err := json.Unmarshal([]byte(decoded.Response), &generated); err != nil {
		return nil, fmt.Errorf("decode generated note json: %w", err)
	}

	return &AnalystNote{
		Status:                    StatusGenerated,
		Model:                     decoded.Model,
		PromptVersion:             promptVersion,
		Note:                      strings.TrimSpace(generated.Note),
		EvidenceSummary:           trimStringSlice(generated.EvidenceSummary),
		MissingInformationSummary: trimStringSlice(generated.MissingInformationSummary),
		NextStepRationale:         strings.TrimSpace(generated.NextStepRationale),
	}, nil
}

func noteSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"note": map[string]any{"type": "string"},
			"evidence_summary": map[string]any{
				"type": "array",
				"items": map[string]any{"type": "string"},
			},
			"missing_information_summary": map[string]any{
				"type": "array",
				"items": map[string]any{"type": "string"},
			},
			"next_step_rationale": map[string]any{"type": "string"},
		},
		"required": []string{"note", "evidence_summary", "missing_information_summary", "next_step_rationale"},
	}
}

func trimStringSlice(in []string) []string {
	var out []string
	for _, v := range in {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
