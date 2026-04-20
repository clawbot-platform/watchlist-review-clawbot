package scoring

import "github.com/clawbot-platform/watchlist-review-clawbot/internal/features"

func computeDataSufficiency(fx *features.ExtractedFeatures, result *Result) {
	if fx == nil || result == nil {
		return
	}

	result.ScreenedCompletenessScore = scoreScreenedCompleteness(fx)
	result.MatchedCompletenessScore = scoreMatchedCompleteness(fx)
	result.IdentifierQualityScore = scoreIdentifierQuality(fx)
	result.GeographyQualityScore = scoreGeographyQuality(fx)
	result.SupportingContextScore = scoreSupportingContext(fx)

	total := result.ScreenedCompletenessScore +
		result.MatchedCompletenessScore +
		result.IdentifierQualityScore +
		result.GeographyQualityScore +
		result.SupportingContextScore

	if total > 100 {
		total = 100
	}
	result.DataSufficiencyScore = total
}

func scoreScreenedCompleteness(fx *features.ExtractedFeatures) int {
	score := 0
	if fx.ScreenedCompleteness.NamePresent {
		score += 8
	}
	if fx.ScreenedCompleteness.AliasCount > 0 {
		score += 4
	}
	if fx.ScreenedCompleteness.DatePresent {
		score += dateCompletenessPoints(fx)
	}
	if fx.ScreenedCompleteness.CountryCount > 0 {
		score += countryCompletenessPoints(fx)
	}
	if fx.ScreenedCompleteness.AddressCount > 0 {
		score += addressCompletenessPoints(fx)
	}
	if fx.ScreenedCompleteness.IdentifierCount > 0 {
		score += identifierCompletenessPoints(fx)
	}
	if score > 30 {
		score = 30
	}
	return score
}

func scoreMatchedCompleteness(fx *features.ExtractedFeatures) int {
	score := 0
	if fx.MatchedCompleteness.NamePresent {
		score += 6
	}
	if fx.MatchedCompleteness.AliasCount > 0 {
		score += 4
	}
	if fx.MatchedCompleteness.DatePresent {
		score += matchedDateCompletenessPoints(fx)
	}
	if fx.MatchedCompleteness.CountryCount > 0 {
		score += matchedCountryCompletenessPoints(fx)
	}
	if fx.MatchedCompleteness.AddressCount > 0 {
		score += matchedAddressCompletenessPoints(fx)
	}
	if fx.MatchedCompleteness.IdentifierCount > 0 {
		score += 6
	}
	if score > 25 {
		score = 25
	}
	return score
}

func scoreIdentifierQuality(fx *features.ExtractedFeatures) int {
	switch {
	case len(fx.Identifiers.ExactMatches) > 0:
		return 20
	case len(fx.Identifiers.ScreenedByType) > 0 && len(fx.Identifiers.MatchedByType) > 0:
		return 12
	case len(fx.Identifiers.ScreenedByType) > 0 || len(fx.Identifiers.MatchedByType) > 0:
		return 8
	default:
		return 0
	}
}

func scoreGeographyQuality(fx *features.ExtractedFeatures) int {
	switch {
	case len(fx.Geography.CountryOverlap) > 0 && len(fx.Geography.AddressCountries) > 0:
		return 10
	case len(fx.Geography.CountryOverlap) > 0:
		return 8
	case len(fx.Geography.ScreenedCountries) > 0 || len(fx.Geography.MatchedCountries) > 0:
		return 4
	case len(fx.Geography.AddressCountries) > 0:
		return 3
	default:
		return 0
	}
}

func scoreSupportingContext(fx *features.ExtractedFeatures) int {
	score := 0
	if len(fx.MatchFlags) > 0 {
		score += 3
	}
	if fx.VendorScore != nil {
		score += 2
	}
	if fx.NameScore != nil {
		score += 2
	}
	if fx.ScreenedCompleteness.SupportingNoteCount > 0 || fx.MatchedCompleteness.SupportingNoteCount > 0 {
		score += 6
	}
	if fx.Context.IsACH {
		score += 2
	}
	if score > 15 {
		score = 15
	}
	return score
}

func dateCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "organization_onboarding" {
		return 5
	}
	return 6
}

func countryCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "ach_party" {
		return 5
	}
	return 4
}

func addressCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "organization_onboarding" {
		return 5
	}
	if fx.Kind == "ach_party" {
		return 3
	}
	return 4
}

func identifierCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "ach_party" {
		return 7
	}
	return 4
}

func matchedDateCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "organization_onboarding" {
		return 4
	}
	return 5
}

func matchedCountryCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "organization_onboarding" || fx.Kind == "ach_party" {
		return 5
	}
	return 4
}

func matchedAddressCompletenessPoints(fx *features.ExtractedFeatures) int {
	if fx.Kind == "ach_party" {
		return 0
	}
	return 4
}
