package notes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

const promptVersion = "granite-analyst-note-v5-local-json"

type PromptInput struct {
	Alert            *alerts.CanonicalAlert
	Features         *features.ExtractedFeatures
	Score            *scoring.Result
	Compare          *identity.CompareResponse
	Screening        *identity.ScreenOFACResponse
	RetrievalContext *retrieval.PromptContext
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ModelPrompt struct {
	Messages         []ChatMessage `json:"messages"`
	ResponseJSONHint string        `json:"response_json_hint,omitempty"`
}

func BuildPrompt(input PromptInput) (string, error) {
	mp, err := BuildModelPrompt(input)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(mp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal model prompt: %w", err)
	}
	return string(b), nil
}

func BuildModelPrompt(input PromptInput) (ModelPrompt, error) {
	if input.Alert == nil {
		return ModelPrompt{}, fmt.Errorf("alert is required")
	}
	if input.Score == nil {
		return ModelPrompt{}, fmt.Errorf("score is required")
	}

	system := buildSystemInstructions(input)
	userPayload, err := buildUserPayload(input)
	if err != nil {
		return ModelPrompt{}, err
	}

	return ModelPrompt{
		Messages: []ChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: userPayload},
		},
		ResponseJSONHint: analystNoteJSONSchemaHint(),
	}, nil
}

func buildSystemInstructions(input PromptInput) string {
	allowedNames := allowedEntityNames(input.Alert)
	retrievalRules := "Use retrieval snippets only as supporting context. Do not copy instructional text, template text, or formatting rules into the output."
	nameRules := "Only mention names that appear in the screened party or matched party. Never introduce names found only in screening candidates or retrieved examples."
	missingInfoRules := "If there is no genuine missing information, return an empty array for missing_information_summary. Do not write placeholder phrases such as 'no missing information found' or 'all necessary details are present'."
	formatRules := "Return strict JSON only. Do not include markdown. Do not include bullet prefixes. Do not echo instructions."

	var b strings.Builder
	b.WriteString("You are an AML/watchlist analyst-note generator.\n")
	b.WriteString("Your job is to explain the deterministic decision without changing it.\n")
	b.WriteString("The deterministic scorer is authoritative for decision_label and next_step.\n")
	b.WriteString("Write a grounded analyst note using only the supplied alert, score, identity evidence, and retrieval context.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("1. Do not change the deterministic disposition.\n")
	b.WriteString("2. ")
	b.WriteString(nameRules)
	b.WriteString("\n")
	b.WriteString("3. ")
	b.WriteString(retrievalRules)
	b.WriteString("\n")
	b.WriteString("4. ")
	b.WriteString(missingInfoRules)
	b.WriteString("\n")
	b.WriteString("5. ")
	b.WriteString(formatRules)
	b.WriteString("\n")
	if len(allowedNames) > 0 {
		b.WriteString("6. Allowed names: ")
		b.WriteString(strings.Join(allowedNames, "; "))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString("Output keys:\n")
	b.WriteString("- status\n")
	b.WriteString("- note\n")
	b.WriteString("- evidence_summary\n")
	b.WriteString("- missing_information_summary\n")
	b.WriteString("- next_step_rationale\n")
	return b.String()
}

func buildUserPayload(input PromptInput) (string, error) {
	payload := map[string]any{
		"alert": map[string]any{
			"kind":             safeAlertKind(input.Alert),
			"metadata":         input.Alert.Metadata,
			"screened_party":   input.Alert.ScreenedParty,
			"matched_party":    input.Alert.MatchedParty,
			"screening_feats":  input.Alert.ScreeningFeatures,
			"transaction":      alertTransactionMap(input.Alert),
			"allowed_names":    allowedEntityNames(input.Alert),
			"screened_name":    screenedName(input.Alert),
			"matched_name":     matchedName(input.Alert),
			"screened_type":    safeScreenedType(input.Alert),
			"matched_type":     safeMatchedType(input.Alert),
			"source_program":   safeProgram(input.Alert),
			"source_jurisdion": input.Alert.Metadata.Jurisdiction,
		},
		"deterministic_score": map[string]any{
			"decision_label":         input.Score.DecisionLabel,
			"decision_reason":        input.Score.DecisionReason,
			"next_step":              input.Score.NextStep,
			"match_strength_score":   input.Score.MatchStrengthScore,
			"data_sufficiency_score": input.Score.DataSufficiencyScore,
			"evidence_for":           input.Score.EvidenceFor,
			"evidence_against":       input.Score.EvidenceAgainst,
			"missing_information":    input.Score.MissingInformation,
			"contradictions":         input.Score.Contradictions,
		},
		"features":          input.Features,
		"identity_evidence": buildIdentityEvidence(input),
		"retrieval_context": buildRetrievalContext(input.RetrievalContext),
		"output_contract": map[string]any{
			"status": "generated",
			"note":   "1 concise grounded paragraph",
			"evidence_summary": []string{
				"3-5 grounded evidence items only",
			},
			"missing_information_summary": []string{
				"0-4 grounded missing-information items only",
			},
			"next_step_rationale": "1 concise sentence aligned to deterministic next_step",
		},
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal prompt payload: %w", err)
	}
	return string(b), nil
}

func buildIdentityEvidence(input PromptInput) map[string]any {
	out := map[string]any{}
	if input.Compare != nil {
		out["compare"] = input.Compare
	}
	if input.Screening != nil {
		out["screening"] = input.Screening
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildRetrievalContext(ctx *retrieval.PromptContext) map[string]any {
	if ctx == nil {
		return nil
	}
	return map[string]any{
		"query_text": ctx.QueryText,
		"snippets":   ctx.Snippets,
	}
}

func analystNoteJSONSchemaHint() string {
	return `{
  "status": "generated",
  "note": "string",
  "evidence_summary": ["string"],
  "missing_information_summary": ["string"],
  "next_step_rationale": "string"
}`
}

func allowedEntityNames(alert *alerts.CanonicalAlert) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}

	if alert == nil {
		return nil
	}
	add(alert.ScreenedParty.Name.FullName)
	add(alert.MatchedParty.Name.FullName)
	return out
}

func safeAlertKind(alert *alerts.CanonicalAlert) string {
	if alert == nil {
		return ""
	}
	return string(alert.Kind)
}

func safeScreenedType(alert *alerts.CanonicalAlert) string {
	if alert == nil {
		return ""
	}
	return string(alert.ScreenedParty.EntityType)
}

func safeMatchedType(alert *alerts.CanonicalAlert) string {
	if alert == nil {
		return ""
	}
	return string(alert.MatchedParty.EntityType)
}

func safeProgram(alert *alerts.CanonicalAlert) string {
	if alert == nil {
		return ""
	}
	return alert.MatchedParty.Program
}

func alertTransactionMap(alert *alerts.CanonicalAlert) map[string]any {
	if alert == nil || alert.Transaction == nil {
		return nil
	}

	return map[string]any{
		"transaction_id": alert.Transaction.TransactionID,
		"rail_type":      alert.Transaction.RailType,
		"amount":         alert.Transaction.Amount,
		"currency":       alert.Transaction.Currency,
	}
}

func screenedName(alert *alerts.CanonicalAlert) string {
	if alert == nil {
		return ""
	}
	return strings.TrimSpace(alert.ScreenedParty.Name.FullName)
}

func matchedName(alert *alerts.CanonicalAlert) string {
	if alert == nil {
		return ""
	}
	return strings.TrimSpace(alert.MatchedParty.Name.FullName)
}
