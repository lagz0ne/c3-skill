package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	c3HookStart       = "# BEGIN C3 MANAGED BLOCK"
	c3HookEnd         = "# END C3 MANAGED BLOCK"
	c3AttributesStart = "# BEGIN C3 MANAGED ATTRIBUTES"
	c3AttributesEnd   = "# END C3 MANAGED ATTRIBUTES"
	c3GitignoreStart  = "# BEGIN C3 MANAGED IGNORE"
	c3GitignoreEnd    = "# END C3 MANAGED IGNORE"
	preCommitHookName = "pre-commit"
)

// RunGitInstall installs thin Git guardrails for the canonical C3 workflow.
func RunGitInstall(projectDir, c3Dir string, w io.Writer) error {
	gitDir, err := resolveGitDir(projectDir)
	if err != nil {
		return err
	}

	hookPath := filepath.Join(gitDir, "hooks", preCommitHookName)
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		return fmt.Errorf("git install: create hooks dir: %w", err)
	}
	hookContent := buildPreCommitHook()
	if err := upsertManagedBlockFile(hookPath, c3HookStart, c3HookEnd, hookContent, true); err != nil {
		return fmt.Errorf("git install: write pre-commit hook: %w", err)
	}
	fmt.Fprintf(w, "Installed %s\n", hookPath)

	attrPath := filepath.Join(projectDir, ".gitattributes")
	attrBlock := strings.Join([]string{
		c3AttributesStart,
		".c3/c3.db binary linguist-generated",
		c3AttributesEnd,
		"",
	}, "\n")
	if err := upsertManagedBlockFile(attrPath, c3AttributesStart, c3AttributesEnd, attrBlock, false); err != nil {
		return fmt.Errorf("git install: write .gitattributes: %w", err)
	}
	fmt.Fprintf(w, "Updated %s\n", attrPath)

	ignorePath := filepath.Join(c3Dir, ".gitignore")
	ignoreBlock := strings.Join([]string{
		c3GitignoreStart,
		"c3.db",
		"c3.db.bak-*",
		c3GitignoreEnd,
		"",
	}, "\n")
	if err := upsertManagedBlockFile(ignorePath, c3GitignoreStart, c3GitignoreEnd, ignoreBlock, false); err != nil {
		return fmt.Errorf("git install: write .c3/.gitignore: %w", err)
	}
	fmt.Fprintf(w, "Updated %s\n", ignorePath)

	fmt.Fprintf(w, "Installed C3 Git guardrails for %s\n", c3Dir)
	return nil
}

func resolveGitDir(projectDir string) (string, error) {
	gitPath := filepath.Join(projectDir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return "", fmt.Errorf("git install: .git not found in %s", projectDir)
	}
	if info.IsDir() {
		return gitPath, nil
	}

	data, err := os.ReadFile(gitPath)
	if err != nil {
		return "", fmt.Errorf("git install: read .git file: %w", err)
	}
	line := strings.TrimSpace(string(data))
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("git install: unsupported .git file format")
	}
	target := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if !filepath.IsAbs(target) {
		target = filepath.Join(projectDir, target)
	}
	return filepath.Clean(target), nil
}

func buildPreCommitHook() string {
	return strings.Join([]string{
		c3HookStart,
		`repo_root="$(git rev-parse --show-toplevel)"`,
		`if git diff --cached --name-only -- .c3/c3.db | grep -q .; then`,
		`  echo "c3: .c3/c3.db is local cache only; unstage it." >&2`,
		`  exit 1`,
		`fi`,
		`c3x verify --c3-dir "$repo_root/.c3"`,
		`if ! git diff --quiet -- .c3; then`,
		`  echo "c3: canonical .c3/ files changed; review and stage them before committing." >&2`,
		`  exit 1`,
		`fi`,
		c3HookEnd,
		"",
	}, "\n")
}

func upsertManagedBlockFile(path, startMarker, endMarker, managedBlock string, executable bool) error {
	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	} else if !os.IsNotExist(err) {
		return err
	}

	content := existing
	if start := strings.Index(content, startMarker); start >= 0 {
		end := strings.Index(content[start:], endMarker)
		if end < 0 {
			return fmt.Errorf("managed block start found without end marker in %s", path)
		}
		end += start + len(endMarker)
		content = content[:start] + managedBlock + content[end:]
	} else {
		if strings.TrimSpace(content) == "" {
			if executable {
				content = "#!/bin/sh\nset -eu\n\n" + managedBlock
			} else {
				content = managedBlock
			}
		} else {
			if executable && !strings.HasPrefix(content, "#!") {
				content = "#!/bin/sh\nset -eu\n\n" + strings.TrimLeft(content, "\n")
			}
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n" + managedBlock
		}
	}

	mode := os.FileMode(0644)
	if executable {
		mode = 0755
	}
	return os.WriteFile(path, []byte(content), mode)
}
