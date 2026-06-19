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
	// (b) a non-resolving cite of any shape flags ungrounded — by resolution, not shape.
	if !refIssuesHave(cite("ref-nope"), "ungrounded") {
		t.Error("a non-resolving reference must flag 'ungrounded'")
	}
	// (c) pure prose grounds nothing → ungrounded.
	if !refIssuesHave(cite("the checkout flow"), "ungrounded") {
		t.Error("a pure-prose reference cell must be flagged ungrounded")
	}
}

// A custom-type id whose tail contains a builtin prefix — `a11y-rule-ar-focus-visible`
// contains `rule-ar-focus-visible` — must not be split into a phantom builtin reference
// and flagged. The whole token resolves; the embedded substring is not a separate cite.
// (Regression: the design-system eval wired uc-button → a11y-rule-ar-focus-visible and
// got a false "unknown entity reference: rule-ar-focus-visible".)
func TestValidateColumn_ReferenceCustomTypeWithBuiltinTail(t *testing.T) {
	s, _ := openStoreC3(t)
	mustInsert(t, s, &store.Entity{ID: "a11y-rule-ar-focus-visible", Type: "a11y-rule", Title: "Focus visible", Status: "active", Metadata: "{}"})
	col := schema.ColumnDef{Name: "Follows", Type: "reference", Edge: "follows"}
	tbl := &markdown.Table{Headers: []string{"Follows"}, Rows: []map[string]string{{"Follows": "a11y-rule-ar-focus-visible"}}}
	issues := validateColumn(col, tbl, &store.Entity{ID: "uc-button", Type: "ui-component"}, CheckOptions{Store: s}, map[string]string{})
	for _, is := range issues {
		t.Errorf("a resolvable custom-type id with a builtin tail must not warn, got: %s", is.Message)
	}
	// A genuinely broken builtin-shaped cite is still caught.
	bad := validateColumn(col, &markdown.Table{Headers: []string{"Follows"}, Rows: []map[string]string{{"Follows": "rule-does-not-exist"}}},
		&store.Entity{ID: "uc-button", Type: "ui-component"}, CheckOptions{Store: s}, map[string]string{})
	if !refIssuesHave(bad, "ungrounded") {
		t.Error("a non-resolving builtin-shaped id must still flag 'ungrounded'")
	}
}

// An `entity_id` column honors the `N.A - <reason>` sentinel, like every other typed
// reference column — a docs-only row that intentionally names no entity is not an
// "unknown entity reference". (Regression: c3-0 membership row "N.A - docs-only".)
func TestValidateColumn_EntityIDHonorsNASentinel(t *testing.T) {
	s, _ := openStoreC3(t)
	col := schema.ColumnDef{Name: "ID", Type: "entity_id"}
	tbl := &markdown.Table{Headers: []string{"ID"}, Rows: []map[string]string{{"ID": "N.A - docs-only"}}}
	issues := validateColumn(col, tbl, &store.Entity{ID: "c3-0", Type: "system"}, CheckOptions{Store: s}, map[string]string{})
	if refIssuesHave(issues, "unknown entity reference") {
		t.Errorf("an entity_id N.A sentinel must not warn, got: %v", issues)
	}
	// A real unknown id in the same column is still caught.
	bad := validateColumn(col, &markdown.Table{Headers: []string{"ID"}, Rows: []map[string]string{{"ID": "c3-999"}}},
		&store.Entity{ID: "c3-0", Type: "system"}, CheckOptions{Store: s}, map[string]string{})
	if !refIssuesHave(bad, "unknown entity reference") {
		t.Error("a real unknown entity_id must still flag 'unknown entity reference'")
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
