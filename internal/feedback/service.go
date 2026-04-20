package feedback

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/artifacts"
)

type Service struct {
	store          artifacts.Store
	manifest       artifacts.ManifestWriter
	eventPublisher EventPublisher
	now            func() time.Time
}

func NewService(store artifacts.Store, extras ...any) *Service {
	s := &Service{
		store: store,
		now:   time.Now,
	}
	for _, extra := range extras {
		switch value := extra.(type) {
		case artifacts.ManifestWriter:
			s.manifest = value
		case EventPublisher:
			s.eventPublisher = value
		}
	}
	return s
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateResult, error) {
	if s == nil || s.store == nil {
		return CreateResult{}, fmt.Errorf("feedback store is not configured")
	}
	feedback, err := buildFeedback(s.now().UTC(), input)
	if err != nil {
		return CreateResult{}, err
	}

	var refs []artifacts.ArtifactRef
	var warnings []string

	ref, err := s.store.WriteJSON(artifacts.WriteInput{
		TenantID:      feedback.TenantID,
		CaseID:        feedback.CaseID,
		AlertID:       feedback.AlertID,
		CorrelationID: feedback.CorrelationID,
		Kind:          artifacts.KindAnalystFeedback,
		Payload:       feedback,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("persist feedback artifact: %w", err)
	}
	refs = append(refs, ref)
	warnings = append(warnings, s.publishEvent(ctx, feedback, &ref)...)

	if s.manifest != nil {
		manifestRef, err := s.manifest.UpsertCaseManifest(artifacts.ManifestInput{
			TenantID:      feedback.TenantID,
			CaseID:        feedback.CaseID,
			CorrelationID: feedback.CorrelationID,
			NewArtifacts:  refs,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("persist_feedback_manifest_failed: %v", err))
		} else {
			refs = append(refs, manifestRef)
		}
	}

	return CreateResult{
		Feedback:     feedback,
		ArtifactRefs: refs,
		Warnings:     dedupeStrings(warnings),
	}, nil
}

func buildFeedback(now time.Time, input CreateInput) (AnalystFeedback, error) {
	if strings.TrimSpace(input.TenantID) == "" {
		return AnalystFeedback{}, fmt.Errorf("tenant_id is required")
	}
	if strings.TrimSpace(input.CaseID) == "" {
		return AnalystFeedback{}, fmt.Errorf("case_id is required")
	}
	switch input.DecisionAgreement {
	case DecisionAgreementAgree, DecisionAgreementDisagree, DecisionAgreementPartial:
	default:
		return AnalystFeedback{}, fmt.Errorf("decision_agreement must be one of agree, disagree, partial")
	}
	if input.NoteRating < 0 || input.NoteRating > 5 {
		return AnalystFeedback{}, fmt.Errorf("note_rating must be between 0 and 5")
	}
	if input.OutcomeRating < 0 || input.OutcomeRating > 5 {
		return AnalystFeedback{}, fmt.Errorf("outcome_rating must be between 0 and 5")
	}

	feedback := AnalystFeedback{
		FeedbackID:        fmt.Sprintf("fb_%d", now.UnixNano()),
		TenantID:          strings.TrimSpace(input.TenantID),
		CaseID:            strings.TrimSpace(input.CaseID),
		AlertID:           strings.TrimSpace(input.AlertID),
		CorrelationID:     strings.TrimSpace(input.CorrelationID),
		CreatedAt:         now,
		AnalystID:         strings.TrimSpace(input.AnalystID),
		SystemDecision:    strings.TrimSpace(input.SystemDecision),
		DecisionAgreement: input.DecisionAgreement,
		CorrectedLabel:    strings.TrimSpace(input.CorrectedLabel),
		NoteRating:        input.NoteRating,
		OutcomeRating:     input.OutcomeRating,
		Comment:           strings.TrimSpace(input.Comment),
		Tags:              cleanTags(input.Tags),
	}
	feedback.DerivedSignals = deriveSignals(feedback)
	return feedback, nil
}

func deriveSignals(feedback AnalystFeedback) []string {
	var out []string
	if feedback.DecisionAgreement != DecisionAgreementAgree {
		out = append(out, "decision_policy_tuning_candidate")
	}
	if feedback.NoteRating > 0 && feedback.NoteRating <= 3 {
		out = append(out, "note_quality_tuning_candidate")
	}
	if feedback.OutcomeRating > 0 && feedback.OutcomeRating <= 3 {
		out = append(out, "outcome_quality_tuning_candidate")
	}
	for _, tag := range feedback.Tags {
		switch strings.ToLower(tag) {
		case "retrieval_gap":
			out = append(out, "retrieval_tuning_candidate")
		case "prompt_issue":
			out = append(out, "prompt_tuning_candidate")
		case "false_positive":
			out = append(out, "false_positive_learning_candidate")
		case "false_negative":
			out = append(out, "false_negative_learning_candidate")
		}
	}
	return dedupeStrings(out)
}

func cleanTags(tags []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tag)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (s *Service) publishEvent(ctx context.Context, feedback AnalystFeedback, ref *artifacts.ArtifactRef) []string {
	if s == nil || s.eventPublisher == nil {
		return nil
	}
	err := s.eventPublisher.PublishFeedbackCreated(ctx, FeedbackCreatedEvent{
		EventID:        fmt.Sprintf("evt_%d", s.now().UTC().UnixNano()),
		EventType:      "clawbot.watchlist.review.feedback.created.v1",
		OccurredAt:     s.now().UTC(),
		TenantID:       feedback.TenantID,
		CaseID:         feedback.CaseID,
		AlertID:        feedback.AlertID,
		CorrelationID:  feedback.CorrelationID,
		SystemDecision: feedback.SystemDecision,
		Feedback:       feedback,
		ArtifactRef:    ref,
	})
	if err != nil {
		return []string{fmt.Sprintf("publish_feedback_created_failed: %v", err)}
	}
	return nil
}

func dedupeStrings(in []string) []string {
	var out []string
	seen := map[string]struct{}{}
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
