package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// MigrateV2Options configures the migrate-v2 command.
type MigrateV2Options struct {
	Store  *store.Store
	DryRun bool
}

// RunMigrateV2 recomputes node trees for entities that don't have any yet.
// Entities that already have nodes are skipped.
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

		// No nodes and no Body field — nothing to migrate.
		skipped++
	}

	if opts.DryRun {
		fmt.Fprintf(w, "\ndry-run: would migrate %d entities, skipped %d (already have nodes or no content)\n", migrated, skipped)
	} else {
		fmt.Fprintf(w, "\nmigrated %d entities, skipped %d (already have nodes or no content)\n", migrated, skipped)
	}
	return nil
}
