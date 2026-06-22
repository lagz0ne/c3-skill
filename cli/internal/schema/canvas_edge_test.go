package schema

import (
	"strings"
	"testing"
)

// edgeCanvas builds a canvas whose Governance table declares one citation column
// (`edge: uses`, targets ref/rule). If dupEdge is true it declares a SECOND
// `edge: uses` column, which must be rejected (ambiguous writer).
func edgeCanvas(dupEdge bool) string {
	second := ""
	if dupEdge {
		second = `        - name: AlsoReference
          type: reference
          edge: uses
`
	}
	return `---
id: edge-demo
type: canvas
description: A canvas exercising edge-column metadata.
---

domain: software
sections:
    - name: Governance
      content_type: table
      required: true
      purpose: Refs and rules governing this entity
      columns:
        - name: Reference
          type: reference
          edge: uses
          targets:
            - ref
            - rule
        - name: Notes
          type: text
` + second
}

// TestColumnDef_EdgeTargetsParse — a column's `edge`/`targets` metadata parses and
// is exposed on ColumnDef; `type` (parse) and `edge` (relationship) are distinct.
func TestColumnDef_EdgeTargetsParse(t *testing.T) {
	canvas, err := ParseCanvasDocument("canvases/edge-demo.md", edgeCanvas(false))
	if err != nil {
		t.Fatalf("expected edge canvas to parse, got: %v", err)
	}
	if err := ValidateCanvas(canvas); err != nil {
		t.Fatalf("single edge:uses writer should validate, got: %v", err)
	}
	var ref *ColumnDef
	for i := range canvas.Sections {
		for j := range canvas.Sections[i].Columns {
			if canvas.Sections[i].Columns[j].Name == "Reference" {
				ref = &canvas.Sections[i].Columns[j]
			}
		}
	}
	if ref == nil {
		t.Fatal("expected a Reference column")
	}
	if ref.Type != "reference" {
		t.Errorf("Type should stay the parse primitive, got %q", ref.Type)
	}
	if ref.Edge != "uses" {
		t.Errorf("Edge should be the relationship %q, got %q", "uses", ref.Edge)
	}
	if strings.Join(ref.Targets, ",") != "ref,rule" {
		t.Errorf("Targets should be [ref rule], got %v", ref.Targets)
	}
}

// TestValidateCanvas_RejectsDuplicateEdgeWriter — two columns claiming the same
// relationship makes `c3 wire` ambiguous, so loading the canvas must be rejected
// (ParseCanvasDocument validates).
func TestValidateCanvas_RejectsDuplicateEdgeWriter(t *testing.T) {
	_, err := ParseCanvasDocument("canvases/edge-demo.md", edgeCanvas(true))
	if err == nil {
		t.Fatal("two edge:uses columns must be rejected (one unambiguous writer per relationship)")
	}
	if !strings.Contains(err.Error(), "at most one writer") {
		t.Fatalf("expected unambiguous-writer error, got: %v", err)
	}
}
