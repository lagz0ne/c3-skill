package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// docTypeToStoreType maps frontmatter.DocType to the store entity type string.
func docTypeToStoreType(dt frontmatter.DocType) string {
	switch dt {
	case frontmatter.DocContext:
		return "system"
	case frontmatter.DocContainer:
		return "container"
	case frontmatter.DocComponent:
		return "component"
	case frontmatter.DocRef:
		return "ref"
	case frontmatter.DocADR:
		return "adr"
	case frontmatter.DocRule:
		return "rule"
	case frontmatter.DocRecipe:
		return "recipe"
	default:
		return ""
	}
}

// RunMigrate migrates file-based .c3/ docs into a SQLite database.
// If keepOriginals is false, the original .md files and code-map.yaml are removed after migration.
func RunMigrate(c3Dir string, keepOriginals bool, w io.Writer) error {
	dbPath := filepath.Join(c3Dir, "c3.db")
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("error: %s already exists — migration already done?", dbPath)
	}

	// Walk .c3/ files
	result, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return fmt.Errorf("error: walking %s: %w", c3Dir, err)
	}

	if len(result.Warnings) > 0 {
		for _, warn := range result.Warnings {
			fmt.Fprintf(w, "warning: skipping %s (failed to parse frontmatter)\n", warn.Path)
		}
	}

	if len(result.Docs) == 0 {
		return fmt.Errorf("error: no documents found in %s", c3Dir)
	}

	// Open store
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: opening database: %w", err)
	}
	defer s.Close()

	// Track migrated file paths for cleanup
	var migratedFiles []string

	// Import entities
	entityCount := 0
	relCount := 0

	for _, doc := range result.Docs {
		fm := doc.Frontmatter
		dt := frontmatter.ClassifyDoc(fm)
		storeType := docTypeToStoreType(dt)
		if storeType == "" {
			fmt.Fprintf(w, "warning: skipping %s (unknown type)\n", doc.Path)
			continue
		}

		slug := walker.SlugFromPath(doc.Path)
		title := fm.Title
		if title == "" {
			title = fm.ID
		}

		// Serialize Extra as JSON metadata
		metadata := "{}"
		if len(fm.Extra) > 0 {
			if data, err := json.Marshal(fm.Extra); err == nil {
				metadata = string(data)
			}
		}

		entity := &store.Entity{
			ID:          fm.ID,
			Type:        storeType,
			Title:       title,
			Slug:        slug,
			Category:    fm.Category,
			ParentID:    fm.Parent,
			Goal:        fm.Goal,
			Summary:     fm.Summary,
			Description: fm.Description,
			Body:        doc.Body,
			Status:      fm.Status,
			Boundary:    fm.Boundary,
			Date:        fm.Date,
			Metadata:    metadata,
		}
		if entity.Status == "" {
			entity.Status = "active"
		}

		if err := s.InsertEntity(entity); err != nil {
			return fmt.Errorf("error: inserting entity %s: %w", fm.ID, err)
		}
		entityCount++
		migratedFiles = append(migratedFiles, doc.Path)

		fmt.Fprintf(w, "  migrated %s (%s)\n", fm.ID, storeType)
	}

	// Import relationships (second pass — all entities exist now)
	for _, doc := range result.Docs {
		fm := doc.Frontmatter
		dt := frontmatter.ClassifyDoc(fm)
		if docTypeToStoreType(dt) == "" {
			continue
		}

		// uses (fm.Refs maps to yaml field "uses:")
		for _, ref := range fm.Refs {
			if err := addRelSafe(s, fm.ID, ref, "uses"); err != nil {
				fmt.Fprintf(w, "warning: %s\n", err)
			} else {
				relCount++
			}
		}

		// affects
		for _, affected := range fm.Affects {
			if err := addRelSafe(s, fm.ID, affected, "affects"); err != nil {
				fmt.Fprintf(w, "warning: %s\n", err)
			} else {
				relCount++
			}
		}

		// scope (strip anchors)
		for _, sc := range fm.Scope {
			target := frontmatter.StripAnchor(sc)
			if err := addRelSafe(s, fm.ID, target, "scope"); err != nil {
				fmt.Fprintf(w, "warning: %s\n", err)
			} else {
				relCount++
			}
		}

		// sources (strip anchors)
		for _, src := range fm.Sources {
			target := frontmatter.StripAnchor(src)
			if err := addRelSafe(s, fm.ID, target, "sources"); err != nil {
				fmt.Fprintf(w, "warning: %s\n", err)
			} else {
				relCount++
			}
		}

		// via from Extra
		if viaVal, ok := fm.Extra["via"]; ok {
			switch v := viaVal.(type) {
			case string:
				if err := addRelSafe(s, fm.ID, v, "via"); err != nil {
					fmt.Fprintf(w, "warning: %s\n", err)
				} else {
					relCount++
				}
			case []interface{}:
				for _, item := range v {
					if vs, ok := item.(string); ok {
						if err := addRelSafe(s, fm.ID, vs, "via"); err != nil {
							fmt.Fprintf(w, "warning: %s\n", err)
						} else {
							relCount++
						}
					}
				}
			}
		}
	}

	// Import code-map
	cmPath := filepath.Join(c3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err != nil {
		fmt.Fprintf(w, "warning: failed to parse code-map.yaml: %v\n", err)
	} else {
		cmCount := 0
		for id, globs := range cm {
			if id == "_exclude" {
				for _, pattern := range globs {
					if pattern != "" {
						if err := s.AddExclude(pattern); err != nil {
							fmt.Fprintf(w, "warning: adding exclude %q: %v\n", pattern, err)
						}
					}
				}
				continue
			}
			// Filter empty globs
			var nonEmpty []string
			for _, g := range globs {
				if g != "" {
					nonEmpty = append(nonEmpty, g)
				}
			}
			if len(nonEmpty) == 0 {
				continue
			}
			// Only set code map for entities that exist in the store
			if _, err := s.GetEntity(id); err != nil {
				fmt.Fprintf(w, "warning: code-map entity %s not found in store, skipping\n", id)
				continue
			}
			if err := s.SetCodeMap(id, nonEmpty); err != nil {
				fmt.Fprintf(w, "warning: setting code map for %s: %v\n", id, err)
			} else {
				cmCount++
			}
		}
		if cmCount > 0 {
			fmt.Fprintf(w, "  imported code-map for %d entities\n", cmCount)
		}
	}

	fmt.Fprintf(w, "\nmigrated %d entities, %d relationships -> %s\n", entityCount, relCount, dbPath)

	// Cleanup if requested
	if !keepOriginals {
		// Remove migrated .md files
		for _, relPath := range migratedFiles {
			absPath := filepath.Join(c3Dir, relPath)
			os.Remove(absPath)
		}
		// Remove code-map.yaml
		os.Remove(cmPath)
		// Remove _index/ directory
		os.RemoveAll(filepath.Join(c3Dir, "_index"))
		// Remove empty directories (walk bottom-up)
		removeEmptyDirs(c3Dir)
		fmt.Fprintln(w, "removed original files (kept c3.db and config.yaml)")
	}

	return nil
}

