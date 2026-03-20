package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
