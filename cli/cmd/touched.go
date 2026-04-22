package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ResolveTouchedTargets returns entity IDs affected by files changed since
// `since` (empty = uncommitted: staged + unstaged + untracked).
func ResolveTouchedTargets(projectDir, c3Dir, since string) ([]string, error) {
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	defer s.Close()
	return resolveTouchedTargetsWithStore(projectDir, c3Dir, since, s)
}

func resolveTouchedTargetsWithStore(projectDir, c3Dir, since string, s *store.Store) ([]string, error) {
	files, err := gitTouchedFiles(projectDir, since)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var targets []string
	addID := func(id string) {
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		targets = append(targets, id)
	}

	c3Rel, _ := filepath.Rel(projectDir, c3Dir)
	c3Rel = filepath.ToSlash(c3Rel)
	if c3Rel == "" {
		c3Rel = ".c3"
	}

	for _, f := range files {
		f = filepath.ToSlash(f)
		if strings.HasPrefix(f, c3Rel+"/") && strings.HasSuffix(f, ".md") {
			abs := filepath.Join(projectDir, f)
			if data, rerr := os.ReadFile(abs); rerr == nil {
				if fm, _ := frontmatter.ParseFrontmatter(string(data)); fm != nil && fm.ID != "" {
					addID(fm.ID)
					continue
				}
			}
		}
		ids, _ := s.LookupByFile(f)
		for _, id := range ids {
			addID(id)
		}
	}
	sort.Strings(targets)
	return targets, nil
}

// gitTouchedFiles returns files changed since `since`; empty since means
// uncommitted (staged + unstaged + untracked).
func gitTouchedFiles(projectDir, since string) ([]string, error) {
	cmds := [][]string{}
	if since == "" {
		cmds = append(cmds,
			[]string{"diff", "--name-only", "HEAD"},
			[]string{"ls-files", "--others", "--exclude-standard"},
		)
	} else {
		cmds = append(cmds,
			[]string{"diff", "--name-only", since + "..HEAD"},
			[]string{"diff", "--name-only"},
		)
	}

	seen := map[string]bool{}
	var all []string
	for _, args := range cmds {
		full := append([]string{"-C", projectDir}, args...)
		out, err := exec.Command("git", full...).Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			seen[line] = true
			all = append(all, line)
		}
	}
	return all, nil
}
