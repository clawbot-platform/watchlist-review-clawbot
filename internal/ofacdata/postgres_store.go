package ofacdata

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) UpsertSnapshot(ctx context.Context, snap Snapshot) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("ofac sql store is not configured")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, p := range snap.Parties {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO ofac_parties (
                party_id, list_uid, primary_name, party_type, distinct_party_id, created_at, updated_at
            ) VALUES ($1,$2,$3,$4,$5,$6,$7)
            ON CONFLICT (party_id) DO UPDATE SET
                list_uid = EXCLUDED.list_uid,
                primary_name = EXCLUDED.primary_name,
                party_type = EXCLUDED.party_type,
                distinct_party_id = EXCLUDED.distinct_party_id,
                updated_at = EXCLUDED.updated_at`,
			p.PartyID, p.ListUID, p.PrimaryName, p.PartyType, p.DistinctPartyID, p.CreatedAt, p.UpdatedAt,
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM ofac_party_aliases WHERE party_id = $1`, p.PartyID); err != nil {
			return err
		}
		for _, alias := range p.Aliases {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO ofac_party_aliases (party_id, alias) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
				p.PartyID, alias,
			); err != nil {
				return err
			}
		}
	}

	if err := replaceByParty(ctx, tx, `DELETE FROM ofac_identifiers WHERE party_id = $1`, collectPartyIDsFromIdentifiers(snap.Identifiers)); err != nil {
		return err
	}
	for _, x := range snap.Identifiers {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO ofac_identifiers (party_id, identifier_type, identifier_value, issuing_country)
            VALUES ($1,$2,$3,$4)
            ON CONFLICT (party_id, identifier_type, identifier_value) DO UPDATE SET
                issuing_country = EXCLUDED.issuing_country`,
			x.PartyID, x.IdentifierType, x.IdentifierValue, x.IssuingCountry,
		); err != nil {
			return err
		}
	}

	if err := replaceByParty(ctx, tx, `DELETE FROM ofac_documents WHERE party_id = $1`, collectPartyIDsFromDocuments(snap.Documents)); err != nil {
		return err
	}
	for _, x := range snap.Documents {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO ofac_documents (party_id, document_type, document_value, country)
            VALUES ($1,$2,$3,$4)
            ON CONFLICT (party_id, document_type, document_value) DO UPDATE SET
                country = EXCLUDED.country`,
			x.PartyID, x.DocumentType, x.DocumentValue, x.Country,
		); err != nil {
			return err
		}
	}

	if err := replaceByParty(ctx, tx, `DELETE FROM ofac_locations WHERE party_id = $1`, collectPartyIDsFromLocations(snap.Locations)); err != nil {
		return err
	}
	for _, x := range snap.Locations {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO ofac_locations (party_id, country_code, city, address_text, province_or_st)
            VALUES ($1,$2,$3,$4,$5)
            ON CONFLICT (party_id, country_code, city, address_text) DO UPDATE SET
                province_or_st = EXCLUDED.province_or_st`,
			x.PartyID, x.CountryCode, x.City, x.AddressText, x.ProvinceOrSt,
		); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ofac_relationships`); err != nil {
		return err
	}
	for _, x := range snap.Relationships {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO ofac_relationships (relationship_id, from_party_id, to_party_id, relationship, source_list_uid)
            VALUES ($1,$2,$3,$4,$5)
            ON CONFLICT (relationship_id) DO UPDATE SET
                relationship = EXCLUDED.relationship,
                source_list_uid = EXCLUDED.source_list_uid`,
			x.RelationshipID, x.FromPartyID, x.ToPartyID, x.Relationship, x.SourceListUID,
		); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ofac_sanctions_entries`); err != nil {
		return err
	}
	for _, x := range snap.SanctionsEntrys {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO ofac_sanctions_entries (entry_id, party_id, program, list_source, published_at)
            VALUES ($1,$2,$3,$4,$5)
            ON CONFLICT (entry_id) DO UPDATE SET
                program = EXCLUDED.program,
                list_source = EXCLUDED.list_source,
                published_at = EXCLUDED.published_at`,
			x.EntryID, x.PartyID, x.Program, x.ListSource, x.PublishedAt,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func replaceByParty(ctx context.Context, tx *sql.Tx, query string, partyIDs []string) error {
	seen := map[string]struct{}{}
	for _, partyID := range partyIDs {
		if partyID == "" {
			continue
		}
		if _, ok := seen[partyID]; ok {
			continue
		}
		seen[partyID] = struct{}{}
		if _, err := tx.ExecContext(ctx, query, partyID); err != nil {
			return err
		}
	}
	return nil
}

func collectPartyIDsFromIdentifiers(items []Identifier) []string {
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, x.PartyID)
	}
	return out
}
func collectPartyIDsFromDocuments(items []Document) []string {
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, x.PartyID)
	}
	return out
}
func collectPartyIDsFromLocations(items []Location) []string {
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, x.PartyID)
	}
	return out
}
