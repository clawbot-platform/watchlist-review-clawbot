# Fix application notes

## 1. Wire the contradiction policy into the live scoring path

Add this right before the deterministic result is returned from the real scoring function:

```go
result := scoreDeterministically(...)
scoring.ApplySparseContradictionPolicy(&result)
return result
```

If the call site lives inside the `scoring` package already, drop the package prefix:

```go
result := scoreDeterministically(...)
ApplySparseContradictionPolicy(&result)
return result
```

## 2. Restart the live worker

After wiring the scoring path:

```bash
pkill -f 'cmd/review-worker-api' || true
go run ./cmd/review-worker-api
```

## 3. Re-run the calibrated suite

```bash
export REVIEW_BATCH_SPEC_PATH='eval/goldset/manifest.calibrated.v1.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
export REVIEW_BATCH_EVAL_TIMEOUT='90s'
go run ./cmd/review-batch-eval
```

## 4. Promotion rerun

```bash
export REVIEW_BATCH_SPEC_PATH='eval/goldset/manifest.calibrated.v1.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
export REVIEW_BATCH_EVAL_TIMEOUT='90s'
export REVIEW_REPORTS_DIR='eval/reports'
go run ./cmd/review-regression-runner
```
