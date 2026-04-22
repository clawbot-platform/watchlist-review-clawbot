package ofacdata

import (
	"context"
	"encoding/xml"
	"fmt"
)

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
	if i == nil || i.store == nil {
		return fmt.Errorf("ofac ingestor store is not configured")
	}
	var root XMLRoot
	if err := xml.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("parse OFAC XML: %w", err)
	}
	snapshot := NormalizeXML(root)
	if err := i.store.UpsertSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("persist OFAC snapshot: %w", err)
	}
	return nil
}
