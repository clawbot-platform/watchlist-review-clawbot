package history

import (
    "context"
    "testing"
    "time"
)

type stubStore struct {
    result *LookupResult
}

func (s *stubStore) LookupSamePair(_ context.Context, _ LookupRequest) (*LookupResult, error) {
    if s.result == nil {
        return &LookupResult{}, nil
    }
    return s.result, nil
}

func (s *stubStore) SaveDisposition(_ context.Context, _ CaseDisposition) error {
    return nil
}

func TestRecencyWeight(t *testing.T) {
    now := time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)

    got := RecencyWeight(now.Add(-10*24*time.Hour), now)
    if got != 1.0 {
        t.Fatalf("RecencyWeight(10d) = %v, want 1.0", got)
    }

    got = RecencyWeight(now.Add(-120*24*time.Hour), now)
    if got != 0.4 {
        t.Fatalf("RecencyWeight(120d) = %v, want 0.4", got)
    }
}

func TestServiceLookup(t *testing.T) {
    svc := NewService(&stubStore{
        result: &LookupResult{EscalateCount: 2},
    })

    got, err := svc.Lookup(context.Background(), LookupRequest{})
    if err != nil {
        t.Fatalf("Lookup() error = %v", err)
    }
    if got.EscalateCount != 2 {
        t.Fatalf("EscalateCount = %d, want 2", got.EscalateCount)
    }
}
