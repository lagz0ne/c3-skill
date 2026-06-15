package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// Item 8 — Migration: SWEEP & CLEAR ALL.
//
// The migration visits every entity once: it CLEARS each fact's `active` status
// (facts have no status), MAPS each change doc's legacy status onto the canonical
// set recording the lossy provisioned->done collapse, GRANDFATHERS old terminal
// ADRs to `done` with NO retro success-check, RECONCILES already-materialized
// old-grammar canvases explicitly, and RE-SEALS every entity. Every step is
// itemized in a loud report; an unmappable status FAILS loud. The migration is the
// ONLY path (besides status/supersede/auto-done) that may move the status column,
// and the ONLY path that may rewrite a terminal status.

// seedMigrationFact inserts a fact (non-change-doc) entity carrying the legacy
// `active` status with a small body so re-seal has nodes to hash.
func seedMigrationFact(t *testing.T, s *store.Store, id, status string) {
	t.Helper()
	e := &store.Entity{
		ID:       id,
		Type:     "component",
		Title:    id,
		Slug:     id,
		Category: "feature",
		ParentID: "c3-1",
		Status:   status,
		Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("seed fact %s: %v", id, err)
	}
	if err := content.WriteEntity(s, id, "# "+id+"\n\n## Goal\n\nA fact body.\n"); err != nil {
		t.Fatalf("seed fact body %s: %v", id, err)
	}
}

// seedMigrationADR inserts an ADR change doc carrying a legacy status, with a body.
func seedMigrationADR(t *testing.T, s *store.Store, id, status string) {
	t.Helper()
	e := &store.Entity{
		ID:       id,
		Type:     "adr",
		Title:    id,
		Slug:     id,
		Status:   status,
		Date:     "20260101",
		Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("seed adr %s: %v", id, err)
	}
	if err := content.WriteEntity(s, id, "# "+id+"\n\n## Context\n\nLegacy decision.\n"); err != nil {
		t.Fatalf("seed adr body %s: %v", id, err)
	}
}

// runMigrate runs the migration over the store and returns the report and output.
func runMigrate(t *testing.T, s *store.Store) (MigrationReport, string) {
	t.Helper()
	var buf bytes.Buffer
	report, err := RunMigrate(MigrateOptions{Store: s}, &buf)
	if err != nil {
		t.Fatalf("RunMigrate failed: %v\n%s", err, buf.String())
	}
	return report, buf.String()
}

// findMigrationEntry locates the report entry for an entity id.
func findMigrationEntry(report MigrationReport, id string) (MigrationEntry, bool) {
	for _, e := range report.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return MigrationEntry{}, false
}

