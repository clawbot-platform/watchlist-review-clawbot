package ofacdata

import "encoding/xml"

type XMLRoot struct {
    XMLName         xml.Name           `xml:"SanctionsList"`
    DistinctParties []XMLDistinctParty `xml:"DistinctParties>DistinctParty"`
    Profiles        []XMLProfile       `xml:"Profiles>Profile"`
}

type XMLDistinctParty struct {
    ID          string                 `xml:"FixedRef"`
    PartyType   string                 `xml:"DistinctPartyType>Value"`
    Names       []XMLDistinctPartyName `xml:"DistinctPartyName"`
    BirthDates  []XMLDate              `xml:"DateOfBirth"`
    Addresses   []XMLAddress           `xml:"Address"`
    Identifiers []XMLIdentifier        `xml:"IdentityDocument"`
}

type XMLDistinctPartyName struct {
    Primary bool   `xml:"Primary,attr"`
    Value   string `xml:"NamePartValue"`
}

type XMLProfile struct {
    ID             string            `xml:"FixedRef"`
    PartyID        string            `xml:"DistinctPartyFixedRef"`
    ListUID        string            `xml:"ListUniqueID"`
    Programs       []XMLProgram      `xml:"SanctionsProgram"`
    Relationships  []XMLRelationship `xml:"ProfileRelationship"`
    Documents      []XMLDocument     `xml:"IdentityDocument"`
    Addresses      []XMLAddress      `xml:"Address"`
    Features       []XMLFeature      `xml:"Feature"`
}

type XMLProgram struct {
    Value string `xml:"Value"`
}

type XMLRelationship struct {
    ToProfileID  string `xml:"TargetProfileFixedRef"`
    Relationship string `xml:"ProfileRelationshipTypeDescription"`
}

type XMLDocument struct {
    Type    string `xml:"Type"`
    Number  string `xml:"Number"`
    Country string `xml:"IssuingCountry"`
}

type XMLIdentifier struct {
    Type    string `xml:"Type"`
    Number  string `xml:"Number"`
    Country string `xml:"IssuingCountry"`
}

type XMLAddress struct {
    Address1 string `xml:"Address1"`
    City     string `xml:"City"`
    State    string `xml:"StateOrProvince"`
    Country  string `xml:"Country"`
}

type XMLFeature struct {
    Type  string `xml:"Type"`
    Value string `xml:"Value"`
}

type XMLDate struct {
    Value string `xml:"Date"`
}
