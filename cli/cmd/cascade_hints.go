package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func agentHints(hints []HelpHint) []HelpHint {
	if !isAgentMode() || len(hints) == 0 {
		return nil
	}
	return hints
}

func writeAgentHints(w io.Writer, hints []HelpHint) {
	writeHints(w, hints)
}

func cascadeReviewHints() []HelpHint {
	return []HelpHint{
		{Command: "cascade review", Description: "for each changed component, record ADR Parent Delta: updated or no-delta with evidence"},
		{Command: "git diff --name-only -- . ':(exclude).c3/c3.db'", Description: "find changed files before declaring docs synced"},
		{Command: "c3x check --only <id>", Description: "prove focused docs while unrelated branch docs or ADRs are still in progress"},
		{Command: "c3x check", Description: "structural pass; still prove Parent Delta evidence"},
	}
}

func cascadeHintsForEntity(entity *store.Entity) []HelpHint {
	if entity == nil {
		return nil
	}

	switch entity.Type {
	case "adr":
		return adrHints(entity.ID)
	case "component":
		var hints []HelpHint
		if entity.ParentID != "" {
			hints = append(hints,
				HelpHint{Command: fmt.Sprintf("c3x read %s", entity.ParentID), Description: "TOP-DOWN first: read parent container responsibility before component detail"},
				HelpHint{Command: fmt.Sprintf("c3x graph %s --format mermaid", entity.ParentID), Description: "verify container/component relationship"},
			)
		}
		hints = append(hints, HelpHint{Command: fmt.Sprintf("c3x read %s", entity.ID), Description: "read matched/changed component contract after parent context"})
		hints = append(hints, HelpHint{Command: "ADR Parent Delta", Description: "record parent update or no-delta evidence before done"})
		return hints
	case "container":
		return []HelpHint{
			{Command: fmt.Sprintf("c3x graph %s --format mermaid", entity.ID), Description: "verify component membership and cited refs/rules"},
			{Command: "ADR Parent Delta", Description: "compare changed components against Components and Responsibilities sections"},
		}
	case "ref", "rule":
		return []HelpHint{
			{Command: fmt.Sprintf("c3x graph %s --format mermaid", entity.ID), Description: "show all citing components before changing shared constraint"},
			{Command: "cascade review", Description: "check every citing component for compliance drift"},
		}
	default:
		if entity.ParentID != "" {
			return []HelpHint{
				{Command: fmt.Sprintf("c3x read %s", entity.ParentID), Description: "read parent before changing child entity"},
				{Command: "ADR Parent Delta", Description: "record parent update or no-delta evidence before done"},
			}
		}
		return cascadeReviewHints()
	}
}

func adrHints(entityID string) []HelpHint {
	schemaCommand := "c3x schema adr"
	return []HelpHint{
		{Command: schemaCommand, Description: "authoritative ADR canvas contract for required sections, tables, and rejection rules"},
		{Command: fmt.Sprintf("c3x read %s --full", entityID), Description: "inspect the complete ADR work order, including why each compliance row is required"},
		{Command: fmt.Sprintf("c3x write %s < adr.md", entityID), Description: "replace the full ADR only if the complete work order must change"},
		{Command: fmt.Sprintf("c3x check --include-adr --only %s", entityID), Description: "prove this ADR while other branch docs are still in progress"},
		{Command: "c3x check --include-adr", Description: "prove ADR compliance rows, structural coverage, and canonical sync before final handoff"},
	}
}

func newComponentTopDownHints(entity *store.Entity) []HelpHint {
	if entity == nil || entity.Type != "component" {
		return cascadeHintsForEntity(entity)
	}
	hints := []HelpHint{}
	if entity.ParentID != "" {
		hints = append(hints,
			HelpHint{Command: fmt.Sprintf("c3x read %s", entity.ParentID), Description: "TOP-DOWN required for new component: update container Components/Responsibilities first"},
			HelpHint{Command: fmt.Sprintf("c3x graph %s --format mermaid", entity.ParentID), Description: "prove parent container now sees the new component"},
		)
	}
	hints = append(hints,
		HelpHint{Command: fmt.Sprintf("c3x read %s", entity.ID), Description: "then refine strict component Goal, Parent Fit, Governance, Contract, and Change Safety"},
		HelpHint{Command: "ADR Parent Delta", Description: "new component cannot finish without parent update evidence"},
	)
	return hints
}

func cascadeHintsForID(s *store.Store, id string) []HelpHint {
	entity, err := s.GetEntity(id)
	if err != nil {
		return nil
	}
	return cascadeHintsForEntity(entity)
}

func lookupMissHints(filePath string) []HelpHint {
	return []HelpHint{
		{Command: "c3x eval", Description: "coverage gap: no eval spec's code globs map this path — add the binding in .c3/eval/<fact>.yaml"},
		{Command: fmt.Sprintf("c3x lookup %q", filePath), Description: "rerun after updating the eval spec's code globs"},
		{Command: "ADR Parent Delta", Description: "uncharted files need explicit ownership evidence before done"},
	}
}
