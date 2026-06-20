package cmd

import (
	"fmt"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// retireGate refuses a retire whose committed result would strand the graph — a child
// orphaned (its parent gone) or a citer left dangling (its body still names a retired
// fact). It checks the POST-APPLY state by replaying the whole unit in a preview
// transaction, so the unit's OWN reparents and re-cites count: retiring a fact AND
// re-pointing its children/citers away in the same unit is allowed; retiring it and
// leaving them stranded is refused. (A pre-apply check on the current graph would
// wrongly block the legitimate re-point-then-retire flow.) Integrity is the tool's:
// the destruction lands all-or-nothing only when nothing is left dangling.
func retireGate(s *store.Store, c3Dir string, patches []changeset.Patch) []string {
	retired := map[string]bool{}
	for _, p := range patches {
		if p.Scope == changeset.ScopeRetire {
			retired[p.Target] = true
		}
	}
	if len(retired) == 0 {
		return nil
	}

	// The children of, and citers to, each retired fact — re-examined after the unit
	// applies, where any reparent/re-cite in this same unit has landed.
	childOf := map[string]string{} // child id -> the retired parent it had
	suspects := map[string]bool{}
	for id := range retired {
		kids, _ := s.Children(id)
		for _, k := range kids {
			childOf[k.ID] = id
			suspects[k.ID] = true
		}
		citers, _ := s.RelationshipsTo(id)
		for _, r := range citers {
			suspects[r.FromID] = true
		}
	}

	rejectSet := map[string]bool{}
	_ = s.WithPreviewTx(func(ts *store.Store) error {
		if err := changeset.Apply(ts, patches, applyHooks(c3Dir)); err != nil {
			// The unit doesn't apply for some OTHER reason — a different gate reports it;
			// don't double-report here.
			return nil
		}
		for sid := range suspects {
			e, err := ts.GetEntity(sid)
			if err != nil {
				continue // the suspect was itself retired in this unit — fine
			}
			// Orphan: a child of a retired fact now has no live parent — retiring the
			// parent NULLs the child's parent_id (ON DELETE SET NULL) — and it was not
			// reparented to a live one in this unit.
			if oldParent, wasChild := childOf[sid]; wasChild {
				if e.ParentID == "" {
					rejectSet[fmt.Sprintf("retire would orphan child %s (parent %s retired) — reparent or retire %s in this unit", sid, oldParent, sid)] = true
				} else if _, perr := ts.GetEntity(e.ParentID); perr != nil {
					rejectSet[fmt.Sprintf("retire would orphan child %s (parent %s gone) — reparent or retire %s in this unit", sid, e.ParentID, sid)] = true
				}
			}
			// Dangle: body still cites a retired fact (this unit did not re-point it).
			body, err := content.ReadEntity(ts, sid)
			if err != nil {
				continue
			}
			for rid := range retired {
				if bodyReferencesID(body, rid) {
					rejectSet[fmt.Sprintf("retire of %s would dangle %s's citation — drop or re-point it in this unit", rid, sid)] = true
				}
			}
		}
		return nil
	})

	rejects := make([]string, 0, len(rejectSet))
	for r := range rejectSet {
		rejects = append(rejects, r)
	}
	sort.Strings(rejects)
	return rejects
}

// bodyReferencesID reports whether any table cell in the body names id as a reference
// token (id alone, or comma/space/pipe separated) — how a reference/edge column cites
// a fact. Used to detect a citation left dangling by a retire.
func bodyReferencesID(body, id string) bool {
	for _, sec := range markdown.ParseSections(body) {
		tbl, err := markdown.ParseTable(sec.Content)
		if err != nil || tbl == nil {
			continue
		}
		for _, row := range tbl.Rows {
			for _, cell := range row {
				for _, tok := range referenceTokens(cell) {
					if tok == id {
						return true
					}
				}
			}
		}
	}
	return false
}
