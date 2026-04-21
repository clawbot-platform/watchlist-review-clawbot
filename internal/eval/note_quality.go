package eval

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/api"
)

var likelyEntityNamePattern = regexp.MustCompile(`\b(?:[A-Z][a-z]+|[A-Z]{2,}|[A-Z][a-z]+[.-][A-Z][a-z]+)(?:\s+(?:[A-Z][a-z]+|[A-Z]{2,}|[A-Z][a-z]+[.-][A-Z][a-z]+|LLC|LTD|INC|CO|CORP|SA|PLC|BV|NV)){1,5}\b`)

var ignoredPseudoNames = map[string]struct{}{
	"main st":                     {},
	"the dob":                     {},
	"date of birth":               {},
	"united states":               {},
	"jersey city":                 {},
	"new york":                    {},
	"venezuela":                   {},
	"syria":                       {},
	"cuba":                        {},
	"iran":                        {},
	"united states of america":    {},
	"passport number":             {},
	"passport id":                 {},
	"passport identifier":         {},
	"identifier match":            {},
	"exact identifier match":      {},
	"normalized name":             {},
	"exact normalized name":       {},
	"screened party":              {},
	"matched party":               {},
	"screened name":               {},
	"matched name":                {},
	"entity type":                 {},
	"hard contradiction":          {},
	"country mismatch":            {},
	"country match":               {},
	"address support":             {},
	"country support":             {},
	"entity type support":         {},
	"exact date match":            {},
	"exact normalized name match": {},
}

var ignoredNameTokens = map[string]struct{}{
	"st":         {},
	"street":     {},
	"road":       {},
	"rd":         {},
	"ave":        {},
	"avenue":     {},
	"blvd":       {},
	"city":       {},
	"country":    {},
	"state":      {},
	"dob":        {},
	"passport":   {},
	"id":         {},
	"identifier": {},
	"match":      {},
	"exact":      {},
	"normalized": {},
	"name":       {},
	"screened":   {},
	"matched":    {},
	"party":      {},
	"program":    {},
	"address":    {},
	"support":    {},
}

func EvaluateCase(spec CaseSpec, resp *api.ReviewResponse) ([]string, []string) {
	var errors []string
	var warnings []string

	if resp == nil {
		return []string{"nil response"}, nil
	}

	if spec.ExpectDecisionLabel != "" && resp.DecisionLabel != spec.ExpectDecisionLabel {
		errors = append(errors, "unexpected decision_label: got="+resp.DecisionLabel+" want="+spec.ExpectDecisionLabel)
	}

	if spec.RequireGeneratedNote {
		if resp.AnalystNote == nil {
			errors = append(errors, "missing analyst_note")
		} else {
			n := resp.AnalystNote

			if n.Status != "generated" {
				errors = append(errors, "analyst_note.status="+n.Status+" want=generated")
			}
			if strings.TrimSpace(n.Note) == "" {
				errors = append(errors, "analyst_note.note is empty")
			}
			if len(n.EvidenceSummary) == 0 {
				errors = append(errors, "analyst_note.evidence_summary is empty")
			}
			if strings.TrimSpace(n.NextStepRationale) == "" {
				errors = append(errors, "analyst_note.next_step_rationale is empty")
			}

			for _, item := range n.EvidenceSummary {
				trimmed := strings.TrimSpace(item)
				if strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "•") {
					errors = append(errors, "analyst_note.evidence_summary contains bullet-prefixed item: "+trimmed)
					break
				}
				if isInstructionLeak(trimmed) {
					errors = append(errors, "analyst_note.evidence_summary contains instruction leakage")
					break
				}
			}

			for _, item := range n.MissingInformationSummary {
				trimmed := strings.TrimSpace(item)
				if isInstructionLeak(trimmed) {
					errors = append(errors, "analyst_note.missing_information_summary contains instruction leakage")
					break
				}
				if isPlaceholderMissingInfo(trimmed) {
					errors = append(errors, "analyst_note.missing_information_summary contains placeholder text")
					break
				}
			}

			if noteContainsPlaceholderBody(n.Note) {
				errors = append(errors, "analyst_note.note contains placeholder no-missing-information text")
			}

			if rationaleLooksTruncated(n.NextStepRationale) {
				errors = append(errors, "analyst_note.next_step_rationale appears truncated")
			}

			if contaminationErr := detectCrossEntityContamination(resp); contaminationErr != "" {
				errors = append(errors, contaminationErr)
			}

			for _, warning := range n.Warnings {
				if warning == "granite_analyst_note_potentially_inconsistent_with_decision" {
					errors = append(errors, "analyst note inconsistent with deterministic decision")
				}
			}
		}
	}

	if spec.RequireRetrieval {
		raw, _ := json.Marshal(resp.ReviewContext)
		text := string(raw)
		if !strings.Contains(text, "retrieval_context") {
			errors = append(errors, "retrieval_context missing from review_context")
		} else {
			if strings.Contains(text, "retrieval_failed") || strings.Contains(text, "retrieval_not_configured") {
				errors = append(errors, "retrieval_context contains retrieval failure or not-configured warning")
			}
			if !strings.Contains(text, "\"snippets\"") {
				errors = append(errors, "retrieval_context does not contain snippets")
			}
		}
	}

	return dedupe(errors), dedupe(warnings)
}

