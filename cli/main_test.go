package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
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

func TestRun_NoDatabaseBlocked(t *testing.T) {
	// Create .c3/ without a database
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	err := run([]string{"--c3-dir", c3Dir, "list"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when no database exists")
	}
	if !strings.Contains(err.Error(), "no database found") {
		t.Errorf("error should mention 'no database found': %v", err)
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
		{ID: "c3-0", Type: "system", Title: "TestProject", Slug: "", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "api", Slug: "api", ParentID: "c3-0", Goal: "Serve API", Boundary: "service", Status: "active", Metadata: "{}"},
		{ID: "c3-101", Type: "component", Title: "auth", Slug: "auth", Category: "foundation", ParentID: "c3-1", Status: "active", Metadata: "{}"},
		{ID: "ref-jwt", Type: "ref", Title: "JWT", Slug: "jwt", Goal: "JWT tokens", Status: "active", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	s.AddRelationship(&store.Relationship{FromID: "ref-jwt", ToID: "c3-1", RelType: "scope"})

	// Populate node trees
	bodies := map[string]string{
		"c3-0":    "# TestProject\n\n## Goal\n\nTest.\n\n## Containers\n\n| ID | Name | Boundary | Goal |\n|----|------|----------|------|\n",
		"c3-1":    "# api\n\n## Goal\n\nServe API.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n",
		"c3-101":  "# auth\n\n## Goal\n\nHandle auth.\n\n## Related Refs\n\n| Ref | Role |\n|-----|------|\n",
		"ref-jwt": "# JWT\n\n## Goal\n\nJWT tokens.\n",
	}
	for id, body := range bodies {
		if err := content.WriteEntity(s, id, body); err != nil {
			t.Fatalf("seed nodes %s: %v", id, err)
		}
	}

	return c3Dir
}
