package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRunMigrateDryRun_DetectsGoalGap(t *testing.T) {
	c3Dir := createFixture(t)
	// The fixture has entities with goal in body but not frontmatter
	// Remove frontmatter goal from container to test gap detection
	writeFile(t, c3Dir+"/c3-1-api/README.md", `---
id: c3-1
title: api
type: container
boundary: service
parent: c3-0
---

# api

## Goal

Serve API requests with low latency.

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
`)

	var buf bytes.Buffer
	err := RunMigrateDryRun(c3Dir, false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-1") {
		t.Error("should report c3-1")
	}
	if !strings.Contains(output, "`goal:` empty") {
		t.Error("should flag empty goal frontmatter")
	}
	if !strings.Contains(output, "Serve API requests") {
		t.Error("should show body excerpt")
	}
}

func TestRunMigrateDryRun_JSON(t *testing.T) {
	c3Dir := createFixture(t)

	var buf bytes.Buffer
	err := RunMigrateDryRun(c3Dir, true, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result MigrateDryRunResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if result.Total == 0 {
		t.Error("should have entities")
	}
}

func TestRunMigrateDryRun_CleanEntity(t *testing.T) {
	c3Dir := createFixture(t)
	// Add a ref with all required sections filled
	writeFile(t, c3Dir+"/refs/ref-clean.md", `---
id: ref-clean
title: Clean Ref
goal: Already has goal in frontmatter
---

# Clean Ref

## Goal

Already has goal in frontmatter.

## Choice

Use the clean approach.

## Why

Because it's clean.
`)

	var buf bytes.Buffer
	err := RunMigrateDryRun(c3Dir, true, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result MigrateDryRunResult
	json.Unmarshal(buf.Bytes(), &result)
	if result.Clean == 0 {
		t.Error("ref-clean should be counted as clean")
	}
}

func TestRunMigrateDryRun_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	c3Dir := dir + "/.c3"
	os.MkdirAll(c3Dir, 0755)

	var buf bytes.Buffer
	err := RunMigrateDryRun(c3Dir, false, &buf)
	if err == nil {
		t.Error("expected error for empty .c3/")
	}
}

func TestRunMigrateDryRun_MissingSections(t *testing.T) {
	c3Dir := createFixture(t)
	// Create a ref with no Choice or Why sections
	writeFile(t, c3Dir+"/refs/ref-bare.md", `---
id: ref-bare
title: Bare Ref
goal: Has goal but no sections
---

# Bare Ref

## Goal

Has goal but no sections.
`)

	var buf bytes.Buffer
	RunMigrateDryRun(c3Dir, true, &buf)

	var result MigrateDryRunResult
	json.Unmarshal(buf.Bytes(), &result)

	// Find ref-bare in results
	for _, e := range result.Entities {
		if e.ID == "ref-bare" {
			hasMissing := false
			for _, g := range e.Gaps {
				if g.Status == "missing_section" {
					hasMissing = true
				}
			}
			if !hasMissing {
				t.Error("ref-bare should have missing_section gaps (Choice, Why)")
			}
			return
		}
	}
	t.Error("ref-bare not found in report")
}
