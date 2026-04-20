package features

import (
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

func TestExtractIndividualOnboarding(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{AlertID: "alert-1", SourceSystem: "firco", CreatedAt: time.Now().UTC()},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name: alerts.Name{FullName: "Jane Citizen", Aliases: []string{"J Citizen"}},
			DateOfBirth: "1988-07-14",
			Countries: []string{"US"},
			Identifiers: []alerts.Identifier{{Type: alerts.IdentifierTypePassport, Value: "P1234567"}},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "ofac",
			EntityType: alerts.EntityTypeIndividual,
			Name: alerts.Name{FullName: "Jane Citizen", Aliases: []string{"Jane Citizen Alias"}},
			DateOfBirth: "1988-07-14",
			Countries: []string{"US"},
			Identifiers: []alerts.Identifier{{Type: alerts.IdentifierTypePassport, Value: "P1234567"}},
		},
	}
	got, err := Extract(alert)
	if err != nil { t.Fatalf("Extract() error = %v", err) }
	if got.ScreenedName.Canonical != "JANE CITIZEN" { t.Fatalf("canonical = %q", got.ScreenedName.Canonical) }
	if !got.Date.ExactMatch { t.Fatal("expected exact date match") }
	if len(got.Identifiers.ExactMatches) != 1 { t.Fatalf("expected exact identifier match") }
	if !got.Geography.HasCountrySupport { t.Fatal("expected country support") }
	if len(got.Contradictions) != 0 { t.Fatalf("unexpected contradictions: %+v", got.Contradictions) }
}

func TestExtractWrongContextConflict(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindACHParty,
		Metadata: alerts.AlertMetadata{AlertID: "alert-2", SourceSystem: "firco", CreatedAt: time.Now().UTC()},
		ScreenedParty: alerts.Party{EntityType: alerts.EntityTypeIndividual, Name: alerts.Name{FullName: "Mercury"}, Countries: []string{"US"}},
		MatchedParty: alerts.MatchedParty{ListSource: "ofac", EntityType: alerts.EntityTypeVessel, Name: alerts.Name{FullName: "Mercury"}, Countries: []string{"RU"}},
		Transaction: &alerts.TransactionContext{TransactionID: "txn-1", RailType: "ach"},
	}
	got, err := Extract(alert)
	if err != nil { t.Fatalf("Extract() error = %v", err) }
	if !got.Context.EntityTypeConflict { t.Fatal("expected entity type conflict") }
	if len(got.Contradictions) == 0 { t.Fatal("expected contradictions") }
}
