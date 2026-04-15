package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunImport_RebuildsDBFromMarkdown(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	for _, sub := range []string{
		c3Dir,
		filepath.Join(c3Dir, "c3-1-api"),
		filepath.Join(c3Dir, "refs"),
		filepath.Join(c3Dir, "adr"),
	} {
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
c3-version: 4
c3-seal: 9f39b7a2c70bb6a149f7d1be9d65f1f3052d0af5dc8b8c42a98b56d3544f523e
title: TestProject
summary: System summary
---

# TestProject

## Goal

System goal.
`)
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "README.md"), `---
id: c3-1
c3-version: 4
c3-seal: 5e5101f95da1ad295378ef4d0b67dd35e2887d85c4a6a4cb767f520f70338698
title: api
type: container
boundary: service
parent: c3-0
summary: Container summary
---

# api

## Goal

Serve API requests.
`)
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
c3-version: 4
c3-seal: d8c7fe5947b3126ece5935a4b2faee1b8f57da11db4ec70f884de74357d49fd8
title: JWT Authentication
summary: Ref summary
description: Token standard
scope: [c3-1]
---

# JWT Authentication

## Goal

Standardize auth tokens.
`)
	writeFile(t, filepath.Join(c3Dir, "adr", "adr-20260312-use-go.md"), `---
id: adr-20260312-use-go
c3-version: 4
c3-seal: d598e3de379f3293c604460b57d394a00df1b0b97ccf412ad77b852cc79f7799
title: Use Go for CLI
type: adr
date: 2026-03-12T00:00:00Z
summary: ADR summary
---

# Use Go for CLI

## Context

Need fast CLI.
`)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-1:\n  - cli/**\n_exclude:\n  - dist/**\n")

	// Seed a wrong DB to prove import repairs from markdown instead of reusing stale data.
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open stale db: %v", err)
	}
	if err := s.InsertEntity(&store.Entity{
		ID:       "c3-0",
		Type:     "system",
		Title:    "Stale",
		Status:   "active",
		Metadata: "{}",
	}); err != nil {
		t.Fatalf("insert stale entity: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close stale db: %v", err)
	}

	var buf bytes.Buffer
	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true}, &buf); err != nil {
		t.Fatalf("RunImport: %v", err)
	}
	if !strings.Contains(buf.String(), "Imported") {
		t.Fatalf("expected import summary, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Backed up") {
		t.Fatalf("expected backup summary, got %q", buf.String())
	}

	s, err = store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open repaired db: %v", err)
	}
	defer s.Close()

	var export bytes.Buffer
	if err := RunExport(ExportOptions{Store: s, OutputDir: filepath.Join(dir, "export")}, &export); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	readme, err := os.ReadFile(filepath.Join(dir, "export", "README.md"))
	if err != nil {
		t.Fatalf("read exported readme: %v", err)
	}
	if !strings.Contains(string(readme), "summary: System summary") {
		t.Fatalf("expected repaired summary in export:\n%s", string(readme))
	}
	if !strings.Contains(string(readme), "c3-version: 4") {
		t.Fatalf("expected repaired c3-version in export:\n%s", string(readme))
	}

	refDoc, err := os.ReadFile(filepath.Join(dir, "export", "refs", "ref-jwt.md"))
	if err != nil {
		t.Fatalf("read exported ref: %v", err)
	}
	if !strings.Contains(string(refDoc), "description: Token standard") {
		t.Fatalf("expected repaired description in ref export:\n%s", string(refDoc))
	}

	diffOut := &bytes.Buffer{}
	if err := RunDiff(s, false, "", false, diffOut); err != nil {
		t.Fatalf("RunDiff: %v", err)
	}
	if !strings.Contains(diffOut.String(), "No uncommitted changes.") {
		t.Fatalf("expected clean changelog after import, got:\n%s", diffOut.String())
	}

	backups, err := filepath.Glob(filepath.Join(c3Dir, "c3.db.bak-*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one backup file, got %v", backups)
	}
}

func TestRunImport_SurfacesLayerDisconnectAfterRebuild(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	for _, sub := range []string{
		c3Dir,
		filepath.Join(c3Dir, "c3-1-api"),
	} {
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}
	}

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: TestProject
---

# TestProject

## Goal

System goal.

## Containers

| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |
| --- | --- | --- | --- | --- | --- |
| c3-1 | api | service | active | Serve API | API layer |
`)
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "README.md"), `---
id: c3-1
title: api
type: container
parent: c3-0
---

# api

## Goal

Serve API requests.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |

## Responsibilities

Serve API requests.
`)
	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
---

# auth

## Goal

Authenticate users.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| OUT | session | c3-1 |
`)

	var buf bytes.Buffer
	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true}, &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(),
		"Imported 3 entities",
		"Layer integration issues after rebuild",
		"c3-101",
		"missing from c3-1 Components table",
	)
}

func TestRunImport_RejectsBrokenSealWithoutForce(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
c3-seal: broken
title: TestProject
---

# TestProject

## Goal

System goal.
`)
	err := RunImport(ImportOptions{C3Dir: c3Dir}, io.Discard)
	if err == nil {
		t.Fatal("expected broken seal error")
	}
	if !strings.Contains(err.Error(), "broken C3 seal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunImport_RequiresForceWhenDatabaseExists(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: TestProject
---

# TestProject

## Goal

System goal.
`)
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open stale db: %v", err)
	}
	_ = s.Close()

	err = RunImport(ImportOptions{C3Dir: c3Dir}, io.Discard)
	if err == nil {
		t.Fatal("expected force guard error")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected --force hint, got %v", err)
	}
}

func TestRunImport_RejectsUnsealedFileWhenDatabaseMissing(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: TestProject
---

# TestProject

## Goal

System goal.
`)
	err := RunImport(ImportOptions{C3Dir: c3Dir}, io.Discard)
	if err == nil {
		t.Fatal("expected unsealed file error")
	}
	if !strings.Contains(err.Error(), "unsealed C3 file") {
		t.Fatalf("unexpected error: %v", err)
	}
}
