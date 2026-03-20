package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunMigrate(t *testing.T) {
	c3Dir := createFixture(t)

	var buf bytes.Buffer
	err := RunMigrate(c3Dir, true, &buf)
	if err != nil {
		t.Fatalf("RunMigrate failed: %v", err)
	}

	// Open the created DB and verify entities
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer s.Close()

	// Verify all 7 entities were imported
	all, err := s.AllEntities()
	if err != nil {
		t.Fatalf("AllEntities: %v", err)
	}
	if len(all) != 7 {
		t.Errorf("expected 7 entities, got %d", len(all))
		for _, e := range all {
			t.Logf("  %s (%s)", e.ID, e.Type)
		}
	}

	// Verify specific entities
	tests := []struct {
		id       string
		typ      string
		title    string
		parentID string
	}{
		{"c3-0", "system", "TestProject", ""},
		{"c3-1", "container", "api", "c3-0"},
		{"c3-2", "container", "web", "c3-0"},
		{"c3-101", "component", "auth", "c3-1"},
		{"c3-110", "component", "users", "c3-1"},
		{"ref-jwt", "ref", "JWT Authentication", ""},
		{"adr-20260226-use-go", "adr", "Use Go for CLI", ""},
	}

	for _, tt := range tests {
		e, err := s.GetEntity(tt.id)
		if err != nil {
			t.Errorf("GetEntity(%s): %v", tt.id, err)
			continue
		}
		if e.Type != tt.typ {
			t.Errorf("%s: type = %q, want %q", tt.id, e.Type, tt.typ)
		}
		if e.Title != tt.title {
			t.Errorf("%s: title = %q, want %q", tt.id, e.Title, tt.title)
		}
		if e.ParentID != tt.parentID {
			t.Errorf("%s: parentID = %q, want %q", tt.id, e.ParentID, tt.parentID)
		}
	}

	// Verify relationships
	// c3-101 -> ref-jwt (uses)
	rels, err := s.RelationshipsFrom("c3-101")
	if err != nil {
		t.Fatalf("RelationshipsFrom(c3-101): %v", err)
	}
	if !hasRel(rels, "c3-101", "ref-jwt", "uses") {
		t.Error("missing relationship: c3-101 -> ref-jwt (uses)")
	}

	// ref-jwt -> c3-1 (scope)
	rels, err = s.RelationshipsFrom("ref-jwt")
	if err != nil {
		t.Fatalf("RelationshipsFrom(ref-jwt): %v", err)
	}
	if !hasRel(rels, "ref-jwt", "c3-1", "scope") {
		t.Error("missing relationship: ref-jwt -> c3-1 (scope)")
	}

	// adr -> c3-0 (affects)
	rels, err = s.RelationshipsFrom("adr-20260226-use-go")
	if err != nil {
		t.Fatalf("RelationshipsFrom(adr): %v", err)
	}
	if !hasRel(rels, "adr-20260226-use-go", "c3-0", "affects") {
		t.Error("missing relationship: adr-20260226-use-go -> c3-0 (affects)")
	}

	// Verify entity fields
	e, _ := s.GetEntity("c3-1")
	if e.Boundary != "service" {
		t.Errorf("c3-1 boundary = %q, want %q", e.Boundary, "service")
	}
	if e.Goal != "Serve API requests" {
		t.Errorf("c3-1 goal = %q, want %q", e.Goal, "Serve API requests")
	}

	e, _ = s.GetEntity("c3-101")
	if e.Category != "foundation" {
		t.Errorf("c3-101 category = %q, want %q", e.Category, "foundation")
	}

	e, _ = s.GetEntity("adr-20260226-use-go")
	if e.Status != "proposed" {
		t.Errorf("adr status = %q, want %q", e.Status, "proposed")
	}

	// Verify output contains summary
	output := buf.String()
	if !strings.Contains(output, "migrated") {
		t.Error("output should contain migration summary")
	}
}

func TestRunMigrate_KeepOriginals(t *testing.T) {
	c3Dir := createFixture(t)

	var buf bytes.Buffer
	err := RunMigrate(c3Dir, true, &buf)
	if err != nil {
		t.Fatalf("RunMigrate failed: %v", err)
	}

	// Original files should still exist
	originals := []string{
		filepath.Join(c3Dir, "README.md"),
		filepath.Join(c3Dir, "c3-1-api", "README.md"),
		filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"),
		filepath.Join(c3Dir, "refs", "ref-jwt.md"),
		filepath.Join(c3Dir, "adr", "adr-20260226-use-go.md"),
	}
	for _, path := range originals {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to still exist with keepOriginals=true", path)
		}
	}

	// DB should also exist
	if _, err := os.Stat(filepath.Join(c3Dir, "c3.db")); os.IsNotExist(err) {
		t.Error("c3.db should exist")
	}
}

func TestRunMigrate_RemovesOriginals(t *testing.T) {
	c3Dir := createFixture(t)

	var buf bytes.Buffer
	err := RunMigrate(c3Dir, false, &buf)
	if err != nil {
		t.Fatalf("RunMigrate failed: %v", err)
	}

	// Original .md files should be removed
	removedFiles := []string{
		filepath.Join(c3Dir, "README.md"),
		filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"),
		filepath.Join(c3Dir, "refs", "ref-jwt.md"),
		filepath.Join(c3Dir, "adr", "adr-20260226-use-go.md"),
	}
	for _, path := range removedFiles {
		if _, err := os.Stat(path); err == nil {
			t.Errorf("expected %s to be removed with keepOriginals=false", path)
		}
	}

	// DB should exist
	if _, err := os.Stat(filepath.Join(c3Dir, "c3.db")); os.IsNotExist(err) {
		t.Error("c3.db should exist after migration")
	}
}

