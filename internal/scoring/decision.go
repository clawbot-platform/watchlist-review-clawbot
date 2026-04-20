package scoring

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
)

const (
	DecisionClose               = "close"
	DecisionInvestigateNextStep = "investigate_next_step"
	DecisionEscalate            = "escalate"
	DecisionReviewPending       = "review_pending"
)

func Evaluate(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures) *Result {
	result := &Result{}
	if alert == nil || fx == nil {
		result.DecisionLabel = DecisionReviewPending
		result.DecisionReason = "alert or extracted features unavailable"
		result.NextStep = "Review worker inputs and retry"
		return result
	}

	result.Contradictions = uniqueStrings(append([]string(nil), fx.Contradictions...))
	computeMatchStrength(fx, result)
	computeDataSufficiency(fx, result)
	result.MissingInformation = deriveMissingInformation(fx)
	decide(alert, fx, result)

	return result
}

func decide(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures, result *Result) {
	hasHardConflict := hasAny(result.Contradictions, "entity_type_conflict", "identifier_conflict")
	hasAnyConflict := len(result.Contradictions) > 0

	switch {
	case result.MatchStrengthScore >= 70 &&
		result.DataSufficiencyScore >= 50 &&
		!hasAnyConflict &&
		hasNonNameCorroboration(result):
		result.DecisionLabel = DecisionEscalate
		result.DecisionReason = "Strong corroborated match with sufficient data and no contradictions."
		result.NextStep = "Escalate for analyst review."

	case hasHardConflict:
		if result.MatchStrengthScore <= 30 {
			result.DecisionLabel = DecisionClose
			result.DecisionReason = "Hard contradiction present and overall match strength is weak."
			result.NextStep = "Close as likely false positive and document contradiction."
		} else {
			result.DecisionLabel = DecisionInvestigateNextStep
			result.DecisionReason = "Hard contradiction present; do not escalate without further verification."
			result.NextStep = "Investigate contradiction and gather additional corroboration."
		}

	case organizationGrayZone(alert, fx, result):
		result.DecisionLabel = DecisionInvestigateNextStep
		result.DecisionReason = "Organization name overlap exists, but corroboration is too weak for escalation and too meaningful for immediate closure."
		result.NextStep = "Investigate organization identifiers, registration details, and geography before closing."

	case result.MatchStrengthScore <= 30:
		result.DecisionLabel = DecisionClose
		result.DecisionReason = "Low match strength with insufficient evidence for escalation."
		result.NextStep = "Close as likely false positive."

	case result.DataSufficiencyScore < 50 || hasAnyConflict:
		result.DecisionLabel = DecisionInvestigateNextStep
		result.DecisionReason = "Case is plausible but under-evidenced or contradictory."
		result.NextStep = "Investigate next step and gather missing data."

	default:
		result.DecisionLabel = DecisionInvestigateNextStep
		result.DecisionReason = "Case falls in the gray zone and requires conservative follow-up."
		result.NextStep = "Investigate next step and gather corroborating evidence."
	}

	result.EvidenceFor = uniqueStrings(result.EvidenceFor)
	result.EvidenceAgainst = uniqueStrings(result.EvidenceAgainst)
	result.MissingInformation = uniqueStrings(result.MissingInformation)

	sort.Strings(result.Contradictions)
	sort.Strings(result.EvidenceFor)
	sort.Strings(result.EvidenceAgainst)
	sort.Strings(result.MissingInformation)
}

func organizationGrayZone(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures, result *Result) bool {
	if alert == nil || fx == nil || result == nil {
		return false
	}
	if alert.Kind != alerts.AlertKindOrganizationOnboarding {
		return false
	}
	if hasAny(result.Contradictions, "entity_type_conflict", "identifier_conflict") {
		return false
	}
	if hasNonNameCorroboration(result) {
		return false
	}

	// Organization overlap cases should stay conservative.
	// If there is any meaningful business-name overlap and the case has
	// enough structure to review, prefer investigate_next_step over close.
	return result.NameMatchScore >= 4 &&
		result.MatchStrengthScore <= 30 &&
		result.DataSufficiencyScore >= 30
}
func deriveMissingInformation(fx *features.ExtractedFeatures) []string {
	if fx == nil {
		return nil
	}
	var missing []string
	seen := map[string]struct{}{}
	for _, field := range fx.MissingFields {
		addMissing(seen, &missing, field)
	}
	if !fx.Date.ExactMatch && !fx.Date.YearMatch {
		addMissing(seen, &missing, "date corroboration")
	}
	if len(fx.Identifiers.ExactMatches) == 0 {
		addMissing(seen, &missing, "strong identifier corroboration")
	}
	if !fx.Geography.HasCountrySupport {
		addMissing(seen, &missing, "geography corroboration")
	}
	if fx.ScreenedCompleteness.AddressCount == 0 && fx.MatchedCompleteness.AddressCount == 0 {
		addMissing(seen, &missing, "address context")
	}
	return missing
}

func addMissing(seen map[string]struct{}, out *[]string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if _, ok := seen[value]; ok {
		return
	}
	seen[value] = struct{}{}
	*out = append(*out, value)
}

func hasAny(values []string, wanted ...string) bool {
	set := map[string]struct{}{}
	for _, v := range values {
		set[strings.TrimSpace(v)] = struct{}{}
	}
	for _, want := range wanted {
		if _, ok := set[want]; ok {
			return true
		}
	}
	return false
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (r *Result) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("MSS=%d DSS=%d label=%s", r.MatchStrengthScore, r.DataSufficiencyScore, r.DecisionLabel)
}
