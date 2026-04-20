# Review Batch Evaluation

This update tightens note-quality checks. Batch evaluation now fails on:
- bullet-prefixed `evidence_summary` items
- truncated `next_step_rationale`
- placeholder-style `missing_information_summary`
- retrieval failures when `require_retrieval = true`

## Run

```bash
export REVIEW_BATCH_SPEC_PATH='eval/specs/review-batch-example.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
go run ./cmd/review-batch-eval
```

## Recommended local setup for retrieval-required cases

Terminal 1:
```bash
export RETRIEVAL_GATEWAY_BACKEND='local_json'
export RETRIEVAL_LOCAL_JSON_PATH='eval/retrieval/snippets.json'
go run ./cmd/retrieval-gateway
```

Terminal 2:
```bash
export ENABLE_REVIEW_RETRIEVAL='true'
export REVIEW_RETRIEVAL_BASE_URL='http://127.0.0.1:8088'
go run ./cmd/review-worker-api
```

Terminal 3:
```bash
go run ./cmd/review-batch-eval
```
