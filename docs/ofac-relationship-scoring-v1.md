# OFAC Relationship Scoring v1

## Objective

Use relationship-bearing OFAC data to enrich deterministic scoring without making the LLM a scoring dependency.

## New deterministic score blocks

Add the following score components to the main deterministic result:

- `relationship_support_score`
- `relationship_conflict_penalty`
- `official_doc_link_score`
- `program_context_score`
- `relationship_reasons[]`

## Data sources

This deliverable assumes ingest of relationship-bearing OFAC content into normalized relational tables:

- `ofac_party`
- `ofac_sanctions_entry`
- `ofac_identifier`
- `ofac_document`
- `ofac_location`
- `ofac_relationship`

Optional materialized helper views:

- `ofac_party_neighbors`
- `ofac_party_program_context`
- `ofac_linked_docs`

## Scoring behavior

### Positive support
Examples:
- direct relationship to the matched sanctioned profile
- linked registration/tax/passport/vessel/aircraft identifiers
- coherent program context
- shared location or legal-registration evidence that reinforces the matched profile

### Conflict / penalty
Examples:
- relationship neighborhood inconsistent with screened entity context
- linked docs that contradict the screened profile
- neighborhood evidence pointing to a different entity family than the candidate under review

## First integration path

1. Parse and normalize OFAC data in `internal/ofacdata`
2. Build relationship evidence in `internal/ofacgraph`
3. Score evidence in `internal/scoring/relationship_support.go`
4. Attach evidence to `review_context.relationship_evidence`
5. Add tests and eval cases

## Initial weights

Starter-only guidance:

- `relationship_support_score`: 0..15
- `relationship_conflict_penalty`: 0..10
- `official_doc_link_score`: 0..10
- `program_context_score`: 0..5

Actual weights should be tuned inside the main repo against eval outcomes.
