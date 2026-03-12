package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Layer 2: Schema validation — check validates decorated content
// =============================================================================

func TestRunCheck_EmptyRequiredSection(t *testing.T) {
	c3Dir := createRichFixture(t)

	// Make c3-110's Goal section empty
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-110-users.md"), `---
id: c3-110
title: users
type: component
category: feature
parent: c3-1
---

# users

## Goal

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-110") {
		t.Errorf("should flag c3-110, got: %s", output)
	}
}

func TestRunCheck_DefaultSkipsADR(t *testing.T) {
	c3Dir := createRichFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	for _, issue := range result.Issues {
		if strings.Contains(issue.Entity, "adr-") {
			t.Errorf("default check should skip ADR validation, but found issue for: %s", issue.Entity)
		}
	}
}

func TestRunCheck_IncludeADRValidatesADR(t *testing.T) {
	c3Dir := createRichFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: true, IncludeADR: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	// The fixture ADR has "Context" section but schema requires "Goal" — should produce a warning
	hasADRIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Entity, "adr-") {
			hasADRIssue = true
			break
		}
	}
	if !hasADRIssue {
		t.Error("--include-adr should validate ADR entities, expected issues for adr-20260226-use-go")
	}
}

func TestRunCheck_EmptyRequiredTable(t *testing.T) {
	c3Dir := createRichFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// c3-110 and c3-201 have empty Dependencies tables — should warn about them
	if !strings.Contains(output, "empty required table") {
		t.Errorf("should warn about empty required tables, got: %s", output)
	}
	// Code References no longer in schema — should NOT appear
	if strings.Contains(output, "Code References") {
		t.Errorf("should not warn about Code References (removed from schema), got: %s", output)
	}
}

func TestRunCheck_MissingRequiredSection_Ref(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "refs"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
`)

	// Ref missing required "Choice" and "Why" sections
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-incomplete.md"), `---
id: ref-incomplete
title: Incomplete Ref
goal: Some pattern
---

# Incomplete Ref

## Goal

Some pattern.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref-incomplete") {
		t.Errorf("should flag ref-incomplete, got: %s", output)
	}
	if !strings.Contains(output, "Choice") {
		t.Errorf("should mention missing Choice section, got: %s", output)
	}
}

// =============================================================================
// Layer 3: Typed content validation
// =============================================================================

func TestRunCheck_EntityIdNotInGraph(t *testing.T) {
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

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-999 |
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-999") {
		t.Errorf("should flag nonexistent entity in Dependencies, got: %s", output)
	}
}

func TestRunCheck_RefIdNotInGraph(t *testing.T) {
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

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-1 |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-nonexistent | Token format |
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref-nonexistent") {
		t.Errorf("should flag nonexistent ref in Related Refs, got: %s", output)
	}
	if !strings.Contains(output, "unknown ref reference") {
		t.Errorf("should use 'unknown ref reference' message, got: %s", output)
	}
}

func TestRunCheck_SuggestsByTitle(t *testing.T) {
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

	// Component references "api" by title instead of "c3-1" by ID
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth

## Goal

Auth.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | api |
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "did you mean c3-1?") {
		t.Errorf("should suggest c3-1 for 'api', got: %s", output)
	}
}

