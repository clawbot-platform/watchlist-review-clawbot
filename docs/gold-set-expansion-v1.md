# Gold Set Expansion v1

This pack expands the gold set from the current green 2-case baseline to a 12-case starter suite.

## Coverage

- individual true match
- individual retrieval-required true match
- individual sparse false positive
- individual wrong-context overlap
- organization true match
- organization false-positive name overlap
- ACH sparse false positive
- ACH strong true match
- real-OFAC-inspired individual false positive
- real-OFAC-inspired wrong-context vessel case
- alias-plus-identifier individual true match
- organization wrong-context country conflict

## Run

```bash
export REVIEW_BATCH_SPEC_PATH='eval/goldset/manifest.expanded.v1.json'
export REVIEW_BATCH_EVAL_BASE_URL='http://127.0.0.1:8090'
export REVIEW_BATCH_EVAL_TIMEOUT='60s'
go run ./cmd/review-batch-eval
```

## Expanded local retrieval path

```bash
export RETRIEVAL_GATEWAY_BACKEND='local_json'
export RETRIEVAL_LOCAL_JSON_PATH='eval/retrieval/snippets.expanded.v1.json'
go run ./cmd/retrieval-gateway
```

## Notes

This is an additive expansion pack. It does not replace the current promoted 2-case baseline automatically.
Use it to calibrate the next promoted manifest.
