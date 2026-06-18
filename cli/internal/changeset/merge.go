package changeset

import (
	"fmt"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// MergedBody returns the markdown body a patch would produce, WITHOUT writing —
// so the canvas gate can validate the merged result before any commit. It mirrors
// applyBlock exactly: an empty body DELETES the cited node (and its descendants);
// a table-row edit is normalized to the stored cell form; otherwise the node's
// content is swapped. Every other node is untouched.
func MergedBody(s *store.Store, p Patch) (string, error) {
	switch p.Scope {
	case ScopeBlock:
		_, nodeID, _, _, ok := ParseCiteHandle(p.Base)
		if !ok {
			return "", fmt.Errorf("patch %s: malformed base handle", p.Source)
		}
		nodes, err := s.NodesForEntity(p.Target)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(p.Content) == "" {
			// Deletion: drop the cited node and every descendant (DeleteNode cascades).
			drop := map[int64]bool{nodeID: true}
			for changed := true; changed; {
				changed = false
				for _, n := range nodes {
					if n.ParentID.Valid && drop[n.ParentID.Int64] && !drop[n.ID] {
						drop[n.ID] = true
						changed = true
					}
				}
			}
			var merged []*store.Node
			for _, n := range nodes {
				if drop[n.ID] {
					continue
				}
				cp := *n
				merged = append(merged, &cp)
			}
			return content.RenderMarkdown(merged), nil
		}
		merged := make([]*store.Node, len(nodes))
		for i, n := range nodes {
			cp := *n
			if cp.ID == nodeID {
				c := p.Content
				if cp.Type == "table_row" || cp.Type == "table_header" {
					c = normalizeTableRowContent(c)
				}
				cp.Content = c
				cp.Hash = store.ComputeNodeHash(c, cp.Type)
			}
			merged[i] = &cp
		}
		return content.RenderMarkdown(merged), nil
	default:
		return "", fmt.Errorf("MergedBody: scope %q not yet supported", p.Scope)
	}
}
