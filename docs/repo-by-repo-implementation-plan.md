# Repo-by-Repo Implementation Plan

This document turns the phased architecture delta into a concrete implementation plan with:

- repo-by-repo responsibilities
- exact package trees
- first file stubs
- sequencing by deliverable
- explicit ownership boundaries between:
  - `watchlist-review-clawbot`
  - `claw-identity`
  - optional future `clawmem` interfaces

This plan assumes the current baseline in `watchlist-review-clawbot` remains the contract:

- deterministic scoring remains authoritative
- LLM note generation remains downstream of deterministic evidence
- the current batch/regression suite remains the promotion gate

---

## 1. Architecture summary

### `watchlist-review-clawbot`
Owns:

- deterministic scoring policy
- final decision thresholds and labels
- watchlist-specific policy hooks
- runtime orchestration of evidence sources
- API review context assembly
- evaluation and promotion gates
- analyst note generation and normalization

This repo remains the **policy brain**.

### `claw-identity`
Owns:

- normalized entity comparison
- canonical matched-profile identity evidence
- OFAC profile and relationship evidence generation
- linked-document / linked-identifier evidence
- same-pair and same-profile historical evidence services once stabilized

This repo becomes the **identity evidence engine**.

### `clawmem`
Owns later:

- memory/retrieval substrate
- prior-case retrieval at scale
- example retrieval for explanation generation
- optional graph-context retrieval
- optional contradiction-pattern recall across large history

This repo becomes the **memory/retrieval layer**, but not the first home of deterministic policy.

---

## 2. Delivery order

### Priority 1 — Previous disposition scoring
Add a new deterministic score block:

- `previous_disposition_score`
- `previous_disposition_reason[]`

Apply it only when:

- the same screened entity and matched list UID are confidently linked to prior reviewed cases

### Priority 2 — OFAC relationship support
Ingest relationship-bearing OFAC XML/SLS into normalized tables and add:

- `relationship_support_score`
- `relationship_conflict_penalty`

### Priority 3 — Contradiction-pattern memory
Use normalized contradiction signatures to improve deterministic handling of:

- sparse false positives
- recurring wrong-context alerts
- repeated contradiction signatures

---

## 3. Deliverables

## Deliverable A — Disposition Memory Scoring v1
Add:

- normalized alert-history store
- same-pair recurrence features
- recency-decayed prior disposition score
- deterministic policy hooks

Acceptance:

- no regression in current 12-case suite
- add 4–6 new tests for repeated-pair scenarios

## Deliverable B — OFAC Relationship Scoring v1
Add:

- OFAC XML/SLS ingestion into normalized relationship tables
- direct-relationship and linked-doc scoring
- graph-derived explanation fields

Acceptance:

- organization and network-style cases score better deterministically
- no regression on current suite

## Deliverable C — Deterministic History + Relationship Promotion Suite
Add new eval cases for:

- repeated confirmed true match
- repeated false positive with same contradiction pattern
- related-party reinforcement
- wrong-context relationship neighborhood

Acceptance:

- expanded promotion suite passes
- deterministic regressions remain zero on prior calibrated cases

---

# 4. Repo 1 — `watchlist-review-clawbot`

## 4.1 Role in target state

`watchlist-review-clawbot` should be the place where final deterministic policy is decided.

It should:

- orchestrate compare/screening/history/relationship evidence
- compute watchlist-specific scores
- decide labels such as `escalate`, `investigate_next_step`, `close`
- expose the full deterministic explanation block in review context
- own evals and promotion rules

It should **not** become the long-term source of reusable identity graph logic.
That gets extracted later into `claw-identity`.

---

## 4.2 Package tree

