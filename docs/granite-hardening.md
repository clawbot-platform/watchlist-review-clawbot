# Granite Hardening

This pack hardens the analyst-note layer without changing deterministic decisions.

## What is added

- note normalization and validation
- bullet stripping and deduplication for summaries
- consistency warning when note text appears to conflict with deterministic label
- API tests for:
  - `analyst_note.status = generated`
  - `analyst_note.status = skipped`
  - `analyst_note.status = failed`
- confirmation that deterministic labels remain present even when Granite fails

## Intent

Granite remains advisory.
The worker's governing outputs continue to be:
- deterministic scores
- contradictions
- decision label
- next step

Granite should never be able to silently override deterministic outcomes.
