package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunInitDB_CreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	var buf bytes.Buffer

	err := RunInitDB(c3Dir, "test-project", &buf)
	if err != nil {
		t.Fatalf("RunInitDB failed: %v", err)
	}

	// Verify c3.db exists
	dbPath := filepath.Join(c3Dir, "c3.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("c3.db should exist")
	}

	// Verify config.yaml does NOT exist (only SQLite DB should be created)
	if _, err := os.Stat(filepath.Join(c3Dir, "config.yaml")); err == nil {
		t.Fatal("config.yaml should not exist — only c3.db should be written")
	}

	// Verify entities in DB
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer s.Close()

	// c3-0 context entity
	e, err := s.GetEntity("c3-0")
	if err != nil {
		t.Fatalf("GetEntity(c3-0): %v", err)
	}
	if e.Type != "system" {
		t.Errorf("c3-0 type = %q, want %q", e.Type, "system")
	}
	if e.Title != "test-project" {
		t.Errorf("c3-0 title = %q, want %q", e.Title, "test-project")
	}

	// ADR entity
	adr, err := s.GetEntity("adr-00000000-c3-adoption")
	if err != nil {
		t.Fatalf("GetEntity(adr-00000000-c3-adoption): %v", err)
	}
	if adr.Type != "adr" {
		t.Errorf("adr type = %q, want %q", adr.Type, "adr")
	}

	// ADR -> c3-0 relationship
	rels, err := s.RelationshipsFrom("adr-00000000-c3-adoption")
	if err != nil {
		t.Fatalf("RelationshipsFrom: %v", err)
	}
	found := false
	for _, r := range rels {
		if r.ToID == "c3-0" && r.RelType == "affects" {
			found = true
		}
	}
	if !found {
		t.Error("missing adr -> c3-0 (affects) relationship")
	}
}

func TestRunInitDB_FailsIfExists(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)
	var buf bytes.Buffer

	err := RunInitDB(c3Dir, "test-project", &buf)
	if err == nil {
		t.Fatal("expected error when .c3/ already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestRunInitDB_Output(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	var buf bytes.Buffer

	if err := RunInitDB(c3Dir, "test-project", &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created .c3/") {
		t.Error("output should contain 'Created .c3/'")
	}
	if !strings.Contains(output, "c3.db") {
		t.Error("output should mention c3.db")
	}
}