func TestRunCheck_FixReplacesByTitle(t *testing.T) {
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

	compPath := filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")
	writeFile(t, compPath, `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth

## Goal

Auth.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | api |
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, Fix: true, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "fixed 1 reference") {
		t.Errorf("should report 1 fix, got: %s", output)
	}

	// Verify the file was actually modified
	data, err := os.ReadFile(compPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "| c3-1 |") {
		t.Errorf("file should contain '| c3-1 |' after fix, got: %s", content)
	}
	if strings.Contains(content, "| api |") {
		t.Errorf("file should NOT contain '| api |' after fix, got: %s", content)
	}
}

func TestRunCheck_EnhancedJSON(t *testing.T) {
	c3Dir := createRichFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	// Should have issues from schema validation (empty tables, etc.)
	hasSchemaIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "empty") {
			hasSchemaIssue = true
			break
		}
	}
	if !hasSchemaIssue {
		t.Error("JSON output should include schema validation issues")
	}
	// Code References removed from schema — should not appear
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "Code References") {
			t.Errorf("should not have Code References issues (removed from schema), got: %s", issue.Message)
		}
	}
}

// =============================================================================
// Code-map integration tests
// =============================================================================

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

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir, C3Dir: c3Dir}
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

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "nonexistent.ts") {
		t.Errorf("should flag missing file in code-map, got: %s", output)
	}
}

// =============================================================================
// RED: Codex-identified issue — C3Dir should be distinct from ProjectDir
// =============================================================================

func TestRunCheck_CodeMapCustomC3Dir(t *testing.T) {
	// Bug: RunCheckV2 hardcodes code-map path as ProjectDir/.c3/code-map.yaml.
	// When .c3/ is at a custom location (--c3-dir), code-map is NOT found.
	//
	// Setup: ProjectDir has NO .c3/ subdirectory. The .c3/ docs live elsewhere.
	// code-map.yaml with a valid entry exists in the custom c3 dir.
	// Because RunCheckV2 looks at ProjectDir/.c3/code-map.yaml, it won't find it.

	projectDir := t.TempDir()
	customC3Dir := t.TempDir() // separate, not under projectDir

	// Create source file in project dir
	os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0755)
	writeFile(t, filepath.Join(projectDir, "src", "auth", "jwt.ts"), "export function validate() {}")

	// Create C3 docs in the custom c3 dir
	os.MkdirAll(filepath.Join(customC3Dir, "c3-1-api"), 0755)

	writeFile(t, filepath.Join(customC3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
`)

	writeFile(t, filepath.Join(customC3Dir, "c3-1-api", "README.md"), `---
id: c3-1
title: api
type: container
parent: c3-0
---

# api
`)

	writeFile(t, filepath.Join(customC3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth

## Goal

Auth.
`)

	// code-map.yaml lives inside the custom C3 dir (NOT under projectDir/.c3/)
	writeFile(t, filepath.Join(customC3Dir, "code-map.yaml"), `c3-101:
  - src/auth/jwt.ts
`)

	docs := loadDocs(t, customC3Dir)
	graph := loadGraph(t, customC3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{
		Graph:      graph,
		Docs:       docs,
		JSON:       true,
		ProjectDir: projectDir,
		C3Dir:      customC3Dir,
	}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	// The code-map at customC3Dir/code-map.yaml has c3-101 mapped to src/auth/jwt.ts.
	// If CheckOptions supported C3Dir, it would find and validate this code-map.
	// We expect to see zero code-map errors (file exists, ID is valid).
	//
	// BUG: The current code looks at projectDir/.c3/code-map.yaml which doesn't
	// exist, so it silently skips code-map validation entirely. This test asserts
	// that code-map validation DID run (i.e., CheckOptions should accept a C3Dir).
	//
	// To detect whether code-map validation ran, we intentionally put a BOGUS
	// extra entry in the code-map that should produce an error.
	// But with the simpler approach: we verify the total issue count reflects
	// that the engine processed code-map entries.

	// Actually, the simplest RED assertion: CheckOptions SHOULD have a C3Dir field.
	// Since it doesn't, we test that the struct is missing the field by checking
	// that code-map validation was skipped (no code-map issues at all when there
	// should be some).

	// Add a second code-map with an UNKNOWN ID — this should trigger an error
	// if code-map validation ran.
	writeFile(t, filepath.Join(customC3Dir, "code-map.yaml"), `c3-101:
  - src/auth/jwt.ts
c3-999:
  - src/unknown.ts
`)

	// Re-run
	buf.Reset()
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	// If code-map validation ran from the custom C3Dir, c3-999 would be flagged.
	foundCodeMapIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "c3-999") {
			foundCodeMapIssue = true
		}
	}
	if !foundCodeMapIssue {
		t.Error("code-map validation should run from custom C3Dir and flag unknown ID c3-999; " +
			"CheckOptions needs a C3Dir field so RunCheckV2 resolves code-map.yaml from the correct directory")
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

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "code-map") {
		t.Errorf("should have no code-map issues, got: %s", output)
	}
}

// =============================================================================
// Output quality: summary, hints, legend
// =============================================================================

func TestRunCheck_CleanOutputSummary(t *testing.T) {
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

## Containers

| ID | Name | Purpose |
|----|------|---------|
|  | core | Core |

## Abstract Constraints

Keep it simple.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "all clear") {
		t.Errorf("clean run should say 'all clear', got: %s", output)
	}
	if !strings.Contains(output, "Checked") {
		t.Errorf("clean run should show 'Checked N docs', got: %s", output)
	}
}

func TestRunCheck_IssuesSummaryAndLegend(t *testing.T) {
	c3Dir := createRichFixture(t)

	// Make c3-110's Goal section empty to guarantee a warning
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-110-users.md"), `---
id: c3-110
title: users
type: component
category: feature
parent: c3-1
---

# users

## Goal

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Checked") {
		t.Errorf("should have summary header, got: %s", output)
	}
	if !strings.Contains(output, "warning") {
		t.Errorf("summary should mention warnings, got: %s", output)
	}
	if !strings.Contains(output, "Legend:") {
		t.Errorf("should have legend footer, got: %s", output)
	}
}

func TestRunCheck_HintsInTextOutput(t *testing.T) {
	c3Dir := createRichFixture(t)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-110-users.md"), `---
id: c3-110
title: users
type: component
category: feature
parent: c3-1
---

# users

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// Missing Goal should get a hint
	if !strings.Contains(output, "→") {
		t.Errorf("should have hint lines with →, got: %s", output)
	}
	if !strings.Contains(output, "add a ## Goal section") {
		t.Errorf("should have specific hint for missing Goal, got: %s", output)
	}
}

func TestRunCheck_HintsInJSONOutput(t *testing.T) {
	c3Dir := createRichFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	hasHint := false
	for _, issue := range result.Issues {
		if issue.Hint != "" {
			hasHint = true
			break
		}
	}
	if !hasHint {
		t.Error("JSON output should include hint fields on issues")
	}
}

func TestHintFor(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"broken YAML frontmatter: file has --- delimiters but failed to parse", "check for unquoted colons in values"},
		{"missing required section: Goal", "add a ## Goal section with content"},
		{"empty required section: Overview", "add content to the ## Overview section"},
		{"empty required table: Dependencies (headers only, no data rows)", "add at least one data row below the table headers"},
		{"unknown entity reference: c3-999", "verify the ID with 'c3x list'; check for typos"},
		{"unknown ref reference: ref-missing", "use a ref-* ID (e.g., ref-jwt); verify with 'c3x list'"},
		{"file does not exist: src/foo.ts", "create the file or fix the path"},
		{"code-map parse error: yaml: unmarshal error", "fix YAML syntax in .c3/code-map.yaml"},
		{"something unknown", ""},
	}
	for _, tt := range tests {
		got := hintFor(tt.message)
		if got != tt.expected {
			t.Errorf("hintFor(%q) = %q, want %q", tt.message, got, tt.expected)
		}
	}
}

