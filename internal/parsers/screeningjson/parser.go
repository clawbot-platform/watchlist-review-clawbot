package screeningjson

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) SourceSystem() string {
	return "screening_json"
}

type payload struct {
	AlertID       string              `json:"alert_id"`
	SourceSystem  string              `json:"source_system"`
	CreatedAt     string              `json:"created_at"`
	CaseID        string              `json:"case_id,omitempty"`
	AlertType     string              `json:"alert_type,omitempty"`
	Jurisdiction  string              `json:"jurisdiction,omitempty"`
	Priority      string              `json:"priority,omitempty"`
	ScreenedParty partyPayload        `json:"screened_party"`
	MatchedParty  matchedPayload      `json:"matched_party"`
	Screening     screeningPayload    `json:"screening_features,omitempty"`
	Transaction   *transactionPayload `json:"transaction,omitempty"`
	Artifacts     []artifactPayload   `json:"artifacts,omitempty"`
}

type partyPayload struct {
	EntityType        string              `json:"entity_type"`
	FullName          string              `json:"full_name"`
	Aliases           []string            `json:"aliases,omitempty"`
	NativeName        string              `json:"native_name,omitempty"`
	DateOfBirth       string              `json:"date_of_birth,omitempty"`
	BirthYear         string              `json:"birth_year,omitempty"`
	IncorporationDate string              `json:"incorporation_date,omitempty"`
	Countries         []string            `json:"countries,omitempty"`
	Nationalities     []string            `json:"nationalities,omitempty"`
	Addresses         []addressPayload    `json:"addresses,omitempty"`
	Identifiers       []identifierPayload `json:"identifiers,omitempty"`
	SourceRecordID    string              `json:"source_record_id,omitempty"`
	SourceSystem      string              `json:"source_system,omitempty"`
	Attributes        map[string]any      `json:"attributes,omitempty"`
}

type matchedPayload struct {
	ListSource        string              `json:"list_source"`
	Program           string              `json:"program,omitempty"`
	ListUID           string              `json:"list_uid,omitempty"`
	EntityType        string              `json:"entity_type"`
	FullName          string              `json:"full_name"`
	Aliases           []string            `json:"aliases,omitempty"`
	NativeName        string              `json:"native_name,omitempty"`
	DateOfBirth       string              `json:"date_of_birth,omitempty"`
	BirthYear         string              `json:"birth_year,omitempty"`
	IncorporationDate string              `json:"incorporation_date,omitempty"`
	Countries         []string            `json:"countries,omitempty"`
	Addresses         []addressPayload    `json:"addresses,omitempty"`
	Identifiers       []identifierPayload `json:"identifiers,omitempty"`
	Attributes        map[string]any      `json:"attributes,omitempty"`
}

