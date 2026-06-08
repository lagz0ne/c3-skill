package schema

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

const builtInCanvasDir = "builtin/canvases"

//go:embed builtin/canvases/*.md
var builtInCanvasFS embed.FS

var builtInDefinitions = mustLoadBuiltInDefinitions()

func mustLoadBuiltInDefinitions() map[string]Canvas {
	entries, err := fs.ReadDir(builtInCanvasFS, builtInCanvasDir)
	if err != nil {
		panic(fmt.Sprintf("load built-in canvases: %v", err))
	}
	defs := make(map[string]Canvas, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.ToSlash(filepath.Join(builtInCanvasDir, entry.Name()))
		data, err := builtInCanvasFS.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("read built-in canvas %s: %v", entry.Name(), err))
		}
		canvas, err := ParseCanvasDocument(path, string(data))
		if err != nil {
			panic(err)
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		if canvas.ID != id {
			panic(fmt.Sprintf("built-in canvas %s declares id %q", entry.Name(), canvas.ID))
		}
		if _, exists := defs[canvas.ID]; exists {
			panic(fmt.Sprintf("duplicate built-in canvas id %q", canvas.ID))
		}
		canvas.Source = "built-in"
		defs[canvas.ID] = canvas
	}
	return defs
}

func builtInDefinitionIDs() []string {
	out := make([]string, 0, len(builtInDefinitions))
	for id := range builtInDefinitions {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

var definitionAliases = map[string]string{
	"context": "system",
	"c3-adr":  "adr",
}

func canonicalDefinitionID(entityType string) string {
	entityType = strings.TrimSpace(entityType)
	if alias, ok := definitionAliases[entityType]; ok {
		return alias
	}
	return entityType
}
