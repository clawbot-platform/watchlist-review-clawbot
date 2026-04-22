package ofacdata

import (
    "fmt"
    "strings"
    "time"
)

func NormalizeXML(root XMLRoot) Snapshot {
    now := time.Now().UTC()
    snap := Snapshot{}

    partiesByDistinct := map[string]*Party{}
    profileToParty := map[string]string{}

    for _, dp := range root.DistinctParties {
        partyID := strings.TrimSpace(dp.ID)
        if partyID == "" {
            continue
        }
        primary := firstNonEmptyDistinctName(dp.Names)
        party := Party{
            PartyID:         partyID,
            ListUID:         partyID,
            PrimaryName:     primary,
            Aliases:         collectDistinctAliases(dp.Names, primary),
            PartyType:       normalizePartyType(dp.PartyType),
            DistinctPartyID: partyID,
            CreatedAt:       now,
            UpdatedAt:       now,
        }
        partiesByDistinct[partyID] = &party

        for _, ident := range dp.Identifiers {
            if strings.TrimSpace(ident.Number) == "" {
                continue
            }
            snap.Identifiers = append(snap.Identifiers, Identifier{
                PartyID:         partyID,
                IdentifierType:  normalizeLabel(ident.Type),
                IdentifierValue: normalizeValue(ident.Number),
                IssuingCountry:  normalizeCountry(ident.Country),
            })
        }
        for _, addr := range dp.Addresses {
            snap.Locations = append(snap.Locations, Location{
                PartyID:      partyID,
                CountryCode:  normalizeCountry(addr.Country),
                City:         normalizeLabel(addr.City),
                AddressText:  strings.TrimSpace(strings.Join([]string{addr.Address1, addr.City, addr.State, addr.Country}, ", ")),
                ProvinceOrSt: normalizeLabel(addr.State),
            })
        }
    }

    for _, profile := range root.Profiles {
        partyID := strings.TrimSpace(profile.PartyID)
        if partyID == "" {
            partyID = strings.TrimSpace(profile.ID)
        }
        if partyID == "" {
            continue
        }
        if _, ok := partiesByDistinct[partyID]; !ok {
            partiesByDistinct[partyID] = &Party{
                PartyID:         partyID,
                ListUID:         strings.TrimSpace(profile.ListUID),
                PrimaryName:     strings.TrimSpace(profile.ID),
                PartyType:       PartyTypeOrganization,
                DistinctPartyID: partyID,
                CreatedAt:       now,
                UpdatedAt:       now,
            }
        }

        profileToParty[strings.TrimSpace(profile.ID)] = partyID
        if partiesByDistinct[partyID].ListUID == "" {
            partiesByDistinct[partyID].ListUID = strings.TrimSpace(profile.ListUID)
        }

        for _, program := range profile.Programs {
            p := normalizeLabel(program.Value)
            if p == "" {
                continue
            }
            snap.SanctionsEntrys = append(snap.SanctionsEntrys, SanctionsEntry{
                EntryID:     fmt.Sprintf("%s|%s", partyID, p),
                PartyID:     partyID,
                Program:     p,
                ListSource:  "OFAC_XML",
                PublishedAt: now,
            })
        }
        for _, doc := range profile.Documents {
            if strings.TrimSpace(doc.Number) == "" {
                continue
            }
            snap.Documents = append(snap.Documents, Document{
                PartyID:       partyID,
                DocumentType:  normalizeLabel(doc.Type),
                DocumentValue: normalizeValue(doc.Number),
                Country:       normalizeCountry(doc.Country),
            })
        }
        for _, addr := range profile.Addresses {
            snap.Locations = append(snap.Locations, Location{
                PartyID:      partyID,
                CountryCode:  normalizeCountry(addr.Country),
                City:         normalizeLabel(addr.City),
                AddressText:  strings.TrimSpace(strings.Join([]string{addr.Address1, addr.City, addr.State, addr.Country}, ", ")),
                ProvinceOrSt: normalizeLabel(addr.State),
            })
        }
    }

    for _, profile := range root.Profiles {
        from := profileToParty[strings.TrimSpace(profile.ID)]
        if from == "" {
            continue
        }
        for _, rel := range profile.Relationships {
            to := profileToParty[strings.TrimSpace(rel.ToProfileID)]
            if to == "" {
                continue
            }
            relationship := normalizeLabel(rel.Relationship)
            snap.Relationships = append(snap.Relationships, Relationship{
                RelationshipID: fmt.Sprintf("%s|%s|%s", from, to, relationship),
                FromPartyID:    from,
                ToPartyID:      to,
                Relationship:   relationship,
                SourceListUID:  partiesByDistinct[from].ListUID,
            })
        }
    }

    for _, party := range partiesByDistinct {
        snap.Parties = append(snap.Parties, *party)
    }

    return dedupeSnapshot(snap)
}

