package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// LegacyCheckOptions holds parameters for file-based (pre-migration) validation.
type LegacyCheckOptions struct {
	Docs          []frontmatter.ParsedDoc
	Graph         *walker.C3Graph
	JSON          bool
	ProjectDir    string
	C3Dir         string
	ParseWarnings []walker.ParseWarning
	IncludeADR    bool
	Fix           bool
}

// RunLegacyCheck validates file-based .c3/ docs using the walker.
// Used pre-migration to surface malformed docs that need LLM assistance.
func RunLegacyCheck(opts LegacyCheckOptions, w io.Writer) error {
	var issues []Issue

	// Phase 1: Parse warnings (files that failed to parse)
	for _, pw := range opts.ParseWarnings {
		issues = append(issues, Issue{
			Severity: "error",
			Entity:   pw.Path,
			Message:  "failed to parse YAML frontmatter",
			Hint:     "Fix YAML frontmatter before migration",
		})
	}

	// Phase 2: Validate each entity
	titleMap := buildLegacyTitleMap(opts.Graph)
	for _, entity := range opts.Graph.All() {
		// Skip ADRs unless requested
		if entity.Type == frontmatter.DocADR && !opts.IncludeADR {
			continue
		}

		// Check required frontmatter fields
		if entity.Frontmatter.Title == "" {
			issues = append(issues, Issue{
				Severity: "error",
				Entity:   entity.ID,
				Message:  "missing title in frontmatter",
				Hint:     "Add 'title: <name>' to frontmatter",
			})
		}

		// Check schema sections
		entityType := entity.Type.String()
		sections := schema.ForType(entityType)
		bodyParts := markdown.ParseSections(entity.Body)
		bodyMap := make(map[string]bool)
		for _, s := range bodyParts {
			bodyMap[strings.ToLower(s.Name)] = true
		}
		for _, sec := range sections {
			if sec.Required {
				if !bodyMap[strings.ToLower(sec.Name)] {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("missing required section: %s", sec.Name),
						Hint:     fmt.Sprintf("Add '## %s' section to body", sec.Name),
					})
				}
			}
		}

		// Check relationship targets exist
		for _, rel := range entity.Relationships {
			if opts.Graph.Get(rel) == nil {
				suggestion := suggestLegacyFix(rel, titleMap)
				hint := "Remove or fix this reference"
				if suggestion != "" {
					hint = fmt.Sprintf("Did you mean '%s'?", suggestion)
				}
				issues = append(issues, Issue{
					Severity: "error",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("broken reference: %s", rel),
					Hint:     hint,
				})
			}
		}

		// Check origin references (for rules)
		if entity.Type == frontmatter.DocRule {
			for _, orig := range entity.Frontmatter.Origin {
				target := frontmatter.StripAnchor(orig)
				if opts.Graph.Get(target) == nil {
					issues = append(issues, Issue{
						Severity: "error",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("origin reference %q not found", target),
						Hint:     "origin should reference an existing ref or ADR entity",
					})
				}
			}
		}

		// Check parent exists (for components)
		if entity.Frontmatter.Parent != "" && opts.Graph.Get(entity.Frontmatter.Parent) == nil {
			issues = append(issues, Issue{
				Severity: "error",
				Entity:   entity.ID,
				Message:  fmt.Sprintf("parent %s not found", entity.Frontmatter.Parent),
			})
		}
	}

	// Output
	result := CheckResult{
		Total:  len(issues),
		Issues: issues,
	}

	if opts.JSON {
		return writeJSON(w, result)
	}

	if len(issues) == 0 {
		fmt.Fprintln(w, "✓ All docs valid — ready for migration (c3x migrate)")
		return nil
	}

	errors := 0
	warnings := 0
	for _, issue := range issues {
		marker := "⚠"
		if issue.Severity == "error" {
			marker = "✗"
			errors++
		} else {
			warnings++
		}
		entity := issue.Entity
		fmt.Fprintf(w, "  %s %s: %s\n", marker, entity, issue.Message)
		if issue.Hint != "" {
			fmt.Fprintf(w, "    → %s\n", issue.Hint)
		}
	}
	fmt.Fprintf(w, "\n%d error(s), %d warning(s)\n", errors, warnings)
	if errors > 0 {
		fmt.Fprintln(w, "\nFix errors before running c3x migrate.")
		fmt.Fprintln(w, "Use /c3 in Claude Code for LLM-assisted repair.")
	}
	return nil
}

func buildLegacyTitleMap(graph *walker.C3Graph) map[string]string {
	m := make(map[string]string)
	for _, e := range graph.All() {
		lower := strings.ToLower(e.Title)
		if existing, exists := m[lower]; exists {
			if existing != e.ID {
				m[lower] = ""
			}
		} else {
			m[lower] = e.ID
		}
	}
	return m
}

func suggestLegacyFix(badRef string, titleMap map[string]string) string {
	lower := strings.ToLower(badRef)
	// Try exact title match
	if id, ok := titleMap[lower]; ok && id != "" {
		return id
	}
	// Try prefix match
	for title, id := range titleMap {
		if id != "" && strings.HasPrefix(title, lower) {
			return id
		}
	}
	return ""
}
