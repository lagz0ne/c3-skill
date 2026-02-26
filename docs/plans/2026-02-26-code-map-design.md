# Code Map Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace `## Code References` sections in component docs with a centralized `.c3/code-map.yaml` file, validated by `c3x check`.

**Architecture:** Remove Code References from component schema/templates. Add a new code-map parser that loads `.c3/code-map.yaml` and validates: keys are valid component IDs, paths resolve to real files. Integrate into existing `RunCheckV2`.

**Tech Stack:** Go, YAML (gopkg.in/yaml.v3 — already a dependency)

---

### Task 1: Remove Code References from component schema

**Files:**
- Modify: `cli/cmd/schema.go:39-42`

**Step 1: Write the failing test**

Update `cli/cmd/schema_test.go` — the component schema should NOT contain "Code References":

```go
func TestRunSchema_Component_NoCodeReferences(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("component", true, &buf); err != nil {
		t.Fatal(err)
	}

	var schema SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	for _, s := range schema.Sections {
		if s.Name == "Code References" {
			t.Error("component schema should not contain Code References section")
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd cli && go test ./cmd/ -run TestRunSchema_Component_NoCodeReferences -v`
Expected: FAIL — Code References still in schema

**Step 3: Remove Code References from schema registry**

In `cli/cmd/schema.go`, delete lines 39-42 (the Code References SectionDef).

**Step 4: Fix existing tests that expect Code References**

- `TestRunSchema_Component`: Remove `"Code References"` from the expected sections list
- `TestRunSchema_JSON_CodeRefColumns`: Delete this entire test — it validates Code References columns
- `TestRunCheck_EmptyRequiredTable`: Update — no longer warns about Code References
- `TestRunCheck_CodeRefFileNotExist`: Delete — Code References table no longer validated in schema
- `TestRunCheck_CodeRefFileExists`: Delete — same reason
- `TestRunCheck_EnhancedJSON`: Update assertion — no longer checks for Code References issues

**Step 5: Run all tests**

Run: `cd cli && go test ./cmd/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add cli/cmd/schema.go cli/cmd/schema_test.go cli/cmd/check_enhanced_test.go
git commit -m "feat: remove Code References from component schema"
```

---

### Task 2: Remove Code References from component template

**Files:**
- Modify: `cli/templates/component.md`

**Step 1: Remove the section**

Delete lines 79-83 from `cli/templates/component.md`:

```markdown
## Code References

<!-- List concrete code files that implement this component -->
| File | Purpose |
|------|---------|
```

**Step 2: Run template-related tests**

Run: `cd cli && go test ./... -v`
Expected: PASS (templates are embedded, not tested against schema)

**Step 3: Commit**

```bash
git add cli/templates/component.md
git commit -m "feat: remove Code References section from component template"
```

---

### Task 3: Add code-map parser

**Files:**
- Create: `cli/internal/codemap/codemap.go`
- Create: `cli/internal/codemap/codemap_test.go`

**Step 1: Write the failing test**

```go
package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCodeMap_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `# .c3/code-map.yaml
c3-101:
  - src/lib/logger.ts
  - src/lib/logger.test.ts
c3-102:
  - src/lib/config.ts
`
	path := filepath.Join(dir, "code-map.yaml")
	os.WriteFile(path, []byte(content), 0644)

	cm, err := ParseCodeMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cm) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(cm))
	}
	if len(cm["c3-101"]) != 2 {
		t.Errorf("c3-101 should have 2 files, got %d", len(cm["c3-101"]))
	}
}

func TestParseCodeMap_NotExist(t *testing.T) {
	cm, err := ParseCodeMap("/nonexistent/code-map.yaml")
	if err != nil {
		t.Fatal("missing file should not error")
	}
	if len(cm) != 0 {
		t.Errorf("missing file should return empty map, got %d entries", len(cm))
	}
}

func TestParseCodeMap_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "code-map.yaml")
	os.WriteFile(path, []byte(""), 0644)

	cm, err := ParseCodeMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cm) != 0 {
		t.Errorf("empty file should return empty map, got %d entries", len(cm))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/codemap/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement codemap parser**

```go
package codemap

import (
	"os"

	"gopkg.in/yaml.v3"
)

// CodeMap maps component IDs to their source file paths.
type CodeMap map[string][]string

// ParseCodeMap reads and parses a code-map.yaml file.
// Returns an empty map if the file does not exist.
func ParseCodeMap(path string) (CodeMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return CodeMap{}, nil
		}
		return nil, err
	}

	var cm CodeMap
	if err := yaml.Unmarshal(data, &cm); err != nil {
		return nil, err
	}
	if cm == nil {
		cm = CodeMap{}
	}
	return cm, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd cli && go test ./internal/codemap/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cli/internal/codemap/