```text
watchlist-review-clawbot/
├── cmd/
│   ├── review-worker-api/
│   ├── review-batch-eval/
│   ├── review-regression-runner/
│   └── export-note-tuning-set/
├── internal/
│   ├── api/
│   │   ├── types.go
│   │   ├── review_response.go
│   │   └── deterministic_evidence.go
│   ├── runtime/
│   │   ├── flow.go
│   │   ├── evidence_pipeline.go
│   │   └── dependencies.go
│   ├── scoring/
│   │   ├── types.go
│   │   ├── decision.go
│   │   ├── previous_disposition.go
│   │   ├── relationship_support.go
│   │   ├── contradiction_pattern.go
│   │   ├── policy.go
│   │   └── explanation.go
│   ├── history/
│   │   ├── types.go
│   │   ├── normalize.go
│   │   ├── lookup.go
│   │   ├── recency.go
│   │   ├── patterns.go
│   │   └── store.go
│   ├── ofacdata/
│   │   ├── xml_types.go
│   │   ├── ingest.go
│   │   ├── normalize.go
│   │   └── store.go
│   ├── ofacgraph/
│   │   ├── evidence.go
│   │   ├── neighbors.go
│   │   ├── documents.go
│   │   └── explanations.go
│   ├── notes/
│   │   ├── prompt.go
│   │   ├── service.go
│   │   ├── normalize.go
│   │   └── grounding_sanitize.go
│   ├── eval/
│   │   ├── batch.go
│   │   ├── note_quality.go
│   │   ├── promotion.go
│   │   └── deterministic_history_checks.go
│   └── models/
│       └── ollama/
└── eval/
    ├── requests/
    ├── specs/
    └── reports/
```

---

## 4.3 Package responsibilities

### `internal/api`
Responsibilities:

- response types for review output
- deterministic evidence block contracts
- history/relationship evidence exposure in API payloads

Add fields such as:

- `previous_disposition_summary`
- `relationship_evidence`
- `contradiction_pattern_evidence`
- `deterministic_explanation`

### `internal/runtime`
Responsibilities:

- orchestrate the evidence pipeline
- fetch compare/screening from `claw-identity`
- fetch history evidence locally at first
- fetch OFAC relationship evidence locally at first
- call scoring with a full evidence bundle

### `internal/scoring`
Responsibilities:

- compute all deterministic score components
- maintain score weights and threshold logic
- build decision reasons and explanation blocks

This package remains the canonical home for:

- `previous_disposition_score`
- `relationship_support_score`
- `relationship_conflict_penalty`
- contradiction-pattern adjustments

### `internal/history`
Responsibilities:

- normalized alert history lookup
- same-pair recurrence lookup
- same-profile recurrence lookup
- recency-decay helpers
- contradiction-pattern normalization and lookup

### `internal/ofacdata`
Responsibilities:

- ingest OFAC XML/SLS payloads
- normalize parties, identifiers, documents, sanctions entries, locations, relationships
- store them in relational tables

### `internal/ofacgraph`
Responsibilities:

- derive scoring evidence from normalized OFAC data
- direct relationship support
- linked-doc support
- neighborhood conflict indicators
- deterministic graph explanation fragments

### `internal/eval`
Responsibilities:

- preserve existing gold suite
- add history and relationship specific promotion checks
- fail on deterministic regressions

---

## 4.4 First file stubs

### `internal/scoring/previous_disposition.go`

```go
package scoring

import "github.com/clawbot-platform/watchlist-review-clawbot/internal/history"

type PreviousDispositionEvidence struct {
	SamePairCount                 int
	SamePairEscalateCount         int
	SamePairCloseCount            int
	SamePairInvestigateCount      int
	MostRecentDecisionLabel       string
	MostRecentAgeDays             int
	SamePairConfidenceQualified   bool
	RecencyWeightedDispositionHit float64
	Reasons                       []string
}

func ComputePreviousDispositionScore(ev *PreviousDispositionEvidence) (score int, reasons []string) {
	if ev == nil || !ev.SamePairConfidenceQualified {
		return 0, nil
	}

	// TODO: implement recency-decayed prior disposition score.
	_ = history.PairDispositionSummary{}
	return 0, nil
}
```

### `internal/scoring/relationship_support.go`

```go
package scoring

import "github.com/clawbot-platform/watchlist-review-clawbot/internal/ofacgraph"

type RelationshipEvidence struct {
	DirectRelationshipHits   int
	LinkedDocumentHits       int
	ProgramContextHits       int
	NeighborhoodConflicts    int
	RelationshipReasons      []string
	RelationshipPenaltyNotes []string
}

func ComputeRelationshipSupport(ev *ofacgraph.ScoreEvidence) (support int, penalty int, reasons []string) {
	if ev == nil {
		return 0, 0, nil
	}

	// TODO: implement direct-relationship and linked-doc scoring.
	return 0, 0, nil
}
```

