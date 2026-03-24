package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// seedNodes inserts a set of nodes for entity c3-101 into the store.
// Returns the nodes for verification.
func seedNodes(t *testing.T, s *store.Store) []*store.Node {
	t.Helper()
	// Insert top-level nodes first (unique per parent_id=NULL + seq)
	h1 := &store.Node{EntityID: "c3-101", Type: "heading", Level: 2, Seq: 0, Content: "Goal", Hash: "aaa1"}
	h2 := &store.Node{EntityID: "c3-101", Type: "heading", Level: 2, Seq: 1, Content: "Dependencies", Hash: "aaa3"}
	for _, n := range []*store.Node{h1, h2} {
		if _, err := s.InsertNode(n); err != nil {
			t.Fatalf("seed node: %v", err)
		}
	}
	// Insert children (parent_id set, seq unique within parent)
	p := &store.Node{EntityID: "c3-101", Type: "paragraph", Level: 0, Seq: 0, Content: "Handle authentication.", Hash: "aaa2",
		ParentID: sql.NullInt64{Int64: h1.ID, Valid: true}}
	tbl := &store.Node{EntityID: "c3-101", Type: "table", Level: 0, Seq: 0, Content: "", Hash: "aaa4",
		ParentID: sql.NullInt64{Int64: h2.ID, Valid: true}}
	for _, n := range []*store.Node{p, tbl} {
		if _, err := s.InsertNode(n); err != nil {
			t.Fatalf("seed node: %v", err)
		}
	}
	return []*store.Node{h1, p, h2, tbl}
}

func TestRunNodes(t *testing.T) {
	s := createDBFixture(t)
	seedNodes(t, s)

	var buf bytes.Buffer
	err := RunNodes(NodesOptions{Store: s, EntityID: "c3-101", JSON: false}, &buf)
	if err != nil {
		t.Fatalf("RunNodes: %v", err)
	}
	out := buf.String()
	// Should contain header line
	if !strings.Contains(out, "ID") || !strings.Contains(out, "TYPE") {
		t.Errorf("expected table header, got:\n%s", out)
	}
	// Should contain our nodes
	if !strings.Contains(out, "heading") {
		t.Errorf("expected heading node, got:\n%s", out)
	}
	if !strings.Contains(out, "Goal") {
		t.Errorf("expected 'Goal' content, got:\n%s", out)
	}
	if !strings.Contains(out, "paragraph") {
		t.Errorf("expected paragraph node, got:\n%s", out)
	}
}

func TestRunNodes_JSON(t *testing.T) {
	s := createDBFixture(t)
	seedNodes(t, s)

	var buf bytes.Buffer
	err := RunNodes(NodesOptions{Store: s, EntityID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatalf("RunNodes JSON: %v", err)
	}
	var nodes []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &nodes); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(nodes))
	}
}

func TestRunNodes_Empty(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunNodes(NodesOptions{Store: s, EntityID: "c3-101", JSON: false}, &buf)
	if err != nil {
		t.Fatalf("RunNodes empty: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "No nodes") {
		t.Errorf("expected 'No nodes' message, got:\n%s", out)
	}
}

func TestRunNodes_MissingEntity(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunNodes(NodesOptions{Store: s, EntityID: "nonexistent", JSON: false}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}
