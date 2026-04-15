package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// MigrateGap represents a single quality gap found in an entity.
type MigrateGap struct {
	Field       string `json:"field"`
	Status      string `json:"status"`
	BodyExcerpt string `json:"body_excerpt,omitempty"`
	Hint        string `json:"hint"`
}

// MigrateEntityReport is the dry-run analysis for one entity.
type MigrateEntityReport struct {
	ID   string       `json:"id"`
	File string       `json:"file"`
	Type string       `json:"type"`
	Gaps []MigrateGap `json:"gaps"`
}

// MigrateCodeMapIssue is a code-map validation issue.
type MigrateCodeMapIssue struct {
	EntityID string `json:"entity_id"`
	Status   string `json:"status"` // "unknown_entity", "no_file_matches"
	Pattern  string `json:"pattern,omitempty"`
	Hint     string `json:"hint"`
}

// MigrateDryRunResult is the full dry-run output.
type MigrateDryRunResult struct {
	Total         int                   `json:"total"`
	WithGaps      int                   `json:"with_gaps"`
	Clean         int                   `json:"clean"`
	TotalGaps     int                   `json:"total_gaps"`
	Entities      []MigrateEntityReport `json:"entities"`
	CodeMapIssues []MigrateCodeMapIssue `json:"code_map_issues,omitempty"`
}

// RunMigrateDryRun walks .c3/ files and reports quality gaps without migrating.
func RunMigrateDryRun(c3Dir string, jsonOut bool, w io.Writer) error {
	result, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return fmt.Errorf("error: walking %s: %w", c3Dir, err)
	}

	if len(result.Warnings) > 0 {
		for _, warn := range result.Warnings {
			fmt.Fprintf(w, "warning: skipping %s (failed to parse frontmatter)\n", warn.Path)
		}
	}

	if len(result.Docs) == 0 {
		return fmt.Errorf("error: no documents found in %s", c3Dir)
	}

	graph := walker.BuildGraph(result.Docs)

	var report MigrateDryRunResult
	report.Total = len(result.Docs)

	for _, doc := range result.Docs {
		fm := doc.Frontmatter
		dt := frontmatter.ClassifyDoc(fm)
		storeType := docTypeToStoreType(dt)
		if storeType == "" {
			continue
		}

		gaps := analyzeGaps(fm, doc.Body, storeType)
		gaps = append(gaps, analyzeIntegrity(fm, graph)...)

		if len(gaps) > 0 {
			report.Entities = append(report.Entities, MigrateEntityReport{
				ID:   fm.ID,
				File: doc.Path,
				Type: storeType,
				Gaps: gaps,
			})
			report.WithGaps++
			report.TotalGaps += len(gaps)
		} else {
			report.Clean++
		}
	}

	report.CodeMapIssues = analyzeCodeMap(c3Dir, graph)
	report.TotalGaps += len(report.CodeMapIssues)

	if jsonOut {
		return writeJSON(w, report)
	}

	return renderMarkdownReport(report, w)
}

// analyzeGaps checks an entity for quality gaps between frontmatter and body.
func analyzeGaps(fm *frontmatter.Frontmatter, body, storeType string) []MigrateGap {
	var gaps []MigrateGap

	sections := markdown.ParseSections(body)
	sectionMap := make(map[string]string)
	for _, s := range sections {
		if s.Name != "" {
			sectionMap[s.Name] = strings.TrimSpace(s.Content)
		}
	}

	// Check goal: frontmatter vs body
	if fm.Goal == "" {
		bodyGoal := sectionMap["Goal"]
		if bodyGoal != "" {
			excerpt := firstLine(bodyGoal, 80)
			gaps = append(gaps, MigrateGap{
				Field:       "goal",
				Status:      "empty_frontmatter_rich_body",
				BodyExcerpt: excerpt,
				Hint:        fmt.Sprintf("add `goal: %s` to frontmatter", excerpt),
			})
		}
	}

	// Check schema-required sections
	schemaSections := schema.ForType(storeType)
	for _, sec := range schemaSections {
		if !sec.Required {
			continue
		}
		content, exists := sectionMap[sec.Name]
		if !exists {
			gaps = append(gaps, MigrateGap{
				Field:  sec.Name,
				Status: "missing_section",
				Hint:   fmt.Sprintf("add `## %s` section — %s", sec.Name, sec.Purpose),
			})
			continue
		}
		if content == "" {
			gaps = append(gaps, MigrateGap{
				Field:  sec.Name,
				Status: "empty_section",
				Hint:   fmt.Sprintf("add content to `## %s` — %s", sec.Name, sec.Purpose),
			})
			continue
		}
		if sec.ContentType == "table" {
			table, err := markdown.ParseTable(content)
			if err == nil && len(table.Rows) == 0 {
				gaps = append(gaps, MigrateGap{
					Field:  sec.Name,
					Status: "empty_table",
					Hint:   fmt.Sprintf("`## %s` has headers but no data rows", sec.Name),
				})
			}
		}
	}

	return gaps
}

