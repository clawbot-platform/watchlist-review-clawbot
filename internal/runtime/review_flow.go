package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

type IdentityEvidence struct {
	Compare   *identity.CompareResponse    `json:"compare,omitempty"`
	Screening *identity.ScreenOFACResponse `json:"screening,omitempty"`
	Warnings  []string                     `json:"warnings,omitempty"`
}

type ReviewInput struct {
	TenantID      string
	CaseID        string
	CorrelationID string
	Alert         *alerts.CanonicalAlert
}

type ReviewContext struct {
	Alert              *alerts.CanonicalAlert      `json:"alert"`
	Features           *features.ExtractedFeatures `json:"features"`
	IdentityEvidence   IdentityEvidence            `json:"identity_evidence"`
	DeterministicScore *scoring.Result             `json:"deterministic_score,omitempty"`
	AnalystNote        *notes.AnalystNote          `json:"analyst_note,omitempty"`
}

type Flow struct {
	Identity *identity.Client
	Notes    *notes.Service
}

func NewFlow(identityClient *identity.Client, noteService ...*notes.Service) *Flow {
	var ns *notes.Service
	if len(noteService) > 0 {
		ns = noteService[0]
	}
	return &Flow{Identity: identityClient, Notes: ns}
}

func (f *Flow) BuildReviewContext(ctx context.Context, input ReviewInput) (*ReviewContext, error) {
	if input.Alert == nil {
		return nil, fmt.Errorf("alert is required")
	}
	if err := input.Alert.Validate(); err != nil {
		return nil, err
	}

	extracted, err := features.Extract(input.Alert)
	if err != nil {
		return nil, fmt.Errorf("extract features: %w", err)
	}

	evidence, err := f.EnrichWithIdentity(ctx, input)
	if err != nil {
		return nil, err
	}

	deterministic := scoring.Evaluate(input.Alert, extracted)
	analystNote := f.generateAnalystNote(ctx, input.Alert, extracted, deterministic, evidence)

	return &ReviewContext{
		Alert:              input.Alert,
		Features:           extracted,
		IdentityEvidence:   evidence,
		DeterministicScore: deterministic,
		AnalystNote:        analystNote,
	}, nil
}

func (f *Flow) generateAnalystNote(
	ctx context.Context,
	alert *alerts.CanonicalAlert,
	fx *features.ExtractedFeatures,
	score *scoring.Result,
	evidence IdentityEvidence,
) *notes.AnalystNote {
	if f == nil || f.Notes == nil {
		return &notes.AnalystNote{
			Status:   notes.StatusSkipped,
			Warnings: []string{"granite_analyst_note_not_configured"},
		}
	}
	return f.Notes.Generate(ctx, alert, fx, score, evidence.Compare, evidence.Screening)
}

func (f *Flow) EnrichWithIdentity(ctx context.Context, input ReviewInput) (IdentityEvidence, error) {
	if input.Alert == nil {
		return IdentityEvidence{}, fmt.Errorf("alert is required")
	}

	evidence := IdentityEvidence{}
	if f == nil || f.Identity == nil || strings.TrimSpace(f.Identity.BaseURL()) == "" {
		evidence.Warnings = append(evidence.Warnings, "claw_identity_not_configured")
		return evidence, nil
	}

	screenReq, err := identity.BuildScreenOFACRequest(input.Alert, strings.TrimSpace(input.TenantID), strings.TrimSpace(input.CaseID))
	if err != nil {
		return IdentityEvidence{}, fmt.Errorf("build ofac screening request: %w", err)
	}

	screenResp, err := f.Identity.ScreenOFAC(ctx, screenReq, input.CorrelationID)
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "ofac_screening_failed")
	} else {
		evidence.Screening = &screenResp
	}

	compareReq, err := identity.BuildCompareRequest(input.Alert, strings.TrimSpace(input.TenantID), true)
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "compare_request_unavailable")
		return evidence, nil
	}

	compareResp, err := f.Identity.Compare(ctx, compareReq, input.CorrelationID, screenReq.CaseID)
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "compare_failed")
		return evidence, nil
	}

	evidence.Compare = &compareResp
	return evidence, nil
}
