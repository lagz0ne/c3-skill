package store

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// ComputeNodeHash returns SHA256 hex digest of the node content.
// For container nodes (empty content), hashes the type name.
func ComputeNodeHash(content, nodeType string) string {
	input := content
	if input == "" {
		input = nodeType
	}
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h)
}

// ComputeRootMerkle computes the entity-level root hash from all node hashes.
// Sorts hashes lexicographically then hashes the concatenation.
func ComputeRootMerkle(nodeHashes []string) string {
	if len(nodeHashes) == 0 {
		return ""
	}
	sorted := make([]string, len(nodeHashes))
	copy(sorted, nodeHashes)
	sort.Strings(sorted)
	h := sha256.Sum256([]byte(strings.Join(sorted, "")))
	return fmt.Sprintf("%x", h)
}

// HashNodes recomputes hashes from content for all nodes and returns the root merkle.
// This mutates each node's Hash field.
func HashNodes(nodes []*Node) string {
	hashes := make([]string, len(nodes))
	for i, n := range nodes {
		n.Hash = ComputeNodeHash(n.Content, n.Type)
		hashes[i] = n.Hash
	}
	return ComputeRootMerkle(hashes)
}
