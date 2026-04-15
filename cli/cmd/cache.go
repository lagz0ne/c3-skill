package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func RunCacheClear(c3Dir string, w io.Writer) error {
	patterns := []string{
		filepath.Join(c3Dir, "c3.db*"),
		filepath.Join(c3Dir, ".c3.import.tmp.db*"),
	}
	seen := map[string]bool{}
	var paths []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("error: invalid cache clear pattern %q: %w", pattern, err)
		}
		for _, match := range matches {
			if seen[match] {
				continue
			}
			seen[match] = true
			paths = append(paths, match)
		}
	}
	sort.Strings(paths)

	deleted := 0
	for _, path := range paths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error: inspecting cache file %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("error: removing cache file %s: %w", path, err)
		}
		deleted++
		rel, err := filepath.Rel(c3Dir, path)
		if err != nil {
			rel = path
		}
		fmt.Fprintf(w, "removed %s\n", rel)
	}
	if deleted == 0 {
		fmt.Fprintln(w, "No local C3 cache files found.")
		return nil
	}
	fmt.Fprintf(w, "Cleared %d local C3 cache file(s).\n", deleted)
	return nil
}