### `internal/scoring/contradiction_pattern.go`

```go
package scoring

import "github.com/clawbot-platform/watchlist-review-clawbot/internal/history"

type ContradictionPatternEvidence struct {
	PatternKey              string
	PriorFalsePositiveCount int
	PriorEscalateCount      int
	PriorInvestigateCount   int
	Reasons                 []string
}

func ComputeContradictionPatternScore(ev *history.PatternOutcomeSummary) (score int, reasons []string) {
	if ev == nil {
		return 0, nil
	}

	// TODO: implement contradiction-pattern memory scoring.
	return 0, nil
}
```

### `internal/history/types.go`

```go
package history

import "time"

type PairKey struct {
	TenantID                   string
	ScreenedEntityFingerprint  string
	MatchedListUID             string
	MatchedProgram             string
}

type DispositionEvent struct {
	CaseID               string
	AlertID              string
	DecisionLabel        string
	DecisionReason       string
	ContradictionPattern string
	OccurredAt           time.Time
}

type PairDispositionSummary struct {
	Key                    PairKey
	Events                 []DispositionEvent
	MostRecent             *DispositionEvent
	RecencyWeightedSupport float64
}

type PatternOutcomeSummary struct {
	PatternKey         string
	FalsePositiveCount int
	EscalateCount      int
	InvestigateCount   int
}
```

### `internal/history/store.go`

```go
package history

import "context"

type Store interface {
	LookupSamePair(ctx context.Context, key PairKey) (*PairDispositionSummary, error)
	LookupPattern(ctx context.Context, patternKey string) (*PatternOutcomeSummary, error)
	RecordDisposition(ctx context.Context, key PairKey, ev DispositionEvent) error
}
```

### `internal/ofacdata/ingest.go`

```go
package ofacdata

import "context"

type IngestSource struct {
	SourceName string
	Bytes      []byte
}

type IngestResult struct {
	PartyCount         int
	RelationshipCount  int
	IdentifierCount    int
	DocumentCount      int
	LocationCount      int
	SanctionsEntryCount int
}

func IngestXML(ctx context.Context, src IngestSource, store Store) (*IngestResult, error) {
	// TODO: parse relationship-bearing OFAC XML/SLS payloads.
	_ = ctx
	_ = src
	_ = store
	return &IngestResult{}, nil
}
```

### `internal/ofacgraph/evidence.go`

```go
package ofacgraph

import "context"

type ScoreEvidence struct {
	MatchedProfileID       string
	DirectRelationshipHits int
	LinkedDocumentHits     int
	ProgramContextHits     int
	NeighborhoodConflicts  int
	Reasons                []string
}

type Builder interface {
	BuildScoreEvidence(ctx context.Context, matchedProfileID string) (*ScoreEvidence, error)
}
```

### `internal/runtime/evidence_pipeline.go`

```go
package runtime

import (
	"context"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/history"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/ofacgraph"
)

type EvidencePipeline struct {
	HistoryStore history.Store
	OFACGraph    ofacgraph.Builder
}

func (p *EvidencePipeline) LoadDeterministicEvidence(ctx context.Context) error {
	// TODO: orchestrate prior-disposition and relationship evidence loading.
	_ = ctx
	return nil
}
```

### `internal/eval/deterministic_history_checks.go`

```go
package eval

func ValidateDeterministicHistoryUpgrade() error {
	// TODO: add promotion checks for repeated-pair and relationship cases.
	return nil
}
```

---

## 4.5 New eval assets

Add request cases such as:

```text
eval/requests/history/
├── repeated_confirmed_true_match.json
├── repeated_false_positive_same_pattern.json
├── related_party_reinforcement.json
└── wrong_context_relationship_neighborhood.json
```

Add specs such as:

```text
eval/specs/
├── deterministic-history-v1.json
└── deterministic-history-relationships-v1.json
```

