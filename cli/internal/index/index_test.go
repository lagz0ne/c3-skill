package index

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// --- Test helpers ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}

func createFixture(t *testing.T) (string, codemap.CodeMap) {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")

	for _, d := range []string{
		c3Dir,
		filepath.Join(c3Dir, "c3-1-api"),
		filepath.Join(c3Dir, "refs"),
	} {
		os.MkdirAll(d, 0755)
	}

	// Context
	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: TestProject
---

# TestProject

## Goal

Test the system.

## Containers

| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |
|----|------|----------|--------|------------------|-------------------|
| c3-1 | api | service | active | Serve API | Core API |
`)

	// Container
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "README.md"), `---
id: c3-1
title: api
type: container
boundary: service
parent: c3-0
goal: Serve API requests
---

# api

## Goal

Serve API requests.

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-101 | auth | foundation | active | Authentication |

## Responsibilities

Handle all API traffic.
`)

	// Component with refs and filled blocks
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

Handle authentication.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | user credentials | c3-110 |
`)

	// Component without refs
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-110-users.md"), `---
id: c3-110
title: users
type: component
category: feature
parent: c3-1
---

# users

## Goal

Manage user accounts.
`)

	// Ref with scope
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
title: JWT Authentication
goal: Standardize auth tokens
scope: [c3-1]
---

# JWT Authentication

## Goal

Standardize auth tokens.

## Choice

Use RS256 signed JWTs.

## Why

Industry standard.
`)

	cm := codemap.CodeMap{
		"c3-101": {"src/auth/**/*.ts"},
		"c3-110": {"src/users/**/*.ts"},
	}

	return c3Dir, cm
}

func loadGraph(t *testing.T, c3Dir string) *walker.C3Graph {
	t.Helper()
	docs, err := walker.WalkC3Docs(c3Dir)
	if err != nil {
		t.Fatalf("WalkC3Docs: %v", err)
	}
	return walker.BuildGraph(docs)
}

// --- Tests ---

func TestBuild_EntityEntries(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	// Check component c3-101
	e, ok := idx.Entities["c3-101"]
	if !ok {
		t.Fatal("missing entity c3-101")
	}
	if e.Type != "component" {
		t.Errorf("c3-101 type = %q, want component", e.Type)
	}
	if e.Container != "c3-1" {
		t.Errorf("c3-101 container = %q, want c3-1", e.Container)
	}
	if e.Context != "c3-0" {
		t.Errorf("c3-101 context = %q, want c3-0", e.Context)
	}
	if len(e.Refs) != 1 || e.Refs[0] != "ref-jwt" {
		t.Errorf("c3-101 refs = %v, want [ref-jwt]", e.Refs)
	}
	if len(e.CodePaths) != 1 || e.CodePaths[0] != "src/auth/**/*.ts" {
		t.Errorf("c3-101 code paths = %v, want [src/auth/**/*.ts]", e.CodePaths)
	}

	// Check constraints from includes context, container, and ref
	if !containsStr(e.ConstraintsFrom, "c3-0") {
		t.Errorf("c3-101 constraints should include c3-0, got %v", e.ConstraintsFrom)
	}
	if !containsStr(e.ConstraintsFrom, "c3-1") {
		t.Errorf("c3-101 constraints should include c3-1, got %v", e.ConstraintsFrom)
	}
	if !containsStr(e.ConstraintsFrom, "ref-jwt") {
		t.Errorf("c3-101 constraints should include ref-jwt, got %v", e.ConstraintsFrom)
	}

	// Check block fill
	if e.BlockFill == nil {
		t.Fatal("c3-101 block fill is nil")
	}
	if !e.BlockFill["Goal"] {
		t.Error("c3-101 Goal block should be filled")
	}
	if !e.BlockFill["Dependencies"] {
		t.Error("c3-101 Dependencies block should be filled")
	}
}

func TestBuild_ContainerEntry(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	e, ok := idx.Entities["c3-1"]
	if !ok {
		t.Fatal("missing entity c3-1")
	}
	if e.Type != "container" {
		t.Errorf("c3-1 type = %q, want container", e.Type)
	}
	if e.Context != "c3-0" {
		t.Errorf("c3-1 context = %q, want c3-0", e.Context)
	}
	if e.Container != "" {
		t.Errorf("c3-1 container should be empty, got %q", e.Container)
	}
}

func TestBuild_RefEntry(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	ref, ok := idx.Refs["ref-jwt"]
	if !ok {
		t.Fatal("missing ref entry ref-jwt")
	}
	if !containsStr(ref.Citers, "c3-101") {
		t.Errorf("ref-jwt citers should include c3-101, got %v", ref.Citers)
	}
	if !containsStr(ref.Scope, "c3-1") {
		t.Errorf("ref-jwt scope should include c3-1, got %v", ref.Scope)
	}
}

func TestBuild_FileMap(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	fe, ok := idx.Files["src/auth/**/*.ts"]
	if !ok {
		t.Fatal("missing file entry src/auth/**/*.ts")
	}
	if !containsStr(fe.Entities, "c3-101") {
		t.Errorf("file entry entities should include c3-101, got %v", fe.Entities)
	}
	if !containsStr(fe.Refs, "ref-jwt") {
		t.Errorf("file entry refs should include ref-jwt, got %v", fe.Refs)
	}
}