// TestMigrate_ClearsFactActiveStatus — migration CLEARS every fact's `active`
// status (facts have no status), itemized in the report — not remapped to a
// change-doc state.
func TestMigrate_ClearsFactActiveStatus(t *testing.T) {
	s := createDBFixture(t) // c3-0/c3-1/c3-2/c3-101/c3-110/ref-jwt all default active
	report, _ := runMigrate(t, s)

	for _, id := range []string{"c3-0", "c3-1", "c3-101", "c3-110", "ref-jwt"} {
		e, err := s.GetEntity(id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		if e.Status != "" {
			t.Errorf("fact %s status = %q after migration, want cleared (empty)", id, e.Status)
		}
		entry, ok := findMigrationEntry(report, id)
		if !ok {
			t.Errorf("fact %s missing from itemized report", id)
			continue
		}
		if entry.Action != "cleared" {
			t.Errorf("fact %s report action = %q, want %q", id, entry.Action, "cleared")
		}
	}
}

// TestMigrate_MapsActiveWithItemizedReport — running migration produces a clear,
// itemized report of what each entity mapped to / was cleared to (no silent
// change). Every visited entity appears in the report with a from/to pair.
func TestMigrate_MapsActiveWithItemizedReport(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-legacy-proposed", "proposed")

	report, out := runMigrate(t, s)

	all, err := s.AllEntities()
	if err != nil {
		t.Fatalf("all entities: %v", err)
	}
	if len(report.Entries) != len(all) {
		t.Errorf("itemized report covers %d entities, want %d (one entry per entity)", len(report.Entries), len(all))
	}

	entry, ok := findMigrationEntry(report, "adr-legacy-proposed")
	if !ok {
		t.Fatalf("adr-legacy-proposed missing from report")
	}
	if entry.From != "proposed" || entry.To != "open" {
		t.Errorf("adr mapping report = %q->%q, want proposed->open", entry.From, entry.To)
	}
	if entry.Action != "mapped" {
		t.Errorf("adr report action = %q, want %q", entry.Action, "mapped")
	}

	// The textual report must be loud / non-silent: it names the mapping.
	if !strings.Contains(out, "adr-legacy-proposed") {
		t.Errorf("loud report should name adr-legacy-proposed, got:\n%s", out)
	}
	if !strings.Contains(out, "proposed") || !strings.Contains(out, "open") {
		t.Errorf("loud report should show proposed->open, got:\n%s", out)
	}
}

// TestMigrate_MapsAdrActiveDoesNotUnfreezeProvisioned — migration maps ADR legacy
// statuses but does NOT silently un-freeze a provisioned ADR. A provisioned ADR is
// rewritten only through the announced privileged path, and lands on the terminal
// `done` (collapsed), never on a non-terminal state.
func TestMigrate_MapsAdrActiveDoesNotUnfreezeProvisioned(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-prov", "provisioned")

	runMigrate(t, s)

	e, err := s.GetEntity("adr-prov")
	if err != nil {
		t.Fatalf("get adr-prov: %v", err)
	}
	// provisioned is terminal; it must collapse to the terminal `done`, never to a
	// non-terminal (open/accepted) state that would un-freeze it.
	if e.Status != "done" {
		t.Errorf("provisioned ADR status = %q after migration, want %q (collapsed terminal)", e.Status, "done")
	}
}

// TestMigrate_ProvisionedReportedLossy — the provisioned->done mapping is reported
// as LOSSY (the provisioned distinction is collapsed, per item 1's flag).
func TestMigrate_ProvisionedReportedLossy(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-prov", "provisioned")

	report, out := runMigrate(t, s)

	entry, ok := findMigrationEntry(report, "adr-prov")
	if !ok {
		t.Fatalf("adr-prov missing from report")
	}
	if !entry.Lossy {
		t.Errorf("provisioned->done entry should be flagged lossy, got lossy=false")
	}
	if entry.To != "done" {
		t.Errorf("provisioned should map to done, got %q", entry.To)
	}
	if !strings.Contains(strings.ToLower(out), "lossy") {
		t.Errorf("loud report should call out the lossy collapse, got:\n%s", out)
	}

	// A non-lossy mapping must NOT be flagged lossy.
	s2 := createDBFixture(t)
	seedMigrationADR(t, s2, "adr-clean", "proposed")
	report2, _ := runMigrate(t, s2)
	clean, ok := findMigrationEntry(report2, "adr-clean")
	if !ok {
		t.Fatalf("adr-clean missing from report")
	}
	if clean.Lossy {
		t.Errorf("proposed->open is a clean fold; must not be flagged lossy")
	}
}

// TestMigrate_GrandfathersTerminalAdrNoRetroCheck — old terminal ADRs are
// grandfathered to `done` with NO retro success-check: the item-6b auto-done latch
// (After-cite resolution) is NOT run against them (they have no After block). An
// `implemented` ADR with no resolvable After cites still grandfathers to `done`.
func TestMigrate_GrandfathersTerminalAdrNoRetroCheck(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-impl", "implemented")

	report, _ := runMigrate(t, s)

	e, err := s.GetEntity("adr-impl")
	if err != nil {
		t.Fatalf("get adr-impl: %v", err)
	}
	if e.Status != "done" {
		t.Errorf("implemented ADR grandfathered status = %q, want %q", e.Status, "done")
	}
	entry, ok := findMigrationEntry(report, "adr-impl")
	if !ok {
		t.Fatalf("adr-impl missing from report")
	}
	// Grandfathered, not auto-done: the report must record it as a mapping, not an
	// auto-done latch flip (no retro After-check was run).
	if entry.AutoDone {
		t.Errorf("grandfathered terminal ADR must NOT be marked auto-done (no retro check), got AutoDone=true")
	}
}

// TestMigrate_PlainWriteStillBlockedOnProvisioned — a plain RunSet attempting the
// same provisioned rewrite is still a no-op on status per item 2: only the
// privileged migration path may rewrite a terminal status.
func TestMigrate_PlainWriteStillBlockedOnProvisioned(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-prov", "provisioned")

	var buf bytes.Buffer
	// provisioned is terminal in the legal-jump table; a plain set to done is illegal.
	err := RunSet(SetOptions{Store: s, ID: "adr-prov", Field: "status", Value: "done"}, &buf)
	if err == nil {
		t.Fatalf("plain RunSet provisioned->done must be rejected (only migration may rewrite a terminal)")
	}
	e, _ := s.GetEntity("adr-prov")
	if e.Status != "provisioned" {
		t.Errorf("plain RunSet must leave provisioned unchanged, got %q", e.Status)
	}
}

// seedOldGrammarCanvas writes an on-disk canvas at .c3/canvases/<id>.md carrying
// the embedded default's definition BODY (so it is UNcustomized) but in old
// grammar: no `status:` frontmatter and a stale, non-verifying seal. Returns the
// file path.
func seedOldGrammarCanvas(t *testing.T, c3Dir, id string) string {
	t.Helper()
	canvasDir := filepath.Join(c3Dir, "canvases")
	if err := os.MkdirAll(canvasDir, 0755); err != nil {
		t.Fatalf("mkdir canvases: %v", err)
	}
	embedded, ok := schema.CanvasFor(id)
	if !ok {
		t.Fatalf("no embedded canvas %q", id)
	}
	// Old grammar: same definition body via the canonical renderer (valid
	// frontmatter), but with the `status:` set stripped and a stale, non-verifying
	// seal — simulating a pre-status-grammar materialization the user never edited.
	rendered := renderCanvasDoc(embedded, true)
	oldGrammar := stripCanvasStatusLine(rendered)
	oldGrammar = staleSealCanvas(oldGrammar)
	if oldGrammar == rendered {
		t.Fatalf("precondition: expected to strip the status set / re-seal for %q", id)
	}
	path := filepath.Join(canvasDir, id+".md")
	if err := os.WriteFile(path, []byte(oldGrammar), 0644); err != nil {
		t.Fatalf("write old-grammar canvas: %v", err)
	}
	return path
}

// stripCanvasStatusLine removes the `status: [...]` frontmatter line from a
// rendered canvas doc, simulating the pre-status (old) grammar.
func stripCanvasStatusLine(doc string) string {
	var out []string
	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "status:") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// staleSealCanvas replaces the rendered canvas's c3-seal with a stale,
// non-verifying value.
func staleSealCanvas(doc string) string {
	return canvasSealRE.ReplaceAllString(doc, "c3-seal: deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
}

var canvasSealRE = regexp.MustCompile(`c3-seal: [a-f0-9]+`)

// TestMigrate_ReconcilesSealedCanvasExplicitly — a project that already
// materialized an UNCUSTOMIZED old-grammar canvas (its definition still matches
// the embedded default) is reconciled explicitly (re-sealed/updated, gaining the
// new `status` frontmatter + FREE/STRICT markers); the write-if-absent freeze is
// NOT silently bypassed: reconciliation is an explicit, announced rewrite, not a
// default-materialize no-op.
func TestMigrate_ReconcilesSealedCanvasExplicitly(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	path := seedOldGrammarCanvas(t, c3Dir, "adr")

	s := createDBFixture(t)
	var buf bytes.Buffer
	if _, err := RunMigrate(MigrateOptions{Store: s, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatalf("RunMigrate (reconcile): %v\n%s", err, buf.String())
	}

	reconciled, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read reconciled canvas: %v", err)
	}
	got := string(reconciled)
	// The reconciled canvas must now carry the declared status set (new grammar).
	if !strings.Contains(got, "status:") {
		t.Errorf("reconciled adr canvas must gain a declared status set, got:\n%s", got)
	}
	// And it must be a fresh, valid seal — not the stale deadbeef one.
	if strings.Contains(got, "deadbeefdeadbeef") {
		t.Errorf("reconciled canvas must be re-sealed, still carries the stale seal:\n%s", got)
	}
	// And the reconciliation must be announced (not silent).
	if !strings.Contains(buf.String(), "adr") || !strings.Contains(strings.ToLower(buf.String()), "reconcil") {
		t.Errorf("reconciliation must be announced loudly, got:\n%s", buf.String())
	}
}

// TestMigrate_DoesNotClobberCustomizedCanvas — a user-CUSTOMIZED embedded-id
// canvas (its definition diverges from the embedded default — e.g. a hand-edited
// component.md) is NEVER overwritten by the reconcile sweep. Its content is left
// byte-for-byte intact and the sweep emits a LOUD itemized line telling the user
// to reconcile it by hand. Non-silent, no data loss.
func TestMigrate_DoesNotClobberCustomizedCanvas(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	canvasDir := filepath.Join(c3Dir, "canvases")
	if err := os.MkdirAll(canvasDir, 0755); err != nil {
		t.Fatalf("mkdir canvases: %v", err)
	}

	// A CUSTOMIZED component canvas: the user changed the domain, so its definition
	// diverges from the embedded default.
	embedded, ok := schema.CanvasFor("component")
	if !ok {
		t.Fatalf("no embedded component canvas")
	}
	rendered := renderCanvasDoc(embedded, true)
	customized := strings.Replace(rendered, "domain: software", "domain: my-custom-domain", 1)
	if customized == rendered {
		t.Fatalf("precondition: expected to alter the embedded body's domain")
	}
	path := filepath.Join(canvasDir, "component.md")
	if err := os.WriteFile(path, []byte(customized), 0644); err != nil {
		t.Fatalf("write customized canvas: %v", err)
	}

	s := createDBFixture(t)
	var buf bytes.Buffer
	report, err := RunMigrate(MigrateOptions{Store: s, C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatalf("RunMigrate (customized): %v\n%s", err, buf.String())
	}

	// The customized canvas must be left byte-for-byte intact — NOT clobbered.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read customized canvas: %v", err)
	}
	if string(after) != customized {
		t.Errorf("customized canvas must NOT be overwritten; content changed:\n%s", string(after))
	}
	// It must NOT appear in the reconciled list.
	for _, id := range report.Reconciled {
		if id == "component" {
			t.Errorf("customized canvas component must not be auto-reconciled")
		}
	}
	// The sweep must report the manual-reconcile requirement loudly.
	out := buf.String()
	if !strings.Contains(out, "component") ||
		!(strings.Contains(out, "MANUAL") || strings.Contains(strings.ToLower(out), "customized")) {
		t.Errorf("sweep must loudly flag the customized canvas for manual reconcile, got:\n%s", out)
	}
}

// TestMigrate_ReSealsAllAfterSweep — every entity is re-sealed after the sweep;
// post-migration seal verification is clean. After migration, exporting the store
// and verifying every doc's on-disk seal recomputes intact.
func TestMigrate_ReSealsAllAfterSweep(t *testing.T) {
	s := createRichDBFixture(t)
	seedMigrationADR(t, s, "adr-impl", "implemented")
	runMigrate(t, s)

	c3Dir := exportThenStore(t, s)
	issues := checkIssues(t, c3Dir)
	if hasSealWarn(issues, "") {
		t.Fatalf("post-migration export must carry intact seals, got seal warns: %+v", issues)
	}

	// Every entity must also carry a non-empty RootMerkle (re-sealed in store).
	all, _ := s.AllEntities()
	for _, e := range all {
		if e.RootMerkle == "" {
			t.Errorf("entity %s has empty RootMerkle after re-seal sweep", e.ID)
		}
	}
}

// TestMigrate_AmbiguousStatusFailsLoud — a change doc whose status cannot be
// mapped FAILS loudly rather than silently coercing. The error names the offending
// entity and its unmappable status, and the store is left untouched on that entity.
func TestMigrate_AmbiguousStatusFailsLoud(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-weird", "frobnicated")

	var buf bytes.Buffer
	_, err := RunMigrate(MigrateOptions{Store: s}, &buf)
	if err == nil {
		t.Fatalf("migration must FAIL loud on an unmappable status, got nil error")
	}
	if !strings.Contains(err.Error(), "adr-weird") {
		t.Errorf("failure must name the offending entity adr-weird, got: %v", err)
	}
	if !strings.Contains(err.Error(), "frobnicated") {
		t.Errorf("failure must name the unmappable status, got: %v", err)
	}

	// Loud, not silent: the offending entity's status must be left as-is, not coerced.
	e, _ := s.GetEntity("adr-weird")
	if e.Status != "frobnicated" {
		t.Errorf("unmappable status must NOT be silently coerced, got %q", e.Status)
	}
}

// TestMigrate_PostMigrationCheckPasses — after migration, c3x check passes items
// 1–7 rules with no historical doc trapped in an unreachable-failing state. A
// store with a legacy terminal ADR, after migration, exports and checks clean
// (the terminal ADR is grandfathered + frozen, facts cleared, all re-sealed).
func TestMigrate_PostMigrationCheckPasses(t *testing.T) {
	s := createRichDBFixture(t)
	seedMigrationADR(t, s, "adr-impl", "implemented")
	runMigrate(t, s)

	c3Dir := exportThenStore(t, s)
	issues := checkIssues(t, c3Dir)
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("post-migration check produced a FAIL: %+v", iss)
		}
	}
}

// TestMigrate_DoesNotJudgeHistoricalSemantics (NEGATIVE / the LINE) — migration
// applies the declared mapping rule only; it does NOT decide the semantic meaning
// of a historical `active` fact, and never silently coerces. The report/output
// carries no semantic judgment about whether a historical status was "right".
func TestMigrate_DoesNotJudgeHistoricalSemantics(t *testing.T) {
	s := createDBFixture(t)
	seedMigrationADR(t, s, "adr-impl", "implemented")
	_, out := runMigrate(t, s)

	lower := strings.ToLower(out)
	for _, verdict := range []string{"should have been", "was wrong", "incorrect status", "really done", "not really", "bad decision"} {
		if strings.Contains(lower, verdict) {
			t.Errorf("migration must not judge historical semantics, found verdict %q in:\n%s", verdict, out)
		}
	}
}
