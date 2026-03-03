package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// SchemaOutput is the JSON structure returned by the schema command.
type SchemaOutput struct {
	Type     string             `json:"type"`
	Sections []schema.SectionDef `json:"sections"`
}

// RunSchema outputs the section schema for a given entity type.
func RunSchema(entityType string, jsonOutput bool, w io.Writer) error {
	sections := schema.ForType(entityType)
	if sections == nil {
		return fmt.Errorf("unknown entity type: %q", entityType)
	}

	if jsonOutput {
		out := SchemaOutput{
			Type:     entityType,
			Sections: sections,
		}
		return writeJSON(w, out)
	}

	// Text output
	fmt.Fprintf(w, "Schema: %s\n", entityType)
	for _, s := range sections {
		req := ""
		if s.Required {
			req = " (required)"
		}
		fmt.Fprintf(w, "  %s [%s]%s\n", s.Name, s.ContentType, req)
		for _, col := range s.Columns {
			fmt.Fprintf(w, "    - %s (%s)\n", col.Name, col.Type)
		}
	}
	return nil
}
