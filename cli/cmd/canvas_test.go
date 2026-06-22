package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

func TestRunCanvasListRead_BuiltinsCoverTargetUseCases(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "list", JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(),
		"system",
		"container",
		"component",
		"ref",
		"rule",
		"adr",
		"atomic-design-change",
		"pm-requirement",
		"prd",
		"user-story",
	)

	buf.Reset()
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "read", ID: "prd"}, &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(), "type: canvas", "edge<requirement|story>", "cite", "check")
}

func TestRunCanvasList_AgentTOONIncludesHelpHints(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	t.Setenv("C3X_MODE", "agent")
	var buf bytes.Buffer

	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "list", JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help[",
		"c3x canvas read <type>",
		"c3x schema <type>",
	)
}

func TestAllCanvases_AllowsMaterializedBuiltInOverrides(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	def, ok := mustDefinitionForTest(t, "component")
	if !ok {
		t.Fatal("component definition missing")
	}
	if _, err := MaterializeDefinitions(c3Dir, []schema.Canvas{def}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "list", JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(), "component", "canvases/component.md")
}

func TestRunCanvasAddListRead_SealsProjectCanvas(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	if err := RunCanvas(CanvasOptions{
		C3Dir: c3Dir,
		Sub:   "add",
		ID:    "research-note",
		Body:  strings.NewReader(researchCanvasDoc()),
	}, &buf); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(c3Dir, "canvases", "research-note.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	requireAll(t, string(data), "id: research-note", "type: canvas", "c3-seal:", "edge<fact|decision>")

	buf.Reset()
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "read", ID: "research-note"}, &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(), "Research finding", "Evidence")
}

func TestRunCanvasAdd_RejectsUnsupportedPrimitive(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	doc := strings.Replace(researchCanvasDoc(), "edge<fact|decision>", "script", 1)

	err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "add", ID: "research-note", Body: strings.NewReader(doc)}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected invalid primitive to fail")
	}
	requireAll(t, err.Error(), "unsupported type", "script")
}

func TestRunCanvasAdd_IDMismatchIncludesHint(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	doc := strings.Replace(researchCanvasDoc(), "id: research-note", "id: different-note", 1)

	err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "add", ID: "research-note", Body: strings.NewReader(doc)}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected id mismatch to fail")
	}
	requireAll(t, err.Error(), "canvas id mismatch", "hint:", "c3x canvas write different-note --file canvas.md")
}

func TestRunSyncExport_PreservesCanvases(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "add", ID: "research-note", Body: strings.NewReader(researchCanvasDoc())}, &buf); err != nil {
		t.Fatal(err)
	}
	if err := RunSyncExport(ExportOptions{Store: s, OutputDir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(c3Dir, "canvases", "research-note.md")); err != nil {
		t.Fatal("sync export should preserve project canvases")
	}
}

func TestRunCheck_ValidatesProjectCanvases(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "add", ID: "research-note", Body: strings.NewReader(researchCanvasDoc())}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(c3Dir, "canvases", "research-note.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(data), "type: cite", "type: script", 1)
	if err := os.WriteFile(path, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunCheckV2(CheckOptions{Store: s, C3Dir: c3Dir}, &buf)
	if err == nil {
		t.Fatal("expected invalid project canvas to fail check")
	}
	requireAll(t, buf.String(), "canvas", "unsupported type", "script")
}

func researchCanvasDoc() string {
	return `---
id: research-note
type: canvas
description: Research note canvas.
---
domain: research
sections:
  - name: Finding
    content_type: table
    required: true
    columns:
      - name: Research finding
        type: text
      - name: Evidence
        type: cite
      - name: Trace
        type: edge<fact|decision>
reject_if:
  - Finding lacks evidence
workorder: Cite the source before writing the conclusion.
`
}

func projectComponentCanvasDoc() string {
	return `---
id: component
type: canvas
description: Project component shape.
---
domain: software
sections:
  - name: Goal
    content_type: text
    required: true
  - name: Custom Project Section
    content_type: text
    required: true
`
}

func projectADRCanvasDoc() string {
	return `---
id: adr
type: canvas
description: Project ADR shape.
---
domain: software
sections:
  - name: Decision Note
    content_type: text
    required: true
`
}

func researchNoteCanvasDoc() string {
	return `---
id: research-note
type: canvas
description: Evidence-backed research note for concrete investigation findings.
---
domain: research
sections:
  - name: Summary
    content_type: text
    required: true
    purpose: One-paragraph result and confidence level.
  - name: Findings
    content_type: table
    required: true
    purpose: Evidence-backed latency observations.
    columns:
      - name: Finding
        type: text
      - name: Evidence
        type: cite
      - name: Trace
        type: edge<fact|decision>
  - name: Measurements
    content_type: table
    required: true
    purpose: Concrete measured values that make the finding reproducible.
    columns:
      - name: Metric
        type: text
      - name: Value
        type: text
      - name: Evidence
        type: cite
  - name: Decision Pressure
    content_type: table
    required: true
    purpose: Decisions or follow-up checks implied by findings.
    columns:
      - name: Pressure
        type: text
      - name: Target
        type: entity_id
      - name: Confidence
        type: enum
        values: [low, medium, high, N.A - <reason>]
      - name: Result
        type: check
reject_if:
  - Summary lacks the measured latency delta and concrete suspected cause.
  - Findings rows lack citation handles from c3x read --cite.
  - Decision Pressure names a decision without a target entity and check result.
workorder: Ground claims with c3x read --cite, then record only measured facts and named decision pressure.
`
}

func researchNoteBody(ownerCitation, refCitation string) string {
	return `## Summary

Checkout API p95 increased from 180 ms to 420 ms after the 2026-05-31 connection-pool change. Span evidence points to DB pool wait, not JWT validation; confidence is high because lookup and graph context both resolve the source path to c3-101 plus its latency ref and trace rule.

## Findings

| Finding | Evidence | Trace |
|---|---|---|
| Checkout API p95 rose from 180 ms to 420 ms after the pool change. | ` + ownerCitation + ` | fact:p95-latency -> decision:pool-wait-investigation |
| The mapped owner for src/api/handlers/latency.go is c3-101. | ` + refCitation + ` | fact:owner-map -> decision:c3-101-governance-review |

## Measurements

| Metric | Value | Evidence |
|---|---|---|
| p95 latency | 420 ms at 2026-06-05T09:40:00Z; previous baseline 180 ms | ` + ownerCitation + ` |
| db pool wait | 210 ms median wait during checkout summary fan-out | ` + refCitation + ` |

## Decision Pressure

| Pressure | Target | Confidence | Result |
|---|---|---|---|
| Reduce checkout DB pool wait before release gate. | c3-101 | high | go test ./cmd ./internal/store ./internal/index -run TestRunLookup passed |
`
}

func mustDefinitionForTest(t *testing.T, entityType string) (schema.Canvas, bool) {
	t.Helper()
	return schema.DefinitionFor(entityType)
}
