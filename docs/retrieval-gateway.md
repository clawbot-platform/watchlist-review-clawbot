# Retrieval Gateway

This gateway exposes `POST /v1/search` so the worker can use a real retrieval backend without changing the worker-facing retrieval contract.

Supported backends:
- `clawmem_http`
- `pgvector`
- `qdrant`

## Common env

```bash
export RETRIEVAL_GATEWAY_ADDR=':8088'
export RETRIEVAL_GATEWAY_BACKEND='clawmem_http'
```

## Clawmem HTTP backend

```bash
export RETRIEVAL_GATEWAY_BACKEND='clawmem_http'
export CLAWMEM_BASE_URL='http://thinkpad-p50:8087'
```

## pgvector backend

```bash
export RETRIEVAL_GATEWAY_BACKEND='pgvector'
export PGVECTOR_DSN='postgres://postgres:postgres@localhost:5432/clawmem?sslmode=disable'
export PGVECTOR_TABLE='rag_documents'
export RETRIEVAL_EMBED_BASE_URL='http://ai-precision:11434'
export RETRIEVAL_EMBED_MODEL='embeddinggemma'
```

Expected table contract:

```sql
CREATE TABLE rag_documents (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  source text,
  title text,
  body text NOT NULL,
  tags jsonb NOT NULL DEFAULT '[]'::jsonb,
  embedding vector(768) NOT NULL
);
```

## Qdrant backend

```bash
export RETRIEVAL_GATEWAY_BACKEND='qdrant'
export QDRANT_BASE_URL='http://thinkpad-p50:6333'
export QDRANT_COLLECTION='rag_documents'
export RETRIEVAL_EMBED_BASE_URL='http://ai-precision:11434'
export RETRIEVAL_EMBED_MODEL='embeddinggemma'
```
