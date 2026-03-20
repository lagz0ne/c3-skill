package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRun_Version(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"--version"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "dev") {
		t.Errorf("version output = %q", buf.String())
	}
}

func TestRun_Help(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"--help"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "c3x") {
		t.Error("help should mention c3x")
	}
}

func TestRun_EmptyArgs(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Commands:") {
		t.Error("empty args should show help")
	}
}

func TestRun_Capabilities(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"capabilities"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Command") {
		t.Error("capabilities should show command table")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	err := run([]string{"--c3-dir", t.TempDir(), "nonexistent"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestRun_NoC3Dir(t *testing.T) {
	err := run([]string{"--c3-dir", filepath.Join(t.TempDir(), "nope"), "list"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when no .c3/ found")
	}
}

func TestRun_ListWithDB(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "list", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "c3-0") {
		t.Error("list should include c3-0")
	}
}

func TestRun_CheckWithDB(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "check", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Schema(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "schema", "component"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_GraphMissingID(t *testing.T) {
	c3Dir := setupC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "graph"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for graph without entity ID")
	}
}

func TestRun_LookupMissingArg(t *testing.T) {
	c3Dir := setupC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "lookup"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for lookup without file path")
	}
}

func TestRun_LegacyFormatBlocked(t *testing.T) {
	// Create .c3/ with markdown files but no DB
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)
	os.WriteFile(filepath.Join(c3Dir, "README.md"), []byte("---\nid: c3-0\ntitle: test\n---\n"), 0644)

	err := run([]string{"--c3-dir", c3Dir, "list"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for legacy format without DB")
	}
	if !strings.Contains(err.Error(), "markdown files") {
		t.Errorf("error should mention markdown files: %v", err)
	}
}

func TestRun_LegacyCheckAllowed(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)
	os.WriteFile(filepath.Join(c3Dir, "README.md"), []byte("---\nid: c3-0\ntitle: test\n---\n# test\n\n## Goal\n\nTest.\n"), 0644)

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "check"}, &buf)
	if err != nil {
		t.Fatalf("legacy check should work: %v", err)
	}
}

func TestRun_MarketplaceHelp(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"marketplace"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "exists.txt")
	os.WriteFile(f, []byte("x"), 0644)

	if !fileExists(f) {
		t.Error("should return true for existing file")
	}
	if fileExists(filepath.Join(tmp, "nope.txt")) {
		t.Error("should return false for missing file")
	}
}

func TestHasMarkdownFiles(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		dir := t.TempDir()
		if hasMarkdownFiles(dir) {
			t.Error("empty dir should not have markdown files")
		}
	})

	t.Run("top-level md", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "README.md"), []byte("test"), 0644)
		if !hasMarkdownFiles(dir) {
			t.Error("should detect top-level .md files")
		}
	})

	t.Run("nested md", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "containers")
		os.MkdirAll(sub, 0755)
		os.WriteFile(filepath.Join(sub, "c3-1.md"), []byte("test"), 0644)
		if !hasMarkdownFiles(dir) {
			t.Error("should detect .md files in subdirectories")
		}
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		if hasMarkdownFiles("/nonexistent/path") {
			t.Error("should return false for nonexistent dir")
		}
	})

	t.Run("_index dir skipped", func(t *testing.T) {
		dir := t.TempDir()
		idx := filepath.Join(dir, "_index")
		os.MkdirAll(idx, 0755)
		os.WriteFile(filepath.Join(idx, "index.md"), []byte("test"), 0644)
		if hasMarkdownFiles(dir) {
			t.Error("_index dir should be skipped")
		}
	})
}

func TestRun_Add(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "add", "ref", "rate-limiting"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_AddJSON(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "add", "ref", "caching", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ref-caching") {
		t.Errorf("JSON add output should contain ref-caching: %s", buf.String())
	}
}

func TestRun_AddRich(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "add", "ref", "logging", "--goal", "Structured logging"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Set(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "goal", "Updated goal"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Wire(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Unwire(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	// Wire first
	run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "ref-jwt"}, &buf)
	buf.Reset()
	err := run([]string{"--c3-dir", c3Dir, "unwire", "c3-101", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_WireThreeArgs(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "cite", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Delete(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "delete", "ref-jwt", "--dry-run"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Query(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "query", "auth", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Diff(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "diff"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_DiffMark(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "diff", "--mark"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Impact(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "impact", "c3-101", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Export(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "exported")
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "export", outDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Graph(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "graph", "c3-0"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Codemap(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "codemap"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Lookup(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "lookup", "src/main.go"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_MarketplaceList(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"marketplace", "list"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_ListFlat(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "list", "--flat"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_ListCompact(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "list", "--compact"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_SetWithSection(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "--section", "Goal", "New goal text"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_WireRemoveFlag(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	// Wire first
	run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "ref-jwt"}, &buf)
	buf.Reset()
	err := run([]string{"--c3-dir", c3Dir, "wire", "--remove", "c3-101", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

// setupC3DB creates a temp .c3/ dir with a SQLite DB containing a minimal fixture.
func setupC3DB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	dbPath := filepath.Join(c3Dir, "c3.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer s.Close()

	s.InsertEntity(&store.Entity{
		ID: "c3-0", Type: "system", Title: "TestProject",
		Slug: "", Status: "active", Metadata: "{}",
		Body: "# TestProject\n\n## Goal\n\nTest.\n",
	})

	return c3Dir
}

// setupRichC3DB creates a .c3/ dir with DB containing containers, components, refs.
func setupRichC3DB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)
	// Create container dirs for add commands that write files
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	dbPath := filepath.Join(c3Dir, "c3.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer s.Close()

	entities := []*store.Entity{
		{ID: "c3-0", Type: "system", Title: "TestProject", Slug: "", Body: "# TestProject\n\n## Goal\n\nTest.\n\n## Containers\n\n| ID | Name | Boundary | Goal |\n|----|------|----------|------|\n", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "api", Slug: "api", ParentID: "c3-0", Goal: "Serve API", Boundary: "service", Body: "# api\n\n## Goal\n\nServe API.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n", Status: "active", Metadata: "{}"},
		{ID: "c3-101", Type: "component", Title: "auth", Slug: "auth", Category: "foundation", ParentID: "c3-1", Body: "# auth\n\n## Goal\n\nHandle auth.\n\n## Related Refs\n\n| Ref | Role |\n|-----|------|\n", Status: "active", Metadata: "{}"},
		{ID: "ref-jwt", Type: "ref", Title: "JWT", Slug: "jwt", Goal: "JWT tokens", Body: "# JWT\n\n## Goal\n\nJWT tokens.\n", Status: "active", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	s.AddRelationship(&store.Relationship{FromID: "ref-jwt", ToID: "c3-1", RelType: "scope"})

	return c3Dir
}
