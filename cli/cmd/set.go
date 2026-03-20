package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
}

// RunSet updates a frontmatter field or section content on an entity.
func RunSet(opts SetOptions, w io.Writer) error {
	entity, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("entity %q not found", opts.ID)
	}

	if opts.Section != "" {
		return runSetSection(entity, opts, w)
	}
	return runSetField(entity, opts, w)
}

// runSetField updates a frontmatter field via the store.
// See also: applyFrontmatter in write.go maps the same fields from a Frontmatter struct.
func runSetField(entity *store.Entity, opts SetOptions, w io.Writer) error {
	// Map field name to entity field
	switch opts.Field {
	case "goal":
		entity.Goal = opts.Value
	case "summary":
		entity.Summary = opts.Value
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
	case "description":
		entity.Description = opts.Value
	default:
		return fmt.Errorf("unknown field %q", opts.Field)
	}

	if err := opts.Store.UpdateEntity(entity); err != nil {
		return fmt.Errorf("updating entity: %w", err)
	}

	fmt.Fprintf(w, "Updated %s field %q\n", opts.ID, opts.Field)
	return nil
}

// runSetSection updates a markdown section's content.
func runSetSection(entity *store.Entity, opts SetOptions, w io.Writer) error {
	body := entity.Body
	var newBody string
	var err error

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

	entity.Body = newBody

	if opts.Section == "Goal" {
		promoteGoalIfEmpty(entity)
	}

	issues := validateContent(entity)
	if len(issues) > 0 {
		return formatValidationError(opts.ID, issues)
	}

	if err := opts.Store.UpdateEntity(entity); err != nil {
		return fmt.Errorf("updating entity body: %w", err)
	}

	fmt.Fprintf(w, "Updated %s section %q\n", opts.ID, opts.Section)
	return nil
}
