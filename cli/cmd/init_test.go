package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
