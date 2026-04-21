package history

import "context"

type Service struct {
    store Store
}

func NewService(store Store) *Service {
    return &Service{store: store}
}

func (s *Service) Lookup(ctx context.Context, req LookupRequest) (*LookupResult, error) {
    if s == nil || s.store == nil {
        return &LookupResult{}, nil
    }
    return s.store.LookupSamePair(ctx, req)
}
