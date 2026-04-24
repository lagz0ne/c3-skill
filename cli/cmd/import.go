package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/content"
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

type ImportOptions struct {
	C3Dir         string
	Force         bool
	SkipBackup    bool
	AllowADRDrift bool
	Only          []string
}

// RunImport rebuilds c3.db from the markdown tree in c3Dir.
func RunImport(opts ImportOptions, w io.Writer) error {
	c3Dir := opts.C3Dir
	result, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return fmt.Errorf("error: walking %s: %w", c3Dir, err)
	}
	if len(result.Docs) == 0 {
		return fmt.Errorf("error: no documents found in %s", c3Dir)
	}
	for _, warn := range result.Warnings {
		fmt.Fprintf(w, "warning: skipping %s (failed to parse frontmatter)\n", warn.Path)
	}
	for _, doc := range result.Docs {
		actual, expected := verifyParsedDocSeal(doc)
		if actual == expected && actual != "" {
			continue
		}
		if opts.AllowADRDrift && isADRCanonicalPath(doc.Path) {
			continue
		}
		if len(opts.Only) > 0 && !verifyTargetMatchesDoc(opts.Only, doc.Frontmatter.ID, doc.Path) {
			continue
		}
		if opts.Force {
			fmt.Fprintf(w, "warning: resealing %s (expected %s)\n", doc.Path, expected[:12])
			continue
		}
		if actual == "" {
			return fmt.Errorf("error: unsealed C3 file %s\nhint: run 'c3x repair' to reseal canonical files, or use 'c3x import --force' for advanced recovery", doc.Path)
		}
		return fmt.Errorf("error: broken C3 seal in %s\nhint: run 'c3x check' to inspect drift, then 'c3x repair' to reseal canonical files", doc.Path)
	}

	dbPath := filepath.Join(c3Dir, "c3.db")
	if pathExists(dbPath) && !opts.Force {
		return fmt.Errorf("error: import rebuilds %s and resets DB-only history\nhint: re-run with --force to create a backup and replace the database", dbPath)
	}
	if pathExists(dbPath) && !opts.SkipBackup {
		backupPath, err := backupDBFile(dbPath)
		if err != nil {
			return fmt.Errorf("backup database: %w", err)
		}
		fmt.Fprintf(w, "Backed up %s to %s\n", dbPath, backupPath)
	}
	tmpDB := filepath.Join(c3Dir, ".c3.import.tmp.db")
	_ = os.Remove(tmpDB)
	_ = os.Remove(tmpDB + "-wal")
	_ = os.Remove(tmpDB + "-shm")

	s, err := store.Open(tmpDB)
	if err != nil {
		return fmt.Errorf("error: opening temp database: %w", err)
	}
	if err := importDocsToStore(s, c3Dir, result); err != nil {
		s.Close()
		_ = os.Remove(tmpDB)
		_ = os.Remove(tmpDB + "-wal")
		_ = os.Remove(tmpDB + "-shm")
		return err
	}
	if _, err := s.DB().Exec(`PRAGMA wal_checkpoint(FULL)`); err != nil {
		s.Close()
		return fmt.Errorf("checkpoint temp database: %w", err)
	}
	if err := s.Close(); err != nil {
		return fmt.Errorf("close temp database: %w", err)
	}

	_ = os.Remove(dbPath)
	if err := os.Rename(tmpDB, dbPath); err != nil {
		return fmt.Errorf("replace database: %w", err)
	}
	_ = os.Remove(tmpDB + "-wal")
	_ = os.Remove(tmpDB + "-shm")

	fmt.Fprintf(w, "Imported %d entities into %s\n", len(result.Docs), dbPath)
	if err := reportLayerDisconnectsAfterRebuild(dbPath, w); err != nil {
		return err
	}
	return nil
}

func reportLayerDisconnectsAfterRebuild(dbPath string, w io.Writer) error {
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open rebuilt database: %w", err)
	}
	defer s.Close()

	issues := checkLayerDisconnectsStore(s)
	if len(issues) == 0 {
		return nil
	}
	fmt.Fprintln(w, "Layer integration issues after rebuild:")
	for _, issue := range issues {
		fmt.Fprintf(w, "  ! %s: %s\n", issue.Entity, issue.Message)
		if hint := hintFor(issue.Message); hint != "" {
			fmt.Fprintf(w, "    -> %s\n", hint)
		}
	}
	return nil
}

