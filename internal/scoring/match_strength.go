package scoring

import (
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
)

func computeMatchStrength(fx *features.ExtractedFeatures, result *Result) {
	if fx == nil || result == nil {
		return
	}

	result.NameMatchScore, result.EvidenceFor = scoreName(fx, result.EvidenceFor)
	result.DateMatchScore, result.EvidenceFor, result.EvidenceAgainst = scoreDate(fx, result.EvidenceFor, result.EvidenceAgainst)
	result.IdentifierMatchScore, result.EvidenceFor, result.EvidenceAgainst = scoreIdentifiers(fx, result.EvidenceFor, result.EvidenceAgainst)
	result.GeographyMatchScore, result.EvidenceFor, result.EvidenceAgainst = scoreGeography(fx, result.EvidenceFor, result.EvidenceAgainst)
	result.AddressMatchScore, result.EvidenceFor = scoreAddress(fx, result.EvidenceFor)
	result.ContextSupportScore, result.EvidenceFor, result.EvidenceAgainst = scoreContext(fx, result.EvidenceFor, result.EvidenceAgainst)

	total := result.NameMatchScore +
		result.DateMatchScore +
		result.IdentifierMatchScore +
		result.GeographyMatchScore +
		result.AddressMatchScore +
		result.ContextSupportScore

	if total > 100 {
		total = 100
	}
	if total > 69 && !hasNonNameCorroboration(result) {
		total = 69
	}

	result.MatchStrengthScore = total
}

func scoreName(fx *features.ExtractedFeatures, evidenceFor []string) (int, []string) {
	switch {
	case fx.ScreenedName.Canonical != "" && fx.ScreenedName.Canonical == fx.MatchedName.Canonical:
		return 35, append(evidenceFor, "exact normalized name match")

	case fx.ScreenedName.HasExactAliasMatch || fx.MatchedName.HasExactAliasMatch:
		return 30, append(evidenceFor, "exact alias-to-name match")

	case fx.ScreenedName.NativeCanonical != "" && fx.ScreenedName.NativeCanonical == fx.MatchedName.NativeCanonical:
		return 28, append(evidenceFor, "native-script exact match")

	case fx.ScreenedName.TokenSorted != "" && fx.ScreenedName.TokenSorted == fx.MatchedName.TokenSorted:
		return 20, append(evidenceFor, "reordered or partial strong token match")

	default:
		overlap := tokenOverlapRatio(fx.ScreenedName.Tokens, fx.MatchedName.Tokens)

		// Organization names often need a slightly more conservative-but-reviewable
		// treatment when they share meaningful business tokens, even without strong
		// corroboration. This keeps org-overlap cases from being scored as pure noise.
		if string(fx.Kind) == "organization_onboarding" && hasMeaningfulOrganizationTokenOverlap(fx) {
			switch {
			case overlap >= 0.60:
				return 10, append(evidenceFor, "organization name overlap")
			case overlap > 0:
				return 6, append(evidenceFor, "organization token overlap")
			}
		}

		switch {
		case overlap >= 0.90:
			return 12, append(evidenceFor, "strong token overlap")
		case overlap >= 0.60:
			return 8, append(evidenceFor, "partial token overlap")
		case overlap > 0:
			return 4, evidenceFor
		default:
			return 0, evidenceFor
		}
	}
}

func scoreDate(fx *features.ExtractedFeatures, evidenceFor, evidenceAgainst []string) (int, []string, []string) {
	if fx.Date.ExactMatch {
		return 20, append(evidenceFor, "exact date match"), evidenceAgainst
	}
	if fx.Date.YearMatch {
		return 10, append(evidenceFor, "birth year match"), evidenceAgainst
	}
	if fx.Date.HasConflict {
		evidenceAgainst = append(evidenceAgainst, "date conflict")
	}
	return 0, evidenceFor, evidenceAgainst
}

