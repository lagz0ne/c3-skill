package cmd

import (
	"fmt"
	"io"
	"slices"
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

func runSetField(entity *store.Entity, opts SetOptions, w io.Writer) error {
	// Codemap is a special field — stored in code_map table, not frontmatter.
	if opts.Field == "codemap" {
		return runSetCodemap(entity, opts, w)
	}
	if entity.Type == "component" {
		if issues := validateCurrentEntityBody(opts.Store, entity); len(issues) > 0 {
			return formatValidationError(opts.ID, issues)
		}
	}

	switch opts.Field {
	case "goal":
		entity.Goal = opts.Value
	case "status":
		entity.Status = opts.Value
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

// runSetCodemap handles codemap pattern updates: replace, append, or remove.
func runSetCodemap(entity *store.Entity, opts SetOptions, w io.Writer) error {
	if entity.Type == "component" {
		if issues := validateCurrentEntityBody(opts.Store, entity); len(issues) > 0 {
			return formatValidationError(opts.ID, issues)
		}
	}
	if opts.Append && opts.Remove {
		return fmt.Errorf("cannot use --append and --remove together")
	}

	if opts.Remove {
		existing, err := opts.Store.CodeMapFor(entity.ID)
		if err != nil {
			return fmt.Errorf("reading codemap: %w", err)
		}
		found := false
		var filtered []string
		for _, p := range existing {
			if p == opts.Value {
				found = true
			} else {
				filtered = append(filtered, p)
			}
		}
		if !found {
			return fmt.Errorf("pattern %q not found in codemap for %s", opts.Value, entity.ID)
		}
		if err := opts.Store.SetCodeMap(entity.ID, filtered); err != nil {
			return fmt.Errorf("updating codemap: %w", err)
		}
		fmt.Fprintf(w, "Removed codemap pattern %q from %s (%d remaining)\n", opts.Value, entity.ID, len(filtered))
		writeAgentHints(w, cascadeHintsForEntity(entity))
		return nil
	}

	if opts.Append {
		existing, err := opts.Store.CodeMapFor(entity.ID)
		if err != nil {
			return fmt.Errorf("reading codemap: %w", err)
		}
		if slices.Contains(existing, opts.Value) {
			fmt.Fprintf(w, "Codemap pattern %q already exists on %s\n", opts.Value, entity.ID)
			writeAgentHints(w, cascadeHintsForEntity(entity))
			return nil
		}
		existing = append(existing, opts.Value)
		if err := opts.Store.SetCodeMap(entity.ID, existing); err != nil {
			return fmt.Errorf("updating codemap: %w", err)
		}
		fmt.Fprintf(w, "Added codemap pattern %q to %s (%d total)\n", opts.Value, entity.ID, len(existing))
		return nil
	}

	// Replace all patterns (comma-separated)
	var patterns []string
	if opts.Value != "" {
		for _, p := range strings.Split(opts.Value, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				patterns = append(patterns, p)
			}
		}
	}
	if err := opts.Store.SetCodeMap(entity.ID, patterns); err != nil {
		return fmt.Errorf("updating codemap: %w", err)
	}
	fmt.Fprintf(w, "Updated %s codemap (%d patterns)\n", entity.ID, len(patterns))
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
