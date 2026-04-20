package notes

import (
	"encoding/json"
	"fmt"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/identity"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

const promptVersion = "granite-analyst-note-v2-rag"

type PromptInput struct {
	Alert            *alerts.CanonicalAlert
	Features         *features.ExtractedFeatures
	Score            *scoring.Result
	Compare          *identity.CompareResponse
	Screening        *identity.ScreenOFACResponse
	RetrievalContext *retrieval.PromptContext
}

func BuildPrompt(input PromptInput) (string, error) {
	if input.Alert == nil || input.Features == nil || input.Score == nil {
		return "", fmt.Errorf("alert, features, and score are required")
	}

	payload := map[string]any{
		"prompt_version": promptVersion,
		"alert_metadata": map[string]any{
			"kind":         input.Alert.Kind,
			"alert_id":     input.Alert.Metadata.AlertID,
			"case_id":      input.Alert.Metadata.CaseID,
			"alert_type":   input.Alert.Metadata.AlertType,
			"jurisdiction": input.Alert.Metadata.Jurisdiction,
		},
		"screened_party": map[string]any{
			"entity_type":        input.Alert.ScreenedParty.EntityType,
			"name":               input.Alert.ScreenedParty.Name.FullName,
			"aliases":            input.Alert.ScreenedParty.Name.Aliases,
			"date_of_birth":      input.Alert.ScreenedParty.DateOfBirth,
			"incorporation_date": input.Alert.ScreenedParty.IncorporationDate,
			"countries":          input.Alert.ScreenedParty.Countries,
		},
		"matched_party": map[string]any{
			"entity_type":        input.Alert.MatchedParty.EntityType,
			"name":               input.Alert.MatchedParty.Name.FullName,
			"aliases":            input.Alert.MatchedParty.Name.Aliases,
			"date_of_birth":      input.Alert.MatchedParty.DateOfBirth,
			"incorporation_date": input.Alert.MatchedParty.IncorporationDate,
			"countries":          input.Alert.MatchedParty.Countries,
			"program":            input.Alert.MatchedParty.Program,
		},
		"deterministic_result": map[string]any{
			"decision_label":         input.Score.DecisionLabel,
			"decision_reason":        input.Score.DecisionReason,
			"match_strength_score":   input.Score.MatchStrengthScore,
			"data_sufficiency_score": input.Score.DataSufficiencyScore,
			"contradictions":         input.Score.Contradictions,
			"evidence_for":           input.Score.EvidenceFor,
			"evidence_against":       input.Score.EvidenceAgainst,
			"missing_information":    input.Score.MissingInformation,
			"next_step":              input.Score.NextStep,
		},
		"identity_evidence": buildIdentityEvidenceSummary(input),
		"retrieval_context": buildRetrievalPayload(input.RetrievalContext),
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal prompt payload: %w", err)
	}

	prompt := fmt.Sprintf(`You are drafting an analyst note for a watchlist review case.

Rules:
- Do not change the deterministic decision label.
- Treat the deterministic decision as the governing outcome.
- Write for a compliance analyst.
- Be concise and factual.
- Do not invent facts not present in the payload.
- Use retrieval_context only as supporting background. Do not let it override deterministic evidence.
- If retrieval snippets conflict with deterministic evidence, prefer deterministic evidence and note the conflict cautiously.
- Return only JSON matching the requested schema.

Create:
1. "note": 3-5 sentences, concise analyst note.
2. "evidence_summary": 2-6 short bullets summarizing the strongest evidence.
3. "missing_information_summary": 0-5 short bullets for missing or weak information.
4. "next_step_rationale": one short sentence explaining the deterministic next step.

Payload:
%s`, string(encoded))

	return prompt, nil
}

func buildIdentityEvidenceSummary(input PromptInput) map[string]any {
	out := map[string]any{}
	if input.Compare != nil {
		out["compare"] = map[string]any{
			"disposition":       input.Compare.Disposition,
			"confidence_band":   input.Compare.ConfidenceBand,
			"decision_trace_id": input.Compare.DecisionTraceID,
		}
		if input.Compare.Explanation != nil {
			out["compare_explanation"] = map[string]any{
				"summary": input.Compare.Explanation.Summary,
				"why":     input.Compare.Explanation.Why,
				"why_not": input.Compare.Explanation.WhyNot,
			}
		}
	}
	if input.Screening != nil {
		out["screening"] = map[string]any{
			"decision":          input.Screening.Decision,
			"decision_trace_id": input.Screening.DecisionTraceID,
			"screening_id":      input.Screening.ScreeningID,
			"candidate_count":   len(input.Screening.Candidates),
		}
		if len(input.Screening.Candidates) > 0 {
			out["top_candidate"] = input.Screening.Candidates[0]
		}
	}
	return out
}

func buildRetrievalPayload(ctx *retrieval.PromptContext) map[string]any {
	if ctx == nil {
		return map[string]any{"warnings": []string{"retrieval_not_requested"}}
	}
	out := map[string]any{
		"query_text": ctx.QueryText,
		"warnings":   ctx.Warnings,
	}
	if len(ctx.Snippets) > 0 {
		var snippets []map[string]any
		for _, snip := range ctx.Snippets {
			snippets = append(snippets, map[string]any{
				"snippet_id": snip.SnippetID,
				"source":     snip.Source,
				"title":      snip.Title,
				"text":       snip.Text,
				"score":      snip.Score,
				"tags":       snip.Tags,
			})
		}
		out["snippets"] = snippets
	}
	return out
}
