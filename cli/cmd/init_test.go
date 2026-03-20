package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunInit_CreatesStructure(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := RunInit(dir, &buf)
	if err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	// Verify directory structure
	expected := []string{
		".c3",
		".c3/config.yaml",
		".c3/README.md",
		".c3/refs",
		".c3/rules",
		".c3/adr",
		".c3/adr/adr-00000000-c3-adoption.md",
	}

	for _, path := range expected {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", path)
		}
	}
}

func TestRunInit_ConfigYaml(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	if err := RunInit(dir, &buf); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".c3", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "# C3 configuration\n" {
		t.Errorf("unexpected config.yaml content: %q", content)
	}
}

func TestRunInit_ReadmeHasFrontmatter(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	if err := RunInit(dir, &buf); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".c3", "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(content), "---\n") {
		t.Error("README.md should start with frontmatter delimiters")
	}
	if !strings.Contains(string(content), "id: c3-0") {
		t.Error("README.md should contain c3-0 id")
	}
}

func TestRunInit_AdrHasSubstitutions(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	if err := RunInit(dir, &buf); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".c3", "adr", "adr-00000000-c3-adoption.md"))
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if strings.Contains(s, "${DATE}") {
		t.Error("ADR should have ${DATE} replaced")
	}
	if strings.Contains(s, "${PROJECT}") {
		t.Error("ADR should have ${PROJECT} replaced")
	}
	// Should contain the project dir name
	projectName := filepath.Base(dir)
	if !strings.Contains(s, projectName) {
		t.Errorf("ADR should contain project name %q", projectName)
	}
}

func TestRunInit_FailsIfExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".c3"), 0755)
	var buf bytes.Buffer

	err := RunInit(dir, &buf)
	if err == nil {
		t.Fatal("expected error when .c3/ already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestRunInit_Output(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	if err := RunInit(dir, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created .c3/") {
		t.Error("output should contain 'Created .c3/'")
	}
	if !strings.Contains(output, "config.yaml") {
		t.Error("output should mention config.yaml")
	}
	if !strings.Contains(output, "README.md") {
		t.Error("output should mention README.md")
	}
	if !strings.Contains(output, "adr-00000000-c3-adoption.md") {
		t.Error("output should mention adoption ADR")
	}
}

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

	// Verify config.yaml exists
	if _, err := os.Stat(filepath.Join(c3Dir, "config.yaml")); os.IsNotExist(err) {
		t.Fatal("config.yaml should exist")
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
	if !strings.Contains(output, "config.yaml") {
		t.Error("output should mention config.yaml")
	}
}
