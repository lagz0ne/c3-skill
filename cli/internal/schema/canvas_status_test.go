package schema

import (
	"strings"
	"testing"
)

// declaringCanvasDoc returns a minimal canvas markdown that declares the status
// legal set in its frontmatter (keyword `status`), plus one STRICT change-set
// table so it satisfies the change-doc shape introduced in slice 0b.
func declaringCanvasDoc(t *testing.T) string {
	t.Helper()
	return `---
id: demo-change
type: canvas
description: A demo change-doc canvas declaring the status set.
status: [open, accepted, done, superseded]
---

domain: software
sections:
    - name: Rationale
      content_type: text
      required: true
      free: true
      purpose: Free-form reasoning
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

// TestParseCanvasDocument_DeclaresStatusField — a canvas frontmatter declaring
// `status: [open, accepted, done, superseded]` parses without error and the
// parsed Canvas exposes the declared status set.
func TestParseCanvasDocument_DeclaresStatusField(t *testing.T) {
	canvas, err := ParseCanvasDocument("canvases/demo-change.md", declaringCanvasDoc(t))
	if err != nil {
		t.Fatalf("expected canvas declaring status to parse, got error: %v", err)
	}
	want := []string{"open", "accepted", "done", "superseded"}
	if len(canvas.Status) != len(want) {
		t.Fatalf("declared status set = %v, want %v", canvas.Status, want)
	}
	for i := range want {
		if canvas.Status[i] != want[i] {
			t.Fatalf("declared status[%d] = %q, want %q (full = %v)", i, canvas.Status[i], want[i], canvas.Status)
		}
	}
}

// TestParseCanvasDocument_StatusKeyNoLongerRejected — declaring the `status`
// frontmatter keyword must NOT trip the "frontmatter allows only id, type,
// description, and c3-seal" rejection.
func TestParseCanvasDocument_StatusKeyNoLongerRejected(t *testing.T) {
	_, err := ParseCanvasDocument("canvases/demo-change.md", declaringCanvasDoc(t))
	if err != nil && strings.Contains(err.Error(), "frontmatter allows only") {
		t.Fatalf("status keyword should be an allowed frontmatter declaration, got rejection: %v", err)
	}
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

// TestParseCanvasDocument_MalformedStatusFails — an empty or duplicate-state
// status declaration FAILS parse with a clear message.
func TestParseCanvasDocument_MalformedStatusFails(t *testing.T) {
	cases := map[string]string{
		"empty status list": `---
id: bad-empty
type: canvas
description: Bad canvas with empty status set.
status: []
---

domain: software
sections:
    - name: Change Set
      content_type: table
      required: true
      columns:
        - name: Entity
          type: text
`,
		"duplicate status state": `---
id: bad-dup
type: canvas
description: Bad canvas with duplicate status state.
status: [open, accepted, open, done]
---

domain: software
sections:
    - name: Change Set
      content_type: table
      required: true
      columns:
        - name: Entity
          type: text
`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseCanvasDocument("canvases/bad.md", raw)
			if err == nil {
				t.Fatalf("expected malformed status declaration to FAIL parse, got nil error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), "status") {
				t.Fatalf("malformed-status error should mention status, got: %v", err)
			}
		})
	}
}

// TestIsChangeDoc_KeysOnDeclarationNotColumn — IsChangeDoc keys on the declared
// status field: true for change docs, false for fact / story canvases.
func TestIsChangeDoc_KeysOnDeclarationNotColumn(t *testing.T) {
	changeDocs := []string{"adr", "prd", "atomic-design-change"}
	for _, id := range changeDocs {
		if !IsChangeDoc(id) {
			t.Errorf("IsChangeDoc(%q) = false, want true (declares status)", id)
		}
	}
	facts := []string{"user-story", "component", "system"}
	for _, id := range facts {
		if IsChangeDoc(id) {
			t.Errorf("IsChangeDoc(%q) = true, want false (no declared status)", id)
		}
	}
}

// TestIsChangeDoc_StatusColumnIsNotAStatusField — THE keystone false-positive
// guard. A canvas carrying a table column literally named "Status" but NO
// declared status field must yield IsChangeDoc == false.
func TestIsChangeDoc_StatusColumnIsNotAStatusField(t *testing.T) {
	// system canvas carries a real "Status" table column but declares no status set.
	canvas, ok := DefinitionFor("system")
	if !ok {
		t.Fatal("system canvas should exist")
	}
	hasStatusColumn := false
	for _, s := range canvas.Sections {
		for _, c := range s.Columns {
			if c.Name == "Status" {
				hasStatusColumn = true
			}
		}
	}
	if !hasStatusColumn {
		t.Fatal("precondition: system canvas should carry a column named \"Status\"")
	}
	if IsChangeDoc("system") {
		t.Fatal("a table column named \"Status\" must NOT make IsChangeDoc true (keystone false-positive)")
	}
}
