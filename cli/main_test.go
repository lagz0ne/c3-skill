package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/content"
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

func TestRun_EmptyArgs_WithC3Dir(t *testing.T) {
	// When .c3/ with DB exists, empty args shows status dashboard
	c3Dir := setupC3DB(t)

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "TestProject") {
		t.Errorf("empty args with .c3/ should show status dashboard, got: %s", buf.String())
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

func TestRun_CheckWithDB(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "check", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
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

func TestRun_ListRebuildsMissingDatabaseFromCanonicalFiles(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", c3Dir}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(c3Dir, "c3.db")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "list", "--json"}, &buf); err != nil {
		t.Fatalf("expected canonical files to rebuild cache, got %v", err)
	}
	if !fileExists(filepath.Join(c3Dir, "c3.db")) {
		t.Fatal("expected missing c3.db to be rebuilt from canonical files")
	}
	if !strings.Contains(buf.String(), "c3-101") {
		t.Fatalf("expected list output after rebuild, got %q", buf.String())
	}
}

func TestRun_VerifyRebuildSurfacesLayerDisconnect(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-1", "# api\n\n## Goal\n\nServe API requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| c3-101 | auth | foundation | active | Authentication |\n\n## Responsibilities\n\nServe API requests.\n"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if err := os.Remove(filepath.Join(c3Dir, "c3.db")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "verify"}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, want := range []string{
		"Rebuilt local C3 cache from canonical .c3/",
		"layer disconnect",
		"missing from c3-0 Containers table",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output:\n%s", want, output)
		}
	}
}

func TestRun_ImportWithoutDatabase(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c3Dir, "README.md"), []byte(`---
id: c3-0
title: Test
---

# Test

## Goal

Hello.
`), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "import"}, &buf)
	if err == nil {
		t.Fatal("expected unsealed import to fail")
	}
	if !strings.Contains(err.Error(), "unsealed C3 file") {
		t.Fatalf("expected unsealed error, got %v", err)
	}
}

func TestRun_ImportRequiresForceWithExistingDatabase(t *testing.T) {
	c3Dir := setupC3DB(t)
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"--c3-dir", c3Dir, "import"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected force guard")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected --force hint, got %v", err)
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

func TestRun_MutatingCommandBypassesBrokenCanonicalPreverify(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	readmePath := filepath.Join(c3Dir, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(data), "c3-seal:", "c3-seal: broken-", 1)
	if err := os.WriteFile(readmePath, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}

	err = run([]string{"--c3-dir", c3Dir, "list"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "canonical .c3/ is not ready") {
		t.Fatalf("read-only command should still preverify broken canonical docs, got %v", err)
	}

	body := "Updated through section repair.\n"
	r, w, _ := os.Pipe()
	w.WriteString(body)
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "write", "c3-0", "--section", "Goal"}, &buf); err != nil {
		t.Fatalf("section write should reach command validation/export despite broken canonical docs: %v", err)
	}
	if !strings.Contains(buf.String(), `Updated c3-0 section "Goal"`) {
		t.Fatalf("expected section write output, got:\n%s", buf.String())
	}

	data, err = os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	broken = strings.Replace(string(data), "c3-seal:", "c3-seal: broken-", 1)
	if err := os.WriteFile(readmePath, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}

	body = "## Goal\n\nDocument the repair work order.\n"
	r, w, _ = os.Pipe()
	w.WriteString(body)
	w.Close()
	os.Stdin = r

	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "add", "adr", "repair-workflow"}, &buf); err != nil {
		t.Fatalf("mutating add should reach command validation/export despite broken canonical docs: %v", err)
	}
	if !strings.Contains(buf.String(), "adr-") {
		t.Fatalf("expected add output, got:\n%s", buf.String())
	}

	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "verify"}, &buf); err != nil {
		t.Fatalf("expected post-mutation canonical export to verify: %v\n%s", err, buf.String())
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

func TestRun_VerifyRebuildsMissingDB(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(c3Dir, "c3.db")); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "verify"}, &buf); err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(c3Dir, "c3.db")) {
		t.Fatal("expected verify to rebuild local db cache")
	}
	if !strings.Contains(buf.String(), "OK: canonical markdown is in sync") {
		t.Fatalf("expected verify success, got %q", buf.String())
	}
}

