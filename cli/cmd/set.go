package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// SetOptions holds parameters for the set command.
type SetOptions struct {
	Store  *store.Store
	C3Dir  string
	ID     string
	Field  string
	Value  string
	Append bool
	Remove bool
}

// RunSet updates a frontmatter field on an entity.
func RunSet(opts SetOptions, w io.Writer) error {
	entity, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("entity %q not found", opts.ID)
	}
	return runSetField(entity, opts, w)
}

// ResolveSetArgs normalizes set arguments the same way the command runner does.
func ResolveSetArgs(opts Options) (id, field, value string) {
	if len(opts.Args) >= 1 {
		id = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		value = opts.Args[1]
	}
	field = opts.Field
	if field == "" && len(opts.Args) >= 2 {
		field = opts.Args[1]
		if len(opts.Args) >= 3 {
			value = opts.Args[2]
		}
	}
	return id, field, value
}

func runSetField(entity *store.Entity, opts SetOptions, w io.Writer) error {
	if entity.Type == "component" {
		if issues := validateCurrentEntityBody(opts.Store, entity); len(issues) > 0 {
			return formatValidationError(opts.ID, issues)
		}
	}

	switch opts.Field {
	case "goal":
		entity.Goal = opts.Value
	case "status":
		// The status command is the manual legal-jump path. Status is edit-proof:
		// it is NOT written by UpdateEntity, so it must move through the
		// dedicated SetEntityStatus writer.
		if !statusTransitionLegal(entity.Status, opts.Value) {
			next := legalNextStates(entity.Status)
			if len(next) == 0 {
				return fmt.Errorf("error: cannot transition %s from %q (terminal) to %q\nhint: this status is terminal and has no legal next state", opts.ID, entity.Status, opts.Value)
			}
			return fmt.Errorf("error: cannot transition %s from %q directly to %q\nhint: legal next state(s): %s; run: c3x set %s status %s",
				opts.ID, entity.Status, opts.Value, strings.Join(next, ", "), opts.ID, next[0])
		}
		if err := opts.Store.SetEntityStatus(opts.ID, opts.Value); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}
		fmt.Fprintf(w, "Updated %s field %q\n", opts.ID, opts.Field)
		writeAgentHints(w, cascadeHintsForEntity(entity))
		return nil
	case "boundary":
		entity.Boundary = opts.Value
	case "category":
		entity.Category = opts.Value
	case "title":
		entity.Title = opts.Value
	case "date":
		entity.Date = opts.Value
	default:
		return fmt.Errorf("unknown field %q", opts.Field)
	}

	if err := opts.Store.UpdateEntity(entity); err != nil {
		return fmt.Errorf("updating entity: %w", err)
	}

	fmt.Fprintf(w, "Updated %s field %q\n", opts.ID, opts.Field)
	writeAgentHints(w, cascadeHintsForEntity(entity))
	return nil
}

func validateCurrentEntityBody(s *store.Store, entity *store.Entity) []Issue {
	body, err := content.ReadEntity(s, entity.ID)
	if err != nil {
		return nil
	}
	return validateBodyContent(body, entity.Type)
}
