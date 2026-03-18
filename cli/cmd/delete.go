package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
	"github.com/lagz0ne/c3-design/cli/internal/writer"
)

// DeleteOptions holds options for the delete command.
type DeleteOptions struct {
	C3Dir  string
	ID     string
	Graph  *walker.C3Graph
	DryRun bool
}

// RunDelete removes an entity and cleans all references to it.
func RunDelete(opts DeleteOptions, w io.Writer) error {
	id := opts.ID
	if id == "" {
		return fmt.Errorf("usage: c3x delete <id> [--dry-run]")
	}

	// Safety: refuse to delete context root
	if id == "c3-0" {
		return fmt.Errorf("refusing to delete c3-0 (context root)")
	}

	// Find entity
	entity := opts.Graph.Get(id)
	if entity == nil {
		return fmt.Errorf("entity %q not found", id)
	}

	entityPath, err := findEntityFile(opts.C3Dir, id)
	if err != nil {
		return err
	}

	// Safety: refuse to delete containers with children
	children := opts.Graph.Children(id)
	if len(children) > 0 {
		var ids []string
		for _, c := range children {
			ids = append(ids, c.ID)
		}
		return fmt.Errorf("refusing to delete %s: has %d children (%s) — delete them first",
			id, len(children), strings.Join(ids, ", "))
	}

	prefix := ""
	if opts.DryRun {
		prefix = "[dry-run] "
	}

	// Compute cleanup plan: all entities referencing this one
	referencers := opts.Graph.Reverse(id)

	for _, ref := range referencers {
		refPath, err := findEntityFile(opts.C3Dir, ref.ID)
		if err != nil {
			continue
		}

		// Clean uses[], affects[], scope[]
		for _, field := range []string{"uses", "affects", "scope"} {
			arr := getFieldValues(ref.Frontmatter, field)
			if containsStr(arr, id) {
				fmt.Fprintf(w, "%sRemove %s from %s.%s\n", prefix, id, ref.ID, field)
				if !opts.DryRun {
					if err := writer.RemoveFromArrayField(refPath, field, id); err != nil {
						fmt.Fprintf(w, "  warning: %v\n", err)
					}
				}
			}
		}

		// Clean sources[] (handles anchored refs like c3-101#section)
		if hasSourceRef(ref.Frontmatter.Sources, id) {
			fmt.Fprintf(w, "%sRemove %s from %s.sources\n", prefix, id, ref.ID)
			if !opts.DryRun {
				if err := removeSourceRefs(refPath, id); err != nil {
					fmt.Fprintf(w, "  warning: %v\n", err)
				}
			}
		}

		// Clean "Related Refs" table rows where Ref=id
		if containsStr(ref.Frontmatter.Refs, id) || ref.Frontmatter.Parent == id {
			fmt.Fprintf(w, "%sRemove %s from %s Related Refs table\n", prefix, id, ref.ID)
			if !opts.DryRun {
				_ = removeTableRow(refPath, "Related Refs", "Ref", id)
			}
		}

		// Clean parent reference (component's parent field pointing to deleted container)
		// This shouldn't happen because we refuse containers with children above
	}

	// If component: remove row from parent container's Components table
	if entity.Frontmatter.Parent != "" {
		parentEntity := opts.Graph.Get(entity.Frontmatter.Parent)
		if parentEntity != nil && parentEntity.Type == frontmatter.DocContainer {
			parentPath, err := findEntityFile(opts.C3Dir, entity.Frontmatter.Parent)
			if err == nil {
				fmt.Fprintf(w, "%sRemove %s from %s Components table\n", prefix, id, entity.Frontmatter.Parent)
				if !opts.DryRun {
					_ = removeTableRow(parentPath, "Components", "ID", id)
				}
			}
		}
	}

	// Remove entry from code-map.yaml
	codemapPath := filepath.Join(opts.C3Dir, "code-map.yaml")
	if removed, err := removeCodemapEntry(codemapPath, id, opts.DryRun); err == nil && removed {
		fmt.Fprintf(w, "%sRemove %s from code-map.yaml\n", prefix, id)
	}

	// Delete the entity file
	fmt.Fprintf(w, "%sDelete %s (%s)\n", prefix, id, entityPath)
	if !opts.DryRun {
		if err := os.Remove(entityPath); err != nil {
			return fmt.Errorf("delete file: %w", err)
		}

		// If parent directory is now empty (empty container dir), remove it
		dir := filepath.Dir(entityPath)
		if isEmptyDir(dir) && dir != opts.C3Dir {
			fmt.Fprintf(w, "Remove empty directory %s\n", dir)
			os.Remove(dir)
		}
	}

	if opts.DryRun {
		fmt.Fprintf(w, "\nDry run complete — no files modified\n")
	} else {
		fmt.Fprintf(w, "Deleted %s\n", id)
	}

	return nil
}

// getFieldValues returns the values for a known array field from frontmatter.
func getFieldValues(fm *frontmatter.Frontmatter, field string) []string {
	switch field {
	case "uses", "refs":
		return fm.Refs
	case "affects":
		return fm.Affects
	case "scope":
		return fm.Scope
	case "sources":
		return fm.Sources
	}
	return nil
}

func containsStr(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}

// hasSourceRef checks if any source references the given entity ID (with or without anchor).
func hasSourceRef(sources []string, id string) bool {
	for _, src := range sources {
		if frontmatter.StripAnchor(src) == id {
			return true
		}
	}
	return false
}

// removeSourceRefs removes all sources referencing the given ID (including anchored variants).
func removeSourceRefs(path string, id string) error {
	fm, body, err := readFMBody(path)
	if err != nil {
		return err
	}

	var filtered []string
	for _, src := range fm.Sources {
		if frontmatter.StripAnchor(src) != id {
			filtered = append(filtered, src)
		}
	}
	fm.Sources = filtered
	return writer.WriteBack(path, fm, body)
}

// readFMBody reads a file and returns parsed frontmatter + body.
func readFMBody(path string) (*frontmatter.Frontmatter, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	fm, body := frontmatter.ParseFrontmatter(string(data))
	if fm == nil {
		return nil, "", fmt.Errorf("no valid frontmatter in %s", path)
	}
	return fm, body, nil
}

// removeCodemapEntry removes an entity's entry from code-map.yaml using line-level removal.
// Returns true if an entry was found (and removed, if not dry-run).
func removeCodemapEntry(path string, id string, dryRun bool) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	found := false
	skip := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect the start of our entry: "id:" at the top level (no leading whitespace)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.HasPrefix(trimmed, id+":") {
			found = true
			skip = true
			continue
		}

		// If skipping, continue until we hit the next top-level key or end
		if skip {
			if trimmed == "" {
				// blank line — could be between entries, keep skipping
				continue
			}
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				// indented continuation of the entry we're removing
				continue
			}
			// New top-level key — stop skipping
			skip = false
		}

		result = append(result, line)
	}

	if !found {
		return false, nil
	}

	if dryRun {
		return true, nil
	}

	return true, os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// isEmptyDir checks if a directory contains no files.
func isEmptyDir(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err != nil // io.EOF means empty
}

