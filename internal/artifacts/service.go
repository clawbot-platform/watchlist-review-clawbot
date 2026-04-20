package artifacts

import (
	"context"
	"fmt"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/notes"
)

type Service struct {
	store          Store
	manifest       ManifestWriter
	eventPublisher EventPublisher
	now            func() time.Time
}

func NewService(store Store, extras ...any) *Service {
	service := &Service{
		store: store,
		now:   time.Now,
	}
	for _, extra := range extras {
		switch value := extra.(type) {
		case ManifestWriter:
			service.manifest = value
		case EventPublisher:
			service.eventPublisher = value
		}
	}
	return service
}

func (s *Service) Persist(ctx context.Context, input PersistInput) ([]ArtifactRef, []string) {
	if s == nil || s.store == nil {
		return nil, nil
	}

	var refs []ArtifactRef
	var warnings []string

	reviewPayload := map[string]any{
		"decision_label": input.DecisionLabel,
		"review_context": input.ReviewContext,
	}
	ref, err := s.store.WriteJSON(WriteInput{
		TenantID:      input.TenantID,
		CaseID:        input.CaseID,
		AlertID:       input.AlertID,
		CorrelationID: input.CorrelationID,
		Kind:          KindReviewOutput,
		Payload:       reviewPayload,
	})
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("persist_review_output_failed: %v", err))
	} else {
		refs = append(refs, ref)
		warnings = append(warnings, s.publishCreated(ctx, input, ref)...)
	}

	if input.AnalystNote != nil {
		ref, err := s.store.WriteJSON(WriteInput{
			TenantID:      input.TenantID,
			CaseID:        input.CaseID,
			AlertID:       input.AlertID,
			CorrelationID: input.CorrelationID,
			Kind:          KindAnalystNote,
			Payload:       input.AnalystNote,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("persist_analyst_note_failed: %v", err))
		} else {
			refs = append(refs, ref)
			warnings = append(warnings, s.publishCreated(ctx, input, ref)...)
		}
	}

	if s.manifest != nil && len(refs) > 0 {
		ref, err := s.manifest.UpsertCaseManifest(ManifestInput{
			TenantID:      input.TenantID,
			CaseID:        input.CaseID,
			CorrelationID: input.CorrelationID,
			NewArtifacts:  refs,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("persist_case_manifest_failed: %v", err))
		} else {
			refs = append(refs, ref)
			warnings = append(warnings, s.publishCreated(ctx, input, ref)...)
		}
	}

	return refs, dedupeStrings(warnings)
}

func (s *Service) publishCreated(ctx context.Context, input PersistInput, ref ArtifactRef) []string {
	if s == nil || s.eventPublisher == nil {
		return nil
	}
	err := s.eventPublisher.PublishArtifactCreated(ctx, ArtifactCreatedEvent{
		EventID:       fmt.Sprintf("evt_%d", s.now().UTC().UnixNano()),
		EventType:     "clawbot.watchlist.review.artifact.created.v1",
		OccurredAt:    s.now().UTC(),
		TenantID:      input.TenantID,
		CaseID:        input.CaseID,
		AlertID:       input.AlertID,
		CorrelationID: input.CorrelationID,
		DecisionLabel: input.DecisionLabel,
		Artifact:      ref,
	})
	if err != nil {
		return []string{fmt.Sprintf("publish_artifact_created_failed: %v", err)}
	}
	return nil
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, item := range in {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func PersistableNote(note *notes.AnalystNote) *notes.AnalystNote {
	if note == nil {
		return nil
	}
	return note
}
