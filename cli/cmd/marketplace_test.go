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

// --- Additional coverage tests (no git required) ---

func setupMarketplaceFixture(t *testing.T) string {
	t.Helper()
	baseDir := filepath.Join(t.TempDir(), "marketplace")
	os.MkdirAll(baseDir, 0755)

	sourcesYAML := `sources:
  - name: go-patterns
    url: https://github.com/org/go-patterns
    fetched: 2026-03-01T00:00:00Z
`
	os.WriteFile(filepath.Join(baseDir, "sources.yaml"), []byte(sourcesYAML), 0644)

	cacheDir := filepath.Join(baseDir, "go-patterns")
	os.MkdirAll(cacheDir, 0755)

	manifestYAML := `name: go-patterns
description: Go design patterns
tags: [go, patterns]
rules:
  - id: rule-error-wrapping
    title: Error Wrapping
    category: error-handling
    tags: [errors]
    summary: Always wrap errors with context
`
	os.WriteFile(filepath.Join(cacheDir, "marketplace.yaml"), []byte(manifestYAML), 0644)

	ruleContent := `---
id: rule-error-wrapping
title: Error Wrapping
---

# Error Wrapping

Always wrap errors.
`
	os.WriteFile(filepath.Join(cacheDir, "rule-error-wrapping.md"), []byte(ruleContent), 0644)

	return baseDir
}

func TestRunMarketplaceList_Empty(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "marketplace")
	os.MkdirAll(baseDir, 0755)

	var buf bytes.Buffer
	err := RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No marketplace sources") {
		t.Errorf("empty list should say no sources, got: %s", buf.String())
	}
}

func TestRunMarketplaceList_WithSources(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "go-patterns") {
		t.Error("should list go-patterns source")
	}
	if !strings.Contains(output, "rule-error-wrapping") {
		t.Error("should list rule-error-wrapping")
	}
}

func TestRunMarketplaceList_JSON(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"name"`) {
		t.Error("JSON output should contain name field")
	}
}

func TestRunMarketplaceList_FilterBySource(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir, SourceName: "nonexistent"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No marketplace sources") {
		t.Error("filtering by non-existent source should show no sources")
	}
}

func TestRunMarketplaceList_FilterByTag(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir, Tag: "errors"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "rule-error-wrapping") {
		t.Error("should find rule by tag")
	}
}

func TestRunMarketplaceList_FilterByTag_NoMatch(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir, Tag: "python"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "rule-error-wrapping") {
		t.Error("should not find rule with non-matching tag")
	}
}

func TestRunMarketplaceShow_FromFixture(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir, RuleID: "rule-error-wrapping"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Error Wrapping") {
		t.Error("should show rule content")
	}
}

func TestRunMarketplaceShow_NotFound(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir, RuleID: "rule-nonexistent"}, &buf)
	if err == nil {
		t.Error("expected error for nonexistent rule")
	}
}

func TestRunMarketplaceShow_NoRuleID(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir}, &buf)
	if err == nil {
		t.Error("expected error when no rule ID")
	}
}

func TestRunMarketplaceShow_FilterBySource(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir, SourceName: "go-patterns", RuleID: "rule-error-wrapping"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Error Wrapping") {
		t.Error("should show rule content when filtered by source")
	}
}

func TestRunMarketplaceShow_FilterByWrongSource(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir, SourceName: "wrong-source", RuleID: "rule-error-wrapping"}, &buf)
	if err == nil {
		t.Error("expected error when filtering by wrong source")
	}
}

func TestRunMarketplaceRemove_NoName(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceRemove(MarketplaceOptions{BaseDir: baseDir}, &buf)
	if err == nil {
		t.Error("expected error when no source name")
	}
}

func TestRunMarketplaceRemove_NonexistentSource(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceRemove(MarketplaceOptions{BaseDir: baseDir, SourceName: "nonexistent"}, &buf)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestRunMarketplaceRemove_FromFixture(t *testing.T) {
	baseDir := setupMarketplaceFixture(t)

	var buf bytes.Buffer
	err := RunMarketplaceRemove(MarketplaceOptions{BaseDir: baseDir, SourceName: "go-patterns"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Removed") {
		t.Error("should confirm removal")
	}
}

func TestRunMarketplaceUpdate_NoSources(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "marketplace")
	os.MkdirAll(baseDir, 0755)

	var buf bytes.Buffer
	err := RunMarketplaceUpdate(MarketplaceOptions{BaseDir: baseDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No marketplace sources") {
		t.Error("should say no sources")
	}
}

func TestRunMarketplaceAdd_NoURL(t *testing.T) {
	var buf bytes.Buffer
	err := RunMarketplaceAdd(MarketplaceOptions{BaseDir: t.TempDir()}, &buf)
	if err == nil {
		t.Error("expected error when no URL")
	}
}

func TestContainsTag(t *testing.T) {
	if !containsTag([]string{"Go", "Patterns"}, "go") {
		t.Error("should match case-insensitively")
	}
	if containsTag([]string{"Go"}, "python") {
		t.Error("should not match")
	}
	if containsTag(nil, "go") {
		t.Error("nil tags should not match")
	}
	if !containsTag([]string{"errors"}, "errors") {
		t.Error("exact match should work")
	}
}

func TestResolveBaseDir(t *testing.T) {
	got := resolveBaseDir("/custom/path")
	if got != "/custom/path" {
		t.Errorf("resolveBaseDir with override = %q", got)
	}

	got = resolveBaseDir("")
	if got == "" {
		t.Error("resolveBaseDir without override should return default")
	}
}
