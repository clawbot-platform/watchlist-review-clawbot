package screeningjson

import (
	"context"
	"testing"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

func TestParseIndividualOnboarding(t *testing.T) {
	raw := []byte(`{
		"alert_id":"screening-json-1",
		"source_system":"screening_json",
		"created_at":"2026-04-14T12:00:00Z",
		"case_id":"case-1",
		"alert_type":"individual_onboarding",
		"jurisdiction":"us",
		"screened_party":{
			"entity_type":"individual",
			"full_name":"Jane Citizen",
			"aliases":["J Citizen"],
			"date_of_birth":"1988-07-14",
			"countries":["us"],
			"identifiers":[{"type":"passport","value":"P1234567"}]
		},
		"matched_party":{
			"list_source":"ofac",
			"program":"sdn",
			"list_uid":"uid-1",
			"entity_type":"individual",
			"full_name":"Jane Citizen",
			"aliases":["Jane Q Citizen"],
			"date_of_birth":"1988-07-14",
			"countries":["us"],
			"identifiers":[{"type":"passport","value":"P1234567"}]
		},
		"screening_features":{
			"vendor_score":88,
			"name_score":95,
			"match_flags":["name","dob","identifier"]
		}
	}`)

	alert, err := New().Parse(context.Background(), raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if alert.Kind != alerts.AlertKindIndividualOnboarding {
		t.Fatalf("kind = %q, want %q", alert.Kind, alerts.AlertKindIndividualOnboarding)
	}
	if alert.Metadata.SourceSystem != "screening_json" {
		t.Fatalf("source_system = %q, want %q", alert.Metadata.SourceSystem, "screening_json")
	}
	if alert.Metadata.Jurisdiction != "US" {
		t.Fatalf("jurisdiction = %q, want %q", alert.Metadata.Jurisdiction, "US")
	}
	if alert.MatchedParty.Program != "SDN" {
		t.Fatalf("program = %q, want %q", alert.MatchedParty.Program, "SDN")
	}
}

func TestParseACHAlert(t *testing.T) {
	raw := []byte(`{
		"alert_id":"screening-json-ach-1",
		"source_system":"screening_json",
		"created_at":"2026-04-14T12:00:00Z",
		"alert_type":"ach_payment_screening",
		"screened_party":{
			"entity_type":"individual",
			"full_name":"Jane Citizen",
			"countries":["us"]
		},
		"matched_party":{
			"list_source":"ofac",
			"entity_type":"individual",
			"full_name":"Jane Citizen"
		},
		"transaction":{
			"transaction_id":"txn-1",
			"rail_type":"ach",
			"currency":"usd",
			"originator_role":"originator",
			"country_corridor":"us-us"
		}
	}`)

	alert, err := New().Parse(context.Background(), raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if alert.Kind != alerts.AlertKindACHParty {
		t.Fatalf("kind = %q, want %q", alert.Kind, alerts.AlertKindACHParty)
	}
	if alert.Transaction == nil {
		t.Fatal("expected transaction to be present")
	}
	if alert.Transaction.RailType != "ach" {
		t.Fatalf("transaction.rail_type = %q, want %q", alert.Transaction.RailType, "ach")
	}
	if alert.Metadata.SourceSystem != "screening_json" {
		t.Fatalf("source_system = %q, want %q", alert.Metadata.SourceSystem, "screening_json")
	}
}