package marketplace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// createBareRepo creates a bare git repo with a marketplace.yaml and one rule file.
func createBareRepo(t *testing.T) string {
	t.Helper()

	work := filepath.Join(t.TempDir(), "work")
	os.MkdirAll(work, 0755)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	os.WriteFile(filepath.Join(work, "marketplace.yaml"), []byte("name: test-rules\ndescription: Test\nrules:\n  - id: rule-test-one\n    summary: First test rule\n"), 0644)
	os.WriteFile(filepath.Join(work, "rule-test-one.md"), []byte("---\nid: rule-test-one\ntype: rule\ntitle: Test Rule One\ngoal: Test\n---\n\n# Test Rule One\n\n## Rule\n\nAlways test.\n"), 0644)

	run("add", ".")
	run("commit", "-m", "init")

	bare := filepath.Join(t.TempDir(), "bare.git")
	cmd := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone --bare: %v\n%s", err, out)
	}

	return bare
}

func TestClone(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	bare := createBareRepo(t)
	dest := filepath.Join(t.TempDir(), "cloned")

	err := Clone(bare, dest)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "marketplace.yaml")); err != nil {
		t.Fatal("marketplace.yaml not found in clone")
	}
	if _, err := os.Stat(filepath.Join(dest, "rule-test-one.md")); err != nil {
		t.Fatal("rule-test-one.md not found in clone")
	}
}

func TestPull(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	bare := createBareRepo(t)
	dest := filepath.Join(t.TempDir(), "cloned")

	if err := Clone(bare, dest); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	err := Pull(dest)
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
}

func TestCloneInvalidURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dest := filepath.Join(t.TempDir(), "bad")
	err := Clone("/nonexistent/repo.git", dest)
	if err == nil {
		t.Fatal("expected error for invalid repo")
	}
}
