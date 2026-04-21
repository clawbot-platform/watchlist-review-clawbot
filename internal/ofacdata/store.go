package ofacdata

import "context"

type SQLStore struct{}

func (s *SQLStore) UpsertSnapshot(ctx context.Context, snap Snapshot) error {
	// TODO: write normalized rows into Postgres
	_ = ctx
	_ = snap
	return nil
}
