package runtime

import "context"

type RelationshipBuilder interface {
	BuildRelationshipEvidence(ctx context.Context, in any) (any, error)
}

// TODO:
// - resolve build input from alert/features/identity evidence
// - fetch relationship evidence
// - apply scoring hook
// - attach graph explanation fields to review_context
