package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRetireGate_RefusesOrphaningChild(t *testing.T) {
	s, _ := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "api", Status: "active", Metadata: "{}"})
	mustInsert(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", ParentID: "c3-1", Status: "active", Metadata: "{}"})

	// Retire the container alone → would orphan its child.
	if r := retireGate(s, []changeset.Patch{{Target: "c3-1", Scope: changeset.ScopeRetire}}); len(r) == 0 || !strings.Contains(r[0], "orphan") {
		t.Fatalf("retiring a parent with a live child must be refused, got: %v", r)
	}
	// Retire parent + child together → clean.
	if r := retireGate(s, []changeset.Patch{
		{Target: "c3-1", Scope: changeset.ScopeRetire},
		{Target: "c3-101", Scope: changeset.ScopeRetire},
	}); len(r) != 0 {
		t.Errorf("retiring parent + child together must be allowed, got: %v", r)
	}
	// Reparent the child away in the same unit → clean.
	mustInsert(t, s, &store.Entity{ID: "c3-2", Type: "container", Title: "web", Status: "active", Metadata: "{}"})
	if r := retireGate(s, []changeset.Patch{
		{Target: "c3-1", Scope: changeset.ScopeRetire},
		{Target: "c3-101", Scope: changeset.ScopeFrontmatter, Parent: "c3-2"},
	}); len(r) != 0 {
		t.Errorf("retiring a parent whose child is reparented away in the unit must be allowed, got: %v", r)
	}
}

func TestRetireGate_RefusesDanglingCiter(t *testing.T) {
	s, _ := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "ref-jwt", Type: "ref", Title: "jwt", Status: "active", Metadata: "{}"})
	mustInsert(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", Status: "active", Metadata: "{}"})
	if err := s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "ref-jwt", RelType: "uses"}); err != nil {
		t.Fatal(err)
	}

	// Retire the cited ref alone → the citer's citation would dangle.
	if r := retireGate(s, []changeset.Patch{{Target: "ref-jwt", Scope: changeset.ScopeRetire}}); len(r) == 0 || !strings.Contains(r[0], "cites") {
		t.Fatalf("retiring a cited ref must be refused, got: %v", r)
	}
	// Retire ref + its only citer together → clean.
	if r := retireGate(s, []changeset.Patch{
		{Target: "ref-jwt", Scope: changeset.ScopeRetire},
		{Target: "c3-101", Scope: changeset.ScopeRetire},
	}); len(r) != 0 {
		t.Errorf("retiring a ref together with its only citer must be allowed, got: %v", r)
	}
}

func mustInsert(t *testing.T, s *store.Store, e *store.Entity) {
	t.Helper()
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert %s: %v", e.ID, err)
	}
}
