# Review Batch Evaluation

This CLI runs review requests from a JSON spec and checks:
- expected deterministic decision label
- analyst note generated/non-empty
- evidence summary presence
- next-step rationale presence
- inconsistency warning absence
- retrieval context presence when required

## Run

```bash
export REVIEW_BATCH_SPEC_PATH='eval/specs/review-batch-example.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
go run ./cmd/review-batch-eval
```
