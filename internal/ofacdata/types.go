package ofacdata

import "time"

type PartyType string

const (
	PartyTypeIndividual   PartyType = "individual"
	PartyTypeOrganization PartyType = "organization"
	PartyTypeVessel       PartyType = "vessel"
	PartyTypeAircraft     PartyType = "aircraft"
)

type Party struct {
	PartyID         string
	ListUID         string
	PrimaryName     string
	Aliases         []string
	PartyType       PartyType
	Programs        []string
	DistinctPartyID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Identifier struct {
	PartyID         string
	IdentifierType  string
	IdentifierValue string
	IssuingCountry  string
}

type Document struct {
	PartyID       string
	DocumentType  string
	DocumentValue string
	Country       string
}

type Location struct {
	PartyID      string
	CountryCode  string
	City         string
	AddressText  string
	ProvinceOrSt string
}

type Relationship struct {
	RelationshipID string
	FromPartyID    string
	ToPartyID      string
	Relationship   string
	SourceListUID  string
}

type SanctionsEntry struct {
	EntryID     string
	PartyID     string
	Program     string
	ListSource  string
	PublishedAt time.Time
}

type Snapshot struct {
	Parties         []Party
	Identifiers     []Identifier
	Documents       []Document
	Locations       []Location
	Relationships   []Relationship
	SanctionsEntrys []SanctionsEntry
}
