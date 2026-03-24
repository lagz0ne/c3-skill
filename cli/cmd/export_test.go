package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunExport(t *testing.T) {
	s := createDBFixture(t)
	outDir := filepath.Join(t.TempDir(), "exported")

	var buf bytes.Buffer
	err := RunExport(ExportOptions{Store: s, OutputDir: outDir}, &buf)
	if err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Exported") {
		t.Errorf("expected 'Exported' summary, got:\n%s", out)
	}

	// Verify system context → README.md
	readmePath := filepath.Join(outDir, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("expected README.md, got error: %v", err)
	}
	if !strings.Contains(string(content), "id: c3-0") {
		t.Errorf("README.md should contain 'id: c3-0', got:\n%s", string(content))
	}

	// Verify container → c3-1-api/README.md
	containerPath := filepath.Join(outDir, "c3-1-api", "README.md")
	content, err = os.ReadFile(containerPath)
	if err != nil {
		t.Fatalf("expected container README.md, got error: %v", err)
	}
	if !strings.Contains(string(content), "id: c3-1") {
		t.Errorf("container README should contain 'id: c3-1', got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "boundary: service") {
		t.Errorf("container README should contain 'boundary: service', got:\n%s", string(content))
	}

	// Verify component → c3-1-api/c3-101-auth.md
	compPath := filepath.Join(outDir, "c3-1-api", "c3-101-auth.md")
	content, err = os.ReadFile(compPath)
	if err != nil {
		t.Fatalf("expected component file, got error: %v", err)
	}
	if !strings.Contains(string(content), "id: c3-101") {
		t.Errorf("component file should contain 'id: c3-101', got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "uses: [ref-jwt]") {
		t.Errorf("component file should contain 'uses: [ref-jwt]', got:\n%s", string(content))
	}

	// Verify ref → refs/ref-jwt.md
	refPath := filepath.Join(outDir, "refs", "ref-jwt.md")
	content, err = os.ReadFile(refPath)
	if err != nil {
		t.Fatalf("expected ref file, got error: %v", err)
	}
	if !strings.Contains(string(content), "id: ref-jwt") {
		t.Errorf("ref file should contain 'id: ref-jwt', got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "scope: [c3-1]") {
		t.Errorf("ref file should contain 'scope: [c3-1]', got:\n%s", string(content))
	}

	// Verify ADR → adr/adr-20260226-use-go.md
	adrPath := filepath.Join(outDir, "adr", "adr-20260226-use-go.md")
	content, err = os.ReadFile(adrPath)
	if err != nil {
		t.Fatalf("expected ADR file, got error: %v", err)
	}
	if !strings.Contains(string(content), "id: adr-20260226-use-go") {
		t.Errorf("ADR file should contain 'id: adr-20260226-use-go', got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "status: proposed") {
		t.Errorf("ADR file should contain 'status: proposed', got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "affects: [c3-0]") {
		t.Errorf("ADR file should contain 'affects: [c3-0]', got:\n%s", string(content))
	}
}

func TestRunExport_EntityCount(t *testing.T) {
	s := createDBFixture(t)
	outDir := filepath.Join(t.TempDir(), "exported")

	var buf bytes.Buffer
	err := RunExport(ExportOptions{Store: s, OutputDir: outDir}, &buf)
	if err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	// Fixture has 7 entities
	if !strings.Contains(buf.String(), "Exported 7 entities") {
		t.Errorf("expected 'Exported 7 entities', got:\n%s", buf.String())
	}
}

func TestRunExport_CodeMap(t *testing.T) {
	s := createDBFixture(t)
	// Add a code map entry
	if err := s.SetCodeMap("c3-101", []string{"src/auth/**"}); err != nil {
		t.Fatalf("set code map: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "exported")
	var buf bytes.Buffer
	err := RunExport(ExportOptions{Store: s, OutputDir: outDir}, &buf)
	if err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	cmPath := filepath.Join(outDir, "code-map.yaml")
	content, err := os.ReadFile(cmPath)
	if err != nil {
		t.Fatalf("expected code-map.yaml, got error: %v", err)
	}
	if !strings.Contains(string(content), "c3-101") {
		t.Errorf("code-map.yaml should contain c3-101, got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "src/auth/**") {
		t.Errorf("code-map.yaml should contain pattern, got:\n%s", string(content))
	}
}

func TestRunExport_AllTypes(t *testing.T) {
	s := createRichDBFixture(t)
	// Add recipe and rule entities
	s.InsertEntity(&store.Entity{
		ID: "recipe-auth-flow", Type: "recipe", Title: "Auth Flow",
		Slug: "auth-flow", Goal: "End-to-end auth", Status: "active", Metadata: "{}",
	})
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Logging",
		Slug: "logging", Goal: "Structured logging", Status: "active", Metadata: "{}",
	})
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	outDir := filepath.Join(t.TempDir(), "export")
	var buf bytes.Buffer
	err := RunExport(ExportOptions{Store: s, OutputDir: outDir}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Check recipe file exists
	if _, err := os.Stat(filepath.Join(outDir, "recipes", "recipe-auth-flow.md")); err != nil {
		t.Error("recipe file should be exported")
	}
	// Check rule file exists
	if _, err := os.Stat(filepath.Join(outDir, "rules", "rule-logging.md")); err != nil {
		t.Error("rule file should be exported")
	}
	// Check ADR file exists (with date)
	adrFiles, _ := filepath.Glob(filepath.Join(outDir, "adr", "adr-*.md"))
	if len(adrFiles) == 0 {
		t.Error("ADR file should be exported")
	}
	// Check code-map
	if _, err := os.Stat(filepath.Join(outDir, "code-map.yaml")); err != nil {
		t.Error("code-map.yaml should be exported")
	}
}

func TestEntityExportPath_OrphanComponent(t *testing.T) {
	e := &store.Entity{ID: "c3-999", Type: "component", Slug: "orphan", ParentID: "c3-missing"}
	path := entityExportPath("/out", e, map[string]string{})
	if !strings.Contains(path, "orphans") {
		t.Errorf("orphan component path = %q, want 'orphans' dir", path)
	}
}

func TestEntityExportPath_UnknownType(t *testing.T) {
	e := &store.Entity{ID: "x", Type: "unknown", Slug: "x"}
	path := entityExportPath("/out", e, map[string]string{})
	if path != "" {
		t.Errorf("unknown type should return empty path, got %q", path)
	}
}

func TestEntityExportPath_System(t *testing.T) {
	e := &store.Entity{ID: "c3-0", Type: "system", Slug: ""}
	path := entityExportPath("/out", e, map[string]string{})
	if path != filepath.Join("/out", "README.md") {
		t.Errorf("system path = %q", path)
	}
}

func TestEntityExportPath_Container(t *testing.T) {
	e := &store.Entity{ID: "c3-1", Type: "container", Slug: "api"}
	path := entityExportPath("/out", e, map[string]string{})
	if path != filepath.Join("/out", "c3-1-api", "README.md") {
		t.Errorf("container path = %q", path)
	}
}

func TestEntityExportPath_Ref(t *testing.T) {
	e := &store.Entity{ID: "ref-jwt", Type: "ref", Slug: "jwt"}
	path := entityExportPath("/out", e, map[string]string{})
	if path != filepath.Join("/out", "refs", "ref-jwt.md") {
		t.Errorf("ref path = %q", path)
	}
}

func TestEntityExportPath_ADRWithDate(t *testing.T) {
	e := &store.Entity{ID: "adr-20260226-use-go", Type: "adr", Slug: "use-go", Date: "20260226"}
	path := entityExportPath("/out", e, map[string]string{})
	expected := filepath.Join("/out", "adr", "adr-20260226-use-go.md")
	if path != expected {
		t.Errorf("adr path with date = %q, want %q", path, expected)
	}
}

func TestEntityExportPath_ADRWithoutDate(t *testing.T) {
	e := &store.Entity{ID: "adr-use-go", Type: "adr", Slug: "use-go"}
	path := entityExportPath("/out", e, map[string]string{})
	expected := filepath.Join("/out", "adr", "adr-use-go.md")
	if path != expected {
		t.Errorf("adr path without date = %q, want %q", path, expected)
	}
}

func TestEntityExportPath_Recipe(t *testing.T) {
	e := &store.Entity{ID: "recipe-auth", Type: "recipe", Slug: "auth"}
	path := entityExportPath("/out", e, map[string]string{})
	expected := filepath.Join("/out", "recipes", "recipe-auth.md")
	if path != expected {
		t.Errorf("recipe path = %q, want %q", path, expected)
	}
}

func TestEntityExportPath_Rule(t *testing.T) {
	e := &store.Entity{ID: "rule-logging", Type: "rule", Slug: "logging"}
	path := entityExportPath("/out", e, map[string]string{})
	expected := filepath.Join("/out", "rules", "rule-logging.md")
	if path != expected {
		t.Errorf("rule path = %q, want %q", path, expected)
	}
}
