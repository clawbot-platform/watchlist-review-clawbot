package history

import "context"

type Store interface {
    LookupSamePair(ctx context.Context, req LookupRequest) (*LookupResult, error)
    SaveDisposition(ctx context.Context, disposition CaseDisposition) error
}
