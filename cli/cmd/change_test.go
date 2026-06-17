package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func openStoreC3(t *testing.T) (*store.Store, string) {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, t.TempDir()
}

func seedRef(t *testing.T, s *store.Store, id, goal, choice, why string) {
	t.Helper()
	if err := s.InsertEntity(&store.Entity{ID: id, Type: "ref", Title: id, Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("# %s\n\n## Goal\n\n%s\n\n## Choice\n\n%s\n\n## Why\n\n%s\n", id, goal, choice, why)
	if err := content.WriteEntity(s, id, body); err != nil {
		t.Fatal(err)
	}
}

func citeFor(t *testing.T, s *store.Store, id, snippet string) string {
	t.Helper()
	e, _ := s.GetEntity(id)
	nodes, _ := s.NodesForEntity(id)
	for _, n := range nodes {
		if n.Type == "paragraph" && strings.Contains(n.Content, snippet) {
			return fmt.Sprintf("%s#n%d@v%d:sha256:%s", e.ID, n.ID, e.Version, n.Hash)
		}
	}
	t.Fatalf("no paragraph %q in %s", snippet, id)
	return ""
}

func paraHash(t *testing.T, s *store.Store, id, snippet string) string {
	t.Helper()
	nodes, _ := s.NodesForEntity(id)
	for _, n := range nodes {
		if strings.Contains(n.Content, snippet) {
			return n.Hash
		}
	}
	return ""
}

func writePatch(t *testing.T, c3Dir, unitID, name, body string) {
	t.Helper()
	dir := filepath.Join(c3Dir, "changes", unitID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// S1 — edit a block; the target block changes, siblings stay frozen.
func TestRunChangeApply_S1_BlockEdit(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize token verification so components do not reinvent it.", "Use RS256 signed JWTs verified against a shared public key.", "Old rationale about asymmetric signing here.")
	base := citeFor(t, s, "ref-jwt", "Old rationale")
	beforeChoice := paraHash(t, s, "ref-jwt", "RS256 signed JWTs")

	writePatch(t, c3Dir, "adr-1", "01-why.patch.md",
		"---\ntarget: ref-jwt\nscope: block\nbase: "+base+"\n---\nAsymmetric signing lets each service verify without holding the signing secret, unlike HS256.\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err != nil {
		t.Fatalf("apply: %v\noutput: %s", err, buf.String())
	}
	got, _ := content.ReadEntity(s, "ref-jwt")
	if !strings.Contains(got, "unlike HS256") {
		t.Errorf("target block not updated:\n%s", got)
	}
	if after := paraHash(t, s, "ref-jwt", "RS256 signed JWTs"); after != beforeChoice {
		t.Error("sibling Choice block must stay frozen")
	}
}

// Codemap carrier — a .codemap.md binds external code through change apply. This
// also covers the carrier-only unit (no .patch.md): it must apply, not silently do
// nothing.
func TestRunChangeApply_Codemap_Binds(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize token verification.", "Use RS256 JWTs.", "Asymmetric rationale.")

	writePatch(t, c3Dir, "adr-1", "01.codemap.md",
		"---\ntarget: ref-jwt\n---\ncli/internal/auth/**\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err != nil {
		t.Fatalf("apply: %v\noutput: %s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "bound 01.codemap.md") {
		t.Errorf("expected a bound-codemap report, got: %s", buf.String())
	}
	patterns, _ := s.CodeMapFor("ref-jwt")
	if len(patterns) != 1 || patterns[0] != "cli/internal/auth/**" {
		t.Errorf("codemap = %v, want [cli/internal/auth/**]", patterns)
	}
}

// A codemap carrier whose target neither exists nor is created by the unit is
// rejected at preflight with a clear message (not a raw FK error at apply).
func TestRunChangeApply_Codemap_RejectsMissingTarget(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	writePatch(t, c3Dir, "adr-1", "01.codemap.md", "---\ntarget: c3-nope\n---\nsrc/**\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err == nil {
		t.Fatalf("expected a missing-target reject; output: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "does not exist") {
		t.Errorf("expected a clear missing-target reject, got: %s", buf.String())
	}
}

// Two codemap carriers targeting the same entity would full-replace each other
// (last wins, silently) — the unit must reject so its external footprint is clear.
func TestRunChangeApply_Codemap_RejectsDuplicateTarget(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize tokens.", "RS256.", "Asymmetric.")
	writePatch(t, c3Dir, "adr-1", "01.codemap.md", "---\ntarget: ref-jwt\n---\ncli/a/**\n")
	writePatch(t, c3Dir, "adr-1", "02.codemap.md", "---\ntarget: ref-jwt\n---\ncli/b/**\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err == nil {
		t.Fatalf("two carriers for one target must be rejected; output: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "both target") {
		t.Errorf("expected a duplicate-target reject, got: %s", buf.String())
	}
}

// S3 — a drifted anchor rejects the whole unit; nothing changes.
func TestRunChangeApply_S3_DriftRejects(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize token verification across the platform.", "Use RS256 signed JWTs for all services.", "Asymmetric verification rationale goes here.")
	base := citeFor(t, s, "ref-jwt", "Asymmetric verification")
	stale := strings.Replace(base, base[strings.LastIndex(base, ":")+1:], strings.Repeat("0", 64), 1)

	writePatch(t, c3Dir, "adr-1", "01.patch.md",
		"---\ntarget: ref-jwt\nscope: block\nbase: "+stale+"\n---\nShould not land.\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err == nil {
		t.Fatalf("expected drift rejection; output: %s", buf.String())
	}
	got, _ := content.ReadEntity(s, "ref-jwt")
	if strings.Contains(got, "Should not land.") {
		t.Error("drifted patch must not modify the target")
	}
}

// S4 — one drifted patch blocks the whole set (atomic).
func TestRunChangeApply_S4_AtomicOneDrifted(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-a", "Goal A spanning a few words here.", "Choice A spanning a few words.", "Why A original rationale text.")
	seedRef(t, s, "ref-b", "Goal B spanning a few words here.", "Choice B spanning a few words.", "Why B original rationale text.")
	baseA := citeFor(t, s, "ref-a", "Why A original")
	baseB := citeFor(t, s, "ref-b", "Why B original")
	staleB := strings.Replace(baseB, baseB[strings.LastIndex(baseB, ":")+1:], strings.Repeat("0", 64), 1)

	writePatch(t, c3Dir, "adr-1", "01-a.patch.md",
		"---\ntarget: ref-a\nscope: block\nbase: "+baseA+"\n---\nNew rationale A that is fresh and valid.\n")
	writePatch(t, c3Dir, "adr-1", "02-b.patch.md",
		"---\ntarget: ref-b\nscope: block\nbase: "+staleB+"\n---\nNew rationale B that would land.\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err == nil {
		t.Fatalf("expected atomic rejection; output: %s", buf.String())
	}
	gotA, _ := content.ReadEntity(s, "ref-a")
	if strings.Contains(gotA, "fresh and valid") {
		t.Error("atomic: ref-a must be unchanged when sibling patch drifts")
	}
}

// S6 — the freeze guard refuses a direct mutation of a fact (apply-only).
func TestGuardFactMutation_RefusesFacts(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Goal text spanning several words here now.", "Choice text spanning several words now.", "Why rationale text here now today.")
	err := GuardFactMutation(s, c3Dir, "ref-jwt")
	if err == nil {
		t.Fatal("a direct mutation of a fact must be refused")
	}
	if !strings.Contains(err.Error(), "frozen") {
		t.Errorf("refusal should explain facts are frozen, got: %v", err)
	}
}

func TestFactIsFrozen_Classification(t *testing.T) {
	dir := t.TempDir()
	if !FactIsFrozen(dir, "ref") || !FactIsFrozen(dir, "component") {
		t.Error("facts (ref/component) must be frozen")
	}
	if FactIsFrozen(dir, "adr") {
		t.Error("a change-doc (adr) is the authoring surface, not frozen")
	}
	if FactIsFrozen(dir, "canvas") {
		t.Error("a canvas is the contract, not frozen")
	}
}

// S2 — create a brand-new fact on an empty project via a no-base create patch.
func TestRunChangeApply_S2_CreateNewFact(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	body := "# ref-new\n\n## Goal\n\nStandardize a brand new pattern across components here now.\n\n## Choice\n\nUse the chosen concrete approach for this new pattern.\n\n## Why\n\nRationale explaining why this choice beats the alternatives here.\n"
	writePatch(t, c3Dir, "adr-1", "01-create.patch.md",
		"---\ntarget: ref-new\nscope: whole\ntype: ref\ntitle: New Ref\n---\n"+body)

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err != nil {
		t.Fatalf("create apply failed on empty project: %v\n%s", err, buf.String())
	}
	e, err := s.GetEntity("ref-new")
	if err != nil {
		t.Fatalf("created fact missing: %v", err)
	}
	if e.RootMerkle == "" {
		t.Error("a created fact must be born sealed (root merkle set)")
	}
	got, _ := content.ReadEntity(s, "ref-new")
	if !strings.Contains(got, "Standardize a brand new pattern") {
		t.Errorf("created body wrong:\n%s", got)
	}
}

// S7 — composed rename + re-edge via a frontmatter patch, applied atomically.
func TestRunChangeApply_S7_ComposedRenameReEdge(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Goal text spanning several words here now today.", "Choice text spanning several words now today.", "Why rationale text here now today indeed.")
	seedRef(t, s, "ref-crypto", "Crypto goal spanning several words here today.", "Crypto choice spanning several words now.", "Crypto why rationale text here now today.")
	e, _ := s.GetEntity("ref-jwt")
	entityBase := fmt.Sprintf("ref-jwt@v%d:sha256:%s", e.Version, e.RootMerkle)

	writePatch(t, c3Dir, "adr-1", "01-meta.patch.md",
		"---\ntarget: ref-jwt\nscope: frontmatter\nbase: "+entityBase+"\ntitle: JWT (renamed)\nuses:\n  - ref-crypto\n---\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err != nil {
		t.Fatalf("frontmatter apply: %v\n%s", err, buf.String())
	}
	e2, _ := s.GetEntity("ref-jwt")
	if e2.Title != "JWT (renamed)" {
		t.Errorf("rename not applied, title = %q", e2.Title)
	}
	rels, _ := s.RelationshipsFrom("ref-jwt")
	found := false
	for _, r := range rels {
		if r.RelType == "uses" && r.ToID == "ref-crypto" {
			found = true
		}
	}
	if !found {
		t.Error("re-edge to ref-crypto not applied")
	}
}

// Seal-state status: pending before apply, applied after (derived from hashes).
func TestRunChangeStatus_PendingThenApplied(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Goal text spanning several words here now.", "Choice text spanning several words now.", "Old why rationale text here now.")
	base := citeFor(t, s, "ref-jwt", "Old why rationale")
	writePatch(t, c3Dir, "adr-1", "01.patch.md",
		"---\ntarget: ref-jwt\nscope: block\nbase: "+base+"\n---\nFresh rationale that lands cleanly for the why section here.\n")

	var before strings.Builder
	if err := RunChangeStatus(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &before); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(before.String(), "pending") {
		t.Errorf("expected pending before apply:\n%s", before.String())
	}

	var ab strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &ab); err != nil {
		t.Fatalf("apply: %v\n%s", err, ab.String())
	}

	var after strings.Builder
	if err := RunChangeStatus(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &after); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(after.String(), "applied") {
		t.Errorf("expected applied after apply:\n%s", after.String())
	}
}

func TestRunChangeStatus_Drifted(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Goal text spanning several words here now.", "Choice text spanning several words now.", "Why rationale text original here now.")
	base := citeFor(t, s, "ref-jwt", "Why rationale text original")
	stale := strings.Replace(base, base[strings.LastIndex(base, ":")+1:], strings.Repeat("0", 64), 1)
	writePatch(t, c3Dir, "adr-1", "01.patch.md", "---\ntarget: ref-jwt\nscope: block\nbase: "+stale+"\n---\nx body\n")

	var b strings.Builder
	if err := RunChangeStatus(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "drifted") {
		t.Errorf("expected drifted status:\n%s", b.String())
	}
}

// A declared result-hash that the applied content does not seal to is rejected
// (the landing must equal what was reviewed).
func TestRunChangeApply_LandingMismatchRejects(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Goal text spanning several words here now.", "Choice text spanning several words now.", "Why rationale text original here now.")
	base := citeFor(t, s, "ref-jwt", "Why rationale text original")
	writePatch(t, c3Dir, "adr-1", "01.patch.md",
		"---\ntarget: ref-jwt\nscope: block\nbase: "+base+"\nresult: sha256:"+strings.Repeat("0", 64)+"\n---\nNew content that will not match the declared result hash.\n")

	var b strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &b); err == nil {
		t.Fatalf("expected landing-mismatch rejection; output: %s", b.String())
	}
}

// S5 — a patch whose merged result violates the canvas is rejected (second gate).
func TestRunChangeApply_S5_CanvasGate(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize token verification across all the services.", "Use RS256 signed JWTs for verification.", "Original why rationale paragraph here.")
	base := citeFor(t, s, "ref-jwt", "Original why rationale")

	// Empty content on the required Why section → merged body has an empty
	// required section → canvas gate must reject.
	writePatch(t, c3Dir, "adr-1", "01.patch.md",
		"---\ntarget: ref-jwt\nscope: block\nbase: "+base+"\n---\n")

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err == nil {
		t.Fatalf("expected canvas-gate rejection; output: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "canvas") {
		t.Errorf("rejection should cite the canvas gate, got: %s", buf.String())
	}
	got, _ := content.ReadEntity(s, "ref-jwt")
	if !strings.Contains(got, "Original why rationale") {
		t.Error("canvas-rejected patch must leave the fact untouched")
	}
}
