package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// buildChangeUnitView projects a unit's external arm: a carrier is "applied" only
// when the live code-map equals its declared globs, and each declared glob that
// resolves to no file on disk is surfaced as unresolved.
func TestBuildChangeUnitView_ExternalArm(t *testing.T) {
	projectRoot := t.TempDir()
	c3Dir := filepath.Join(projectRoot, ".c3")
	if err := os.MkdirAll(filepath.Join(projectRoot, "src", "real"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "src", "real", "f.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.InsertEntity(&store.Entity{ID: "c3-1", Type: "component", Title: "x", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-1", "# x\n\n## Goal\n\nx.\n"); err != nil {
		t.Fatal(err)
	}

	codemaps := []changeset.CodemapChange{{
		Source: "01.codemap.md", Target: "c3-1",
		Globs: []string{"src/real/**", "src/ghost/**"},
	}}
	opts := ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "u"}

	// Live code-map differs from the declared set → not yet applied.
	if err := s.SetCodeMap("c3-1", []string{"src/real/**"}); err != nil {
		t.Fatal(err)
	}
	view := buildChangeUnitView(opts, nil, codemaps)
	if len(view.Codemaps) != 1 {
		t.Fatalf("want 1 codemap arm, got %d", len(view.Codemaps))
	}
	arm := view.Codemaps[0]
	if arm.Applied {
		t.Error("declared set differs from live code-map → must not be applied")
	}
	if len(arm.Unresolved) != 1 || arm.Unresolved[0] != "src/ghost/**" {
		t.Errorf("unresolved = %v, want [src/ghost/**]", arm.Unresolved)
	}

	// Now make the live code-map equal the declared set → applied.
	if err := s.SetCodeMap("c3-1", []string{"src/real/**", "src/ghost/**"}); err != nil {
		t.Fatal(err)
	}
	view2 := buildChangeUnitView(opts, nil, codemaps)
	if !view2.Codemaps[0].Applied {
		t.Error("live code-map equals declared set → must be applied")
	}
}

func TestSameStringSet(t *testing.T) {
	cases := []struct {
		a, b []string
		want bool
	}{
		{[]string{"a", "b"}, []string{"b", "a"}, true},
		{[]string{"a"}, []string{"a", "b"}, false},
		{[]string{"a", "b"}, []string{"a", "a"}, false},
		{nil, nil, true},
	}
	for _, c := range cases {
		if got := sameStringSet(c.a, c.b); got != c.want {
			t.Errorf("sameStringSet(%v,%v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
