package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// A `reference` edge column grounds by RESOLUTION, not by a builtin id-shape pattern:
// a cite to a custom-type fact (a requirement / objective / design-token id that does
// not look like c3-/ref-/rule-) must ground when it resolves — that is what lets a
// custom canvas (QA/design/PM doc types) wire a real traceability graph instead of
// drowning in false "ungrounded" warnings.
func TestValidateColumn_ReferenceGroundsByResolution(t *testing.T) {
	s, _ := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "req-checkout-01", Type: "requirement", Title: "Checkout completes", Status: "active", Metadata: "{}"})
	col := schema.ColumnDef{Name: "Verifies", Type: "reference", Edge: "verifies"}
	cite := func(v string) []Issue {
		tbl := &markdown.Table{Headers: []string{"Verifies"}, Rows: []map[string]string{{"Verifies": v}}}
		return validateColumn(col, tbl, &store.Entity{ID: "tc-1", Type: "test-case"}, CheckOptions{Store: s}, map[string]string{})
	}

	// (a) resolvable custom-type id → grounded, no warning.
	for _, is := range cite("req-checkout-01") {
		if strings.Contains(is.Message, "ungrounded") {
			t.Errorf("a cite to a resolvable custom-type fact must ground, got: %s", is.Message)
		}
	}
	// (b) a builtin-shaped id that does not resolve still flags unknown.
	if !refIssuesHave(cite("ref-nope"), "unknown entity reference") {
		t.Error("a fake builtin-shaped id must still flag 'unknown entity reference'")
	}
	// (c) pure prose grounds nothing → ungrounded.
	if !refIssuesHave(cite("the checkout flow"), "ungrounded") {
		t.Error("a pure-prose reference cell must be flagged ungrounded")
	}
}

func refIssuesHave(issues []Issue, sub string) bool {
	for _, is := range issues {
		if strings.Contains(is.Message, sub) {
			return true
		}
	}
	return false
}
