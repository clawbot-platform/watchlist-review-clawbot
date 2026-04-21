package ofacgraph

import "context"

type QueryService interface {
	BuildRelationshipEvidence(ctx context.Context, in BuildInput) (RelationshipEvidence, error)
}

type BuildInput struct {
	MatchedListUID           string
	ScreenedEntityType       string
	ScreenedCountries        []string
	NormalizedIdentifiers    map[string][]string
	Program                  string
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) BuildRelationshipEvidence(ctx context.Context, in BuildInput) (RelationshipEvidence, error) {
	// TODO:
	// - resolve matched OFAC party/profile from matched list UID
	// - pull direct neighbors and linked documents
	// - compute support/conflict explanations
	_ = ctx
	_ = in
	return RelationshipEvidence{}, nil
}