git commit -m "feat: add code-map.yaml parser"
```

---

### Task 4: Add code-map validation to c3x check

**Files:**
- Modify: `cli/cmd/check_enhanced.go`
- Create: `cli/internal/codemap/validate.go`
- Create: `cli/internal/codemap/validate_test.go`

**Step 1: Write validation tests**

```go
package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_InvalidID(t *testing.T) {
	cm := CodeMap{"c3-999": {"src/foo.ts"}}
	knownIDs := map[string]bool{"c3-101": true}

	issues := Validate(cm, knownIDs, "")
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-999" {
			found = true
		}
	}
	if !found {
		t.Error("should flag unknown component ID c3-999")
	}
}

func TestValidate_MissingFile(t *testing.T) {
	dir := t.TempDir()
	cm := CodeMap{"c3-101": {"src/nonexistent.ts"}}
	knownIDs := map[string]bool{"c3-101": true}

	issues := Validate(cm, knownIDs, dir)
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-101" {
			found = true
		}
	}
	if !found {
		t.Error("should flag nonexistent file")
	}
}

func TestValidate_FileExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "foo.ts"), []byte(""), 0644)

	cm := CodeMap{"c3-101": {"src/foo.ts"}}
	knownIDs := map[string]bool{"c3-101": true}

	issues := Validate(cm, knownIDs, dir)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestValidate_NonComponentID(t *testing.T) {
	cm := CodeMap{"ref-jwt": {"src/foo.ts"}}
	knownIDs := map[string]bool{"ref-jwt": true}

	issues := Validate(cm, knownIDs, "")
	found := false
	for _, issue := range issues {
		if issue.Entity == "ref-jwt" {
			found = true
		}
	}
	if !found {
		t.Error("should flag non-component ID in code-map")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/codemap/ -run TestValidate -v`
Expected: FAIL

**Step 3: Implement validation**

```go
package codemap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Issue represents a validation finding from code-map.
type Issue struct {
	Severity string
	Entity   string
	Message  string
}

// Validate checks a CodeMap against known entity IDs and filesystem.
func Validate(cm CodeMap, knownIDs map[string]bool, projectDir string) []Issue {
	var issues []Issue

	for id, paths := range cm {
		// ID must exist in graph
		if !knownIDs[id] {
			issues = append(issues, Issue{
				Severity: "error",
				Entity:   id,
				Message:  fmt.Sprintf("code-map: unknown entity ID %q", id),
			})
			continue
		}

		// ID must be a component (c3-NNN pattern, not ref-*, adr-*, c3-0, or bare container c3-N)
		if !isComponentID(id) {
			issues = append(issues, Issue{
				Severity: "warning",
				Entity:   id,
				Message:  fmt.Sprintf("code-map: %q is not a component ID", id),
			})
		}

		// Validate file paths exist
		if projectDir == "" {
			continue
		}
		for _, p := range paths {
			absPath := filepath.Join(projectDir, p)
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   id,
					Message:  fmt.Sprintf("code-map: file does not exist: %s", p),
				})
			}
		}
	}

	return issues
}

// isComponentID returns true for c3-NNN+ IDs (3+ digit suffix = component).
func isComponentID(id string) bool {
	if !strings.HasPrefix(id, "c3-") {
		return false
	}
	suffix := id[3:]
	// Components have 3+ digit numeric suffix (101+)
	if len(suffix) < 3 {
		return false
	}
	for _, c := range suffix {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
```

**Step 4: Run tests**

Run: `cd cli && go test ./internal/codemap/ -v`
Expected: PASS

**Step 5: Wire code-map validation into RunCheckV2**

In `cli/cmd/check_enhanced.go`, add code-map loading and validation after the existing checks:

```go
// After existing Layer 3 checks, add:

// Layer 4: Code-map validation
codeMapPath := filepath.Join(filepath.Dir(opts.Graph.All()[0].Path), "..", "code-map.yaml")
// Better: pass c3Dir into CheckOptions, construct path from there
```

Actually — `CheckOptions` needs a `C3Dir` field (it already has `ProjectDir`). The code-map lives at `.c3/code-map.yaml`, and `ProjectDir` is the repo root. So the path is `filepath.Join(ProjectDir, ".c3", "code-map.yaml")`.

Add to `check_enhanced.go` after the bidirectional consistency block:

```go
// Code-map validation
if opts.ProjectDir != "" {
    codeMapPath := filepath.Join(opts.ProjectDir, ".c3", "code-map.yaml")
    cm, err := codemap.ParseCodeMap(codeMapPath)
    if err != nil {
        issues = append(issues, Issue{
            Severity: "error",
            Entity:   "code-map",
            Message:  fmt.Sprintf("failed to parse code-map.yaml: %v", err),
        })
    } else if len(cm) > 0 {
        knownIDs := make(map[string]bool)
        for _, e := range entities {
            knownIDs[e.ID] = true
        }
        for _, cmIssue := range codemap.Validate(cm, knownIDs, opts.ProjectDir) {
            issues = append(issues, Issue{
                Severity: cmIssue.Severity,
                Entity:   cmIssue.Entity,
                Message:  cmIssue.Message,
            })
        }
    }
}
```

**Step 6: Add integration test**

In `cli/cmd/check_enhanced_test.go`:

```go
func TestRunCheck_CodeMapInvalidID(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
`)

	// code-map with unknown ID
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), `c3-999:
  - src/foo.ts
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-999") {
		t.Errorf("should flag unknown ID in code-map, got: %s", output)
	}
}

func TestRunCheck_CodeMapMissingFile(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
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

Auth.
`)

	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), `c3-101:
  - src/auth/nonexistent.ts
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "nonexistent.ts") {
		t.Errorf("should flag missing file in code-map, got: %s", output)
	}
}

