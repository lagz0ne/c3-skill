package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGlobPattern(t *testing.T) {
	globs := []string{"src/**", "src/*.ts", "a?b.go"}
	for _, p := range globs {
		if !IsGlobPattern(p) {
			t.Errorf("%q should be a glob pattern", p)
		}
	}
	literals := []string{"src/auth.go", "pages/[id].tsx", "README.md"}
	for _, p := range literals {
		if IsGlobPattern(p) {
			t.Errorf("%q should be treated as a literal path (brackets are literal)", p)
		}
	}
}

func TestGlobFiles(t *testing.T) {
	dir := t.TempDir()
	// Create files including one with bracket in name
	os.WriteFile(filepath.Join(dir, "normal.ts"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "pages"), 0755)
	os.WriteFile(filepath.Join(dir, "pages", "[id].tsx"), []byte(""), 0644)

	fsys := os.DirFS(dir)

	// Normal glob
	matches, err := GlobFiles(fsys, "*.ts")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0] != "normal.ts" {
		t.Errorf("normal glob = %v", matches)
	}

	// Bracket pattern (framework route param)
	matches, err = GlobFiles(fsys, "pages/[id].tsx")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Errorf("bracket glob should match, got %v", matches)
	}
}
