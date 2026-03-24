package content

import (
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func seedEntity(t *testing.T, s *store.Store, id, typ string) {
	t.Helper()
	err := s.InsertEntity(&store.Entity{
		ID: id, Type: typ, Title: "Test", Slug: "test", Status: "active", Metadata: "{}",
	})
	if err != nil {
		t.Fatalf("seed entity: %v", err)
	}
}

func TestWriteEntity_StoresNodes(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	md := "## Goal\n\nServe requests\n\n## Dependencies\n\n- auth\n- db\n"
	if err := WriteEntity(s, "test-1", md); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	nodes, err := s.NodesForEntity("test-1")
	if err != nil {
		t.Fatalf("NodesForEntity: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected nodes in DB, got 0")
	}

	// Should have headings, paragraph, list, list items
	types := map[string]int{}
	for _, n := range nodes {
		types[n.Type]++
	}
	if types["heading"] < 2 {
		t.Errorf("expected at least 2 headings, got %d", types["heading"])
	}
	if types["paragraph"] < 1 {
		t.Errorf("expected at least 1 paragraph, got %d", types["paragraph"])
	}
}

func TestWriteEntity_CreatesVersion(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	md := "## Goal\n\nServe requests\n"
	if err := WriteEntity(s, "test-1", md); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	v, err := s.GetVersion("test-1", 1)
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if v.RootMerkle == "" {
		t.Error("expected non-empty root merkle on version")
	}
	if v.Content == "" {
		t.Error("expected non-empty content snapshot on version")
	}
}

func TestWriteEntity_UpdatesMerkle(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	md := "## Goal\n\nServe requests\n"
	if err := WriteEntity(s, "test-1", md); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	e, err := s.GetEntity("test-1")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if e.RootMerkle == "" {
		t.Error("expected entity.RootMerkle to be set")
	}
	if e.Version != 1 {
		t.Errorf("expected entity.Version=1, got %d", e.Version)
	}
}

func TestWriteEntity_SyncsGoal(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	md := "## Goal\n\nServe API requests\n\n## Dependencies\n\n- auth\n"
	if err := WriteEntity(s, "test-1", md); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	e, err := s.GetEntity("test-1")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if e.Goal != "Serve API requests" {
		t.Errorf("expected goal 'Serve API requests', got %q", e.Goal)
	}
}

func TestWriteEntity_SecondWrite(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	if err := WriteEntity(s, "test-1", "## Goal\n\nFirst\n"); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteEntity(s, "test-1", "## Goal\n\nSecond\n"); err != nil {
		t.Fatalf("second write: %v", err)
	}

	v, err := s.LatestVersion("test-1")
	if err != nil {
		t.Fatalf("LatestVersion: %v", err)
	}
	if v != 2 {
		t.Errorf("expected version 2 after two writes, got %d", v)
	}

	// Nodes should reflect second write only
	nodes, err := s.NodesForEntity("test-1")
	if err != nil {
		t.Fatalf("NodesForEntity: %v", err)
	}
	found := false
	for _, n := range nodes {
		if n.Type == "paragraph" && n.Content == "Second" {
			found = true
		}
	}
	if !found {
		t.Error("expected paragraph 'Second' in nodes after second write")
	}
}

func TestReadEntity_RoundTrip(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	md := "## Goal\n\nServe requests\n\n## Dependencies\n\n- auth\n- db\n"
	if err := WriteEntity(s, "test-1", md); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	got, err := ReadEntity(s, "test-1")
	if err != nil {
		t.Fatalf("ReadEntity: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty markdown from ReadEntity")
	}

	// Round-trip: re-parse the output and check structural equivalence
	tree1 := ParseMarkdown("test-1", md)
	tree2 := ParseMarkdown("test-1", got)

	if len(tree1.Nodes) != len(tree2.Nodes) {
		t.Errorf("node count mismatch: input=%d, round-trip=%d", len(tree1.Nodes), len(tree2.Nodes))
	}

	// Check merkle equivalence
	merkle1 := store.HashNodes(tree1.Nodes)
	merkle2 := store.HashNodes(tree2.Nodes)
	if merkle1 != merkle2 {
		t.Errorf("merkle mismatch: input=%s, round-trip=%s", merkle1, merkle2)
	}
}

func TestReadEntity_Empty(t *testing.T) {
	s := testStore(t)
	seedEntity(t, s, "test-1", "component")

	got, err := ReadEntity(s, "test-1")
	if err != nil {
		t.Fatalf("ReadEntity: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for entity with no nodes, got %q", got)
	}
}
