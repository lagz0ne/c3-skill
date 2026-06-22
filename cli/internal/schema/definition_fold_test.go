package schema

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Slice 9 prep — DefinitionFor must carry Description/Domain for EVERY entity
// type, so a definition can round-trip as a materialized canvas doc (which
// ParseCanvasDocument requires to have a description).
func TestDefinitionFor_CarriesMetadata(t *testing.T) {
	for _, et := range []string{
		"context", "container", "component", "ref", "rule", "adr",
		"prd", "user-story", "atomic-design-change", "pm-requirement",
	} {
		def, ok := DefinitionFor(et)
		if !ok {
			t.Errorf("%s: not found", et)
			continue
		}
		if strings.TrimSpace(def.Description) == "" {
			t.Errorf("%s: DefinitionFor has empty Description (cannot materialize as a canvas doc)", et)
		}
	}
}

// Slice 8 — DefinitionFor becomes dir-aware: a project-local materialized
// definition at .c3/canvases/<type>.md (user-owned, frozen) overrides the
// embedded seed; absent that file, the embedded definition is the fallback.
func TestDefinitionForDir_PrefersProjectOverride(t *testing.T) {
	dir := t.TempDir()
	canvasDir := filepath.Join(dir, CanvasesDir)
	if err := os.MkdirAll(canvasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := "---\n" +
		"id: component\n" +
		"type: canvas\n" +
		"description: project-local component definition override\n" +
		"---\n" +
		"domain: software\n" +
		"sections:\n" +
		"  - name: Goal\n" +
		"    content_type: text\n" +
		"    required: true\n" +
		"  - name: Custom Project Section\n" +
		"    content_type: text\n" +
		"    required: true\n"
	if err := os.WriteFile(filepath.Join(canvasDir, "component.md"), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}

	def, ok := DefinitionForDir(dir, "component")
	if !ok {
		t.Fatal("DefinitionForDir(component): not found")
	}
	if !slices.Contains(sectionNames(def.Sections), "Custom Project Section") {
		t.Errorf("expected project-local override sections, got %v", sectionNames(def.Sections))
	}

	// No override file for ref -> embedded fallback.
	embedded, ok := DefinitionForDir(dir, "ref")
	if !ok {
		t.Fatal("DefinitionForDir(ref): embedded fallback not found")
	}
	if len(embedded.Sections) == 0 {
		t.Error("embedded ref fallback has no sections")
	}
	// Sanity: embedded fallback equals the embedded DefinitionFor.
	want, _ := DefinitionFor("ref")
	if !slices.Equal(sectionNames(embedded.Sections), sectionNames(want.Sections)) {
		t.Errorf("ref fallback diverged from embedded DefinitionFor")
	}
}

// Slice 7 — DefinitionFor is the single source of entity-type definitions.
// Document canvas types (prd, user-story, ...) are first-class entity types,
// resolvable via DefinitionFor just like structural entities.
func TestDefinitionFor_DocTypesAreEntityTypes(t *testing.T) {
	for _, et := range []string{"prd", "user-story", "atomic-design-change", "pm-requirement"} {
		def, ok := DefinitionFor(et)
		if !ok {
			t.Errorf("DefinitionFor(%q): not found; doc types must be first-class entity types", et)
			continue
		}
		if len(def.Sections) == 0 {
			t.Errorf("DefinitionFor(%q): no sections", et)
		}
	}
}

func sectionNames(secs []SectionDef) []string {
	out := make([]string, len(secs))
	for i, s := range secs {
		out[i] = s.Name
	}
	return out
}

// adr reconciliation: the adr entity definition and the c3-adr canvas resolve to
// the same shape — one source, no divergence.
func TestDefinitionFor_AdrAndCanvasSingleSource(t *testing.T) {
	adrDef, ok := DefinitionFor("adr")
	if !ok {
		t.Fatal("DefinitionFor(adr): not found")
	}
	c3adr, ok := CanvasFor("c3-adr")
	if !ok {
		t.Fatal("CanvasFor(c3-adr): not found")
	}
	if !slices.Equal(sectionNames(adrDef.Sections), sectionNames(c3adr.Sections)) {
		t.Errorf("adr definition %v diverges from c3-adr canvas %v",
			sectionNames(adrDef.Sections), sectionNames(c3adr.Sections))
	}
}

// Every built-in canvas doc type must be resolvable as an entity type via
// DefinitionFor (one registry, keyed by entity type; c3-adr aliases adr).
func TestDefinitionFor_CoversAllCanvasDocTypes(t *testing.T) {
	canvasToEntity := map[string]string{
		"c3-adr":               "adr",
		"atomic-design-change": "atomic-design-change",
		"pm-requirement":       "pm-requirement",
		"prd":                  "prd",
		"user-story":           "user-story",
	}
	for cid, et := range canvasToEntity {
		canvas, cok := CanvasFor(cid)
		def, dok := DefinitionFor(et)
		if !dok {
			t.Errorf("canvas %q (entity type %q) not resolvable via DefinitionFor", cid, et)
			continue
		}
		// the canvas and the entity definition must be the same shape
		if !slices.Equal(sectionNames(canvas.Sections), sectionNames(def.Sections)) {
			t.Errorf("canvas %q sections %v diverge from DefinitionFor(%q) %v",
				cid, sectionNames(canvas.Sections), et, sectionNames(def.Sections))
		}
		_ = cok
	}
}
