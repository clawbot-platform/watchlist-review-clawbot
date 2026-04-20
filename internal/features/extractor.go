package features

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

type ScriptFamily string

const (
	ScriptFamilyUnknown  ScriptFamily = "unknown"
	ScriptFamilyLatin    ScriptFamily = "latin"
	ScriptFamilyCyrillic ScriptFamily = "cyrillic"
	ScriptFamilyArabic   ScriptFamily = "arabic_persian"
	ScriptFamilyHan      ScriptFamily = "han"
	ScriptFamilyMixed    ScriptFamily = "mixed"
)

type NameFeatures struct {
	Original           string       `json:"original"`
	Canonical          string       `json:"canonical"`
	CanonicalAliases   []string     `json:"canonical_aliases,omitempty"`
	NativeCanonical    string       `json:"native_canonical,omitempty"`
	Tokens             []string     `json:"tokens,omitempty"`
	TokenSorted        string       `json:"token_sorted,omitempty"`
	ScriptFamily       ScriptFamily `json:"script_family"`
	HasExactAliasMatch bool         `json:"has_exact_alias_match"`
}

type DateFeatures struct {
	ScreenedExact string `json:"screened_exact,omitempty"`
	ScreenedYear  string `json:"screened_year,omitempty"`
	MatchedExact  string `json:"matched_exact,omitempty"`
	MatchedYear   string `json:"matched_year,omitempty"`
	ExactMatch    bool   `json:"exact_match"`
	YearMatch     bool   `json:"year_match"`
	HasConflict   bool   `json:"has_conflict"`
}

type IdentifierMatch struct {
	Type  alerts.IdentifierType `json:"type"`
	Value string                `json:"value"`
}

type IdentifierFeatures struct {
	ScreenedByType     map[alerts.IdentifierType][]string `json:"screened_by_type,omitempty"`
	MatchedByType      map[alerts.IdentifierType][]string `json:"matched_by_type,omitempty"`
	ExactMatches       []IdentifierMatch                  `json:"exact_matches,omitempty"`
	PotentialConflicts []string                           `json:"potential_conflicts,omitempty"`
}

type GeographyFeatures struct {
	ScreenedCountries []string `json:"screened_countries,omitempty"`
	MatchedCountries  []string `json:"matched_countries,omitempty"`
	CountryOverlap    []string `json:"country_overlap,omitempty"`
	AddressCountries  []string `json:"address_countries,omitempty"`
	HasCountrySupport bool     `json:"has_country_support"`
}

type ContextFeatures struct {
	ScreenedEntityType alerts.EntityType `json:"screened_entity_type"`
	MatchedEntityType  alerts.EntityType `json:"matched_entity_type"`
	Kind               alerts.AlertKind  `json:"kind"`
	IsACH              bool              `json:"is_ach"`
	RailType           string            `json:"rail_type,omitempty"`
	EntityTypeConflict bool              `json:"entity_type_conflict"`
}

type CompletenessFeatures struct {
	NamePresent         bool `json:"name_present"`
	AliasCount          int  `json:"alias_count"`
	DatePresent         bool `json:"date_present"`
	CountryCount        int  `json:"country_count"`
	AddressCount        int  `json:"address_count"`
	IdentifierCount     int  `json:"identifier_count"`
	SupportingNoteCount int  `json:"supporting_note_count"`
}

type ExtractedFeatures struct {
	AlertID              string               `json:"alert_id"`
	Kind                 alerts.AlertKind     `json:"kind"`
	ScreenedName         NameFeatures         `json:"screened_name"`
	MatchedName          NameFeatures         `json:"matched_name"`
	Date                 DateFeatures         `json:"date"`
	Identifiers          IdentifierFeatures   `json:"identifiers"`
	Geography            GeographyFeatures    `json:"geography"`
	Context              ContextFeatures      `json:"context"`
	ScreenedCompleteness CompletenessFeatures `json:"screened_completeness"`
	MatchedCompleteness  CompletenessFeatures `json:"matched_completeness"`
	MissingFields        []string             `json:"missing_fields,omitempty"`
	Contradictions       []string             `json:"contradictions,omitempty"`
	VendorScore          *float64             `json:"vendor_score,omitempty"`
	NameScore            *float64             `json:"name_score,omitempty"`
	MatchFlags           []string             `json:"match_flags,omitempty"`
}

