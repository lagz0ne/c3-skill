package changeset

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// seedEntity inserts a typed entity (parent first, so a child's ParentID FK resolves)
// and writes its body. Body write resets nodes, goal, and seal.
func seedEntity(t *testing.T, s *store.Store, e *store.Entity, body string) {
	t.Helper()
	if e.Metadata == "" {
		e.Metadata = "{}"
	}
	if e.Status == "" {
		e.Status = "active"
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert %s: %v", e.ID, err)
	}
	if body != "" {
		if err := content.WriteEntity(s, e.ID, body); err != nil {
			t.Fatalf("write %s: %v", e.ID, err)
		}
	}
}

const emptyComponents = "# API\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n| --- | --- | --- | --- | --- |\n"

// reconcileHook is the changeset-level membership hook (no canvas validation — that
// is the cmd hook's job); it lets a test drive the deterministic apply saga.
func reconcileHook() *ApplyHooks {
	return &ApplyHooks{ReconcileMembership: func(ts *store.Store, parentID string) error {
		e, err := ts.GetEntity(parentID)
		if err != nil {
			return nil
		}
		section, childType := MembershipSection(e.Type)
		if section == "" {
			return nil
		}
		_, err = ReconcileMembershipBody(ts, parentID, section, childType)
		return err
	}}
}

func TestReconcile_AddsRowForNewChild(t *testing.T) {
	s := openMem(t)
	seedEntity(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "API"}, emptyComponents)
	seedEntity(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", Category: "foundation", ParentID: "c3-1"},
		"# auth\n\n## Goal\n\nVerifies request tokens.\n")

	changed, err := ReconcileMembershipBody(s, "c3-1", "Components", "component")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("reconcile must add the missing child row")
	}
	body, _ := content.ReadEntity(s, "c3-1")
	for _, want := range []string{"c3-101", "auth", "foundation", "active", "Verifies request tokens"} {
		if !strings.Contains(body, want) {
			t.Errorf("synthesized row missing %q (identity derived from child, descriptive from Goal):\n%s", want, body)
		}
	}
}

func TestReconcile_DropsOrphanRow(t *testing.T) {
	s := openMem(t)
	seedEntity(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "API"},
		emptyComponents+"| c3-101 | auth | foundation | active | does auth |\n| c3-199 | ghost | feature | active | nonexistent |\n")
	// Only c3-101 is a real child; c3-199 has no entity ⇒ its row is an orphan.
	seedEntity(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", Category: "foundation", ParentID: "c3-1"},
		"# auth\n\n## Goal\n\nVerifies tokens.\n")

	changed, err := ReconcileMembershipBody(s, "c3-1", "Components", "component")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("reconcile must drop the orphan row")
	}
	body, _ := content.ReadEntity(s, "c3-1")
	if strings.Contains(body, "c3-199") || strings.Contains(body, "ghost") {
		t.Errorf("orphan row must be dropped:\n%s", body)
	}
	if !strings.Contains(body, "c3-101") {
		t.Errorf("real child row must remain:\n%s", body)
	}
}

func TestReconcile_PreservesAuthoredRefreshesIdentity(t *testing.T) {
	s := openMem(t)
	seedEntity(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "API"},
		emptyComponents+"| c3-101 | oldname | foundation | active | the parent's own framing |\n")
	// The child's live title/category differ from the row's stale identity cells.
	seedEntity(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "newname", Category: "feature", ParentID: "c3-1"},
		"# newname\n\n## Goal\n\nNew goal.\n")

	changed, err := ReconcileMembershipBody(s, "c3-1", "Components", "component")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("identity refresh should change the table")
	}
	body, _ := content.ReadEntity(s, "c3-1")
	if !strings.Contains(body, "newname") || strings.Contains(body, "oldname") {
		t.Errorf("identity Name must refresh to the live title:\n%s", body)
	}
	if !strings.Contains(body, "feature") {
		t.Errorf("identity Category must refresh:\n%s", body)
	}
	if !strings.Contains(body, "the parent's own framing") {
		t.Errorf("authored Goal Contribution must be preserved, not overwritten:\n%s", body)
	}
}

