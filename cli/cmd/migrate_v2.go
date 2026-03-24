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

	migrated := 0
	skipped := 0

	for _, e := range entities {
		nodes, err := opts.Store.NodesForEntity(e.ID)
		if err != nil {
			return fmt.Errorf("checking nodes for %s: %w", e.ID, err)
		}
		if len(nodes) > 0 {
			skipped++
			continue
		}

		body := opts.Store.LegacyBody(e.ID)
		if body == "" {
			skipped++
			continue
		}

		if opts.DryRun {
			fmt.Fprintf(w, "  would migrate: %s (%s)\n", e.ID, e.Title)
			migrated++
			continue
		}

		if err := content.WriteEntity(opts.Store, e.ID, body); err != nil {
			fmt.Fprintf(w, "  FAILED: %s — %v\n", e.ID, err)
			continue
		}
		fmt.Fprintf(w, "  migrated: %s (%s)\n", e.ID, e.Title)
		migrated++
	}

	if opts.DryRun {
		fmt.Fprintf(w, "\ndry-run: would migrate %d, skipped %d\n", migrated, skipped)
	} else {
		fmt.Fprintf(w, "\nmigrated %d, skipped %d\n", migrated, skipped)
	}
	return nil
}