func TestRun_RepairResealsBrokenCanonicalTree(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}
	readmePath := filepath.Join(c3Dir, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(data), "c3-seal:", "c3-seal: broken-", 1)
	if err := os.WriteFile(readmePath, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "repair"}, &buf); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "verify"}, &buf); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "BROKEN_SEAL") {
		t.Fatalf("expected repaired seals, got %q", buf.String())
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

func TestRun_Unwire(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	// Wire first
	run([]string{"--c3-dir", c3Dir, "wire", "c3-101", "ref-jwt"}, &buf)
	buf.Reset()
	err := run([]string{"--c3-dir", c3Dir, "unwire", "c3-101", "ref-jwt"}, &buf)
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

func TestRun_Query(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "query", "auth", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Diff(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "diff"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_DiffMark(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "diff", "--mark"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Impact(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "impact", "c3-101", "--json"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_Export(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "exported")
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "export", outDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun_SyncExport(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "synced")
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "sync", "export", outDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Synced canonical markdown") {
		t.Fatalf("expected sync summary, got %q", buf.String())
	}
	if !fileExists(filepath.Join(outDir, "README.md")) {
		t.Fatal("expected synced README.md")
	}
}

func TestRun_SyncCheck(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "synced")
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", outDir}, &buf); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "sync", "check", outDir}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "OK: canonical markdown is in sync") {
		t.Fatalf("expected sync check success, got %q", buf.String())
	}
}

func TestRun_SyncCheckDetectsDrift(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "synced")
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", outDir}, &buf); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "README.md"), []byte("drift\n"), 0644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	err := run([]string{"--c3-dir", c3Dir, "sync", "check", outDir}, &buf)
	if err == nil {
		t.Fatal("expected sync drift error")
	}
	if !strings.Contains(buf.String(), "DIFFERS README.md") {
		t.Fatalf("expected diff report, got %q", buf.String())
	}
}

func TestRun_SyncCheckDetectsBrokenSeal(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "synced")
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", outDir}, &buf); err != nil {
		t.Fatal(err)
	}
	readmePath := filepath.Join(outDir, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(data), "c3-seal:", "c3-seal: broken-", 1)
	if err := os.WriteFile(readmePath, []byte(broken), 0644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	err = run([]string{"--c3-dir", c3Dir, "sync", "check", outDir}, &buf)
	if err == nil {
		t.Fatal("expected broken seal error")
	}
	if !strings.Contains(buf.String(), "BROKEN_SEAL README.md") {
		t.Fatalf("expected broken seal report, got %q", buf.String())
	}
}

func TestRun_SyncExport_RemovesStaleCanonicalFiles(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	staleADRDir := filepath.Join(c3Dir, "adr")
	if err := os.MkdirAll(staleADRDir, 0755); err != nil {
		t.Fatal(err)
	}
	stalePath := filepath.Join(staleADRDir, "adr-00000000-stale.md")
	if err := os.WriteFile(stalePath, []byte("stale\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}
	if fileExists(stalePath) {
		t.Fatal("expected stale canonical file to be removed by sync export")
	}
}

func TestRun_SyncCheck_IgnoresIndexFiles(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	outDir := filepath.Join(t.TempDir(), "synced")
	var buf bytes.Buffer
	if err := run([]string{"--c3-dir", c3Dir, "sync", "export", outDir}, &buf); err != nil {
		t.Fatal(err)
	}
	indexDir := filepath.Join(outDir, "_index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(indexDir, "structural.md"), []byte("ignored\n"), 0644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := run([]string{"--c3-dir", c3Dir, "sync", "check", outDir}, &buf); err != nil {
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

func TestRun_SetWithSection(t *testing.T) {
	c3Dir := setupRichC3DB(t)
	var buf bytes.Buffer
	err := run([]string{"--c3-dir", c3Dir, "set", "c3-0", "--section", "Goal", "New goal text"}, &buf)
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
