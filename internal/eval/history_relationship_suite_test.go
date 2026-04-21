package eval

import "testing"

func TestLoadHistoryRelationshipSpec_ParsesCases(t *testing.T) {
    spec, err := LoadBatchSpec("../../eval/specs/deterministic-history-relationship-v1.json")
    if err != nil {
        t.Fatalf("LoadBatchSpec() error = %v", err)
    }
    if len(spec.Cases) != 4 {
        t.Fatalf("len(spec.Cases) = %d, want 4", len(spec.Cases))
    }
}