type addressPayload struct {
	AddressText string `json:"address_text,omitempty"`
	City        string `json:"city,omitempty"`
	Region      string `json:"region,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Country     string `json:"country,omitempty"`
}

type identifierPayload struct {
	Type           string `json:"type"`
	Value          string `json:"value"`
	IssuingCountry string `json:"issuing_country,omitempty"`
}

type screeningPayload struct {
	VendorScore    *float64 `json:"vendor_score,omitempty"`
	NameScore      *float64 `json:"name_score,omitempty"`
	MatchFlags     []string `json:"match_flags,omitempty"`
	ReasonCodes    []string `json:"reason_codes,omitempty"`
	AnalystNotes   []string `json:"analyst_notes,omitempty"`
	ReviewComments []string `json:"review_comments,omitempty"`
}

type transactionPayload struct {
	TransactionID    string  `json:"transaction_id"`
	RailType         string  `json:"rail_type"`
	Amount           float64 `json:"amount,omitempty"`
	Currency         string  `json:"currency,omitempty"`
	OriginatorRole   string  `json:"originator_role,omitempty"`
	BeneficiaryRole  string  `json:"beneficiary_role,omitempty"`
	PaymentReference string  `json:"payment_reference,omitempty"`
	CountryCorridor  string  `json:"country_corridor,omitempty"`
	Institution      string  `json:"institution,omitempty"`
}

type artifactPayload struct {
	Kind        string `json:"kind"`
	URI         string `json:"uri"`
	Description string `json:"description,omitempty"`
}

func (p *Parser) Parse(_ context.Context, raw []byte) (*alerts.CanonicalAlert, error) {
	var in payload
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("decode screening json payload: %w", err)
	}

	createdAt, err := parseTimestamp(in.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	out := &alerts.CanonicalAlert{
		Kind: inferKind(in),
		Metadata: alerts.AlertMetadata{
			AlertID:      in.AlertID,
			SourceSystem: firstNonEmpty(in.SourceSystem, "screening_json"),
			CreatedAt:    createdAt,
			AlertType:    in.AlertType,
			Jurisdiction: in.Jurisdiction,
			CaseID:       in.CaseID,
			Priority:     in.Priority,
		},
		ScreenedParty: mapParty(in.ScreenedParty),
		MatchedParty:  mapMatched(in.MatchedParty),
		ScreeningFeatures: alerts.ScreeningFeatures{
			VendorScore:    in.Screening.VendorScore,
			NameScore:      in.Screening.NameScore,
			MatchFlags:     in.Screening.MatchFlags,
			ReasonCodes:    in.Screening.ReasonCodes,
			AnalystNotes:   in.Screening.AnalystNotes,
			ReviewComments: in.Screening.ReviewComments,
		},
		Artifacts: mapArtifacts(in.Artifacts),
	}

	if in.Transaction != nil {
		tx := alerts.TransactionContext{
			TransactionID:    in.Transaction.TransactionID,
			RailType:         in.Transaction.RailType,
			Amount:           in.Transaction.Amount,
			Currency:         in.Transaction.Currency,
			OriginatorRole:   in.Transaction.OriginatorRole,
			BeneficiaryRole:  in.Transaction.BeneficiaryRole,
			PaymentReference: in.Transaction.PaymentReference,
			CountryCorridor:  in.Transaction.CountryCorridor,
			Institution:      in.Transaction.Institution,
		}
		out.Transaction = &tx
	}

	out.Normalize()
	if err := out.Validate(); err != nil {
		return nil, err
	}

	return out, nil
}

func inferKind(in payload) alerts.AlertKind {
	alertType := strings.ToLower(strings.TrimSpace(in.AlertType))

	switch {
	case in.Transaction != nil || strings.Contains(alertType, "ach"):
		return alerts.AlertKindACHParty
	case strings.Contains(alertType, "organization"), strings.Contains(alertType, "business"):
		return alerts.AlertKindOrganizationOnboarding
	case strings.EqualFold(in.ScreenedParty.EntityType, string(alerts.EntityTypeOrganization)):
		return alerts.AlertKindOrganizationOnboarding
	case strings.EqualFold(in.ScreenedParty.EntityType, string(alerts.EntityTypeIndividual)):
		return alerts.AlertKindIndividualOnboarding
	default:
		return alerts.AlertKindUnknown
	}
}

func mapParty(in partyPayload) alerts.Party {
	return alerts.Party{
		EntityType: alerts.EntityType(strings.ToLower(strings.TrimSpace(in.EntityType))),
		Name: alerts.Name{
			FullName:   in.FullName,
			Aliases:    in.Aliases,
			NativeName: in.NativeName,
		},
		DateOfBirth:       in.DateOfBirth,
		BirthYear:         in.BirthYear,
		IncorporationDate: in.IncorporationDate,
		Countries:         in.Countries,
		Nationalities:     in.Nationalities,
		Addresses:         mapAddresses(in.Addresses),
		Identifiers:       mapIdentifiers(in.Identifiers),
		SourceRecordID:    in.SourceRecordID,
		SourceSystem:      in.SourceSystem,
		AdditionalAttributes: in.Attributes,
	}
}

func mapMatched(in matchedPayload) alerts.MatchedParty {
	return alerts.MatchedParty{
		ListSource: in.ListSource,
		Program:    in.Program,
		ListUID:    in.ListUID,
		EntityType: alerts.EntityType(strings.ToLower(strings.TrimSpace(in.EntityType))),
		Name: alerts.Name{
			FullName:   in.FullName,
			Aliases:    in.Aliases,
			NativeName: in.NativeName,
		},
		DateOfBirth:       in.DateOfBirth,
		BirthYear:         in.BirthYear,
		IncorporationDate: in.IncorporationDate,
		Countries:         in.Countries,
		Addresses:         mapAddresses(in.Addresses),
		Identifiers:       mapIdentifiers(in.Identifiers),
		AdditionalAttributes: in.Attributes,
	}
}

func mapAddresses(in []addressPayload) []alerts.Address {
	if len(in) == 0 {
		return nil
	}

	out := make([]alerts.Address, 0, len(in))
	for _, a := range in {
		out = append(out, alerts.Address{
			AddressText: a.AddressText,
			City:        a.City,
			Region:      a.Region,
			PostalCode:  a.PostalCode,
			Country:     a.Country,
		})
	}
	return out
}

func mapIdentifiers(in []identifierPayload) []alerts.Identifier {
	if len(in) == 0 {
		return nil
	}

	out := make([]alerts.Identifier, 0, len(in))
	for _, id := range in {
		out = append(out, alerts.Identifier{
			Type:           mapIdentifierType(id.Type),
			Value:          id.Value,
			IssuingCountry: id.IssuingCountry,
		})
	}
	return out
}

func mapArtifacts(in []artifactPayload) []alerts.ArtifactRef {
	if len(in) == 0 {
		return nil
	}

	out := make([]alerts.ArtifactRef, 0, len(in))
	for _, a := range in {
		out = append(out, alerts.ArtifactRef{
			Kind:        a.Kind,
			URI:         a.URI,
			Description: a.Description,
		})
	}
	return out
}

func mapIdentifierType(v string) alerts.IdentifierType {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "passport":
		return alerts.IdentifierTypePassport
	case "national_id", "national-id", "nid":
		return alerts.IdentifierTypeNationalID
	case "tax_id", "tax-id", "tin":
		return alerts.IdentifierTypeTaxID
	case "registration_id", "registration-id", "company_id", "company-id":
		return alerts.IdentifierTypeRegistrationID
	case "customer_id", "customer-id":
		return alerts.IdentifierTypeCustomerID
	case "imo":
		return alerts.IdentifierTypeIMO
	case "", "unknown":
		return alerts.IdentifierTypeUnknown
	default:
		return alerts.IdentifierTypeOther
	}
}

func parseTimestamp(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, fmt.Errorf("value is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if ts, err := time.Parse(layout, v); err == nil {
			return ts.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported timestamp %q", v)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}