// =============================================================================
// Note health validation
// =============================================================================

func TestRunCheck_NoteOrphanedSource(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)
	os.MkdirAll(filepath.Join(c3Dir, "_index", "notes"), 0755)

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

	// Note with a valid source (c3-1) and an orphaned source (c3-999)
	writeFile(t, filepath.Join(c3Dir, "_index", "notes", "auth-flow.md"), `---
topic: authentication-flow
sources:
  - c3-1#Goal
  - c3-999#Overview
source_hash: sha256:abc123
---

# Authentication Flow

Some content about auth.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-999") {
		t.Errorf("should flag orphaned source c3-999, got: %s", output)
	}
	if strings.Contains(output, "c3-1") && strings.Contains(output, "nonexistent entity: c3-1") {
		t.Errorf("should NOT flag valid source c3-1, got: %s", output)
	}
}

func TestRunCheck_NoteNoNotesDir(t *testing.T) {
	// When _index/notes/ doesn't exist, check should pass without errors
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

## Containers

| ID | Name | Purpose |
|----|------|---------|
|  | core | Core |

## Abstract Constraints

Keep it simple.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "all clear") {
		t.Errorf("should be all clear when no notes dir exists, got: %s", output)
	}
}

func TestRunCheck_NoteNoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "_index", "notes"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.

## Containers

| ID | Name | Purpose |
|----|------|---------|
|  | core | Core |

## Abstract Constraints

Keep it simple.
`)

	// Note without frontmatter — should be skipped
	writeFile(t, filepath.Join(c3Dir, "_index", "notes", "plain.md"), `# Just a plain note

No frontmatter here.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "all clear") {
		t.Errorf("note without frontmatter should be skipped, got: %s", output)
	}
}

func TestRunCheck_NoteSourceWithoutAnchor(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "_index", "notes"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
`)

	// Source without # anchor — use whole string as entity ID
	writeFile(t, filepath.Join(c3Dir, "_index", "notes", "overview.md"), `---
topic: overview
sources:
  - c3-0
  - nonexistent-entity
---

# Overview
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "nonexistent-entity") {
		t.Errorf("should flag nonexistent-entity, got: %s", output)
	}
	if strings.Contains(output, "nonexistent entity: c3-0") {
		t.Errorf("should NOT flag valid c3-0, got: %s", output)
	}
}

func TestRunCheck_NoteOrphanedJSON(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "_index", "notes"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
`)

	writeFile(t, filepath.Join(c3Dir, "_index", "notes", "test.md"), `---
topic: test
sources:
  - c3-gone#Details
---

# Test Note
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: true, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	found := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "c3-gone") {
			found = true
			if issue.Hint == "" {
				t.Error("note orphan issue should have a hint in JSON output")
			}
			if issue.Severity != "warning" {
				t.Errorf("note orphan should be warning, got: %s", issue.Severity)
			}
		}
	}
	if !found {
		t.Error("JSON output should include orphaned note issue for c3-gone")
	}
}

// =============================================================================
// Recipe source validation
// =============================================================================

func TestRunCheck_RecipeValidSources(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)
	os.MkdirAll(filepath.Join(c3Dir, "recipes"), 0755)

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

	writeFile(t, filepath.Join(c3Dir, "recipes", "recipe-auth.md"), `---
id: recipe-auth
title: Auth Flow
description: End-to-end auth trace
sources:
  - c3-1#Goal
  - c3-0
---

# Auth Flow

## Goal

Trace auth end-to-end.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "recipe references nonexistent") {
		t.Errorf("should NOT flag valid recipe sources, got: %s", output)
	}
}

