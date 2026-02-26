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
	// c3-110 and c3-201 have empty Code References tables
	if !strings.Contains(output, "Code References") {
		t.Errorf("should warn about empty Code References, got: %s", output)
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

func TestRunCheck_CodeRefFileNotExist(t *testing.T) {
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

## Code References

| File | Purpose |
|------|---------|
| src/auth/nonexistent.ts | This file does not exist |
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
		t.Errorf("should flag nonexistent code reference, got: %s", output)
	}
}

func TestRunCheck_CodeRefFileExists(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	// Create the referenced file
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

## Code References

| File | Purpose |
|------|---------|
| src/auth/jwt.ts | JWT validation |
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false, ProjectDir: dir}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// Should NOT flag the existing file
	if strings.Contains(output, "nonexistent") || strings.Contains(output, "does not exist") {
		t.Errorf("should not flag existing code reference, got: %s", output)
	}
}

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

func TestRunCheck_BidirectionalInconsistency(t *testing.T) {
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
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
refs: [ref-jwt]
---

# auth

## Goal

Auth.
`)

	// ref-jwt Cited By is empty — inconsistent with c3-101 citing it
	writeFile(t, filepath.Join(c3Dir, "refs", "ref-jwt.md"), `---
id: ref-jwt
title: JWT
goal: Tokens
scope: [c3-1]
---

# JWT

## Goal

Tokens.

## Choice

JWT RS256.

## Why

Standard.

## Cited By

| Component | Usage |
|-----------|-------|
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := CheckOptions{Graph: graph, Docs: docs, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// Should detect: c3-101 cites ref-jwt but ref-jwt's Cited By doesn't list c3-101
	// Use || not && — either entity name should appear in the inconsistency warning
	if !strings.Contains(output, "c3-101") || !strings.Contains(output, "ref-jwt") {
		t.Errorf("should detect bidirectional inconsistency between c3-101 and ref-jwt, got: %s", output)
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

	// Should have issues from schema validation
	hasSchemaIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "Code References") ||
			strings.Contains(issue.Message, "empty") {
			hasSchemaIssue = true
			break
		}
	}
	if !hasSchemaIssue {
		t.Error("JSON output should include schema validation issues")
	}
}
