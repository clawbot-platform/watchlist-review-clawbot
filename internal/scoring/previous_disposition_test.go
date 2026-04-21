package scoring

import (
    "testing"
    "time"

    "github.com/clawbot-platform/watchlist-review-clawbot/internal/history"
)

func TestApplyPreviousDispositionScore_SamePairEscalateBoost(t *testing.T) {
    now := time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)
    result := &Result{}

    ApplyPreviousDispositionScore(result, PreviousDispositionInput{
        LookupResult: &history.LookupResult{
            SamePairCases: []history.CaseDisposition{
                {
                    DecisionLabel: history.DispositionEscalate,
                    ReviewedAt:    now.Add(-7 * 24 * time.Hour),
                },
            },
            LatestSamePair: &history.CaseDisposition{
                DecisionLabel: history.DispositionEscalate,
                ReviewedAt:    now.Add(-7 * 24 * time.Hour),
            },
        },
        IdentityLinkConfidenceStrong: true,
        Now:                          now,
    })

    if result.PreviousDispositionScore <= 0 {
        t.Fatalf("PreviousDispositionScore = %d, want > 0", result.PreviousDispositionScore)
    }
}

func TestApplyPreviousDispositionScore_FalsePositivePenalty(t *testing.T) {
    now := time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)
    result := &Result{}

    ApplyPreviousDispositionScore(result, PreviousDispositionInput{
        LookupResult: &history.LookupResult{
            SamePairCases: []history.CaseDisposition{
                {
                    DecisionLabel:        history.DispositionClose,
                    ContradictionPattern: "country_mismatch+identifier_conflict",
                    ReviewedAt:           now.Add(-7 * 24 * time.Hour),
                },
            },
            LatestSamePair: &history.CaseDisposition{
                DecisionLabel:        history.DispositionClose,
                ContradictionPattern: "country_mismatch+identifier_conflict",
                ReviewedAt:           now.Add(-7 * 24 * time.Hour),
            },
        },
        CurrentContradictionPattern:  "country_mismatch+identifier_conflict",
        IdentityLinkConfidenceStrong: true,
        Now:                          now,
    })

    if result.PreviousDispositionScore >= 0 {
        t.Fatalf("PreviousDispositionScore = %d, want < 0", result.PreviousDispositionScore)
    }
}
