# Watchlist Review Clawbot Strategy Baseline

**Status:** Working baseline  
**Date:** 2026-04-14  
**Purpose:** Capture the current agreed strategy for `watchlist-review-clawbot` so implementation stays aligned to design decisions and drift can be detected early.

---

## 1. Why this document exists

This document records the current product and technical strategy for Watchlist Review Clawbot.

It is intended to:
- preserve the current design decisions in one place
- define what the first implementation should and should not do
- make scoring, thresholds, and alert scope explicit
- provide a progress checklist for future implementation phases
- prevent accidental drift back into old assumptions or scope creep

This strategy updates the earlier Watchlist Review Clawbot concept to align with the current Clawbot platform:
- `clawbot-server` as control plane
- `claw-identity` for compare and OFAC screening
- `clawmem` for scoped memory
- `thinkpad-p50` as data / RAG / memory node
- `ai-precision` as inference node
- Granite model usage aligned with `ach-trust-lab`

---

## 2. Core product boundary

Watchlist Review Clawbot is a **review assistant** for sanctions and watchlist alerts produced by existing screening systems.

It is **not**:
- a sanctions screening engine
- an entity-matching engine
- a control-plane system
- a memory engine
- a fraud or AML transaction-monitoring engine

Its job is to:
- normalize vendor-shaped alerts
- extract deterministic evidence
- call shared identity and screening services
- retrieve policy and operational context
- generate conservative, human-reviewable recommendations
- package outputs as structured JSON plus analyst-readable note

---

## 3. Platform-aligned architecture

### 3.1 Shared platform responsibilities

#### `clawbot-server`
Owns:
- run and cycle orchestration
- policy and governance
- reviewer workflow actions
- artifact registration
- audit sequencing
- platform APIs
- async event coordination over NATS

#### `claw-identity`
Owns:
- record-to-record compare
- OFAC screening candidate retrieval and scoring
- explanation packaging
- identity and screening event emission

#### `clawmem`
Owns:
- case and cycle continuity memory
- reviewer memory and summaries
- scoped memory namespaces

#### `thinkpad-p50`
Hosts:
- Qdrant
- Neo4j
- RAG corpora
- memory and retrieval data
- policy and sanctions knowledge indexes

#### `ai-precision`
Hosts:
- Granite 3.3 8B primary reasoning model
- Granite Guardian 3.3 8B guardrail model
- Granite 4 3B helper model
- inference routing / model serving

### 3.2 `watchlist-review-clawbot` responsibilities

This repo should own:
- alert normalization
- canonical alert schema
- deterministic feature extraction
- retrieval request construction
- prompt assembly
- review output parsing
- recommendation packaging
- analyst note generation
- evaluation harness
- fixtures and vertical documentation

This repo should **not** own:
- identity matching internals
- OFAC candidate matching internals
- control-plane orchestration
- artifact registry internals
- memory engine internals

---

## 4. Types of alerts in scope

### 4.1 Primary alert categories

The initial worker should support **watchlist screening alerts**, not generic transaction-monitoring or fraud alerts.

Two alert families are in scope first:

#### A. Party-based screening alerts
Not tied to a live transaction.

Examples:
- customer onboarding screening hit
- business onboarding screening hit
- vendor onboarding screening hit
- periodic rescreening hit
- customer profile change rescreening hit

#### B. ACH-linked party screening alerts
Tied to a payment context, but still fundamentally screening alerts.

Examples:
- ACH originator screening hit
- ACH beneficiary screening hit
- ACH business/company-name screening hit
- ACH alert with sparse party data
- ACH alert with conflicting party data

### 4.2 Explicitly out of scope for initial implementation

- generic AML transaction-monitoring alerts
- fraud alerts
- anomaly detection alerts
- behavior/risk engine outputs
- non-screening entity matching workflows

### 4.3 Recommended first implementation mix

Start with:
- individual onboarding alerts
- organization onboarding alerts
- ACH originator alerts
- ACH beneficiary alerts

This gives enough variety for:
- canonical schema design
- deterministic extraction
- identity enrichment
- conservative review logic

---

## 5. Canonical alert structure

Each alert should be represented in three layers.

### 5.1 Alert metadata
- `alert_id`
- `source_system`
- `created_at`
- `alert_type`
- `jurisdiction`
- `review_priority`
- `case_id` if present

### 5.2 Screened party and matched party

#### Screened party
- `name`
- `aliases`
- `dob` or incorporation date
- `countries`
- `addresses`
- `identifiers`

#### Matched party
- `list_source`
- `program`
- `name`
- `aliases`
- `dob`
- `countries`
- `identifiers`

### 5.3 Optional transaction context
Only for ACH-linked or future transaction-linked alerts.

