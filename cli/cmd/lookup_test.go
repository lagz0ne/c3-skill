package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createLookupFixture builds a fixture with goal/summary frontmatter and a code-map.yaml.
// Returns (c3Dir, projectDir).
func createLookupFixture(t *testing.T) (string, string) {
	t.Helper()
	c3Dir := createFixture(t)
	projectDir := filepath.Dir(c3Dir)

	// Patch c3-101 to have goal + summary in frontmatter
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
goal: Handle authentication and JWT issuance
summary: JWT-based auth with Redis session store
refs: [ref-jwt]
---

# auth

## Goal

Handle authentication and JWT issuance.
`)

	return c3Dir, projectDir
}

func TestRunLookup_ExactMatch(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/login.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/auth/login.ts",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101, got:\n%s", out)
	}
	if !strings.Contains(out, "Handle authentication") {
		t.Errorf("expected goal in output, got:\n%s", out)
	}
	if !strings.Contains(out, "JWT-based auth") {
		t.Errorf("expected summary in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ref-jwt") {
		t.Errorf("expected ref-jwt listed, got:\n%s", out)
	}
	if !strings.Contains(out, "Standardize auth tokens") {
		t.Errorf("expected ref goal in output, got:\n%s", out)
	}
}

func TestRunLookup_GlobStar(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/auth/login.ts",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "c3-101") {
		t.Errorf("*.ts glob should match login.ts, got:\n%s", buf.String())
	}
}

func TestRunLookup_DoubleStar(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/**/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/auth/handlers/login.ts",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "c3-101") {
		t.Errorf("** glob should match nested file, got:\n%s", buf.String())
	}
}

func TestRunLookup_NoMatch(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/payments/stripe.go",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("expected no-match message, got:\n%s", buf.String())
	}
}

func TestRunLookup_NoCodeMap(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// No code-map.yaml written
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/auth/login.ts",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("missing code-map should produce no-match, got:\n%s", buf.String())
	}
}

func TestRunLookup_MultipleComponents(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// Both c3-101 and c3-110 claim src/auth/**
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/auth/**/*.ts\nc3-110:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/auth/login.ts",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in multi-match output, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-110") {
		t.Errorf("expected c3-110 in multi-match output, got:\n%s", out)
	}
}

func TestRunLookup_JSON(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/auth/login.ts",
		JSON:     true,
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result LookupResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if result.File != "src/auth/login.ts" {
		t.Errorf("expected file field, got %q", result.File)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Matches))
	}

	m := result.Matches[0]
	if m.ID != "c3-101" {
		t.Errorf("expected id c3-101, got %q", m.ID)
	}
	if m.Goal == "" {
		t.Error("expected non-empty goal")
	}
	if len(m.Refs) != 1 || m.Refs[0].ID != "ref-jwt" {
		t.Errorf("expected ref-jwt, got %v", m.Refs)
	}
	if m.Refs[0].Goal == "" {
		t.Error("expected non-empty ref goal")
	}
}

// --- Malformed / partial docs ---

func TestRunLookup_MissingGoal(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// Overwrite c3-101 with no goal frontmatter
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
refs: [ref-jwt]
---
# auth
`)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Graph: graph, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("component should still appear without goal, got:\n%s", out)
	}
	if strings.Contains(out, "goal:") {
		t.Errorf("no goal line expected when frontmatter goal is missing, got:\n%s", out)
	}
}

func TestRunLookup_MissingSummary(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// Overwrite c3-101 with goal but no summary
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
goal: Handle authentication
refs: [ref-jwt]
---
# auth
`)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Graph: graph, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "goal:") {
		t.Errorf("goal should appear, got:\n%s", out)
	}
	if strings.Contains(out, "summary:") {
		t.Errorf("no summary line expected when frontmatter summary is missing, got:\n%s", out)
	}
}

func TestRunLookup_NoRefs(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// Overwrite c3-101 with no refs field
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
goal: Handle authentication
summary: JWT auth
---
# auth
`)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Graph: graph, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("component should appear without refs, got:\n%s", out)
	}
	if strings.Contains(out, "refs:") {
		t.Errorf("no refs section expected when component has no refs, got:\n%s", out)
	}
}

