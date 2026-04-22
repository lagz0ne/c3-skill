package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAdrFromDiff_GroupsFilesByComponent(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGit(t, projectDir)

	s := createRichDBFixture(t)
	if err := s.SetCodeMap("c3-101", []string{"api/**"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetCodeMap("c3-201", []string{"web/**"}); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(projectDir, "api"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "web"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("init\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, projectDir, "add", ".")
	gitCmd(t, projectDir, "commit", "-q", "-m", "init")

	// Now create uncommitted changes across two components
	os.WriteFile(filepath.Join(projectDir, "api", "auth.go"), []byte("package api\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "web", "renderer.ts"), []byte("export {}\n"), 0644)

	var buf bytes.Buffer
	err := RunAdrFromDiff(AdrFromDiffOptions{
		Store: s, C3Dir: c3Dir, ProjectDir: projectDir, Slug: "refactor-auth",
	}, &buf)
	if err != nil {
		t.Fatalf("RunAdrFromDiff: %v", err)
	}

	out := buf.String()
	// Frontmatter
	if !strings.Contains(out, "id: adr-") {
		t.Errorf("expected id prefix, got:\n%s", out)
	}
	if !strings.Contains(out, "title: Refactor Auth") {
		t.Errorf("expected humanized title, got:\n%s", out)
	}
	if !strings.Contains(out, "affects: [c3-1, c3-2]") {
		t.Errorf("expected affects with both parents, got:\n%s", out)
	}
	// Context lists touched files per component
	if !strings.Contains(out, "c3-101 (auth): api/auth.go") {
		t.Errorf("expected c3-101 entry with file, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-201 (renderer): web/renderer.ts") {
		t.Errorf("expected c3-201 entry with file, got:\n%s", out)
	}
	// Parent Delta default (#9)
	if !strings.Contains(out, "no-delta: no responsibility change") {
		t.Errorf("expected Parent Delta no-delta default, got:\n%s", out)
	}
}

func TestRunAdrFromDiff_UnmappedFiles(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGit(t, projectDir)

	s := createRichDBFixture(t)
	os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("init\n"), 0644)
	gitCmd(t, projectDir, "add", ".")
	gitCmd(t, projectDir, "commit", "-q", "-m", "init")

	os.WriteFile(filepath.Join(projectDir, "scripts.sh"), []byte("echo hi\n"), 0644)

	var buf bytes.Buffer
	if err := RunAdrFromDiff(AdrFromDiffOptions{
		Store: s, C3Dir: c3Dir, ProjectDir: projectDir, Slug: "scripts",
	}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "(unmapped): scripts.sh") {
		t.Errorf("expected unmapped file entry, got:\n%s", buf.String())
	}
}

func TestRunAdrFromDiff_NoFiles(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGit(t, projectDir)
	s := createRichDBFixture(t)

	var buf bytes.Buffer
	if err := RunAdrFromDiff(AdrFromDiffOptions{
		Store: s, C3Dir: c3Dir, ProjectDir: projectDir, Slug: "empty",
	}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No touched files detected.") {
		t.Errorf("expected no-files message, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "affects: []") {
		t.Errorf("expected empty affects, got:\n%s", buf.String())
	}
}
