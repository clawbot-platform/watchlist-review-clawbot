package ofacgraph

import "strings"

type Neighbor struct {
    FromPartyID   string
    ToPartyID     string
    Relationship  string
    TargetName    string
    TargetProgram string
}

func scoreDirectNeighbors(neighbors []Neighbor) (int, []string) {
    if len(neighbors) == 0 {
        return 0, nil
    }
    score := 0
    var reasons []string
    for _, n := range neighbors {
        rel := strings.ToLower(strings.TrimSpace(n.Relationship))
        switch {
        case strings.Contains(rel, "owner"), strings.Contains(rel, "operator"), strings.Contains(rel, "controller"):
            score += 6
            reasons = append(reasons, "direct ownership/control relationship in OFAC neighborhood")
        case strings.Contains(rel, "director"), strings.Contains(rel, "manager"), strings.Contains(rel, "associate"):
            score += 4
            reasons = append(reasons, "direct managerial/associate relationship in OFAC neighborhood")
        default:
            score += 2
            reasons = append(reasons, "direct related-party linkage present in OFAC relationship graph")
        }
    }
    if score > 12 {
        score = 12
    }
    return score, dedupeStrings(reasons)
}
