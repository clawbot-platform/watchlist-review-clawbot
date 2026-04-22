package ofacgraph

import (
    "context"
    "fmt"
    "sort"
    "strings"
)

type QueryService interface {
    BuildRelationshipEvidence(ctx context.Context, in BuildInput) (RelationshipEvidence, error)
}

type Store interface {
    ResolveMatchedParty(ctx context.Context, listUID string) (MatchedPartyRecord, error)
    LoadNeighbors(ctx context.Context, partyID string) ([]Neighbor, error)
    LoadLinkedDocuments(ctx context.Context, partyID string) ([]LinkedDocument, error)
    LoadProgramContext(ctx context.Context, partyID string) ([]string, error)
    LoadCountries(ctx context.Context, partyID string) ([]string, error)
}

type BuildInput struct {
    MatchedListUID        string
    ScreenedEntityType    string
    ScreenedCountries     []string
    NormalizedIdentifiers map[string][]string
    Program               string
}

type MatchedPartyRecord struct {
    PartyID    string
    ListUID    string
    PartyType  string
    PrimaryName string
}

type Service struct {
    store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) BuildRelationshipEvidence(ctx context.Context, in BuildInput) (RelationshipEvidence, error) {
    if s == nil || s.store == nil {
        return RelationshipEvidence{}, fmt.Errorf("ofacgraph store is not configured")
    }
    matched, err := s.store.ResolveMatchedParty(ctx, in.MatchedListUID)
    if err != nil {
        return RelationshipEvidence{}, err
    }
    if matched.PartyID == "" {
        return RelationshipEvidence{}, nil
    }

    neighbors, err := s.store.LoadNeighbors(ctx, matched.PartyID)
    if err != nil {
        return RelationshipEvidence{}, err
    }
    docs, err := s.store.LoadLinkedDocuments(ctx, matched.PartyID)
    if err != nil {
        return RelationshipEvidence{}, err
    }
    programs, err := s.store.LoadProgramContext(ctx, matched.PartyID)
    if err != nil {
        return RelationshipEvidence{}, err
    }
    countries, err := s.store.LoadCountries(ctx, matched.PartyID)
    if err != nil {
        return RelationshipEvidence{}, err
    }

    supportScore, supportReasons := scoreDirectNeighbors(neighbors)
    docScore, docReasons := scoreLinkedDocuments(docs, in.NormalizedIdentifiers)
    programScore, programReasons := scoreProgramContext(programs, in.Program)
    conflictPenalty, conflictReasons := scoreCountryConflict(countries, in.ScreenedCountries)

    linkedDocs := collectLinkedDocumentStrings(docs)
    ev := RelationshipEvidence{
        MatchedPartyID:              matched.PartyID,
        DirectRelationshipSupport:   supportScore > 0,
        RelationshipSupportScore:    supportScore,
        RelationshipConflictPenalty: conflictPenalty,
        OfficialDocLinkScore:        docScore,
        ProgramContextScore:         programScore,
        NeighborCount:               len(neighbors),
        Reasons:                     append(append(supportReasons, docReasons...), programReasons...),
        Conflicts:                   conflictReasons,
        LinkedDocuments:             linkedDocs,
        ProgramContext:              dedupeStrings(programs),
    }
    return ev, nil
}

func scoreProgramContext(programs []string, expected string) (int, []string) {
    expected = strings.TrimSpace(strings.ToLower(expected))
    if expected == "" {
        return 0, nil
    }
    for _, p := range programs {
        if strings.TrimSpace(strings.ToLower(p)) == expected {
            return 4, []string{"program context aligns with matched sanctions program"}
        }
    }
    return 0, nil
}

func scoreCountryConflict(matchedCountries, screenedCountries []string) (int, []string) {
    matchedSet := map[string]struct{}{}
    for _, c := range matchedCountries {
        c = strings.ToUpper(strings.TrimSpace(c))
        if c != "" { matchedSet[c] = struct{}{} }
    }
    if len(matchedSet) == 0 {
        return 0, nil
    }
    for _, c := range screenedCountries {
        c = strings.ToUpper(strings.TrimSpace(c))
        if c == "" { continue }
        if _, ok := matchedSet[c]; ok {
            return 0, nil
        }
    }
    return 6, []string{"relationship neighborhood countries conflict with screened party context"}
}

func collectLinkedDocumentStrings(docs []LinkedDocument) []string {
    var out []string
    for _, d := range docs {
        val := strings.TrimSpace(strings.Join([]string{d.DocumentType, d.DocumentValue}, ": "))
        if val != ":" && val != "" {
            out = append(out, val)
        }
    }
    sort.Strings(out)
    return dedupeStrings(out)
}

func dedupeStrings(in []string) []string {
    seen := map[string]struct{}{}
    var out []string
    for _, s := range in {
        key := strings.ToLower(strings.TrimSpace(s))
        if key == "" {
            continue
        }
        if _, ok := seen[key]; ok {
            continue
        }
        seen[key] = struct{}{}
        out = append(out, strings.TrimSpace(s))
    }
    return out
}
