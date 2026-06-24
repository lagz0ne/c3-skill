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
		_, hintID, _, expected, ok := ParseCiteHandle(p.Base)
		if !ok {
			return "", fmt.Errorf("patch %s: malformed base handle", p.Source)
		}
		nodes, err := s.NodesForEntity(p.Target)
		if err != nil {
			return "", err
		}
		// Anchor by hash (id is a hint) — same as CheckDrift/applyBlock, so the gate
		// preview matches what apply will actually mutate even after renumbering.
		target, err := resolveCitedNode(s, p.Target, hintID, expected)
		if err != nil {
			return "", fmt.Errorf("patch %s: %v", p.Source, err)
		}
		nodeID := target.ID
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
	case ScopeInsert:
		// Block-base insert (e.g. a new table row after a cited row): render the body
		// with the new node spliced in after the anchor, mirroring applyInsert.
		_, hintID, _, expected, ok := ParseCiteHandle(p.Base)
		if !ok {
			return "", fmt.Errorf("patch %s: MergedBody insert here needs a block base", p.Source)
		}
		nodes, err := s.NodesForEntity(p.Target)
		if err != nil {
			return "", err
		}
		after, err := resolveCitedNode(s, p.Target, hintID, expected)
		if err != nil {
			return "", fmt.Errorf("patch %s: %v", p.Source, err)
		}
		body := p.Content
		if after.Type == "table_row" || after.Type == "table_header" {
			body = normalizeTableRowContent(body)
		}
		nodeType := after.Type
		if after.Type == "table_header" {
			nodeType = "table_row"
		}
		newNode := &store.Node{EntityID: p.Target, ParentID: after.ParentID, Type: nodeType, Level: after.Level, Content: body}
		var merged []*store.Node
		for _, n := range nodes {
			cp := *n
			merged = append(merged, &cp)
			if n.ID == after.ID {
				merged = append(merged, newNode)
			}
		}
		return content.RenderMarkdown(merged), nil
	default:
		return "", fmt.Errorf("MergedBody: scope %q not yet supported", p.Scope)
	}
}
