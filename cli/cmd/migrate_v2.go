package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

type MigrateV2Options struct {
	Store  *store.Store
	DryRun bool
}

func RunMigrateV2(opts MigrateV2Options, w io.Writer) error {
	entities, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("listing entities: %w", err)
	}

	migrated, hasNodes, empty, failed := 0, 0, 0, 0
	var emptyIDs []string
	var dirtyIDs []string

	for _, e := range entities {
		nodes, err := opts.Store.NodesForEntity(e.ID)
		if err != nil {
			return fmt.Errorf("checking nodes for %s: %w", e.ID, err)
		}
		if len(nodes) > 0 {
			hasNodes++
			continue
		}

		body := opts.Store.LegacyBody(e.ID)
		if body == "" {
			empty++
			emptyIDs = append(emptyIDs, e.ID)
			continue
		}

		if hasStaleFrontmatter(body) {
			dirtyIDs = append(dirtyIDs, e.ID)
		}

		if opts.DryRun {
			fmt.Fprintf(w, "  will migrate: %s (%s)\n", e.ID, e.Title)
			migrated++
			continue
		}

		if err := content.WriteEntity(opts.Store, e.ID, body); err != nil {
			fmt.Fprintf(w, "  FAILED %s: %v\n", e.ID, err)
			failed++
			continue
		}
		fmt.Fprintf(w, "  migrated: %s\n", e.ID)
		migrated++
	}

	fmt.Fprintln(w)
	if opts.DryRun {
		fmt.Fprintf(w, "dry-run: %d to migrate", migrated)
	} else {
		fmt.Fprintf(w, "%d migrated", migrated)
	}
	if hasNodes > 0 {
		fmt.Fprintf(w, ", %d already have nodes (ok)", hasNodes)
	}
	if failed > 0 {
		fmt.Fprintf(w, ", %d failed", failed)
	}
	fmt.Fprintln(w)

	if len(dirtyIDs) > 0 {
		fmt.Fprintf(w, "\nWARNING: %d entities had stale frontmatter in body (auto-stripped during migration).\n", len(dirtyIDs))
		fmt.Fprintln(w, "Review and rewrite with accurate content:")
		for _, id := range dirtyIDs {
			fmt.Fprintf(w, "  c3x read %s        # review current content\n", id)
			fmt.Fprintf(w, "  c3x write %s       # pipe corrected markdown\n", id)
		}
	}

	if empty > 0 {
		fmt.Fprintf(w, "\n%d entities have no content yet:\n", empty)
		for _, id := range emptyIDs {
			fmt.Fprintf(w, "  c3x write %s\n", id)
		}
	}

	return nil
}

func hasStaleFrontmatter(body string) bool {
	if strings.HasPrefix(body, "---\n") {
		return true
	}
	lines := strings.SplitN(body, "\n", 5)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			return false
		}
		idx := strings.Index(l, ":")
		if idx > 0 && !strings.Contains(l[:idx], " ") {
			return true
		}
	}
	return false
}
