CREATE TABLE IF NOT EXISTS review_case_dispositions (
    tenant_id TEXT NOT NULL,
    case_id TEXT NOT NULL,
    alert_id TEXT NOT NULL,
    screened_entity_fingerprint TEXT NOT NULL,
    matched_list_uid TEXT NOT NULL,
    matched_program TEXT NOT NULL DEFAULT '',
    decision_label TEXT NOT NULL,
    contradiction_pattern TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, case_id)
);

CREATE INDEX IF NOT EXISTS idx_review_case_dispositions_same_pair
    ON review_case_dispositions (tenant_id, screened_entity_fingerprint, matched_list_uid, reviewed_at DESC);

CREATE INDEX IF NOT EXISTS idx_review_case_dispositions_profile
    ON review_case_dispositions (tenant_id, matched_list_uid, reviewed_at DESC);

CREATE TABLE IF NOT EXISTS ofac_parties (
    party_id TEXT PRIMARY KEY,
    list_uid TEXT NOT NULL,
    primary_name TEXT NOT NULL,
    party_type TEXT NOT NULL,
    distinct_party_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ofac_parties_list_uid ON ofac_parties (list_uid);
CREATE INDEX IF NOT EXISTS idx_ofac_parties_distinct_party_id ON ofac_parties (distinct_party_id);

CREATE TABLE IF NOT EXISTS ofac_party_aliases (
    party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    alias TEXT NOT NULL,
    PRIMARY KEY (party_id, alias)
);

CREATE TABLE IF NOT EXISTS ofac_identifiers (
    party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    identifier_type TEXT NOT NULL,
    identifier_value TEXT NOT NULL,
    issuing_country TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (party_id, identifier_type, identifier_value)
);

CREATE INDEX IF NOT EXISTS idx_ofac_identifiers_lookup
    ON ofac_identifiers (identifier_type, identifier_value);

CREATE TABLE IF NOT EXISTS ofac_documents (
    party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    document_type TEXT NOT NULL,
    document_value TEXT NOT NULL,
    country TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (party_id, document_type, document_value)
);

CREATE INDEX IF NOT EXISTS idx_ofac_documents_lookup
    ON ofac_documents (document_type, document_value);

CREATE TABLE IF NOT EXISTS ofac_locations (
    party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    country_code TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    address_text TEXT NOT NULL DEFAULT '',
    province_or_st TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (party_id, country_code, city, address_text)
);

CREATE INDEX IF NOT EXISTS idx_ofac_locations_country ON ofac_locations (country_code);

CREATE TABLE IF NOT EXISTS ofac_relationships (
    relationship_id TEXT PRIMARY KEY,
    from_party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    to_party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    relationship TEXT NOT NULL,
    source_list_uid TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_ofac_relationships_from_party ON ofac_relationships (from_party_id);
CREATE INDEX IF NOT EXISTS idx_ofac_relationships_to_party ON ofac_relationships (to_party_id);

CREATE TABLE IF NOT EXISTS ofac_sanctions_entries (
    entry_id TEXT PRIMARY KEY,
    party_id TEXT NOT NULL REFERENCES ofac_parties(party_id) ON DELETE CASCADE,
    program TEXT NOT NULL,
    list_source TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ofac_sanctions_entries_party ON ofac_sanctions_entries (party_id);
CREATE INDEX IF NOT EXISTS idx_ofac_sanctions_entries_program ON ofac_sanctions_entries (program);