- `transaction_id`
- `rail_type`
- `amount`
- `currency`
- `originator_role`
- `beneficiary_role`
- `payment_reference`
- `country_corridor`
- `institution_context`

---

## 6. Label set

### 6.1 Primary analyst outcome labels
- `close`
- `investigate_next_step`
- `escalate`

### 6.2 System/governance status label
- `review_pending`

Use `review_pending` only when the system cannot safely complete a normal recommendation, for example:
- guardrail failure
- retrieval failure
- required evidence unavailable
- runtime degradation preventing safe recommendation

### 6.3 Intended label behavior

#### False positives
Mostly:
- `close`

Sometimes:
- `investigate_next_step`

Rarely:
- `escalate`

#### True matches
Mostly:
- `escalate`

Sometimes:
- `investigate_next_step`

Almost never:
- `close`

---

## 7. Real-match vs false-positive ratios

Use **three datasets**, not one universal ratio.

### 7.1 Workflow simulation set
Use this to approximate analyst queue conditions.

- **10% true matches**
- **90% false positives**

### 7.2 Balanced reasoning/evaluation set
Use this for prompt behavior, schema checks, and deterministic scoring evaluation.

- **35% true matches**
- **65% false positives**

### 7.3 Adversarial hard-case set
Use this to stress ambiguity handling.

- **20% true matches**
- **80% false positives**

### 7.4 Initial starter gold set recommendation
Start with a **100-case gold set**:
- 35 true matches
- 65 false positives

Then expand to a **250-case workflow simulation set**:
- 25 true matches
- 225 false positives

---

## 8. False-positive taxonomy

False positives should be categorized explicitly.

### 8.1 Overall recommended false-positive mix
- **35% Name similarity**
- **20% Transliteration / script variation**
- **25% Incomplete data**
- **10% Wrong context / wrong entity type**
- **10% Mixed / compound cases**

### 8.2 Category definitions

#### Name similarity
Examples:
- common first/last name collisions
- weak token overlap
- reordered names
- company-name overlap without corroboration

#### Transliteration / script variation
Examples:
- Arabic / Persian romanization variation
- Cyrillic to Latin variation
- alternate transliterations of the same source-script name

#### Incomplete data
Examples:
- missing DOB
- missing identifier
- missing address
- sparse ACH party data
- missing organization registration number

#### Wrong context / wrong entity type
Examples:
- vessel matched to individual
- individual matched to organization
- company matched to individual alias
- single-word vessel names causing bad hits

#### Mixed / compound false positives
Examples:
- moderate name similarity plus sparse data
- transliteration ambiguity plus wrong context
- company-name overlap plus missing identifiers

### 8.3 Alert-family-specific false-positive mix

#### Onboarding / party screening alerts
- 40% name similarity
- 25% transliteration
- 25% incomplete data
- 5% wrong context
- 5% mixed

#### ACH-linked screening alerts
- 30% name similarity
- 15% transliteration
- 30% incomplete data
- 15% wrong context
- 10% mixed

---

## 9. True-match composition

True matches should not all be easy.

### Recommended true-match split
- **40% strong true matches**
- **40% moderate true matches**
- **20% weak but plausible true matches**

### Strong true matches
Examples:
- exact DOB + identifier + country support
- exact alias and strong corroborative data

### Moderate true matches
Examples:
- no exact identifier, but multiple aligned attributes
- alias + DOB year + geography + program context

### Weak but plausible true matches
Examples:
- sparse but concerning evidence
- enough combined support to prevent safe closure

---

## 10. Scoring strategy overview

The worker should not use one single score.

Use:
- **match_strength_score (MSS)**
- **data_sufficiency_score (DSS)**
- explicit **contradiction flags**
- optional **context consistency signal**

### 10.1 Why separate scores are needed

A case can have:
- medium match strength
- low data sufficiency

This is exactly the type of case that should become `investigate_next_step` rather than `escalate`.

---

## 11. Match strength score (MSS)

### 11.1 Purpose

MSS answers:

**How plausible is it that the screened party and the matched watchlist party are the same entity?**

### 11.2 MSS components

#### Name evidence: 0–35
Use:
- exact normalized full-name match
- exact alias-to-name match
- transliteration-equivalent match
- reordered or partial strong token match
- weak fuzzy support only as supporting evidence

#### Date evidence: 0–20
Use:
- exact DOB
- year-only or partial date support
- incorporation date where appropriate

#### Identifier evidence: 0–20
Use:
- passport
- national ID
- tax ID
- registration number
- IMO number
- other strong identifiers

#### Country / geography evidence: 0–10
Use:
- nationality
- country
- registration country
- geography alignment

#### Address evidence: 0–10
Use:
- exact or strong partial address support
- city/state/country support where usable

#### Context support: 0–5
Use:
- entity-type alignment
- alert context fit
- role/context plausibility

