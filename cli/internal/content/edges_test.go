package content

import (
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// TestDeclaredEdges_GovernanceCitationWiresUses — authoring a ref in the live
// component canvas's Governance reference column declares a `uses` edge; a
// container cited as policy in the same column is filtered out by targets
// [ref, rule] (a component does not "use" its container).
func TestDeclaredEdges_GovernanceCitationWiresUses(t *testing.T) {
	def, ok := schema.DefinitionFor("component")
	if !ok {
		t.Fatal("component canvas should be embedded")
	}
	body := "## Governance\n\n" +
		"| Reference | Type | Governs | Precedence | Notes |\n" +
		"| --- | --- | --- | --- | --- |\n" +
		"| ref-jwt | ref | auth | x | y |\n" +
		"| rule-logging | rule | logs | x | y |\n" +
		"| c3-1 | policy | container | x | y |\n"

	edges := DeclaredEdges("c3-101", def, body)
	got := map[string]string{}
	for _, e := range edges {
		got[e.To] = e.Rel
	}
	if got["ref-jwt"] != "uses" {
		t.Errorf("expected uses edge to ref-jwt, got %+v", edges)
	}
	if got["rule-logging"] != "uses" {
		t.Errorf("expected uses edge to rule-logging, got %+v", edges)
	}
	if _, wired := got["c3-1"]; wired {
		t.Errorf("c3-1 (container, policy) must NOT wire a uses edge (targets [ref,rule]), got %+v", edges)
	}
	if len(edges) != 2 {
		t.Errorf("expected exactly 2 edges (ref + rule), got %d: %+v", len(edges), edges)
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
