package scoring

import (
    "fmt"
    "strings"
    "time"

    "github.com/clawbot-platform/watchlist-review-clawbot/internal/history"
)

type PreviousDispositionInput struct {
    LookupResult                 *history.LookupResult
    CurrentContradictionPattern  string
    IdentityLinkConfidenceStrong bool
    Now                          time.Time
}

func ApplyPreviousDispositionScore(result *Result, input PreviousDispositionInput) {
    if result == nil || input.LookupResult == nil || !input.IdentityLinkConfidenceStrong {
        return
    }

    evidence := computePreviousDispositionEvidence(input)
    if !evidence.Applied {
        return
    }

    result.PreviousDispositionScore = evidence.PreviousDispositionScore
    result.PreviousDispositionReasons = append(result.PreviousDispositionReasons, evidence.PreviousDispositionReasons...)
    result.MatchStrengthScore += evidence.PreviousDispositionScore

    if evidence.PreviousDispositionScore > 0 {
        result.EvidenceFor = append(result.EvidenceFor, "prior reviewed disposition support")
    } else if evidence.PreviousDispositionScore < 0 {
        result.EvidenceAgainst = append(result.EvidenceAgainst, "prior reviewed false-positive history")
    }
}

func computePreviousDispositionEvidence(input PreviousDispositionInput) history.Evidence {
    if input.LookupResult == nil || len(input.LookupResult.SamePairCases) == 0 {
        return history.Evidence{}
    }

    now := input.Now
    if now.IsZero() {
        now = time.Now().UTC()
    }

    evidence := history.Evidence{
        Applied:              true,
        SamePairHistoryFound: true,
    }

    latest := input.LookupResult.LatestSamePair
    if latest == nil {
        return evidence
    }

    weight := history.RecencyWeight(latest.ReviewedAt, now)
    switch latest.DecisionLabel {
    case history.DispositionEscalate:
        evidence.PreviousDispositionScore += int(12 * weight)
        evidence.PreviousDispositionReasons = append(evidence.PreviousDispositionReasons,
            fmt.Sprintf("same pair previously escalated on %s", latest.ReviewedAt.Format(time.RFC3339)))
    case history.DispositionClose:
        if sameContradictionPattern(input.CurrentContradictionPattern, latest.ContradictionPattern) {
            evidence.PreviousDispositionScore -= int(12 * weight)
            evidence.PreviousDispositionReasons = append(evidence.PreviousDispositionReasons,
                "same pair previously closed as false positive with matching contradiction pattern")
        } else {
            evidence.PreviousDispositionScore -= int(6 * weight)
            evidence.PreviousDispositionReasons = append(evidence.PreviousDispositionReasons,
                "same pair previously closed as false positive")
        }
    case history.DispositionInvestigateNextStep:
        evidence.PreviousDispositionScore += int(2 * weight)
        evidence.PreviousDispositionReasons = append(evidence.PreviousDispositionReasons,
            "same pair previously required investigation")
    }

    if evidence.PreviousDispositionScore == 0 {
        evidence.Applied = false
    }
    return evidence
}

func sameContradictionPattern(a string, b string) bool {
    a = strings.TrimSpace(strings.ToLower(a))
    b = strings.TrimSpace(strings.ToLower(b))
    return a != "" && a == b
}
