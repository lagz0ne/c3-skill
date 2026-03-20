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

// ExportOptions holds parameters for the export command.
type ExportOptions struct {
	Store     *store.Store
	OutputDir string
	JSON      bool
}

// RunExport writes all entities from the store back to markdown files.
func RunExport(opts ExportOptions, w io.Writer) error {
	entities, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("export: list entities: %w", err)
	}

	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("export: create output dir: %w", err)
	}

	// Build a lookup of parent entities for path resolution
	parentSlug := make(map[string]string) // id -> "c3-N-slug"
	for _, e := range entities {
		if e.Type == "container" {
			parentSlug[e.ID] = fmt.Sprintf("%s-%s", e.ID, e.Slug)
		}
	}

	count := 0
	for _, e := range entities {
		path := entityExportPath(opts.OutputDir, e, parentSlug)
		if path == "" {
			continue
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("export: mkdir %s: %w", dir, err)
		}

		content := buildExportContent(opts.Store, e)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("export: write %s: %w", path, err)
		}
		count++
	}

	// Export code map
	codeMap, err := opts.Store.AllCodeMap()
	if err == nil && len(codeMap) > 0 {
		cmPath := filepath.Join(opts.OutputDir, "code-map.yaml")
		cmContent := buildCodeMapYAML(codeMap)
		if err := os.WriteFile(cmPath, []byte(cmContent), 0644); err != nil {
			return fmt.Errorf("export: write code-map: %w", err)
		}
	}

	fmt.Fprintf(w, "Exported %d entities to %s\n", count, opts.OutputDir)
	return nil
}

// entityExportPath determines the output file path for an entity.
func entityExportPath(outDir string, e *store.Entity, parentSlug map[string]string) string {
	switch e.Type {
	case "system":
		return filepath.Join(outDir, "README.md")
	case "container":
		dirName := fmt.Sprintf("%s-%s", e.ID, e.Slug)
		return filepath.Join(outDir, dirName, "README.md")
	case "component":
		parentDir := parentSlug[e.ParentID]
		if parentDir == "" {
			parentDir = "orphans"
		}
		fileName := fmt.Sprintf("%s-%s.md", e.ID, e.Slug)
		return filepath.Join(outDir, parentDir, fileName)
	case "ref":
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join(outDir, "refs", fileName)
	case "adr":
		slug := e.Slug
		if e.Date != "" {
			fileName := fmt.Sprintf("adr-%s-%s.md", e.Date, slug)
			return filepath.Join(outDir, "adr", fileName)
		}
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join(outDir, "adr", fileName)
	case "recipe":
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join(outDir, "recipes", fileName)
	case "rule":
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join(outDir, "rules", fileName)
	default:
		return ""
	}
}

// buildExportContent constructs the YAML frontmatter + body for an entity.
func buildExportContent(s *store.Store, e *store.Entity) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: %s\n", e.ID))
	b.WriteString(fmt.Sprintf("title: %s\n", e.Title))
	if e.Type != "system" {
		b.WriteString(fmt.Sprintf("type: %s\n", e.Type))
	}
	if e.Category != "" {
		b.WriteString(fmt.Sprintf("category: %s\n", e.Category))
	}
	if e.ParentID != "" {
		b.WriteString(fmt.Sprintf("parent: %s\n", e.ParentID))
	}
	if e.Goal != "" {
		b.WriteString(fmt.Sprintf("goal: %s\n", e.Goal))
	}
	if e.Summary != "" {
		b.WriteString(fmt.Sprintf("summary: %s\n", e.Summary))
	}
	if e.Boundary != "" {
		b.WriteString(fmt.Sprintf("boundary: %s\n", e.Boundary))
	}
	if e.Status != "" && e.Status != "active" {
		b.WriteString(fmt.Sprintf("status: %s\n", e.Status))
	}
	if e.Date != "" {
		b.WriteString(fmt.Sprintf("date: \"%s\"\n", e.Date))
	}
	if e.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", e.Description))
	}

	// Relationships
	rels, _ := s.RelationshipsFrom(e.ID)
	relsByType := make(map[string][]string)
	for _, r := range rels {
		relsByType[r.RelType] = append(relsByType[r.RelType], r.ToID)
	}
	for _, relType := range []string{"uses", "affects", "scope", "sources", "origin"} {
		if ids, ok := relsByType[relType]; ok {
			sort.Strings(ids)
			b.WriteString(fmt.Sprintf("%s: [%s]\n", relType, strings.Join(ids, ", ")))
		}
	}

	b.WriteString("---\n")

	if e.Body != "" {
		b.WriteString("\n")
		b.WriteString(e.Body)
	}

	return b.String()
}

// buildCodeMapYAML renders the code map as YAML.
func buildCodeMapYAML(codeMap map[string][]string) string {
	var b strings.Builder

	// Sort entity IDs for deterministic output
	var ids []string
	for id := range codeMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		patterns := codeMap[id]
		sort.Strings(patterns)
		b.WriteString(fmt.Sprintf("%s:\n", id))
		for _, p := range patterns {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}
	return b.String()
}
