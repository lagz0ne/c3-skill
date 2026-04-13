package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
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
	excludes, exErr := opts.Store.Excludes()
	if err == nil && (len(codeMap) > 0 || len(excludes) > 0) {
		cmPath := filepath.Join(opts.OutputDir, "code-map.yaml")
		if exErr != nil {
			return fmt.Errorf("export: read excludes: %w", exErr)
		}
		cmContent := buildCodeMapYAML(codeMap, excludes)
		if err := os.WriteFile(cmPath, []byte(cmContent), 0644); err != nil {
			return fmt.Errorf("export: write code-map: %w", err)
		}
	}

	fmt.Fprintf(w, "Exported %d entities to %s\n", count, opts.OutputDir)
	return nil
}

// entityExportPath determines the output file path for an entity.
func entityExportPath(outDir string, e *store.Entity, parentSlug map[string]string) string {
	rel := entityRelativePath(e, parentSlug)
	if rel == "" {
		return ""
	}
	return filepath.Join(outDir, rel)
}

func entityRelativePath(e *store.Entity, parentSlug map[string]string) string {
	switch e.Type {
	case "system":
		return "README.md"
	case "container":
		dirName := fmt.Sprintf("%s-%s", e.ID, e.Slug)
		return filepath.Join(dirName, "README.md")
	case "component":
		parentDir := parentSlug[e.ParentID]
		if parentDir == "" {
			parentDir = "orphans"
		}
		fileName := fmt.Sprintf("%s-%s.md", e.ID, e.Slug)
		return filepath.Join(parentDir, fileName)
	case "ref":
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join("refs", fileName)
	case "adr":
		slug := e.Slug
		if e.Date != "" {
			fileName := fmt.Sprintf("adr-%s-%s.md", canonicalDateSlug(e.Date), slug)
			return filepath.Join("adr", fileName)
		}
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join("adr", fileName)
	case "recipe":
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join("recipes", fileName)
	case "rule":
		fileName := fmt.Sprintf("%s.md", e.ID)
		return filepath.Join("rules", fileName)
	default:
		return ""
	}
}

// buildExportContent constructs the YAML frontmatter + body for an entity.
func buildExportContent(s *store.Store, e *store.Entity) string {
	metadata := parseMetadataMap(e.Metadata)
	rels, _ := s.RelationshipsFrom(e.ID)
	relsByType := make(map[string][]string)
	for _, r := range rels {
		relsByType[r.RelType] = append(relsByType[r.RelType], r.ToID)
	}
	body, err := content.ReadEntity(s, e.ID)
	if err != nil {
		body = ""
	}
	return renderCanonicalDoc(canonicalDoc{
		ID:            e.ID,
		Title:         e.Title,
		Type:          e.Type,
		Category:      e.Category,
		ParentID:      e.ParentID,
		Goal:          e.Goal,
		Boundary:      e.Boundary,
		Status:        e.Status,
		Date:          e.Date,
		Body:          body,
		C3Version:     metadata["c3-version"],
		Summary:       metadata["summary"],
		Description:   metadata["description"],
		Relationships: relsByType,
		Extra:         copyMetadataExcludingKnown(metadata),
	}, true)
}

// buildCodeMapYAML renders the code map as YAML.
func buildCodeMapYAML(codeMap map[string][]string, excludes []string) string {
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
	if len(excludes) > 0 {
		sort.Strings(excludes)
		b.WriteString("_exclude:\n")
		for _, p := range excludes {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}
	return b.String()
}
