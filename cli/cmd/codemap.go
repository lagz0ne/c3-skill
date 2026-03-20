package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// CodemapOptions holds parameters for the codemap command.
type CodemapOptions struct {
	C3Dir string
	Store *store.Store
	JSON  bool
}

// CodemapResult is the output of the codemap scaffold command.
type CodemapResult struct {
	File     string   `json:"file"`
	Added    []string `json:"added"`
	Existing []string `json:"existing"`
}

// RunCodemap scaffolds or updates code-map entries in the store for all
// components and refs. Existing entries are preserved.
func RunCodemap(opts CodemapOptions, w io.Writer) error {
	allEntities, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("listing entities: %w", err)
	}

	var components, refs, rules []*store.Entity
	for _, e := range allEntities {
		switch e.Type {
		case "component":
			components = append(components, e)
		case "ref":
			refs = append(refs, e)
		case "rule":
			rules = append(rules, e)
		}
	}
	sort.Slice(components, func(i, j int) bool { return components[i].ID < components[j].ID })
	sort.Slice(refs, func(i, j int) bool { return refs[i].ID < refs[j].ID })
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	existingCM, _ := opts.Store.AllCodeMap()

	var added, existing []string
	for _, e := range append(append(components, refs...), rules...) {
		if _, ok := existingCM[e.ID]; ok {
			existing = append(existing, e.ID)
		} else {
			// Add empty entry
			if err := opts.Store.SetCodeMap(e.ID, []string{}); err != nil {
				return fmt.Errorf("setting code map for %s: %w", e.ID, err)
			}
			added = append(added, e.ID)
		}
	}

	// Also write code-map.yaml file for backward compat
	cmPath := filepath.Join(opts.C3Dir, "code-map.yaml")
	updatedCM, _ := opts.Store.AllCodeMap()
	if err := writeCodeMapFromStore(cmPath, components, refs, rules, updatedCM, opts.Store); err != nil {
		return fmt.Errorf("write code-map: %w", err)
	}

	result := CodemapResult{
		File:     cmPath,
		Added:    sliceOrEmpty(added),
		Existing: sliceOrEmpty(existing),
	}

	if opts.JSON || os.Getenv("HUMAN") == "" {
		return writeJSON(w, result)
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

// writeCodeMapFromStore serialises the code-map to disk from store data.
func writeCodeMapFromStore(path string, components, refs, rules []*store.Entity, cm map[string][]string, s *store.Store) error {
	var sb strings.Builder

	sb.WriteString("# C3 code-map: maps component, ref, and rule IDs to source file glob patterns.\n")
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

	if len(rules) > 0 {
		sb.WriteString("\n# Rules\n")
		for _, e := range rules {
			writeCodeMapEntry(&sb, e.ID, cm[e.ID])
		}
	}

	// Preserve _exclude
	excl, _ := s.Excludes()
	if len(excl) > 0 {
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
