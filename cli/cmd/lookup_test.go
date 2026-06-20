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

// createLookupFixture creates a DB fixture with goal/summary and a .c3/eval dir for
// fact→code bindings. The fact→code binding lives in .c3/eval/<fact>.yaml `code:`,
// which is what `lookup` resolves through (replacing the code-map).
func createLookupFixture(t *testing.T) (*store.Store, string) {
	t.Helper()
	s := createDBFixture(t)

	// Update c3-101 with goal + summary
	entity, _ := s.GetEntity("c3-101")
	entity.Goal = "Handle authentication and JWT issuance"
	s.UpdateEntity(entity)

	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	return s, c3Dir
}

// bindCode writes a .c3/eval/<fact>.yaml carrying the fact's code globs — the
// binding `lookup` resolves a file against.
func bindCode(t *testing.T, c3Dir, fact string, globs ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("fact: " + fact + "\ncode:\n")
	for _, g := range globs {
		b.WriteString("  - " + g + "\n")
	}
	path := filepath.Join(c3Dir, "eval", fact+".yaml")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunLookup_ExactMatch(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/login.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101, got:\n%s", out)
	}
	if !strings.Contains(out, "Handle authentication") {
		t.Errorf("expected goal in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ref-jwt") {
		t.Errorf("expected ref-jwt listed, got:\n%s", out)
	}
	if !strings.Contains(out, "Standardize auth tokens") {
		t.Errorf("expected ref goal in output, got:\n%s", out)
	}
}

func TestRunLookup_GlobStar(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "c3-101") {
		t.Errorf("*.ts glob should match login.ts, got:\n%s", buf.String())
	}
}

func TestRunLookup_DoubleStar(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/**/*.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/handlers/login.ts", C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "c3-101") {
		t.Errorf("** glob should match nested file, got:\n%s", buf.String())
	}
}

func TestRunLookup_NoMatch(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/payments/stripe.go", C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("expected no-match message, got:\n%s", buf.String())
	}
}

func TestRunLookup_NoBinding(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	// No eval bindings.

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("missing binding should produce no-match, got:\n%s", buf.String())
	}
}

func TestRunLookup_MultipleComponents(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/**/*.ts")
	bindCode(t, c3Dir, "c3-110", "src/auth/*.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in multi-match output, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-110") {
		t.Errorf("expected c3-110 in multi-match output, got:\n%s", out)
	}
}

func TestRunLookup_JSON(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", JSON: true, C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result LookupResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if result.File != "src/auth/login.ts" {
		t.Errorf("expected file field, got %q", result.File)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Matches))
	}

	m := result.Matches[0]
	if m.ID != "c3-101" {
		t.Errorf("expected id c3-101, got %q", m.ID)
	}
	if m.Goal == "" {
		t.Error("expected non-empty goal")
	}
	if len(m.Refs) != 1 || m.Refs[0].ID != "ref-jwt" {
		t.Errorf("expected ref-jwt, got %v", m.Refs)
	}
}

func TestRunLookup_GlobInput_MultipleFiles(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")
	projectDir := filepath.Dir(c3Dir)

	// Create actual source files for glob expansion
	os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0755)
	writeFile(t, filepath.Join(projectDir, "src", "auth", "jwt.ts"), "// jwt")
	writeFile(t, filepath.Join(projectDir, "src", "auth", "middleware.ts"), "// middleware")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Store:      s,
		FilePath:   "src/auth/*.ts",
		ProjectDir: projectDir,
		C3Dir:      c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "pattern: src/auth/*.ts") {
		t.Errorf("expected pattern header, got:\n%s", out)
	}
	if !strings.Contains(out, "2 file(s) matched") {
		t.Errorf("expected 2 files matched, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in components, got:\n%s", out)
	}
}

func TestRunLookup_JSONNoMatch(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/other/file.ts", JSON: true, C3Dir: c3Dir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result LookupResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(result.Matches) != 0 {
		t.Errorf("expected empty matches, got %v", result.Matches)
	}
}

func TestRunLookup_WithRulesInOutput(t *testing.T) {
	s := createDBFixture(t)
	// Add a rule entity and wire it to c3-101
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging", Slug: "logging",
		Goal: "Structured logging", Status: "active", Metadata: "{}",
	})
	s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "rule-logging", RelType: "uses"})

	// Give c3-101 a goal
	entity, _ := s.GetEntity("c3-101")
	entity.Goal = "Authentication component"
	s.UpdateEntity(entity)

	// Bind code so lookup can find c3-101
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	bindCode(t, c3Dir, "c3-101", "src/auth/**")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Store:    s,
		FilePath: "src/auth/login.ts",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "rules:") {
		t.Errorf("should show rules section in output, got:\n%s", output)
	}
	if !strings.Contains(output, "rule-logging") {
		t.Errorf("should show rule-logging in output, got:\n%s", output)
	}
}

func TestLookupSeparatesRulesFromRefs(t *testing.T) {
	s := createDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging", Slug: "logging",
		Goal: "Structured logging", Status: "active", Metadata: "{}",
	})
	s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "rule-logging", RelType: "uses"})

	entity, _ := s.GetEntity("c3-101")
	match := buildMatchFromStore(entity, s)

	if len(match.Refs) != 1 || match.Refs[0].ID != "ref-jwt" {
		t.Errorf("Refs = %v, want [ref-jwt]", match.Refs)
	}
	if len(match.Rules) != 1 || match.Rules[0].ID != "rule-logging" {
		t.Errorf("Rules = %v, want [rule-logging]", match.Rules)
	}
}

// factsForFile is the resolution behind lookup: a file maps to a fact iff one of
// the fact's eval-spec code globs matches it, and the result is sorted.
func TestFactsForFile(t *testing.T) {
	bindings := map[string][]string{
		"c3-101": {"src/auth/**/*.ts"},
		"c3-110": {"src/auth/*.ts", "src/shared/**"},
		"c3-200": {"src/payments/**"},
	}
	got := factsForFile(bindings, "src/auth/login.ts")
	if len(got) != 2 || got[0] != "c3-101" || got[1] != "c3-110" {
		t.Errorf("factsForFile = %v, want sorted [c3-101 c3-110]", got)
	}
	if got := factsForFile(bindings, "src/payments/stripe.go"); len(got) != 1 || got[0] != "c3-200" {
		t.Errorf("factsForFile = %v, want [c3-200]", got)
	}
	if got := factsForFile(bindings, "src/other/x.go"); len(got) != 0 {
		t.Errorf("factsForFile = %v, want []", got)
	}
}
