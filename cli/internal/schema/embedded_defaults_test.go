package schema

import (
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestBuiltInDefinitionsLoadFromEmbeddedCanvasDocuments(t *testing.T) {
	entries, err := fs.ReadDir(builtInCanvasFS, builtInCanvasDir)
	if err != nil {
		t.Fatalf("read embedded built-in canvases: %v", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(entry.Name(), ".md"))
	}
	slices.Sort(ids)

	want := BuiltInDefinitionIDs()
	if !slices.Equal(ids, want) {
		t.Fatalf("embedded canvas ids = %v, want built-in ids %v", ids, want)
	}
	legacyIDs := []string{
		"adr",
		"atomic-design-change",
		"component",
		"container",
		"pm-requirement",
		"prd",
		"recipe",
		"ref",
		"rule",
		"system",
		"user-story",
	}
	if !slices.Equal(want, legacyIDs) {
		t.Fatalf("built-in ids = %v, want legacy ids %v", want, legacyIDs)
	}

	for _, id := range ids {
		path := builtInCanvasDir + "/" + id + ".md"
		data, err := fs.ReadFile(builtInCanvasFS, path)
		if err != nil {
			t.Fatalf("read embedded canvas %q: %v", id, err)
		}
		parsed, err := ParseCanvasDocument(path, string(data))
		if err != nil {
			t.Fatalf("parse embedded canvas %q: %v", id, err)
		}
		parsed.Source = "built-in"

		def, ok := DefinitionFor(id)
		if !ok {
			t.Fatalf("DefinitionFor(%q): not found", id)
		}
		if !reflect.DeepEqual(def, parsed) {
			t.Errorf("DefinitionFor(%q) does not match parsed embedded canvas document", id)
		}
		if def.Source != "built-in" {
			t.Errorf("DefinitionFor(%q).Source = %q, want built-in", id, def.Source)
		}
		if strings.TrimSpace(def.Description) == "" {
			t.Errorf("DefinitionFor(%q).Description is empty", id)
		}
		if len(def.Sections) == 0 {
			t.Errorf("DefinitionFor(%q).Sections is empty", id)
		}
	}
}
