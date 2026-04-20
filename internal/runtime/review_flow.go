package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
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
	ArtifactRefs       []artifacts.ArtifactRef     `json:"artifact_refs,omitempty"`
	ArtifactWarnings   []string                    `json:"artifact_warnings,omitempty"`
}

type Flow struct {
	Identity  *identity.Client
	Notes     *notes.Service
	Artifacts *artifacts.Service
}

func NewFlow(identityClient *identity.Client, noteService *notes.Service, artifactService *artifacts.Service) *Flow {
	return &Flow{
		Identity:  identityClient,
		Notes:     noteService,
		Artifacts: artifactService,
	}
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

	reviewContext := &ReviewContext{
		Alert:              input.Alert,
		Features:           extracted,
		IdentityEvidence:   evidence,
		DeterministicScore: deterministic,
		AnalystNote:        analystNote,
	}
	reviewContext.ArtifactRefs, reviewContext.ArtifactWarnings = f.persistArtifacts(ctx, input, reviewContext)

	return reviewContext, nil
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

func (f *Flow) persistArtifacts(ctx context.Context, input ReviewInput, reviewContext *ReviewContext) ([]artifacts.ArtifactRef, []string) {
	if f == nil || f.Artifacts == nil || input.Alert == nil || reviewContext == nil {
		return nil, nil
	}
	return f.Artifacts.Persist(ctx, artifacts.PersistInput{
		TenantID:      input.TenantID,
		CaseID:        firstNonEmpty(input.CaseID, input.Alert.Metadata.CaseID),
		AlertID:       input.Alert.Metadata.AlertID,
		CorrelationID: input.CorrelationID,
		DecisionLabel: reviewContext.DeterministicScore.DecisionLabel,
		ReviewContext: reviewContext,
		AnalystNote:   artifacts.PersistableNote(reviewContext.AnalystNote),
	})
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
