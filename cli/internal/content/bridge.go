package content

import (
	"fmt"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// WriteEntity parses markdown, then atomically replaces the entity's nodes,
// snapshots a version, and reseals the entity merkle — all in one transaction.
// When called inside a change apply it enlists in that open transaction, so a body
// write commits (or rolls back) as one unit and never leaves a half-updated fact.
func WriteEntity(s *store.Store, entityID, markdown string) error {
	tree := ParseMarkdown(entityID, stripFrontmatter(markdown))
	merkle := collectMerkle(tree.Nodes)
	rendered := RenderMarkdown(tree.Nodes)

	return s.WithTx(func(ts *store.Store) error {
		if err := ts.InsertNodeTree(entityID, tree.Nodes, tree.ParentIndex); err != nil {
			return fmt.Errorf("write nodes: %w", err)
		}
		ver, err := ts.CreateVersion(entityID, rendered, merkle)
		if err != nil {
			return fmt.Errorf("create version: %w", err)
		}
		entity, err := ts.GetEntity(entityID)
		if err != nil {
			return fmt.Errorf("get entity: %w", err)
		}
		entity.RootMerkle = merkle
		entity.Version = ver.Version
		syncGoalFromNodes(entity, tree)
		if err := ts.UpdateEntity(entity); err != nil {
			return fmt.Errorf("update entity: %w", err)
		}
		return nil
	})
}

// ReadEntity reads nodes from the DB and renders them to markdown.
func ReadEntity(s *store.Store, entityID string) (string, error) {
	nodes, err := s.NodesForEntity(entityID)
	if err != nil {
		return "", fmt.Errorf("read nodes: %w", err)
	}
	return RenderMarkdown(nodes), nil
}

func collectMerkle(nodes []*store.Node) string {
	hashes := make([]string, len(nodes))
	for i, n := range nodes {
		hashes[i] = n.Hash
	}
	return store.ComputeRootMerkle(hashes)
}

// stripFrontmatter removes YAML frontmatter (--- delimited) and any
// stale key: value lines before the first markdown heading.
func stripFrontmatter(md string) string {
	// Strip --- delimited frontmatter block.
	if strings.HasPrefix(md, "---\n") {
		if idx := strings.Index(md[4:], "\n---"); idx >= 0 {
			md = strings.TrimLeft(md[idx+8:], "\n")
		}
	}
	// Strip stale key: value lines before the first heading.
	lines := strings.Split(md, "\n")
	start := 0
	for start < len(lines) {
		l := strings.TrimSpace(lines[start])
		if l == "" || strings.HasPrefix(l, "#") {
			break
		}
		// Keep lines that don't look like YAML key: value.
		if !isYAMLLine(l) {
			break
		}
		start++
	}
	if start > 0 {
		md = strings.Join(lines[start:], "\n")
	}
	return strings.TrimLeft(md, "\n")
}

func isYAMLLine(line string) bool {
	// Matches: "key: value", "key:", "- item" (YAML array items).
	if strings.HasPrefix(line, "- ") {
		return true
	}
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return false
	}
	key := line[:idx]
	return !strings.Contains(key, " ") || strings.HasPrefix(key, "##")
}

func syncGoalFromNodes(entity *store.Entity, tree *NodeTree) {
	for i, node := range tree.Nodes {
		if node.Type == "heading" && node.Content == "Goal" {
			for j := i + 1; j < len(tree.Nodes); j++ {
				if tree.ParentIndex[j] == i && tree.Nodes[j].Type == "paragraph" {
					entity.Goal = tree.Nodes[j].Content
					return
				}
			}
			return
		}
	}
}
