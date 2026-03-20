package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

// =============================================================================
// SetFrontmatterField: update a single frontmatter field in a file
// =============================================================================

func TestSetField_StringField(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
goal: Old goal
---

# auth

## Goal

Handle authentication.
`)

	err := SetField(fp, "goal", "New JWT-based auth goal")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, body := frontmatter.ParseFrontmatter(string(content))
	if fm == nil {
		t.Fatal("frontmatter should still be parseable")
	}
	if fm.Goal != "New JWT-based auth goal" {
		t.Errorf("goal = %q, want %q", fm.Goal, "New JWT-based auth goal")
	}
	// Body must be preserved exactly
	if !strings.Contains(body, "Handle authentication.") {
		t.Error("body should be preserved")
	}
}

func TestSetField_EmptyToPopulated(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth
`)

	err := SetField(fp, "goal", "New goal")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm == nil {
		t.Fatal("frontmatter should be parseable")
	}
	if fm.Goal != "New goal" {
		t.Errorf("goal = %q, want %q", fm.Goal, "New goal")
	}
}

func TestSetField_StatusTransition(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "adr.md")
	writeFile(t, fp, `---
id: adr-20260226-auth
title: Add Auth
type: adr
status: proposed
date: "20260226"
affects: [c3-1]
---

# Add Auth

## Context

Need authentication.
`)

	err := SetField(fp, "status", "accepted")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm == nil {
		t.Fatal("frontmatter should be parseable")
	}
	if fm.Status != "accepted" {
		t.Errorf("status = %q, want %q", fm.Status, "accepted")
	}
	// Other fields must survive
	if fm.Title != "Add Auth" {
		t.Errorf("title should survive, got %q", fm.Title)
	}
	if len(fm.Affects) != 1 || fm.Affects[0] != "c3-1" {
		t.Errorf("affects should survive, got %v", fm.Affects)
	}
}

func TestSetField_PreservesBody(t *testing.T) {
	bodyContent := `
# Complex Body

## Section with special chars

Here's some code:

` + "```go" + `
func main() {
    fmt.Println("---")
}
` + "```" + `

## Another section

| Col1 | Col2 |
|------|------|
| a    | b    |

<!-- HTML comment -->
`
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, "---\nid: c3-1\ntitle: test\n---\n"+bodyContent)

	err := SetField(fp, "goal", "new goal")
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	_, body := frontmatter.ParseFrontmatter(string(content))
	if body != bodyContent {
		t.Errorf("body changed after SetField.\nwant:\n%s\ngot:\n%s", bodyContent, body)
	}
}

func TestSetField_PreservesOtherFields(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
boundary: service
summary: Auth service
---

# auth
`)

	err := SetField(fp, "goal", "Handle auth")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm.ID != "c3-101" {
		t.Errorf("id changed: %q", fm.ID)
	}
	if fm.Title != "auth" {
		t.Errorf("title changed: %q", fm.Title)
	}
	if fm.Type != "component" {
		t.Errorf("type changed: %q", fm.Type)
	}
	if fm.Category != "foundation" {
		t.Errorf("category changed: %q", fm.Category)
	}
	if fm.Parent != "c3-1" {
		t.Errorf("parent changed: %q", fm.Parent)
	}
}

func TestSetField_FileNotFound(t *testing.T) {
	err := SetField("/nonexistent/path.md", "goal", "test")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSetField_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, "Just plain markdown, no frontmatter.")

	err := SetField(fp, "goal", "test")
	if err == nil {
		t.Error("expected error for file without frontmatter")
	}
}

func TestSetField_InvalidFieldName(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-1
title: test
---

# test
`)

	err := SetField(fp, "nonexistent_field", "value")
	if err == nil {
		t.Error("expected error for unknown field name")
	}
}

// =============================================================================
// SetArrayField: update array frontmatter fields (refs, affects, scope)
// =============================================================================

func TestSetArrayField_AddToRefs(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
parent: c3-1
uses: [ref-jwt]
---

# auth
`)

	err := AddToArrayField(fp, "uses", "ref-logging")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if fm == nil {
		t.Fatal("frontmatter should be parseable")
	}
	if len(fm.Refs) != 2 {
		t.Fatalf("uses count = %d, want 2", len(fm.Refs))
	}
	if fm.Refs[0] != "ref-jwt" || fm.Refs[1] != "ref-logging" {
		t.Errorf("uses = %v, want [ref-jwt ref-logging]", fm.Refs)
	}
}

func TestSetArrayField_AddToEmptyArray(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth
`)

	err := AddToArrayField(fp, "uses", "ref-jwt")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Refs) != 1 || fm.Refs[0] != "ref-jwt" {
		t.Errorf("uses = %v, want [ref-jwt]", fm.Refs)
	}
}

