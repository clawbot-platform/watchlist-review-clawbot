# Granite-Assisted Analyst Notes

This layer is advisory only.

The worker keeps:
- deterministic scoring
- deterministic contradictions
- deterministic decision label

Granite adds:
- analyst note draft
- evidence summary
- missing information summary
- next-step rationale phrasing

## Suggested env

```bash
export ENABLE_GRANITE_ANALYST_NOTES='true'
export MODEL_PROVIDER='ollama'
export INFERENCE_BASE_URL='http://127.0.0.1:11434'
export PRIMARY_MODEL='granite3.3:8b'
```

The Ollama implementation uses `/api/generate` with:
- `stream: false`
- a JSON schema in `format`

This keeps outputs structured and easier to validate.