func firstNonEmptyDistinctName(names []XMLDistinctPartyName) string {
    for _, name := range names {
        if name.Primary && strings.TrimSpace(name.Value) != "" {
            return strings.TrimSpace(name.Value)
        }
    }
    for _, name := range names {
        if strings.TrimSpace(name.Value) != "" {
            return strings.TrimSpace(name.Value)
        }
    }
    return ""
}

func collectDistinctAliases(names []XMLDistinctPartyName, primary string) []string {
    seen := map[string]struct{}{}
    var out []string
    for _, name := range names {
        value := strings.TrimSpace(name.Value)
        if value == "" || strings.EqualFold(value, primary) {
            continue
        }
        key := strings.ToLower(value)
        if _, ok := seen[key]; ok {
            continue
        }
        seen[key] = struct{}{}
        out = append(out, value)
    }
    return out
}

func normalizePartyType(v string) PartyType {
    switch strings.ToLower(strings.TrimSpace(v)) {
    case "individual", "person":
        return PartyTypeIndividual
    case "vessel":
        return PartyTypeVessel
    case "aircraft":
        return PartyTypeAircraft
    default:
        return PartyTypeOrganization
    }
}

func normalizeCountry(v string) string { return strings.ToUpper(strings.TrimSpace(v)) }
func normalizeLabel(v string) string { return strings.TrimSpace(v) }
func normalizeValue(v string) string { return strings.ToUpper(strings.TrimSpace(v)) }

func dedupeSnapshot(s Snapshot) Snapshot {
    seenParty := map[string]struct{}{}
    var parties []Party
    for _, p := range s.Parties {
        if _, ok := seenParty[p.PartyID]; ok { continue }
        seenParty[p.PartyID] = struct{}{}
        parties = append(parties, p)
    }
    s.Parties = parties

    seenIdentifier := map[string]struct{}{}
    var identifiers []Identifier
    for _, x := range s.Identifiers {
        key := x.PartyID + "|" + x.IdentifierType + "|" + x.IdentifierValue
        if _, ok := seenIdentifier[key]; ok { continue }
        seenIdentifier[key] = struct{}{}
        identifiers = append(identifiers, x)
    }
    s.Identifiers = identifiers

    seenDocument := map[string]struct{}{}
    var docs []Document
    for _, x := range s.Documents {
        key := x.PartyID + "|" + x.DocumentType + "|" + x.DocumentValue
        if _, ok := seenDocument[key]; ok { continue }
        seenDocument[key] = struct{}{}
        docs = append(docs, x)
    }
    s.Documents = docs

    seenLocation := map[string]struct{}{}
    var locations []Location
    for _, x := range s.Locations {
        key := x.PartyID + "|" + x.CountryCode + "|" + x.City + "|" + x.AddressText
        if _, ok := seenLocation[key]; ok { continue }
        seenLocation[key] = struct{}{}
        locations = append(locations, x)
    }
    s.Locations = locations

    seenRel := map[string]struct{}{}
    var rels []Relationship
    for _, x := range s.Relationships {
        if _, ok := seenRel[x.RelationshipID]; ok { continue }
        seenRel[x.RelationshipID] = struct{}{}
        rels = append(rels, x)
    }
    s.Relationships = rels

    seenEntry := map[string]struct{}{}
    var entries []SanctionsEntry
    for _, x := range s.SanctionsEntrys {
        if _, ok := seenEntry[x.EntryID]; ok { continue }
        seenEntry[x.EntryID] = struct{}{}
        entries = append(entries, x)
    }
    s.SanctionsEntrys = entries

    return s
}
