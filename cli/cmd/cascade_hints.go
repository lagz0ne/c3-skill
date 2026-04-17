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
		{Command: "c3x diff", Description: "find component-only deltas before declaring docs synced"},
		{Command: "c3x check", Description: "structural pass; still verify Parent Delta evidence"},
	}
}

func cascadeHintsForEntity(entity *store.Entity) []HelpHint {
	if entity == nil {
		return nil
	}

	switch entity.Type {
	case "adr":
		return []HelpHint{
			{Command: "c3x schema adr", Description: "authoritative ADR creation contract from the CLI"},
			{Command: fmt.Sprintf("c3x read %s --full", entity.ID), Description: "inspect the complete ADR work order before execution"},
			{Command: fmt.Sprintf("c3x write %s < adr.md", entity.ID), Description: "replace the full ADR only if the complete work order must change"},
			{Command: "c3x check --include-adr && c3x verify", Description: "validate ADR detail and canonical sync before done"},
		}
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
		{Command: "c3x codemap", Description: "coverage gap: map or explicitly exclude the surfaced path"},
		{Command: fmt.Sprintf("c3x lookup %q", filePath), Description: "rerun after codemap update"},
		{Command: "ADR Parent Delta", Description: "uncharted files need explicit ownership evidence before done"},
	}
}
