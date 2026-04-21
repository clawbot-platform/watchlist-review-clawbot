package identityevidence

import "context"

type Request struct {
	TenantID            string
	CaseID              string
	AlertID             string
	ScreenedFingerprint string
	MatchedListUID      string
	Program             string
}

type Response struct {
	MatchedProfileID string
	ConfidenceBand   string

	PreviousDispositionScore    int
	PreviousDispositionReasons  []string

	RelationshipSupportScore    int
	RelationshipConflictPenalty int
	OfficialDocLinkScore        int
	ProgramContextScore         int
	RelationshipReasons         []string

	Warnings []string
}

type Client interface {
	BuildEvidence(context.Context, Request) (*Response, error)
}