func scoreIdentifiers(fx *features.ExtractedFeatures, evidenceFor, evidenceAgainst []string) (int, []string, []string) {
	if len(fx.Identifiers.ExactMatches) > 0 {
		match := fx.Identifiers.ExactMatches[0]
		evidenceFor = append(evidenceFor, "exact identifier match on "+string(match.Type))
		return 20, evidenceFor, evidenceAgainst
	}
	if len(fx.Identifiers.PotentialConflicts) > 0 {
		evidenceAgainst = append(evidenceAgainst, "identifier conflict")
	}
	if len(fx.Identifiers.ScreenedByType) > 0 && len(fx.Identifiers.MatchedByType) > 0 {
		return 6, evidenceFor, evidenceAgainst
	}
	return 0, evidenceFor, evidenceAgainst
}

func scoreGeography(fx *features.ExtractedFeatures, evidenceFor, evidenceAgainst []string) (int, []string, []string) {
	if fx.Geography.HasCountrySupport {
		evidenceFor = append(evidenceFor, "country support")
		return 8, evidenceFor, evidenceAgainst
	}
	if len(fx.Geography.ScreenedCountries) > 0 && len(fx.Geography.MatchedCountries) > 0 {
		evidenceAgainst = append(evidenceAgainst, "country mismatch")
	}
	return 0, evidenceFor, evidenceAgainst
}

func scoreAddress(fx *features.ExtractedFeatures, evidenceFor []string) (int, []string) {
	if fx.ScreenedCompleteness.AddressCount > 0 && fx.MatchedCompleteness.AddressCount > 0 && fx.Geography.HasCountrySupport {
		return 8, append(evidenceFor, "address/geography support")
	}
	if fx.ScreenedCompleteness.AddressCount > 0 && fx.Geography.HasCountrySupport {
		return 5, append(evidenceFor, "address support")
	}
	return 0, evidenceFor
}

func scoreContext(fx *features.ExtractedFeatures, evidenceFor, evidenceAgainst []string) (int, []string, []string) {
	if fx.Context.EntityTypeConflict {
		evidenceAgainst = append(evidenceAgainst, "entity type conflict")
		return 0, evidenceFor, evidenceAgainst
	}
	if fx.Context.ScreenedEntityType != "" && fx.Context.ScreenedEntityType == fx.Context.MatchedEntityType {
		return 5, append(evidenceFor, "entity type support"), evidenceAgainst
	}
	return 2, evidenceFor, evidenceAgainst
}

func hasNonNameCorroboration(result *Result) bool {
	return result.DateMatchScore > 0 ||
		result.IdentifierMatchScore > 0 ||
		result.GeographyMatchScore > 0 ||
		result.AddressMatchScore > 0
}

func tokenOverlapRatio(left, right []string) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	leftSet := map[string]struct{}{}
	for _, token := range left {
		leftSet[strings.TrimSpace(token)] = struct{}{}
	}
	matches := 0
	seen := map[string]struct{}{}
	for _, token := range right {
		token = strings.TrimSpace(token)
		if _, done := seen[token]; done {
			continue
		}
		if _, ok := leftSet[token]; ok {
			matches++
			seen[token] = struct{}{}
		}
	}
	maxLen := len(left)
	if len(right) > maxLen {
		maxLen = len(right)
	}
	return float64(matches) / float64(maxLen)
}

func hasMeaningfulOrganizationTokenOverlap(fx *features.ExtractedFeatures) bool {
	if fx == nil {
		return false
	}

	stopwords := map[string]struct{}{
		"INC": {}, "LLC": {}, "LTD": {}, "LIMITED": {}, "CO": {}, "COMPANY": {},
		"CORP": {}, "CORPORATION": {}, "GROUP": {}, "HOLDINGS": {}, "HOLDING": {},
		"SERVICES": {}, "SERVICE": {}, "TRADING": {}, "INDUSTRIES": {}, "INDUSTRY": {},
	}

	left := map[string]struct{}{}
	for _, token := range fx.ScreenedName.Tokens {
		token = strings.TrimSpace(strings.ToUpper(token))
		if token == "" {
			continue
		}
		if _, skip := stopwords[token]; skip {
			continue
		}
		left[token] = struct{}{}
	}

	for _, token := range fx.MatchedName.Tokens {
		token = strings.TrimSpace(strings.ToUpper(token))
		if token == "" {
			continue
		}
		if _, skip := stopwords[token]; skip {
			continue
		}
		if _, ok := left[token]; ok {
			return true
		}
	}

	return false
}