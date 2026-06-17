package schema

import (
	"os"
	"path/filepath"
	"testing"
)

// A project-local canvas that declares a status set IS a change-doc, even though
// it is not a built-in type. The built-in-only IsChangeDoc is blind to it; the
// project-aware IsChangeDocDir must recognize it.
func TestIsChangeDocDir_RecognizesProjectCanvas(t *testing.T) {
	dir := t.TempDir()
	canvasDir := filepath.Join(dir, CanvasesDir)
	if err := os.MkdirAll(canvasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canvasDir, "demo-change.md"), []byte(declaringCanvasDoc(t)), 0o644); err != nil {
		t.Fatal(err)
	}

	if IsChangeDoc("demo-change") {
		t.Fatal("built-in IsChangeDoc should not recognize a project-only canvas")
	}
	if !IsChangeDocDir(dir, "demo-change") {
		t.Fatal("IsChangeDocDir must recognize a project canvas that declares a status set")
	}
}

// A built-in change-doc type is still recognized through the dir-aware path.
func TestIsChangeDocDir_FallsBackToBuiltIn(t *testing.T) {
	if !IsChangeDocDir(t.TempDir(), "adr") {
		t.Fatal("IsChangeDocDir must fall back to the built-in adr change-doc")
	}
	if IsChangeDocDir(t.TempDir(), "component") {
		t.Fatal("a fact canvas (component) is not a change-doc")
	}
}
