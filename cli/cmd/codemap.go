package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// CodemapOptions holds parameters for the codemap command.
type CodemapOptions struct {
	C3Dir string
	Graph *walker.C3Graph
	JSON  bool
}

// CodemapResult is the output of the codemap scaffold command.
type CodemapResult struct {
	File     string   `json:"file"`
	Added    []string `json:"added"`
	Existing []string `json:"existing"`
}

// RunCodemap scaffolds or updates .c3/code-map.yaml with stubs for all
// components and refs in the graph. Existing entries are preserved.
func RunCodemap(opts CodemapOptions, w io.Writer) error {
	cmPath := filepath.Join(opts.C3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err != nil {
		return fmt.Errorf("code-map parse error: %w", err)
	}

	var components, refs []*walker.C3Entity
	for _, e := range opts.Graph.All() {
		switch frontmatter.ClassifyDoc(e.Frontmatter) {
		case frontmatter.DocComponent:
			components = append(components, e)
		case frontmatter.DocRef:
			refs = append(refs, e)
		}
	}
	sort.Slice(components, func(i, j int) bool { return components[i].ID < components[j].ID })
	sort.Slice(refs, func(i, j int) bool { return refs[i].ID < refs[j].ID })

	var added, existing []string
	for _, e := range append(components, refs...) {
		if _, ok := cm[e.ID]; ok {
			existing = append(existing, e.ID)
		} else {
			cm[e.ID] = []string{}
			added = append(added, e.ID)
		}
	}

	if err := writeCodeMap(cmPath, components, refs, cm); err != nil {
		return fmt.Errorf("write code-map: %w", err)
	}

	result := CodemapResult{
		File:     cmPath,
		Added:    sliceOrEmpty(added),
		Existing: sliceOrEmpty(existing),
	}

	if opts.JSON || os.Getenv("HUMAN") == "" {
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(w, string(out))
		return nil
	}

	fmt.Fprintf(w, "code-map: %s\n", cmPath)
	fmt.Fprintf(w, "  added:    %d\n", len(added))
	fmt.Fprintf(w, "  existing: %d\n", len(existing))
	if len(added) > 0 {
		fmt.Fprintln(w, "\nadded:")
		for _, id := range added {
			fmt.Fprintf(w, "  %s\n", id)
		}
	}
	return nil
}

// writeCodeMap serialises the code-map to disk, grouping components then refs.
// Existing patterns are preserved; new entries get an empty list.
func writeCodeMap(path string, components, refs []*walker.C3Entity, cm codemap.CodeMap) error {
	var sb strings.Builder

	sb.WriteString("# C3 code-map: maps component and ref IDs to source file glob patterns.\n")
	sb.WriteString("# Edit patterns, then verify with: c3x coverage\n")

	if len(components) > 0 {
		sb.WriteString("\n# Components\n")
		for _, e := range components {
			writeCodeMapEntry(&sb, e.ID, cm[e.ID])
		}
	}

	if len(refs) > 0 {
		sb.WriteString("\n# Refs\n")
		for _, e := range refs {
			writeCodeMapEntry(&sb, e.ID, cm[e.ID])
		}
	}

	// Preserve _exclude if it exists; otherwise leave a commented example.
	if excl, ok := cm["_exclude"]; ok {
		sb.WriteString("\n# Exclusions (not counted in coverage)\n")
		writeCodeMapEntry(&sb, "_exclude", excl)
	} else {
		sb.WriteString("\n# Exclusions (not counted in coverage)\n")
		sb.WriteString("# _exclude:\n")
		sb.WriteString("#   - \"**/*.test.ts\"\n")
		sb.WriteString("#   - dist/**\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func writeCodeMapEntry(sb *strings.Builder, id string, patterns []string) {
	if len(patterns) == 0 {
		sb.WriteString(id + ": []\n")
		return
	}
	sb.WriteString(id + ":\n")
	for _, p := range patterns {
		sb.WriteString("  - " + p + "\n")
	}
}

func sliceOrEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
