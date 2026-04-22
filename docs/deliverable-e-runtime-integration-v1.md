# Deliverable E — Runtime Integration + Persistence v1

## What this bundle changes
- Replaces starter-only history storage with Postgres-backed persistence + lookup.
- Replaces starter-only OFAC XML ingest with XML parse → normalize → persist flow.
- Adds real normalized OFAC tables for parties, aliases, identifiers, documents, locations, relationships, and sanctions entries.
- Adds a Postgres-backed relationship query service and runtime enrichers that apply history + relationship evidence to live scoring.

## Intended pipeline order
1. Parse alert
2. Extract features
3. Resolve matched list UID / identity evidence
4. Run `HistoryEnricher.Enrich(...)`
5. Run `RelationshipEnricher.Enrich(...)`
6. Continue with deterministic decision derivation / analyst note generation

## Exit criteria
- Organization/network-style cases improve deterministically
- No regression on existing suite
- Review context includes both `history_evidence` and `relationship_evidence`
