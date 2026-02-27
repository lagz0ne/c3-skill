package codemap

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

// CodeMap maps component IDs to their source file paths.
type CodeMap map[string][]string

// ParseCodeMap reads and parses a code-map.yaml file.
// Returns an empty map if the file does not exist.
func ParseCodeMap(path string) (CodeMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return CodeMap{}, nil
		}
		return nil, err
	}

	cm := CodeMap{}
	if len(data) == 0 {
		return cm, nil
	}

	if err := yaml.Unmarshal(data, &cm); err != nil {
		return nil, err
	}

	if cm == nil {
		return CodeMap{}, nil
	}

	return cm, nil
}

// Exclude returns the exclude patterns from a code-map (_exclude key).
func Exclude(cm CodeMap) []string {
	return cm["_exclude"]
}

// matchGlob wraps doublestar.Match with fallback for literal brackets.
// Framework route params like [id], [...slug] (Next.js, SvelteKit) contain
// brackets that doublestar interprets as glob character classes.
// Tries the pattern as-is first, then retries with brackets escaped.
func matchGlob(pattern, name string) (bool, error) {
	matched, err := doublestar.Match(pattern, name)
	if matched || !strings.ContainsAny(pattern, "[]") {
		return matched, err
	}
	escaped := strings.NewReplacer("[", "\\[", "]", "\\]").Replace(pattern)
	return doublestar.Match(escaped, name)
}

// GlobFiles wraps doublestar.Glob with fallback for literal brackets.
func GlobFiles(fsys fs.FS, pattern string) ([]string, error) {
	matches, err := doublestar.Glob(fsys, pattern)
	if (err != nil || len(matches) == 0) && strings.ContainsAny(pattern, "[]") {
		escaped := strings.NewReplacer("[", "\\[", "]", "\\]").Replace(pattern)
		escapedMatches, escapedErr := doublestar.Glob(fsys, escaped)
		if len(escapedMatches) > 0 {
			return escapedMatches, escapedErr
		}
	}
	return matches, err
}

// Match returns sorted component IDs whose file patterns match filePath.
// filePath should be relative to the project root.
// Patterns support glob syntax including **.
// Literal brackets (e.g. Next.js [id] routes) are handled automatically.
// Keys prefixed with _ (e.g. _exclude) are skipped.
func Match(cm CodeMap, filePath string) []string {
	filePath = filepath.ToSlash(filepath.Clean(filePath))

	var matches []string
	for id, patterns := range cm {
		if strings.HasPrefix(id, "_") {
			continue
		}
		for _, pattern := range patterns {
			pattern = filepath.ToSlash(strings.TrimSpace(pattern))
			if pattern == "" {
				continue
			}
			matched, _ := matchGlob(pattern, filePath)
			if matched {
				matches = append(matches, id)
				break
			}
		}
	}
	sort.Strings(matches)
	return matches
}
