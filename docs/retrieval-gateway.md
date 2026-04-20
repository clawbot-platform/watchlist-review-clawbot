# Retrieval Gateway

This update fixes the local developer path so `/v1/search` returns snippets instead of 404 by adding a `local_json` backend for the gateway.

Supported backends:
- `local_json`
- `clawmem_http`
- `pgvector`
- `qdrant`

## Fast local path

```bash
export RETRIEVAL_GATEWAY_ADDR=':8088'
export RETRIEVAL_GATEWAY_BACKEND='local_json'
export RETRIEVAL_LOCAL_JSON_PATH='eval/retrieval/snippets.json'
go run ./cmd/retrieval-gateway
```

Then point the worker at the gateway:

```bash
export ENABLE_REVIEW_RETRIEVAL='true'
export REVIEW_RETRIEVAL_BASE_URL='http://127.0.0.1:8088'
export REVIEW_RETRIEVAL_TIMEOUT='5s'
```

## Real backend paths

### Clawmem HTTP

```bash
export RETRIEVAL_GATEWAY_BACKEND='clawmem_http'
export CLAWMEM_BASE_URL='http://thinkpad-p50:8087'
```

### pgvector

```bash
export RETRIEVAL_GATEWAY_BACKEND='pgvector'
export PGVECTOR_DSN='postgres://postgres:postgres@localhost:5432/clawmem?sslmode=disable'
export PGVECTOR_TABLE='rag_documents'
export RETRIEVAL_EMBED_BASE_URL='http://ai-precision:11434'
export RETRIEVAL_EMBED_MODEL='embeddinggemma'
```

### Qdrant

```bash
export RETRIEVAL_GATEWAY_BACKEND='qdrant'
export QDRANT_BASE_URL='http://thinkpad-p50:6333'
export QDRANT_COLLECTION='rag_documents'
export RETRIEVAL_EMBED_BASE_URL='http://ai-precision:11434'
export RETRIEVAL_EMBED_MODEL='embeddinggemma'
```
