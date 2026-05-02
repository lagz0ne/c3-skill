package main

import (
	"bytes"
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
	t.Setenv("C3X_COORDINATOR_IDLE_MS", "75")

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	runSet := func(value string) {
		defer wg.Done()
		var buf bytes.Buffer
		errs <- run([]string{"--c3-dir", c3Dir, "set", "c3-101", "goal", value}, &buf)
	}

	wg.Add(2)
	go runSet("Handle API authentication requests plus sessions.")
	go runSet("Handle API authentication requests plus tokens.")
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent set failed: %v", err)
		}
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	entity, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(entity.Goal, "sessions") && !strings.Contains(entity.Goal, "tokens") {
		t.Fatalf("goal was not updated by either concurrent mutation: %q", entity.Goal)
	}
}

func TestRun_CoordinatorForwardsPipedInput(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	if err := coord.Cleanup(c3Dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("C3X_COORDINATOR_IDLE_MS", "10")

	body := richComponentBody("auth", "Handle auth with forwarded stdin.")
	var buf bytes.Buffer
	err := runWithIO(
		[]string{"--c3-dir", c3Dir, "write", "c3-101"},
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
	entity, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if entity.Goal != "Handle auth with forwarded stdin." {
		t.Fatalf("stdin write did not update goal: %q", entity.Goal)
	}
}

func TestRun_WriteFromFile(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	body := richComponentBody("auth", "Handle auth via --file.")
	bodyPath := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(bodyPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runWithIO(
		[]string{"--c3-dir", c3Dir, "write", "c3-101", "--file", bodyPath},
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
	entity, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if entity.Goal != "Handle auth via --file." {
		t.Fatalf("goal = %q, want %q", entity.Goal, "Handle auth via --file.")
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

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "goal", "Updated despite drift"}, &buf)
	if err != nil {
		t.Fatalf("mutation must bypass preverify, got: %v", err)
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

func TestRun_MarketplaceHelp(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"marketplace"}, &buf)
	if err != nil {
		t.Fatal(err)
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

func TestRun_Set(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "goal", "Updated goal"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(c3Dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "goal: Updated goal") {
		t.Fatalf("expected canonical export to include updated goal, got:\n%s", string(data))
	}
}

func TestRun_SetRejectsSection(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "--section", "Goal", "New goal"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected set --section to be rejected")
	}
	if !strings.Contains(err.Error(), "no longer accepts --section") {
		t.Fatalf("expected section-rejection error, got: %v", err)
	}
}

func TestRun_Wire(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_WireRemoveFlag(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	// Wire first
	run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "ref-jwt"}, &buf)
	buf.Reset()
	err := run([]string{"--c3-dir", c3Dir, "wire", "--remove", "c3-101", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_WireThreeArgs(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "cite", "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Delete(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "delete", "ref-jwt", "--dry-run"}, &buf)
	if err != nil {
		t.Fatal(err)
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

func TestRun_MarketplaceList(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"marketplace", "list"}, &buf)
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
