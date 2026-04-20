# Watchlist Review Clawbot: Scoring Matrix

## Purpose

This document defines the deterministic scoring baseline for Watchlist Review Clawbot.

It is intended to:
- make implementation choices explicit
- keep alert review behavior conservative and explainable
- prevent drift away from agreed scoring and threshold logic
- provide a stable baseline before adding Granite reasoning and guardrails

This worker remains a **review assistant**, not a screening engine or entity-matching engine. Existing screening systems remain the system of record, and conservative human-reviewable output remains mandatory.

---

## 1. Core scoring outputs

For every alert, produce:

- `match_strength_score` (`MSS`) → 0 to 100
- `data_sufficiency_score` (`DSS`) → 0 to 100
- `context_consistency` → `positive`, `neutral`, `negative`
- `contradictions[]` → typed contradictions/blockers
- `decision_label` → `close`, `investigate_next_step`, `escalate`, or `review_pending`

### 1.1 Match Strength Score (MSS)

MSS answers:

> How plausible is it that the screened party and the matched party are the same entity?

### 1.2 Data Sufficiency Score (DSS)

DSS answers:

> Do we have enough reliable data to make a safe recommendation?

### 1.3 Contradictions

Contradictions are **not** hidden inside MSS or DSS.

Track them explicitly:

- `dob_conflict`
- `identifier_conflict`
- `entity_type_conflict`
- `geography_conflict`
- `context_conflict`

### 1.4 Decision thresholds

Baseline decision logic:

- `close`
  - MSS <= 30
  - no hard blocker requiring review

- `investigate_next_step`
  - MSS 31–69
  - or DSS < 50
  - or ambiguity/contradiction exists

- `escalate`
  - MSS >= 70
  - DSS >= 50
  - no hard contradiction
  - evidence must include more than name-only similarity

- `review_pending`
  - runtime or guardrail failure
  - required downstream evidence unavailable
  - worker cannot safely finalize a recommendation

### 1.5 Name-only cap

If MSS would exceed 69 based on name / alias / transliteration evidence **without** corroboration from date, identifier, geography, or context, cap MSS at **69**.

This prevents unsafe escalation from strong-name-only false positives.

---

## 2. Shared first-pass name scoring rules

These rules apply across all alert families, with different emphasis depending on alert type.

### 2.1 Name bucket rules

Only the **highest** applicable name bucket should be used.

| Name evidence type | Score range | Notes |
|---|---:|---|
| Exact normalized full-name match | 35 | Canonical exact equality after normalization |
| Exact alias-to-name match | 30 | Screened-party name exactly equals canonical alias |
| Strong transliteration-equivalent match | 24–28 | Cyrillic / Arabic / Persian / cross-script equivalent |
| Reordered or partial strong token match | 15–22 | Token sort, surname-first, dropped middle tokens |
| Weak fuzzy/token overlap | 5–12 | Supporting only, never enough for escalation |
| Noise only | 0–4 | Very weak overlap; should not drive decisions |

### 2.2 Date bucket rules

| Date evidence type | Score |
|---|---:|
| Exact full DOB/incorp date match | 20 |
| Exact month+year or strong approximate match | 12 |
| Year-only match | 8–10 |
| Missing | 0 |
| Contradiction | 0 and add contradiction flag |

### 2.3 Identifier bucket rules

| Identifier evidence type | Score |
|---|---:|
| Exact strong identifier match | 20 |
| Exact secondary identifier match | 10–15 |
| Partial but plausible support | 4–8 |
| Missing | 0 |
| Explicit mismatch | 0 and add contradiction flag |

### 2.4 Geography / address bucket rules

| Geography or address evidence type | Score |
|---|---:|
| Strong address / city / country alignment | 8–10 |
| Country-only or weak geography support | 3–7 |
| Missing | 0 |
| Contradiction | 0 and add contradiction flag |

### 2.5 Context / entity-type support

| Context support type | Score |
|---|---:|
| Strong context / entity type support | 5 |
| Neutral / unknown | 0–2 |
| Wrong-context suspicion | 0 |
| Clear mismatch | 0 and add contradiction flag |

---

## 3. DSS scoring framework

DSS is composed of five buckets.

| DSS bucket | Max points | What it measures |
|---|---:|---|
| Screened party completeness | 30 | How complete the alert subject is |
| Matched party completeness | 25 | How complete the list-side data is |
| Identifier quality | 20 | Whether identifiers are meaningful and reliable |
| Geography / address quality | 10 | Whether location fields are usable |
| Supporting context | 15 | Notes, attachments, vendor flags, payment context |