func TestReconcile_IdempotentAndDeterministic(t *testing.T) {
	seal := func() string {
		s := openMem(t)
		seedEntity(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "API"}, emptyComponents)
		seedEntity(t, s, &store.Entity{ID: "c3-101", Type: "component", Title: "auth", Category: "foundation", ParentID: "c3-1"},
			"# auth\n\n## Goal\n\nVerifies tokens.\n")
		seedEntity(t, s, &store.Entity{ID: "c3-102", Type: "component", Title: "store", Category: "foundation", ParentID: "c3-1"},
			"# store\n\n## Goal\n\nPersists data.\n")
		if changed, _ := ReconcileMembershipBody(s, "c3-1", "Components", "component"); !changed {
			t.Fatal("first reconcile should add both rows")
		}
		// Second reconcile is a no-op — the table already matches.
		if changed, _ := ReconcileMembershipBody(s, "c3-1", "Components", "component"); changed {
			t.Error("reconcile must be idempotent (second pass is a no-op)")
		}
		e, _ := s.GetEntity("c3-1")
		return e.RootMerkle
	}
	// Same base, twice ⇒ identical seal: apply is a deterministic saga.
	if a, b := seal(), seal(); a != b {
		t.Errorf("deterministic saga: same base must seal identically, got %s != %s", a, b)
	}
}

// The integrity-by-construction proof: a change-unit whose ONLY material is a
// child's parent: declaration (no membership-row patch) still produces a consistent
// parent table at commit. The author never touches the table — the tool does.
func TestApply_MembershipSynthesizedFromParentEdgeAlone(t *testing.T) {
	s := openMem(t)
	seedEntity(t, s, &store.Entity{ID: "c3-0", Type: "system", Title: "Platform"},
		"# Platform\n\n## Containers\n\n| ID | Name | Status |\n| --- | --- | --- |\n")
	seedEntity(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "API"},
		"# API\n\n## Goal\n\nServes requests.\n")
	e1, _ := s.GetEntity("c3-1")
	base := fmt.Sprintf("c3-1@v%d:sha256:%s", e1.Version, e1.RootMerkle)

	p := Patch{Target: "c3-1", Scope: ScopeFrontmatter, Base: base, Parent: "c3-0", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, reconcileHook()); err != nil {
		t.Fatalf("apply: %v", err)
	}
	body, _ := content.ReadEntity(s, "c3-0")
	if !strings.Contains(body, "c3-1") || !strings.Contains(body, "API") {
		t.Errorf("declaring parent: alone must synthesize the parent's membership row:\n%s", body)
	}
}

// Reparenting heals the parent a child LEAVES: the old parent's row is dropped and
// the new parent's row appears — in one atomic apply, with no membership patch.
func TestApply_ReparentHealsOldAndNewParent(t *testing.T) {
	s := openMem(t)
	seedEntity(t, s, &store.Entity{ID: "c3-0a", Type: "system", Title: "Old"},
		"# Old\n\n## Containers\n\n| ID | Name | Status |\n| --- | --- | --- |\n")
	seedEntity(t, s, &store.Entity{ID: "c3-0b", Type: "system", Title: "New"},
		"# New\n\n## Containers\n\n| ID | Name | Status |\n| --- | --- | --- |\n")
	seedEntity(t, s, &store.Entity{ID: "c3-1", Type: "container", Title: "API", ParentID: "c3-0a"},
		"# API\n\n## Goal\n\nServes requests.\n")
	// First make c3-0a consistent (it owns c3-1).
	if _, err := ReconcileMembershipBody(s, "c3-0a", "Containers", "container"); err != nil {
		t.Fatal(err)
	}
	if body, _ := content.ReadEntity(s, "c3-0a"); !strings.Contains(body, "c3-1") {
		t.Fatalf("setup: c3-0a should list c3-1:\n%s", body)
	}

	e1, _ := s.GetEntity("c3-1")
	base := fmt.Sprintf("c3-1@v%d:sha256:%s", e1.Version, e1.RootMerkle)
	p := Patch{Target: "c3-1", Scope: ScopeFrontmatter, Base: base, Parent: "c3-0b", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, reconcileHook()); err != nil {
		t.Fatalf("apply: %v", err)
	}
	oldBody, _ := content.ReadEntity(s, "c3-0a")
	newBody, _ := content.ReadEntity(s, "c3-0b")
	if strings.Contains(oldBody, "c3-1") {
		t.Errorf("old parent must lose the reparented child's row:\n%s", oldBody)
	}
	if !strings.Contains(newBody, "c3-1") {
		t.Errorf("new parent must gain the row:\n%s", newBody)
	}
}
