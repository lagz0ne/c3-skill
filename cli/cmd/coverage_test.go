package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func createCoverageFixture(t *testing.T) (*store.Store, string) {
	t.Helper()
	s := createDBFixture(t)
	projectDir := t.TempDir()

	// Create source files in the project
	for _, f := range []string{
		"src/auth/login.ts",
		"src/auth/login.test.ts",
		"src/billing/invoice.ts",
		"src/shared/utils.ts",
	} {
		full := filepath.Join(projectDir, filepath.FromSlash(f))
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("// "+f), 0644)
	}

	return s, projectDir
}

func TestRunCoverage_DefaultJSON(t *testing.T) {
	s, projectDir := createCoverageFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/login.ts"})
	s.AddExclude("**/*.test.ts")

	t.Setenv("HUMAN", "")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		Store:      s,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result codemap.CoverageResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("default output should be JSON: %v\n%s", err, buf.String())
	}
	if result.Total == 0 {
		t.Error("expected non-zero total")
	}
}

func TestRunCoverage_HumanOutput(t *testing.T) {
	s, projectDir := createCoverageFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/login.ts"})
	s.AddExclude("**/*.test.ts")

	t.Setenv("HUMAN", "1")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		Store:      s,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "C3 Code-Map Coverage") {
		t.Errorf("expected header, got:\n%s", out)
	}
	if !strings.Contains(out, "mapped:") {
		t.Errorf("expected mapped line, got:\n%s", out)
	}
}

func TestRunCoverage_NoCodeMap(t *testing.T) {
	s, projectDir := createCoverageFixture(t)
	// No code-map entries

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		Store:      s,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result codemap.CoverageResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if result.Mapped != 0 {
		t.Errorf("expected 0 mapped with no code-map, got %d", result.Mapped)
	}
}

func TestRunCoverage_AllMapped(t *testing.T) {
	s, projectDir := createCoverageFixture(t)
	s.SetCodeMap("c3-101", []string{"src/**/*.ts"})

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		Store:      s,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result codemap.CoverageResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if result.Unmapped != 0 {
		t.Errorf("expected 0 unmapped, got %d: %v", result.Unmapped, result.UnmappedFiles)
	}
	if result.CoveragePct != 100 {
		t.Errorf("expected 100%% coverage, got %.1f%%", result.CoveragePct)
	}
}
