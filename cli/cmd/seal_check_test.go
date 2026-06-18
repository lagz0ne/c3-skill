package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// exportFixtureToDisk exports the rich DB fixture to a c3Dir on disk so the
// on-disk facts carry real, intact seals. Returns the store and c3Dir.
func exportFixtureToDisk(t *testing.T) string {
	t.Helper()
	s := createRichDBFixture(t)
	c3Dir := filepath.Join(t.TempDir(), ".c3")
	var buf bytes.Buffer
	if err := RunExport(ExportOptions{Store: s, OutputDir: c3Dir}, &buf); err != nil {
		t.Fatalf("RunExport: %v", err)
	}
	return c3Dir
}

func checkIssues(t *testing.T, c3Dir string, only ...string) []Issue {
	t.Helper()
	// Re-import the on-disk docs into a fresh store so the store mirrors disk,
	// then run check against that store with C3Dir wired for the seal pass.
	dbDir := t.TempDir()
	tmpC3 := filepath.Join(dbDir, ".c3")
	if err := copyTree(t, c3Dir, tmpC3); err != nil {
		t.Fatalf("copyTree: %v", err)
	}
	s := importDir(t, tmpC3)
	var buf bytes.Buffer
	opts := CheckOptions{Store: s, C3Dir: tmpC3, JSON: true, Only: only}
	_ = RunCheckV2(opts, &buf)
	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid check JSON: %v\n%s", err, buf.String())
	}
	return result.Issues
}

func copyTree(t *testing.T, src, dst string) error {
	t.Helper()
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

func importDir(t *testing.T, c3Dir string) *store.Store {
	t.Helper()
	var buf bytes.Buffer
	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true, SkipBackup: true}, &buf); err != nil {
		t.Fatalf("RunImport: %v\n%s", err, buf.String())
	}
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open imported store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func hasSealWarn(issues []Issue, entityID string) bool {
	for _, iss := range issues {
		if iss.Severity != "warning" {
			continue
		}
		if !strings.Contains(strings.ToLower(iss.Message), "seal") {
			continue
		}
		if entityID == "" || iss.Entity == entityID {
			return true
		}
	}
	return false
}

// TestRunCheck_WarnsOnFactHandEditedSinceSeal — a fact whose on-disk content no
// longer matches its committed c3-seal (hand-edited after sealing) surfaces a
// seal-mismatch WARN from `c3x check`.
// TestFilterCanvasPaths — canvas paths are dropped from a broken-seal list (the
// fix that keeps the seal list consistent with the canvas-excluded sync diff).
func TestFilterCanvasPaths(t *testing.T) {
	in := []string{"canvases/adr.md", "c3-1-api/c3-101.md", "canvases/component.md", "refs/ref-x.md"}
	out := filterCanvasPaths(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 non-canvas paths, got %v", out)
	}
	for _, p := range out {
		if isCanvasCanonicalPath(p) {
			t.Errorf("filterCanvasPaths left a canvas path: %s", p)
		}
	}
}

