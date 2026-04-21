# Disposition Memory Scoring v1

## Scope

This phase adds deterministic history-aware scoring for reviewed cases.

The first implementation should only use prior reviewed history when:
- the screened entity fingerprint matches
- the matched list UID matches
- the prior case is for the same tenant
- the linked pair meets a minimum confidence threshold

## New scoring fields

- `previous_disposition_score`
- `previous_disposition_reasons[]`

## Inputs

Normalized history lookup should return:
- same-pair prior cases
- same matched profile prior cases
- most recent reviewed disposition
- counts by disposition label
- contradiction-pattern summaries
- review timestamps for recency decay

## First-pass weighting proposal

- `+12` same pair recently escalated/confirmed
- `-12` same pair recently closed as false positive with same contradiction pattern
- `+4` same matched profile with repeated positive history
- `+2` unresolved same-pair investigations
- decay weights using recency bands:
  - 0–30 days: 1.0
  - 31–90 days: 0.7
  - 91–180 days: 0.4
  - 181+ days: 0.2

## Guardrails

Do not apply history scoring when:
- the matched list UID differs
- the screened entity fingerprint differs materially
- the prior history is unresolved and identity linkage is weak
- the tenant differs

## Runtime integration point

1. runtime builds current alert features
2. runtime derives screened entity fingerprint
3. runtime loads same-pair history
4. scoring computes disposition-memory contribution
5. runtime attaches history evidence into `review_context`

## API / review context additions

Add:
- `previous_disposition_summary`
- `history_evidence`
- `deterministic_score.previous_disposition_score`
- `deterministic_score.previous_disposition_reasons`

## Eval additions

Add cases for:
- repeated confirmed true match
- repeated false positive same contradiction pattern
- repeated unresolved same pair
- history should not apply because pair is different
