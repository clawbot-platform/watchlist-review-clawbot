package ofacgraph

import (
    "context"
    "database/sql"
)

type PostgresStore struct {
    db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
    return &PostgresStore{db: db}
}

func (s *PostgresStore) ResolveMatchedParty(ctx context.Context, listUID string) (MatchedPartyRecord, error) {
    var rec MatchedPartyRecord
    if s == nil || s.db == nil {
        return rec, nil
    }
    err := s.db.QueryRowContext(ctx, `
        SELECT party_id, list_uid, party_type, primary_name
        FROM ofac_parties
        WHERE list_uid = $1
        LIMIT 1`, listUID,
    ).Scan(&rec.PartyID, &rec.ListUID, &rec.PartyType, &rec.PrimaryName)
    if err == sql.ErrNoRows {
        return MatchedPartyRecord{}, nil
    }
    return rec, err
}

func (s *PostgresStore) LoadNeighbors(ctx context.Context, partyID string) ([]Neighbor, error) {
    if s == nil || s.db == nil {
        return nil, nil
    }
    rows, err := s.db.QueryContext(ctx, `
        SELECT r.from_party_id, r.to_party_id, r.relationship,
               p.primary_name,
               COALESCE(se.program, '')
        FROM ofac_relationships r
        JOIN ofac_parties p ON p.party_id = r.to_party_id
        LEFT JOIN ofac_sanctions_entries se ON se.party_id = r.to_party_id
        WHERE r.from_party_id = $1`, partyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Neighbor
    for rows.Next() {
        var n Neighbor
        if err := rows.Scan(&n.FromPartyID, &n.ToPartyID, &n.Relationship, &n.TargetName, &n.TargetProgram); err != nil {
            return nil, err
        }
        out = append(out, n)
    }
    return out, rows.Err()
}

func (s *PostgresStore) LoadLinkedDocuments(ctx context.Context, partyID string) ([]LinkedDocument, error) {
    if s == nil || s.db == nil {
        return nil, nil
    }
    rows, err := s.db.QueryContext(ctx, `
        SELECT party_id, document_type, document_value, country
        FROM ofac_documents
        WHERE party_id = $1`, partyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []LinkedDocument
    for rows.Next() {
        var d LinkedDocument
        if err := rows.Scan(&d.PartyID, &d.DocumentType, &d.DocumentValue, &d.IssuingCountry); err != nil {
            return nil, err
        }
        out = append(out, d)
    }
    return out, rows.Err()
}

func (s *PostgresStore) LoadProgramContext(ctx context.Context, partyID string) ([]string, error) {
    if s == nil || s.db == nil {
        return nil, nil
    }
    rows, err := s.db.QueryContext(ctx, `SELECT program FROM ofac_sanctions_entries WHERE party_id = $1`, partyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []string
    for rows.Next() {
        var v string
        if err := rows.Scan(&v); err != nil {
            return nil, err
        }
        out = append(out, v)
    }
    return out, rows.Err()
}

func (s *PostgresStore) LoadCountries(ctx context.Context, partyID string) ([]string, error) {
    if s == nil || s.db == nil {
        return nil, nil
    }
    rows, err := s.db.QueryContext(ctx, `SELECT country_code FROM ofac_locations WHERE party_id = $1`, partyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []string
    for rows.Next() {
        var v string
        if err := rows.Scan(&v); err != nil {
            return nil, err
        }
        out = append(out, v)
    }
    return out, rows.Err()
}
