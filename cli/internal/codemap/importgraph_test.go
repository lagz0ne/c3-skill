package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, root, rel, contents string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte(contents), 0644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
}

func TestDeriveCallers_FindsImports(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "web/renderer/html.ts", "export function render() {}\n")
	writeFile(t, dir, "api/caller.ts",
		"import { render } from '../web/renderer/html';\nrender();\n")
	writeFile(t, dir, "api/unrelated.ts", "console.log('hi');\n")

	callers, err := DeriveCallers(dir, []string{"web/renderer/html.ts"})
	if err != nil {
		t.Fatalf("DeriveCallers: %v", err)
	}
	if !contains(callers, "api/caller.ts") {
		t.Errorf("expected api/caller.ts in callers, got %v", callers)
	}
	if contains(callers, "api/unrelated.ts") {
		t.Errorf("unrelated file should not be a caller, got %v", callers)
	}
	if contains(callers, "web/renderer/html.ts") {
		t.Errorf("target file should not be its own caller, got %v", callers)
	}
}

func TestDeriveCallers_EmptyTargets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.ts", "hello\n")
	callers, err := DeriveCallers(dir, nil)
	if err != nil {
		t.Fatalf("DeriveCallers: %v", err)
	}
	if len(callers) != 0 {
		t.Errorf("expected no callers for empty targets, got %v", callers)
	}
}

func TestDeriveCallers_MatchesExtensionlessImports(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/util/format.go", "package util\n")
	writeFile(t, dir, "cmd/main.go",
		"package main\nimport _ \"example.com/app/pkg/util/format\"\n")

	callers, err := DeriveCallers(dir, []string{"pkg/util/format.go"})
	if err != nil {
		t.Fatalf("DeriveCallers: %v", err)
	}
	if !contains(callers, "cmd/main.go") {
		t.Errorf("expected extensionless match to find cmd/main.go, got %v", callers)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