func TestBuild_ReverseDeps(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	// c3-1 should have reverse deps from its children (c3-101, c3-110)
	e := idx.Entities["c3-1"]
	if !containsStr(e.ReverseDeps, "c3-101") {
		t.Errorf("c3-1 reverse deps should include c3-101, got %v", e.ReverseDeps)
	}
	if !containsStr(e.ReverseDeps, "c3-110") {
		t.Errorf("c3-1 reverse deps should include c3-110, got %v", e.ReverseDeps)
	}
}

func TestBuild_HashDeterminism(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)

	idx1 := Build(graph, cm, c3Dir)
	idx2 := Build(graph, cm, c3Dir)

	if idx1.Hash != idx2.Hash {
		t.Errorf("hash not deterministic: %s != %s", idx1.Hash, idx2.Hash)
	}
	if !strings.HasPrefix(idx1.Hash, "sha256:") {
		t.Errorf("hash should start with sha256:, got %s", idx1.Hash)
	}
}

func TestBuild_EmptyCodeMap(t *testing.T) {
	c3Dir, _ := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, codemap.CodeMap{}, c3Dir)

	if len(idx.Files) != 0 {
		t.Errorf("expected empty file map with empty code-map, got %d entries", len(idx.Files))
	}
	e := idx.Entities["c3-101"]
	if len(e.CodePaths) != 0 {
		t.Errorf("expected no code paths with empty code-map, got %v", e.CodePaths)
	}
}

func TestWriteMarkdown(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, idx); err != nil {
		t.Fatalf("WriteMarkdown: %v", err)
	}
	out := buf.String()

	// Check header
	if !strings.Contains(out, "# C3 Structural Index") {
		t.Error("missing header")
	}
	if !strings.Contains(out, "<!-- hash: sha256:") {
		t.Error("missing hash comment")
	}

	// Check entity section
	if !strings.Contains(out, "## c3-101 — auth (component)") {
		t.Error("missing c3-101 entity section")
	}
	if !strings.Contains(out, "container: c3-1") {
		t.Error("missing container line for c3-101")
	}
	if !strings.Contains(out, "refs: ref-jwt") {
		t.Error("missing refs line for c3-101")
	}

	// Check file map
	if !strings.Contains(out, "## File Map") {
		t.Error("missing File Map section")
	}
	if !strings.Contains(out, "src/auth/**/*.ts → c3-101") {
		t.Error("missing file map entry")
	}

	// Check ref map
	if !strings.Contains(out, "## Ref Map") {
		t.Error("missing Ref Map section")
	}
	if !strings.Contains(out, "ref-jwt cited by:") {
		t.Error("missing ref map entry")
	}
}

func TestWriteJSON(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	var buf bytes.Buffer
	if err := WriteJSON(&buf, idx); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var decoded StructuralIndex
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}
	if decoded.Hash != idx.Hash {
		t.Errorf("JSON hash = %q, want %q", decoded.Hash, idx.Hash)
	}
	if len(decoded.Entities) != len(idx.Entities) {
		t.Errorf("JSON entities count = %d, want %d", len(decoded.Entities), len(idx.Entities))
	}
}

func TestWriteTo(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)

	if err := WriteTo(c3Dir, idx); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	outPath := filepath.Join(c3Dir, "_index", "structural.md")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(data), "# C3 Structural Index") {
		t.Error("written file missing header")
	}
}

func TestRefGovernance_AllGoverned(t *testing.T) {
	c3Dir, cm := createFixture(t)
	// The fixture has c3-101 with refs:[ref-jwt] and c3-110 without refs.
	// Add a ref to c3-110 to make all governed.
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-110-users.md"), `---
id: c3-110
title: users
type: component
category: feature
parent: c3-1
refs: [ref-jwt]
---

# users

## Goal

Manage user accounts.
`)

	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)
	result := RefGovernance(idx)

	if result.TotalComponents != 2 {
		t.Errorf("total = %d, want 2", result.TotalComponents)
	}
	if result.Governed != 2 {
		t.Errorf("governed = %d, want 2", result.Governed)
	}
	if result.GovernancePct != 100 {
		t.Errorf("pct = %f, want 100", result.GovernancePct)
	}
	if len(result.UngovernedComponents) != 0 {
		t.Errorf("ungoverned = %v, want empty", result.UngovernedComponents)
	}
}

func TestRefGovernance_SomeUngoverned(t *testing.T) {
	c3Dir, cm := createFixture(t)
	graph := loadGraph(t, c3Dir)
	idx := Build(graph, cm, c3Dir)
	result := RefGovernance(idx)

	// c3-101 has refs, c3-110 does not
	if result.TotalComponents != 2 {
		t.Errorf("total = %d, want 2", result.TotalComponents)
	}
	if result.Governed != 1 {
		t.Errorf("governed = %d, want 1", result.Governed)
	}
	if result.GovernancePct != 50 {
		t.Errorf("pct = %f, want 50", result.GovernancePct)
	}
	if !containsStr(result.UngovernedComponents, "c3-110") {
		t.Errorf("ungoverned should include c3-110, got %v", result.UngovernedComponents)
	}
}

func TestRefGovernance_NoComponents(t *testing.T) {
	// Empty index — no components at all
	idx := &StructuralIndex{
		Entities: map[string]EntityEntry{},
		Files:    map[string]FileEntry{},
		Refs:     map[string]RefEntry{},
	}
	result := RefGovernance(idx)

	if result.TotalComponents != 0 {
		t.Errorf("total = %d, want 0", result.TotalComponents)
	}
	if result.GovernancePct != 0 {
		t.Errorf("pct = %f, want 0", result.GovernancePct)
	}
}

func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
