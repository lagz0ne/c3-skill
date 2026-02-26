package config

import (
	"os"
	"path/filepath"
)

// FindC3Dir walks up from startDir looking for a .c3/ directory.
// Returns the absolute path to .c3/ or empty string if not found.
func FindC3Dir(startDir string) string {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, ".c3")
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// ResolveC3Dir returns the override path if set, otherwise auto-discovers .c3/ from startDir.
func ResolveC3Dir(startDir string, override string) string {
	if override != "" {
		return override
	}
	return FindC3Dir(startDir)
}

// ProjectDir returns the project root (parent of .c3/ directory).
func ProjectDir(c3Dir string) string {
	return filepath.Dir(c3Dir)
}
