# End-to-End Local Run Against Live `claw-identity`

## Prerequisites

- `claw-identity` is already running locally and reachable
- `compare-bridge` is already running if `claw-identity` is in bridge mode
- local Postgres data is seeded as previously validated
- the fixture alert file exists at:
  - `test/fixtures/screeningjson/alert_with_source_refs.json`

## 1. Start the worker HTTP API

```bash
export WATCHLIST_REVIEW_HTTP_ADDR=':8090'
export WATCHLIST_REVIEW_DEFAULT_TENANT='test-tenant'
export CLAWBOT_IDENTITY_BASE_URL='http://localhost:8080'

go run ./cmd/review-worker-api
```

## 2. Health check

```bash
curl -s http://localhost:8090/healthz
```

Expected:

```json
{"ok":true}
```

## 3. Invoke `/v1/reviews`

```bash
jq -n   --arg tenant "test-tenant"   --arg caseid "case-screening-001"   --arg source "screening_json"   --slurpfile alert test/fixtures/screeningjson/alert_with_source_refs.json   '{
    tenant_id: $tenant,
    case_id: $caseid,
    source_system: $source,
    raw_alert: $alert[0],
    options: {
      explain: true,
      mode: "deterministic"
    }
  }' | curl -s -X POST http://localhost:8090/v1/reviews   -H 'Content-Type: application/json'   -H 'X-Correlation-ID: review-worker-e2e-001'   --data @-
```

## 4. Expected behavior

The response should include:

- `status: "review_context_built"`
- `alert_id`
- `case_id`
- `review_context.features`
- `identity_trace_refs.screening_id`
- `identity_trace_refs.decision_trace_id`
- usually `identity_trace_refs.explanation_id` if compare succeeds
- empty or small `warnings`

Because the fixture includes:
- `screened_party.source_system = "kyc_applications"`
- `screened_party.source_record_id = "left-record"`
- `matched_party.list_source = "watchlist_candidates"`
- `matched_party.list_uid = "right-record"`

the compare path should be available in addition to OFAC screening.

---

# Thin `clawbot-server` integration

## Planned route

```text
POST /api/v1/watchlist-review/reviews
```

This route should simply:
- accept tenant/case/source_system/raw_alert/options
- generate or propagate `X-Correlation-ID`
- call the worker's `POST /v1/reviews`
- return the worker response

See the `clawbot-server/` folder in this pack for thin client and handler code.
