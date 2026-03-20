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
// Kept for migration tests that need file-based fixtures.
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
// Kept for migration tests.
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

// createRichDBFixture creates a store with the same entities as createRichFixture,
// including body content with tables for wire/set/check tests.
func createRichDBFixture(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	entities := []*store.Entity{
		{ID: "c3-0", Type: "system", Title: "TestProject", Slug: "", Body: "# TestProject\n\n## Goal\n\nTest the system.\n\n## Containers\n\n| ID | Name | Boundary | Goal |\n|----|------|----------|------|\n| c3-1 | api | service | Serve API requests |\n| c3-2 | web | app | Web frontend |\n", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "api", Slug: "api", ParentID: "c3-0", Goal: "Serve API requests", Boundary: "service", Body: "# api\n\n## Goal\n\nServe API requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| c3-101 | auth | foundation | active | Authentication |\n| c3-110 | users | feature | active | User management |\n", Status: "active", Metadata: "{}"},
		{ID: "c3-2", Type: "container", Title: "web", Slug: "web", ParentID: "c3-0", Boundary: "app", Body: "# web\n\n## Goal\n\nWeb frontend.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| c3-201 | renderer | feature | active | Renders pages |\n", Status: "active", Metadata: "{}"},
		{ID: "c3-101", Type: "component", Title: "auth", Slug: "auth", Category: "foundation", ParentID: "c3-1", Body: "# auth\n\n## Goal\n\nHandle authentication.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n| IN | user credentials | c3-110 |\n\n## Related Refs\n\n| Ref | Role |\n|-----|------|\n| ref-jwt | Token format |\n", Status: "active", Metadata: "{}"},
		{ID: "c3-110", Type: "component", Title: "users", Slug: "users", Category: "feature", ParentID: "c3-1", Body: "# users\n\n## Goal\n\nManage user accounts.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n", Status: "active", Metadata: "{}"},
		{ID: "c3-201", Type: "component", Title: "renderer", Slug: "renderer", Category: "feature", ParentID: "c3-2", Body: "# renderer\n\n## Goal\n\nRender pages.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n\n## Related Refs\n\n| Ref | Role |\n|-----|------|\n", Status: "active", Metadata: "{}"},
		{ID: "ref-jwt", Type: "ref", Title: "JWT Authentication", Slug: "jwt", Goal: "Standardize auth tokens", Body: "# JWT Authentication\n\n## Goal\n\nStandardize auth tokens.\n\n## Choice\n\nUse RS256 signed JWTs.\n\n## Why\n\nIndustry standard, asymmetric verification.\n", Status: "active", Metadata: "{}"},
		{ID: "ref-error-handling", Type: "ref", Title: "Error Handling", Slug: "error-handling", Goal: "Consistent error responses", Body: "# Error Handling\n\n## Goal\n\nConsistent error responses.\n\n## Choice\n\nRFC 7807 Problem Details.\n\n## Why\n\nMachine-readable error format.\n", Status: "active", Metadata: "{}"},
		{ID: "adr-20260226-use-go", Type: "adr", Title: "Use Go for CLI", Slug: "use-go", Status: "proposed", Date: "20260226", Body: "# Use Go for CLI\n\n## Context\n\nNeed fast CLI.\n", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed entity %s: %v", e.ID, err)
		}
	}

	rels := []*store.Relationship{
		{FromID: "c3-101", ToID: "ref-jwt", RelType: "uses"},
		{FromID: "ref-jwt", ToID: "c3-1", RelType: "scope"},
		{FromID: "ref-error-handling", ToID: "c3-1", RelType: "scope"},
		{FromID: "ref-error-handling", ToID: "c3-2", RelType: "scope"},
		{FromID: "adr-20260226-use-go", ToID: "c3-0", RelType: "affects"},
	}
	for _, r := range rels {
		if err := s.AddRelationship(r); err != nil {
			t.Fatalf("seed rel %s->%s: %v", r.FromID, r.ToID, err)
		}
	}

	return s
}

// createDBFixtureWithC3Dir creates a DB fixture AND a c3Dir on disk (needed for add commands that write files).
func createDBFixtureWithC3Dir(t *testing.T) (*store.Store, string) {
	t.Helper()
	s := createDBFixture(t)

	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	// Create container directories matching the fixture
	for _, d := range []string{
		c3Dir,
		filepath.Join(c3Dir, "c3-1-api"),
		filepath.Join(c3Dir, "c3-2-web"),
	} {
		os.MkdirAll(d, 0755)
	}

	return s, c3Dir
}
