# Disposition Memory Scoring v1 Eval Additions

Add at least 4–6 cases:

1. repeated confirmed true match
2. repeated false positive with same contradiction pattern
3. repeated unresolved same pair
4. history should not apply because pair differs
5. history should not apply because tenant differs
6. stale history with weak decay

Suggested naming:
- `same_pair_prior_escalate_recent.json`
- `same_pair_prior_false_positive_same_pattern.json`
- `same_pair_prior_investigate_recent.json`
- `different_pair_no_history_application.json`
