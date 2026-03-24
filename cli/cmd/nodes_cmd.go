package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// NodesOptions holds parameters for the nodes command.
type NodesOptions struct {
	Store    *store.Store
	EntityID string
	JSON     bool
}

// NodeRow is the JSON representation of a node.
type NodeRow struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id,omitempty"`
	Type     string `json:"type"`
	Level    int    `json:"level,omitempty"`
	Seq      int    `json:"seq"`
	Hash     string `json:"hash"`
	Content  string `json:"content"`
}

// RunNodes lists all nodes for an entity as a tree.
func RunNodes(opts NodesOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("usage: c3x nodes <entity-id>")
	}

	if _, err := opts.Store.GetEntity(opts.EntityID); err != nil {
		return fmt.Errorf("entity %q not found", opts.EntityID)
	}

	nodes, err := opts.Store.NodesForEntity(opts.EntityID)
	if err != nil {
		return fmt.Errorf("fetching nodes: %w", err)
	}

	if len(nodes) == 0 {
		fmt.Fprintln(w, "No nodes for "+opts.EntityID)
		return nil
	}

	if opts.JSON {
		rows := make([]NodeRow, len(nodes))
		for i, n := range nodes {
			rows[i] = NodeRow{
				ID:      n.ID,
				Type:    nodeTypeLabel(n),
				Seq:     n.Seq,
				Hash:    n.Hash,
				Content: n.Content,
			}
			if n.ParentID.Valid {
				rows[i].ParentID = n.ParentID.Int64
			}
			if n.Level > 0 {
				rows[i].Level = n.Level
			}
		}
		return writeJSON(w, rows)
	}

	depthMap := buildDepthMap(nodes)
	fmt.Fprintf(w, "%-6s %-16s %-4s %-10s %s\n", "ID", "TYPE", "SEQ", "HASH", "CONTENT")
	for _, n := range nodes {
		depth := depthMap[n.ID]
		indent := strings.Repeat("  ", depth)
		label := nodeTypeLabel(n)
		hash := n.Hash
		if len(hash) > 8 {
			hash = hash[:8]
		}
		content := n.Content
		if len(content) > 40 {
			content = content[:40] + "..."
		}
		fmt.Fprintf(w, "%-6d %s%-*s %-4d %-10s %s\n",
			n.ID, indent, 16-len(indent), label, n.Seq, hash, content)
	}
	return nil
}

// nodeTypeLabel returns the type with level suffix for headings.
func nodeTypeLabel(n *store.Node) string {
	if n.Type == "heading" && n.Level > 0 {
		return fmt.Sprintf("heading[%d]", n.Level)
	}
	return n.Type
}

// buildDepthMap computes the tree depth for each node by ID.
func buildDepthMap(nodes []*store.Node) map[int64]int {
	dm := make(map[int64]int)
	for _, n := range nodes {
		if !n.ParentID.Valid {
			dm[n.ID] = 0
		} else {
			dm[n.ID] = dm[n.ParentID.Int64] + 1
		}
	}
	return dm
}
