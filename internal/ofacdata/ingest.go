package ofacdata

import "context"

type Store interface {
	UpsertSnapshot(ctx context.Context, snap Snapshot) error
}

type Ingestor struct {
	store Store
}

func NewIngestor(store Store) *Ingestor {
	return &Ingestor{store: store}
}

func (i *Ingestor) IngestXML(ctx context.Context, raw []byte) error {
	// TODO:
	// 1. parse XML into XMLRoot
	// 2. normalize into Snapshot
	// 3. upsert into relational tables
	_ = raw
	_ = ctx
	return nil
}
