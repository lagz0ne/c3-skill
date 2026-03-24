package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// PruneOptions holds parameters for the prune command.
type PruneOptions struct {
	Store    *store.Store
	EntityID string
	Keep     int
}

// RunPrune prunes old versions, keeping the last N.
func RunPrune(opts PruneOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("usage: c3x prune <entity-id> --keep <n>")
	}

	if _, err := opts.Store.GetEntity(opts.EntityID); err != nil {
		return fmt.Errorf("entity %q not found", opts.EntityID)
	}

	if opts.Keep < 1 {
		return fmt.Errorf("--keep must be at least 1")
	}

	pruned, err := opts.Store.PruneVersions(opts.EntityID, opts.Keep)
	if err != nil {
		return fmt.Errorf("pruning versions: %w", err)
	}

	fmt.Fprintf(w, "Pruned %d version(s) for %s (keeping last %d)\n", pruned, opts.EntityID, opts.Keep)
	return nil
}
