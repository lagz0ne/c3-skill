package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestResolveTouchedTargets_SourceEdit(t *testing.T) {
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

	srcPath := filepath.Join(projectDir, "api", "auth.go")
	if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPath, []byte("package api\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, projectDir, "add", ".")
	gitCmd(t, projectDir, "commit", "-q", "-m", "init")
	if err := os.WriteFile(srcPath, []byte("package api\n// edit\n"), 0644); err != nil {
		t.Fatal(err)
	}

	targets, err := resolveTouchedTargetsWithStore(projectDir, c3Dir, "", s)
	if err != nil {
		t.Fatalf("resolveTouchedTargets: %v", err)
	}
	if !containsStr2(targets, "c3-101") {
		t.Errorf("expected c3-101 (codemap mapped from api/auth.go), got %v", targets)
	}
}

func TestResolveTouchedTargets_CanonicalEdit(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755); err != nil {
		t.Fatal(err)
	}
	initGit(t, projectDir)

	s := createRichDBFixture(t)
	authPath := filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")
	initial := "---\nid: c3-101\ntitle: auth\n---\n\n# auth\n"
	if err := os.WriteFile(authPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, projectDir, "add", ".")
	gitCmd(t, projectDir, "commit", "-q", "-m", "init")
	if err := os.WriteFile(authPath, []byte(initial+"\n## Goal\n\nNew goal\n"), 0644); err != nil {
		t.Fatal(err)
	}

	targets, err := resolveTouchedTargetsWithStore(projectDir, c3Dir, "", s)
	if err != nil {
		t.Fatalf("resolveTouchedTargets: %v", err)
	}
	sort.Strings(targets)
	if !containsStr2(targets, "c3-101") {
		t.Errorf("expected c3-101 from canonical edit, got %v", targets)
	}
}

func initGit(t *testing.T, dir string) {
	t.Helper()
	gitCmd(t, dir, "init", "-q")
	gitCmd(t, dir, "config", "user.email", "test@test")
	gitCmd(t, dir, "config", "user.name", "test")
	gitCmd(t, dir, "config", "commit.gpgsign", "false")
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}
