package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
)

func createCoverageFixture(t *testing.T) (c3Dir string, projectDir string) {
	t.Helper()
	c3Dir = createFixture(t)
	projectDir = filepath.Dir(c3Dir)

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

	return c3Dir, projectDir
}

func TestRunCoverage_DefaultJSON(t *testing.T) {
	c3Dir, projectDir := createCoverageFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/auth/login.ts\n_exclude:\n  - \"**/*.test.ts\"\n")

	// Ensure HUMAN is not set (default = JSON)
	t.Setenv("HUMAN", "")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
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
	c3Dir, projectDir := createCoverageFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/auth/login.ts\n_exclude:\n  - \"**/*.test.ts\"\n")

	t.Setenv("HUMAN", "1")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
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
	if !strings.Contains(out, "excluded:") {
		t.Errorf("expected excluded line, got:\n%s", out)
	}
	if !strings.Contains(out, "unmapped files:") {
		t.Errorf("expected unmapped files section, got:\n%s", out)
	}
}

func TestRunCoverage_JSONFlagOverridesHuman(t *testing.T) {
	c3Dir, projectDir := createCoverageFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/auth/login.ts\n")

	t.Setenv("HUMAN", "1")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
		JSON:       true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result codemap.CoverageResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("--json should force JSON even with HUMAN set: %v\n%s", err, buf.String())
	}
}

func TestRunCoverage_NoCodeMap(t *testing.T) {
	c3Dir, projectDir := createCoverageFixture(t)
	// No code-map.yaml written

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Default is JSON
	var result codemap.CoverageResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if result.Mapped != 0 {
		t.Errorf("expected 0 mapped with no code-map, got %d", result.Mapped)
	}
}

func TestRunCoverage_AllMapped(t *testing.T) {
	c3Dir, projectDir := createCoverageFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/**/*.ts\n")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
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
		t.Errorf("expected 0 unmapped when all files match, got %d: %v", result.Unmapped, result.UnmappedFiles)
	}
	if result.CoveragePct != 100 {
		t.Errorf("expected 100%% coverage, got %.1f%%", result.CoveragePct)
	}
}

func TestRunCoverage_RefGovernance(t *testing.T) {
	c3Dir, projectDir := createCoverageFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/auth/login.ts\n_exclude:\n  - \"**/*.test.ts\"\n")

	t.Setenv("HUMAN", "")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	gov, ok := output["ref_governance"]
	if !ok {
		t.Fatal("expected ref_governance in output")
	}

	govMap, ok := gov.(map[string]interface{})
	if !ok {
		t.Fatal("ref_governance should be an object")
	}

	if govMap["total_components"] == nil {
		t.Error("expected total_components in ref_governance")
	}
}

func TestRunCoverage_RefGovernanceHuman(t *testing.T) {
	c3Dir, projectDir := createCoverageFixture(t)
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"),
		"c3-101:\n  - src/auth/login.ts\n")

	t.Setenv("HUMAN", "1")

	var buf bytes.Buffer
	err := RunCoverage(CoverageOptions{
		C3Dir:      c3Dir,
		ProjectDir: projectDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Ref Governance") {
		t.Errorf("expected Ref Governance section, got:\n%s", out)
	}
}
