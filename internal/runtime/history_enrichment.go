package runtime

import (
    "context"

    "github.com/clawbot-platform/watchlist-review-clawbot/internal/history"
)

type HistoryProvider interface {
    Lookup(ctx context.Context, req history.LookupRequest) (*history.LookupResult, error)
}

// Example integration flow:
//
// 1. Build screened fingerprint from the current alert/features.
// 2. Lookup same-pair reviewed history.
// 3. Pass lookup result into scoring.ApplyPreviousDispositionScore.
// 4. Attach summary/evidence into review_context for downstream note generation.
