package notes

import (
	"context"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

type Generator interface {
	Generate(context.Context, PromptInput) (*AnalystNote, error)
}

type Service struct {
	generator Generator
}

func NewService(generator Generator) *Service {
	return &Service{generator: generator}
}

func (s *Service) Generate(
	ctx context.Context,
	alert *alerts.CanonicalAlert,
	fx *features.ExtractedFeatures,
	score *scoring.Result,
	compare *identity.CompareResponse,
	screening *identity.ScreenOFACResponse,
	retrievalContext *retrieval.PromptContext,
) *AnalystNote {
	if s == nil || s.generator == nil {
		return &AnalystNote{
			Status:   StatusSkipped,
			Warnings: []string{"granite_analyst_note_not_configured"},
		}
	}

	note, err := s.generator.Generate(ctx, PromptInput{
		Alert:            alert,
		Features:         fx,
		Score:            score,
		Compare:          compare,
		Screening:        screening,
		RetrievalContext: retrievalContext,
	})
	if err != nil {
		return &AnalystNote{
			Status:   StatusFailed,
			Warnings: []string{"granite_analyst_note_failed", err.Error()},
		}
	}

	sanitizeGroundedNames(note, alert, screening, score)
	return NormalizeAndValidateWithContext(note, alert, fx, score)
}
