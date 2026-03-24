package cmd

import (
	"fmt"
	"io"

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

	if empty > 0 {
		fmt.Fprintf(w, "\n%d entities have no content yet:\n", empty)
		for _, id := range emptyIDs {
			fmt.Fprintf(w, "  %s — write content with: c3x write %s\n", id, id)
		}
	}

	return nil
}
