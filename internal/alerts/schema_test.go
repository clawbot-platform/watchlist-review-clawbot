package alerts

import (
	"testing"
	"time"
)

func TestCanonicalAlertValidateIndividualOnboarding(t *testing.T) {
	alert := &CanonicalAlert{
		Kind: AlertKindIndividualOnboarding,
		Metadata: AlertMetadata{AlertID: "alert-1", SourceSystem: "firco", CreatedAt: time.Now().UTC()},
		ScreenedParty: Party{EntityType: EntityTypeIndividual, Name: Name{FullName: "Jane Citizen"}},
		MatchedParty:  MatchedParty{ListSource: "ofac", EntityType: EntityTypeIndividual, Name: Name{FullName: "Jane Citizen"}},
	}
	if err := alert.Validate(); err != nil { t.Fatalf("Validate() error = %v", err) }
}

func TestCanonicalAlertValidateACHRequiresTransaction(t *testing.T) {
	alert := &CanonicalAlert{
		Kind: AlertKindACHParty,
		Metadata: AlertMetadata{AlertID: "alert-2", SourceSystem: "actimize", CreatedAt: time.Now().UTC()},
		ScreenedParty: Party{EntityType: EntityTypeIndividual, Name: Name{FullName: "Jane Citizen"}},
		MatchedParty:  MatchedParty{ListSource: "ofac", EntityType: EntityTypeIndividual, Name: Name{FullName: "Jane Citizen"}},
	}
	if err := alert.Validate(); err == nil { t.Fatal("expected validation error") }
}