// TestCheckFactSeals_SkipsCanvases — the fact-seal-on-disk check ignores canvases
// (user-owned definitions), so a canvas with a mismatched seal is not flagged.
func TestCheckFactSeals_SkipsCanvases(t *testing.T) {
	c3Dir := exportFixtureToDisk(t)
	canvasDir := filepath.Join(c3Dir, "canvases")
	if err := os.MkdirAll(canvasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A canvas with a deliberately wrong seal must NOT surface as a fact-seal mismatch.
	bad := "---\nid: component\nc3-seal: " + strings.Repeat("0", 64) + "\ntype: canvas\ndescription: 'Component definition'\n---\n\ndomain: software\nsections: []\n"
	if err := os.WriteFile(filepath.Join(canvasDir, "component.md"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, i := range checkFactSealsOnDisk(c3Dir) {
		if i.Entity == "component" {
			t.Fatalf("a canvas seal mismatch must be skipped by the fact-seal check, got: %+v", i)
		}
	}
}

func TestRunCheck_WarnsOnFactHandEditedSinceSeal(t *testing.T) {
	c3Dir := exportFixtureToDisk(t)
	compPath := filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")
	data, err := os.ReadFile(compPath)
	if err != nil {
		t.Fatalf("read component: %v", err)
	}
	// Hand-edit the body without touching the seal — breaks the seal.
	edited := string(data) + "\n\n## Sneaky Hand Edit\n\nAdded after sealing.\n"
	if err := os.WriteFile(compPath, []byte(edited), 0644); err != nil {
		t.Fatalf("write edited component: %v", err)
	}

	issues := checkIssues(t, c3Dir)
	if !hasSealWarn(issues, "c3-101") {
		t.Fatalf("expected seal-mismatch WARN for hand-edited fact c3-101, got: %+v", issues)
	}
}

// TestRunCheck_WarnsOnChangeDocHandEditedSinceSeal — a change doc (e.g. an ADR)
// whose on-disk body was hand-edited after sealing surfaces a seal-mismatch WARN
// from `c3x check`, not only at import. The change-doc seal exemption is removed:
// tampering with a sealed change-doc body must be caught during routine check.
func TestRunCheck_WarnsOnChangeDocHandEditedSinceSeal(t *testing.T) {
	c3Dir := exportFixtureToDisk(t)
	adrPath := findDocByID(t, c3Dir, "adr-20260226-use-go")

	data, err := os.ReadFile(adrPath)
	if err != nil {
		t.Fatalf("read adr: %v", err)
	}
	// Hand-edit the body without touching the seal — breaks the change-doc seal.
	edited := string(data) + "\n\n## Sneaky Edit\n\nTampered after sealing.\n"
	if err := os.WriteFile(adrPath, []byte(edited), 0644); err != nil {
		t.Fatalf("write edited adr: %v", err)
	}

	issues := checkIssues(t, c3Dir)
	if !hasSealWarn(issues, "adr-20260226-use-go") {
		t.Fatalf("expected seal-mismatch WARN for hand-edited change doc adr-20260226-use-go, got: %+v", issues)
	}
	// It must be a WARN, never a hard FAIL (provenance is judgment).
	for _, iss := range issues {
		if iss.Entity == "adr-20260226-use-go" && strings.Contains(strings.ToLower(iss.Message), "seal") {
			if iss.Severity != "warning" {
				t.Fatalf("change-doc seal mismatch must be severity \"warning\", got %q: %+v", iss.Severity, iss)
			}
		}
	}
}

// findDocByID walks the on-disk .c3 tree and returns the path of the markdown file
// whose frontmatter id matches the given id.
func findDocByID(t *testing.T, c3Dir, id string) string {
	t.Helper()
	var found string
	err := filepath.Walk(c3Dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(data), "id: "+id+"\n") {
			found = p
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk c3Dir: %v", err)
	}
	if found == "" {
		t.Fatalf("doc %q not found on disk under %s", id, c3Dir)
	}
	return found
}

// TestRunCheck_NoSealWarnForIntactFact — false-positive guard: a fact whose seal
// is intact (untouched export) produces NO seal-mismatch WARN.
func TestRunCheck_NoSealWarnForIntactFact(t *testing.T) {
	c3Dir := exportFixtureToDisk(t)
	issues := checkIssues(t, c3Dir)
	if hasSealWarn(issues, "") {
		t.Fatalf("expected no seal WARN for intact facts, got: %+v", issues)
	}
}

// TestRunCheck_SealMismatchSurfacesAsWarnNotFail — a seal mismatch is a WARN
// (severity "warning"), never a hard error/FAIL: provenance is judgment, c3x
// only reports the mechanical mismatch.
func TestRunCheck_SealMismatchSurfacesAsWarnNotFail(t *testing.T) {
	c3Dir := exportFixtureToDisk(t)
	compPath := filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md")
	data, err := os.ReadFile(compPath)
	if err != nil {
		t.Fatalf("read component: %v", err)
	}
	edited := string(data) + "\n\n## Drift\n\nContent changed post-seal.\n"
	if err := os.WriteFile(compPath, []byte(edited), 0644); err != nil {
		t.Fatalf("write edited component: %v", err)
	}

	issues := checkIssues(t, c3Dir)
	sawSealWarn := false
	for _, iss := range issues {
		if !strings.Contains(strings.ToLower(iss.Message), "seal") {
			continue
		}
		if iss.Entity != "c3-101" {
			continue
		}
		sawSealWarn = true
		if iss.Severity != "warning" {
			t.Fatalf("seal mismatch must be severity \"warning\", got %q: %+v", iss.Severity, iss)
		}
	}
	if !sawSealWarn {
		t.Fatalf("expected a seal-mismatch issue for c3-101, got: %+v", issues)
	}
}

// TestRunCheck_SealCheckDoesNotJudgeProvenance — NEGATIVE / the line. The
// seal-mismatch WARN is purely mechanical: it does NOT label the edit as
// legitimate vs sneaky, and does NOT escalate to a hard FAIL. The message must
// not assert provenance/intent, and the check must still succeed (no error
// return) on a seal mismatch alone.
func TestRunCheck_SealCheckDoesNotJudgeProvenance(t *testing.T) {
	c3Dir := exportFixtureToDisk(t)

	dbDir := t.TempDir()
	tmpC3 := filepath.Join(dbDir, ".c3")
	if err := copyTree(t, c3Dir, tmpC3); err != nil {
		t.Fatalf("copyTree: %v", err)
	}
	compPath := filepath.Join(tmpC3, "c3-1-api", "c3-101-auth.md")
	data, err := os.ReadFile(compPath)
	if err != nil {
		t.Fatalf("read component: %v", err)
	}
	s := importDir(t, tmpC3)
	// Edit AFTER import so the store mirrors the original seal but disk drifted.
	edited := string(data) + "\n\n## Reseal Or Sneak\n\nAmbiguous edit.\n"
	if err := os.WriteFile(compPath, []byte(edited), 0644); err != nil {
		t.Fatalf("write edited component: %v", err)
	}

	var buf bytes.Buffer
	err = RunCheckV2(CheckOptions{Store: s, C3Dir: tmpC3, JSON: true}, &buf)
	if err != nil {
		t.Fatalf("seal mismatch must not FAIL the check, got error: %v", err)
	}
	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid check JSON: %v\n%s", err, buf.String())
	}
	if !hasSealWarn(result.Issues, "c3-101") {
		t.Fatalf("expected a seal-mismatch WARN for c3-101, got: %+v", result.Issues)
	}
	for _, iss := range result.Issues {
		if !strings.Contains(strings.ToLower(iss.Message), "seal") {
			continue
		}
		lower := strings.ToLower(iss.Message + " " + iss.Hint)
		for _, judgement := range []string{"sneaky", "legitimate", "illegit", "unauthorized", "tamper", "malicious"} {
			if strings.Contains(lower, judgement) {
				t.Fatalf("seal WARN must not judge provenance (found %q): %+v", judgement, iss)
			}
		}
		if iss.Severity == "error" {
			t.Fatalf("seal mismatch must never be severity \"error\": %+v", iss)
		}
	}
}

// TestHintFor_MisfitPointsToChangeDocNotTableEdit — the layer-disconnect
// (misfit) hint no longer instructs a direct parent-table / child-parent edit;
// it points at opening a change doc that amends the parent top-down.
func TestHintFor_MisfitPointsToChangeDocNotTableEdit(t *testing.T) {
	hint := hintFor("layer disconnect: child component c3-110 has parent c3-1 but is missing from c3-1 Components table")
	if hint == "" {
		t.Fatal("expected a non-empty layer-disconnect hint")
	}
	lower := strings.ToLower(hint)
	for _, banned := range []string{"update parent table", "fix the child parent field", "fix the child parent"} {
		if strings.Contains(lower, banned) {
			t.Fatalf("layer-disconnect hint must not instruct a direct table edit (found %q): %q", banned, hint)
		}
	}
	if !strings.Contains(lower, "change doc") {
		t.Fatalf("layer-disconnect hint should point at opening a change doc, got: %q", hint)
	}
}
