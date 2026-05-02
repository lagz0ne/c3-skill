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

	writeRejectIfBlock(w, entityType)

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
	}

	return nil
}

// writeRejectIfBlock prints the REJECT IF rejection contract for adr, ref, and rule.
func writeRejectIfBlock(w io.Writer, entityType string) {
	switch entityType {
	case "adr":
		fmt.Fprintln(w)
		fmt.Fprintln(w, "REJECT IF:")
		fmt.Fprintln(w, "  - Any required section absent or filled with TBD/TODO/\"see above\"/\"as needed\"")
		fmt.Fprintln(w, "  - Compliance rows must say why the ref/rule applies, unless the whole row is N.A - <reason>")
		fmt.Fprintln(w, "  - Affected Topology rows must say why the entity is affected, unless the whole row is N.A - <reason>")
		fmt.Fprintln(w, "  - Verification has no executable command, smoke check, or named artifact")
		fmt.Fprintln(w, "  - Alternatives Considered rows have no repo-specific rejection reason")
		fmt.Fprintln(w, "  - Underlay C3 Changes lacks the exact validators/tests/help that enforce the decision")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Run c3x schema adr before drafting; do not draft ADR prose first and reconcile later.")
		fmt.Fprintln(w, "Treat each 'fill' line as required authoring guidance, not optional commentary.")
		fmt.Fprintln(w, "ADR creation is all-or-nothing: thin sections fail at creation, no incremental fill later.")
		fmt.Fprintln(w)
	case "ref":
		fmt.Fprintln(w)
		fmt.Fprintln(w, "REJECT IF:")
		fmt.Fprintln(w, "  - 'Why' restates 'Choice' instead of giving rationale (the ref becomes a rule)")
		fmt.Fprintln(w, "  - 'Goal' describes what code does instead of what problem the pattern standardizes")
		fmt.Fprintln(w, "  - 'Choice' is generic ('use best practices') instead of naming a concrete option")
		fmt.Fprintln(w, "  - No file path or grep evidence backs the 'How' pattern (one-off, not a ref)")
		fmt.Fprintln(w, "  - Pattern is primarily about enforcement (golden code, anti-patterns) — that's a rule, not a ref")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Refs are rationale documents. If you cannot answer 'why this pattern over alternatives'")
		fmt.Fprintln(w, "you do not have a ref yet — discover first, then draft.")
		fmt.Fprintln(w)
	case "rule":
		fmt.Fprintln(w)
		fmt.Fprintln(w, "REJECT IF:")
		fmt.Fprintln(w, "  - 'Golden Example' is paraphrased instead of literal code copied from a real file")
		fmt.Fprintln(w, "  - 'Rule' is multi-clause or aspirational ('should generally') instead of one-line enforceable")
		fmt.Fprintln(w, "  - No 1-3 YES/NO compliance question can be derived from 'Rule' + 'Golden Example'")
		fmt.Fprintln(w, "  - Rule is primarily about rationale (why this approach) — that's a ref, not a rule")
		fmt.Fprintln(w, "  - 'Goal' describes a single component instead of a project-wide standard")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Rules are enforceable standards. Find the canonical code in the codebase FIRST.")
		fmt.Fprintln(w, "If no real example exists, the rule is premature — author the first instance, then extract the rule.")
		fmt.Fprintln(w)
	}
}
