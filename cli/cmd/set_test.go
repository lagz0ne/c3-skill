package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

// =============================================================================
// RunSet: unified set command for frontmatter fields and sections
// =============================================================================

func TestRunSet_FrontmatterField(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		C3Dir:   c3Dir,
		ID:      "c3-101",
		Field:   "goal",
		Value:   "Handle JWT authentication",
	}

	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm.Goal != "Handle JWT authentication" {
		t.Errorf("goal = %q, want %q", fm.Goal, "Handle JWT authentication")
	}

	output := buf.String()
	if !strings.Contains(output, "Updated") {
		t.Errorf("should print Updated message, got: %s", output)
	}
}

func TestRunSet_EntityNotFound(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{C3Dir: c3Dir, ID: "c3-999", Field: "goal", Value: "test"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunSet_UpdatesContainerGoal(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{C3Dir: c3Dir, ID: "c3-1", Field: "goal", Value: "Serve high-performance API requests"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm.Goal != "Serve high-performance API requests" {
		t.Errorf("container goal = %q", fm.Goal)
	}
}

func TestRunSet_UpdatesRefGoal(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{C3Dir: c3Dir, ID: "ref-jwt", Field: "goal", Value: "Standardize JWT token format"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "refs", "ref-jwt.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm.Goal != "Standardize JWT token format" {
		t.Errorf("ref goal = %q", fm.Goal)
	}
}

func TestRunSet_UpdatesAdrStatus(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{C3Dir: c3Dir, ID: "adr-20260226-use-go", Field: "status", Value: "accepted"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "adr", "adr-20260226-use-go.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm.Status != "accepted" {
		t.Errorf("adr status = %q, want %q", fm.Status, "accepted")
	}
}

func TestRunSet_PreservesBody(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	originalContent, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, originalBody := frontmatter.ParseFrontmatter(string(originalContent))

	opts := SetOptions{C3Dir: c3Dir, ID: "c3-101", Field: "goal", Value: "New goal"}
	RunSet(opts, &buf)

	updatedContent, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, updatedBody := frontmatter.ParseFrontmatter(string(updatedContent))

	if originalBody != updatedBody {
		t.Errorf("body changed after set.\noriginal: %q\nupdated: %q", originalBody, updatedBody)
	}
}

func TestRunSet_InvalidField(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{C3Dir: c3Dir, ID: "c3-101", Field: "nonexistent_field", Value: "test"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for unknown field name")
	}
}

// =============================================================================
// RunSet with Section target: set section content
// =============================================================================

func TestRunSet_SectionText(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		C3Dir:   c3Dir,
		ID:      "c3-101",
		Section: "Goal",
		Value:   "Provide JWT-based authentication for all API endpoints.",
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Provide JWT-based authentication for all API endpoints.") {
		t.Error("section content should be updated")
	}
}

func TestRunSet_SectionTable(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "README.md"), `---
id: c3-1
title: api
type: container
parent: c3-0
---

# api
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth

## Goal

Handle auth.

## Code References

| File | Purpose |
|------|---------|
`)

	var buf bytes.Buffer
	tableJSON := `[{"File":"src/auth/jwt.ts","Purpose":"JWT validation"},{"File":"src/auth/middleware.ts","Purpose":"Auth middleware"}]`

	opts := SetOptions{
		C3Dir:   c3Dir,
		ID:      "c3-101",
		Section: "Code References",
		Value:   tableJSON,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "jwt.ts") {
		t.Error("should contain new code reference")
	}
	if !strings.Contains(s, "middleware.ts") {
		t.Error("should contain new code reference")
	}
	if !strings.Contains(s, "Handle auth.") {
		t.Error("Goal section should survive section update")
	}
}

func TestRunSet_SectionMalformedJSON(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		C3Dir:   c3Dir,
		ID:      "c3-101",
		Section: "Code References",
		Value:   `[{"File": broken json`,
	}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestRunSet_SectionNotFound(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		C3Dir:   c3Dir,
		ID:      "c3-101",
		Section: "NonExistent Section",
		Value:   "content",
	}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for nonexistent section")
	}
}

func TestRunSet_SectionAppend(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		C3Dir:   c3Dir,
		ID:      "c3-101",
		Section: "Dependencies",
		Value:   `{"Direction":"OUT","What":"events","From/To":"c3-103"}`,
		Append:  true,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	// Original row should survive
	if !strings.Contains(s, "user credentials") {
		t.Error("existing row should be preserved")
	}
	// New row should be added
	if !strings.Contains(s, "events") {
		t.Error("new row should be appended")
	}
}