func detectCrossEntityContamination(resp *api.ReviewResponse) string {
	if resp == nil || resp.AnalystNote == nil {
		return ""
	}

	allowed := allowedNamesFromReviewContext(resp.ReviewContext)
	if len(allowed) == 0 {
		return ""
	}

	// 1) Explicit candidate names from review context.
	for _, candidate := range extractCandidateNames(resp.ReviewContext) {
		norm := normalizeName(candidate)
		if norm == "" || isIgnoredPseudoName(norm) {
			continue
		}
		if _, ok := allowed[norm]; ok {
			continue
		}
		if noteContainsLiteral(resp, candidate) {
			return fmt.Sprintf("analyst_note appears contaminated by unrelated entity reference: %s", norm)
		}
	}

	// 2) Fallback lexical scan, but only for likely entity names.
	for _, mention := range extractLikelyEntityMentions(resp) {
		norm := normalizeName(mention)
		if norm == "" || isIgnoredPseudoName(norm) {
			continue
		}
		if _, ok := allowed[norm]; ok {
			continue
		}
		return fmt.Sprintf("analyst_note appears contaminated by unrelated entity reference: %s", norm)
	}

	return ""
}

func allowedNamesFromReviewContext(reviewContext any) map[string]struct{} {
	out := map[string]struct{}{}

	add := func(name string) {
		name = normalizeName(name)
		if name == "" {
			return
		}
		out[name] = struct{}{}
	}

	add(extractNestedString(reviewContext, "alert", "screened_party", "name", "full_name"))
	add(extractNestedString(reviewContext, "alert", "matched_party", "name", "full_name"))

	for _, alias := range extractNestedStringSlice(reviewContext, "alert", "screened_party", "name", "aliases") {
		add(alias)
	}
	for _, alias := range extractNestedStringSlice(reviewContext, "alert", "matched_party", "name", "aliases") {
		add(alias)
	}

	return out
}

func noteContainsLiteral(resp *api.ReviewResponse, name string) bool {
	if resp == nil || resp.AnalystNote == nil {
		return false
	}
	target := strings.ToLower(strings.TrimSpace(name))
	if target == "" {
		return false
	}
	for _, text := range noteTexts(resp) {
		if strings.Contains(strings.ToLower(text), target) {
			return true
		}
	}
	return false
}

func extractLikelyEntityMentions(resp *api.ReviewResponse) []string {
	if resp == nil || resp.AnalystNote == nil {
		return nil
	}

	var out []string
	seen := map[string]struct{}{}

	for _, text := range noteTexts(resp) {
		for _, match := range likelyEntityNamePattern.FindAllString(text, -1) {
			norm := normalizeName(match)
			if norm == "" || isIgnoredPseudoName(norm) || !looksLikeEntityName(match) {
				continue
			}
			if _, ok := seen[norm]; ok {
				continue
			}
			seen[norm] = struct{}{}
			out = append(out, strings.TrimSpace(match))
		}
	}

	return out
}

