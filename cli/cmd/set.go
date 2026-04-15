package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// isJSONArray checks if a string is a valid JSON array.
// Uses actual parsing instead of prefix-sniffing to avoid misrouting
// plain text that happens to start with '['.
func isJSONArray(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") {
		return false
	}
	var v []json.RawMessage
	return json.Unmarshal([]byte(s), &v) == nil
}

// SetOptions holds parameters for the set command.
type SetOptions struct {
	Store   *store.Store
	C3Dir   string
	ID      string
	Field   string
	Section string
	Value   string
	Append  bool
	Remove  bool
	Stdin   bool
}

// BatchSetPayload is the JSON format for batch set via stdin.
type BatchSetPayload struct {
	Fields   map[string]string `json:"fields,omitempty"`
	Sections map[string]string `json:"sections,omitempty"`
	Codemap  *[]string         `json:"codemap,omitempty"`
}

// RunSet updates a frontmatter field or section content on an entity.
func RunSet(opts SetOptions, w io.Writer) error {
	entity, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("entity %q not found", opts.ID)
	}

	if opts.Stdin {
		return runSetBatch(entity, opts, w)
	}

	if opts.Section != "" {
		return runSetSection(entity, opts, w)
	}
	return runSetField(entity, opts, w)
}

func runSetBatch(entity *store.Entity, opts SetOptions, w io.Writer) error {
	var payload BatchSetPayload
	if err := json.Unmarshal([]byte(opts.Value), &payload); err != nil {
		return fmt.Errorf("error: invalid JSON payload\nhint: expected {\"fields\": {...}, \"sections\": {...}}")
	}

	// Apply fields
	for field, value := range payload.Fields {
		switch field {
		case "goal":
			entity.Goal = value
		case "status":
			entity.Status = value
		case "boundary":
			entity.Boundary = value
		case "category":
			entity.Category = value
		case "title":
			entity.Title = value
		case "date":
			entity.Date = value
		default:
			return fmt.Errorf("unknown field %q", field)
		}
	}

	// Apply sections through node tree
	if len(payload.Sections) > 0 {
		body, err := content.ReadEntity(opts.Store, opts.ID)
		if err != nil {
			body = ""
		}

		for section, sectionContent := range payload.Sections {
			newBody, err := markdown.ReplaceSection(body, section, sectionContent)
			if err != nil {
				return fmt.Errorf("error: section %q not found in %s\nhint: available sections: %s",
					section, opts.ID, availableSections(body))
			}
			body = newBody
		}

		if err := content.WriteEntity(opts.Store, opts.ID, body); err != nil {
			return fmt.Errorf("writing content: %w", err)
		}

		// Re-fetch to pick up rendered body/merkle/version.
		entity, err = opts.Store.GetEntity(opts.ID)
		if err != nil {
			return fmt.Errorf("re-fetch entity: %w", err)
		}

		// Re-apply fields (WriteEntity may have overwritten metadata).
		for field, value := range payload.Fields {
			switch field {
			case "goal":
				entity.Goal = value
			case "status":
				entity.Status = value
			case "boundary":
				entity.Boundary = value
			case "category":
				entity.Category = value
			case "title":
				entity.Title = value
			case "date":
				entity.Date = value
			}
		}
	}

	// Promote goal if updated via section
	if _, hasGoalSection := payload.Sections["Goal"]; hasGoalSection {
		promoteGoalIfEmpty(entity, opts.Store)
	}

	if err := opts.Store.UpdateEntity(entity); err != nil {
		return fmt.Errorf("updating entity: %w", err)
	}

	if payload.Codemap != nil {
		if err := opts.Store.SetCodeMap(entity.ID, *payload.Codemap); err != nil {
			return fmt.Errorf("updating codemap: %w", err)
		}
	}

	parts := fmt.Sprintf("%d fields, %d sections", len(payload.Fields), len(payload.Sections))
	if payload.Codemap != nil {
		parts += fmt.Sprintf(", %d codemap patterns", len(*payload.Codemap))
	}
	fmt.Fprintf(w, "Updated %s (%s)\n", opts.ID, parts)
	writeAgentHints(w, cascadeHintsForEntity(entity))
	return nil
}

func runSetField(entity *store.Entity, opts SetOptions, w io.Writer) error {
	// Codemap is a special field — stored in code_map table, not frontmatter.
	if opts.Field == "codemap" {
		return runSetCodemap(entity, opts, w)
	}

	// Map field name to entity field
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

// runSetSection updates a markdown section's content.
func runSetSection(entity *store.Entity, opts SetOptions, w io.Writer) error {
	// Read current content from node tree.
	body, err := content.ReadEntity(opts.Store, opts.ID)
	if err != nil {
		body = ""
	}

	var newBody string

	if opts.Append {
		// Append mode: parse single JSON object, append as row
		var row map[string]string
		if err := json.Unmarshal([]byte(opts.Value), &row); err != nil {
			return fmt.Errorf("malformed JSON for append: %w", err)
		}
		newBody, err = markdown.AppendTableRow(body, opts.Section, row)
		if err != nil {
			return err
		}
	} else if isJSONArray(opts.Value) {
		// JSON array: parse as table rows
		var rows []map[string]string
		if err := json.Unmarshal([]byte(opts.Value), &rows); err != nil {
			return fmt.Errorf("malformed JSON for table: %w", err)
		}
		// Extract existing table to get headers
		existingTable, err := markdown.ExtractTableFromSection(body, opts.Section)
		if err != nil {
			return err
		}
		var headers []string
		if existingTable != nil {
			headers = existingTable.Headers
		} else {
			// Derive headers from first row
			if len(rows) > 0 {
				for k := range rows[0] {
					headers = append(headers, k)
				}
			}
		}
		table := &markdown.Table{
			Headers: headers,
			Rows:    rows,
		}
		newBody, err = markdown.SetTableInSection(body, opts.Section, table)
		if err != nil {
			return err
		}
	} else {
		// Plain text: replace section content
		newBody, err = markdown.ReplaceSection(body, opts.Section, opts.Value)
		if err != nil {
			return err
		}
	}

	// Write through node tree.
	if err := content.WriteEntity(opts.Store, opts.ID, newBody); err != nil {
		return fmt.Errorf("writing content: %w", err)
	}

	if opts.Section == "Goal" {
		entity, err = opts.Store.GetEntity(opts.ID)
		if err != nil {
			return fmt.Errorf("re-fetch entity: %w", err)
		}
		promoteGoalIfEmpty(entity, opts.Store)
		if err := opts.Store.UpdateEntity(entity); err != nil {
			return fmt.Errorf("updating entity: %w", err)
		}
	}

	fmt.Fprintf(w, "Updated %s section %q\n", opts.ID, opts.Section)
	writeAgentHints(w, cascadeHintsForEntity(entity))
	return nil
}
