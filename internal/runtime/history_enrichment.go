package runtime

import (
    "context"
    "time"

    "github.com/clawbot-platform/watchlist-review-clawbot/internal/history"
    "github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

type HistoryProvider interface {
    Lookup(ctx context.Context, req history.LookupRequest) (*history.LookupResult, error)
}

type HistoryEnrichmentInput struct {
    TenantID                  string
    ScreenedEntityFingerprint string
    MatchedListUID            string
    MatchedProgram            string
    CurrentContradiction      string
    IdentityLinkConfidenceStrong bool
    Score                     *scoring.Result
    ReviewContext             map[string]any
    Now                       time.Time
}

type HistoryEnricher struct {
    provider HistoryProvider
}

func NewHistoryEnricher(provider HistoryProvider) *HistoryEnricher {
    return &HistoryEnricher{provider: provider}
}

func (e *HistoryEnricher) Enrich(ctx context.Context, in HistoryEnrichmentInput) (*history.LookupResult, error) {
    if e == nil || e.provider == nil {
        return &history.LookupResult{}, nil
    }
    lookup, err := e.provider.Lookup(ctx, history.LookupRequest{
        TenantID:                  in.TenantID,
        ScreenedEntityFingerprint: in.ScreenedEntityFingerprint,
        MatchedListUID:            in.MatchedListUID,
        MatchedProgram:            in.MatchedProgram,
    })
    if err != nil {
        return nil, err
    }

    scoring.ApplyPreviousDispositionScore(in.Score, scoring.PreviousDispositionInput{
        LookupResult:                 lookup,
        CurrentContradictionPattern:  in.CurrentContradiction,
        IdentityLinkConfidenceStrong: in.IdentityLinkConfidenceStrong,
        Now:                          in.Now,
    })

    if in.ReviewContext != nil {
        in.ReviewContext["history_evidence"] = map[string]any{
            "same_pair_cases":                 lookup.SamePairCases,
            "same_matched_profile_cases":      lookup.SameMatchedProfileCases,
            "latest_same_pair":                lookup.LatestSamePair,
            "escalate_count":                  lookup.EscalateCount,
            "investigate_count":               lookup.InvestigateCount,
            "close_count":                     lookup.CloseCount,
            "review_pending_count":            lookup.ReviewPendingCount,
            "most_recent_reviewed_at":         lookup.MostRecentReviewedAt,
            "recurring_contradiction_keys":    lookup.RecurringContradictionKeys,
        }
    }

    return lookup, nil
}
