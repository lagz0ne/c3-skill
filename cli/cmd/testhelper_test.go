package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// createFixture sets up a .c3/ directory with common test data.
// Returns the .c3/ directory path.
func createFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")

	// Create directory structure
	for _, d := range []string{
		c3Dir,
		filepath.Join(c3Dir, "c3-1-api"),
		filepath.Join(c3Dir, "c3-2-web"),
		filepath.Join(c3Dir, "refs"),
		filepath.Join(c3Dir, "adr"),
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
`)

	// Container 1
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

Serve API requests

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
`)

	// Component c3-101
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
uses: [ref-jwt]
---

# auth

## Goal

Handle authentication.
`)

	// Component c3-110
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

	// Container 2
	writeFile(t, filepath.Join(c3Dir, "c3-2-web", "README.md"), `---
id: c3-2
title: web
type: container
boundary: app
parent: c3-0
---

# web

## Goal

Web frontend.
`)

	// Ref
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
title: JWT Authentication
goal: Standardize auth tokens
scope: [c3-1]
---

# JWT Authentication

## Goal

Standardize auth tokens.
`)

	// ADR
	writeFile(t, filepath.Join(c3Dir, "adr", "adr-20260226-use-go.md"), `---
id: adr-20260226-use-go
title: Use Go for CLI
type: adr
status: proposed
date: "20260226"
affects: [c3-0]
---

# Use Go for CLI

## Context

Need fast CLI.
`)

	return c3Dir
}

// loadDocs walks a .c3/ directory and returns parsed docs.
func loadDocs(t *testing.T, c3Dir string) []frontmatter.ParsedDoc {
	t.Helper()
	docs, err := walker.WalkC3Docs(c3Dir)
	if err != nil {
		t.Fatalf("WalkC3Docs failed: %v", err)
	}
	return docs
}

// loadGraph builds a C3Graph from a .c3/ directory.
func loadGraph(t *testing.T, c3Dir string) *walker.C3Graph {
	t.Helper()
	return walker.BuildGraph(loadDocs(t, c3Dir))
}

// createRichFixture sets up a .c3/ directory with decorated content sections (tables, refs, etc.)
// Used by tests that exercise schema validation, wire/unwire, and enhanced check.
func createRichFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")

	for _, d := range []string{
		c3Dir,
		filepath.Join(c3Dir, "c3-1-api"),
		filepath.Join(c3Dir, "c3-2-web"),
		filepath.Join(c3Dir, "refs"),
		filepath.Join(c3Dir, "adr"),
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

| ID | Name | Boundary | Goal |
|----|------|----------|------|
| c3-1 | api | service | Serve API requests |
| c3-2 | web | app | Web frontend |
`)

	// Container 1 — with Components table
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
| c3-110 | users | feature | active | User management |
`)

	// Component c3-101 — with Dependencies, Code References, Related Refs
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
category: foundation
parent: c3-1
uses: [ref-jwt]
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
`)

	// Component c3-110 — with empty Code References
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

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
`)

	// Container 2
	writeFile(t, filepath.Join(c3Dir, "c3-2-web", "README.md"), `---
id: c3-2
title: web
type: container
boundary: app
parent: c3-0
---

# web

## Goal

Web frontend.

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-201 | renderer | feature | active | Renders pages |
`)

	// Component c3-201 — with empty Code References
	writeFile(t, filepath.Join(c3Dir, "c3-2-web", "c3-201-renderer.md"), `---
id: c3-201
title: renderer
type: component
category: feature
parent: c3-2
---

# renderer

## Goal

Render pages.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|

## Related Refs

| Ref | Role |
|-----|------|
`)

	// Ref
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

Industry standard, asymmetric verification.
`)

	// Second ref — for wire tests
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-error-handling.md"), `---
id: ref-error-handling
title: Error Handling
goal: Consistent error responses
scope: [c3-1, c3-2]
---

# Error Handling

## Goal

Consistent error responses.

## Choice

RFC 7807 Problem Details.

## Why

Machine-readable error format.
`)

	// ADR
	writeFile(t, filepath.Join(c3Dir, "adr", "adr-20260226-use-go.md"), `---
id: adr-20260226-use-go
title: Use Go for CLI
type: adr
status: proposed
date: "20260226"
affects: [c3-0]
---

# Use Go for CLI

## Context

Need fast CLI.
`)

	return c3Dir
}

// containsStr2 checks if a string slice contains a value.
func containsStr2(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}

// createDBFixture creates an in-memory store populated with the same entities
// and relationships as createFixture: c3-0, c3-1, c3-2, c3-101, c3-110,
// ref-jwt, adr-20260226-use-go.
func createDBFixture(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	entities := []*store.Entity{
		{ID: "c3-0", Type: "system", Title: "TestProject", Slug: "", Goal: "", Summary: "", Body: "# TestProject\n\n## Goal\n\nTest the system.\n", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "api", Slug: "api", ParentID: "c3-0", Goal: "Serve API requests", Boundary: "service", Status: "active", Metadata: "{}"},
		{ID: "c3-2", Type: "container", Title: "web", Slug: "web", ParentID: "c3-0", Boundary: "app", Status: "active", Metadata: "{}"},
		{ID: "c3-101", Type: "component", Title: "auth", Slug: "auth", Category: "foundation", ParentID: "c3-1", Status: "active", Metadata: "{}"},
		{ID: "c3-110", Type: "component", Title: "users", Slug: "users", Category: "feature", ParentID: "c3-1", Status: "active", Metadata: "{}"},
		{ID: "ref-jwt", Type: "ref", Title: "JWT Authentication", Slug: "jwt", Goal: "Standardize auth tokens", Status: "active", Metadata: "{}"},
		{ID: "adr-20260226-use-go", Type: "adr", Title: "Use Go for CLI", Slug: "use-go", Status: "proposed", Date: "20260226", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed entity %s: %v", e.ID, err)
		}
	}

	rels := []*store.Relationship{
		{FromID: "c3-101", ToID: "ref-jwt", RelType: "uses"},
		{FromID: "ref-jwt", ToID: "c3-1", RelType: "scope"},
		{FromID: "adr-20260226-use-go", ToID: "c3-0", RelType: "affects"},
	}
	for _, r := range rels {
		if err := s.AddRelationship(r); err != nil {
			t.Fatalf("seed rel %s->%s: %v", r.FromID, r.ToID, err)
		}
	}

	return s
}
