package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCheck_Clean(t *testing.T) {
	c3Dir := createFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunCheck(graph, docs, false, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "entities") {
		t.Errorf("output should mention entity count: %s", output)
	}
}

func TestRunCheck_DuplicateID(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---
# Test
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "README.md"), `---
id: c3-1
title: api
type: container
parent: c3-0
---
# api
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
---
# auth
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-duplicate.md"), `---
id: c3-101
title: duplicate
type: component
parent: c3-1
---
# duplicate
Content.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunCheck(graph, docs, false, &buf)

	output := buf.String()
	if !strings.Contains(output, "duplicate ID") {
		t.Errorf("should detect duplicate ID, got: %s", output)
	}
}

func TestRunCheck_BrokenLink(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---
# Test
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "README.md"), `---
id: c3-1
title: api
type: container
parent: c3-0
---
# api
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"), `---
id: c3-101
title: auth
type: component
parent: c3-1
refs: [ref-nonexistent]
---
# auth
Content.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunCheck(graph, docs, false, &buf)

	output := buf.String()
	if !strings.Contains(output, "broken link") {
		t.Errorf("should detect broken link, got: %s", output)
	}
	if !strings.Contains(output, "ref-nonexistent") {
		t.Errorf("should mention ref-nonexistent, got: %s", output)
	}
}

func TestRunCheck_MissingParent(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---
# Test
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "orphan-component.md"), `---
id: c3-101
title: orphan
type: component
---
# orphan
Content.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunCheck(graph, docs, false, &buf)

	output := buf.String()
	if !strings.Contains(output, "missing parent") {
		t.Errorf("should detect missing parent, got: %s", output)
	}
}

func TestRunCheck_EmptyBody(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunCheck(graph, docs, false, &buf)

	output := buf.String()
	if !strings.Contains(output, "empty content body") {
		t.Errorf("should detect empty body, got: %s", output)
	}
}

func TestRunCheck_IDFilenameMismatch(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "refs"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---
# Test
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "refs", "ref-wrong.md"), `---
id: ref-right
title: Wrong Name
---
# Wrong
Content.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunCheck(graph, docs, false, &buf)

	output := buf.String()
	if !strings.Contains(output, "ID/filename mismatch") {
		t.Errorf("should detect ID/filename mismatch, got: %s", output)
	}
}

func TestRunCheck_JSON(t *testing.T) {
	c3Dir := createFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunCheck(graph, docs, true, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if result.Total < 5 {
		t.Errorf("expected at least 5 entities, got %d", result.Total)
	}
}

func TestRunCheck_OrphanDetection(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "refs"), 0755)

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: Test
---
# Test
Content.
`)

	writeFile(t, filepath.Join(c3Dir, "refs", "ref-orphan.md"), `---
id: ref-orphan
title: Orphan Ref
---
# Orphan
Content.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunCheck(graph, docs, false, &buf)

	output := buf.String()
	if !strings.Contains(output, "orphan") {
		t.Errorf("should detect orphan ref, got: %s", output)
	}
}
