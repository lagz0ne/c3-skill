package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/writer"
)

// SetOptions holds parameters for the set command.
type SetOptions struct {
	C3Dir   string
	ID      string
	Field   string
	Section string
	Value   string
	Append  bool
}

// RunSet updates a frontmatter field or section content on an entity.
func RunSet(opts SetOptions, w io.Writer) error {
	// Find the file for this entity
	path, err := findEntityFile(opts.C3Dir, opts.ID)
	if err != nil {
		return err
	}

	if opts.Section != "" {
		return runSetSection(path, opts, w)
	}
	return runSetField(path, opts, w)
}

// runSetField updates a frontmatter field.
func runSetField(path string, opts SetOptions, w io.Writer) error {
	if err := writer.SetField(path, opts.Field, opts.Value); err != nil {
		return err
	}
	fmt.Fprintf(w, "Updated %s field %q on %s\n", opts.ID, opts.Field, filepath.Base(path))
	return nil
}

// runSetSection updates a markdown section's content.
func runSetSection(path string, opts SetOptions, w io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fm, body := frontmatter.ParseFrontmatter(string(data))
	if fm == nil {
		return fmt.Errorf("no valid frontmatter in %s", path)
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
	} else if strings.HasPrefix(strings.TrimSpace(opts.Value), "[") {
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

	// Write back: re-serialize frontmatter + new body
	return writeEntityFile(path, fm, newBody)
}

