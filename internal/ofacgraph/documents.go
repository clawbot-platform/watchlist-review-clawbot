package ofacgraph

type LinkedDocument struct {
	PartyID        string
	DocumentType   string
	DocumentValue  string
	IssuingCountry string
}

func scoreLinkedDocuments(docs []LinkedDocument, normalizedIdentifiers map[string][]string) (int, []string) {
	// TODO: exact linked official docs should boost support materially
	_ = docs
	_ = normalizedIdentifiers
	return 0, nil
}
