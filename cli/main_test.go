package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/coord"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRun_Version(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"--version"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "dev") {
		t.Errorf("version output = %q", buf.String())
	}
}

func TestRun_Help(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"--help"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "c3x") {
		t.Error("help should mention c3x")
	}
}

func TestRun_EmptyArgs_NoC3Dir(t *testing.T) {
	// When --c3-dir points to nonexistent dir, empty args shows help
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", filepath.Join(t.TempDir(), "no-c3")}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Commands:") {
		t.Errorf("empty args without .c3/ should show help, got:\n%s", buf.String())
	}
}

func TestRun_Capabilities(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"capabilities"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Command") {
		t.Error("capabilities should show command table")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	err := run([]string{"--c3-dir", t.TempDir(), "nonexistent"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestRun_NoC3Dir(t *testing.T) {
	err := run([]string{"--c3-dir", filepath.Join(t.TempDir(), "nope"), "list"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when no .c3/ found")
	}
}

func TestRun_ListWithDB(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "list", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "c3-0") {
		t.Error("list should include c3-0")
	}
}

func TestRun_ConcurrentMutationsUseShortLivedCoordinator(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	if err := coord.Cleanup(c3Dir); err != nil {
		t.Fatal(err)
	}
	requireCoordinatorAvailable(t, c3Dir)
	t.Setenv("C3X_COORDINATOR_IDLE_MS", "75")

	// Export canonical first so both serialized mutations refresh from a
	// consistent base (each mutating command rebuilds its cache from canonical).
	s0, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s0, OutputDir: c3Dir}, io.Discard); err != nil {
		s0.Close()
		t.Fatal(err)
	}
	s0.Close()

	// Two independent change-units each create a fact (non-frozen, coordinated).
	mkCreate := func(unit, ref string) {
		dir := filepath.Join(c3Dir, "changes", unit)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "# " + ref + "\n\n## Goal\n\nConcurrent create goal spanning several words now.\n\n## Choice\n\nThe concrete approach chosen for this pattern here.\n\n## Why\n\nRationale why this choice beats the alternatives here.\n"
		if err := os.WriteFile(filepath.Join(dir, "01.patch.md"), []byte("---\ntarget: "+ref+"\nscope: whole\ntype: ref\n---\n"+body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkCreate("adr-ca", "ref-conc-a")
	mkCreate("adr-cb", "ref-conc-b")

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	apply := func(unit string) {
		defer wg.Done()
		var buf bytes.Buffer
		errs <- run([]string{"--c3-dir", c3Dir, "change", "apply", unit}, &buf)
	}

	wg.Add(2)
	go apply("adr-ca")
	go apply("adr-cb")
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent change apply failed: %v", err)
		}
	}

	// Both commands completed via the coordinator without error (checked above);
	// the system is consistent and at least one create landed.
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_, aErr := s.GetEntity("ref-conc-a")
	_, bErr := s.GetEntity("ref-conc-b")
	if aErr != nil && bErr != nil {
		t.Fatalf("neither concurrent create landed: a=%v b=%v", aErr, bErr)
	}
}

func TestRun_CoordinatorForwardsPipedInput(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	if err := coord.Cleanup(c3Dir); err != nil {
		t.Fatal(err)
	}
	requireCoordinatorAvailable(t, c3Dir)
	t.Setenv("C3X_COORDINATOR_IDLE_MS", "10")

	// add reads its body from stdin and is not a frozen-fact mutation, so it is a
	// clean vehicle for exercising coordinator stdin forwarding.
	body := "## Goal\n\nForwarded stdin ref goal spanning several words here now.\n\n## Choice\n\nThe concrete chosen approach for this pattern.\n\n## Why\n\nRationale why this choice beats the alternatives here.\n"
	var buf bytes.Buffer
	err := runWithIO(
		[]string{"--c3-dir", c3Dir, "add", "ref", "piped-x"},
		strings.NewReader(body),
		false,
		&buf,
		io.Discard,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	entity, err := s.GetEntity("ref-piped-x")
	if err != nil {
		t.Fatalf("piped add did not create the ref: %v", err)
	}
	if !strings.Contains(entity.Goal, "Forwarded stdin ref goal") {
		t.Fatalf("stdin was not forwarded to add: goal = %q", entity.Goal)
	}
}

func requireCoordinatorAvailable(t *testing.T, c3Dir string) {
	t.Helper()
	leader, err := coord.NewLeader(c3Dir)
	if errors.Is(err, coord.ErrUnavailable) {
		t.Skipf("unix socket coordinator unavailable in this environment: %v", err)
	}
	if err != nil {
		t.Fatalf("probe coordinator: %v", err)
	}
	if err := leader.Close(); err != nil {
		t.Fatalf("close coordinator probe: %v", err)
	}
	if err := coord.Cleanup(c3Dir); err != nil {
		t.Fatalf("cleanup coordinator probe: %v", err)
	}
}

func TestRun_WriteFromFile(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	// --file is exercised via add (not frozen) since write on a fact is refused.
	body := "## Goal\n\nFiled ref goal spanning several words here now today.\n\n## Choice\n\nThe concrete approach chosen for this pattern here.\n\n## Why\n\nRationale why this choice beats the alternatives here.\n"
	bodyPath := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(bodyPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runWithIO(
		[]string{"--c3-dir", c3Dir, "add", "ref", "filed-x", "--file", bodyPath},
		strings.NewReader(""),
		true,
		&buf,
		io.Discard,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	entity, err := s.GetEntity("ref-filed-x")
	if err != nil {
		t.Fatalf("--file add did not create the ref: %v", err)
	}
	if !strings.Contains(entity.Goal, "Filed ref goal") {
		t.Fatalf("--file body not applied: goal = %q", entity.Goal)
	}
}

func TestRun_AddFromFile(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	body := "## Goal\nShared JWT ref.\n\n## Choice\nHS256.\n\n## Why\nSimple.\n"
	bodyPath := filepath.Join(t.TempDir(), "ref.md")
	if err := os.WriteFile(bodyPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runWithIO(
		[]string{"--c3-dir", c3Dir, "add", "ref", "jwt-hs256", "--file", bodyPath},
		strings.NewReader(""),
		true,
		&buf,
		io.Discard,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.GetEntity("ref-jwt-hs256"); err != nil {
		t.Fatalf("ref-jwt-hs256 not created: %v", err)
	}
}

func TestRun_CheckWithDB(t *testing.T) {
	c3Dir := setupC3DB(t)
	// Seed canonical README so pre-heal succeeds.
	seedCanonicalReadme(t, c3Dir)
	var buf bytes.Buffer
	_ = run([]string{"--c3-dir", c3Dir, "check", "--json"}, &buf)
}

// check must surface the failing entity + message, not just "1 error(s)".
func TestRun_CheckSurfacesValidatorIssues(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)

	// Break c3-101 by stripping its strict body so check returns errors with details.
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nThin.\n"); err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Close()

	var buf bytes.Buffer
	err = run([]string{"--c3-dir", c3Dir, "check"}, &buf)
	if err == nil {
		t.Fatal("expected check to fail on broken c3-101 body")
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("check output must name the failing entity, got:\n%s", out)
	}
	if !strings.Contains(out, "missing required section") && !strings.Contains(out, "empty required section") {
		t.Errorf("check output must describe what failed, got:\n%s", out)
	}
}

// check --json must include the issue list with entity + message.
func TestRun_CheckJSONIncludesIssues(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nThin.\n"); err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Close()

	var buf bytes.Buffer
	err = run([]string{"--c3-dir", c3Dir, "check", "--json"}, &buf)
	if err == nil {
		t.Fatal("expected check --json to fail")
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("check --json must include entity id, got:\n%s", out)
	}
	if !strings.Contains(out, "\"issues\"") && !strings.Contains(out, "issues:") {
		t.Errorf("check --json must include issues list, got:\n%s", out)
	}
}

// Mutations must not be gated by canonical preverify failures, per ADR
// mutation-preverify-repair-bypass — the mutation itself may be the fix.
func TestRun_MutationBypassesPreverify(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)

	c101Path := filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")
	broken := "---\nid: c3-101\nc3-seal: deadbeef\ntitle: auth\ntype: component\ncategory: foundation\nparent: c3-1\n---\n\n# auth\n\n## Goal\n\nThin.\n"
	if err := os.WriteFile(c101Path, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}

	// A mutation (change apply, creating a fact) must still run despite the broken
	// canonical c3-101 — mutations bypass the preverify-repair step.
	dir := filepath.Join(c3Dir, "changes", "adr-preverify")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# ref-pv\n\n## Goal\n\nPreverify-bypass ref goal spanning several words now.\n\n## Choice\n\nThe concrete approach chosen here for the pattern.\n\n## Why\n\nRationale why this choice beats the alternatives here.\n"
	if err := os.WriteFile(filepath.Join(dir, "01.patch.md"), []byte("---\ntarget: ref-pv\nscope: whole\ntype: ref\n---\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "change", "apply", "adr-preverify"}, &buf)
	if err != nil {
		t.Fatalf("mutation must bypass preverify, got: %v\n%s", err, buf.String())
	}
}

func TestBDD_ReadOnlyLookupSkipsCanonicalPreverifyWhenCacheExists(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	entity, err := s.GetEntity("c3-101")
	if err != nil {
		s.Close()
		t.Fatal(err)
	}
	entity.Title = "api-latency-gateway"
	entity.Goal = "Own API request routing and latency instrumentation."
	if err := s.UpdateEntity(entity); err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := s.SetCodeMap("c3-101", []string{"src/api/handlers/latency.go"}); err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Close()

	c101Path := filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")
	broken := "---\nid: c3-101\nc3-seal: deadbeef\ntitle: auth\ntype: component\ncategory: foundation\nparent: c3-1\n---\n\n# auth\n\n## Goal\n\nThin.\n"
	if err := os.WriteFile(c101Path, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = runWithIO(
		[]string{"--c3-dir", c3Dir, "lookup", "src/api/handlers/latency.go"},
		strings.NewReader(""),
		true,
		&stdout,
		&stderr,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "c3-101") || !strings.Contains(stdout.String(), "api-latency-gateway") {
		t.Fatalf("lookup should use cache-backed result, stdout:\n%s", stdout.String())
	}
	if strings.Contains(stderr.String(), ".c3/ drift detected") {
		t.Fatalf("read-only lookup should not preverify canonical drift, stderr:\n%s", stderr.String())
	}
}

func TestBDD_RunSearchHybridJSONDispatch(t *testing.T) {
	t.Setenv("C3_SEMANTIC_CACHE_DIR", t.TempDir())
	t.Setenv("C3_SEMANTIC_OFFLINE", "1")
	c3Dir := setupRichC3DB(t)
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	seedMainHybridSearchFixture(t, s)
	s.Close()

	var stdout, stderr bytes.Buffer
	err = runWithIO(
		[]string{"--c3-dir", c3Dir, "search", "pool wait p95 latency", "--json", "--limit", "3"},
		strings.NewReader(""),
		true,
		&stdout,
		&stderr,
		false,
	)
	if err != nil {
		t.Fatalf("search dispatch failed: %v\nstderr:\n%s", err, stderr.String())
	}
	var out cmd.SearchOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, stdout.String())
	}
	if len(out.Results) == 0 || out.Results[0].ID != "research-note-20260605-api-latency" {
		t.Fatalf("unexpected search result: %+v", out.Results)
	}
	if out.Results[0].Context.Component.ID != "c3-101" || out.Results[0].Context.Path != "src/api/handlers/latency.go" {
		t.Fatalf("hybrid context missing: %+v", out.Results[0].Context)
	}
}

// c3x repair must be a real command, not "unknown command".
func TestRun_RepairCommandExists(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "repair"}, &buf)
	if err != nil && strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("c3x repair must be wired up, got: %v", err)
	}
}

// repair rewrites canonical files and the cache, so it must classify as mutating
// (gives it the rollback snapshot + coordinator gate). Otherwise a failed
// repair can leave .c3/ partially rewritten with no way back.
func TestCommandMutatesCanonical_RepairIsMutating(t *testing.T) {
	if !commandMutatesCanonical(cmd.Options{Command: "repair"}) {
		t.Fatal("repair must be classified as mutating: it rewrites canonical files")
	}
}

// supersede and migrate rewrite store status (and migrate rewrites canvases), so
// they must classify as mutating to get the rollback snapshot + coordinator gate.
func TestCommandMutatesCanonical_SupersedeAndMigrateAreMutating(t *testing.T) {
	for _, c := range []string{"supersede", "migrate"} {
		if !commandMutatesCanonical(cmd.Options{Command: c}) {
			t.Fatalf("%s must be classified as mutating", c)
		}
	}
}

// c3x supersede must be a wired command, not "unknown command", and must dispatch
// to RunSupersede.
func TestRun_SupersedeCommandDispatches(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	// Seed a terminal change doc (done) and a successor (open) into the store +
	// canonical, so the dispatch path can flip + backlink.
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []*store.Entity{
		{ID: "adr-old", Type: "adr", Title: "Old decision", Slug: "old", Status: "done", Date: "20260101", Metadata: "{}"},
		{ID: "adr-new", Type: "adr", Title: "New decision", Slug: "new", Status: "open", Date: "20260601", Metadata: "{}"},
	} {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed %s: %v", e.ID, err)
		}
		if err := content.WriteEntity(s, e.ID, "# "+e.Title+"\n\n## Context\n\nBody.\n"); err != nil {
			t.Fatalf("seed body %s: %v", e.ID, err)
		}
	}
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Close()

	var buf bytes.Buffer
	err = run([]string{"--c3-dir", c3Dir, "supersede", "adr-new", "adr-old"}, &buf)
	if err != nil && strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("c3x supersede must be wired up, got: %v", err)
	}
	if err != nil {
		t.Fatalf("supersede dispatch failed: %v\n%s", err, buf.String())
	}

	s2, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	old, err := s2.GetEntity("adr-old")
	if err != nil {
		t.Fatal(err)
	}
	if old.Status != "superseded" {
		t.Fatalf("supersede should flip adr-old to superseded, got %q", old.Status)
	}
}

// c3x supersede without two args reports a usage error, not "unknown command".
func TestRun_SupersedeMissingArgs(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)
	err := run([]string{"--c3-dir", c3Dir, "supersede", "adr-new"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected usage error for supersede with one arg")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("supersede must be wired, got: %v", err)
	}
}

// c3x migrate must be a wired command and dispatch to RunMigrate, sweeping the
// fact 'active' statuses to empty (loud report).
func TestRun_MigrateCommandDispatches(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "migrate"}, &buf)
	if err != nil && strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("c3x migrate must be wired up, got: %v", err)
	}
	if err != nil {
		t.Fatalf("migrate dispatch failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "MIGRATION") {
		t.Fatalf("migrate should emit a loud migration report, got:\n%s", buf.String())
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if e.Status != "" {
		t.Fatalf("migrate should clear fact c3-101 active status, got %q", e.Status)
	}
}

// c3x change apply must dispatch through run() and drive the full mutation path
// (snapshot → apply → canonical export). A no-base create patch creates a fact.
func TestRun_ChangeApplyDispatches(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Close()

	unitID := "adr-20260616-newref"
	dir := filepath.Join(c3Dir, "changes", unitID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# ref-newx\n\n## Goal\n\nStandardize a brand new pattern across components here now.\n\n## Choice\n\nUse the chosen concrete approach for this new pattern.\n\n## Why\n\nRationale explaining why this choice beats the realistic alternatives here.\n"
	if err := os.WriteFile(filepath.Join(dir, "01-create.patch.md"), []byte("---\ntarget: ref-newx\nscope: whole\ntype: ref\ntitle: New Ref\n---\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "change", "apply", unitID}, &buf); err != nil {
		t.Fatalf("change apply dispatch failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "applied") {
		t.Errorf("expected apply confirmation, got: %s", buf.String())
	}
	s2, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	if _, err := s2.GetEntity("ref-newx"); err != nil {
		t.Errorf("created fact ref-newx not found after dispatch: %v", err)
	}
}

func TestRun_ChangeUsageError(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	seedCanonicalReadme(t, c3Dir)
	err := run([]string{"--c3-dir", c3Dir, "change", "bogus", "x"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected usage error for unknown change subcommand")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("change must be wired, got: %v", err)
	}
}

func TestRun_Schema(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "schema", "component"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_GraphMissingID(t *testing.T) {
	c3Dir := setupC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "graph"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for graph without entity ID")
	}
}

func TestRun_LookupMissingArg(t *testing.T) {
	c3Dir := setupC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "lookup"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for lookup without file path")
	}
}

func TestRun_GitInstall(t *testing.T) {
	c3Dir := setupC3DB(t)
	projectDir := filepath.Dir(c3Dir)
	if err := os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "git", "install"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Installed C3 Git guardrails") {
		t.Fatalf("expected git install summary, got %q", buf.String())
	}
	if !fileExists(filepath.Join(projectDir, ".git", "hooks", "pre-commit")) {
		t.Fatal("expected pre-commit hook")
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "exists.txt")
	os.WriteFile(f, []byte("x"), 0644)

	if !fileExists(f) {
		t.Error("should return true for existing file")
	}
	if fileExists(filepath.Join(tmp, "nope.txt")) {
		t.Error("should return false for missing file")
	}
}

func TestRun_Add(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	body := "## Goal\nRate limiting strategy.\n\n## Choice\nToken bucket.\n\n## Why\nSimple and effective.\n"
	r, w, _ := os.Pipe()
	w.WriteString(body)
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	err := run([]string{"--c3-dir", c3Dir, "add", "ref", "rate-limiting"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(c3Dir, "README.md")) {
		t.Fatal("expected canonical export after add")
	}
}

func TestRun_AddJSON(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	body := "## Goal\nCaching strategy.\n\n## Choice\nRedis.\n\n## Why\nFast.\n"
	r, w, _ := os.Pipe()
	w.WriteString(body)
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	err := run([]string{"--c3-dir", c3Dir, "add", "ref", "caching", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ref-caching") {
		t.Errorf("JSON add output should contain ref-caching: %s", buf.String())
	}
}

func TestRun_AddAgentModeReturnsTOON(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	body := "## Goal\nAgent cache strategy.\n\n## Choice\nSQLite cache.\n\n## Why\nLocal and deterministic.\n"
	r, w, _ := os.Pipe()
	w.WriteString(body)
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	err := run([]string{"--c3-dir", c3Dir, "add", "ref", "agent-cache"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("agent add must not emit JSON:\n%s", out)
	}
	if !strings.Contains(out, "id: ref-agent-cache") {
		t.Fatalf("agent add should emit TOON id, got:\n%s", out)
	}
}

func TestRun_AgentModeExplicitJSONStillReturnsTOON(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	c3Dir := setupRichC3DB(t)

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetCodeMap("c3-101", []string{"src/auth/login.ts"}); err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Close()

	cases := []struct {
		name string
		argv []string
		want string
	}{
		{name: "list", argv: []string{"--c3-dir", c3Dir, "list", "--json"}, want: "totalCount:"},
		{name: "read", argv: []string{"--c3-dir", c3Dir, "read", "c3-101", "--json"}, want: "id: c3-101"},
		{name: "lookup", argv: []string{"--c3-dir", c3Dir, "lookup", "src/auth/login.ts", "--json"}, want: "file: src/auth/login.ts"},
		{name: "schema", argv: []string{"--c3-dir", c3Dir, "schema", "component", "--json"}, want: "type: component"},
		{name: "graph", argv: []string{"--c3-dir", c3Dir, "graph", "c3-101", "--json"}, want: "nodes["},
		{name: "search", argv: []string{"--c3-dir", c3Dir, "search", "auth", "--no-semantic", "--json"}, want: "query: auth"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := run(tc.argv, &buf); err != nil {
				t.Fatal(err)
			}
			out := strings.TrimSpace(buf.String())
			if strings.HasPrefix(out, "{") || strings.HasPrefix(out, "[") {
				t.Fatalf("agent %s --json must emit TOON, got JSON:\n%s", tc.name, out)
			}
			if !strings.Contains(out, tc.want) {
				t.Fatalf("agent %s --json missing %q in:\n%s", tc.name, tc.want, out)
			}
		})
	}
}

// Facts are frozen: set/delete on a fact are refused at the CLI; the change
// is a change-unit. (set --section is still meaningful on a non-frozen change-doc.)
func TestRun_Set(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "goal", "Updated goal"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("set on a fact must be refused (facts are frozen)")
	}
	if !strings.Contains(err.Error(), "frozen") {
		t.Fatalf("expected a frozen-fact refusal, got: %v", err)
	}
}

func TestRun_SetCodemapOnFrozenFact(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer

	err := runWithIO([]string{"--c3-dir", c3Dir, "set", "c3-101", "codemap", "src/auth/**"}, strings.NewReader(""), true, &buf, io.Discard, false)
	if err != nil {
		t.Fatalf("set codemap on a frozen fact should be allowed, got: %v", err)
	}
	if strings.Contains(buf.String(), "facts are frozen") {
		t.Fatalf("codemap set should not emit frozen-fact refusal:\n%s", buf.String())
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	patterns, err := s.CodeMapFor("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) != 1 || patterns[0] != "src/auth/**" {
		t.Fatalf("codemap patterns = %v, want [src/auth/**]", patterns)
	}
}

func TestRun_SetCodemapOnFrozenFactWithFieldFlag(t *testing.T) {
	c3Dir := setupRichC3DB(t)

	err := runWithIO([]string{"--c3-dir", c3Dir, "set", "c3-101", "src/ui/**", "--field", "codemap"}, strings.NewReader(""), true, &bytes.Buffer{}, io.Discard, false)
	if err != nil {
		t.Fatalf("set --field codemap on a frozen fact should be allowed, got: %v", err)
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	patterns, err := s.CodeMapFor("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) != 1 || patterns[0] != "src/ui/**" {
		t.Fatalf("codemap patterns = %v, want [src/ui/**]", patterns)
	}
}

func TestRun_SetRejectsSection(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	// adr is a change-doc (not frozen), so the dispatch guard passes and the
	// command's own --section rejection fires.
	err := run([]string{"--c3-dir", c3Dir, "set", "adr-20260226-use-go", "--section", "Goal", "New goal"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected set --section to be rejected")
	}
	if !strings.Contains(err.Error(), "no longer accepts --section") {
		t.Fatalf("expected section-rejection error, got: %v", err)
	}
}

func TestRun_Delete(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "delete", "ref-jwt", "--dry-run"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("delete of a fact must be refused (facts are frozen)")
	}
	if !strings.Contains(err.Error(), "frozen") {
		t.Fatalf("expected a frozen-fact refusal, got: %v", err)
	}
}

func TestRun_Graph(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "graph", "c3-0"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Codemap(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "codemap"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Lookup(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "lookup", "src/main.go"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_ListFlat(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "list", "--flat"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_ListCompact(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "list", "--compact"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

// seedCanonicalReadme writes a minimal sealed README.md to the c3Dir via the
// store's export helper so pre-heal verification passes.
func seedCanonicalReadme(t *testing.T, c3Dir string) {
	t.Helper()
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		t.Fatal(err)
	}
}

// setupC3DB creates a temp .c3/ dir with a SQLite DB containing a minimal fixture.
func setupC3DB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	dbPath := filepath.Join(c3Dir, "c3.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer s.Close()

	s.InsertEntity(&store.Entity{
		ID: "c3-0", Type: "system", Title: "TestProject",
		Slug: "", Status: "active", Metadata: "{}",
	})

	return c3Dir
}

// setupRichC3DB creates a .c3/ dir with DB containing containers, components, refs.
func setupRichC3DB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)
	// Create container dirs for add commands that write files
	os.MkdirAll(filepath.Join(c3Dir, "c3-1-api"), 0755)

	dbPath := filepath.Join(c3Dir, "c3.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer s.Close()

	entities := []*store.Entity{
		{ID: "c3-0", Type: "system", Title: "TestProject", Slug: "", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "api", Slug: "api", ParentID: "c3-0", Goal: "Serve API", Boundary: "service", Status: "active", Metadata: "{}"},
		{ID: "c3-101", Type: "component", Title: "auth", Slug: "auth", Category: "foundation", ParentID: "c3-1", Status: "active", Metadata: "{}"},
		{ID: "ref-jwt", Type: "ref", Title: "JWT", Slug: "jwt", Goal: "JWT tokens", Status: "active", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	s.AddRelationship(&store.Relationship{FromID: "ref-jwt", ToID: "c3-1", RelType: "scope"})

	// Populate node trees
	bodies := map[string]string{
		"c3-0":    "# TestProject\n\n## Goal\n\nTest.\n\n## Containers\n\n| ID | Name | Boundary | Goal |\n|----|------|----------|------|\n",
		"c3-1":    "# api\n\n## Goal\n\nServe API.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n",
		"c3-101":  richComponentBody("auth", "Handle API authentication requests."),
		"ref-jwt": "# JWT\n\n## Goal\n\nJWT tokens.\n",
	}
	for id, body := range bodies {
		if err := content.WriteEntity(s, id, body); err != nil {
			t.Fatalf("seed nodes %s: %v", id, err)
		}
	}

	return c3Dir
}

func seedMainHybridSearchFixture(t *testing.T, s *store.Store) {
	t.Helper()
	for _, e := range []*store.Entity{
		{ID: "rule-trace-context", Type: "rule", Title: "Trace Context Propagation", Slug: "trace-context", Goal: "Every outbound API call carries traceparent and request_id.", Status: "active", Metadata: "{}"},
		{ID: "ref-latency-budget", Type: "ref", Title: "Latency Budget", Slug: "latency-budget", Goal: "Keep API p95 under 250 ms before checkout release.", Status: "active", Metadata: "{}"},
		{ID: "research-note-20260605-api-latency", Type: "research-note", Title: "API Latency Investigation", Slug: "api-latency", Goal: "Investigate checkout API latency pool wait regression.", Status: "active", Metadata: "{}"},
	} {
		if err := s.InsertEntity(e); err != nil {
			t.Fatal(err)
		}
	}
	auth, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	auth.Title = "api-latency-gateway"
	auth.Goal = "Own API request routing and latency instrumentation."
	if err := s.UpdateEntity(auth); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []*store.Relationship{
		{FromID: "research-note-20260605-api-latency", ToID: "c3-101", RelType: "affects"},
		{FromID: "research-note-20260605-api-latency", ToID: "ref-latency-budget", RelType: "uses"},
		{FromID: "research-note-20260605-api-latency", ToID: "rule-trace-context", RelType: "uses"},
	} {
		if err := s.AddRelationship(rel); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetCodeMap("c3-101", []string{"src/api/handlers/latency.go"}); err != nil {
		t.Fatal(err)
	}
	body := "## Summary\n\nCheckout API p95 increased from 180 ms to 420 ms after the connection-pool change. Span evidence points to DB pool wait.\n"
	if err := content.WriteEntity(s, "research-note-20260605-api-latency", body); err != nil {
		t.Fatal(err)
	}
}

func richComponentBody(title, goal string) string {
	return "# " + title + "\n\n" +
		"## Goal\n\n" + goal + "\n\n" +
		"## Parent Fit\n\n" +
		"| Field | Value |\n|-------|-------|\n" +
		"| Parent | c3-1 |\n| Role | Owns authentication behavior for the API container. |\n| Boundary | Keeps identity decisions inside the API boundary. |\n| Collaboration | Coordinates with JWT reference rules and caller-facing API flows. |\n\n" +
		"## Purpose\n\nProvide agent-ready authentication documentation so generated code can preserve identity flow, boundaries, verification, and governing references.\n\n" +
		"## Foundational Flow\n\n" +
		"| Aspect | Detail | Reference |\n|--------|--------|-----------|\n" +
		"| Input | Accept caller credentials and token material. | ref-jwt |\n" +
		"| State | Keep auth state explicit at API boundaries. | ref-jwt |\n" +
		"| Output | Return verified identity context to downstream handlers. | ref-jwt |\n" +
		"| Failure | Reject invalid or expired token material. | ref-jwt |\n\n" +
		"## Business Flow\n\n" +
		"| Aspect | Detail | Reference |\n|--------|--------|-----------|\n" +
		"| Actor | API caller asks for authenticated access. | ref-jwt |\n" +
		"| Decision | Authentication gates protected behavior. | ref-jwt |\n" +
		"| Outcome | Authorized requests continue with identity context. | ref-jwt |\n" +
		"| Exception | Unauthorized requests receive a clear rejection. | ref-jwt |\n\n" +
		"## Governance\n\n" +
		"| Reference | Type | Governs | Precedence | Notes |\n|-----------|------|---------|------------|-------|\n" +
		"| ref-jwt | ref | Token format and validation expectations. | Required | Applied to all auth decisions. |\n\n" +
		"## Contract\n\n" +
		"| Surface | Direction | Contract | Boundary | Evidence |\n|---------|-----------|----------|----------|----------|\n" +
		"| Credentials | IN | Caller supplies credentials or token material. | API boundary | ref-jwt |\n" +
		"| Identity | OUT | Component returns verified identity context or rejection. | API boundary | go test ./... |\n\n" +
		"## Change Safety\n\n" +
		"| Risk | Trigger | Detection | Required Verification |\n|------|---------|-----------|-----------------------|\n" +
		"| Token drift | JWT reference changes. | Contract review catches mismatched fields. | go test ./... |\n" +
		"| Boundary leak | Identity state bypasses auth. | Code review traces caller paths. | c3x check --include-adr |\n\n" +
		"## Derived Materials\n\n" +
		"| Material | Must derive from | Allowed variance | Evidence |\n|----------|------------------|------------------|----------|\n" +
		"| Auth handlers | Goal and Contract sections. | Names may vary by framework. | go test ./cmd |\n"
}
