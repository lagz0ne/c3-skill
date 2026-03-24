package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// HashOptions holds parameters for the hash command.
type HashOptions struct {
	Store     *store.Store
	EntityID  string
	Recompute bool
}

// RunHash outputs the entity's root_merkle hash.
// With Recompute, recomputes from nodes and compares.
func RunHash(opts HashOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("usage: c3x hash <entity-id>")
	}

	entity, err := opts.Store.GetEntity(opts.EntityID)
	if err != nil {
		return fmt.Errorf("entity %q not found", opts.EntityID)
	}

	if !opts.Recompute {
		fmt.Fprintln(w, entity.RootMerkle)
		return nil
	}

	nodes, err := opts.Store.NodesForEntity(opts.EntityID)
	if err != nil {
		return fmt.Errorf("fetching nodes: %w", err)
	}
	computed := store.HashNodes(nodes)

	if computed == entity.RootMerkle {
		fmt.Fprintf(w, "OK  %s\n", computed)
	} else {
		fmt.Fprintf(w, "DRIFT  stored=%s  computed=%s\n", entity.RootMerkle, computed)
	}
	return nil
}
