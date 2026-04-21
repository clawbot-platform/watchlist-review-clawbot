package ofacdata

func NormalizeXML(root XMLRoot) Snapshot {
	// TODO:
	// - map distinct parties to canonical party rows
	// - map identifiers/documents/locations
	// - map profile relationships into Relationship rows
	// - map sanctions entries and program context
	_ = root
	return Snapshot{}
}
