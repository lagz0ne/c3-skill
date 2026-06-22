package codemap

import (
	"io/fs"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// CodeMap maps fact IDs to their source file globs. The fact→code binding lives in
// the eval specs (.c3/eval/<fact>.yaml `code:` field); this type is the in-memory
// shape eval and lookup share.
type CodeMap map[string][]string

// IsGlobPattern reports whether p contains glob metacharacters.
// Brackets are excluded — [id] style paths (Next.js, SvelteKit) are literal.
func IsGlobPattern(p string) bool {
	return strings.ContainsAny(p, "*?")
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
