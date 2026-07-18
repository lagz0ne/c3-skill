package walker

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

// ParseWarning records a .md file that has YAML frontmatter delimiters but failed to parse.
type ParseWarning struct {
	Path string // relative to .c3/
}

// WalkResult holds both successfully parsed docs and files that failed to parse.
type WalkResult struct {
	Docs     []frontmatter.ParsedDoc
	Warnings []ParseWarning
}

// WalkC3Docs recursively walks a .c3/ directory and parses all .md files.
func WalkC3Docs(c3Dir string) ([]frontmatter.ParsedDoc, error) {
	result, err := WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return nil, err
	}
	return result.Docs, nil
}

// WalkC3DocsWithWarnings is like WalkC3Docs but also returns parse warnings.
func WalkC3DocsWithWarnings(c3Dir string) (*WalkResult, error) {
	result := &WalkResult{}

	err := filepath.Walk(c3Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "_index" {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		raw := string(content)
		fm, body := frontmatter.ParseFrontmatter(raw)
		rel, _ := filepath.Rel(c3Dir, path)
		if fm != nil {
			result.Docs = append(result.Docs, frontmatter.ParsedDoc{
				Frontmatter: fm,
				Body:        body,
				Path:        rel,
			})
		} else if strings.HasPrefix(raw, "---\n") {
			result.Warnings = append(result.Warnings, ParseWarning{Path: rel})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

var slugPattern = regexp.MustCompile(`^(c3-\d+-|c3-\d+|ref-|rule-|adr-\d+-|README)`)

// SlugFromPath derives a slug from a file path by stripping the ID prefix.
// For README.md files (containers), the slug is derived from the parent directory name.
func SlugFromPath(filePath string) string {
	base := strings.TrimSuffix(filepath.Base(filePath), ".md")
	if base == "README" {
		dir := filepath.Dir(filePath)
		if dir == "." || dir == "" || dir == "/" {
			return "" // top-level context README
		}
		dirBase := filepath.Base(dir)
		return slugPattern.ReplaceAllString(dirBase, "")
	}
	return slugPattern.ReplaceAllString(base, "")
}

// SlugFromEntityPath derives the slug relative to the entity's declared ID.
// This keeps cacheless imports idempotent for legacy or custom IDs that the
// standard numeric-ID pattern cannot recognize.
func SlugFromEntityPath(filePath, entityID string) string {
	segment := strings.TrimSuffix(filepath.Base(filePath), ".md")
	if segment == "README" {
		dir := filepath.Dir(filePath)
		if dir == "." || dir == "" || dir == "/" {
			return ""
		}
		segment = filepath.Base(dir)
	}

	if segment == entityID {
		// Canonical ADR ids contain their date and slug in the id itself. Keep
		// the suffix so export does not collapse same-date ADRs to one path.
		if slug := SlugFromPath(filePath); strings.HasPrefix(entityID, "adr-") && slug != segment {
			return slug
		}
		return ""
	}
	if prefix := entityID + "-"; strings.HasPrefix(segment, prefix) {
		return strings.TrimPrefix(segment, prefix)
	}
	return SlugFromPath(filePath)
}
