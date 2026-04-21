package ofacgraph

type Neighbor struct {
	FromPartyID   string
	ToPartyID     string
	Relationship  string
	TargetName    string
	TargetProgram string
}

func scoreDirectNeighbors(neighbors []Neighbor) (int, []string) {
	// TODO: weight direct relationship support by relationship type and program coherence
	_ = neighbors
	return 0, nil
}
