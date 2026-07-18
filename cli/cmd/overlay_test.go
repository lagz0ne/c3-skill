package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// TestWithUnitOverlay_PreviewsStagedThenRollsBack — a change-unit's staged
// create-patch is visible inside the overlay (the preview applies the real apply
// path) but never committed (always rolled back). This is the lens that lets
// `graph --unit X` show staged work before apply.
func TestWithUnitOverlay_PreviewsStagedThenRollsBack(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	body := "# ref-new\n\n## Goal\n\nStandardize a brand new pattern across components here now.\n\n## Choice\n\nUse the chosen concrete approach for this new pattern.\n\n## Why\n\nRationale explaining why this choice beats the alternatives here.\n"
	writePatch(t, c3Dir, "adr-x", "01-create.patch.md",
		"---\ntarget: ref-new\nscope: whole\ntype: ref\ntitle: New Ref\n---\n"+body)

	if _, err := s.GetEntity("ref-new"); err == nil {
		t.Fatal("precondition: ref-new must not exist before the overlay")
	}

	seen := false
	if err := WithUnitOverlay(s, c3Dir, "adr-x", func(ts *store.Store) error {
		if _, err := ts.GetEntity("ref-new"); err == nil {
			seen = true
		}
		return nil
	}); err != nil {
		t.Fatalf("overlay: %v", err)
	}
	if !seen {
		t.Error("overlay must preview the staged ref-new")
	}
	if _, err := s.GetEntity("ref-new"); err == nil {
		t.Error("overlay must roll back — ref-new leaked into the committed store")
	}
}

func TestWithUnitOverlay_RejectsBodyOwnedFrontmatterUses(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedComponentWithBodyOwnedUses(t, s)
	e, _ := s.GetEntity("c3-101")
	writePatch(t, c3Dir, "adr-edge", "01-edge.patch.md",
		"---\ntarget: c3-101\nscope: frontmatter\nbase: c3-101@v1:sha256:"+e.RootMerkle+"\nuses:\n  - ref-new\n---\n")

	err := WithUnitOverlay(s, c3Dir, "adr-edge", func(*store.Store) error {
		t.Fatal("forbidden overlay must not invoke its callback")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "body-owned") || !strings.Contains(err.Error(), "Governance") {
		t.Fatalf("overlay must share the apply edge-source gate, got %v", err)
	}
	rels, _ := s.RelationshipsFrom("c3-101")
	if len(rels) != 1 || rels[0].ToID != "ref-old" {
		t.Fatalf("rejected overlay changed committed edges: %+v", rels)
	}
}

// TestWithUnitOverlay_MissingUnitFailsLoud — an unknown/empty unit fails loudly
// rather than silently showing applied state.
func TestWithUnitOverlay_MissingUnitFailsLoud(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	if err := WithUnitOverlay(s, c3Dir, "adr-does-not-exist", func(*store.Store) error { return nil }); err == nil {
		t.Fatal("overlay of a non-existent unit must error, never silently use applied state")
	}
}