func TestRunCheck_RecipeInvalidSources(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "recipes"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---

# Test

## Goal

Test.
`)

	writeFile(t, filepath.Join(c3Dir, "recipes", "recipe-auth.md"), `---
id: recipe-auth
title: Auth Flow
description: End-to-end auth trace
sources:
  - c3-0
  - c3-999#Overview
---

# Auth Flow

## Goal

Trace auth end-to-end.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-999") {
		t.Errorf("should flag nonexistent entity c3-999, got: %s", output)
	}
	if !strings.Contains(output, "recipe references nonexistent entity") {
		t.Errorf("should have recipe-specific message, got: %s", output)
	}
}

func TestRunCheck_NoteEndingAtEOF(t *testing.T) {
	// Test fix for #8: frontmatter ending with --- at EOF (no trailing newline)
	content := "---\ntopic: test\nsources:\n  - c3-1#Goal\n---"
	sources := parseNoteSources(content)
	if len(sources) != 1 {
		t.Fatalf("expected 1 source from EOF-terminated frontmatter, got %d: %v", len(sources), sources)
	}
	if sources[0] != "c3-1#Goal" {
		t.Errorf("expected c3-1#Goal, got %s", sources[0])
	}
}

func TestParseNoteSources(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "list style",
			content: "---\ntopic: test\nsources:\n  - c3-101#Goal\n  - ref-jwt#Choice\n---\n",
			want:    []string{"c3-101#Goal", "ref-jwt#Choice"},
		},
		{
			name:    "inline style",
			content: "---\ntopic: test\nsources: [c3-1, c3-2]\n---\n",
			want:    []string{"c3-1", "c3-2"},
		},
		{
			name:    "no frontmatter",
			content: "# Just text\n",
			want:    nil,
		},
		{
			name:    "no sources key",
			content: "---\ntopic: test\n---\n",
			want:    nil,
		},
		{
			name:    "empty sources",
			content: "---\ntopic: test\nsources:\nstatus: current\n---\n",
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNoteSources(tt.content)
			if len(got) != len(tt.want) {
				t.Fatalf("parseNoteSources() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseNoteSources()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// =============================================================================
// Scope cross-check: ref scopes container but child doesn't cite
// =============================================================================

func TestRunCheck_ScopeCrossCheck(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)
	os.MkdirAll(filepath.Join(c3Dir, "refs"), 0755)

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

## Goal

Serve API.

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-101 | auth | foundation | active | Auth |
| c3-110 | users | feature | active | Users |

## Responsibilities

Handle API.
`)

	// c3-101 cites ref-jwt
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
refs: [ref-jwt]
---

# auth

## Goal

Auth.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-110 |
`)

	// c3-110 does NOT cite ref-jwt
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-110-users.md"), `---
id: c3-110
title: users
type: component
category: feature
parent: c3-1
---

# users

## Goal

Users.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-101 |
`)

	// ref-jwt scopes c3-1 (the container)
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
title: JWT Auth
goal: Auth tokens
scope: [c3-1]
---

# JWT Auth

## Goal

Auth tokens.

## Choice

RS256.

## Why

Standard.
`)

	// code-map needed for index build
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), `c3-101:
  - src/auth/**
c3-110:
  - src/users/**
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// c3-110 should be warned: ref-jwt scopes c3-1 but c3-110 doesn't cite it
	if !strings.Contains(output, "ref-jwt scopes c3-1 but c3-110 does not cite it") {
		t.Errorf("should warn about c3-110 not citing ref-jwt, got: %s", output)
	}
	// c3-101 DOES cite ref-jwt — should NOT be warned
	if strings.Contains(output, "c3-101 does not cite it") {
		t.Errorf("should NOT warn about c3-101, got: %s", output)
	}
}

func TestRunCheck_ScopeCrossCheck_AllCited(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)
	os.MkdirAll(filepath.Join(c3Dir, "refs"), 0755)

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

## Goal

Serve API.

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-101 | auth | foundation | active | Auth |

## Responsibilities

Handle API.
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
refs: [ref-jwt]
---

# auth

## Goal

Auth.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-1 |
`)

	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
title: JWT Auth
goal: Auth tokens
scope: [c3-1]
---

# JWT Auth

## Goal

Auth tokens.

## Choice

RS256.

## Why

Standard.
`)

	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), `c3-101:
  - src/auth/**
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir, C3Dir: c3Dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "does not cite it") {
		t.Errorf("all components cite the ref, should have no scope warnings, got: %s", output)
	}
}