func Extract(alert *alerts.CanonicalAlert) (*ExtractedFeatures, error) {
	if alert == nil { return nil, fmt.Errorf("alert is required") }
	alert.Normalize()
	if err := alert.Validate(); err != nil { return nil, err }

	out := &ExtractedFeatures{
		AlertID: alert.Metadata.AlertID,
		Kind: alert.Kind,
		ScreenedName: buildNameFeatures(alert.ScreenedParty.Name, alert.MatchedParty.Name),
		MatchedName: buildNameFeatures(alert.MatchedParty.Name, alert.ScreenedParty.Name),
		Date: buildDateFeatures(alert),
		Identifiers: buildIdentifierFeatures(alert),
		Geography: buildGeographyFeatures(alert),
		Context: buildContextFeatures(alert),
		ScreenedCompleteness: buildPartyCompleteness(alert.ScreenedParty, alert.ScreeningFeatures),
		MatchedCompleteness: buildMatchedCompleteness(alert.MatchedParty, alert.ScreeningFeatures),
		VendorScore: alert.ScreeningFeatures.VendorScore,
		NameScore: alert.ScreeningFeatures.NameScore,
		MatchFlags: append([]string(nil), alert.ScreeningFeatures.MatchFlags...),
	}
	out.MissingFields = collectMissingFields(alert)
	out.Contradictions = collectContradictions(out)
	return out, nil
}

func buildNameFeatures(name alerts.Name, counterpart alerts.Name) NameFeatures {
	canonical := canonicalName(name.FullName)
	var aliases []string
	for _, alias := range name.Aliases {
		if c := canonicalName(alias); c != "" { aliases = append(aliases, c) }
	}
	sort.Strings(aliases)
	counterpartCanonical := canonicalName(counterpart.FullName)
	hasExactAliasMatch := false
	for _, alias := range aliases {
		if alias == counterpartCanonical && alias != "" { hasExactAliasMatch = true; break }
	}
	tokens := tokenizeCanonical(canonical)
	return NameFeatures{
		Original: name.FullName,
		Canonical: canonical,
		CanonicalAliases: aliases,
		NativeCanonical: canonicalName(name.NativeName),
		Tokens: tokens,
		TokenSorted: tokenSort(tokens),
		ScriptFamily: inferScriptFamily(firstNonEmpty(name.NativeName, name.FullName)),
		HasExactAliasMatch: hasExactAliasMatch,
	}
}

func buildDateFeatures(alert *alerts.CanonicalAlert) DateFeatures {
	screenedExact := firstNonEmpty(alert.ScreenedParty.DateOfBirth, alert.ScreenedParty.IncorporationDate)
	matchedExact := firstNonEmpty(alert.MatchedParty.DateOfBirth, alert.MatchedParty.IncorporationDate)
	screenedYear := firstNonEmpty(alert.ScreenedParty.BirthYear, yearFromDate(screenedExact))
	matchedYear := firstNonEmpty(alert.MatchedParty.BirthYear, yearFromDate(matchedExact))
	exact := screenedExact != "" && matchedExact != "" && screenedExact == matchedExact
	yearMatch := screenedYear != "" && matchedYear != "" && screenedYear == matchedYear
	conflict := screenedExact != "" && matchedExact != "" && screenedExact != matchedExact && !exact
	return DateFeatures{ScreenedExact: screenedExact, ScreenedYear: screenedYear, MatchedExact: matchedExact, MatchedYear: matchedYear, ExactMatch: exact, YearMatch: yearMatch, HasConflict: conflict && !yearMatch}
}

func buildIdentifierFeatures(alert *alerts.CanonicalAlert) IdentifierFeatures {
	screened := bucketIdentifiers(alert.ScreenedParty.Identifiers)
	matched := bucketIdentifiers(alert.MatchedParty.Identifiers)
	var exact []IdentifierMatch
	var conflicts []string
	for idType, screenedValues := range screened {
		matchedValues := matched[idType]
		if len(matchedValues) == 0 { continue }
		matchedSet := make(map[string]struct{}, len(matchedValues))
		for _, v := range matchedValues { matchedSet[v] = struct{}{} }
		matchesForType := false
		for _, v := range screenedValues {
			if _, ok := matchedSet[v]; ok {
				exact = append(exact, IdentifierMatch{Type: idType, Value: v})
				matchesForType = true
			}
		}
		if !matchesForType && len(screenedValues) > 0 && len(matchedValues) > 0 {
			conflicts = append(conflicts, string(idType))
		}
	}
	sort.Slice(exact, func(i, j int) bool {
		if exact[i].Type == exact[j].Type { return exact[i].Value < exact[j].Value }
		return exact[i].Type < exact[j].Type
	})
	sort.Strings(conflicts)
	return IdentifierFeatures{ScreenedByType: screened, MatchedByType: matched, ExactMatches: exact, PotentialConflicts: uniqueStrings(conflicts)}
}

