package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunEval_AgentTOONIncludesHelpHints(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	codePath := filepath.Join(projectDir, "src", "auth.go")
	if err := os.MkdirAll(filepath.Dir(codePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codePath, []byte("package auth\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-101.yaml"), []byte("fact: c3-101\ncode:\n  - src/auth.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help[",
		"c3x eval <fact-id>",
		"c3x lookup <file-or-glob>",
	)
}