func TestRunMigrate_AlreadyExists(t *testing.T) {
	c3Dir := createFixture(t)

	// Create a fake c3.db
	os.WriteFile(filepath.Join(c3Dir, "c3.db"), []byte("fake"), 0644)

	var buf bytes.Buffer
	err := RunMigrate(c3Dir, true, &buf)
	if err == nil {
		t.Fatal("expected error when c3.db already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestRunMigrate_WithCodeMap(t *testing.T) {
	c3Dir := createFixture(t)

	// Add a code-map.yaml
	codeMap := `c3-101:
  - src/auth/**
  - src/lib/jwt.ts
c3-110:
  - src/users/**
_exclude:
  - node_modules/**
  - dist/**
`
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), codeMap)

	var buf bytes.Buffer
	err := RunMigrate(c3Dir, true, &buf)
	if err != nil {
		t.Fatalf("RunMigrate failed: %v", err)
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer s.Close()

	// Verify code map entries
	patterns, err := s.CodeMapFor("c3-101")
	if err != nil {
		t.Fatalf("CodeMapFor(c3-101): %v", err)
	}
	if len(patterns) != 2 {
		t.Errorf("expected 2 code map patterns for c3-101, got %d: %v", len(patterns), patterns)
	}

	// Verify excludes
	excludes, err := s.Excludes()
	if err != nil {
		t.Fatalf("Excludes: %v", err)
	}
	if len(excludes) != 2 {
		t.Errorf("expected 2 excludes, got %d: %v", len(excludes), excludes)
	}
}

func TestDocTypeToStoreType(t *testing.T) {
	tests := []struct {
		dt   frontmatter.DocType
		want string
	}{
		{frontmatter.DocContext, "system"},
		{frontmatter.DocContainer, "container"},
		{frontmatter.DocComponent, "component"},
		{frontmatter.DocRef, "ref"},
		{frontmatter.DocADR, "adr"},
		{frontmatter.DocRule, "rule"},
		{frontmatter.DocRecipe, "recipe"},
		{frontmatter.DocUnknown, ""},
	}
	for _, tt := range tests {
		got := docTypeToStoreType(tt.dt)
		if got != tt.want {
			t.Errorf("docTypeToStoreType(%d) = %q, want %q", tt.dt, got, tt.want)
		}
	}
}

func TestAddRelSafe_EmptyToID(t *testing.T) {
	s := createDBFixture(t)
	err := addRelSafe(s, "c3-101", "", "uses")
	if err != nil {
		t.Error("empty toID should be silently ignored")
	}
}

func TestAddRelSafe_ValidRelationship(t *testing.T) {
	s := createDBFixture(t)
	// Both c3-101 and c3-1 exist in fixture
	err := addRelSafe(s, "c3-101", "c3-1", "depends-on")
	if err != nil {
		t.Errorf("valid relationship should succeed: %v", err)
	}
}

func TestRemoveEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	// Create nested empty dirs
	os.MkdirAll(filepath.Join(dir, "a", "b"), 0755)
	os.MkdirAll(filepath.Join(dir, "c"), 0755)
	// Create non-empty dir
	os.MkdirAll(filepath.Join(dir, "d"), 0755)
	os.WriteFile(filepath.Join(dir, "d", "file.txt"), []byte("data"), 0644)

	removeEmptyDirs(dir)

	// Empty dirs should be removed
	if _, err := os.Stat(filepath.Join(dir, "c")); !os.IsNotExist(err) {
		t.Error("empty dir 'c' should be removed")
	}
	// Non-empty dir should survive
	if _, err := os.Stat(filepath.Join(dir, "d")); err != nil {
		t.Error("non-empty dir 'd' should survive")
	}
}

func TestCreateDBFixture(t *testing.T) {
	s := createDBFixture(t)

	all, err := s.AllEntities()
	if err != nil {
		t.Fatalf("AllEntities: %v", err)
	}
	if len(all) != 7 {
		t.Errorf("expected 7 entities, got %d", len(all))
	}

	// Verify relationships
	rels, _ := s.RelationshipsFrom("c3-101")
	if !hasRel(rels, "c3-101", "ref-jwt", "uses") {
		t.Error("missing c3-101 -> ref-jwt (uses)")
	}

	rels, _ = s.RelationshipsFrom("ref-jwt")
	if !hasRel(rels, "ref-jwt", "c3-1", "scope") {
		t.Error("missing ref-jwt -> c3-1 (scope)")
	}

	rels, _ = s.RelationshipsFrom("adr-20260226-use-go")
	if !hasRel(rels, "adr-20260226-use-go", "c3-0", "affects") {
		t.Error("missing adr-20260226-use-go -> c3-0 (affects)")
	}
}

// hasRel checks if a relationship slice contains a specific relationship.
func hasRel(rels []*store.Relationship, from, to, relType string) bool {
	for _, r := range rels {
		if r.FromID == from && r.ToID == to && r.RelType == relType {
			return true
		}
	}
	return false
}
