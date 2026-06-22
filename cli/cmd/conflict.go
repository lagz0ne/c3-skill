package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// patchBaseContent recovers the content a block patch anchored to — the cited node
// as it was at the patch's base version — by re-hashing the nodes of that version's
// stored body and matching the cite handle's sha256. Empty when unrecoverable (the
// base is an entity handle, the version is gone, or the hash isn't in it). This is
// the BASE arm of the 3-way a conflicted patch needs.
func patchBaseContent(s *store.Store, p changeset.Patch) string {
	entity, _, version, hash, ok := changeset.ParseCiteHandle(p.Base)
	if !ok {
		return ""
	}
	v, err := s.GetVersion(entity, version)
	if err != nil {
		return ""
	}
	tree := content.ParseMarkdown(entity, v.Content)
	for _, n := range tree.Nodes {
		if store.ComputeNodeHash(n.Content, n.Type) == hash {
			return n.Content
		}
	}
	return ""
}

// renderConflict writes the 3-way a drifted patch needs to be re-authored: BASE (the
// block it anchored to, recovered from history), YOURS (the change it carries), and
// the move to re-anchor against CURRENT (the live frozen fact). The author rewrites
// the patch against current; apply then re-runs every gate, so a stale resolution
// still can't land. This is the conflict resolution flow — the saga refuses to apply
// while any patch is in conflict, and this surface is how you clear it.
func renderConflict(w io.Writer, s *store.Store, p changeset.Patch, reason string) {
	fmt.Fprintf(w, "conflict %s → %s (%s)\n  reason: %s\n", p.Source, p.Target, p.Scope, reason)
	if base := patchBaseContent(s, p); base != "" {
		fmt.Fprintf(w, "  you anchored to:\n%s\n", indentBlock(base, "    "))
	}
	if y := strings.TrimSpace(p.Content); y != "" {
		fmt.Fprintf(w, "  your change:\n%s\n", indentBlock(y, "    "))
	}
	fmt.Fprintf(w, "  current moved under you — re-anchor: c3 read %s --cite, then update this patch's base + body and re-apply\n", p.Target)
}

func indentBlock(s, pad string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}