func buildGeographyFeatures(alert *alerts.CanonicalAlert) GeographyFeatures {
	screenedCountries := append([]string(nil), alert.ScreenedParty.Countries...)
	matchedCountries := append([]string(nil), alert.MatchedParty.Countries...)
	var addressCountries []string
	for _, addr := range alert.ScreenedParty.Addresses {
		if addr.Country != "" { addressCountries = append(addressCountries, strings.ToUpper(strings.TrimSpace(addr.Country))) }
	}
	for _, addr := range alert.MatchedParty.Addresses {
		if addr.Country != "" { addressCountries = append(addressCountries, strings.ToUpper(strings.TrimSpace(addr.Country))) }
	}
	screenedSet := sliceSet(screenedCountries)
	var overlap []string
	for _, c := range matchedCountries {
		if _, ok := screenedSet[c]; ok { overlap = append(overlap, c) }
	}
	overlap = uniqueStrings(overlap)
	sort.Strings(overlap)
	addressCountries = uniqueStrings(addressCountries)
	sort.Strings(addressCountries)
	return GeographyFeatures{ScreenedCountries: uniqueStrings(screenedCountries), MatchedCountries: uniqueStrings(matchedCountries), CountryOverlap: overlap, AddressCountries: addressCountries, HasCountrySupport: len(overlap) > 0}
}

func buildContextFeatures(alert *alerts.CanonicalAlert) ContextFeatures {
	out := ContextFeatures{ScreenedEntityType: alert.ScreenedParty.EntityType, MatchedEntityType: alert.MatchedParty.EntityType, Kind: alert.Kind}
	if alert.Transaction != nil {
		out.IsACH = strings.EqualFold(alert.Transaction.RailType, "ach")
		out.RailType = alert.Transaction.RailType
	}
	if alert.ScreenedParty.EntityType != alerts.EntityTypeUnknown &&
		alert.MatchedParty.EntityType != alerts.EntityTypeUnknown &&
		alert.ScreenedParty.EntityType != alert.MatchedParty.EntityType {
		out.EntityTypeConflict = true
	}
	return out
}

func buildPartyCompleteness(p alerts.Party, sf alerts.ScreeningFeatures) CompletenessFeatures {
	return CompletenessFeatures{
		NamePresent: strings.TrimSpace(p.Name.FullName) != "",
		AliasCount: len(p.Name.Aliases),
		DatePresent: firstNonEmpty(p.DateOfBirth, p.IncorporationDate, p.BirthYear) != "",
		CountryCount: len(p.Countries),
		AddressCount: len(p.Addresses),
		IdentifierCount: len(p.Identifiers),
		SupportingNoteCount: len(sf.AnalystNotes) + len(sf.ReviewComments),
	}
}

func buildMatchedCompleteness(p alerts.MatchedParty, sf alerts.ScreeningFeatures) CompletenessFeatures {
	return CompletenessFeatures{
		NamePresent: strings.TrimSpace(p.Name.FullName) != "",
		AliasCount: len(p.Name.Aliases),
		DatePresent: firstNonEmpty(p.DateOfBirth, p.IncorporationDate, p.BirthYear) != "",
		CountryCount: len(p.Countries),
		AddressCount: len(p.Addresses),
		IdentifierCount: len(p.Identifiers),
		SupportingNoteCount: len(sf.AnalystNotes) + len(sf.ReviewComments),
	}
}

