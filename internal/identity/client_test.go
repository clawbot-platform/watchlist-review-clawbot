package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCompareSendsHeadersAndDecodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/compare" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("X-Correlation-ID"); got != "corr-1" {
			t.Fatalf("X-Correlation-ID = %q", got)
		}
		if got := r.Header.Get("X-Case-ID"); got != "case-1" {
			t.Fatalf("X-Case-ID = %q", got)
		}
		var req CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Left.SourceRecordID != "left-record" {
			t.Fatalf("left record = %q", req.Left.SourceRecordID)
		}
		_ = json.NewEncoder(w).Encode(CompareResponse{
			Disposition:     "resolved",
			ConfidenceBand:  "high",
			DecisionTraceID: "dt-1",
			Explanation: &Explanation{ExplanationID: "exp-1"},
		})
	}))
	defer server.Close()

	client := New(server.URL, 0, "tenant-1")
	resp, err := client.Compare(context.Background(), CompareRequest{
		Left:    SourceRef{SourceSystem: "kyc_applications", SourceRecordID: "left-record"},
		Right:   SourceRef{SourceSystem: "watchlist_candidates", SourceRecordID: "right-record"},
		Explain: true,
	}, "corr-1", "case-1")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if resp.DecisionTraceID != "dt-1" {
		t.Fatalf("DecisionTraceID = %q", resp.DecisionTraceID)
	}
}

func TestClientScreenOFACDecodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/watchlist/ofac/screenings" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(ScreenOFACResponse{
			ScreeningID:     "scr-1",
			Decision:        "manual_review",
			DecisionTraceID: "dt-2",
		})
	}))
	defer server.Close()

	client := New(server.URL, 0, "tenant-1")
	resp, err := client.ScreenOFAC(context.Background(), ScreenOFACRequest{
		CaseID: "case-1",
		Subject: OFACSubject{
			Name: "Jane Citizen",
		},
	}, "corr-1")
	if err != nil {
		t.Fatalf("ScreenOFAC() error = %v", err)
	}
	if resp.ScreeningID != "scr-1" {
		t.Fatalf("ScreeningID = %q", resp.ScreeningID)
	}
}
