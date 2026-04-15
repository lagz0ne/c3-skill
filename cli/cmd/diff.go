package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// RunDiff renders the changelog or marks entries as committed.
func RunDiff(s *store.Store, mark bool, commitHash string, jsonOut bool, w io.Writer) error {
	if mark {
		if err := s.MarkChangelog(commitHash); err != nil {
			return fmt.Errorf("mark changelog: %w", err)
		}
		fmt.Fprintf(w, "Changelog marked with commit %s\n", commitHash)
		return nil
	}

	changes, err := s.UnmarkedChanges()
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}

	if len(changes) == 0 {
		fmt.Fprintln(w, "No uncommitted changes.")
		writeAgentHints(w, cascadeReviewHints())
		return nil
	}

	if jsonOut {
		if isAgentMode() {
			return writeJSON(w, struct {
				Changes []*store.ChangeEntry `json:"changes"`
				Help    []HelpHint           `json:"help,omitempty"`
			}{
				Changes: changes,
				Help:    agentHints(cascadeReviewHints()),
			})
		}
		return writeJSON(w, changes)
	}

	// Group consecutive entries by entity_id
	type group struct {
		entityID string
		entries  []*store.ChangeEntry
	}
	var groups []group
	for _, c := range changes {
		if len(groups) == 0 || groups[len(groups)-1].entityID != c.EntityID {
			groups = append(groups, group{entityID: c.EntityID, entries: []*store.ChangeEntry{c}})
		} else {
			groups[len(groups)-1].entries = append(groups[len(groups)-1].entries, c)
		}
	}

	for _, g := range groups {
		first := g.entries[0]
		switch first.Action {
		case "add":
			fmt.Fprintf(w, "+ ADDED %s\n", g.entityID)
		case "delete":
			fmt.Fprintf(w, "- DELETED %s\n", g.entityID)
		default:
			// For modify/update actions, or mixed groups
			hasAdd := false
			hasDelete := false
			for _, e := range g.entries {
				if e.Action == "add" {
					hasAdd = true
				}
				if e.Action == "delete" {
					hasDelete = true
				}
			}
			if hasAdd {
				fmt.Fprintf(w, "+ ADDED %s\n", g.entityID)
			} else if hasDelete {
				fmt.Fprintf(w, "- DELETED %s\n", g.entityID)
			} else {
				fmt.Fprintf(w, "~ MODIFIED %s\n", g.entityID)
			}
			for _, e := range g.entries {
				if e.Action == "update" && e.Field != "" {
					fmt.Fprintf(w, "    %s: %q -> %q\n", e.Field, e.OldValue, e.NewValue)
				}
			}
		}
	}

	writeAgentHints(w, cascadeReviewHints())
	return nil
}
