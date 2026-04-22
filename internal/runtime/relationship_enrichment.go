package runtime

import (
    "context"

    "github.com/clawbot-platform/watchlist-review-clawbot/internal/ofacgraph"
    "github.com/clawbot-platform/watchlist-review-clawbot/internal/scoring"
)

type RelationshipBuilder interface {
    BuildRelationshipEvidence(ctx context.Context, in ofacgraph.BuildInput) (ofacgraph.RelationshipEvidence, error)
}

type RelationshipEnrichmentInput struct {
    MatchedListUID        string
    ScreenedEntityType    string
    ScreenedCountries     []string
    NormalizedIdentifiers map[string][]string
    Program               string
    Score                 *scoring.Result
    ReviewContext         map[string]any
}

type RelationshipEnricher struct {
    builder RelationshipBuilder
}

func NewRelationshipEnricher(builder RelationshipBuilder) *RelationshipEnricher {
    return &RelationshipEnricher{builder: builder}
}

func (e *RelationshipEnricher) Enrich(ctx context.Context, in RelationshipEnrichmentInput) (ofacgraph.RelationshipEvidence, error) {
    if e == nil || e.builder == nil {
        return ofacgraph.RelationshipEvidence{}, nil
    }
    ev, err := e.builder.BuildRelationshipEvidence(ctx, ofacgraph.BuildInput{
        MatchedListUID:        in.MatchedListUID,
        ScreenedEntityType:    in.ScreenedEntityType,
        ScreenedCountries:     in.ScreenedCountries,
        NormalizedIdentifiers: in.NormalizedIdentifiers,
        Program:               in.Program,
    })
    if err != nil {
        return ofacgraph.RelationshipEvidence{}, err
    }

    scoring.ApplyRelationshipEvidence(in.Score, ev)
    if in.ReviewContext != nil {
        in.ReviewContext["relationship_evidence"] = ofacgraph.BuildExplanation(ev)
    }
    return ev, nil
}
