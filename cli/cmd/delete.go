package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// DeleteOptions holds options for the delete command.
type DeleteOptions struct {
	C3Dir  string
	ID     string
	Store  *store.Store
	DryRun bool
}

// RunDelete removes an entity and cleans all references to it.
func RunDelete(opts DeleteOptions, w io.Writer) error {
	id := opts.ID
	if id == "" {
		return fmt.Errorf("usage: c3x delete <id> [--dry-run]")
	}

	// Safety: refuse to delete context root
	if id == "c3-0" {
		return fmt.Errorf("refusing to delete c3-0 (context root)")
	}

	// Find entity
	entity, err := opts.Store.GetEntity(id)
	if err != nil {
		return fmt.Errorf("entity %q not found", id)
	}

	// Safety: refuse to delete containers with children
	children, _ := opts.Store.Children(id)
	if len(children) > 0 {
		var ids []string
		for _, c := range children {
			ids = append(ids, c.ID)
		}
		return fmt.Errorf("refusing to delete %s: has %d children (%s) — delete them first",
			id, len(children), strings.Join(ids, ", "))
	}

	prefix := ""
	if opts.DryRun {
		prefix = "[dry-run] "
	}

	// Find entities that reference this one via relationships
	inbound, _ := opts.Store.RelationshipsTo(id)
	for _, rel := range inbound {
		fmt.Fprintf(w, "%sRemove relationship %s -[%s]-> %s\n", prefix, rel.FromID, rel.RelType, id)
		if !opts.DryRun {
			_ = opts.Store.RemoveRelationship(rel)

			// Also clean up the body of the referencing entity (table rows, etc.)
			refEntity, err := opts.Store.GetEntity(rel.FromID)
			if err != nil {
				continue
			}
			// Clean "Related Refs" table rows where Ref=id
			_ = removeTableRowStore(opts.Store, refEntity, "Related Refs", "Ref", id)
			// Clean "Related Rules" table rows where Rule=id
			if strings.HasPrefix(id, "rule-") {
				_ = removeTableRowStore(opts.Store, refEntity, "Related Rules", "Rule", id)
			}
			// Clean "Components" table if this is a component being deleted
			if entity.Type == "component" {
				_ = removeTableRowStore(opts.Store, refEntity, "Components", "ID", id)
			}
		}
	}

	// If component: remove row from parent container's Components table
	if entity.ParentID != "" {
		parentEntity, err := opts.Store.GetEntity(entity.ParentID)
		if err == nil && parentEntity.Type == "container" {
			fmt.Fprintf(w, "%sRemove %s from %s Components table\n", prefix, id, entity.ParentID)
			if !opts.DryRun {
				_ = removeTableRowStore(opts.Store, parentEntity, "Components", "ID", id)
			}
		}
	}

	// Remove code-map entries (cascaded by FK, but log it)
	codeMapGlobs, _ := opts.Store.CodeMapFor(id)
	if len(codeMapGlobs) > 0 {
		fmt.Fprintf(w, "%sRemove %s from code-map\n", prefix, id)
	}

	// Delete the entity (cascades relationships and code-map via FK)
	fmt.Fprintf(w, "%sDelete %s (%s)\n", prefix, id, entity.Type)
	if !opts.DryRun {
		if err := opts.Store.DeleteEntity(id); err != nil {
			return fmt.Errorf("delete entity: %w", err)
		}
	}

	if opts.DryRun {
		fmt.Fprintf(w, "\nDry run complete — no changes made\n")
	} else {
		fmt.Fprintf(w, "Deleted %s\n", id)
	}

	return nil
}