// analyzeIntegrity checks cross-references: broken refs, parents, origins.
func analyzeIntegrity(fm *frontmatter.Frontmatter, graph *walker.C3Graph) []MigrateGap {
	var gaps []MigrateGap

	// Broken parent
	if fm.Parent != "" && graph.Get(fm.Parent) == nil {
		gaps = append(gaps, MigrateGap{
			Field:  "parent",
			Status: "broken_ref",
			Hint:   fmt.Sprintf("parent `%s` not found — fix or remove", fm.Parent),
		})
	}

	// Broken uses/refs
	for _, ref := range fm.Refs {
		if graph.Get(ref) == nil {
			gaps = append(gaps, MigrateGap{
				Field:  "uses",
				Status: "broken_ref",
				Hint:   fmt.Sprintf("`%s` not found — verify ID with c3x list", ref),
			})
		}
	}

	// Broken affects
	for _, id := range fm.Affects {
		if graph.Get(id) == nil {
			gaps = append(gaps, MigrateGap{
				Field:  "affects",
				Status: "broken_ref",
				Hint:   fmt.Sprintf("`%s` not found", id),
			})
		}
	}

	// Broken scope
	for _, id := range fm.Scope {
		if graph.Get(id) == nil {
			gaps = append(gaps, MigrateGap{
				Field:  "scope",
				Status: "broken_ref",
				Hint:   fmt.Sprintf("`%s` not found", id),
			})
		}
	}

	// Broken origin (rules)
	for _, orig := range fm.Origin {
		target := frontmatter.StripAnchor(orig)
		if graph.Get(target) == nil {
			gaps = append(gaps, MigrateGap{
				Field:  "origin",
				Status: "broken_ref",
				Hint:   fmt.Sprintf("origin `%s` not found — should reference a ref or ADR", target),
			})
		}
	}

	// Broken sources (recipes)
	for _, src := range fm.Sources {
		target := frontmatter.StripAnchor(src)
		if graph.Get(target) == nil {
			gaps = append(gaps, MigrateGap{
				Field:  "sources",
				Status: "broken_ref",
				Hint:   fmt.Sprintf("source `%s` not found", target),
			})
		}
	}

	return gaps
}

// analyzeCodeMap validates code-map.yaml entries against the entity graph.
func analyzeCodeMap(c3Dir string, graph *walker.C3Graph) []MigrateCodeMapIssue {
	cmPath := filepath.Join(c3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err != nil || len(cm) == 0 {
		return nil
	}

	projectDir := filepath.Dir(c3Dir)

	var issues []MigrateCodeMapIssue
	for id, patterns := range cm {
		if id == "_exclude" {
			continue
		}

		// Check entity exists
		if graph.Get(id) == nil {
			issues = append(issues, MigrateCodeMapIssue{
				EntityID: id,
				Status:   "unknown_entity",
				Hint:     fmt.Sprintf("code-map references `%s` but no entity with that ID exists", id),
			})
			continue
		}

		// Check patterns match at least one file
		for _, pattern := range patterns {
			if pattern == "" {
				continue
			}
			matched, _ := codemap.GlobFiles(os.DirFS(projectDir), pattern)
			if len(matched) == 0 {
				issues = append(issues, MigrateCodeMapIssue{
					EntityID: id,
					Status:   "no_file_matches",
					Pattern:  pattern,
					Hint:     fmt.Sprintf("pattern `%s` matches 0 files — stale or wrong glob?", pattern),
				})
			}
		}
	}

	return issues
}

func renderMarkdownReport(report MigrateDryRunResult, w io.Writer) error {
	fmt.Fprintf(w, "# Migration Readiness: %d entities\n\n", report.Total)

	if report.WithGaps == 0 && len(report.CodeMapIssues) == 0 {
		fmt.Fprintln(w, "All entities are clean — ready to migrate.")
		fmt.Fprintln(w, "\nRun: c3x migrate")
		return nil
	}

	totalIssues := report.TotalGaps
	fmt.Fprintf(w, "## Gaps: %d issues across %d entities\n\n", totalIssues, report.WithGaps)

	for _, entity := range report.Entities {
		fmt.Fprintf(w, "### %s (%s) — %s\n", entity.ID, entity.Type, entity.File)
		for _, gap := range entity.Gaps {
			switch gap.Status {
			case "empty_frontmatter_rich_body":
				fmt.Fprintf(w, "- [ ] `%s:` empty — body has: %q\n", gap.Field, gap.BodyExcerpt)
			case "missing_section":
				fmt.Fprintf(w, "- [ ] `## %s` missing — %s\n", gap.Field, gap.Hint)
			case "empty_section":
				fmt.Fprintf(w, "- [ ] `## %s` empty — %s\n", gap.Field, gap.Hint)
			case "empty_table":
				fmt.Fprintf(w, "- [ ] `## %s` table has no data rows\n", gap.Field)
			case "broken_ref":
				fmt.Fprintf(w, "- [ ] `%s` broken — %s\n", gap.Field, gap.Hint)
			}
		}
		fmt.Fprintln(w)
	}

	if len(report.CodeMapIssues) > 0 {
		fmt.Fprintf(w, "## Code-Map Issues: %d\n\n", len(report.CodeMapIssues))
		for _, issue := range report.CodeMapIssues {
			switch issue.Status {
			case "unknown_entity":
				fmt.Fprintf(w, "- [ ] `%s` — %s\n", issue.EntityID, issue.Hint)
			case "no_file_matches":
				fmt.Fprintf(w, "- [ ] `%s` pattern `%s` — %s\n", issue.EntityID, issue.Pattern, issue.Hint)
			}
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "## Clean: %d entities — ready to migrate\n\n", report.Clean)
	fmt.Fprintf(w, "Fix %d issues, then run: `c3x migrate`\n", totalIssues)
	return nil
}

// firstLine returns the first line of text, truncated to maxLen.
func firstLine(s string, maxLen int) string {
	line := strings.SplitN(s, "\n", 2)[0]
	line = strings.TrimSpace(line)
	if len(line) > maxLen {
		return line[:maxLen-3] + "..."
	}
	return line
}
