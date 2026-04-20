package notes

import (
	"context"
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

type fakeGenerator struct {
	note *AnalystNote
	err  error
}

func (f *fakeGenerator) Generate(_ context.Context, _ PromptInput) (*AnalystNote, error) {
	return f.note, f.err
}

func TestServiceGenerateSkippedWhenNotConfigured(t *testing.T) {
	service := NewService(nil)
	note := service.Generate(context.Background(), nil, nil, nil, nil, nil, nil)
	if note == nil || note.Status != StatusSkipped {
		t.Fatalf("note = %+v, want skipped", note)
	}
}

func TestServiceGenerateReturnsNote(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{
			AlertID:      "alert-1",
			SourceSystem: "screening_json",
			CreatedAt:    time.Now().UTC(),
		},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "watchlist_candidates",
			ListUID:    "right-record",
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
	}
	fx, err := features.Extract(alert)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	score := scoring.Evaluate(alert, fx)

	service := NewService(&fakeGenerator{
		note: &AnalystNote{
			Status: StatusGenerated,
			Note:   "Analyst note draft.",
		},
	})
	note := service.Generate(context.Background(), alert, fx, score, nil, nil, nil)
	if note == nil || note.Status != StatusGenerated {
		t.Fatalf("note = %+v, want generated", note)
	}
	if note.Note != "Analyst note draft." {
		t.Fatalf("note text = %q", note.Note)
	}
}

func TestBuildPrompt(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{
			AlertID:      "alert-1",
			SourceSystem: "screening_json",
			CreatedAt:    time.Now().UTC(),
		},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "watchlist_candidates",
			ListUID:    "right-record",
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
	}
	fx, err := features.Extract(alert)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	score := scoring.Evaluate(alert, fx)
	prompt, err := BuildPrompt(PromptInput{
		Alert:            alert,
		Features:         fx,
		Score:            score,
		RetrievalContext: nil,
	})
	if err != nil {
		t.Fatalf("BuildPrompt() error = %v", err)
	}
	if prompt == "" {
		t.Fatal("expected prompt")
	}
}