func TestRunCheck_CodeMapValid(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)
	os.MkdirAll(filepath.Join(dir, "src", "auth"), 0755)
	writeFile(t, filepath.Join(dir, "src", "auth", "jwt.ts"), "export function validate() {}")

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
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

Auth.
`)

	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), `c3-101:
  - src/auth/jwt.ts
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "code-map") {
		t.Errorf("should have no code-map issues, got: %s", output)
	}
}
```

**Step 7: Run all tests**

Run: `cd cli && go test ./... -v`
Expected: PASS

**Step 8: Commit**

```bash
git add cli/internal/codemap/ cli/cmd/check_enhanced.go cli/cmd/check_enhanced_test.go
git commit -m "feat: add code-map.yaml validation to c3x check"
```

---

### Task 5: Update skill references

**Files:**
- Modify: `skills/c3/SKILL.md`
- Modify: `skills/c3/references/onboard.md`
- Modify: `skills/c3/references/query.md`
- Modify: `skills/c3/references/audit.md`
- Modify: `skills/c3/references/change.md`
- Modify: `skills/c3/references/sweep.md`
- Modify: `skills/c3/references/ref.md`

**Step 1: Update each file**

Replace references to `## Code References` with `code-map.yaml` guidance:

- **SKILL.md**: Change "Foundation (01-09): infrastructure, `## Code References` required" to "Foundation (01-09): infrastructure, mapped in `.c3/code-map.yaml`"
- **onboard.md**: Replace Code References mentions with code-map.yaml. Foundation/Feature components get entries in code-map. Refs don't.
- **query.md**: `## Code References` → `.c3/code-map.yaml` as the place to find file paths for components
- **audit.md**: Update "The Code References Test" → "The Code Map Test". Component WITH code-map entry → implemented. Without → provisioned.
- **change.md**: Update Phase 3/4 references
- **sweep.md**: Update code discovery path
- **ref.md**: Update the component vs ref distinction

Each replacement should be mechanical — swap the section name for the file path. The semantic meaning stays the same.

**Step 2: Commit**

```bash
git add skills/c3/
git commit -m "docs: update skill references from Code References to code-map.yaml"
```

---

### Task 6: Remove bidirectional Cited By check

The `## Cited By` section on refs was manual bookkeeping. The graph already tracks `refs:` frontmatter on components and can derive cited-by relationships. Remove the bidirectional consistency check that enforces Cited By tables.

**Files:**
- Modify: `cli/cmd/check_enhanced.go:201-255`
- Modify: `cli/cmd/check_enhanced_test.go`
- Modify: `cli/templates/ref.md`

**Step 1: Remove the bidirectional check block**

Delete lines 201-255 in `check_enhanced.go` (the `Layer 3: Bidirectional consistency` block).

**Step 2: Remove Cited By from ref schema**

In `schema.go`, remove the `Cited By` SectionDef from the ref schema (lines 82-85).

**Step 3: Remove Cited By from ref template**

Delete the `## Cited By` section from `cli/templates/ref.md` (lines 97-100).

**Step 4: Fix tests**

- Delete `TestRunCheck_BidirectionalInconsistency`
- Update any tests that set up Cited By tables (they should still work, just won't be validated)

**Step 5: Run all tests**

Run: `cd cli && go test ./... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add cli/cmd/schema.go cli/cmd/check_enhanced.go cli/cmd/check_enhanced_test.go cli/templates/ref.md
git commit -m "feat: remove Cited By enforcement, graph derives citations from refs: frontmatter"
```
