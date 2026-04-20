package identity

import (
	"testing"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

func TestBuildCompareRequest(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{
			AlertID:      "alert-1",
			SourceSystem: "screening_json",
			CreatedAt:    time.Now().UTC(),
		},
		ScreenedParty: alerts.Party{
			EntityType:     alerts.EntityTypeIndividual,
			Name:           alerts.Name{FullName: "Jane Citizen"},
			SourceSystem:   "kyc_applications",
			SourceRecordID: "left-record",
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "watchlist_candidates",
			ListUID:    "right-record",
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
	}
	req, err := BuildCompareRequest(alert, "tenant-1", true)
	if err != nil {
		t.Fatalf("BuildCompareRequest() error = %v", err)
	}
	if req.Left.SourceRecordID != "left-record" || req.Right.SourceRecordID != "right-record" {
		t.Fatalf("unexpected compare request: %+v", req)
	}
}

func TestBuildScreenOFACRequest(t *testing.T) {
	alert := &alerts.CanonicalAlert{
		Kind: alerts.AlertKindIndividualOnboarding,
		Metadata: alerts.AlertMetadata{
			AlertID:      "alert-1",
			CaseID:       "case-1",
			SourceSystem: "screening_json",
			CreatedAt:    time.Now().UTC(),
		},
		ScreenedParty: alerts.Party{
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen", Aliases: []string{"J Citizen"}},
			DateOfBirth:"1988-07-14",
			Countries:  []string{"US"},
			Identifiers: []alerts.Identifier{
				{Type: alerts.IdentifierTypePassport, Value: "P1234567"},
			},
		},
		MatchedParty: alerts.MatchedParty{
			ListSource: "ofac",
			EntityType: alerts.EntityTypeIndividual,
			Name:       alerts.Name{FullName: "Jane Citizen"},
		},
	}
	req, err := BuildScreenOFACRequest(alert, "tenant-1", "")
	if err != nil {
		t.Fatalf("BuildScreenOFACRequest() error = %v", err)
	}
	if req.CaseID != "case-1" {
		t.Fatalf("CaseID = %q", req.CaseID)
	}
	if req.Subject.Name != "Jane Citizen" {
		t.Fatalf("Subject.Name = %q", req.Subject.Name)
	}
	if req.Subject.Identifiers["passport"] != "P1234567" {
		t.Fatalf("passport = %q", req.Subject.Identifiers["passport"])
	}
}
