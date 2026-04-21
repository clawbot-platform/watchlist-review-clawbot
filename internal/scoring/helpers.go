package scoring

import (
	"sort"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
)

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}

	if len(out) == 0 {
		return nil
	}

	sort.Strings(out)
	return out
}

func hasAny(haystack []string, needles ...string) bool {
	if len(haystack) == 0 || len(needles) == 0 {
		return false
	}

	lookup := make(map[string]struct{}, len(haystack))
	for _, item := range haystack {
		item = strings.ToLower(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		lookup[item] = struct{}{}
	}

	for _, needle := range needles {
		needle = strings.ToLower(strings.TrimSpace(needle))
		if needle == "" {
			continue
		}
		if _, ok := lookup[needle]; ok {
			return true
		}
	}
	return false
}

func deriveMissingInformation(fx *features.ExtractedFeatures) []string {
	if fx == nil {
		return nil
	}

	var out []string

	// Prefer explicit missing fields if the extractor already computed them.
	out = append(out, fx.MissingFields...)

	// Conservative additional hints from missing corroboration.
	if !fx.Date.ExactMatch && fx.Date.ScreenedExact == "" {
		out = append(out, "screened_party.date")
	}
	if !fx.Date.ExactMatch && fx.Date.MatchedExact == "" {
		out = append(out, "matched_party.date")
	}

	if len(fx.Identifiers.ExactMatches) == 0 {
		if len(fx.Identifiers.ScreenedByType) == 0 {
			out = append(out, "screened_party.identifiers")
		}
		if len(fx.Identifiers.MatchedByType) == 0 {
			out = append(out, "matched_party.identifiers")
		}
		out = append(out, "strong identifier corroboration")
	}

	if !fx.Geography.HasCountrySupport {
		out = append(out, "geography corroboration")
	}

	return uniqueStrings(out)
}
