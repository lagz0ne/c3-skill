package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_InvalidID(t *testing.T) {
	cm := CodeMap{"c3-999": {"src/foo.ts"}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, "")
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-999" {
			found = true
		}
	}
	if !found {
		t.Error("should flag unknown component ID c3-999")
	}
}

func TestValidate_MissingFile(t *testing.T) {
	dir := t.TempDir()
	cm := CodeMap{"c3-101": {"src/nonexistent.ts"}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, dir)
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-101" {
			found = true
		}
	}
	if !found {
		t.Error("should flag nonexistent file")
	}
}

func TestValidate_FileExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "foo.ts"), []byte(""), 0644)

	cm := CodeMap{"c3-101": {"src/foo.ts"}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, dir)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestValidate_RefID(t *testing.T) {
	// refs are valid in code-map alongside components — no warning expected
	cm := CodeMap{"ref-jwt": {"src/foo.ts"}}
	entities := map[string]string{"ref-jwt": "ref"}

	issues := Validate(cm, entities, "")
	for _, issue := range issues {
		if issue.Entity == "ref-jwt" && issue.Severity == "warning" {
			t.Errorf("ref-jwt should not be flagged: %s", issue.Message)
		}
	}
}

// =============================================================================
// RED tests for Codex-identified issues
// =============================================================================

// Issue 1: Container c3-100 matches c3-\d{3,} regex but is a container, not a component.
// The fix uses entity type info (map[string]string) instead of regex.
func TestValidate_ContainerID(t *testing.T) {
	cm := CodeMap{"c3-100": {"src/container-stuff.ts"}}
	entities := map[string]string{"c3-100": "container"}

	issues := Validate(cm, entities, "")

	// Should flag c3-100 as non-component (it's a container)
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-100" && issue.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("should flag container ID c3-100 as non-component in code-map")
	}
}

// Issue 3: Empty path in file list should be flagged.
func TestValidate_EmptyPath(t *testing.T) {
	cm := CodeMap{"c3-101": {""}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, "")
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-101" && issue.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("should flag empty string path as warning")
	}
}

// Issue 3: Directory path (not a file) should be flagged.
func TestValidate_DirectoryPath(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	cm := CodeMap{"c3-101": {"src"}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, dir)
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-101" && issue.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("should flag directory path as warning (code-map entries should be files, not directories)")
	}
}

// Issue 4: Absolute path should be flagged.
func TestValidate_AbsolutePath(t *testing.T) {
	cm := CodeMap{"c3-101": {"/etc/passwd"}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, "")
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-101" && issue.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("should flag absolute path /etc/passwd as warning")
	}
}

// Issue 4: Path traversal with ../ should be flagged.
func TestValidate_DotDotPath(t *testing.T) {
	cm := CodeMap{"c3-101": {"../../../etc/passwd"}}
	entities := map[string]string{"c3-101": "component"}

	issues := Validate(cm, entities, "")
	found := false
	for _, issue := range issues {
		if issue.Entity == "c3-101" && issue.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("should flag path traversal ../../../etc/passwd as warning")
	}
}