func looksLikeEntityName(s string) bool {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) < 2 {
		return false
	}
	badCount := 0
	for _, part := range parts {
		if _, ok := ignoredNameTokens[strings.ToLower(strings.Trim(part, ".,:;()[]{}'\""))]; ok {
			badCount++
		}
	}
	return badCount < len(parts)
}

func isIgnoredPseudoName(s string) bool {
	s = normalizeName(s)
	if s == "" {
		return true
	}
	if _, ok := ignoredPseudoNames[s]; ok {
		return true
	}
	return false
}

func noteTexts(resp *api.ReviewResponse) []string {
	if resp == nil || resp.AnalystNote == nil {
		return nil
	}
	var out []string
	out = append(out, resp.AnalystNote.Note)
	out = append(out, resp.AnalystNote.NextStepRationale)
	out = append(out, resp.AnalystNote.EvidenceSummary...)
	out = append(out, resp.AnalystNote.MissingInformationSummary...)
	return out
}

func extractCandidateNames(v any) []string {
	var out []string
	seen := map[string]struct{}{}

	var walk func(any)
	walk = func(node any) {
		switch t := node.(type) {
		case map[string]any:
			if candidates, ok := t["candidates"]; ok {
				if arr, ok := candidates.([]any); ok {
					for _, item := range arr {
						obj, ok := item.(map[string]any)
						if !ok {
							continue
						}
						name, ok := obj["name"].(string)
						if !ok {
							continue
						}
						norm := normalizeName(name)
						if norm == "" {
							continue
						}
						if _, exists := seen[norm]; exists {
							continue
						}
						seen[norm] = struct{}{}
						out = append(out, strings.TrimSpace(name))
					}
				}
			}
			for _, child := range t {
				walk(child)
			}
		case []any:
			for _, child := range t {
				walk(child)
			}
		}
	}

	walk(v)
	return out
}

func extractNestedString(root any, path ...string) string {
	current, ok := root.(map[string]any)
	if !ok {
		return ""
	}

	for i, part := range path {
		val, exists := current[part]
		if !exists {
			return ""
		}
		if i == len(path)-1 {
			s, _ := val.(string)
			return strings.TrimSpace(s)
		}
		next, ok := val.(map[string]any)
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

func extractNestedStringSlice(root any, path ...string) []string {
	current, ok := root.(map[string]any)
	if !ok {
		return nil
	}

	for i, part := range path {
		val, exists := current[part]
		if !exists {
			return nil
		}
		if i == len(path)-1 {
			raw, ok := val.([]any)
			if !ok {
				return nil
			}
			var out []string
			for _, item := range raw {
				s, ok := item.(string)
				if ok && strings.TrimSpace(s) != "" {
					out = append(out, strings.TrimSpace(s))
				}
			}
			return out
		}
		next, ok := val.(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}
	return nil
}

func normalizeName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}

func rationaleLooksTruncated(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	last := s[len(s)-1]
	return last != '.' && last != '!' && last != '?'
}

func isInstructionLeak(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	bad := []string{
		"3-5 grounded evidence items only",
		"0-4 grounded missing-information items only",
		"3-5 concise grounded bullets",
		"0-4 concise grounded items",
		"without markdown bullet prefixes",
		"no placeholder text",
		"return strict json",
		"1 concise sentence aligned to deterministic next_step",
	}
	for _, phrase := range bad {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func isPlaceholderMissingInfo(s string) bool {
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

func noteContainsPlaceholderBody(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	bad := []string{
		"no missing information found",
		"no missing information is noted",
		"all required evidence for decision is present",
		"all necessary details for decision-making are present",
	}
	for _, phrase := range bad {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, item := range in {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