---

# 5. Repo 2 — `claw-identity`

## 5.1 Role in target state

`claw-identity` becomes the canonical service for **identity evidence**, not watchlist policy.

It should answer questions like:

- are these two records likely the same profile?
- what linked documents or identifiers corroborate that?
- what OFAC relationship neighborhood exists around the matched profile?
- what prior reviewed evidence exists for the same pair or same matched profile?

It should **not** decide final watchlist labels.

---

## 5.2 Package tree

```text
claw-identity/
├── cmd/
│   └── claw-identity-api/
├── internal/
│   ├── api/
│   │   ├── compare.go
│   │   ├── screening.go
│   │   ├── history_evidence.go
│   │   └── relationship_evidence.go
│   ├── compare/
│   │   ├── service.go
│   │   ├── name.go
│   │   ├── date.go
│   │   ├── geography.go
│   │   └── identifiers.go
│   ├── screening/
│   │   ├── service.go
│   │   └── ofac_candidates.go
│   ├── profiles/
│   │   ├── types.go
│   │   ├── canonical.go
│   │   └── matching.go
│   ├── ofacdata/
│   │   ├── ingest.go
│   │   ├── xml_types.go
│   │   ├── normalize.go
│   │   └── store.go
│   ├── relationships/
│   │   ├── service.go
│   │   ├── direct.go
│   │   ├── neighborhood.go
│   │   └── explanations.go
│   ├── linkeddocs/
│   │   ├── service.go
│   │   ├── identifiers.go
│   │   └── official_docs.go
│   ├── historyevidence/
│   │   ├── service.go
│   │   ├── pair_lookup.go
│   │   ├── profile_lookup.go
│   │   └── recency.go
│   └── storage/
│       ├── postgres/
│       └── migrations/
└── migrations/
```

---

## 5.3 Package responsibilities

### `internal/compare`
Responsibilities:

- canonical compare of screened vs matched entity
- name/date/geography/identifier comparison
- confidence bands
- source refs

### `internal/profiles`
Responsibilities:

- canonical matched-profile identity
- stable internal profile IDs
- mapping list UIDs and documents into canonical matched profile records

### `internal/ofacdata`
Responsibilities:

- normalized OFAC profile ingestion
- persistent relational representation of parties, entries, documents, relationships

### `internal/relationships`
Responsibilities:

- direct relationship support
- neighborhood analysis
- relationship conflict detection
- graph explanation fragments

### `internal/linkeddocs`
Responsibilities:

- corroboration from registration docs, tax IDs, passports, vessel IDs, etc.
- linked identifier evidence

### `internal/historyevidence`
Responsibilities:

- same-pair historical evidence lookup
- same-profile historical evidence lookup
- recency and support summaries

### `internal/api`
Responsibilities:

- expose evidence contracts to calling repos

---

## 5.4 API contracts to expose

### Compare response extension

```go
package api

type CompareEvidenceResponse struct {
	MatchedProfileID           string   `json:"matched_profile_id,omitempty"`
	ConfidenceBand             string   `json:"confidence_band,omitempty"`
	SamePairQualified          bool     `json:"same_pair_qualified,omitempty"`
	LinkedIdentifierSupport    []string `json:"linked_identifier_support,omitempty"`
	RelationshipSupport        []string `json:"relationship_support,omitempty"`
	RelationshipConflicts      []string `json:"relationship_conflicts,omitempty"`
	NeighborhoodSummary        []string `json:"neighborhood_summary,omitempty"`
	PriorDispositionSummary    []string `json:"prior_disposition_summary,omitempty"`
}
```

This gives `watchlist-review-clawbot` a reusable evidence payload without moving label policy into `claw-identity`.

---

## 5.5 First file stubs

### `internal/relationships/service.go`

```go
package relationships

import "context"

type Evidence struct {
	MatchedProfileID      string
	DirectSupportReasons  []string
	ConflictReasons       []string
	NeighborhoodSummary   []string
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) BuildEvidence(ctx context.Context, matchedProfileID string) (*Evidence, error) {
	// TODO: derive direct and neighborhood relationship evidence.
	_ = ctx
	_ = matchedProfileID
	return &Evidence{MatchedProfileID: matchedProfileID}, nil
}
```

