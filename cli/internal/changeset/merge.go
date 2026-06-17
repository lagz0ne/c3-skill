package changeset

import (
	"fmt"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// MergedBody returns the markdown body a patch would produce, WITHOUT writing —
// so the canvas gate can validate the merged result before any commit. For a
// block patch it clones the node tree, swaps the cited node's content, and
// renders; every other node is untouched.
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
		merged := make([]*store.Node, len(nodes))
		for i, n := range nodes {
			cp := *n
			if cp.ID == nodeID {
				cp.Content = p.Content
				cp.Hash = store.ComputeNodeHash(p.Content, cp.Type)
			}
			merged[i] = &cp
		}
		return content.RenderMarkdown(merged), nil
	default:
		return "", fmt.Errorf("MergedBody: scope %q not yet supported", p.Scope)
	}
}
