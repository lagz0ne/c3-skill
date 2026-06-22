package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// SupersedeOptions holds parameters for the supersede command.
type SupersedeOptions struct {
	Store *store.Store
	NewID string // the successor doc that supersedes
	OldID string // the existing terminal doc being superseded
}

// supersedeTerminalStatuses are the states a doc must be in to be superseded.
// A supersede flips a frozen/historical decision under a successor; superseding
// a still-open decision would lose its in-flight status, so it is rejected.
// Legacy ADR terminals are included so historical ADRs can still be superseded.
var supersedeTerminalStatuses = map[string]bool{
	"done":        true,
	"superseded":  true,
	"implemented": true,
	"provisioned": true,
}

// RunSupersede records that NewID supersedes OldID: it flips the old (terminal)
// doc to `superseded`, writes the backlink NewID --supersedes--> OldID, and
// rejects a non-terminal target or a supersede that would form a cycle.
//
// This is a mechanical operation: it never judges whether NewID is a legitimate
// successor or whether superseding is the right move.
func RunSupersede(opts SupersedeOptions, w io.Writer) error {
	if opts.NewID == "" || opts.OldID == "" {
		return fmt.Errorf("error: usage: c3x supersede <new-id> <old-id>\nhint: run c3x list --include-adr to choose the successor and terminal target ids")
	}
	if opts.NewID == opts.OldID {
		return fmt.Errorf("error: cannot supersede %s with itself (cycle)\nhint: pass a distinct successor id and old terminal doc id", opts.OldID)
	}

	if _, err := opts.Store.GetEntity(opts.NewID); err != nil {
		return fmt.Errorf("error: successor %q not found\nhint: run c3x search %q or c3x list --include-adr to find the successor id", opts.NewID, opts.NewID)
	}
	old, err := opts.Store.GetEntity(opts.OldID)
	if err != nil {
		return fmt.Errorf("error: target %q not found\nhint: run c3x search %q or c3x list --include-adr to find the terminal target id", opts.OldID, opts.OldID)
	}

	if !supersedeTerminalStatuses[old.Status] {
		return fmt.Errorf("error: cannot supersede %s: status %q is not terminal\nhint: only a terminal (done/superseded) doc can be superseded; finish it first", opts.OldID, old.Status)
	}

	// Cycle guard: walk the existing supersede chain reachable from OldID. If
	// NewID is reachable, adding NewID -> OldID would close a cycle.
	if supersedeReaches(opts.Store, opts.OldID, opts.NewID) {
		return fmt.Errorf("error: cannot supersede %s with %s: would form a supersede cycle\nhint: run c3x graph %s --direction forward to inspect the existing supersede chain", opts.OldID, opts.NewID, opts.OldID)
	}

	// Flip the old doc to superseded via the *->superseded edge; status is
	// edit-proof, so it moves through the dedicated status writer.
	fromStatus := old.Status
	if err := opts.Store.SetEntityStatus(opts.OldID, "superseded"); err != nil {
		return fmt.Errorf("error: flipping %s to superseded: %w", opts.OldID, err)
	}

	// Write the backlink successor --supersedes--> superseded.
	if err := opts.Store.AddRelationship(&store.Relationship{
		FromID:  opts.NewID,
		ToID:    opts.OldID,
		RelType: "supersedes",
	}); err != nil {
		return fmt.Errorf("error: linking %s -> %s: %w", opts.NewID, opts.OldID, err)
	}

	fmt.Fprintf(w, "Superseded %s with %s (%s -> superseded)\n", opts.OldID, opts.NewID, fromStatus)
	return nil
}

// supersedeReaches reports whether `target` is reachable from `start` by walking
// outbound `supersedes` edges (DFS). Used to reject a supersede that would close
// a cycle. Visited tracking guards against pre-existing cycles in the data.
func supersedeReaches(s *store.Store, start, target string) bool {
	visited := map[string]bool{}
	var dfs func(id string) bool
	dfs = func(id string) bool {
		if id == target {
			return true
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		rels, err := s.RelationshipsFrom(id)
		if err != nil {
			return false
		}
		for _, r := range rels {
			if r.RelType != "supersedes" {
				continue
			}
			if dfs(r.ToID) {
				return true
			}
		}
		return false
	}
	return dfs(start)
}
