package notes

import (
	"regexp"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

var (
	reDOBClaim        = regexp.MustCompile(`(?i)\b(date of birth|dob|born)\b`)
	reCountryClaim    = regexp.MustCompile(`(?i)\b(country|jurisdiction|venezuela|united states|syria|cuba|iran|us|ve|sy|cu|ir)\b`)
	reAddressClaim    = regexp.MustCompile(`(?i)\b(address|street|st\b|avenue|ave\b|broadway|main st|jersey city|new york)\b`)
	reIdentifierClaim = regexp.MustCompile(`(?i)\b(passport|identifier|tax id|tax_id|registration|imo|document number)\b`)
)

// Backward-compatible entrypoint for existing tests/callers.
func NormalizeAndValidate(note *AnalystNote, score *scoring.Result) *AnalystNote {
	return NormalizeAndValidateWithContext(note, nil, nil, score)
}

// New context-aware entrypoint for runtime use.
func NormalizeAndValidateWithContext(
	note *AnalystNote,
	alert *alerts.CanonicalAlert,
	fx *features.ExtractedFeatures,
	score *scoring.Result,
) *AnalystNote {
	if note == nil {
		return nil
	}

	note.Note = normalizeFreeText(note.Note)
	note.NextStepRationale = normalizeNextStepRationale(note.NextStepRationale, score)
	note.EvidenceSummary = normalizeSummaryItems(note.EvidenceSummary)
	note.MissingInformationSummary = normalizeMissingItems(note.MissingInformationSummary)

	sanitizeGroundedFacts(note, alert, fx, score)

	// Backfill evidence summary if normalization stripped everything.
	if len(note.EvidenceSummary) == 0 {
		note.EvidenceSummary = rebuildEvidenceSummaryFromScore(score)
		if len(note.EvidenceSummary) > 0 {
			appendWarning(note, "granite_analyst_note_evidence_summary_rebuilt_from_deterministic_score")
		}
	}

	if strings.TrimSpace(note.Note) == "" &&
		len(note.EvidenceSummary) == 0 &&
		strings.TrimSpace(note.NextStepRationale) == "" {
		return &AnalystNote{
			Status:   StatusFailed,
			Warnings: []string{"granite_analyst_note_empty_after_normalization"},
		}
	}

	if strings.TrimSpace(note.NextStepRationale) == "" && score != nil {
		note.NextStepRationale = deterministicNextStepRationale(score)
		if note.NextStepRationale != "" {
			appendWarning(note, "granite_analyst_note_next_step_rationale_rewritten")
		}
	}

	if len(note.MissingInformationSummary) == 0 {
		note.MissingInformationSummary = nil
	}

	appendPotentialInconsistencyWarning(note, score)

	if strings.TrimSpace(note.Note) == "" {
		note.Status = StatusFailed
		appendWarning(note, "granite_analyst_note_empty_after_normalization")
	} else if note.Status == "" {
		note.Status = StatusGenerated
	}

	return note
}
func sanitizeGroundedFacts(note *AnalystNote, alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures, score *scoring.Result) {
	if note == nil {
		return
	}

	// Preserve legacy NormalizeAndValidate(note, score) behavior used by old tests.
	// Runtime calls NormalizeAndValidateWithContext(...) and still gets grounding guards.
	if alert == nil && fx == nil {
		return
	}

	allowDOB := false
	allowCountry := false
	allowAddress := false
	allowIdentifier := false

	if fx != nil {
		allowDOB = fx.Date.ExactMatch || fx.Date.ScreenedExact != "" || fx.Date.MatchedExact != ""
		allowCountry = len(fx.Geography.ScreenedCountries) > 0 || len(fx.Geography.MatchedCountries) > 0 || fx.Geography.HasCountrySupport
		allowAddress = len(fx.Geography.AddressCountries) > 0
		allowIdentifier = len(fx.Identifiers.ExactMatches) > 0 || len(fx.Identifiers.ScreenedByType) > 0 || len(fx.Identifiers.MatchedByType) > 0
	} else if alert != nil {
		allowDOB = alert.ScreenedParty.DateOfBirth != "" || alert.MatchedParty.DateOfBirth != ""
		allowCountry = len(alert.ScreenedParty.Countries) > 0 || len(alert.MatchedParty.Countries) > 0
		allowAddress = len(alert.ScreenedParty.Addresses) > 0
		allowIdentifier = len(alert.ScreenedParty.Identifiers) > 0 || len(alert.MatchedParty.Identifiers) > 0
	}

	note.Note = stripUnsupportedClaims(note.Note, allowDOB, allowCountry, allowAddress, allowIdentifier)
	note.NextStepRationale = stripUnsupportedClaims(note.NextStepRationale, allowDOB, allowCountry, allowAddress, allowIdentifier)

	var filteredEvidence []string
	for _, item := range note.EvidenceSummary {
		item = stripUnsupportedClaims(item, allowDOB, allowCountry, allowAddress, allowIdentifier)
		item = normalizeSummaryItem(item)
		if item != "" {
			filteredEvidence = append(filteredEvidence, item)
		}
	}
	note.EvidenceSummary = dedupeNonEmpty(filteredEvidence)

	var filteredMissing []string
	for _, item := range note.MissingInformationSummary {
		item = stripInstructionLeak(item)
		item = normalizeSummaryItem(item)
		if item == "" || isPlaceholderMissingInfoLocal(item) {
			continue
		}
		filteredMissing = append(filteredMissing, item)
	}
	note.MissingInformationSummary = dedupeNonEmpty(filteredMissing)

	if strings.TrimSpace(note.Note) == "" && score != nil {
		note.Note = fallbackGroundedNote(alert, score)
		appendWarning(note, "granite_analyst_note_note_rewritten_for_grounding")
	}
}
func stripUnsupportedClaims(s string, allowDOB, allowCountry, allowAddress, allowIdentifier bool) string {
	s = normalizeFreeText(s)
	if s == "" {
		return ""
	}

	lower := strings.ToLower(strings.TrimSpace(s))

	if !allowDOB && reDOBClaim.MatchString(lower) {
		return ""
	}
	if !allowCountry && reCountryClaim.MatchString(lower) {
		return ""
	}
	if !allowAddress && reAddressClaim.MatchString(lower) {
		return ""
	}
	if !allowIdentifier && reIdentifierClaim.MatchString(lower) {
		return ""
	}

	return s
}

func normalizeSummaryItem(s string) string {
	s = stripInstructionLeak(s)
	s = strings.TrimSpace(s)

	// unwrap repeated markdown / quote wrappers
	for {
		before := s
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "\"'`")
		s = strings.TrimSpace(s)
		s = strings.TrimLeft(s, "-*• ")
		s = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(s, "")
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "*_`")
		s = strings.TrimSpace(s)
		if s == before {
			break
		}
	}

	return strings.TrimSpace(s)
}
func normalizeSummaryItems(items []string) []string {
	var out []string
	for _, item := range items {
		item = normalizeSummaryItem(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return dedupeNonEmpty(out)
}
func normalizeMissingItems(items []string) []string {
	var out []string
	for _, item := range items {
		item = normalizeSummaryItem(item)
		if item == "" || isPlaceholderMissingInfoLocal(item) {
			continue
		}
		out = append(out, item)
	}
	return dedupeNonEmpty(out)
}

func normalizeFreeText(s string) string {
	s = stripInstructionLeak(s)
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func stripInstructionLeak(s string) string {
	if s == "" {
		return ""
	}

	orig := strings.TrimSpace(s)

	normalizeForMatch := func(v string) string {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, "\"'`*_ ")
		v = strings.Join(strings.Fields(v), " ")
		return strings.ToLower(v)
	}

	badStandalone := []string{
		"3-5 grounded evidence items only",
		"0-4 grounded missing-information items only",
		"3-5 concise grounded bullets without markdown bullet prefixes",
		"0-4 concise grounded items; no placeholder text",
		"1 concise sentence aligned to deterministic next_step",
	}

	prefixes := []string{
		"1 concise sentence aligned to deterministic next_step:",
		"3-5 grounded evidence items only:",
		"0-4 grounded missing-information items only:",
		"3-5 concise grounded bullets without markdown bullet prefixes:",
		"0-4 concise grounded items; no placeholder text:",
	}

	norm := normalizeForMatch(orig)

	for _, bad := range badStandalone {
		if norm == normalizeForMatch(bad) {
			return ""
		}
	}

	// Remove leading quoted/scaffold prefixes repeatedly.
	current := orig
	for {
		trimmed := strings.TrimSpace(current)
		trimmed = strings.Trim(trimmed, "\"'`*_ ")
		matched := false

		for _, prefix := range prefixes {
			normTrimmed := normalizeForMatch(trimmed)
			normPrefix := normalizeForMatch(prefix)
			if strings.HasPrefix(normTrimmed, normPrefix) {
				// slice from the original trimmed string using prefix length in plain form
				candidate := strings.TrimSpace(trimmed[len(prefix):])
				candidate = strings.Trim(candidate, "\"'`*_ ")
				current = candidate
				matched = true
				break
			}
		}

		if !matched {
			break
		}
	}

	current = strings.TrimSpace(current)
	current = strings.Trim(current, "\"'`")
	return strings.TrimSpace(current)
}

func isPlaceholderMissingInfoLocal(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	bad := []string{
		"no explicit missing information",
		"no missing information",
		"none identified",
		"all essential data elements provided",
		"all necessary details for decision-making are present",
		"all required evidence for decision is present",
		"no missing information found",
		"no missing information is noted",
	}
	for _, phrase := range bad {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func fallbackGroundedNote(alert *alerts.CanonicalAlert, score *scoring.Result) string {
	if score == nil {
		return ""
	}

	name := ""
	if alert != nil {
		if strings.TrimSpace(alert.ScreenedParty.Name.FullName) != "" {
			name = strings.TrimSpace(alert.ScreenedParty.Name.FullName)
		} else {
			name = strings.TrimSpace(alert.MatchedParty.Name.FullName)
		}
	}

	switch {
	case name != "" && score.DecisionReason != "":
		return ensureSentence(name + ": " + score.DecisionReason)
	case name != "":
		return ensureSentence(name + " requires review based on deterministic scoring")
	case score.DecisionReason != "":
		return ensureSentence(score.DecisionReason)
	default:
		return ""
	}
}

func ensureSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	last := s[len(s)-1]
	if last == '.' || last == '!' || last == '?' {
		return s
	}
	return s + "."
}

func dedupeNonEmpty(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
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
	return out
}
func appendPotentialInconsistencyWarning(note *AnalystNote, score *scoring.Result) {
	if note == nil || score == nil {
		return
	}

	text := strings.ToLower(strings.Join([]string{
		note.Note,
		note.NextStepRationale,
		strings.Join(note.EvidenceSummary, " "),
		strings.Join(note.MissingInformationSummary, " "),
	}, " "))

	switch strings.TrimSpace(score.DecisionLabel) {
	case "escalate":
		if strings.Contains(text, "close") || strings.Contains(text, "false positive") || strings.Contains(text, "clear") {
			appendWarning(note, "granite_analyst_note_potentially_inconsistent_with_decision")
		}
	case "close":
		if strings.Contains(text, "escalate") || strings.Contains(text, "investigate") {
			appendWarning(note, "granite_analyst_note_potentially_inconsistent_with_decision")
		}
	case "investigate_next_step":
		if strings.Contains(text, "close") || strings.Contains(text, "escalate") {
			appendWarning(note, "granite_analyst_note_potentially_inconsistent_with_decision")
		}
	}
}
func rebuildEvidenceSummaryFromScore(score *scoring.Result) []string {
	if score == nil {
		return nil
	}

	var out []string
	for _, item := range score.EvidenceFor {
		item = normalizeSummaryItem(item)
		if item == "" {
			continue
		}
		out = append(out, humanizeEvidenceLabel(item))
	}

	return dedupeNonEmpty(out)
}

func humanizeEvidenceLabel(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "address support":
		return "Address support"
	case "country support":
		return "Country support"
	case "entity type support":
		return "Entity type support"
	case "exact date match":
		return "Exact date match"
	case "exact normalized name match":
		return "Exact normalized name match"
	default:
		return ensureSentence(strings.TrimSpace(s))
	}
}
func normalizeNextStepRationale(raw string, score *scoring.Result) string {
	s := normalizeSummaryItem(raw)
	s = strings.TrimSpace(s)

	if s == "" {
		return deterministicNextStepRationale(score)
	}

	// remove any leftover scaffolding text if it survived partial normalization
	lower := strings.ToLower(s)
	prefix := "1 concise sentence aligned to deterministic next_step:"
	if strings.HasPrefix(lower, prefix) {
		s = strings.TrimSpace(s[len(prefix):])
	}

	s = strings.Trim(s, "\"'`")
	s = strings.TrimSpace(s)
	if s == "" {
		return deterministicNextStepRationale(score)
	}

	return ensureSentence(s)
}