func TestRunLookup_RefMissingGoal(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// Overwrite ref-jwt with no goal field
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
title: JWT Authentication
type: ref
---
# JWT Authentication
`)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Graph: graph, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "ref-jwt") {
		t.Errorf("ref should appear even without goal, got:\n%s", out)
	}
	// Should not have a trailing colon with empty value
	if strings.Contains(out, "ref-jwt: \n") || strings.Contains(out, "ref-jwt:  ") {
		t.Errorf("ref with no goal should not print empty colon, got:\n%s", out)
	}
}

func TestRunLookup_RefNotInGraph(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// Component refs a non-existent ref
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
goal: Handle authentication
refs: [ref-does-not-exist]
---
# auth
`)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Graph: graph, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("component should appear even when its ref is missing from graph, got:\n%s", out)
	}
	if strings.Contains(out, "ref-does-not-exist") {
		t.Errorf("missing ref should be silently skipped, got:\n%s", out)
	}
}

func TestRunLookup_ComponentNotInGraph(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	// code-map points to an ID that has no .md doc
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-999:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Graph: graph, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	// c3-999 matched in code-map but missing from graph → treated as no match
	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("component missing from graph should produce no-match, got:\n%s", buf.String())
	}
}

func TestRunLookup_GlobInput_MultipleFiles(t *testing.T) {
	c3Dir, projectDir := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")

	// Create actual source files for glob expansion
	if err := os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(projectDir, "src", "auth", "jwt.ts"), "// jwt")
	writeFile(t, filepath.Join(projectDir, "src", "auth", "middleware.ts"), "// middleware")

	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:      graph,
		FilePath:   "src/auth/*.ts",
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "pattern: src/auth/*.ts") {
		t.Errorf("expected pattern header, got:\n%s", out)
	}
	if !strings.Contains(out, "2 file(s) matched") {
		t.Errorf("expected 2 files matched, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in components, got:\n%s", out)
	}
	if !strings.Contains(out, "jwt.ts") || !strings.Contains(out, "middleware.ts") {
		t.Errorf("expected both files in file map, got:\n%s", out)
	}
}

func TestRunLookup_GlobInput_NoFilesMatched(t *testing.T) {
	c3Dir, projectDir := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	// No source files created — glob finds nothing
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:      graph,
		FilePath:   "src/auth/*.ts",
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no files matched") {
		t.Errorf("expected no-files-matched message, got:\n%s", buf.String())
	}
}

func TestRunLookup_GlobInput_JSON(t *testing.T) {
	c3Dir, projectDir := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")

	if err := os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(projectDir, "src", "auth", "jwt.ts"), "// jwt")

	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:      graph,
		FilePath:   "src/auth/*.ts",
		JSON:       true,
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result GlobLookupResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if result.Pattern != "src/auth/*.ts" {
		t.Errorf("expected pattern field, got %q", result.Pattern)
	}
	if len(result.Files) != 1 || result.Files[0] != "src/auth/jwt.ts" {
		t.Errorf("expected [src/auth/jwt.ts], got %v", result.Files)
	}
	if len(result.Components) != 1 || result.Components[0].ID != "c3-101" {
		t.Errorf("expected c3-101 component, got %v", result.Components)
	}
	if len(result.FileMap["src/auth/jwt.ts"]) != 1 {
		t.Errorf("expected file_map entry for jwt.ts, got %v", result.FileMap)
	}
}

func TestRunLookup_JSONNoMatch(t *testing.T) {
	c3Dir, _ := createLookupFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - src/auth/*.ts\n")
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Graph:    graph,
		FilePath: "src/other/file.ts",
		JSON:     true,
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result LookupResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(result.Matches) != 0 {
		t.Errorf("expected empty matches, got %v", result.Matches)
	}
}
