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
	if err := Apply(s, []Patch{p}, nil, nil); err != nil {
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

// A frontmatter patch can now edit a frozen fact's boundary / category / date —
// parity with `set`, which the change-unit path previously could not reach.
func TestApply_Frontmatter_SetsAttributes(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nGoal.\n")
	base := entityHandle(t, s, "c3-101")

	p := Patch{Target: "c3-101", Scope: ScopeFrontmatter, Base: base,
		Boundary: "owns auth only", Category: "feature", Date: "2026-06-18", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err != nil {
		t.Fatalf("apply: %v", err)
	}

	e, _ := s.GetEntity("c3-101")
	if e.Boundary != "owns auth only" || e.Category != "feature" || e.Date != "2026-06-18" {
		t.Errorf("frontmatter patch did not set attributes: boundary=%q category=%q date=%q", e.Boundary, e.Category, e.Date)
	}
}

// A table-row block patch accepts the natural markdown forms an author writes and
// reduces them to the bare " | "-joined cells a table_row node stores.
func TestNormalizeTableRowContent(t *testing.T) {
	cases := map[string]string{
		"| a | b | c |":          "a | b | c", // outer pipes stripped
		"a | b | c":              "a | b | c", // already bare
		"  |  a |  b  | ":        "a | b",      // extra whitespace trimmed
		"| x |":                  "x",          // single cell
		"\n| a | b |\n":          "a | b",       // leading/trailing newlines
		"| h1 | h2 |\n| --- | --- |": "h1 | h2", // separator line skipped
	}
	for in, want := range cases {
		if got := normalizeTableRowContent(in); got != want {
			t.Errorf("normalizeTableRowContent(%q) = %q, want %q", in, got, want)
		}
	}
}

// An empty block body DELETES the cited node — "empty body deletes the block".
func TestApply_BlockEmpty_DeletesNode(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nKeep this goal.\n\n## Detail\n\nDelete this detail.\n")
	handle, _ := blockHandle(t, s, "c3-101", "Delete this detail.")

	p := Patch{Target: "c3-101", Scope: ScopeBlock, Base: handle, Content: "", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if nodeHashOf(t, s, "c3-101", "Delete this detail.") != "" {
		t.Error("an empty block patch must DELETE the cited node, not blank it")
	}
	if nodeHashOf(t, s, "c3-101", "Keep this goal.") == "" {
		t.Error("a sibling block must remain after deleting another")
	}
}

func TestApply_Drift_RejectsWholeSet(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nOriginal goal.\n")
	handle, liveHash := blockHandle(t, s, "c3-101", "Original goal.")
	stale := strings.Replace(handle, liveHash, strings.Repeat("0", 64), 1)

	p := Patch{Target: "c3-101", Scope: ScopeBlock, Base: stale, Content: "Should not land.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err == nil {
		t.Fatal("a drifted anchor must reject the apply")
	}
	if nodeHashOf(t, s, "c3-101", "Should not land.") != "" {
		t.Error("drifted patch must not modify the target")
	}
	if nodeHashOf(t, s, "c3-101", "Original goal.") == "" {
		t.Error("original content must be intact after a rejected apply")
	}
}

// TestApply_Atomic_MidApplyLandingMismatchRollsBack exercises the integrity crux:
// a failure that surfaces DURING the write phase (not in the drift preflight). Both
// patches anchor fresh, so the drift gate passes and apply enters the write loop.
// Patch 1 writes c3-101; patch 2 then fails its landing-hash check. Without a single
// transaction, patch 1 would already be committed (half-applied set). With it, the
// whole change-unit rolls back.
func TestApply_Atomic_MidApplyLandingMismatchRollsBack(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nAuth goal.\n")
	seedFact(t, s, "c3-110", "# users\n\n## Goal\n\nUsers goal.\n")
	authHandle, _ := blockHandle(t, s, "c3-101", "Auth goal.")
	usersHandle, _ := blockHandle(t, s, "c3-110", "Users goal.")
	beforeAuthMerkle := entityMerkle(t, s, "c3-101")

	patches := []Patch{
		// Lands cleanly — no result hash to verify.
		{Target: "c3-101", Scope: ScopeBlock, Base: authHandle, Content: "New auth goal.", Source: "01.patch.md"},
		// Fresh anchor (passes drift) but a bogus result hash → fails the landing
		// check inside applyBlock, AFTER patch 1 has written.
		{Target: "c3-110", Scope: ScopeBlock, Base: usersHandle, Content: "New users goal.", Result: "sha256:" + strings.Repeat("b", 64), Source: "02.patch.md"},
	}
	err := Apply(s, patches, nil, nil)
	if err == nil {
		t.Fatal("a landing-hash mismatch on patch 2 must fail the apply")
	}
	if !strings.Contains(err.Error(), "landing mismatch") {
		t.Fatalf("expected a landing mismatch error, got: %v", err)
	}
	// Patch 1's write must have been rolled back with the failed transaction.
	if nodeHashOf(t, s, "c3-101", "New auth goal.") != "" {
		t.Error("atomic: c3-101's new content must NOT persist when patch 2 fails mid-apply")
	}
	if nodeHashOf(t, s, "c3-101", "Auth goal.") == "" {
		t.Error("atomic: c3-101's original content must be intact after rollback")
	}
	if entityMerkle(t, s, "c3-101") != beforeAuthMerkle {
		t.Errorf("atomic: c3-101's seal must be unchanged after rollback")
	}
	// And c3-110 itself stayed put.
	if nodeHashOf(t, s, "c3-110", "New users goal.") != "" {
		t.Error("atomic: c3-110 must be unchanged after a rejected apply")
	}
}

// TestApply_Atomic_DuplicateBaseRejected covers the write-time re-anchor: two
// patches that cite the SAME fresh block base both pass the drift preflight, but
// once patch 1 rewrites the block, patch 2's base no longer matches the live in-tx
// node. It must be rejected (not silently clobber patch 1), and the whole unit
// rolls back.
func TestApply_Atomic_DuplicateBaseRejected(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nOriginal goal.\n")
	handle, _ := blockHandle(t, s, "c3-101", "Original goal.")
	beforeMerkle := entityMerkle(t, s, "c3-101")

	patches := []Patch{
		{Target: "c3-101", Scope: ScopeBlock, Base: handle, Content: "First rewrite.", Source: "01.patch.md"},
		{Target: "c3-101", Scope: ScopeBlock, Base: handle, Content: "Second rewrite.", Source: "02.patch.md"},
	}
	err := Apply(s, patches, nil, nil)
	if err == nil {
		t.Fatal("two patches citing the same base must not both apply")
	}
	if !strings.Contains(err.Error(), "changed before apply") {
		t.Fatalf("expected a write-time re-anchor rejection, got: %v", err)
	}
	if nodeHashOf(t, s, "c3-101", "First rewrite.") != "" {
		t.Error("atomic: patch 1 must roll back when patch 2 is rejected")
	}
	if nodeHashOf(t, s, "c3-101", "Second rewrite.") != "" {
		t.Error("patch 2 must not apply")
	}
	if nodeHashOf(t, s, "c3-101", "Original goal.") == "" {
		t.Error("original content must be intact after rollback")
	}
	if entityMerkle(t, s, "c3-101") != beforeMerkle {
		t.Error("seal must be unchanged after rollback")
	}
}

func entityHandle(t *testing.T, s *store.Store, id string) string {
	t.Helper()
	e, err := s.GetEntity(id)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%s@v%d:sha256:%s", e.ID, e.Version, e.RootMerkle)
}

// TestApply_Insert_AppendsSectionSiblingsFrozen is the climb's core property: an
// insert adds a new section at the end, leaves every existing block hash-frozen, and
// reseals the entity — so a fact can gain a section (e.g. when a canvas rung rises)
// without drifting anything it already had.
func TestApply_Insert_AppendsSectionSiblingsFrozen(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nOriginal goal.\n\n## Detail\n\nDetail body.\n")
	beforeGoal := nodeHashOf(t, s, "c3-101", "Original goal.")
	beforeDetail := nodeHashOf(t, s, "c3-101", "Detail body.")
	beforeMerkle := entityMerkle(t, s, "c3-101")

	p := Patch{Target: "c3-101", Scope: ScopeInsert, Base: entityHandle(t, s, "c3-101"), Content: "## Contract\n\nTokens must be RS256.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err != nil {
		t.Fatalf("insert apply: %v", err)
	}
	body, _ := content.ReadEntity(s, "c3-101")
	if nodeHashOf(t, s, "c3-101", "Tokens must be RS256.") == "" || !strings.Contains(body, "## Contract") {
		t.Errorf("inserted section missing:\n%s", body)
	}
	if nodeHashOf(t, s, "c3-101", "Original goal.") != beforeGoal || nodeHashOf(t, s, "c3-101", "Detail body.") != beforeDetail {
		t.Error("existing blocks must stay frozen across an insert")
	}
	if entityMerkle(t, s, "c3-101") == beforeMerkle {
		t.Error("entity merkle must change after an insert (content was added)")
	}
	if i, j := strings.Index(body, "## Detail"), strings.Index(body, "## Contract"); i < 0 || j < i {
		t.Errorf("inserted section must append at the end:\n%s", body)
	}
}

func TestApply_Insert_StaleAnchorRebases(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nGoal.\n")
	h := entityHandle(t, s, "c3-101")
	stale := strings.Replace(h, h[strings.LastIndex(h, ":")+1:], strings.Repeat("0", 64), 1)
	p := Patch{Target: "c3-101", Scope: ScopeInsert, Base: stale, Content: "## X\n\nbody.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err == nil {
		t.Fatal("a stale entity anchor must reject the insert")
	}
	if nodeHashOf(t, s, "c3-101", "body.") != "" {
		t.Error("a rejected insert must not modify the fact")
	}
}

func TestApply_Insert_RollsBackWithFailedSibling(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nGoal.\n")
	beforeMerkle := entityMerkle(t, s, "c3-101")
	ih := entityHandle(t, s, "c3-101")
	bh, _ := blockHandle(t, s, "c3-101", "Goal.")
	patches := []Patch{
		{Target: "c3-101", Scope: ScopeInsert, Base: ih, Content: "## New\n\nbody.", Source: "01.patch.md"},
		{Target: "c3-101", Scope: ScopeBlock, Base: bh, Content: "won't seal", Result: "sha256:" + strings.Repeat("d", 64), Source: "02.patch.md"},
	}
	if err := Apply(s, patches, nil, nil); err == nil {
		t.Fatal("the landing mismatch on patch 2 must fail the whole unit")
	}
	if nodeHashOf(t, s, "c3-101", "body.") != "" {
		t.Error("the inserted section must roll back when a sibling patch fails")
	}
	if entityMerkle(t, s, "c3-101") != beforeMerkle {
		t.Error("merkle must be unchanged after rollback")
	}
}

func TestApply_Insert_RejectsHeadinglessBody(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nGoal.\n")
	p := Patch{Target: "c3-101", Scope: ScopeInsert, Base: entityHandle(t, s, "c3-101"), Content: "Just a loose paragraph, no heading.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err == nil || !strings.Contains(err.Error(), "section heading") {
		t.Fatalf("a headingless insert body must be rejected (it would land a stray root paragraph), got: %v", err)
	}
}

func TestApply_Insert_RejectsDuplicateSection(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nGoal.\n")
	p := Patch{Target: "c3-101", Scope: ScopeInsert, Base: entityHandle(t, s, "c3-101"), Content: "## Goal\n\nDuplicate goal.", Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("inserting an existing section must be rejected, got: %v", err)
	}
}

func TestApply_Insert_ResultLandingCheck(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nGoal.\n")
	p := Patch{Target: "c3-101", Scope: ScopeInsert, Base: entityHandle(t, s, "c3-101"), Content: "## Contract\n\nbody.", Result: "sha256:" + strings.Repeat("e", 64), Source: "01.patch.md"}
	if err := Apply(s, []Patch{p}, nil, nil); err == nil || !strings.Contains(err.Error(), "landing mismatch") {
		t.Fatalf("a bogus result on an insert must fail the landing check, got: %v", err)
	}
	if nodeHashOf(t, s, "c3-101", "body.") != "" {
		t.Error("a landing-mismatch insert must roll back")
	}
}

// TestApply_CodemapCarrierAppliesInSameUnit proves a unit's internal patch and its
// external codemap carrier both land in one apply.
func TestApply_CodemapCarrierAppliesInSameUnit(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nAuth goal.\n")
	handle, _ := blockHandle(t, s, "c3-101", "Auth goal.")

	patches := []Patch{{Target: "c3-101", Scope: ScopeBlock, Base: handle, Content: "New auth goal.", Source: "01.patch.md"}}
	codemaps := []CodemapChange{{Target: "c3-101", Globs: []string{"src/auth/**"}, Source: "01.codemap.md"}}
	if err := Apply(s, patches, codemaps, nil); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if nodeHashOf(t, s, "c3-101", "New auth goal.") == "" {
		t.Error("the patch did not apply")
	}
	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 1 || patterns[0] != "src/auth/**" {
		t.Errorf("codemap = %v, want [src/auth/**]", patterns)
	}
}

// TestApply_PatchRollsBackWhenCodemapFails is the cross-arm integrity proof: an
// internal fact write (patch) and an external binding write (codemap) are one unit.
// The patch applies, then the codemap carrier fails (FK — its target does not
// exist). The whole transaction rolls back, so the internal fact is NOT left
// changed while its external binding never landed.
func TestApply_PatchRollsBackWhenCodemapFails(t *testing.T) {
	s := openMem(t)
	seedFact(t, s, "c3-101", "# auth\n\n## Goal\n\nAuth goal.\n")
	handle, _ := blockHandle(t, s, "c3-101", "Auth goal.")
	beforeMerkle := entityMerkle(t, s, "c3-101")

	patches := []Patch{{Target: "c3-101", Scope: ScopeBlock, Base: handle, Content: "New auth goal.", Source: "01.patch.md"}}
	codemaps := []CodemapChange{{Target: "c3-999-missing", Globs: []string{"src/x/**"}, Source: "01.codemap.md"}}
	if err := Apply(s, patches, codemaps, nil); err == nil {
		t.Fatal("a codemap carrier with a missing target must fail the apply")
	}
	if nodeHashOf(t, s, "c3-101", "New auth goal.") != "" {
		t.Error("atomic: the patch must roll back when a later codemap write fails")
	}
	if entityMerkle(t, s, "c3-101") != beforeMerkle {
		t.Error("atomic: the entity seal must be unchanged after the cross-arm rollback")
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
	if err := Apply(s, patches, nil, nil); err == nil {
		t.Fatal("one drifted patch must block the whole set")
	}
	// The fresh target must NOT have been written (atomic).
	if nodeHashOf(t, s, "c3-101", "New auth goal.") != "" {
		t.Error("atomic: c3-101 must be unchanged when a sibling patch drifts")
	}
}
