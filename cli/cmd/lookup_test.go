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

// createLookupFixture creates a DB fixture with goal/summary and code-map.
func createLookupFixture(t *testing.T) (*store.Store, string) {
	t.Helper()
	s := createDBFixture(t)

	// Update c3-101 with goal + summary
	entity, _ := s.GetEntity("c3-101")
	entity.Goal = "Handle authentication and JWT issuance"
	s.UpdateEntity(entity)

	projectDir := t.TempDir()
	return s, projectDir
}

func TestRunLookup_ExactMatch(t *testing.T) {
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/login.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts"}, &buf)
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
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/*.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "c3-101") {
		t.Errorf("*.ts glob should match login.ts, got:\n%s", buf.String())
	}
}

func TestRunLookup_DoubleStar(t *testing.T) {
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**/*.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/handlers/login.ts"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "c3-101") {
		t.Errorf("** glob should match nested file, got:\n%s", buf.String())
	}
}

func TestRunLookup_NoMatch(t *testing.T) {
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/*.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/payments/stripe.go"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("expected no-match message, got:\n%s", buf.String())
	}
}

func TestRunLookup_NoCodeMap(t *testing.T) {
	s, _ := createLookupFixture(t)
	// No code-map entries

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "no component mapping found") {
		t.Errorf("missing code-map should produce no-match, got:\n%s", buf.String())
	}
}

func TestRunLookup_MultipleComponents(t *testing.T) {
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**/*.ts"})
	s.SetCodeMap("c3-110", []string{"src/auth/*.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts"}, &buf)
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
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/*.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", JSON: true}, &buf)
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
	s, projectDir := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/*.ts"})

	// Create actual source files for glob expansion
	os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0755)
	writeFile(t, filepath.Join(projectDir, "src", "auth", "jwt.ts"), "// jwt")
	writeFile(t, filepath.Join(projectDir, "src", "auth", "middleware.ts"), "// middleware")

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Store:      s,
		FilePath:   "src/auth/*.ts",
		ProjectDir: projectDir,
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
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/*.ts"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{Store: s, FilePath: "src/other/file.ts", JSON: true}, &buf)
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

	// Set code map so lookup can find c3-101
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	err := RunLookup(LookupOptions{
		Store:    s,
		FilePath: "src/auth/login.ts",
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
