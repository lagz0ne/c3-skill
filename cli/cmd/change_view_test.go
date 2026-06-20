package cmd

import (
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// buildChangeUnitView projects a unit's patch arm: each patch surfaces its source,
// target, and scope. (A fact's external code binding now lives in its eval spec, not
// in the change-unit, so there is no codemap arm.)
func TestBuildChangeUnitView_PatchArm(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	opts := ChangeApplyOptions{Store: s, C3Dir: filepath.Join(t.TempDir(), ".c3"), UnitID: "u"}
	patches := []changeset.Patch{
		{Source: "01.patch.md", Target: "c3-101", Scope: changeset.ScopeWhole, Type: "component", Content: "# auth\n"},
	}
	view := buildChangeUnitView(opts, patches)
	if view.Unit != "u" {
		t.Errorf("unit = %q, want u", view.Unit)
	}
	if len(view.Patches) != 1 {
		t.Fatalf("want 1 patch arm, got %d", len(view.Patches))
	}
	arm := view.Patches[0]
	if arm.Source != "01.patch.md" || arm.Target != "c3-101" || arm.Scope != string(changeset.ScopeWhole) {
		t.Errorf("patch arm = %+v", arm)
	}
}
