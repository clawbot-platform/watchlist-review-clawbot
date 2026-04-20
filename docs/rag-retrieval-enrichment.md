# Retrieval/RAG Enrichment for Granite Notes

This layer enriches the Granite prompt with retrieval context while keeping deterministic decisioning unchanged.

Ollama documents a stable local API base URL and provides a dedicated `POST /api/embed` endpoint for generating embeddings, which is a common local building block for RAG pipelines. The note layer here assumes retrieval is provided by a separate service, such as Clawmem or a future retrieval gateway, and only injects top snippets into the Granite prompt as supporting background. citeturn559137view0turn416933view0

## Design intent

- deterministic scores and labels remain authoritative
- retrieval only enriches the Granite analyst note prompt
- missing retrieval should never break review generation
- retrieval snippets must not override deterministic evidence

## Suggested env

```bash
export ENABLE_REVIEW_RETRIEVAL='true'
export REVIEW_RETRIEVAL_BASE_URL='http://thinkpad-p50:8088'
export REVIEW_RETRIEVAL_TIMEOUT='5s'
```

## Suggested retrieval contract

The worker expects an HTTP retrieval service with:

- `POST /v1/search`

Request:
```json
{
  "tenant_id": "test-tenant",
  "case_id": "case-123",
  "alert_id": "alert-123",
  "text": "individual_onboarding | screened=Jane Citizen | matched=Jane Citizen | decision=escalate",
  "top_k": 4,
  "tags": ["individual_onboarding", "individual", "sdn"]
}
```

Response:
```json
{
  "snippets": [
    {
      "snippet_id": "snip-1",
      "source": "clawmem",
      "title": "Prior review",
      "text": "Prior case involved passport corroboration.",
      "score": 0.91,
      "tags": ["sdn", "passport"]
    }
  ]
}
```

## Notes

This is intentionally retrieval-service-agnostic:
- Clawmem can back it
- a pgvector service can back it
- a Qdrant/Qdrant+metadata service can back it

The Granite prompt receives only the final snippets, not the retrieval implementation details.
