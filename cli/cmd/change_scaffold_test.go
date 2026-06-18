package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// TestRunChangeScaffold_StagesMissingSections covers the climb's scaffolder: a fact
// that sits below its canvas bar (here a ref with only a Goal, while the ref canvas
// requires Goal/Choice/Why) gets an insert-patch staging EMPTY templates for the
// missing required sections, ready to fill.
func TestRunChangeScaffold_StagesMissingSections(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	if err := s.InsertEntity(&store.Entity{ID: "ref-jwt", Type: "ref", Title: "jwt", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	// Write a sub-canvas body directly (content.WriteEntity doesn't gate), so the fact
	// is genuinely below the bar — the laddering case where the canvas rose after the
	// fact was authored.
	if err := content.WriteEntity(s, "ref-jwt", "# jwt\n\n## Goal\n\nStandardize token verification.\n"); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	if err := RunChangeScaffold(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &buf); err != nil {
		t.Fatalf("scaffold: %v\n%s", err, buf.String())
	}

	patch, err := os.ReadFile(filepath.Join(c3Dir, "changes", "climb-1", "ref-jwt-climb.patch.md"))
	if err != nil {
		t.Fatalf("expected a scaffolded climb patch for ref-jwt: %v\noutput: %s", err, buf.String())
	}
	requireAll(t, string(patch), "target: ref-jwt", "scope: insert", "base: ref-jwt@v", "## Choice", "## Why")
}

// TestClimb_ScaffoldGatedThenApplies is the end-to-end climb: scaffold stages empty
// templates, apply REFUSES them (won't apply on error), and once filled the same unit
// lands and the fact reaches the new rung.
func TestClimb_ScaffoldGatedThenApplies(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	if err := s.InsertEntity(&store.Entity{ID: "ref-jwt", Type: "ref", Title: "jwt", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "ref-jwt", "# jwt\n\n## Goal\n\nStandardize token verification.\n"); err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	if err := RunChangeScaffold(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &sb); err != nil {
		t.Fatal(err)
	}

	// Empty templates must NOT apply.
	var ab strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &ab); err == nil {
		t.Fatalf("apply must reject empty scaffolded sections; output: %s", ab.String())
	}

	// Fill the templates (keep the scaffolded frontmatter: target/scope/base).
	patchPath := filepath.Join(c3Dir, "changes", "climb-1", "ref-jwt-climb.patch.md")
	raw, err := os.ReadFile(patchPath)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(raw), "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("unexpected scaffolded patch shape:\n%s", raw)
	}
	filled := "---\n" + parts[1] + "---\n## Choice\n\nUse RS256 signed JWTs verified against a shared public key.\n\n## Why\n\nAsymmetric verification keeps the signing secret off every verifier.\n"
	if err := os.WriteFile(patchPath, []byte(filled), 0o644); err != nil {
		t.Fatal(err)
	}

	// Now it lands.
	var ab2 strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &ab2); err != nil {
		t.Fatalf("filled climb must apply: %v\n%s", err, ab2.String())
	}
	body, _ := content.ReadEntity(s, "ref-jwt")
	requireAll(t, body, "## Choice", "## Why", "Use RS256", "Asymmetric verification")
}

// A fact already at its bar produces no climb patch.
func TestRunChangeScaffold_NothingWhenComplete(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize token verification.", "Use RS256 JWTs.", "Asymmetric rationale.")

	var buf strings.Builder
	if err := RunChangeScaffold(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "nothing to climb") {
		t.Errorf("a fact at its bar should produce no climb patch, got: %s", buf.String())
	}
}

// TestRunChangeScaffold_RerunDoesNotOverwriteFilled — re-running scaffold must not
// clobber an already-staged (possibly filled) climb patch; it skips it.
func TestRunChangeScaffold_RerunDoesNotOverwriteFilled(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	if err := s.InsertEntity(&store.Entity{ID: "ref-jwt", Type: "ref", Title: "jwt", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "ref-jwt", "# jwt\n\n## Goal\n\nStandardize token verification.\n"); err != nil {
		t.Fatal(err)
	}
	var b1 strings.Builder
	if err := RunChangeScaffold(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &b1); err != nil {
		t.Fatal(err)
	}
	patchPath := filepath.Join(c3Dir, "changes", "climb-1", "ref-jwt-climb.patch.md")
	if err := os.WriteFile(patchPath, []byte("FILLED BY HUMAN"), 0o644); err != nil {
		t.Fatal(err)
	}

	var b2 strings.Builder
	if err := RunChangeScaffold(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "climb-1"}, &b2); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b2.String(), "skipped") {
		t.Errorf("re-scaffold should report skipping the existing patch, got: %s", b2.String())
	}
	got, _ := os.ReadFile(patchPath)
	if string(got) != "FILLED BY HUMAN" {
		t.Fatalf("re-scaffold overwrote a filled climb patch: %q", got)
	}
}
