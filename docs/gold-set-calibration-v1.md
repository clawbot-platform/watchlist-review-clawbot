# Gold Set Calibration v1

This pack calibrates the expanded gold set after the first 12-case run showed three main issue classes:

1. deterministic policy closed sparse / contradictory cases that should now investigate
2. ACH fixtures were invalid because `transaction` payloads were missing
3. vessel-style case used an unsupported alert kind
4. non-individual retrieval relevance was weak and caused poor note grounding

## Included changes

- `internal/scoring/policy.go`
- `internal/scoring/policy_test.go`
- `eval/goldset/manifest.calibrated.v1.json`
- repaired request payloads under `eval/requests/goldset-calibrated/`
- `eval/retrieval/snippets.calibrated.v1.json`

## Deterministic policy wiring

Call the helper after the base deterministic score is computed:

```go
result := scoreDeterministically(...)
scoring.ApplySparseContradictionPolicy(&result)
return result
```

## Local retrieval run

```bash
export RETRIEVAL_GATEWAY_BACKEND='local_json'
export RETRIEVAL_LOCAL_JSON_PATH='eval/retrieval/snippets.calibrated.v1.json'
go run ./cmd/retrieval-gateway
```

## Batch-eval run

```bash
export REVIEW_BATCH_SPEC_PATH='eval/goldset/manifest.calibrated.v1.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
export REVIEW_BATCH_EVAL_TIMEOUT='90s'
go run ./cmd/review-batch-eval
```

## Promotion rerun

```bash
export REVIEW_BATCH_SPEC_PATH='eval/goldset/manifest.calibrated.v1.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
export REVIEW_BATCH_EVAL_TIMEOUT='90s'
export REVIEW_REPORTS_DIR='eval/reports'
go run ./cmd/review-regression-runner
```
