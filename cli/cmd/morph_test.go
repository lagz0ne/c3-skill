package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// widgetV2 raises the widget canvas: Risk becomes a second required section.
const widgetV2 = `---
id: widget
type: canvas
description: A widget, now carrying risk.
---

domain: software
sections:
    - name: Spec
      content_type: text
      required: true
    - name: Risk
      content_type: text
      required: true
`

// Morph: raising a canvas (a section becomes required) is refused while an existing
// instance does not yet satisfy it — UNLESS the same unit migrates that instance to the
// new shape. Mirrors the retire gate: the model moves only if every fact can come with it.
func TestMorphGate_RefusesMorphThatStrandsInstance(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "w-1", Type: "widget", Title: "one", Status: "active", Metadata: "{}"})
	if err := content.WriteEntity(s, "w-1", "## Spec\n\nthe spec.\n"); err != nil {
		t.Fatal(err)
	}
	morph := changeset.Patch{Target: "widget", Scope: changeset.ScopeCanvas, Content: widgetV2, Source: "01.canvas.md"}
	morphed, factPatches, rejects := parseMorphs([]changeset.Patch{morph})
	if len(rejects) != 0 {
		t.Fatalf("widgetV2 is a valid canvas; parseMorphs should not reject: %v", rejects)
	}

	// morph alone → w-1 lacks the now-required Risk section → refused.
	if r := morphGate(s, c3Dir, factPatches, morphed); len(r) == 0 || !strings.Contains(joined(r), "w-1") {
		t.Fatalf("morphing a canvas that strands an instance must be refused, got: %v", r)
	}

	// once every instance satisfies the new shape, the same morph lands clean.
	if err := content.WriteEntity(s, "w-1", "## Spec\n\nthe spec.\n\n## Risk\n\nnone yet.\n"); err != nil {
		t.Fatal(err)
	}
	if r := morphGate(s, c3Dir, factPatches, morphed); len(r) != 0 {
		t.Errorf("a morph whose instances all satisfy the new shape must be allowed, got: %v", r)
	}
}

// parseMorphs splits canvas-scope patches from fact patches and rejects a morph whose
// new shape is not a valid canvas for the type it targets — before any gate runs.
func TestParseMorphs_SplitsAndValidates(t *testing.T) {
	canvasP := changeset.Patch{Target: "widget", Scope: changeset.ScopeCanvas, Content: widgetV2, Source: "01.canvas.md"}
	factP := changeset.Patch{Target: "w-1", Scope: changeset.ScopeBlock, Base: "w-1#n1@v0:sha256:" + strings.Repeat("a", 64), Content: "x", Source: "02.block.md"}
	morphed, fact, rejects := parseMorphs([]changeset.Patch{canvasP, factP})
	if len(rejects) != 0 {
		t.Fatalf("unexpected rejects: %v", rejects)
	}
	if _, ok := morphed["widget"]; !ok {
		t.Errorf("widget morph missing from the morphed map")
	}
	if len(fact) != 1 || fact[0].Target != "w-1" {
		t.Errorf("fact patch was not split out: %v", fact)
	}

	// A morph whose new shape is not even a valid canvas is refused outright.
	bad := changeset.Patch{Target: "widget", Scope: changeset.ScopeCanvas, Content: "not a canvas at all", Source: "x.md"}
	if _, _, r := parseMorphs([]changeset.Patch{bad}); len(r) == 0 || !strings.Contains(joined(r), "invalid") {
		t.Fatalf("an invalid morphed canvas must be rejected, got: %v", r)
	}

	// A morph whose canvas id does not match the type it targets is refused.
	mismatch := changeset.Patch{Target: "gadget", Scope: changeset.ScopeCanvas, Content: widgetV2, Source: "y.md"}
	if _, _, r := parseMorphs([]changeset.Patch{mismatch}); len(r) == 0 || !strings.Contains(joined(r), "canvas id") {
		t.Fatalf("a target/canvas-id mismatch must be rejected, got: %v", r)
	}
}
