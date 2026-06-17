package changeset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPatchDir_OrdersByFilename(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("02-users.patch.md", "---\ntarget: c3-110\nscope: whole\ntype: component\n---\nbody2\n")
	write("01-auth.patch.md", "---\ntarget: c3-101\nscope: block\nbase: c3-101#n1@v1:sha256:"+strings.Repeat("a", 64)+"\n---\nbody1\n")
	write("README.md", "not a patch — ignored")

	patches, err := ReadPatchDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(patches) != 2 {
		t.Fatalf("want 2 patches (README.md ignored), got %d", len(patches))
	}
	if patches[0].Target != "c3-101" || patches[1].Target != "c3-110" {
		t.Errorf("patches must be ordered by filename: got %s then %s", patches[0].Target, patches[1].Target)
	}
}

func TestReadPatchDir_MissingDirIsEmpty(t *testing.T) {
	patches, err := ReadPatchDir(filepath.Join(t.TempDir(), "no-such-folder"))
	if err != nil {
		t.Fatalf("a missing patch folder is not an error, got: %v", err)
	}
	if len(patches) != 0 {
		t.Fatalf("missing folder → 0 patches, got %d", len(patches))
	}
}

func TestReadPatchDir_PropagatesParseError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.patch.md"), []byte("---\nscope: block\n---\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPatchDir(dir); err == nil {
		t.Fatal("a malformed patch file must surface as an error")
	}
}
