package changeset

import (
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// PatchState is a patch's lifecycle, derived purely from seal state (hash
// comparison) — never stored, never read from git.
type PatchState string

const (
	StatePending PatchState = "pending" // anchor fresh, not yet applied
	StateApplied PatchState = "applied" // the live block already seals to the patch's result
	StateDrifted PatchState = "drifted" // the anchor moved to something unexpected → rebase
	StateNew     PatchState = "new"     // a create patch whose target does not exist yet
)

// PatchStateOf derives a patch's state by comparing the cited anchor's hash with
// live seal state:
//   - base-hash == live  → pending (fresh, not applied)
//   - result   == live   → applied (it landed)
//   - otherwise          → drifted (anchor moved)
func PatchStateOf(s *store.Store, p Patch) PatchState {
	if p.Base == "" { // create
		if _, err := s.GetEntity(p.Target); err == nil {
			return StateApplied
		}
		return StateNew
	}
	_, nodeID, _, baseHash, ok := ParseCiteHandle(p.Base)
	if !ok {
		return StateDrifted
	}
	node, err := s.GetNode(nodeID)
	if err != nil || node.EntityID != p.Target {
		return StateDrifted
	}
	if node.Hash == baseHash {
		return StatePending
	}
	if p.Scope == ScopeBlock && node.Hash == store.ComputeNodeHash(p.Content, node.Type) {
		return StateApplied
	}
	return StateDrifted
}

// normalizeHash strips an optional "sha256:" prefix so a declared result-hash and
// a stored node hash compare directly.
func normalizeHash(h string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(h), "sha256:"))
}
