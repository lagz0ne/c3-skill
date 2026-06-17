package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestHasErrorSeverity(t *testing.T) {
	if hasErrorSeverity([]Issue{{Severity: "warning"}, {Severity: "info"}}) {
		t.Error("no error-severity issue should report false")
	}
	if !hasErrorSeverity([]Issue{{Severity: "warning"}, {Severity: "error"}}) {
		t.Error("an error-severity issue should report true")
	}
	if hasErrorSeverity(nil) {
		t.Error("empty issues should report false")
	}
}

// TestStrictCodemap_GatesAutoDoneFlip is the regression for the blocking review
// finding: under --strict-codemap, an unresolved external binding on an affected
// entity must GATE the accepted->done flip (not let the doc latch to done and only
// then fail). Without --strict it is WARN-only and the flip proceeds.
func TestStrictCodemap_GatesAutoDoneFlip(t *testing.T) {
	affected := "\n## Affected Topology\n\n" +
		"| Entity | Type | Why affected | Evidence | Governance review |\n" +
		"|--|--|--|--|--|\n" +
		"| c3-101 | component | rebound auth | N.A - none | N.A - none |\n"
	projectDir := t.TempDir() // no files → the binding below resolves to nothing

	// strict: the unresolved binding gates accepted->done.
	s := createRichDBFixture(t)
	if err := s.SetCodeMap("c3-101", []string{"src/GHOST/**"}); err != nil {
		t.Fatal(err)
	}
	entity, body := seedAcceptedPRD(t, s, testCitationForEntity(t, s, "c3-1"), testCitationForEntity(t, s, "c3-101"))
	if err := content.WriteEntity(s, entity.ID, body+affected); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = RunCheckV2(CheckOptions{Store: s, ProjectDir: projectDir, JSON: true, IncludeADR: true, Fix: true, StrictCodemap: true, Only: []string{entity.ID}}, &buf)
	if got, _ := s.GetEntity(entity.ID); got.Status != "accepted" {
		t.Fatalf("--strict-codemap must block accepted->done while a binding is unresolved; got status %q", got.Status)
	}

	// non-strict: same broken binding is WARN-only; the doc still auto-dones.
	s2 := createRichDBFixture(t)
	if err := s2.SetCodeMap("c3-101", []string{"src/GHOST/**"}); err != nil {
		t.Fatal(err)
	}
	entity2, body2 := seedAcceptedPRD(t, s2, testCitationForEntity(t, s2, "c3-1"), testCitationForEntity(t, s2, "c3-101"))
	if err := content.WriteEntity(s2, entity2.ID, body2+affected); err != nil {
		t.Fatal(err)
	}
	var buf2 bytes.Buffer
	_ = RunCheckV2(CheckOptions{Store: s2, ProjectDir: projectDir, JSON: true, IncludeADR: true, Fix: true, StrictCodemap: false, Only: []string{entity2.ID}}, &buf2)
	if got, _ := s2.GetEntity(entity2.ID); got.Status != "done" {
		t.Fatalf("without --strict-codemap an unresolved binding is WARN-only; the doc should still auto-done, got status %q", got.Status)
	}
}

const introspectBody = "## Affected Topology\n\n" +
	"| Entity | Type | Why affected | Evidence | Governance review |\n" +
	"|--------|------|--------------|----------|-------------------|\n" +
	"| c3-101 | component | refactored auth | N.A - none | N.A - none |\n"

func introspectFixture(t *testing.T) (*store.Store, string) {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.InsertEntity(&store.Entity{ID: "c3-101", Type: "component", Title: "auth", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nAuth.\n"); err != nil {
		t.Fatal(err)
	}
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, "src", "exists"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "exists", "foo.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return s, projectDir
}

func TestCodemapIntrospection_WarnsOnUnresolvedGlob(t *testing.T) {
	s, projectDir := introspectFixture(t)
	if err := s.SetCodeMap("c3-101", []string{"src/exists/**", "src/missing/**"}); err != nil {
		t.Fatal(err)
	}
	issues := codemapIntrospection(s, projectDir, introspectBody, false)
	if len(issues) != 1 {
		t.Fatalf("want exactly 1 warning (only the missing glob), got %d: %+v", len(issues), issues)
	}
	if issues[0].Severity != "warning" {
		t.Errorf("severity = %q, want warning", issues[0].Severity)
	}
	if issues[0].Entity != "c3-101" || !strings.Contains(issues[0].Message, "src/missing/**") {
		t.Errorf("unexpected issue: %+v", issues[0])
	}
}

func TestCodemapIntrospection_StrictPromotesToError(t *testing.T) {
	s, projectDir := introspectFixture(t)
	if err := s.SetCodeMap("c3-101", []string{"src/missing/**"}); err != nil {
		t.Fatal(err)
	}
	issues := codemapIntrospection(s, projectDir, introspectBody, true)
	if len(issues) != 1 || issues[0].Severity != "error" {
		t.Fatalf("--strict-codemap must promote the binding warning to an error, got %+v", issues)
	}
}

func TestCodemapIntrospection_AllResolveNoIssues(t *testing.T) {
	s, projectDir := introspectFixture(t)
	if err := s.SetCodeMap("c3-101", []string{"src/exists/**"}); err != nil {
		t.Fatal(err)
	}
	if issues := codemapIntrospection(s, projectDir, introspectBody, false); len(issues) != 0 {
		t.Errorf("a resolving binding must produce no issues, got %+v", issues)
	}
}

func TestCodemapIntrospection_NoCodemapIsSilent(t *testing.T) {
	s, projectDir := introspectFixture(t)
	// c3-101 declares no codemap → nothing to match → silent (whether a fact should
	// bind code is the author's call, not the tool's).
	if issues := codemapIntrospection(s, projectDir, introspectBody, false); len(issues) != 0 {
		t.Errorf("an entity with no codemap must be silent, got %+v", issues)
	}
}

func TestCodemapIntrospection_NoProjectDirNoOp(t *testing.T) {
	if issues := codemapIntrospection(nil, "", introspectBody, false); issues != nil {
		t.Errorf("an empty projectDir must no-op, got %+v", issues)
	}
}
