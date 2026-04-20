package alerts

import "time"

type AlertKind string

const (
	AlertKindUnknown                AlertKind = "unknown"
	AlertKindIndividualOnboarding   AlertKind = "individual_onboarding"
	AlertKindOrganizationOnboarding AlertKind = "organization_onboarding"
	AlertKindACHParty               AlertKind = "ach_party"
)

type EntityType string

const (
	EntityTypeUnknown      EntityType = "unknown"
	EntityTypeIndividual   EntityType = "individual"
	EntityTypeOrganization EntityType = "organization"
	EntityTypeVessel       EntityType = "vessel"
)

type IdentifierType string

const (
	IdentifierTypeUnknown        IdentifierType = "unknown"
	IdentifierTypePassport       IdentifierType = "passport"
	IdentifierTypeNationalID     IdentifierType = "national_id"
	IdentifierTypeTaxID          IdentifierType = "tax_id"
	IdentifierTypeRegistrationID IdentifierType = "registration_id"
	IdentifierTypeCustomerID     IdentifierType = "customer_id"
	IdentifierTypeIMO            IdentifierType = "imo"
	IdentifierTypeOther          IdentifierType = "other"
)

type AlertMetadata struct {
	AlertID      string    `json:"alert_id"`
	SourceSystem string    `json:"source_system"`
	CreatedAt    time.Time `json:"created_at"`
	AlertType    string    `json:"alert_type,omitempty"`
	Jurisdiction string    `json:"jurisdiction,omitempty"`
	CaseID       string    `json:"case_id,omitempty"`
	Priority     string    `json:"priority,omitempty"`
}

type Name struct {
	FullName   string   `json:"full_name"`
	Aliases    []string `json:"aliases,omitempty"`
	NativeName string   `json:"native_name,omitempty"`
}

type Identifier struct {
	Type           IdentifierType `json:"type"`
	Value          string         `json:"value"`
	IssuingCountry string         `json:"issuing_country,omitempty"`
}

type Address struct {
	AddressText string `json:"address_text,omitempty"`
	City        string `json:"city,omitempty"`
	Region      string `json:"region,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Country     string `json:"country,omitempty"`
}

type Party struct {
	EntityType           EntityType      `json:"entity_type"`
	Name                 Name            `json:"name"`
	DateOfBirth          string          `json:"date_of_birth,omitempty"`
	BirthYear            string          `json:"birth_year,omitempty"`
	IncorporationDate    string          `json:"incorporation_date,omitempty"`
	Countries            []string        `json:"countries,omitempty"`
	Nationalities        []string        `json:"nationalities,omitempty"`
	Addresses            []Address       `json:"addresses,omitempty"`
	Identifiers          []Identifier    `json:"identifiers,omitempty"`
	SourceRecordID       string          `json:"source_record_id,omitempty"`
	SourceSystem         string          `json:"source_system,omitempty"`
	AdditionalAttributes map[string]any  `json:"additional_attributes,omitempty"`
}

type MatchedParty struct {
	ListSource           string         `json:"list_source"`
	Program              string         `json:"program,omitempty"`
	ListUID              string         `json:"list_uid,omitempty"`
	EntityType           EntityType     `json:"entity_type"`
	Name                 Name           `json:"name"`
	DateOfBirth          string         `json:"date_of_birth,omitempty"`
	BirthYear            string         `json:"birth_year,omitempty"`
	IncorporationDate    string         `json:"incorporation_date,omitempty"`
	Countries            []string       `json:"countries,omitempty"`
	Addresses            []Address      `json:"addresses,omitempty"`
	Identifiers          []Identifier   `json:"identifiers,omitempty"`
	AdditionalAttributes map[string]any `json:"additional_attributes,omitempty"`
}

type ScreeningFeatures struct {
	VendorScore    *float64 `json:"vendor_score,omitempty"`
	NameScore      *float64 `json:"name_score,omitempty"`
	MatchFlags     []string `json:"match_flags,omitempty"`
	ReasonCodes    []string `json:"reason_codes,omitempty"`
	AnalystNotes   []string `json:"analyst_notes,omitempty"`
	ReviewComments []string `json:"review_comments,omitempty"`
}

type TransactionContext struct {
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

type ArtifactRef struct {
	Kind        string `json:"kind"`
	URI         string `json:"uri"`
	Description string `json:"description,omitempty"`
}

type CanonicalAlert struct {
	Kind              AlertKind           `json:"kind"`
	Metadata          AlertMetadata       `json:"metadata"`
	ScreenedParty     Party               `json:"screened_party"`
	MatchedParty      MatchedParty        `json:"matched_party"`
	ScreeningFeatures ScreeningFeatures   `json:"screening_features,omitempty"`
	Transaction       *TransactionContext `json:"transaction,omitempty"`
	Artifacts         []ArtifactRef       `json:"artifacts,omitempty"`
}
