package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// canvasDocRequiring builds a minimal canvas document whose listed sections are all
// required text sections — enough to drive the canvas/morph gates in a test.
func canvasDocRequiring(id string, sections ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\nid: %s\ntype: canvas\ndescription: %s canvas.\n---\n\ndomain: test\nsections:\n", id, id)
	for _, s := range sections {
		fmt.Fprintf(&b, "    - name: %s\n      content_type: text\n      required: true\n", s)
	}
	return b.String()
}

func mustParseCanvas(t *testing.T, doc string) schema.Canvas {
	t.Helper()
	c, err := schema.ParseCanvasDocument("canvases/x.md", doc)
	if err != nil {
		t.Fatalf("parse canvas: %v", err)
	}
	return c
}

// writeCanvasFile seeds a project canvas on disk (sealed, as `canvas write` renders it).
func writeCanvasFile(t *testing.T, c3Dir, id, doc string) string {
	t.Helper()
	canvas := mustParseCanvas(t, doc)
	path := filepath.Join(c3Dir, schema.CanvasesDir, id+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(renderCanvasDoc(canvas, true)), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// The canvas gate must validate a migrated instance against the shape this unit is
// MORPHING the type to — not the stale on-disk canvas. A body valid under the new
// shape but not the old must pass when the type is being morphed in the same unit.
func TestCanvasGate_UsesMorphedShapeForType(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	writeCanvasFile(t, c3Dir, "policy", canvasDocRequiring("policy", "Basis")) // on-disk requires Basis

	// A create whose body carries Detail (the NEW required section), not Basis.
	p := changeset.Patch{Target: "policy-9", Scope: changeset.ScopeWhole, Type: "policy", Content: "## Detail\n\nthe detail.\n", Source: "c.md"}

	// Against the on-disk canvas (requires Basis) → rejected.
	if err := canvasGate(s, c3Dir, p, nil); err == nil {
		t.Fatal("on-disk canvas requires Basis; a body lacking it must be rejected")
	}
	// Against the in-unit morphed canvas (requires Detail) → accepted.
	morphed := map[string]schema.Canvas{"policy": mustParseCanvas(t, canvasDocRequiring("policy", "Detail"))}
	if err := canvasGate(s, c3Dir, p, morphed); err != nil {
		t.Fatalf("morphed shape requires Detail and the body has it — must pass: %v", err)
	}
}

// applyCanvasMorphs is reversible: the restore closure reverts an overwritten canvas
// to its prior bytes and removes a canvas it newly created — so a failed store apply
// can roll the file side back and the unit lands all-or-nothing.
func TestApplyCanvasMorphs_RestoreReverts(t *testing.T) {
	_, c3Dir := openStoreC3(t)
	path := writeCanvasFile(t, c3Dir, "policy", canvasDocRequiring("policy", "Basis"))
	orig, _ := os.ReadFile(path)

	morphed := map[string]schema.Canvas{
		"policy": mustParseCanvas(t, canvasDocRequiring("policy", "Detail")), // overwrite existing
		"gizmo":  mustParseCanvas(t, canvasDocRequiring("gizmo", "Spec")),    // brand new file
	}
	restore, err := applyCanvasMorphs(c3Dir, morphed)
	if err != nil {
		t.Fatal(err)
	}
	if now, _ := os.ReadFile(path); !strings.Contains(string(now), "Detail") {
		t.Fatal("policy canvas was not morphed on disk")
	}
	gizmoPath := filepath.Join(c3Dir, schema.CanvasesDir, "gizmo.md")
	if _, err := os.Stat(gizmoPath); err != nil {
		t.Fatal("gizmo canvas was not written")
	}

	if err := restore(); err != nil {
		t.Fatal(err)
	}
	if back, _ := os.ReadFile(path); string(back) != string(orig) {
		t.Errorf("restore did not revert the overwritten canvas")
	}
	if _, err := os.Stat(gizmoPath); !os.IsNotExist(err) {
		t.Errorf("restore should remove a newly-created canvas, but %s remains", gizmoPath)
	}
}

// The evolve-unit end to end: a canvas-scope patch reshapes a type and the same unit
// migrates its instance. Morph-alone (no migration) is refused; morph + migration in
// one unit lands the new canvas file AND the migrated instance atomically.
func TestRunChangeApply_EvolveUnit(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	canvasPath := writeCanvasFile(t, c3Dir, "policy", canvasDocRequiring("policy", "Spec"))
	mustInsert(t, s, &store.Entity{ID: "policy-1", Type: "policy", Title: "p1", Status: "active", Metadata: "{}"})
	if err := content.WriteEntity(s, "policy-1", "## Spec\n\nthe spec.\n"); err != nil {
		t.Fatal(err)
	}

	// The morphed canvas: Spec stays, Risk becomes newly required.
	policyV2 := "---\ntarget: policy\nscope: canvas\n---\n" + canvasDocRequiring("policy", "Spec", "Risk")

	// Unit A — morph alone. policy-1 lacks Risk → the morph strands it → refused, and
	// the on-disk canvas is left untouched (gates run before any write).
	writePatch(t, c3Dir, "morph-alone", "01-policy.canvas.patch.md", policyV2)
	var bufA strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "morph-alone"}, &bufA); err == nil {
		t.Fatalf("a morph that strands policy-1 must be refused; output: %s", bufA.String())
	} else if !strings.Contains(bufA.String(), "policy-1") {
		t.Errorf("rejection should name the stranded instance, got: %s", bufA.String())
	}
	if cur, _ := os.ReadFile(canvasPath); strings.Contains(string(cur), "Risk") {
		t.Fatal("a refused morph must NOT have written the canvas file")
	}

	// Unit B — morph + migrate in one unit. Insert Risk into policy-1 alongside the
	// canvas morph; both land atomically.
	e, _ := s.GetEntity("policy-1")
	base := fmt.Sprintf("policy-1@v%d:sha256:%s", e.Version, e.RootMerkle)
	writePatch(t, c3Dir, "morph-migrate", "01-policy.canvas.patch.md", policyV2)
	writePatch(t, c3Dir, "morph-migrate", "02-policy-1.insert.patch.md",
		"---\ntarget: policy-1\nscope: insert\nbase: "+base+"\n---\n## Risk\n\nnewly assessed.\n")

	var bufB strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "morph-migrate"}, &bufB); err != nil {
		t.Fatalf("morph + migration in one unit must land: %v\noutput: %s", err, bufB.String())
	}
	// Canvas file morphed on disk.
	if cur, _ := os.ReadFile(canvasPath); !strings.Contains(string(cur), "Risk") {
		t.Error("canvas file was not morphed to require Risk")
	}
	// Instance migrated: policy-1 now carries the Risk section, valid against the new shape.
	body, err := content.ReadEntity(s, "policy-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "Risk") {
		t.Errorf("policy-1 was not migrated to carry Risk, got: %q", body)
	}
	v2 := mustParseCanvas(t, canvasDocRequiring("policy", "Spec", "Risk"))
	if issues := validateBodyContentWithDefinition(body, "policy", v2.Sections); len(issues) > 0 {
		t.Errorf("migrated policy-1 should satisfy the new shape, got issues: %v", issues)
	}
	if !strings.Contains(bufB.String(), "morphed canvas policy") {
		t.Errorf("apply output should report the canvas morph, got: %s", bufB.String())
	}
}
