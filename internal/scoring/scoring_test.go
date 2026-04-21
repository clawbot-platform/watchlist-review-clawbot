package scoring

import (
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
)

func TestEvaluateStrongIndividualMatchEscalates(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind:     alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{AlertID: "alert-1", SourceSystem: "screening_json", CreatedAt: time.Now().UTC()},
		ScreenedParty: alerts.Party{
			EntityType:  alerts.EntityTypeIndividual,
			Name:        alerts.Name{FullName: "Jane Citizen", Aliases: []string{"J Citizen"}},
			DateOfBirth: "1988-07-14",
			Countries:   []string{"US"},
			Addresses:   []alerts.Address{{AddressText: "1 Main St", Country: "US"}},
			Identifiers: []alerts.Identifier{{Type: alerts.IdentifierTypePassport, Value: "P1234567"}},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource:  "watchlist_candidates",
			ListUID:     "right-record",
			EntityType:  alerts.EntityTypeIndividual,
			Name:        alerts.Name{FullName: "Jane Citizen", Aliases: []string{"Jane Q Citizen"}},
			DateOfBirth: "1988-07-14",
			Countries:   []string{"US"},
			Addresses:   []alerts.Address{{AddressText: "1 Main St", Country: "US"}},
			Identifiers: []alerts.Identifier{{Type: alerts.IdentifierTypePassport, Value: "P1234567"}},
		},
		ScreeningFeatures: alerts.ScreeningFeatures{MatchFlags: []string{"name", "dob", "identifier"}},
	}
	fx, err := features.Extract(alert)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	got := Evaluate(alert, fx)
	if got.DecisionLabel != DecisionEscalate {
		t.Fatalf("DecisionLabel = %q, do not want escalate (%+v)", got.DecisionLabel, got)
	}
	if got.MatchStrengthScore < 70 {
		t.Fatalf("MatchStrengthScore = %d, want >= 70", got.MatchStrengthScore)
	}
	if got.DataSufficiencyScore < 50 {
		t.Fatalf("DataSufficiencyScore = %d, want >= 50", got.DataSufficiencyScore)
	}
}

func TestEvaluateWrongContextDoesNotEscalate(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind:     alerts.AlertKindACHParty,
		Metadata: alerts.AlertMetadata{AlertID: "alert-2", SourceSystem: "screening_json", CreatedAt: time.Now().UTC()},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Mercury"},
			Countries:  []string{"US"},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "ofac",
			EntityType: alerts.EntityTypeVessel,
			Name:       alerts.Name{FullName: "Mercury"},
			Countries:  []string{"RU"},
		},
		Transaction: &alerts.TransactionContext{TransactionID: "txn-1", RailType: "ach"},
	}
	fx, err := features.Extract(alert)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	got := Evaluate(alert, fx)
	if got.DecisionLabel == DecisionEscalate {
		t.Fatalf("DecisionLabel = %q, do not want escalate (%+v)", got.DecisionLabel, got)
	}
	if !hasAny(got.Contradictions, "entity_type_conflict") {
		t.Fatalf("expected entity_type_conflict, got %+v", got.Contradictions)
	}
}

func TestEvaluateNameOnlyCapsAtInvestigate(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind:     alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{AlertID: "alert-3", SourceSystem: "screening_json", CreatedAt: time.Now().UTC()},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Common Name"},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "ofac",
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Common Name"},
		},
	}
	fx, err := features.Extract(alert)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	got := Evaluate(alert, fx)
	if got.MatchStrengthScore > 69 {
		t.Fatalf("MatchStrengthScore = %d, want <= 69", got.MatchStrengthScore)
	}
	if got.DecisionLabel == DecisionEscalate {
		t.Fatalf("DecisionLabel = %q, do not want escalate", got.DecisionLabel)
	}
}
