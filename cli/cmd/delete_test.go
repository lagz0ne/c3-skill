package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

func TestRunDelete_Component(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	// Add codemap entry
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/**\nc3-110:\n  - src/users/**\n")

	var buf bytes.Buffer
	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "c3-101", Graph: graph}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// File should be gone
	if _, err := os.Stat(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")); !os.IsNotExist(err) {
		t.Error("entity file should be deleted")
	}

	// Parent container's Components table should not have c3-101
	containerContent, _ := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "README.md"))
	if strings.Contains(string(containerContent), "c3-101") {
		t.Error("parent container Components table should not contain c3-101")
	}

	// Codemap should not have c3-101
	cmContent, _ := os.ReadFile(filepath.Join(c3Dir, "code-map.yaml"))
	if strings.Contains(string(cmContent), "c3-101") {
		t.Error("code-map.yaml should not contain c3-101")
	}
	// c3-110 should still be in codemap
	if !strings.Contains(string(cmContent), "c3-110") {
		t.Error("code-map.yaml should still contain c3-110")
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted c3-101") {
		t.Errorf("should print Deleted message, got: %s", output)
	}
}

func TestRunDelete_Ref(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "ref-jwt", Graph: graph}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// File should be gone
	if _, err := os.Stat(filepath.Join(c3Dir, "refs", "ref-jwt.md")); !os.IsNotExist(err) {
		t.Error("ref file should be deleted")
	}

	// c3-101 should no longer have ref-jwt in uses[]
	content, _ := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if containsStr2(fm.Refs, "ref-jwt") {
		t.Errorf("c3-101 uses should not contain ref-jwt after delete, got %v", fm.Refs)
	}

	// Related Refs table should not contain ref-jwt
	if strings.Contains(string(content), "ref-jwt") {
		t.Error("c3-101 Related Refs table should not contain ref-jwt")
	}
}

func TestRunDelete_ContainerWithChildren(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "c3-1", Graph: graph}, &buf)
	if err == nil {
		t.Fatal("expected error for container with children")
	}
	if !strings.Contains(err.Error(), "children") {
		t.Errorf("error should mention children: %v", err)
	}
	if !strings.Contains(err.Error(), "c3-101") {
		t.Errorf("error should list child IDs: %v", err)
	}
}

func TestRunDelete_ContextRoot(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "c3-0", Graph: graph}, &buf)
	if err == nil {
		t.Fatal("expected error for c3-0")
	}
	if !strings.Contains(err.Error(), "c3-0") {
		t.Errorf("error should mention c3-0: %v", err)
	}
}

func TestRunDelete_NotFound(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "c3-999", Graph: graph}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunDelete_DryRun(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "c3-101", Graph: graph, DryRun: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// File should still exist
	if _, err := os.Stat(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")); os.IsNotExist(err) {
		t.Error("entity file should NOT be deleted in dry-run mode")
	}

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("output should contain [dry-run] prefix, got: %s", output)
	}
	if !strings.Contains(output, "no files modified") {
		t.Errorf("output should mention 'no files modified', got: %s", output)
	}
}

func TestRunDelete_RuleCleanup(t *testing.T) {
	c3Dir := createRichFixture(t)

	// Add a rule entity
	os.MkdirAll(filepath.Join(c3Dir, "rules"), 0755)
	writeFile(t, filepath.Join(c3Dir, "rules", "rule-logging.md"), `---
id: rule-logging
type: rule
title: Logging Standard
origin: [ref-jwt]
---

# Logging Standard

## Goal

Standardize logging.

## Rule

Use structured JSON logging.

## Golden Example

`+"`"+`go
logger.Info("request", "method", r.Method)
`+"`"+`
`)

	// Update c3-101 to reference rule-logging and have a Related Rules table
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
uses: [ref-jwt, rule-logging]
---

# auth

## Goal

Handle authentication.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | user credentials | c3-110 |

## Related Refs

| Ref | Role |
|-----|------|
| ref-jwt | Token format |

## Related Rules

| Rule | Summary |
|------|---------|
| rule-logging | Logging standard |
`)

	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "rule-logging", Graph: graph}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Rule file should be gone
	if _, err := os.Stat(filepath.Join(c3Dir, "rules", "rule-logging.md")); !os.IsNotExist(err) {
		t.Error("rule file should be deleted")
	}

	// c3-101 should no longer have rule-logging in uses[]
	content, _ := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if containsStr2(fm.Refs, "rule-logging") {
		t.Errorf("c3-101 uses should not contain rule-logging after delete, got %v", fm.Refs)
	}

	// Related Rules table should not contain rule-logging
	if strings.Contains(string(content), "rule-logging") {
		t.Error("c3-101 Related Rules table should not contain rule-logging")
	}

	// ref-jwt should still be in uses
	if !containsStr2(fm.Refs, "ref-jwt") {
		t.Errorf("c3-101 uses should still contain ref-jwt, got %v", fm.Refs)
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted rule-logging") {
		t.Errorf("should print Deleted message, got: %s", output)
	}
}

func TestRunDelete_SourcesCleanup(t *testing.T) {
	c3Dir := createRichFixture(t)

	// Add a recipe with sources referencing c3-101 (including anchored ref)
	os.MkdirAll(filepath.Join(c3Dir, "recipes"), 0755)
	writeFile(t, filepath.Join(c3Dir, "recipes", "recipe-auth-flow.md"), `---
id: recipe-auth-flow
title: Auth Flow
type: recipe
sources: [c3-101#login, c3-101#logout, c3-110]
---

# Auth Flow

## Goal

Trace auth flow.
`)

	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{C3Dir: c3Dir, ID: "c3-101", Graph: graph}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Recipe sources should no longer reference c3-101
	content, _ := os.ReadFile(filepath.Join(c3Dir, "recipes", "recipe-auth-flow.md"))
	fm, _ := frontmatter.ParseFrontmatter(string(content))

	for _, src := range fm.Sources {
		if frontmatter.StripAnchor(src) == "c3-101" {
			t.Errorf("recipe sources should not reference c3-101 after delete, got %v", fm.Sources)
			break
		}
	}

	// c3-110 should still be in sources
	if !containsStr2(fm.Sources, "c3-110") {
		t.Errorf("recipe sources should still contain c3-110, got %v", fm.Sources)
	}
}