### 3.1 Screened party completeness (0–30)

| Field presence | Points |
|---|---:|
| Name present and usable | 8 |
| Aliases present | 4 |
| DOB/incorp date present | 6 |
| Country present | 4 |
| Address present | 4 |
| Identifier present | 4 |

### 3.2 Matched party completeness (0–25)

| Field presence | Points |
|---|---:|
| Name usable | 6 |
| Aliases present | 4 |
| DOB/incorp date present | 5 |
| Country/geography present | 4 |
| Identifier present | 6 |

### 3.3 Identifier quality (0–20)

| Identifier quality | Points |
|---|---:|
| Strong identifiers on both sides | 20 |
| Strong identifier on one side only | 10–14 |
| Weak/noisy identifier support | 3–8 |
| No useful identifiers | 0 |

### 3.4 Geography quality (0–10)

| Geography quality | Points |
|---|---:|
| Specific usable geography on both sides | 8–10 |
| Country-level only | 3–6 |
| No useful geography | 0 |

### 3.5 Supporting context (0–15)

| Supporting context | Points |
|---|---:|
| Strong notes/docs/context | 10–15 |
| Some useful vendor detail | 4–9 |
| Little to no useful support | 0–3 |

---

# 4. Per-field matrices by alert family

## 4.1 Individual onboarding alerts

### 4.1.1 MSS matrix — individual onboarding

**Name bucket max: 35**  
Take the highest applicable rule only.

| Field / bucket | Max | Rule |
|---|---:|---|
| Full legal name exact normalized match | 35 | Same canonical full name |
| Alias-to-name exact | 30 | Screened name equals OFAC alias |
| Strong transliteration-equivalent | 26 | Cross-script equivalent with stable token structure |
| Reordered / strong partial token match | 18 | Reordered first/last or missing middle name |
| Weak fuzzy overlap | 8 | Supporting only |

**Additional evidence buckets**

| Field / bucket | Max | Rule |
|---|---:|---|
| DOB exact | 20 | Full exact DOB |
| DOB year-only | 10 | Year-only support |
| Passport / national ID / tax ID exact | 20 | Strong identifier match |
| Country / nationality support | 8 | Same country or nationality |
| Address support | 8 | Strong city/address alignment |
| Entity-type support | 5 | Individual-to-individual |

**Onboarding MSS notes**
- DOB and identifier evidence are the strongest differentiators for individuals.
- A name-only or transliteration-only match must not escalate.
- Strong DOB contradiction is a blocker.

### 4.1.2 DSS matrix — individual onboarding

| Field / bucket | Max | Rule |
|---|---:|---|
| Screened name present | 8 | Usable legal name |
| Screened aliases present | 4 | Alias list available |
| Screened DOB present | 6 | Full or partial DOB |
| Screened country present | 4 | Country/nationality available |
| Screened address present | 4 | Address/city available |
| Screened identifier present | 4 | Passport, tax ID, national ID |

| Field / bucket | Max | Rule |
|---|---:|---|
| Matched name usable | 6 | OFAC name usable |
| Matched aliases present | 4 | Alias data present |
| Matched DOB present | 5 | Full or partial DOB |
| Matched geography present | 4 | Country/location usable |
| Matched identifiers present | 6 | Identifier fields present |

| Quality/support bucket | Max | Rule |
|---|---:|---|
| Identifier quality | 20 | Strong if exact identifiers exist |
| Geography quality | 10 | Strong if address/city/country are meaningful |
| Supporting context | 15 | Vendor notes, analyst notes, attachments |

**Onboarding DSS notes**
- Cases with missing DOB and missing identifiers should usually land in `investigate_next_step`.
- Exact identifiers without DOB can still produce strong DSS if the identifier is reliable.

---

## 4.2 Organization onboarding alerts

### 4.2.1 MSS matrix — organization onboarding

**Name bucket max: 35**  
Take the highest applicable rule only.

| Field / bucket | Max | Rule |
|---|---:|---|
| Exact normalized organization name | 35 | Canonical exact company name |
| Alias-to-name exact | 30 | OFAC alias equals customer/vendor legal name |
| Strong transliteration-equivalent | 24 | Cross-script business-name equivalent |
| Reordered / partial strong token match | 20 | Removed suffixes or reordered business tokens |
| Weak fuzzy overlap | 8 | Supporting only |

**Additional evidence buckets**

| Field / bucket | Max | Rule |
|---|---:|---|
| Incorporation / registration date exact | 15 | Use when available |
| Registration number / tax ID / company ID exact | 20 | Strong org identifier |
| Country / registration country support | 10 | Same country or registry country |
| Address support | 10 | Registered/business address alignment |
| Entity-type support | 5 | Organization-to-organization |

