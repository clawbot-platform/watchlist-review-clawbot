package ofacdata

// XML model stubs for the relationship-bearing OFAC feed.
//
// These are intentionally incomplete starter structs. The real repo should
// refine them against the exact XML/SLS payloads being ingested.
type XMLRoot struct {
	DistinctParties []XMLDistinctParty `xml:"DistinctParties>DistinctParty"`
	Profiles        []XMLProfile       `xml:"Profiles>Profile"`
}

type XMLDistinctParty struct {
	ID    string `xml:"FixedRef"`
	Names []struct {
		Value string `xml:"NamePartValue"`
	} `xml:"DistinctPartyName"`
}

type XMLProfile struct {
	ID           string               `xml:"FixedRef"`
	PartyID      string               `xml:"DistinctPartyFixedRef"`
	Relationships []XMLRelationship   `xml:"ProfileRelationship"`
}

type XMLRelationship struct {
	ToProfileID   string `xml:"TargetProfileFixedRef"`
	Relationship  string `xml:"ProfileRelationshipTypeDescription"`
}
