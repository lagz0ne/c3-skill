package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGitInstall_DirectoryGitDir(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0755); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunGitInstall(projectDir, c3Dir, &buf); err != nil {
		t.Fatal(err)
	}

	hookPath := filepath.Join(projectDir, ".git", "hooks", preCommitHookName)
	hook, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	hookStr := string(hook)
	for _, want := range []string{
		"#!/bin/sh",
		c3HookStart,
		".c3/c3.db is local cache only",
		"c3x verify",
		`git diff --quiet -- .c3 ':(exclude).c3/c3.db' ':(exclude).c3/c3.db-*'`,
		"review and stage",
		c3HookEnd,
	} {
		if !strings.Contains(hookStr, want) {
			t.Fatalf("hook missing %q:\n%s", want, hookStr)
		}
	}

	attr, err := os.ReadFile(filepath.Join(projectDir, ".gitattributes"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(attr), ".c3/c3.db binary linguist-generated") {
		t.Fatalf("missing c3 db attribute:\n%s", string(attr))
	}

	ignore, err := os.ReadFile(filepath.Join(c3Dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ignore), "c3.db.bak-*") {
		t.Fatalf("missing backup ignore:\n%s", string(ignore))
	}
	if !strings.Contains(string(ignore), "c3.db") {
		t.Fatalf("missing db ignore:\n%s", string(ignore))
	}
}

func TestRunGitInstall_GitFileWorktree(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	realGitDir := filepath.Join(t.TempDir(), "gitdir")
	if err := os.MkdirAll(filepath.Join(realGitDir, "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".git"), []byte("gitdir: "+realGitDir+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RunGitInstall(projectDir, c3Dir, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(realGitDir, "hooks", preCommitHookName)); err != nil {
		t.Fatal(err)
	}
}

func TestRunGitInstall_PreservesExistingContentAndReplacesManagedBlock(t *testing.T) {
	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		t.Fatal(err)
	}
	hooksDir := filepath.Join(projectDir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(hooksDir, preCommitHookName)
	original := "#!/bin/sh\nset -eu\n\necho custom\n\n" + c3HookStart + "\nold\n" + c3HookEnd + "\n"
	if err := os.WriteFile(hookPath, []byte(original), 0755); err != nil {
		t.Fatal(err)
	}

	if err := RunGitInstall(projectDir, c3Dir, io.Discard); err != nil {
		t.Fatal(err)
	}

	hook, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	hookStr := string(hook)
	if !strings.Contains(hookStr, "echo custom") {
		t.Fatalf("custom hook content lost:\n%s", hookStr)
	}
	if strings.Contains(hookStr, "\nold\n") {
		t.Fatalf("old managed block was not replaced:\n%s", hookStr)
	}
	if strings.Count(hookStr, c3HookStart) != 1 {
		t.Fatalf("expected one managed block:\n%s", hookStr)
	}
}

func TestBuildPreCommitHook_IgnoresDisposableCacheDiff(t *testing.T) {
	hook := buildPreCommitHook()
	if strings.Contains(hook, "git diff --quiet -- .c3\n") {
		t.Fatalf("hook checks all .c3 and will block on c3.db cache diffs:\n%s", hook)
	}
	for _, want := range []string{
		`:(exclude).c3/c3.db`,
		`:(exclude).c3/c3.db-*`,
		`:(exclude).c3/*.tmp.db`,
	} {
		if !strings.Contains(hook, want) {
			t.Fatalf("hook missing cache exclude %q:\n%s", want, hook)
		}
	}
}
