package codemap

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// CoverageResult holds code-map coverage statistics.
type CoverageResult struct {
	Total         int      `json:"total"`
	Mapped        int      `json:"mapped"`
	Excluded      int      `json:"excluded"`
	Unmapped      int      `json:"unmapped"`
	CoveragePct   float64  `json:"coverage_pct"`
	UnmappedFiles []string `json:"unmapped_files"`
}

// Coverage computes how many project files are mapped, excluded, or unmapped.
func Coverage(cm CodeMap, projectDir string) (*CoverageResult, error) {
	files, err := ListProjectFiles(projectDir)
	if err != nil {
		return nil, err
	}

	excludePatterns := Exclude(cm)

	var mapped, excluded, unmapped int
	unmappedFiles := []string{}

	for _, f := range files {
		ids := Match(cm, f)
		if len(ids) > 0 {
			mapped++
			continue
		}

		if matchesAny(f, excludePatterns) {
			excluded++
		} else {
			unmapped++
			unmappedFiles = append(unmappedFiles, f)
		}
	}

	total := len(files)
	pct := 0.0
	mappable := total - excluded
	if mappable > 0 {
		pct = float64(mapped) / float64(mappable) * 100
	}

	return &CoverageResult{
		Total:         total,
		Mapped:        mapped,
		Excluded:      excluded,
		Unmapped:      unmapped,
		CoveragePct:   pct,
		UnmappedFiles: unmappedFiles,
	}, nil
}

func matchesAny(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		ok, _ := matchGlob(pattern, filePath)
		if ok {
			return true
		}
	}
	return false
}

var skipPrefixes = []string{".c3/", ".git/", "node_modules/", "dist/"}

func isSkippedPath(p string) bool {
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// ListProjectFiles returns all tracked files in a project directory.
// Tries git ls-files first, falls back to filesystem walk.
// Excludes .c3/, .git/, node_modules/, and dist/ paths in both modes.
func ListProjectFiles(projectDir string) ([]string, error) {
	out, err := exec.Command("git", "-C", projectDir, "ls-files").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var files []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || isSkippedPath(filepath.ToSlash(l)) {
				continue
			}
			files = append(files, filepath.ToSlash(l))
		}
		sort.Strings(files)
		return files, nil
	}

	// Fallback: walk filesystem, skipping common non-source dirs.
	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "dist": true, ".c3": true,
	}

	var files []string
	err = filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(projectDir, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
