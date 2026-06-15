package schema

import (
	"strings"
	"testing"
)

// freeStrictCanvas declares one FREE reasoning section and exactly one STRICT
// change-set block, plus the status set (so it is a change doc).
func freeStrictCanvas() string {
	return `---
id: fs-change
type: canvas
description: A change-doc canvas with a FREE reasoning section and one STRICT change-set.
status: [open, accepted, done, superseded]
---

domain: software
sections:
    - name: Rationale
      content_type: text
      required: true
      free: true
      purpose: Free-form reasoning, any prose
    - name: Change Set
      content_type: table
      required: true
      purpose: The strict change-set block
      columns:
        - name: Entity
          type: text
        - name: Evidence
          type: cite
`
}

// TestSectionDef_FreeStrictMarkerParses — a section declaring the FREE marker
// parses and SectionDef exposes it; a section without it defaults to STRICT.
func TestSectionDef_FreeStrictMarkerParses(t *testing.T) {
	canvas, err := ParseCanvasDocument("canvases/fs-change.md", freeStrictCanvas())
	if err != nil {
		t.Fatalf("expected FREE/STRICT canvas to parse, got: %v", err)
	}
	var rationale, changeSet *SectionDef
	for i := range canvas.Sections {
		switch canvas.Sections[i].Name {
		case "Rationale":
			rationale = &canvas.Sections[i]
		case "Change Set":
			changeSet = &canvas.Sections[i]
		}
	}
	if rationale == nil || changeSet == nil {
		t.Fatalf("expected Rationale and Change Set sections, got %+v", canvas.Sections)
	}
	if !rationale.Free {
		t.Error("Rationale section should be FREE (free: true)")
	}
	if changeSet.Free {
		t.Error("Change Set section should default to STRICT (free absent)")
	}
}

// TestValidateCanvas_FreeSectionSkippedByShapeChecks — a FREE section is
// excluded from typed-column / shape checks. A FREE text section with an
// otherwise-invalid column type (which a STRICT section would reject) must NOT
// fail validation.
func TestValidateCanvas_FreeSectionSkippedByShapeChecks(t *testing.T) {
	// FREE section carries a column with an unsupported type. If shape checks ran
	// on it, ValidateCanvas would reject "unsupported type". Because it is FREE,
	// it must be skipped.
	raw := `---
id: fs-free-skip
type: canvas
description: FREE section must escape typed-column shape checks.
status: [open, accepted, done, superseded]
---

domain: software
sections:
    - name: Rationale
      content_type: table
      required: true
      free: true
      columns:
        - name: Whatever
          type: totally-not-a-primitive
    - name: Change Set
      content_type: table
      required: true
      columns:
        - name: Entity
          type: text
`
	if _, err := ParseCanvasDocument("canvases/fs-free-skip.md", raw); err != nil {
		t.Fatalf("FREE section content must be skipped by shape checks, got: %v", err)
	}
}

// TestValidateCanvas_StrictChangeSetFullyChecked — the STRICT change-set block
// is fully typed-column checked. An unsupported column type in the STRICT block
// (not FREE) must FAIL validation.
func TestValidateCanvas_StrictChangeSetFullyChecked(t *testing.T) {
	raw := `---
id: fs-strict-checked
type: canvas
description: STRICT change-set must be fully typed-column checked.
status: [open, accepted, done, superseded]
---

domain: software
sections:
    - name: Rationale
      content_type: text
      required: true
      free: true
    - name: Change Set
      content_type: table
      required: true
      columns:
        - name: Entity
          type: totally-not-a-primitive
`
	_, err := ParseCanvasDocument("canvases/fs-strict-checked.md", raw)
	if err == nil {
		t.Fatal("STRICT change-set with an unsupported column type must FAIL validation")
	}
	if !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected unsupported-type rejection on the STRICT block, got: %v", err)
	}
}

// TestChangeDoc_RequiresExactlyOneStrictChangeSet — a change-doc canvas (one
// that declares status) must have exactly one STRICT change-set block. A change
// doc with NO STRICT change-set (all sections FREE) is invalid.
func TestChangeDoc_RequiresExactlyOneStrictChangeSet(t *testing.T) {
	noStrict := `---
id: fs-no-strict
type: canvas
description: A change doc with no STRICT change-set block is invalid.
status: [open, accepted, done, superseded]
---

domain: software
sections:
    - name: Rationale
      content_type: text
      required: true
      free: true
    - name: More Reasoning
      content_type: text
      required: true
      free: true
`
	if _, err := ParseCanvasDocument("canvases/fs-no-strict.md", noStrict); err == nil {
		t.Fatal("a change doc with no STRICT change-set block must be invalid")
	}
}

// TestFreeSection_ContentNeverJudged — NEGATIVE / the line. c3x emits no signal
// about the *content* of a FREE section: arbitrary prose of any length/format in
// a FREE section produces zero validation issues. Mechanical-only guard.
func TestFreeSection_ContentNeverJudged(t *testing.T) {
	raw := `---
id: fs-content-free
type: canvas
description: FREE section content is never judged for shape.
status: [open, accepted, done, superseded]
---

domain: software
sections:
    - name: Rationale
      content_type: text
      required: true
      free: true
      min_words: 999
    - name: Change Set
      content_type: table
      required: true
      columns:
        - name: Entity
          type: text
`
	if _, err := ParseCanvasDocument("canvases/fs-content-free.md", raw); err != nil {
		t.Fatalf("FREE section must never be judged on content (min_words ignored), got: %v", err)
	}
}
