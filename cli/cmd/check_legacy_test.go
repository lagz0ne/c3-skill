package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

func TestRunLegacyCheck_ValidDocs(t *testing.T) {
	c3Dir := createFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  false,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// The basic fixture has schema warnings (missing optional sections)
	// but no errors (broken refs, missing titles, etc.)
	if strings.Contains(output, "error") && !strings.Contains(output, "0 error(s)") {
		t.Errorf("valid docs should have no errors, got: %s", output)
	}
}

func TestRunLegacyCheck_JSONOutput(t *testing.T) {
	c3Dir := createFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse failed: %v", err)
	}
}

func TestRunLegacyCheck_BrokenRef(t *testing.T) {
	c3Dir := createFixture(t)
	// Add a component with a broken reference
	writeFile(t, c3Dir+"/c3-1-api/c3-102-broken.md", `---
id: c3-102
title: broken
type: component
parent: c3-1
uses: [ref-nonexistent]
---

# broken

## Goal

Test broken ref.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  false,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "broken reference") {
		t.Errorf("should report broken reference, got: %s", output)
	}
}

func TestRunLegacyCheck_MissingTitle(t *testing.T) {
	c3Dir := createFixture(t)
	// Add a component with missing title
	writeFile(t, c3Dir+"/c3-1-api/c3-103-notitle.md", `---
id: c3-103
type: component
parent: c3-1
---

# notitle

## Goal

Missing title.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  false,
	}, &buf)

	output := buf.String()
	if !strings.Contains(output, "missing title") {
		t.Errorf("should report missing title, got: %s", output)
	}
}

func TestRunLegacyCheck_WithParseWarnings(t *testing.T) {
	c3Dir := createFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  false,
		ParseWarnings: []walker.ParseWarning{
			{Path: "bad-file.md"},
		},
	}, &buf)

	output := buf.String()
	if !strings.Contains(output, "bad-file.md") {
		t.Errorf("should report parse warning, got: %s", output)
	}
}

func TestRunLegacyCheck_IncludeADR(t *testing.T) {
	c3Dir := createFixture(t)
	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	RunLegacyCheck(LegacyCheckOptions{
		Docs:       docs,
		Graph:      graph,
		JSON:       true,
		IncludeADR: true,
	}, &buf)

	var result CheckResult
	json.Unmarshal(buf.Bytes(), &result)
	// With IncludeADR, ADRs should be validated
	// The fixture ADR has status and date but check may report missing sections
}

func TestRunLegacyCheck_BrokenParent(t *testing.T) {
	c3Dir := createFixture(t)
	writeFile(t, c3Dir+"/c3-1-api/c3-104-badparent.md", `---
id: c3-104
title: badparent
type: component
parent: c3-999
---

# badparent

## Goal

Bad parent.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  false,
	}, &buf)

	output := buf.String()
	if !strings.Contains(output, "parent") && !strings.Contains(output, "not found") {
		t.Errorf("should report bad parent, got: %s", output)
	}
}

func TestRunLegacyCheck_RuleOrigin(t *testing.T) {
	c3Dir := createFixture(t)

	// Add a rule with broken origin
	os.MkdirAll(filepath.Join(c3Dir, "rules"), 0755)
	writeFile(t, filepath.Join(c3Dir, "rules", "rule-broken.md"), `---
id: rule-broken
title: Broken Rule
type: rule
origin: [ref-nonexistent]
---

# Broken Rule

## Goal

Test broken origin.
`)

	docs := loadDocs(t, c3Dir)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	RunLegacyCheck(LegacyCheckOptions{
		Docs:  docs,
		Graph: graph,
		JSON:  false,
	}, &buf)

	output := buf.String()
	if !strings.Contains(output, "origin") {
		t.Errorf("should report broken origin, got: %s", output)
	}
}
