package changeset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCodemapCarrier(t *testing.T) {
	raw := "---\ntarget: c3-101\nbase:\n  - old/**\n---\nsrc/auth/**\n# a comment, ignored\n\nsrc/login/**\n"
	c, err := ParseCodemapCarrier("01.codemap.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if c.Target != "c3-101" {
		t.Errorf("target = %q, want c3-101", c.Target)
	}
	if len(c.Globs) != 2 || c.Globs[0] != "src/auth/**" || c.Globs[1] != "src/login/**" {
		t.Errorf("globs = %v, want [src/auth/** src/login/**]", c.Globs)
	}
	if len(c.Base) != 1 || c.Base[0] != "old/**" {
		t.Errorf("base = %v, want [old/**]", c.Base)
	}
}

func TestParseCodemapCarrier_MissingTarget(t *testing.T) {
	if _, err := ParseCodemapCarrier("x.codemap.md", "---\nbase: []\n---\nsrc/**\n"); err == nil {
		t.Fatal("a carrier with no target must error")
	}
}

func TestParseCodemapCarrier_EmptyBodyRejected(t *testing.T) {
	// A carrier full-replaces the target's code-map, so an empty one would silently
	// clear it — reject rather than allow that footgun.
	if _, err := ParseCodemapCarrier("x.codemap.md", "---\ntarget: c3-1\n---\n"); err == nil {
		t.Fatal("an empty carrier body must be rejected")
	}
}

func TestParseCodemapCarrier_DeduplicatesGlobs(t *testing.T) {
	c, err := ParseCodemapCarrier("x.codemap.md", "---\ntarget: c3-1\n---\nsrc/a/**\nsrc/a/**\nsrc/b/**\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Globs) != 2 || c.Globs[0] != "src/a/**" || c.Globs[1] != "src/b/**" {
		t.Errorf("duplicate globs must be collapsed, got %v", c.Globs)
	}
}

func TestReadCodemapDir(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "02-b.codemap.md", "---\ntarget: c3-2\n---\nb/**\n")
	mustWrite(t, dir, "01-a.codemap.md", "---\ntarget: c3-1\n---\na/**\n")
	mustWrite(t, dir, "01-x.patch.md", "---\ntarget: c3-1\nscope: whole\ntype: component\n---\nbody\n")

	changes, err := ReadCodemapDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 {
		t.Fatalf("want 2 carriers (the .patch.md ignored), got %d", len(changes))
	}
	if changes[0].Target != "c3-1" || changes[1].Target != "c3-2" {
		t.Errorf("carriers must be filename-ordered, got %q then %q", changes[0].Target, changes[1].Target)
	}
}

func TestReadCodemapDir_MissingFolder(t *testing.T) {
	changes, err := ReadCodemapDir(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("missing folder must not error: %v", err)
	}
	if changes != nil {
		t.Errorf("missing folder should yield nil, got %v", changes)
	}
}

// A change-unit folder commonly holds both kinds of material; each reader must see
// only its own suffix.
func TestCarrierAndPatchDiscovery_NoCrossContamination(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "01.patch.md", "---\ntarget: c3-1\nscope: whole\ntype: component\n---\nbody\n")
	mustWrite(t, dir, "01.codemap.md", "---\ntarget: c3-1\n---\nsrc/**\n")

	patches, err := ReadPatchDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(patches) != 1 {
		t.Fatalf("ReadPatchDir should see exactly 1 patch (not the carrier), got %d", len(patches))
	}
	codemaps, err := ReadCodemapDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(codemaps) != 1 {
		t.Fatalf("ReadCodemapDir should see exactly 1 carrier (not the patch), got %d", len(codemaps))
	}
}

func mustWrite(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
