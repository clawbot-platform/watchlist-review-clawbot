package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/feedback"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/parsers"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/parsers/screeningjson"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/runtime"
)

type Server struct {
	registry        *parsers.Registry
	flow            *runtime.Flow
	feedbackService *feedback.Service
}

func NewServer(identityClient *identity.Client, extras ...any) (*Server, error) {
	registry, err := parsers.NewRegistry(
		screeningjson.New(),
	)
	if err != nil {
		return nil, err
	}

	var noteService *notes.Service
	var artifactService *artifacts.Service
	var retrievalService *retrieval.Service
	var feedbackService *feedback.Service
	for _, extra := range extras {
		switch value := extra.(type) {
		case *notes.Service:
			noteService = value
		case *artifacts.Service:
			artifactService = value
		case *retrieval.Service:
			retrievalService = value
		case *feedback.Service:
			feedbackService = value
		}
	}

	return &Server{
		registry:        registry,
		flow:            runtime.NewFlow(identityClient, noteService, artifactService, retrievalService),
		feedbackService: feedbackService,
	}, nil
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/v1/reviews", s.review)
	mux.HandleFunc("/v1/feedback", s.feedback)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) review(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(req.SourceSystem) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "source_system is required"})
		return
	}
	if len(req.RawAlert) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "raw_alert is required"})
		return
	}

	alert, err := s.registry.Parse(r.Context(), req.SourceSystem, req.RawAlert)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("parse alert: %v", err)})
		return
	}

	correlationID := correlationIDFromRequest(r)
	ctx, err := s.flow.BuildReviewContext(r.Context(), runtime.ReviewInput{
		TenantID:      req.TenantID,
		CaseID:        chooseCaseID(req.CaseID, alert.Metadata.CaseID),
		CorrelationID: correlationID,
		Alert:         alert,
	})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": fmt.Sprintf("build review context: %v", err)})
		return
	}

	resp := ReviewResponse{
		Status:           "review_context_built",
		CaseID:           chooseCaseID(req.CaseID, alert.Metadata.CaseID),
		AlertID:          alert.Metadata.AlertID,
		Warnings:         append([]string(nil), ctx.IdentityEvidence.Warnings...),
		ReviewContext:    ctx,
		AnalystNote:      ctx.AnalystNote,
		ArtifactRefs:     append([]artifacts.ArtifactRef(nil), ctx.ArtifactRefs...),
		ArtifactWarnings: append([]string(nil), ctx.ArtifactWarnings...),
	}
	if ctx.IdentityEvidence.Compare != nil {
		resp.IdentityTraceRefs.DecisionTraceID = ctx.IdentityEvidence.Compare.DecisionTraceID
		if ctx.IdentityEvidence.Compare.Explanation != nil {
			resp.IdentityTraceRefs.ExplanationID = ctx.IdentityEvidence.Compare.Explanation.ExplanationID
		}
	}
	if ctx.IdentityEvidence.Screening != nil {
		resp.IdentityTraceRefs.ScreeningID = ctx.IdentityEvidence.Screening.ScreeningID
		if resp.IdentityTraceRefs.DecisionTraceID == "" {
			resp.IdentityTraceRefs.DecisionTraceID = ctx.IdentityEvidence.Screening.DecisionTraceID
		}
	}
	if ctx.DeterministicScore != nil {
		resp.MatchStrengthScore = ctx.DeterministicScore.MatchStrengthScore
		resp.DataSufficiencyScore = ctx.DeterministicScore.DataSufficiencyScore
		resp.Contradictions = append([]string(nil), ctx.DeterministicScore.Contradictions...)
		resp.DecisionLabel = ctx.DeterministicScore.DecisionLabel
		resp.DecisionReason = ctx.DeterministicScore.DecisionReason
		resp.EvidenceFor = append([]string(nil), ctx.DeterministicScore.EvidenceFor...)
		resp.EvidenceAgainst = append([]string(nil), ctx.DeterministicScore.EvidenceAgainst...)
		resp.MissingInformation = append([]string(nil), ctx.DeterministicScore.MissingInformation...)
		resp.NextStep = ctx.DeterministicScore.NextStep
	}

	w.Header().Set("X-Correlation-ID", correlationID)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) feedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if s == nil || s.feedbackService == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "feedback capture is not configured"})
		return
	}

	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	result, err := s.feedbackService.Create(r.Context(), feedback.CreateInput{
		TenantID:          req.TenantID,
		CaseID:            req.CaseID,
		AlertID:           req.AlertID,
		CorrelationID:     correlationIDFromRequest(r),
		AnalystID:         req.AnalystID,
		SystemDecision:    req.SystemDecision,
		DecisionAgreement: req.DecisionAgreement,
		CorrectedLabel:    req.CorrectedLabel,
		NoteRating:        req.NoteRating,
		OutcomeRating:     req.OutcomeRating,
		Comment:           req.Comment,
		Tags:              req.Tags,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, FeedbackResponse{
		Status:       "feedback_captured",
		Feedback:     result.Feedback,
		ArtifactRefs: result.ArtifactRefs,
		Warnings:     result.Warnings,
	})
}

func correlationIDFromRequest(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Correlation-ID")); v != "" {
		return v
	}
	return "corr_" + time.Now().UTC().Format("20060102150405.000000000")
}

func chooseCaseID(values ...string) string {
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
