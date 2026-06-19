package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func joined(r []string) string { return strings.Join(r, " | ") }

// Orphan: retiring a parent is refused while a child still points at it — UNLESS the
// same unit reparents or retires that child (the gate checks the post-apply state).
func TestRetireGate_OrphanRefusedUnlessHandledInUnit(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "api", Status: "active", Metadata: "{}"})
	mustInsert(t, s, &store.Entity{ID: "c3-2", Type: "container", Title: "web", Status: "active", Metadata: "{}"})
	mustInsert(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", ParentID: "c3-1", Status: "active", Metadata: "{}"})
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nAuth.\n"); err != nil {
		t.Fatal(err)
	}
	retireParent := changeset.Patch{Target: "c3-1", Scope: changeset.ScopeRetire, Source: "01.patch.md"}

	// retire the parent alone → orphans the child.
	if r := retireGate(s, c3Dir, []changeset.Patch{retireParent}, nil); len(r) == 0 || !strings.Contains(joined(r), "orphan") {
		t.Fatalf("retiring a parent with a live child must be refused, got: %v", r)
	}
	// retire the parent AND reparent the child away in the same unit → allowed.
	reparent := changeset.Patch{Target: "c3-101", Scope: changeset.ScopeFrontmatter, Parent: "c3-2", Source: "02.patch.md"}
	if r := retireGate(s, c3Dir, []changeset.Patch{retireParent, reparent}, nil); len(r) != 0 {
		t.Errorf("retire + reparent-the-child in the SAME unit must be allowed, got: %v", r)
	}
	// retire both parent and child → allowed.
	retireChild := changeset.Patch{Target: "c3-101", Scope: changeset.ScopeRetire, Source: "03.patch.md"}
	if r := retireGate(s, c3Dir, []changeset.Patch{retireParent, retireChild}, nil); len(r) != 0 {
		t.Errorf("retire parent + child together must be allowed, got: %v", r)
	}
}

// Dangle: retiring a cited ref is refused while a citer still names it — UNLESS the
// same unit re-cites that citer away (the post-apply body no longer names the ref).
// This is the design-system / rearchitect finding: re-point + retire in one unit.
func TestRetireGate_DangleRefusedUnlessRecitedInUnit(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "ref-jwt", Type: "ref", Title: "jwt", Status: "active", Metadata: "{}"})
	mustInsert(t, s, &store.Entity{ID: "ref-new", Type: "ref", Title: "new", Status: "active", Metadata: "{}"})
	mustInsert(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", Status: "active", Metadata: "{}"})
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Refs\n\n| Reference | Note |\n| --- | --- |\n| ref-jwt | uses |\n"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "ref-jwt", RelType: "uses"}); err != nil {
		t.Fatal(err)
	}
	retire := changeset.Patch{Target: "ref-jwt", Scope: changeset.ScopeRetire, Source: "01.patch.md"}

	// retire the ref alone → the citer's body still names it → dangle.
	if r := retireGate(s, c3Dir, []changeset.Patch{retire}, nil); len(r) == 0 || !strings.Contains(joined(r), "dangle") {
		t.Fatalf("retiring a cited ref must be refused, got: %v", r)
	}
	// retire the ref AND re-cite the citer to ref-new in the same unit → allowed.
	handle := blockRowHandle(t, s, "c3-101", "ref-jwt")
	recite := changeset.Patch{Target: "c3-101", Scope: changeset.ScopeBlock, Base: handle, Content: "| ref-new | uses |", Source: "02.patch.md"}
	if r := retireGate(s, c3Dir, []changeset.Patch{retire, recite}, nil); len(r) != 0 {
		t.Errorf("retire + re-cite-the-citer in the SAME unit must be allowed, got: %v", r)
	}
}

func mustInsert(t *testing.T, s *store.Store, e *store.Entity) {
	t.Helper()
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert %s: %v", e.ID, err)
	}
}

func blockRowHandle(t *testing.T, s *store.Store, id, contains string) string {
	t.Helper()
	e, _ := s.GetEntity(id)
	nodes, _ := s.NodesForEntity(id)
	for _, n := range nodes {
		if n.Type == "table_row" && strings.Contains(n.Content, contains) {
			return fmt.Sprintf("%s#n%d@v%d:sha256:%s", id, n.ID, e.Version, n.Hash)
		}
	}
	t.Fatalf("no table_row containing %q in %s", contains, id)
	return ""
}