**Organization onboarding MSS notes**
- Organization names are noisier due to suffixes and generic tokens.
- Registration number and registration country matter more than DOB-style fields.
- Normalize suffixes such as `LTD`, `LLC`, `INC`, `CO`, `PJSC`, `OOO`, etc., before exact comparison.

### 4.2.2 DSS matrix — organization onboarding

| Field / bucket | Max | Rule |
|---|---:|---|
| Screened org name present | 8 | Legal entity name present |
| Screened aliases / trade names present | 4 | DBA / alias support |
| Screened registration/incorp date present | 5 | Useful but not mandatory |
| Screened country present | 4 | Registration or operating country |
| Screened address present | 5 | Registered or operating address |
| Screened organization identifier present | 4 | Tax ID, company ID, registration no. |

| Field / bucket | Max | Rule |
|---|---:|---|
| Matched org name usable | 6 | OFAC entity name usable |
| Matched aliases present | 4 | Alias set available |
| Matched registration/incorp date present | 4 | Useful if provided |
| Matched geography present | 5 | Country/location usable |
| Matched identifiers present | 6 | Identifier fields present |

| Quality/support bucket | Max | Rule |
|---|---:|---|
| Identifier quality | 20 | Strong if registry/tax/company IDs are usable |
| Geography quality | 10 | Strong if registered address or country aligns |
| Supporting context | 15 | Corporate structure notes, source docs, vendor notes |

**Organization DSS notes**
- Organization cases can remain high-value even with no date if identifiers and registration geography are strong.
- Generic business names without identifiers should remain low-DSS.

---

## 4.3 ACH party alerts

### 4.3.1 MSS matrix — ACH party alerts

**Name bucket max: 35**  
Take the highest applicable rule only.

| Field / bucket | Max | Rule |
|---|---:|---|
| Exact normalized party name | 35 | Exact originator/beneficiary/company name |
| Alias-to-name exact | 30 | OFAC alias equals party name |
| Strong transliteration-equivalent | 24 | Cross-script party equivalent |
| Reordered / partial strong token match | 18 | Strong token overlap |
| Weak fuzzy overlap | 8 | Supporting only |

**Additional evidence buckets**

| Field / bucket | Max | Rule |
|---|---:|---|
| DOB exact (individual party) | 15 | Often unavailable, so weight slightly lower |
| DOB year-only | 8 | Partial support |
| Identifier exact | 20 | Account-linked customer ID, tax ID, passport, etc. |
| Country / corridor support | 10 | Party country or transaction corridor supports match |
| Address support | 5 | Often sparse in ACH alerts |
| Transaction/context support | 10 | Party role, payment narrative, corridor, institution context |
| Entity-type support | 5 | Person-to-person or org-to-org alignment |

**ACH MSS notes**
- ACH alerts often have sparse party data; transaction context becomes more important.
- Context support can help, but should not rescue a weak name-only hit into escalation.
- If the ACH alert is only a name hit with sparse metadata, it should usually remain `investigate_next_step`.

### 4.3.2 DSS matrix — ACH party alerts

| Field / bucket | Max | Rule |
|---|---:|---|
| Screened party name present | 8 | Party name present |
| Screened aliases present | 2 | Usually limited |
| Screened DOB/incorp date present | 5 | Often missing |
| Screened country present | 5 | Useful if captured |
| Screened address present | 3 | Often limited |
| Screened identifier present | 7 | Customer ID, account-linked ID, tax ID, etc. |

| Field / bucket | Max | Rule |
|---|---:|---|
| Matched name usable | 6 | OFAC party name usable |
| Matched aliases present | 4 | Alias support present |
| Matched DOB/incorp date present | 4 | Useful if present |
| Matched geography present | 5 | Country/location support |
| Matched identifiers present | 6 | OFAC-side identifiers available |

| Quality/support bucket | Max | Rule |
|---|---:|---|
| Identifier quality | 20 | Higher importance in sparse ACH contexts |
| Geography quality | 10 | Country/corridor support |
| Supporting transaction context | 15 | Narrative, role, corridor, counterparty context |

**ACH DSS notes**
- A transaction-linked alert can have usable DSS even with weak address fields if transaction context and customer identifiers are strong.
- Missing DOB should not be fatal if role, identifiers, and corridor context are helpful.
- Narrative-only context should not be over-trusted.

---

## 4.4 Wrong-context cases

Wrong-context cases are special. They are defined by:
- entity-type mismatch
- role mismatch
- vessel vs individual/company confusion
- geography/place mistaken as party
- payment narrative or reference text creating a spurious match