func collectMissingFields(alert *alerts.CanonicalAlert) []string {
	var missing []string
	if strings.TrimSpace(alert.ScreenedParty.Name.FullName) == "" { missing = append(missing, "screened_party.name.full_name") }
	if firstNonEmpty(alert.ScreenedParty.DateOfBirth, alert.ScreenedParty.IncorporationDate, alert.ScreenedParty.BirthYear) == "" { missing = append(missing, "screened_party.date") }
	if len(alert.ScreenedParty.Identifiers) == 0 { missing = append(missing, "screened_party.identifiers") }
	if len(alert.ScreenedParty.Countries) == 0 { missing = append(missing, "screened_party.countries") }
	if strings.TrimSpace(alert.MatchedParty.Name.FullName) == "" { missing = append(missing, "matched_party.name.full_name") }
	if firstNonEmpty(alert.MatchedParty.DateOfBirth, alert.MatchedParty.IncorporationDate, alert.MatchedParty.BirthYear) == "" { missing = append(missing, "matched_party.date") }
	if len(alert.MatchedParty.Identifiers) == 0 { missing = append(missing, "matched_party.identifiers") }
	if len(alert.MatchedParty.Countries) == 0 { missing = append(missing, "matched_party.countries") }
	return uniqueStrings(missing)
}

func collectContradictions(out *ExtractedFeatures) []string {
	var contradictions []string
	if out.Date.HasConflict { contradictions = append(contradictions, "dob_conflict") }
	if len(out.Identifiers.PotentialConflicts) > 0 { contradictions = append(contradictions, "identifier_conflict") }
	if out.Context.EntityTypeConflict { contradictions = append(contradictions, "entity_type_conflict") }
	if len(out.Geography.ScreenedCountries) > 0 && len(out.Geography.MatchedCountries) > 0 && len(out.Geography.CountryOverlap) == 0 { contradictions = append(contradictions, "geography_conflict") }
	return uniqueStrings(contradictions)
}

func bucketIdentifiers(ids []alerts.Identifier) map[alerts.IdentifierType][]string {
	if len(ids) == 0 { return nil }
	out := map[alerts.IdentifierType][]string{}
	for _, id := range ids {
		if id.Type == alerts.IdentifierTypeUnknown || strings.TrimSpace(id.Value) == "" { continue }
		canon := canonicalIdentifier(id.Value)
		if canon == "" { continue }
		out[id.Type] = append(out[id.Type], canon)
	}
	for k, vals := range out {
		out[k] = uniqueStrings(vals)
		sort.Strings(out[k])
	}
	return out
}

func canonicalName(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" { return "" }
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastSpace = false
		default:
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func tokenizeCanonical(s string) []string {
	if s == "" { return nil }
	return strings.Fields(s)
}

func tokenSort(tokens []string) string {
	if len(tokens) == 0 { return "" }
	copied := append([]string(nil), tokens...)
	sort.Strings(copied)
	return strings.Join(copied, " ")
}

func inferScriptFamily(s string) ScriptFamily {
	var hasLatin, hasCyrillic, hasArabic, hasHan bool
	for _, r := range s {
		switch {
		case unicode.In(r, unicode.Latin):
			hasLatin = true
		case unicode.In(r, unicode.Cyrillic):
			hasCyrillic = true
		case unicode.In(r, unicode.Arabic):
			hasArabic = true
		case unicode.In(r, unicode.Han):
			hasHan = true
		}
	}
	count := 0
	for _, v := range []bool{hasLatin, hasCyrillic, hasArabic, hasHan} {
		if v { count++ }
	}
	switch {
	case count == 0:
		return ScriptFamilyUnknown
	case count > 1:
		return ScriptFamilyMixed
	case hasLatin:
		return ScriptFamilyLatin
	case hasCyrillic:
		return ScriptFamilyCyrillic
	case hasArabic:
		return ScriptFamilyArabic
	case hasHan:
		return ScriptFamilyHan
	default:
		return ScriptFamilyUnknown
	}
}

func canonicalIdentifier(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" { return "" }
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) { b.WriteRune(r) }
	}
	return b.String()
}

func yearFromDate(v string) string {
	v = strings.TrimSpace(v)
	if len(v) < 4 { return "" }
	year := v[:4]
	if _, err := strconv.Atoi(year); err != nil { return "" }
	return year
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" { return t }
	}
	return ""
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" { continue }
		if _, ok := seen[v]; ok { continue }
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 { return nil }
	return out
}

func sliceSet(in []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range in {
		if t := strings.TrimSpace(v); t != "" { out[t] = struct{}{} }
	}
	return out
}
