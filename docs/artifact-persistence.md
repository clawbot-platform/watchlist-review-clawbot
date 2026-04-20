# Review Artifact Persistence

This layer persists review outputs and analyst notes as JSON artifacts.

## What is persisted

- `review_output`
  - decision label
  - review context
- `analyst_note`
  - Granite-generated, skipped, or failed note payload

## Suggested env

```bash
export ENABLE_REVIEW_ARTIFACTS='true'
export REVIEW_ARTIFACTS_DIR='./var/artifacts'
```

Artifacts are written under a date/case/alert path and returned as `artifact_refs` in the API response.

This persistence layer is advisory and non-blocking:
- review generation still succeeds if artifact writing fails
- failures appear in `artifact_warnings`