### 11.3 MSS formula

**MSS = Name + Date + Identifier + Country + Address + ContextSupport**

Cap between 0 and 100.

### 11.4 Critical MSS rule

If the score above 70 is driven only by name/transliteration evidence, cap the score at **69**.

This prevents name-only false positives from being escalated too aggressively.

---

## 12. Data sufficiency score (DSS)

### 12.1 Purpose

DSS answers:

**Do we have enough reliable data to make a confident recommendation?**

### 12.2 DSS components

#### Screened party completeness: 0–30
- name
- aliases
- DOB/incorp date
- country
- address
- identifiers

#### Matched party completeness: 0–25
- name
- aliases
- DOB/incorp date
- geography
- identifiers

#### Identifier quality: 0–20
- strong identifiers on one or both sides
- identifier usefulness, not just presence

#### Geography quality: 0–10
- usable address or geography detail
- specific rather than vague context

#### Supporting context: 0–15
- analyst notes
- vendor flags
- supporting attachments
- transaction context for ACH alerts

### 12.3 DSS formula

**DSS = ScreenedCompleteness + MatchedCompleteness + IdentifierQuality + GeoQuality + SupportingContext**

Cap between 0 and 100.

---

## 13. Contradictions and blockers

Do not bury contradictions inside the score alone.

Track them explicitly.

### 13.1 Contradiction types
- `dob_conflict`
- `identifier_conflict`
- `entity_type_conflict`
- `geography_conflict`
- `context_conflict`

### 13.2 Hard negative blockers
These should prevent escalation unless unusually strong counter-evidence exists:
- strong entity-type mismatch
- explicit DOB contradiction
- explicit strong identifier contradiction
- clearly incompatible country profile without supporting identifiers
- clearly wrong alert context

### 13.3 Mandatory-review conditions
These should force `investigate_next_step` or `review_pending`:
- DSS too low
- transliteration-heavy case with sparse corroboration
- mid-range MSS with ambiguity
- model and deterministic logic disagree materially
- guardrail failure or service degradation

---

## 14. Threshold policy

### 14.1 Baseline thresholds

#### `close`
Use when:
- **MSS <= 30**
- no mandatory-review rule triggered
- no strong unresolved contradiction requiring review

#### `investigate_next_step`
Use when:
- **31 <= MSS <= 69**
- or **DSS < 50**
- or contradictions exist
- or the case is sparse, transliteration-heavy, or context-ambiguous

#### `escalate`
Use when:
- **MSS >= 70**
- **DSS >= 50**
- no hard contradiction
- evidence includes more than name-only similarity

#### `review_pending`
Use when:
- guardrail failed
- retrieval failed critically
- required evidence unavailable
- system cannot safely return a normal recommendation

### 14.2 Practical decision table

| MSS | DSS | Context | Label |
|---|---:|---|---|
| 0–30 | any | neutral/negative | `close` |
| 31–49 | any | any | `investigate_next_step` |
| 50–69 | <50 | any | `investigate_next_step` |
| 50–69 | >=50 | strongly negative | `investigate_next_step` |
| 50–69 | >=50 | neutral/positive | `investigate_next_step` |
| 70–100 | <50 | any | `investigate_next_step` |
| 70–100 | >=50 | contradiction present | `investigate_next_step` |
| 70–100 | >=50 | neutral/positive | `escalate` |

---

## 15. Name-matching strategy for first pass

The first pass should use a **hybrid, script-aware approach**, not one universal fuzzy algorithm.

### 15.1 Matching sequence

1. exact canonical full-name match
2. exact alias-to-name match
3. transliteration-equivalent match
4. token reorder / partial strong token match
5. weak fuzzy support as supporting evidence only

### 15.2 Why hybrid matching is required

A pure edit-distance strategy is not good enough for multilingual OFAC-style name matching, especially across scripts. Transliteration and script-family effects must be handled explicitly.

### 15.3 Script-family-aware emphasis

Given the current OFAC distribution, first-pass logic should emphasize:

#### Slavic / Eastern European (Cyrillic)
- transliteration-aware matching
- variant comparison across Cyrillic and Latin forms
- token-order robustness

#### Middle Eastern / Arabic / Persian
- transliteration-aware comparison
- alias-heavy comparison
- conservative handling when only romanized ambiguity is present

#### Latin American / Hispanic
- strong token and surname handling
- double-surname robustness
- token-order awareness

#### Asian
- exact native-script or exact romanized alias where available
- conservative behavior for fuzzy-only romanized ambiguity

#### African / Other / Unknown
- generic hybrid path
- exact + alias + token + limited fuzzy support

### 15.4 Name-match scoring ladder

#### Exact normalized full-name match
Highest score within name bucket.

