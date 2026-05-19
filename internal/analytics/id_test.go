package analytics

import "testing"

func TestNewEventID(t *testing.T) {
	id, err := NewEventID()
	if err != nil {
		t.Fatalf("NewEventID: %v", err)
	}
	if len(id) != 36 {
		t.Fatalf("unexpected id length: %q", id)
	}
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		t.Fatalf("unexpected uuid format: %q", id)
	}
}