func backupDBFile(dbPath string) (string, error) {
	src, err := os.Open(dbPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	backupPath := fmt.Sprintf("%s.bak-%s", dbPath, time.Now().UTC().Format("20060102T150405Z"))
	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return "", err
	}
	if err := dst.Close(); err != nil {
		return "", err
	}
	return backupPath, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func importDocsToStore(s *store.Store, c3Dir string, result *walker.WalkResult) error {
	entityCount := 0
	for _, doc := range result.Docs {
		fm := doc.Frontmatter
		dt := frontmatter.ClassifyDoc(fm)
		storeType := docTypeToStoreType(dt)
		if storeType == "" {
			continue
		}

		title := fm.Title
		if title == "" {
			title = fm.ID
		}

		entity := &store.Entity{
			ID:       fm.ID,
			Type:     storeType,
			Title:    title,
			Slug:     walker.SlugFromPath(doc.Path),
			Category: fm.Category,
			ParentID: fm.Parent,
			Goal:     fm.Goal,
			Status:   fm.Status,
			Boundary: fm.Boundary,
			Date:     fm.Date,
			Metadata: buildMetadataFromFrontmatter(fm.Summary, fm.Description, fm.Extra),
		}
		if entity.Status == "" {
			entity.Status = "active"
		}
		if err := s.InsertEntity(entity); err != nil {
			return fmt.Errorf("error: inserting entity %s: %w", fm.ID, err)
		}
		entityCount++
	}

	for _, doc := range result.Docs {
		fm := doc.Frontmatter
		if docTypeToStoreType(frontmatter.ClassifyDoc(fm)) == "" {
			continue
		}

		for _, rel := range []struct {
			targets []string
			relType string
			strip   bool
		}{
			{fm.Refs, "uses", false},
			{fm.Affects, "affects", false},
			{fm.Scope, "scope", true},
			{fm.Sources, "sources", true},
			{fm.Origin, "origin", true},
		} {
			for _, target := range rel.targets {
				if rel.strip {
					target = frontmatter.StripAnchor(target)
				}
				if err := addRelSafe(s, fm.ID, target, rel.relType); err != nil {
					return err
				}
			}
		}
		if viaVal, ok := fm.Extra["via"]; ok {
			switch v := viaVal.(type) {
			case string:
				if err := addRelSafe(s, fm.ID, v, "via"); err != nil {
					return err
				}
			case []any:
				for _, item := range v {
					if vs, ok := item.(string); ok {
						if err := addRelSafe(s, fm.ID, vs, "via"); err != nil {
							return err
						}
					}
				}
			}
		}

		body := strings.TrimSpace(doc.Body)
		if body == "" {
			continue
		}
		if err := content.WriteEntity(s, fm.ID, doc.Body); err != nil {
			return fmt.Errorf("error: writing content for %s: %w", fm.ID, err)
		}
	}

	cmPath := filepath.Join(c3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err == nil {
		for id, globs := range cm {
			if id == "_exclude" {
				for _, pattern := range globs {
					if pattern != "" {
						if err := s.AddExclude(pattern); err != nil {
							return fmt.Errorf("error: adding exclude %q: %w", pattern, err)
						}
					}
				}
				continue
			}
			var nonEmpty []string
			for _, g := range globs {
				if g != "" {
					nonEmpty = append(nonEmpty, g)
				}
			}
			if len(nonEmpty) == 0 {
				continue
			}
			if _, err := s.GetEntity(id); err != nil {
				continue
			}
			if err := s.SetCodeMap(id, nonEmpty); err != nil {
				return fmt.Errorf("error: setting code map for %s: %w", id, err)
			}
		}
	}

	// Rebuild should establish a clean baseline.
	if _, err := s.DB().Exec(`DELETE FROM changelog`); err != nil {
		return fmt.Errorf("error: clearing changelog: %w", err)
	}
	_ = entityCount
	return nil
}