#### Exact alias-to-name match
Almost as strong as exact full-name match.

#### Strong transliteration-equivalent match
Score below exact/alias exact but above generic fuzzy.

#### Reordered or partial strong token match
Moderate score band.

#### Weak fuzzy overlap
Low supporting score only.

### 15.5 First-pass implementation principle

Do not let plain edit distance dominate multilingual name scoring.

---

## 16. Model strategy

### 16.1 Initial runtime model stack
- **Primary review reasoning:** Granite 3.3 8B
- **Guardrail / compliance gate:** Granite Guardian 3.3 8B
- **Helper / repair model:** Granite 4 3B

### 16.2 Initial adaptation strategy
Prefer:
- deterministic evidence packaging
- prompt engineering
- schema constraints
- retrieval grounding

Delay LoRA / QLoRA until:
- canonical datasets exist
- gold labels exist
- metrics are stable
- output schema is stable

---

## 17. RAG and MCP strategy

### 17.1 RAG
Use retrieval for:
- policy and SOP passages
- sanctions program context
- escalation guidance
- internal review notes
- prior case summaries and memory

### 17.2 MCP
MCP should be used as an internal tool layer, not as the entire runtime.

Planned MCP surfaces:
- Knowledge MCP
- Identity MCP
- Memory MCP
- Artifact / Control MCP

HTTP remains the stable service boundary.
NATS remains the async event boundary.
MCP is the tool abstraction layer for agentic access.

---

## 18. Output contract

The review worker output should include:
- `decision_label`
- `match_strength_score`
- `data_sufficiency_score`
- `decision_reason`
- `evidence_for`
- `evidence_against`
- `missing_information`
- `next_step`
- `identity_trace_refs`
  - `decision_trace_id`
  - `explanation_id`
  - `screening_id`
- `retrieval_refs`
- `analyst_note`
- `guardrail_status`

---

## 19. Implementation phases

### Phase 1 — Canonical alert worker
Deliver:
- repo skeleton
- canonical alert schema
- one parser
- deterministic feature extraction
- `claw-identity` integration
- conservative packaged output without LLM

### Phase 2 — Granite-assisted review MVP
Deliver:
- Granite 3.3 8B reasoning path
- structured JSON recommendation
- analyst note generation

### Phase 3 — Guardrail and dual mode
Deliver:
- Granite Guardian 3.3 8B guardrail
- dual-mode comparison between deterministic evidence and model reasoning

### Phase 4 — RAG + MCP integration
Deliver:
- retrieval integration
- policy and SOP grounding
- MCP tool integrations

### Phase 5 — Memory and reviewer feedback
Deliver:
- case/cycle memory continuity
- reviewer feedback loop

### Phase 6 — Evaluation and portfolio hardening
Deliver:
- false-positive reduction metrics
- escalation accuracy
- JSON validity
- explanation usefulness
- portfolio hardening

---

## 20. Progress tracker

Use this section to track alignment against the current strategy.

### 20.1 Scope and architecture
- [ ] Review-assistant-only boundary preserved
- [ ] No screening-engine logic implemented locally
- [ ] `claw-identity` used for compare / OFAC screening
- [ ] `clawbot-server` kept as control plane
- [ ] `clawmem` kept as memory layer
- [ ] Granite model stack aligned with platform

### 20.2 Alert universe
- [ ] Individual onboarding alerts defined
- [ ] Organization onboarding alerts defined
- [ ] ACH originator alerts defined
- [ ] ACH beneficiary alerts defined
- [ ] Transaction context schema defined

### 20.3 Dataset design
- [ ] False-positive taxonomy finalized
- [ ] 100-case gold set designed
- [ ] Workflow simulation dataset designed
- [ ] Hard-case dataset designed
- [ ] True-match strength tiers defined

### 20.4 Scoring and labels
- [ ] MSS formula implemented
- [ ] DSS formula implemented
- [ ] Contradiction flags implemented
- [ ] Threshold policy implemented
- [ ] Name-matching first pass implemented
- [ ] Wrong-context handling implemented

### 20.5 Model and retrieval
- [ ] Granite 3.3 8B review path implemented
- [ ] Granite Guardian 3.3 8B guardrail path implemented
- [ ] Retrieval path implemented
- [ ] MCP tool plan documented

### 20.6 Evaluation
- [ ] False-positive reduction metric defined
- [ ] Escalation accuracy metric defined
- [ ] JSON validity checks defined
- [ ] Explanation usefulness review defined

---

## 21. Change-control rule

Any future implementation change that materially alters:
- alert scope
- label set
- false-positive ratio assumptions
- score definitions
- threshold policy
- name-matching strategy
- platform boundaries

should update this document.

This document is the baseline. Code should follow it unless there is an explicit decision to revise it.