These cases should be handled with a **context overlay matrix**, not by inflating ordinary name scores.

### 4.4.1 Wrong-context context overlay

| Wrong-context indicator | MSS effect | DSS effect | Decision effect |
|---|---:|---:|---|
| Individual matched to vessel name | set context support to 0 | no change | add `entity_type_conflict` |
| Organization matched to individual list entry | set context support to 0 | no change | add `entity_type_conflict` |
| Vessel matched to company/person | set context support to 0 | no change | add `context_conflict` |
| Geography/place token mistaken for party | reduce name bucket to max 8 | no change | bias toward `close` |
| Narrative/reference caused match, party fields weak | reduce name bucket by 10–20 | reduce supporting-context score to max 5 | bias toward `investigate_next_step` or `close` |
| Strong name overlap but explicit wrong entity type | keep raw name score for audit | no positive context score | escalation blocked unless exceptional corroboration exists |

### 4.4.2 Wrong-context decision rules

| Condition | Result |
|---|---|
| Entity-type conflict and no strong identifier/date corroboration | `close` |
| Entity-type conflict but sparse data prevents safe close | `investigate_next_step` |
| Wrong-context signal plus high name score only | cap MSS at 30–40 |
| Wrong-context signal plus strong identifiers/geography/date | `investigate_next_step`, not immediate `escalate` |
| Any vessel/person/company mismatch with no corroboration | `close` |

**Wrong-context notes**
- Wrong-context cases are where the model must stay humble.
- Strong name overlap must not outweigh a clear entity-type mismatch.
- These cases are one of the main reasons contradictions stay separate from MSS/DSS.

---

# 5. Family-specific threshold overlays

## 5.1 Individual onboarding
- `escalate`:
  - MSS >= 70
  - DSS >= 50
  - no `dob_conflict`
  - no `identifier_conflict`
  - evidence includes DOB or identifier or strong geography support

- `investigate_next_step`:
  - MSS 31–69
  - or DSS < 50
  - or transliteration-only strong match
  - or sparse onboarding/KYC data

## 5.2 Organization onboarding
- `escalate`:
  - MSS >= 70
  - DSS >= 50
  - no strong registration/identifier contradiction
  - evidence includes registration number, registration country, or strong address support

- `investigate_next_step`:
  - generic business-name overlap
  - missing company ID / registration number
  - weak geography or only broad name overlap

## 5.3 ACH party alerts
- `escalate`:
  - MSS >= 70
  - DSS >= 50
  - no hard contradiction
  - evidence includes more than name and narrative alone

- `investigate_next_step`:
  - sparse ACH party metadata
  - high name score but weak supporting fields
  - corridor/context support present but insufficient identifiers

## 5.4 Wrong-context cases
- default to `close` or `investigate_next_step`
- only allow `escalate` if:
  - MSS >= 80
  - DSS >= 60
  - and there is strong corroboration beyond the wrong-context signal
- in practice, these should almost never auto-escalate from first-pass scoring alone

---

# 6. Implementation guidance

## 6.1 Scoring outputs per case

Each scored case should produce:

- `name_match_score`
- `date_match_score`
- `identifier_match_score`
- `geography_match_score`
- `address_match_score`
- `context_support_score`
- `match_strength_score`
- `screened_party_completeness_score`
- `matched_party_completeness_score`
- `identifier_quality_score`
- `geography_quality_score`
- `supporting_context_score`
- `data_sufficiency_score`
- `contradictions[]`
- `decision_label`

## 6.2 Current baseline relationship to `claw-identity`

`claw-identity` already has a simple deterministic first-pass scorer:
- exact name
- partial name
- alias
- DOB / birth year
- country
- address
- identifier
with a review threshold of 60. This matrix keeps that spirit but formalizes it by alert family and adds DSS + contradiction handling.

## 6.3 Expected first implementation order

1. implement family-specific extractors
2. implement name bucket selection
3. implement MSS and DSS calculators
4. implement contradictions
5. implement threshold engine
6. add fixtures for:
   - individual onboarding
   - organization onboarding
   - ACH party
   - wrong-context

---

# 7. Drift checks

Future changes should be reviewed against these questions:

- Are we still keeping exact/alias/transliteration above weak fuzzy evidence?
- Are we still requiring more than name-only similarity to escalate?
- Are wrong-context cases still strongly suppressed?
- Are sparse-data cases still biased toward `investigate_next_step`?
- Are DSS and contradictions still separate from MSS?

If any answer becomes “no,” document the reason before changing the matrix.
