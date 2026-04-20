# Runtime Flow

1. ingest alert/case payload
2. normalize to canonical schema
3. extract deterministic features
4. enrich via `claw-identity`
5. retrieve policy/context
6. run Granite review reasoning
7. apply guardrail and conservative shaping
8. emit structured output + note + trace refs
