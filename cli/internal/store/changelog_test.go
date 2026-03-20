package store

import "testing"

func TestChangelog_LogAndDiff(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// seedFixture inserts 7 entities + 3 relationships = 10 add + 3 add_rel entries.
	changes, err := s.UnmarkedChanges()
	if err != nil {
		t.Fatalf("unmarked: %v", err)
	}
	if len(changes) != 10 {
		t.Errorf("expected 10 unmarked changes after seed, got %d", len(changes))
	}

	// Verify the first entry is for acme-platform.
	if changes[0].EntityID != "acme-platform" {
		t.Errorf("first change entity = %q, want acme-platform", changes[0].EntityID)
	}
	if changes[0].Action != "add" {
		t.Errorf("first change action = %q, want add", changes[0].Action)
	}
}

func TestChangelog_Mark(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	if err := s.MarkChangelog("abc123"); err != nil {
		t.Fatalf("mark: %v", err)
	}

	changes, _ := s.UnmarkedChanges()
	if len(changes) != 0 {
		t.Errorf("expected 0 unmarked after mark, got %d", len(changes))
	}

	// New changes after marking should be unmarked.
	e, _ := s.GetEntity("api-gateway")
	e.Title = "Changed"
	s.UpdateEntity(e)

	changes, _ = s.UnmarkedChanges()
	if len(changes) != 1 {
		t.Errorf("expected 1 unmarked after new update, got %d", len(changes))
	}
}
