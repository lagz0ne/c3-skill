package store

import (
	"database/sql"
	"testing"
)

func TestInsertNode_AndGet(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	n := &Node{
		EntityID: "auth-handler",
		Type:     "heading",
		Level:    2,
		Seq:      0,
		Content:  "Goal",
		Hash:     ComputeNodeHash("Goal", "heading"),
	}
	id, err := s.InsertNode(n)
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := s.GetNode(id)
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if got.EntityID != "auth-handler" || got.Type != "heading" || got.Content != "Goal" {
		t.Errorf("unexpected node: %+v", got)
	}
	if got.Hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestInsertNode_WithParent(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	parent := &Node{EntityID: "auth-handler", Type: "heading", Level: 2, Seq: 0, Content: "Dependencies", Hash: "abc"}
	parentID, err := s.InsertNode(parent)
	if err != nil {
		t.Fatalf("insert parent: %v", err)
	}

	child := &Node{
		EntityID: "auth-handler",
		ParentID: sql.NullInt64{Int64: parentID, Valid: true},
		Type:     "table_row",
		Seq:      0,
		Content:  "IN|auth-svc|c3-102",
		Hash:     "def",
	}
	childID, err := s.InsertNode(child)
	if err != nil {
		t.Fatalf("insert child: %v", err)
	}

	got, err := s.GetNode(childID)
	if err != nil {
		t.Fatalf("get child: %v", err)
	}
	if !got.ParentID.Valid || got.ParentID.Int64 != parentID {
		t.Errorf("parent_id = %v, want %d", got.ParentID, parentID)
	}
}

func TestNodesForEntity(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	nodes := []*Node{
		{EntityID: "auth-handler", Type: "heading", Level: 2, Seq: 0, Content: "Goal", Hash: "a"},
		{EntityID: "auth-handler", Type: "paragraph", Seq: 1, Content: "Authenticate requests", Hash: "b"},
		{EntityID: "auth-handler", Type: "heading", Level: 2, Seq: 2, Content: "Dependencies", Hash: "c"},
	}
	for _, n := range nodes {
		if _, err := s.InsertNode(n); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	got, err := s.NodesForEntity("auth-handler")
	if err != nil {
		t.Fatalf("nodes for entity: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d nodes, want 3", len(got))
	}
	if got[0].Content != "Goal" || got[1].Content != "Authenticate requests" || got[2].Content != "Dependencies" {
		t.Errorf("unexpected order: %v, %v, %v", got[0].Content, got[1].Content, got[2].Content)
	}
}

func TestDeleteNode_CascadesChildren(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	parent := &Node{EntityID: "auth-handler", Type: "list", Seq: 0, Hash: "a"}
	parentID, _ := s.InsertNode(parent)

	child := &Node{
		EntityID: "auth-handler",
		ParentID: sql.NullInt64{Int64: parentID, Valid: true},
		Type:     "list_item",
		Seq:      0,
		Content:  "item 1",
		Hash:     "b",
	}
	childID, _ := s.InsertNode(child)

	if err := s.DeleteNode(parentID); err != nil {
		t.Fatalf("delete parent: %v", err)
	}

	_, err := s.GetNode(childID)
	if err == nil {
		t.Error("expected child to be cascade-deleted")
	}
}

func TestReplaceEntityNodes(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Insert initial nodes.
	s.InsertNode(&Node{EntityID: "auth-handler", Type: "heading", Seq: 0, Content: "Old", Hash: "old"})

	// Replace with new set.
	newNodes := []*Node{
		{Type: "heading", Level: 2, Seq: 0, Content: "Goal", Hash: "a"},
		{Type: "paragraph", Seq: 1, Content: "New goal text", Hash: "b"},
	}
	if err := s.ReplaceEntityNodes("auth-handler", newNodes); err != nil {
		t.Fatalf("replace: %v", err)
	}

	got, err := s.NodesForEntity("auth-handler")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d nodes, want 2", len(got))
	}
	if got[0].Content != "Goal" {
		t.Errorf("got %q, want Goal", got[0].Content)
	}
}

func TestNodeChildren(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	parent := &Node{EntityID: "auth-handler", Type: "table", Seq: 0, Hash: "t"}
	parentID, _ := s.InsertNode(parent)

	for i, content := range []string{"row1", "row2", "row3"} {
		s.InsertNode(&Node{
			EntityID: "auth-handler",
			ParentID: sql.NullInt64{Int64: parentID, Valid: true},
			Type:     "table_row",
			Seq:      i,
			Content:  content,
			Hash:     ComputeNodeHash(content, "table_row"),
		})
	}

	children, err := s.NodeChildren(parentID)
	if err != nil {
		t.Fatalf("children: %v", err)
	}
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3", len(children))
	}
}

func TestUpdateNode(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	n := &Node{EntityID: "auth-handler", Type: "paragraph", Seq: 0, Content: "old", Hash: "old"}
	id, _ := s.InsertNode(n)

	n.ID = id
	n.Content = "new content"
	n.Hash = ComputeNodeHash("new content", "paragraph")
	if err := s.UpdateNode(n); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := s.GetNode(id)
	if got.Content != "new content" {
		t.Errorf("got %q, want 'new content'", got.Content)
	}
}