func TestSetArrayField_NoDuplicates(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
parent: c3-1
uses: [ref-jwt]
---

# auth
`)

	err := AddToArrayField(fp, "uses", "ref-jwt")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Refs) != 1 {
		t.Errorf("should not duplicate, uses = %v", fm.Refs)
	}
}

func TestRemoveFromArrayField(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
parent: c3-1
uses: [ref-jwt, ref-logging]
---

# auth
`)

	err := RemoveFromArrayField(fp, "uses", "ref-jwt")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Refs) != 1 || fm.Refs[0] != "ref-logging" {
		t.Errorf("uses = %v, want [ref-logging]", fm.Refs)
	}
}

func TestRemoveFromArrayField_LastElement(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, `---
id: c3-101
title: auth
type: component
parent: c3-1
uses: [ref-jwt]
---

# auth
`)

	err := RemoveFromArrayField(fp, "uses", "ref-jwt")
	if err != nil {
		t.Fatal(err)
	}

	content, err2 := os.ReadFile(fp)
	if err2 != nil {
		t.Fatalf("ReadFile failed: %v", err2)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Refs) != 0 {
		t.Errorf("uses should be empty, got %v", fm.Refs)
	}
}

// =============================================================================
// helpers
// =============================================================================

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}

func TestSetField_AllStringFields(t *testing.T) {
	fields := map[string]struct {
		check func(fm *frontmatter.Frontmatter) string
	}{
		"summary":  {func(fm *frontmatter.Frontmatter) string { return fm.Summary }},
		"boundary": {func(fm *frontmatter.Frontmatter) string { return fm.Boundary }},
		"category": {func(fm *frontmatter.Frontmatter) string { return fm.Category }},
		"title":    {func(fm *frontmatter.Frontmatter) string { return fm.Title }},
		"date":     {func(fm *frontmatter.Frontmatter) string { return fm.Date }},
	}

	for field, tc := range fields {
		t.Run(field, func(t *testing.T) {
			tmp := t.TempDir()
			fp := filepath.Join(tmp, "test.md")
			writeFile(t, fp, "---\nid: c3-1\ntitle: test\ntype: container\n---\n\n# test\n")

			err := SetField(fp, field, "new-value")
			if err != nil {
				t.Fatal(err)
			}

			content, _ := os.ReadFile(fp)
			fm, _ := frontmatter.ParseFrontmatter(string(content))
			if fm == nil {
				t.Fatal("frontmatter should be parseable")
			}
			if tc.check(fm) != "new-value" {
				t.Errorf("%s = %q, want %q", field, tc.check(fm), "new-value")
			}
		})
	}
}

func TestArrayField_Affects(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, "---\nid: adr-1\ntitle: test\ntype: adr\naffects: [c3-0]\n---\n\n# test\n")

	err := AddToArrayField(fp, "affects", "c3-1")
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Affects) != 2 {
		t.Errorf("affects = %v, want 2 items", fm.Affects)
	}

	// Remove
	err = RemoveFromArrayField(fp, "affects", "c3-0")
	if err != nil {
		t.Fatal(err)
	}

	content, _ = os.ReadFile(fp)
	fm, _ = frontmatter.ParseFrontmatter(string(content))
	if len(fm.Affects) != 1 || fm.Affects[0] != "c3-1" {
		t.Errorf("affects after remove = %v", fm.Affects)
	}
}

func TestArrayField_Scope(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, "---\nid: ref-test\ntitle: test\nscope: [c3-1]\n---\n\n# test\n")

	err := AddToArrayField(fp, "scope", "c3-2")
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Scope) != 2 {
		t.Errorf("scope = %v, want 2 items", fm.Scope)
	}
}

func TestArrayField_Sources(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, "---\nid: rule-test\ntitle: test\ntype: rule\n---\n\n# test\n")

	err := AddToArrayField(fp, "sources", "ref-jwt")
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	if len(fm.Sources) != 1 || fm.Sources[0] != "ref-jwt" {
		t.Errorf("sources = %v", fm.Sources)
	}
}

func TestArrayField_InvalidField(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "test.md")
	writeFile(t, fp, "---\nid: c3-1\ntitle: test\n---\n\n# test\n")

	err := AddToArrayField(fp, "invalid", "value")
	if err == nil {
		t.Error("expected error for invalid array field")
	}

	err = RemoveFromArrayField(fp, "invalid", "value")
	if err == nil {
		t.Error("expected error for invalid array field on remove")
	}
}
