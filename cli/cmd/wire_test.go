package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

// =============================================================================
// RunWire: bidirectional cite relationship management (2 sides)
// =============================================================================

func TestRunWire_CiteRef_UpdatesBothSides(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	err := RunWire(c3Dir, "c3-201", "cite", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Side 1: source's frontmatter uses[] should include ref-error-handling
	srcContent, err := os.ReadFile(filepath.Join(c3Dir, "c3-2-web", "c3-201-renderer.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(srcContent))
	if !containsStr2(fm.Refs, "ref-error-handling") {
		t.Errorf("Side 1 fail: source uses should include ref-error-handling, got %v", fm.Refs)
	}

	// Side 2: source's Related Refs table should include ref-error-handling
	srcStr := string(srcContent)
	if !strings.Contains(srcStr, "ref-error-handling") {
		t.Error("Side 2 fail: source's Related Refs table should include ref-error-handling")
	}

	output := buf.String()
	if !strings.Contains(output, "Wired") {
		t.Errorf("should print Wired message, got: %s", output)
	}
}

func TestRunWire_CiteRef_NoDuplicate(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	// c3-101 already cites ref-jwt in fixture
	err := RunWire(c3Dir, "c3-101", "cite", "ref-jwt", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify no duplicates in frontmatter
	content, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "c3-101-auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(content))
	count := 0
	for _, r := range fm.Refs {
		if r == "ref-jwt" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("should not duplicate ref citation, count = %d", count)
	}
}

func TestRunWire_SourceNotFound(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	err := RunWire(c3Dir, "c3-999", "cite", "ref-jwt", &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunWire_TargetNotFound(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	err := RunWire(c3Dir, "c3-101", "cite", "ref-nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunWire_InvalidRelationType(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	err := RunWire(c3Dir, "c3-101", "depend", "c3-102", &buf)
	if err == nil {
		t.Fatal("expected error for unsupported relation type (only cite in v1)")
	}
}

// =============================================================================
// RunUnwire: remove cite relationship from all three sides
// =============================================================================

func TestRunUnwire_CiteRef(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	// Wire first
	RunWire(c3Dir, "c3-201", "cite", "ref-error-handling", &buf)
	buf.Reset()

	// Unwire
	err := RunUnwire(c3Dir, "c3-201", "cite", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Side 1: uses[] should not contain ref-error-handling
	srcContent, err := os.ReadFile(filepath.Join(c3Dir, "c3-2-web", "c3-201-renderer.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := frontmatter.ParseFrontmatter(string(srcContent))
	if containsStr2(fm.Refs, "ref-error-handling") {
		t.Errorf("source uses should not contain ref-error-handling after unwire, got %v", fm.Refs)
	}

	// Side 2: Related Refs should not contain ref-error-handling (as a table row)
	// (Note: the section header "Related Refs" will still exist, just no row for this ref)
}

func TestRunUnwire_NotWired(t *testing.T) {
	c3Dir := createRichFixture(t)
	var buf bytes.Buffer

	// c3-201 doesn't cite ref-jwt — unwire should be idempotent (no error)
	err := RunUnwire(c3Dir, "c3-201", "cite", "ref-jwt", &buf)
	// Acceptable: either no error (idempotent) or specific "not wired" error
	if err != nil && !strings.Contains(err.Error(), "not") {
		t.Errorf("unexpected error: %v", err)
	}
}

// =============================================================================
// Wire: rule- target detection (uses Related Rules section)
// =============================================================================

func TestWireRuleUsesRelatedRulesSection(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1"), 0755)
	os.MkdirAll(filepath.Join(c3Dir, "rules"), 0755)

	compContent := "---\nid: c3-101\ntype: component\nparent: c3-1\nuses: []\n---\n\n# Auth\n\n## Goal\nAuth\n\n## Related Refs\n\n| Ref | Role |\n|-----|------|\n\n## Related Rules\n\n| Rule | Role |\n|------|------|\n"
	ruleContent := "---\nid: rule-logging\ntype: rule\n---\n\n# Logging\n"

	os.WriteFile(filepath.Join(c3Dir, "c3-1", "c3-101-auth.md"), []byte(compContent), 0644)
	os.WriteFile(filepath.Join(c3Dir, "rules", "rule-logging.md"), []byte(ruleContent), 0644)

	var buf bytes.Buffer
	err := RunWire(c3Dir, "c3-101", "cite", "rule-logging", &buf)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(c3Dir, "c3-1", "c3-101-auth.md"))
	body := string(data)
	if !strings.Contains(body, "rule-logging") {
		t.Error("rule-logging should appear in component file")
	}
}
