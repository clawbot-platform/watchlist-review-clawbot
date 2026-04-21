package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
)

type BatchSpec struct {
	Cases []CaseSpec `json:"cases"`
}

type CaseSpec struct {
	Name                 string `json:"name"`
	RequestPath          string `json:"request_path"`
	ExpectDecisionLabel  string `json:"expect_decision_label,omitempty"`
	RequireGeneratedNote bool   `json:"require_generated_note,omitempty"`
	RequireRetrieval     bool   `json:"require_retrieval,omitempty"`
}

type CaseResult struct {
	Name            string              `json:"name"`
	RequestPath     string              `json:"request_path,omitempty"`
	ExpectedCaseID  string              `json:"expected_case_id,omitempty"`
	ExpectedAlertID string              `json:"expected_alert_id,omitempty"`
	ActualCaseID    string              `json:"actual_case_id,omitempty"`
	ActualAlertID   string              `json:"actual_alert_id,omitempty"`
	Passed          bool                `json:"passed"`
	Errors          []string            `json:"errors,omitempty"`
	Warnings        []string            `json:"warnings,omitempty"`
	Response        *api.ReviewResponse `json:"response,omitempty"`
}

type BatchReport struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Passed      bool         `json:"passed"`
	CaseResults []CaseResult `json:"case_results"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type requestEnvelope struct {
	CaseID   string `json:"case_id"`
	RawAlert struct {
		AlertID   string `json:"alert_id"`
		CaseID    string `json:"case_id"`
		AlertType string `json:"alert_type"`
	} `json:"raw_alert"`
}

type requestMeta struct {
	CaseID    string
	AlertID   string
	AlertType string
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func LoadBatchSpec(path string) (BatchSpec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return BatchSpec{}, fmt.Errorf("read batch spec: %w", err)
	}
	var spec BatchSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return BatchSpec{}, fmt.Errorf("decode batch spec: %w", err)
	}
	return spec, nil
}

func (c *Client) Run(ctx context.Context, spec BatchSpec) (BatchReport, error) {
	report := BatchReport{
		GeneratedAt: time.Now().UTC(),
		Passed:      true,
		CaseResults: make([]CaseResult, 0, len(spec.Cases)),
	}

	for i := range spec.Cases {
		testCase := spec.Cases[i] // bind by value explicitly
		result := CaseResult{
			Name:        testCase.Name,
			RequestPath: testCase.RequestPath,
		}

		reqBody, err := os.ReadFile(testCase.RequestPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("read request: %v", err))
			result.Passed = false
			report.Passed = false
			report.CaseResults = append(report.CaseResults, result)
			continue
		}

		meta, err := extractRequestMeta(reqBody)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("parse request metadata: %v", err))
			result.Passed = false
			report.Passed = false
			report.CaseResults = append(report.CaseResults, result)
			continue
		}
		result.ExpectedCaseID = meta.CaseID
		result.ExpectedAlertID = meta.AlertID

		resp, err := c.callReview(ctx, reqBody)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			result.Passed = false
			report.Passed = false
			report.CaseResults = append(report.CaseResults, result)
			continue
		}

		// Deep-copy before storing so report data cannot be affected by accidental reuse.
		result.Response = cloneReviewResponse(resp)
		if result.Response != nil {
			result.ActualCaseID = result.Response.CaseID
			result.ActualAlertID = result.Response.AlertID
		}

		if mismatch := responseMismatch(meta, result.Response); mismatch != "" {
			result.Errors = append(result.Errors, mismatch)
			result.Passed = false
			report.Passed = false
			report.CaseResults = append(report.CaseResults, result)
			continue
		}

		result.Errors, result.Warnings = EvaluateCase(testCase, result.Response)
		result.Passed = len(result.Errors) == 0
		if !result.Passed {
			report.Passed = false
		}
		report.CaseResults = append(report.CaseResults, result)
	}

	return report, nil
}

func (c *Client) callReview(ctx context.Context, payload []byte) (*api.ReviewResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/reviews", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build review request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute review request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("review status %d: %s", resp.StatusCode, string(body))
	}

	var decoded api.ReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode review response: %w", err)
	}
	return &decoded, nil
}

func extractRequestMeta(body []byte) (requestMeta, error) {
	var env requestEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return requestMeta{}, err
	}
	return requestMeta{
		CaseID:    firstNonEmpty(env.RawAlert.CaseID, env.CaseID),
		AlertID:   env.RawAlert.AlertID,
		AlertType: env.RawAlert.AlertType,
	}, nil
}

func responseMismatch(meta requestMeta, resp *api.ReviewResponse) string {
	if resp == nil {
		return "nil response"
	}
	if meta.CaseID != "" && resp.CaseID != meta.CaseID {
		return fmt.Sprintf("response case_id mismatch: expected=%s got=%s", meta.CaseID, resp.CaseID)
	}
	if meta.AlertID != "" && resp.AlertID != meta.AlertID {
		return fmt.Sprintf("response alert_id mismatch: expected=%s got=%s", meta.AlertID, resp.AlertID)
	}
	return ""
}

func cloneReviewResponse(in *api.ReviewResponse) *api.ReviewResponse {
	if in == nil {
		return nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		clone := *in
		return &clone
	}
	var out api.ReviewResponse
	if err := json.Unmarshal(b, &out); err != nil {
		clone := *in
		return &clone
	}
	return &out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
