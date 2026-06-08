package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// Slice 9 core — MaterializeDefinitions writes definitions to
// .c3/canvases/<id>.md as sealed canonical markdown, WRITE-IF-ABSENT: it never
// overwrites a definition a user already owns (the freeze guarantee), and what
// it writes round-trips back through ParseCanvasDocument.
func TestMaterializeDefinitions_WriteIfAbsentAndRoundTrips(t *testing.T) {
	dir := t.TempDir()
	canvasDir := filepath.Join(dir, schema.CanvasesDir)
	if err := os.MkdirAll(canvasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// a user-owned definition that must NOT be touched (frozen)
	frozen := filepath.Join(canvasDir, "component.md")
	if err := os.WriteFile(frozen, []byte("USER OWNED — do not touch"), 0o644); err != nil {
		t.Fatal(err)
	}

	refDef, _ := schema.DefinitionFor("ref")
	canvases := []schema.Canvas{
		{ID: "component", Description: "embedded seed", Domain: "software", Sections: mustSections(t, "component")},
		{ID: "ref", Description: "reference rationale definition", Domain: "software", Sections: refDef.Sections},
	}

	written, err := MaterializeDefinitions(dir, canvases)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 1 || written[0] != "ref" {
		t.Fatalf("written = %v; want only [ref] (component is frozen)", written)
	}

	// frozen file untouched
	if got, _ := os.ReadFile(frozen); string(got) != "USER OWNED — do not touch" {
		t.Errorf("frozen component.md was overwritten: %q", got)
	}

	// the written ref definition round-trips
	data, err := os.ReadFile(filepath.Join(canvasDir, "ref.md"))
	if err != nil {
		t.Fatal(err)
	}
	canvas, err := schema.ParseCanvasDocument("canvases/ref.md", string(data))
	if err != nil {
		t.Fatalf("materialized ref.md does not round-trip: %v", err)
	}
	if canvas.ID != "ref" || len(canvas.Sections) == 0 {
		t.Errorf("round-tripped canvas wrong: id=%q sections=%d", canvas.ID, len(canvas.Sections))
	}
}

func TestMaterializeDefinitions_WritesAllBuiltIns(t *testing.T) {
	dir := t.TempDir()
	defs, err := schema.AllDefinitions("")
	if err != nil {
		t.Fatal(err)
	}

	written, err := MaterializeDefinitions(dir, defs)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(written)
	want := schema.BuiltInDefinitionIDs()
	if !slices.Equal(written, want) {
		t.Fatalf("written = %v; want all built-ins %v", written, want)
	}

	for _, id := range want {
		data, err := os.ReadFile(filepath.Join(dir, schema.CanvasesDir, id+".md"))
		if err != nil {
			t.Fatalf("read materialized %s: %v", id, err)
		}
		got, err := schema.ParseCanvasDocument(filepath.ToSlash(filepath.Join(schema.CanvasesDir, id+".md")), string(data))
		if err != nil {
			t.Fatalf("parse materialized %s: %v", id, err)
		}
		got.Source = "built-in"
		embedded, ok := schema.DefinitionFor(id)
		if !ok {
			t.Fatalf("DefinitionFor(%q): not found", id)
		}
		if !reflect.DeepEqual(got, embedded) {
			t.Errorf("materialized %s does not round-trip to embedded definition", id)
		}
	}

	adr, _ := schema.DefinitionFor("adr")
	if len(adr.Reject.Bullets) == 0 || adr.Reject.Workorder == "" {
		t.Fatal("materialized built-ins must preserve ADR rejection contract")
	}
}

func mustSections(t *testing.T, entityType string) []schema.SectionDef {
	t.Helper()
	def, ok := schema.DefinitionFor(entityType)
	if !ok {
		t.Fatalf("no definition for %q", entityType)
	}
	return def.Sections
}
