package history

import "time"

func RecencyWeight(reviewedAt time.Time, now time.Time) float64 {
    if reviewedAt.IsZero() || now.IsZero() {
        return 0
    }

    age := now.Sub(reviewedAt)
    switch {
    case age <= 30*24*time.Hour:
        return 1.0
    case age <= 90*24*time.Hour:
        return 0.7
    case age <= 180*24*time.Hour:
        return 0.4
    default:
        return 0.2
    }
}
