package store

import (
	"fmt"
	"testing"
)

// TestWithTx_RollbackUndoesWrite proves read-your-writes inside the tx and full
// rollback on error: a write is visible to a later read within the same closure,
// but vanishes once the closure returns an error.
func TestWithTx_RollbackUndoesWrite(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)
	before, err := s.GetEntity("auth-handler")
	if err != nil {
		t.Fatal(err)
	}

	err = s.WithTx(func(ts *Store) error {
		e, err := ts.GetEntity("auth-handler")
		if err != nil {
			return err
		}
		e.Title = "MUTATED"
		if err := ts.UpdateEntity(e); err != nil {
			return err
		}
		mid, err := ts.GetEntity("auth-handler") // read-your-writes within the tx
		if err != nil {
			return err
		}
		if mid.Title != "MUTATED" {
			t.Errorf("within tx: read should see the uncommitted write, got %q", mid.Title)
		}
		return fmt.Errorf("boom")
	})
	if err == nil {
		t.Fatal("WithTx must surface the closure error")
	}
	after, err := s.GetEntity("auth-handler")
	if err != nil {
		t.Fatal(err)
	}
	if after.Title != before.Title {
		t.Fatalf("rollback failed: title = %q, want %q", after.Title, before.Title)
	}
}

// TestWithTx_NestedSetCodeMapCommitsWithOuter proves a nested WithTx (SetCodeMap
// opens one internally) enlists in the outer tx instead of deadlocking on the
// single pooled connection, and commits with it.
func TestWithTx_NestedSetCodeMapCommitsWithOuter(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.WithTx(func(ts *Store) error {
		return ts.SetCodeMap("auth-handler", []string{"src/auth/**"})
	})
	if err != nil {
		t.Fatalf("nested SetCodeMap inside WithTx must not deadlock or error: %v", err)
	}
	patterns, err := s.CodeMapFor("auth-handler")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) != 1 || patterns[0] != "src/auth/**" {
		t.Fatalf("committed codemap = %v, want [src/auth/**]", patterns)
	}
}

// TestWithTx_NestedSetCodeMapRollsBackWithOuter is the integrity proof behind
// "patches + codemap are all-or-nothing": a SetCodeMap nested inside an outer tx
// that later fails must roll back. If SetCodeMap committed its own transaction
// (the pre-seam behavior), the globs would survive the outer rollback.
func TestWithTx_NestedSetCodeMapRollsBackWithOuter(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.WithTx(func(ts *Store) error {
		if err := ts.SetCodeMap("auth-handler", []string{"src/auth/**"}); err != nil {
			return err
		}
		return fmt.Errorf("boom after codemap write")
	})
	if err == nil {
		t.Fatal("WithTx must surface the closure error")
	}
	patterns, err := s.CodeMapFor("auth-handler")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) != 0 {
		t.Fatalf("nested SetCodeMap must roll back with the outer tx, got %v", patterns)
	}
}

// TestWithTx_MixedNodeAndCodeMapAtomic proves the cross-table guarantee the carrier
// relies on: an entity/seal write and a codemap write in one closure either both
// land or both vanish.
func TestWithTx_MixedNodeAndCodeMapAtomic(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)
	before, _ := s.GetEntity("auth-handler")

	err := s.WithTx(func(ts *Store) error {
		e, _ := ts.GetEntity("auth-handler")
		e.RootMerkle = "deadbeef"
		if err := ts.UpdateEntity(e); err != nil {
			return err
		}
		if err := ts.SetCodeMap("auth-handler", []string{"src/auth/**"}); err != nil {
			return err
		}
		return fmt.Errorf("boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	after, _ := s.GetEntity("auth-handler")
	if after.RootMerkle != before.RootMerkle {
		t.Errorf("seal write must roll back: merkle = %q, want %q", after.RootMerkle, before.RootMerkle)
	}
	patterns, _ := s.CodeMapFor("auth-handler")
	if len(patterns) != 0 {
		t.Errorf("codemap write must roll back together with the seal write, got %v", patterns)
	}
}
