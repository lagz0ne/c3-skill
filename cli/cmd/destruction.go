package cmd

import (
	"fmt"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// retireGate refuses a retire whose consequences would leave the frozen graph
// dangling — integrity is the tool's, so the saga cannot commit a destruction it
// hasn't fully resolved. A fact may not be retired while it still has:
//   - live children (they would be orphaned), unless this same unit also retires
//     them or reparents them away (a frontmatter parent: change); or
//   - live citers (their citations would dangle), unless this same unit also
//     retires the citer.
//
// The membership row drop is automatic (the reconcile on the retired child's old
// parent), so it is never the author's concern — only these author-owned references
// are gated. Resolve them in the same atomic unit, then the retire lands.
func retireGate(s *store.Store, patches []changeset.Patch) []string {
	retired := map[string]bool{}
	reparented := map[string]string{} // child -> new parent, from frontmatter patches in this unit
	for _, p := range patches {
		switch p.Scope {
		case changeset.ScopeRetire:
			retired[p.Target] = true
		case changeset.ScopeFrontmatter:
			if p.Parent != "" {
				reparented[p.Target] = p.Parent
			}
		}
	}
	if len(retired) == 0 {
		return nil
	}

	var rejects []string
	for id := range retired {
		// Children this unit neither retires nor reparents away ⇒ orphans.
		children, _ := s.Children(id)
		for _, c := range children {
			if retired[c.ID] {
				continue
			}
			if np, ok := reparented[c.ID]; ok && np != id {
				continue // moved to a live parent in this same unit
			}
			rejects = append(rejects, fmt.Sprintf("retire %s would orphan child %s — retire or reparent %s in this unit", id, c.ID, c.ID))
		}
		// Citers this unit does not also retire ⇒ dangling citations.
		citers, _ := s.RelationshipsTo(id)
		seen := map[string]bool{}
		for _, r := range citers {
			if retired[r.FromID] || seen[r.FromID] {
				continue
			}
			seen[r.FromID] = true
			rejects = append(rejects, fmt.Sprintf("retire %s: %s still cites it (%s) — drop that citation in this unit before retiring", id, r.FromID, r.RelType))
		}
	}
	sort.Strings(rejects)
	return rejects
}
