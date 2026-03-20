package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// MigrateGap represents a single quality gap found in an entity.
type MigrateGap struct {
	Field       string `json:"field"`
	Status      string `json:"status"`       // "empty_frontmatter_rich_body", "missing_section", "empty_section", "empty_table"
	BodyExcerpt string `json:"body_excerpt"` // first line of body content if available
	Hint        string `json:"hint"`
}

// MigrateEntityReport is the dry-run analysis for one entity.
type MigrateEntityReport struct {
	ID   string       `json:"id"`
	File string       `json:"file"`
	Type string       `json:"type"`
	Gaps []MigrateGap `json:"gaps"`
}

// MigrateDryRunResult is the full dry-run output.
type MigrateDryRunResult struct {
	Total    int                   `json:"total"`
	WithGaps int                   `json:"with_gaps"`
	Clean    int                   `json:"clean"`
	TotalGaps int                  `json:"total_gaps"`
	Entities []MigrateEntityReport `json:"entities"`
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

	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	return renderMarkdownReport(report, w)
}

// analyzeGaps checks an entity for quality gaps between frontmatter and body.
func analyzeGaps(fm *frontmatter.Frontmatter, body, storeType string) []MigrateGap {
	var gaps []MigrateGap

	sections := markdown.ParseSections(body)
	sectionMap := make(map[string]string) // name -> trimmed content
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
		// For table sections, check if there are data rows
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

func renderMarkdownReport(report MigrateDryRunResult, w io.Writer) error {
	fmt.Fprintf(w, "# Migration Readiness: %d entities\n\n", report.Total)

	if report.WithGaps == 0 {
		fmt.Fprintln(w, "All entities are clean — ready to migrate.")
		fmt.Fprintln(w, "\nRun: c3x migrate")
		return nil
	}

	fmt.Fprintf(w, "## Gaps: %d issues across %d entities\n\n", report.TotalGaps, report.WithGaps)

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
			}
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "## Clean: %d entities — ready to migrate\n\n", report.Clean)
	fmt.Fprintf(w, "Fix %d gaps, then run: `c3x migrate`\n", report.TotalGaps)
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
