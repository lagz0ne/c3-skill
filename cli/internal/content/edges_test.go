package content

import (
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// TestDeclaredEdges_ExtractsCandidatesWithTargets — DeclaredEdges is pure
// extraction: every id in the Governance reference column becomes a candidate
// `uses` edge carrying the column's target restriction. (Type filtering — e.g.
// dropping the container cited as policy — happens in the sync, which has the
// store; see TestSync below.) A custom-typed id is extracted too, not dropped.
func TestDeclaredEdges_ExtractsCandidatesWithTargets(t *testing.T) {
	def, ok := schema.DefinitionFor("component")
	if !ok {
		t.Fatal("component canvas should be embedded")
	}
	body := "## Governance\n\n" +
		"| Reference | Type | Governs | Precedence | Notes |\n" +
		"| --- | --- | --- | --- | --- |\n" +
		"| ref-jwt | ref | auth | x | y |\n" +
		"| c3-1 | policy | container | x | y |\n" +
		"| decision-log-api | spec | custom type | x | y |\n"

	got := map[string]DeclaredEdge{}
	for _, e := range DeclaredEdges("c3-101", def, body) {
		got[e.To] = e
	}
	for _, id := range []string{"ref-jwt", "c3-1", "decision-log-api"} {
		e, ok := got[id]
		if !ok {
			t.Fatalf("expected %s extracted as a candidate, got %+v", id, got)
		}
		if e.Rel != "uses" {
			t.Errorf("%s: rel should be uses, got %q", id, e.Rel)
		}
		if len(e.Targets) != 2 {
			t.Errorf("%s: should carry the column targets [ref rule], got %v", id, e.Targets)
		}
	}
}

// TestSyncCanvasOwnedRelationships_FiltersByActualType — the sync wires only
// citations whose ACTUAL stored type matches the column's targets: ref-jwt (ref)
// wires; c3-1 (container, cited as policy) does not.
func TestSyncCanvasOwnedRelationships_FiltersByActualType(t *testing.T) {
	s := testStore(t)
	for _, e := range []*store.Entity{
		{ID: "c3-101", Type: "component", Title: "comp", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "ctr", Status: "active", Metadata: "{}"},
		{ID: "ref-jwt", Type: "ref", Title: "ref", Status: "active", Metadata: "{}"},
	} {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed %s: %v", e.ID, err)
		}
	}
	def, _ := schema.DefinitionFor("component")
	body := "## Governance\n\n" +
		"| Reference | Type | Governs | Precedence | Notes |\n" +
		"| --- | --- | --- | --- | --- |\n" +
		"| ref-jwt | ref | auth | x | y |\n" +
		"| c3-1 | policy | container | x | y |\n"
	if err := SyncCanvasOwnedRelationships(s, "c3-101", def, body); err != nil {
		t.Fatalf("sync: %v", err)
	}
	rels, _ := s.RelationshipsFrom("c3-101")
	to := map[string]bool{}
	for _, r := range rels {
		if r.RelType == "uses" {
			to[r.ToID] = true
		}
	}
	if !to["ref-jwt"] {
		t.Error("ref-jwt (type ref) must wire a uses edge")
	}
	if to["c3-1"] {
		t.Error("c3-1 (type container) must NOT wire a uses edge under targets [ref,rule]")
	}
}

// TestDeclaredEdges_SkipsNAAndBlank — "N.A - <reason>" and blank cells wire
// nothing.
func TestDeclaredEdges_SkipsNAAndBlank(t *testing.T) {
	def, _ := schema.DefinitionFor("component")
	body := "## Governance\n\n" +
		"| Reference | Type | Governs | Precedence | Notes |\n" +
		"| --- | --- | --- | --- | --- |\n" +
		"| N.A - no refs yet | N.A - <reason> | n/a | n/a | seed |\n"
	if edges := DeclaredEdges("c3-101", def, body); len(edges) != 0 {
		t.Errorf("N.A citation must wire nothing, got %+v", edges)
	}
}

// TestCanvasOwnedRelTypes — the component canvas owns `uses` (Governance.Reference
// edge: uses); a canvas with no edge-column owns nothing (legacy frontmatter path).
func TestCanvasOwnedRelTypes(t *testing.T) {
	comp, _ := schema.DefinitionFor("component")
	if owned := CanvasOwnedRelTypes(comp); !owned["uses"] {
		t.Errorf("component canvas should own 'uses', got %v", owned)
	}
	sys, _ := schema.DefinitionFor("system")
	if owned := CanvasOwnedRelTypes(sys); len(owned) != 0 {
		t.Errorf("system canvas declares no edge-column, should own nothing, got %v", owned)
	}
}
