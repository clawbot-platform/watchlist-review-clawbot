# Hardening Notes for Deterministic Slice

This pack adds:

- one additional individual false-positive fixture driven by common-name collision
- one organization false-positive fixture driven by business-name overlap
- one wrong-context fixture driven by vessel-vs-individual mismatch
- one sparse-data fixture driven by name-only overlap
- API-level response assertions for labels and score bands

## Expected deterministic behavior

### false_positive_common_name.json
- label: `investigate_next_step`
- reason: exact name but conflicting date/geography and no strong identifier corroboration

### false_positive_org_name_overlap.json
- label: `investigate_next_step`
- reason: overlapping business tokens but weak corroboration

### wrong_context_vessel_vs_individual.json
- label: `investigate_next_step`
- reason: entity-type conflict prevents escalation even with strong name overlap

### sparse_data_name_only.json
- label: `investigate_next_step`
- reason: name-only hit with very low data sufficiency

These cases intentionally bias toward conservative handling. No new fixture in this pack should auto-escalate.