// addRelSafe adds a relationship, ignoring foreign key errors for missing targets.
func addRelSafe(s *store.Store, fromID, toID, relType string) error {
	if toID == "" {
		return nil
	}
	err := s.AddRelationship(&store.Relationship{
		FromID:  fromID,
		ToID:    toID,
		RelType: relType,
	})
	if err != nil {
		return fmt.Errorf("relationship %s->%s (%s): %v", fromID, toID, relType, err)
	}
	return nil
}

// removeEmptyDirs removes empty directories within root, bottom-up.
// Skips root itself.
func removeEmptyDirs(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == root {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) == 0 {
			os.Remove(path)
		}
		return nil
	})
	// Second pass: parent dirs may now be empty
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == root {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) == 0 {
			os.Remove(path)
		}
		return nil
	})
}

// findC3Dir searches upward from startDir for a .c3/ directory.
// Used by commands that need to locate the project root.
func findC3Dir(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".c3")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// hasDBFile checks whether a c3.db file exists in the given .c3/ directory.
func hasDBFile(c3Dir string) bool {
	_, err := os.Stat(filepath.Join(c3Dir, "c3.db"))
	return err == nil
}

// openStoreFromDir opens the store from a .c3/ directory.
func openStoreFromDir(c3Dir string) (*store.Store, error) {
	dbPath := filepath.Join(c3Dir, "c3.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no c3.db found in %s — run 'c3x migrate' first", c3Dir)
	}
	return store.Open(dbPath)
}

// entityTypeLabel returns a human-friendly label for display.
func entityTypeLabel(t string) string {
	return strings.ToUpper(t[:1]) + t[1:]
}
