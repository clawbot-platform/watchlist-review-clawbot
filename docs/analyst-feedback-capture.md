# Analyst Feedback Capture

This layer adds explicit analyst feedback capture so notes and outcomes can feed future tuning.

## What is captured

- analyst agreement/disagreement with the system decision
- corrected label when the analyst disagrees
- note rating (0-5)
- outcome rating (0-5)
- free-text comment
- tuning tags such as:
  - `prompt_issue`
  - `retrieval_gap`
  - `false_positive`
  - `false_negative`

## Derived tuning signals

The service derives signals such as:
- `decision_policy_tuning_candidate`
- `note_quality_tuning_candidate`
- `outcome_quality_tuning_candidate`
- `retrieval_tuning_candidate`
- `prompt_tuning_candidate`

These are persisted with the feedback record for downstream tuning workflows.

## Endpoint

`POST /v1/feedback`

## Suggested env

```bash
export ENABLE_REVIEW_FEEDBACK_CAPTURE='true'
export ENABLE_REVIEW_ARTIFACTS='true'
```

Optional feedback events:

```bash
export ENABLE_REVIEW_FEEDBACK_EVENTS='true'
export REVIEW_FEEDBACK_EVENTS_NATS_URL='nats://100.67.85.91:4222'
export REVIEW_FEEDBACK_EVENTS_SUBJECT='clawbot.watchlist.review.feedback.created.v1'
```

## Example request

```json
{
  "tenant_id": "test-tenant",
  "case_id": "case-screening-001",
  "alert_id": "screening-json-source-refs-001",
  "analyst_id": "analyst-1",
  "system_decision": "escalate",
  "decision_agreement": "disagree",
  "corrected_label": "investigate_next_step",
  "note_rating": 2,
  "outcome_rating": 3,
  "comment": "Needs better explanation of why the passport evidence matters.",
  "tags": ["prompt_issue", "retrieval_gap"]
}
```
