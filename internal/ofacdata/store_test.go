package ofacdata

import "testing"

func TestNormalizeXML_Starter(t *testing.T) {
	root := XMLRoot{}
	snap := NormalizeXML(root)
	if len(snap.Parties) != 0 {
		t.Fatalf("expected empty snapshot for empty input")
	}
}
