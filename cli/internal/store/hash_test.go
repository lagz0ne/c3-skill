package store

import (
	"testing"
)

func TestComputeNodeHash_Content(t *testing.T) {
	h1 := ComputeNodeHash("Authenticate requests", "paragraph")
	h2 := ComputeNodeHash("Authenticate requests", "paragraph")
	if h1 != h2 {
		t.Error("same content should produce same hash")
	}

	h3 := ComputeNodeHash("Different content", "paragraph")
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}

	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64 (sha256 hex)", len(h1))
	}
}

func TestComputeNodeHash_EmptyContent(t *testing.T) {
	// Container nodes with empty content hash the type name.
	h1 := ComputeNodeHash("", "table")
	h2 := ComputeNodeHash("", "list")
	if h1 == h2 {
		t.Error("different types should produce different hashes")
	}
}

func TestComputeRootMerkle(t *testing.T) {
	hashes := []string{"abc", "def", "ghi"}
	m1 := ComputeRootMerkle(hashes)
	if m1 == "" {
		t.Error("expected non-empty merkle")
	}

	// Order-independent (sorted internally).
	m2 := ComputeRootMerkle([]string{"ghi", "abc", "def"})
	if m1 != m2 {
		t.Error("merkle should be order-independent")
	}

	// Different hashes produce different merkle.
	m3 := ComputeRootMerkle([]string{"abc", "def", "xyz"})
	if m1 == m3 {
		t.Error("different hashes should produce different merkle")
	}
}

func TestComputeRootMerkle_Empty(t *testing.T) {
	m := ComputeRootMerkle(nil)
	if m != "" {
		t.Errorf("empty hashes should produce empty merkle, got %q", m)
	}
}

func TestHashNodes(t *testing.T) {
	nodes := []*Node{
		{Type: "heading", Content: "Goal"},
		{Type: "paragraph", Content: "Authenticate requests"},
		{Type: "table", Content: ""},
	}

	merkle := HashNodes(nodes)
	if merkle == "" {
		t.Error("expected non-empty merkle")
	}

	// Each node should have its hash set.
	for i, n := range nodes {
		if n.Hash == "" {
			t.Errorf("node %d hash empty", i)
		}
	}

	// Container node hash should differ from content node.
	if nodes[0].Hash == nodes[2].Hash {
		t.Error("heading and empty table should have different hashes")
	}
}
