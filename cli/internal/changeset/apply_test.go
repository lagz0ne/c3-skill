package changeset

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func openMem(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func seedFact(t *testing.T, s *store.Store, id, body string) {
	t.Helper()
	if err := s.InsertEntity(&store.Entity{ID: id, Type: "component", Title: id, Status: "active", Metadata: "{}"}); err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
	if err := content.WriteEntity(s, id, body); err != nil {
		t.Fatalf("write %s: %v", id, err)
	}
}

// blockHandle returns the cite handle + live hash for the paragraph node whose
// content contains snippet.
func blockHandle(t *testing.T, s *store.Store, id, snippet string) (handle, hash string) {
	t.Helper()
	entity, err := s.GetEntity(id)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := s.NodesForEntity(id)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		if n.Type == "paragraph" && strings.Contains(n.Content, snippet) {
			return fmt.Sprintf("%s#n%d@v%d:sha256:%s", entity.ID, n.ID, entity.Version, n.Hash), n.Hash
		}
	}
	t.Fatalf("no paragraph containing %q in %s", snippet, id)
	return "", ""
}

func nodeHashOf(t *testing.T, s *store.Store, id, snippet string) string {
	t.Helper()
	nodes, _ := s.NodesForEntity(id)
	for _, n := range nodes {
		if strings.Contains(n.Content, snippet) {
			return n.Hash
		}
	}
	return ""
}

func entityMerkle(t *testing.T, s *store.Store, id string) string {
	t.Helper()
	e, err := s.GetEntity(id)
	if err != nil {
		t.Fatal(err)
	}
	return e.RootMerkle
}

func TestApply_BlockFlip_SiblingsFrozen(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nOriginal goal.\n\n## Detail\n\nDetail body.\n")
	handle, _ := blockHandle(t, s, "c3-101", "Original goal.")
	beforeSibling := nodeHashOf(t, s, "c3-101", "Detail body.")
	beforeMerkle := entityMerkle(t, s, "c3-101")

	p := Patch{Target: "c3-101", Scope: ScopeBlock, Base: handle, Content: "Updated goal text.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Target block changed.
	if nodeHashOf(t, s, "c3-101", "Updated goal text.") == "" {
		t.Error("target block content was not updated")
	}
	// Sibling block frozen (same hash).
	if after := nodeHashOf(t, s, "c3-101", "Detail body."); after != beforeSibling {
		t.Errorf("sibling block hash must be unchanged: before %s after %s", beforeSibling, after)
	}
	// Entity seal moved.
	if entityMerkle(t, s, "c3-101") == beforeMerkle {
		t.Error("entity root merkle must change after a block flip")
	}
}

func TestApply_Drift_RejectsWholeSet(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nOriginal goal.\n")
	handle, liveHash := blockHandle(t, s, "c3-101", "Original goal.")
	stale := strings.Replace(handle, liveHash, strings.Repeat("0", 64), 1)

	p := Patch{Target: "c3-101", Scope: ScopeBlock, Base: stale, Content: "Should not land.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}); err == nil {
		t.Fatal("a drifted anchor must reject the apply")
	}
	if nodeHashOf(t, s, "c3-101", "Should not land.") != "" {
		t.Error("drifted patch must not modify the target")
	}
	if nodeHashOf(t, s, "c3-101", "Original goal.") == "" {
		t.Error("original content must be intact after a rejected apply")
	}
}

func TestApply_Atomic_OneDriftedBlocksAll(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nAuth goal.\n")
	seedFact(t, s, "c3-110", "# users\n\n## Goal\n\nUsers goal.\n")
	freshHandle, _ := blockHandle(t, s, "c3-101", "Auth goal.")
	usersHandle, usersHash := blockHandle(t, s, "c3-110", "Users goal.")
	staleUsers := strings.Replace(usersHandle, usersHash, strings.Repeat("0", 64), 1)

	patches := []Patch{
		{Target: "c3-101", Scope: ScopeBlock, Base: freshHandle, Content: "New auth goal.", Source: "01.patch.md"},
		{Target: "c3-110", Scope: ScopeBlock, Base: staleUsers, Content: "New users goal.", Source: "02.patch.md"},
	}
	if err := Apply(s, patches); err == nil {
		t.Fatal("one drifted patch must block the whole set")
	}
	// The fresh target must NOT have been written (atomic).
	if nodeHashOf(t, s, "c3-101", "New auth goal.") != "" {
		t.Error("atomic: c3-101 must be unchanged when a sibling patch drifts")
	}
}
