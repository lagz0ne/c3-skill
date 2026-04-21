package codemap

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DeriveCallers returns project-relative file paths that appear to reference
// any of the given targetSources (grep-derived, heuristic). targetSources are
// themselves excluded from the result. Paths are forward-slash normalized.
//
// Heuristic: a file is a caller if it contains the target's relative path
// or its extensionless form as a substring. This is coarse enough to catch
// import statements across common languages (Go, TS/JS, Python, Rust, Java)
// without parsing each dialect.
func DeriveCallers(projectDir string, targetSources []string) ([]string, error) {
	if len(targetSources) == 0 {
		return nil, nil
	}

	targetSet := make(map[string]bool, len(targetSources))
	var tokens []string
	tokenSeen := map[string]bool{}
	for _, src := range targetSources {
		src = filepath.ToSlash(strings.TrimSpace(src))
		if src == "" {
			continue
		}
		targetSet[src] = true
		if !tokenSeen[src] {
			tokenSeen[src] = true
			tokens = append(tokens, src)
		}
		if ext := filepath.Ext(src); ext != "" {
			noExt := strings.TrimSuffix(src, ext)
			if !tokenSeen[noExt] {
				tokenSeen[noExt] = true
				tokens = append(tokens, noExt)
			}
		}
	}
	if len(tokens) == 0 {
		return nil, nil
	}

	files, err := ListProjectFiles(projectDir)
	if err != nil {
		return nil, err
	}

	const maxFileBytes = 2 * 1024 * 1024 // 2MB cap per file to bound cost
	var callers []string
	seen := map[string]bool{}
	for _, rel := range files {
		if targetSet[rel] {
			continue
		}
		if !isTextFileCandidate(rel) {
			continue
		}
		abs := filepath.Join(projectDir, rel)
		info, statErr := os.Stat(abs)
		if statErr != nil || !info.Mode().IsRegular() {
			continue
		}
		if info.Size() == 0 || info.Size() > maxFileBytes {
			continue
		}
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			continue
		}
		contents := string(data)
		for _, tok := range tokens {
			if strings.Contains(contents, tok) {
				if !seen[rel] {
					seen[rel] = true
					callers = append(callers, rel)
				}
				break
			}
		}
	}
	sort.Strings(callers)
	return callers, nil
}

var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".pdf": true,
	".zip": true, ".tar": true, ".gz": true, ".ico": true, ".woff": true,
	".woff2": true, ".ttf": true, ".otf": true, ".mp4": true, ".webm": true,
	".wasm": true, ".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".db": true, ".sqlite": true, ".bin": true,
}

func isTextFileCandidate(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	return !binaryExts[ext]
}
