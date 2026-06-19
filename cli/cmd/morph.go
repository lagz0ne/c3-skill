package cmd

import (
	"fmt"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// morphGate refuses a canvas morph whose committed result would leave any existing
// instance of that fact-type invalid against the NEW shape — UNLESS the same unit also
// migrates every such instance. It is the evolve-unit's one new gate, and it mirrors the
// retire gate: the model moves only if every fact can come with it.
//
// A `canvas`-scope patch carries the morphed definition (its `Content`). The gate parses
// that new shape, replays the unit's OWN fact-migrations in a preview transaction, then
// validates every instance of the morphed type against the new shape — so a morph that
// raises the contract AND migrates its instances in the same unit is allowed, while a
// morph that strands them is refused. (A pre-apply check on the current bodies would
// wrongly block the legitimate morph-then-migrate flow.) Integrity is the tool's: the
// schema lands atomically only when nothing it governs is left invalid.
func morphGate(s *store.Store, c3Dir string, patches []changeset.Patch, codemaps []changeset.CodemapChange) []string {
	morphed := map[string][]schema.SectionDef{} // type -> its new sections
	factPatches := make([]changeset.Patch, 0, len(patches))
	rejectSet := map[string]bool{}
	for _, p := range patches {
		if p.Scope == changeset.ScopeCanvas {
			canvas, err := schema.ParseCanvasDocument("canvases/"+p.Target+".md", p.Content)
			if err != nil {
				// The morphed shape is not even a valid canvas — refuse it outright.
				rejectSet[fmt.Sprintf("morph of canvas %s is itself invalid: %v", p.Target, err)] = true
				continue
			}
			morphed[p.Target] = canvas.Sections
			continue
		}
		factPatches = append(factPatches, p)
	}

	if len(morphed) > 0 {
		_ = s.WithPreviewTx(func(ts *store.Store) error {
			// Apply the unit's fact-migrations first, so an instance brought up to the
			// new shape IN THIS UNIT counts as migrated.
			if err := changeset.Apply(ts, factPatches, codemaps, applyHooks(c3Dir)); err != nil {
				// The unit doesn't apply for some OTHER reason — a different gate reports
				// it; don't double-report here.
				return nil
			}
			for typ, sections := range morphed {
				ents, err := ts.EntitiesByType(typ)
				if err != nil {
					continue
				}
				for _, e := range ents {
					body, err := content.ReadEntity(ts, e.ID)
					if err != nil {
						continue
					}
					if issues := validateBodyContentWithDefinition(body, typ, sections); len(issues) > 0 {
						rejectSet[fmt.Sprintf("morph of canvas %s would leave %s invalid (%s) — migrate it in this unit", typ, e.ID, issues[0].Message)] = true
					}
				}
			}
			return nil
		})
	}

	rejects := make([]string, 0, len(rejectSet))
	for r := range rejectSet {
		rejects = append(rejects, r)
	}
	sort.Strings(rejects)
	return rejects
}