### `internal/linkeddocs/service.go`

```go
package linkeddocs

import "context"

type Evidence struct {
	MatchedProfileID string
	SupportReasons   []string
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) BuildEvidence(ctx context.Context, matchedProfileID string) (*Evidence, error) {
	// TODO: derive linked-document and identifier support.
	_ = ctx
	_ = matchedProfileID
	return &Evidence{MatchedProfileID: matchedProfileID}, nil
}
```

### `internal/historyevidence/service.go`

```go
package historyevidence

import "context"

type PairEvidence struct {
	SamePairQualified       bool
	PriorDispositionReasons []string
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LookupSamePair(ctx context.Context, screenedFingerprint, matchedListUID string) (*PairEvidence, error) {
	// TODO: return same-pair history summaries for deterministic policy consumers.
	_ = ctx
	_ = screenedFingerprint
	_ = matchedListUID
	return &PairEvidence{}, nil
}
```

### `internal/ofacdata/store.go`

```go
package ofacdata

import "context"

type Store interface {
	UpsertParty(ctx context.Context, party any) error
	UpsertRelationship(ctx context.Context, relationship any) error
	UpsertDocument(ctx context.Context, document any) error
}
```

---

# 6. Repo 3 — optional future `clawmem` interfaces

## 6.1 Role in target state

`clawmem` should support retrieval and memory once the deterministic needs are proven.

Do **not** make it a mandatory first dependency for:

- same-pair prior disposition scoring
- direct OFAC relationship scoring
- contradiction-pattern scoring v1

Those should work with relational storage first.

`clawmem` comes in when you want:

- large-scale prior-case retrieval
- example retrieval for analyst-note generation
- graph-context retrieval beyond direct relationship scoring
- cross-workflow memory and institutional context

---

## 6.2 Interface tree

```text
clawmem/
├── internal/
│   ├── caseretrieval/
│   │   ├── service.go
│   │   ├── similar_cases.go
│   │   └── pair_history.go
│   ├── examplememory/
│   │   ├── service.go
│   │   ├── accepted_notes.go
│   │   └── explanation_examples.go
│   ├── graphretrieval/
│   │   ├── service.go
│   │   ├── neighborhood.go
│   │   └── path_summary.go
│   ├── patternretrieval/
│   │   ├── service.go
│   │   └── contradiction_patterns.go
│   └── api/
│       ├── retrieval.go
│       ├── history.go
│       └── graph.go
```

---

## 6.3 Suggested future interfaces

### `internal/api/history.go`

```go
package api

type CaseHistoryQuery struct {
	TenantID                  string `json:"tenant_id"`
	ScreenedEntityFingerprint string `json:"screened_entity_fingerprint,omitempty"`
	MatchedListUID            string `json:"matched_list_uid,omitempty"`
	DecisionLabel             string `json:"decision_label,omitempty"`
	Limit                     int    `json:"limit,omitempty"`
}

type CaseHistoryResult struct {
	CaseID         string   `json:"case_id"`
	AlertID        string   `json:"alert_id"`
	DecisionLabel  string   `json:"decision_label"`
	Reasons        []string `json:"reasons,omitempty"`
	OccurredAt     string   `json:"occurred_at"`
}
```

### `internal/api/graph.go`

```go
package api

type GraphNeighborhoodQuery struct {
	MatchedProfileID string `json:"matched_profile_id"`
	HopLimit         int    `json:"hop_limit,omitempty"`
}

type GraphNeighborhoodResult struct {
	MatchedProfileID string   `json:"matched_profile_id"`
	Summary          []string `json:"summary,omitempty"`
}
```

### `internal/api/retrieval.go`

```go
package api

type ExampleRetrievalQuery struct {
	TenantID      string   `json:"tenant_id"`
	DecisionLabel string   `json:"decision_label,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Limit         int      `json:"limit,omitempty"`
}

