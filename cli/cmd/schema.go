package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// SchemaOutput is the JSON structure returned by the schema command.
type SchemaOutput struct {
	Type      string              `json:"type"`
	Sections  []schema.SectionDef `json:"sections"`
	RejectIf  []string            `json:"reject_if,omitempty"`
	Workorder string              `json:"workorder,omitempty"`
}

type SchemaOptions struct {
	EntityType string
	JSON       bool
	C3Dir      string
	Template   string
}

// RunSchema outputs the section schema for a given entity type.
func RunSchema(entityType string, jsonOutput bool, w io.Writer) error {
	return RunSchemaWithOptions(SchemaOptions{EntityType: entityType, JSON: jsonOutput}, w)
}

func RunSchemaWithOptions(opts SchemaOptions, w io.Writer) error {
	entityType := opts.EntityType
	if opts.Template != "" {
		return fmt.Errorf("error: --template has been retired\nhint: use c3x canvas read adr or c3x schema adr")
	}
	def, ok := schema.DefinitionForDir(opts.C3Dir, entityType)
	if !ok {
		return fmt.Errorf("unknown entity type: %q", entityType)
	}
	sections := def.Sections

	if opts.JSON {
		out := SchemaOutput{
			Type:      entityType,
			Sections:  sections,
			RejectIf:  def.Reject.Bullets,
			Workorder: def.Reject.Workorder,
		}
		return writeJSON(w, out)
	}

	// Text output
	fmt.Fprintf(w, "Schema: %s\n", entityType)

	writeRejectIfBlock(w, def.Reject)

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
			fmt.Fprintf(w, "    rejected when: %s\n", s.Failure)
		}
		for _, col := range s.Columns {
			edge := ""
			if col.Edge != "" {
				edge = fmt.Sprintf("  → edge: %s", col.Edge)
				if len(col.Targets) > 0 {
					edge += fmt.Sprintf(" (targets: %s)", strings.Join(col.Targets, ", "))
				}
			}
			if len(col.Values) > 0 {
				fmt.Fprintf(w, "    - %s (%s) values: %s%s\n", col.Name, col.Type, strings.Join(col.Values, ", "), edge)
			} else {
				fmt.Fprintf(w, "    - %s (%s)%s\n", col.Name, col.Type, edge)
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
		fmt.Fprintln(w, "  - A column marked `→ edge: <rel>` IS the citation: authoring its cell materializes that graph edge (no separate c3 wire)")
	}

	return nil
}

// writeRejectIfBlock prints the REJECT IF rejection contract from the resolved
// canvas definition. Empty contracts print nothing.
func writeRejectIfBlock(w io.Writer, rules schema.RejectRules) {
	if len(rules.Bullets) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "REJECT IF:")
	for _, bullet := range rules.Bullets {
		fmt.Fprintf(w, "  - %s\n", bullet)
	}
	if rules.Workorder != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, rules.Workorder)
	}
	fmt.Fprintln(w)
}
