package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
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
refs: [ref-jwt]
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}
