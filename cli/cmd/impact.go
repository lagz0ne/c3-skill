package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ImpactOptions holds parameters for the impact command.
type ImpactOptions struct {
	Store    *store.Store
	EntityID string
	Depth    int
	JSON     bool
}

// RunImpact performs transitive impact analysis on an entity.
func RunImpact(opts ImpactOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("error: impact requires an <entity-id> argument\nhint: run 'c3x impact <entity-id>' to analyze")
	}

	depth := opts.Depth
	if depth <= 0 {
		depth = 3
	}

	results, err := opts.Store.Impact(opts.EntityID, depth)
	if err != nil {
		return fmt.Errorf("impact analysis: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(w, "No affected entities found.")
		return nil
	}

	if opts.JSON {
		return writeJSON(w, results)
	}

	// Get entity title for header
	title := opts.EntityID
	if e, err := opts.Store.GetEntity(opts.EntityID); err == nil {
		title = e.Title
	}

	fmt.Fprintf(w, "Impact of %s (%s):\n", opts.EntityID, title)
	for _, r := range results {
		fmt.Fprintf(w, "  depth %d: %s [%s] %s\n", r.Depth, r.ID, r.Type, r.Title)
	}

	return nil
}
