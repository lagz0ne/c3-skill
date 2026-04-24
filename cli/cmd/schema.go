package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// SchemaOutput is the JSON structure returned by the schema command.
type SchemaOutput struct {
	Type     string              `json:"type"`
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
		constraints := ""
		if s.MinWords > 0 {
			constraints += fmt.Sprintf(" (min %d words)", s.MinWords)
		}
		if s.MinRows > 0 {
			constraints += fmt.Sprintf(" (min %d rows)", s.MinRows)
		}
		fmt.Fprintf(w, "  %s [%s]%s%s\n", s.Name, s.ContentType, req, constraints)
		if s.Purpose != "" {
			fmt.Fprintf(w, "    purpose: %s\n", s.Purpose)
		}
		if s.Fill != "" {
			fmt.Fprintf(w, "    fill: %s\n", s.Fill)
		}
		if s.Failure != "" {
			fmt.Fprintf(w, "    if weak/missing: %s\n", s.Failure)
		}
		for _, col := range s.Columns {
			if len(col.Values) > 0 {
				fmt.Fprintf(w, "    - %s (%s) values: %s\n", col.Name, col.Type, strings.Join(col.Values, ", "))
			} else {
				fmt.Fprintf(w, "    - %s (%s)\n", col.Name, col.Type)
			}
		}
	}

	// Type-specific rules
	switch entityType {
	case "component":
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Component rules:")
		fmt.Fprintln(w, "  - Sections must appear in the order shown above")
		fmt.Fprintln(w, "  - No placeholder words: TBD, TODO, maybe, optional, later, \"if applicable\"")
		fmt.Fprintln(w, "  - Empty cells: use N.A - <reason> (not N/A, n/a, or bare N.A)")
		fmt.Fprintln(w, "  - Evidence columns: must be grounded — name a command, file path, or entity id")
		fmt.Fprintln(w, "  - Reference columns: must cite an entity id (c3-*, ref-*, rule-*) or N.A - <reason>")
	case "adr":
		fmt.Fprintln(w)
		fmt.Fprintln(w, "ADR rules:")
		fmt.Fprintln(w, "  - All sections required at creation (all-or-nothing)")
		fmt.Fprintln(w, "  - Run c3x schema adr before drafting; do not draft ADR prose first and reconcile later")
		fmt.Fprintln(w, "  - Treat each 'fill' line as required authoring guidance, not optional commentary")
		fmt.Fprintln(w, "  - Treat each 'if weak/missing' line as the failure the section is preventing")
		fmt.Fprintln(w, "  - Compliance rows must say why the ref/rule applies, unless the whole row is N.A - <reason>")
		fmt.Fprintln(w, "  - Affected Topology rows must say why the entity is affected, unless the whole row is N.A - <reason>")
	}

	return nil
}
