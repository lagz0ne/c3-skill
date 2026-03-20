package codemap

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// setupCoverageDir creates a temp directory with source files (not a git repo).
func setupCoverageDir(t *testing.T, files []string) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		full := filepath.Join(dir, filepath.FromSlash(f))
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("// "+f), 0644)
	}
	return dir
}

func TestCoverage_MixedMappedExcludedUnmapped(t *testing.T) {
	dir := setupCoverageDir(t, []string{
		"src/auth/login.ts",
		"src/auth/login.test.ts",
		"src/billing/invoice.ts",
	})

	cm := CodeMap{
		"c3-101":   {"src/auth/**/*.ts"},
		"_exclude": {"**/*.test.ts"},
	}

	result, err := Coverage(cm, dir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Total != 3 {
		t.Errorf("total: got %d, want 3", result.Total)
	}
	// Both auth files match src/auth/**/*.ts → mapped; billing is unmapped
	if result.Mapped != 2 {
		t.Errorf("mapped: got %d, want 2", result.Mapped)
	}
	if result.Excluded != 0 {
		t.Errorf("excluded: got %d, want 0", result.Excluded)
	}
	if result.Unmapped != 1 {
		t.Errorf("unmapped: got %d, want 1", result.Unmapped)
	}
	if len(result.UnmappedFiles) != 1 || result.UnmappedFiles[0] != "src/billing/invoice.ts" {
		t.Errorf("unmapped files: got %v, want [src/billing/invoice.ts]", result.UnmappedFiles)
	}
}

func TestCoverage_ExcludeWorks(t *testing.T) {
	dir := setupCoverageDir(t, []string{
		"src/auth/login.ts",
		"src/auth/login.test.ts",
		"src/billing/invoice.ts",
	})

	// Only map login.ts specifically, not the test file
	cm := CodeMap{
		"c3-101":   {"src/auth/login.ts"},
		"_exclude": {"**/*.test.ts"},
	}

	result, err := Coverage(cm, dir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Mapped != 1 {
		t.Errorf("mapped: got %d, want 1", result.Mapped)
	}
	if result.Excluded != 1 {
		t.Errorf("excluded: got %d, want 1", result.Excluded)
	}
	if result.Unmapped != 1 {
		t.Errorf("unmapped: got %d, want 1", result.Unmapped)
	}
}

func TestCoverage_AllMapped(t *testing.T) {
	dir := setupCoverageDir(t, []string{
		"src/auth/login.ts",
		"src/api/handler.ts",
	})

	cm := CodeMap{
		"c3-101": {"src/**/*.ts"},
	}

	result, err := Coverage(cm, dir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Mapped != 2 {
		t.Errorf("mapped: got %d, want 2", result.Mapped)
	}
	if result.Unmapped != 0 {
		t.Errorf("unmapped: got %d, want 0", result.Unmapped)
	}
	if result.CoveragePct != 100 {
		t.Errorf("coverage: got %.1f%%, want 100%%", result.CoveragePct)
	}
}

func TestCoverage_EmptyCodeMap(t *testing.T) {
	dir := setupCoverageDir(t, []string{
		"src/auth/login.ts",
		"src/api/handler.ts",
	})

	cm := CodeMap{}

	result, err := Coverage(cm, dir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Mapped != 0 {
		t.Errorf("mapped: got %d, want 0", result.Mapped)
	}
	if result.Unmapped != 2 {
		t.Errorf("unmapped: got %d, want 2", result.Unmapped)
	}
	if result.CoveragePct != 0 {
		t.Errorf("coverage: got %.1f%%, want 0%%", result.CoveragePct)
	}
	sort.Strings(result.UnmappedFiles)
	if len(result.UnmappedFiles) != 2 {
		t.Errorf("unmapped files count: got %d, want 2", len(result.UnmappedFiles))
	}
}

func TestCoverage_MultipleExcludePatterns(t *testing.T) {
	dir := setupCoverageDir(t, []string{
		"src/auth/login.ts",
		"src/auth/login.test.ts",
		"src/auth/login.spec.ts",
		"types/auth.d.ts",
	})

	cm := CodeMap{
		"c3-101": {"src/auth/login.ts"},
		"_exclude": {
			"**/*.test.ts",
			"**/*.spec.ts",
			"**/*.d.ts",
		},
	}

	result, err := Coverage(cm, dir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Mapped != 1 {
		t.Errorf("mapped: got %d, want 1", result.Mapped)
	}
	if result.Excluded != 3 {
		t.Errorf("excluded: got %d, want 3", result.Excluded)
	}
	if result.Unmapped != 0 {
		t.Errorf("unmapped: got %d, want 0", result.Unmapped)
	}
}

func TestCoverage_NoFiles(t *testing.T) {
	dir := t.TempDir()
	cm := CodeMap{"c3-101": {"src/**"}}

	result, err := Coverage(cm, dir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Total != 0 {
		t.Errorf("total: got %d, want 0", result.Total)
	}
	if result.CoveragePct != 0 {
		t.Errorf("coverage: got %.1f%%, want 0%%", result.CoveragePct)
	}
}

func TestExclude_ReturnsPatterns(t *testing.T) {
	cm := CodeMap{
		"c3-101":   {"src/**"},
		"_exclude": {"**/*.test.ts", "dist/**"},
	}

	patterns := Exclude(cm)
	if len(patterns) != 2 {
		t.Errorf("expected 2 exclude patterns, got %d", len(patterns))
	}
}

func TestExclude_NilWhenMissing(t *testing.T) {
	cm := CodeMap{"c3-101": {"src/**"}}

	patterns := Exclude(cm)
	if patterns != nil {
		t.Errorf("expected nil for missing _exclude, got %v", patterns)
	}
}

func TestListProjectFiles_SkipsC3Dir(t *testing.T) {
	// .c3/ files should be excluded from file listing (not source code)
	dir := setupCoverageDir(t, []string{
		".c3/README.md",
		".c3/c3-1-api/c3-101-auth.md",
		"src/auth/login.ts",
	})

	files, err := ListProjectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.HasPrefix(f, ".c3/") {
			t.Errorf(".c3/ files should be excluded, found: %s", f)
		}
	}
	if len(files) != 1 || files[0] != "src/auth/login.ts" {
		t.Errorf("expected [src/auth/login.ts], got %v", files)
	}
}

func TestMatch_BracketPaths(t *testing.T) {
	// Next.js/SvelteKit route params should match literally
	cm := CodeMap{
		"c3-101": {"src/app/[id]/page.tsx"},
		"c3-102": {"src/app/[...slug]/**/*.ts"},
	}

	// Exact bracket path
	ids := Match(cm, "src/app/[id]/page.tsx")
	if len(ids) != 1 || ids[0] != "c3-101" {
		t.Errorf("expected [c3-101] for bracket path, got %v", ids)
	}

	// Glob with bracket path segment
	ids = Match(cm, "src/app/[...slug]/nested/handler.ts")
	if len(ids) != 1 || ids[0] != "c3-102" {
		t.Errorf("expected [c3-102] for catch-all bracket path, got %v", ids)
	}

	// No match for unrelated path
	ids = Match(cm, "src/app/other/page.tsx")
	if len(ids) != 0 {
		t.Errorf("unrelated path should not match, got %v", ids)
	}
}

func TestMatch_BracketAndGlobMixed(t *testing.T) {
	// Pattern with both literal brackets and glob wildcards
	cm := CodeMap{
		"c3-101": {"src/app/[id]/**/*.tsx"},
	}

	ids := Match(cm, "src/app/[id]/edit/form.tsx")
	if len(ids) != 1 || ids[0] != "c3-101" {
		t.Errorf("expected [c3-101] for mixed bracket+glob, got %v", ids)
	}
}

func TestIsSkippedPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".c3/something", true},
		{".git/HEAD", true},
		{"node_modules/pkg/index.js", true},
		{"dist/bundle.js", true},
		{"src/main.ts", false},
		{"README.md", false},
	}
	for _, tt := range tests {
		if got := isSkippedPath(tt.path); got != tt.want {
			t.Errorf("isSkippedPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestListProjectFiles_Fallback(t *testing.T) {
	dir := t.TempDir()
	// Not a git repo, so it'll fall back to walk
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "app.ts"), []byte(""), 0644)
	// These should be skipped
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg.js"), []byte(""), 0644)

	files, err := ListProjectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if isSkippedPath(f) {
			t.Errorf("should not include skipped path: %s", f)
		}
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestMatch_TraditionalGlobStillWorks(t *testing.T) {
	// Traditional glob character classes should still work
	cm := CodeMap{
		"c3-101": {"src/**/*.ts"},
	}

	ids := Match(cm, "src/auth/login.ts")
	if len(ids) != 1 || ids[0] != "c3-101" {
		t.Errorf("traditional glob should still work, got %v", ids)
	}
}

func TestMatch_SkipsUnderscoreKeys(t *testing.T) {
	cm := CodeMap{
		"c3-101":   {"src/auth/**"},
		"_exclude": {"src/auth/**"},
	}

	ids := Match(cm, "src/auth/login.ts")
	if len(ids) != 1 || ids[0] != "c3-101" {
		t.Errorf("expected [c3-101], got %v", ids)
	}
}
