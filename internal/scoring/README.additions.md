# `internal/scoring` additions from Deliverables A–C

This merged package combines the scoring work introduced across the first three deliverables.

## Deliverable A — Previous disposition scoring

Add these fields to the existing `Result` struct:

```go
PreviousDispositionScore   int      `json:"previous_disposition_score"`
PreviousDispositionReasons []string `json:"previous_disposition_reasons,omitempty"`
```

Recommended integration point:
- compute base deterministic score
- derive contradiction pattern string
- load same-pair history
- call `ApplyPreviousDispositionScore(&result, input)`
- run final decision policy hooks

## Deliverable B — OFAC relationship scoring

Add the following fields to the real `internal/scoring.Result`:

- `RelationshipSupportScore int`
- `RelationshipConflictPenalty int`
- `OfficialDocLinkScore int`
- `ProgramContextScore int`
- `RelationshipReasons []string`

Add relationship-aware scoring after base feature scoring and before final policy thresholds.

## Deliverable C — promotion coverage

Use the added eval/specs and test cases in this merged package to validate:
- repeated confirmed true match
- repeated false positive with same contradiction pattern
- related-party reinforcement
- wrong-context relationship neighborhood

Acceptance:
- no regression on the current calibrated suite
- deterministic regressions remain zero on legacy cases
- new history/relationship cases have explicit coverage
