# Reviewer Learning Loop v1

This code drop adds:
- saved JSON and Markdown regression reports
- a promotion gate over batch-eval output
- a one-command regression runner
- a gold-set manifest path for reviewed cases

## Run

```bash
export REVIEW_BATCH_SPEC_PATH='eval/goldset/manifest.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
export REVIEW_BATCH_EVAL_TIMEOUT='60s'
export REVIEW_REPORTS_DIR='eval/reports'

go run ./cmd/review-regression-runner
```

## Output

The runner writes:
- `report.json`
- `summary.md`
- `promotion.json`

under a timestamped directory inside `eval/reports/`.

## Promotion rules

- zero deterministic regressions
- zero retrieval-required failures
- note-quality pass rate >= 0.90