type ExampleRetrievalResult struct {
	ExampleID string   `json:"example_id"`
	Summary   string   `json:"summary"`
	Tags      []string `json:"tags,omitempty"`
}
```

---

## 6.4 First file stubs

### `internal/caseretrieval/service.go`

```go
package caseretrieval

import (
	"context"

	"github.com/your-org/clawmem/internal/api"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LookupHistory(ctx context.Context, q api.CaseHistoryQuery) ([]api.CaseHistoryResult, error) {
	// TODO: implement scalable history retrieval.
	_ = ctx
	_ = q
	return nil, nil
}
```

### `internal/examplememory/service.go`

```go
package examplememory

import (
	"context"

	"github.com/your-org/clawmem/internal/api"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) RetrieveExamples(ctx context.Context, q api.ExampleRetrievalQuery) ([]api.ExampleRetrievalResult, error) {
	// TODO: implement accepted-example retrieval.
	_ = ctx
	_ = q
	return nil, nil
}
```

### `internal/graphretrieval/service.go`

```go
package graphretrieval

import (
	"context"

	"github.com/your-org/clawmem/internal/api"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Neighborhood(ctx context.Context, q api.GraphNeighborhoodQuery) (*api.GraphNeighborhoodResult, error) {
	// TODO: implement graph neighborhood retrieval for future Neo4j/Qdrant-backed memory.
	_ = ctx
	_ = q
	return &api.GraphNeighborhoodResult{}, nil
}
```

---

# 7. Ownership matrix

| Capability | watchlist-review-clawbot | claw-identity | clawmem |
|---|---|---|---|
| Final decision label policy | Yes | No | No |
| Score weights / thresholds | Yes | No | No |
| Same-pair prior disposition scoring v1 | Yes | Later reusable evidence | No |
| OFAC XML/SLS ingestion first pass | Yes | Later canonical home | No |
| Reusable relationship evidence | Temporary | Yes | No |
| Linked-doc corroboration evidence | Temporary | Yes | No |
| Contradiction-pattern scoring v1 | Yes | No | Later optional retrieval |
| Promotion suite / regression gates | Yes | No | No |
| Prior-case retrieval at scale | No | Limited | Yes |
| Example retrieval for note generation | No | No | Yes |
| Graph-context retrieval beyond 1-hop/2-hop | No | Limited | Yes |

---

# 8. Recommended implementation sequence

## Step 1
Implement **Deliverable A** in `watchlist-review-clawbot`.

Why:

- fastest win
- easiest to validate
- fully protected by existing promotion suite

## Step 2
Implement contradiction-pattern memory in `watchlist-review-clawbot`.

Why:

- helps sparse and wrong-context cases immediately
- no additional external system required

## Step 3
Implement **Deliverable B** locally in `watchlist-review-clawbot`.

Why:

- fastest way to prove relationship-based deterministic value
- lets you iterate before extracting APIs

## Step 4
Extract reusable evidence-building logic into `claw-identity`.

Why:

- relationship and identity evidence should become reusable across future clawbots
- policy thresholds must remain local to watchlist workflow

## Step 5
Define optional `clawmem` interfaces once you know what history/retrieval patterns are worth keeping.

Why:

- prevents early over-engineering
- keeps deterministic scoring clean and testable

---

# 9. Definition of done by repo

## `watchlist-review-clawbot`
Done when:

- previous disposition scoring is in production path
- contradiction-pattern memory is active
- OFAC relationship evidence is active
- new eval cases are added
- promotion still passes on prior suite

## `claw-identity`
Done when:

- reusable compare + relationship + linked-doc + same-pair evidence API exists
- watchlist-review-clawbot can consume identity evidence without reimplementing it

## `clawmem`
Done later when:

- case retrieval / example retrieval / graph retrieval are needed at scale
- those interfaces improve real workflows rather than just mirroring relational queries

---

# 10. Immediate next implementation target

Start in `watchlist-review-clawbot` with:

1. `internal/history`
2. `internal/scoring/previous_disposition.go`
3. `internal/scoring/contradiction_pattern.go`
4. `internal/runtime/evidence_pipeline.go`
5. new eval fixtures for repeated-pair and contradiction-memory scenarios

That gives the highest deterministic ROI with the least architecture churn.
