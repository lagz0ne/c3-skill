package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func createTestMarketplaceRepo(t *testing.T) string {
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

	os.WriteFile(filepath.Join(work, "marketplace.yaml"), []byte(`name: test-rules
description: Test coding rules
tags: [go, testing]
compatibility:
  languages: [go]
rules:
  - id: rule-test-one
    title: Test Rule One
    category: reliability
    tags: [testing]
    summary: Always write tests first
`), 0644)

	os.WriteFile(filepath.Join(work, "rule-test-one.md"), []byte("---\nid: rule-test-one\ntype: rule\ntitle: Test Rule One\ngoal: Ensure test coverage\n---\n\n# Test Rule One\n\n## Rule\n\nWrite tests before implementation.\n\n## Golden Example\n\n```go\nfunc TestFoo(t *testing.T) { ... }\n```\n"), 0644)

	run("add", ".")
	run("commit", "-m", "init")

	bare := filepath.Join(t.TempDir(), "bare.git")
	cmd := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bare: %v\n%s", err, out)
	}
	return bare
}

func TestMarketplaceAddAndList(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := filepath.Join(t.TempDir(), "marketplace")
	repo := createTestMarketplaceRepo(t)

	var buf bytes.Buffer

	// Add
	err := RunMarketplaceAdd(MarketplaceOptions{BaseDir: baseDir, URL: repo}, &buf)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// List (JSON)
	buf.Reset()
	err = RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir, JSON: true}, &buf)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rule-test-one") {
		t.Errorf("List output missing rule-test-one:\n%s", out)
	}
	if !strings.Contains(out, "test-rules") {
		t.Errorf("List output missing source name:\n%s", out)
	}
}

func TestMarketplaceShow(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := filepath.Join(t.TempDir(), "marketplace")
	repo := createTestMarketplaceRepo(t)

	var buf bytes.Buffer
	RunMarketplaceAdd(MarketplaceOptions{BaseDir: baseDir, URL: repo}, &buf)

	buf.Reset()
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir, RuleID: "rule-test-one"}, &buf)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if !strings.Contains(buf.String(), "Golden Example") {
		t.Errorf("Show output missing Golden Example:\n%s", buf.String())
	}
}

func TestMarketplaceRemove(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := filepath.Join(t.TempDir(), "marketplace")
	repo := createTestMarketplaceRepo(t)

	var buf bytes.Buffer
	RunMarketplaceAdd(MarketplaceOptions{BaseDir: baseDir, URL: repo}, &buf)

	buf.Reset()
	err := RunMarketplaceRemove(MarketplaceOptions{BaseDir: baseDir, SourceName: "test-rules"}, &buf)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	cacheDir := filepath.Join(baseDir, "test-rules")
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("cache dir should be removed")
	}
}
