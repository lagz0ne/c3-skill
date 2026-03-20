package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// Issue represents a validation finding.
type Issue struct {
	Severity string `json:"severity"`
	Entity   string `json:"entity,omitempty"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
}

// CheckResult holds the validation output.
type CheckResult struct {
	Total  int     `json:"total"`
	Issues []Issue `json:"issues"`
}

// CheckOptions holds parameters for the enhanced check command.
type CheckOptions struct {
	Store      *store.Store
	JSON       bool
	ProjectDir string
	C3Dir      string
	IncludeADR bool
	Fix        bool
}

// buildTitleMapStore creates a case-insensitive title/slug -> entity ID lookup.
func buildTitleMapStore(s *store.Store) map[string]string {
	m := make(map[string]string)
	entities, _ := s.AllEntities()
	for _, e := range entities {
		lower := strings.ToLower(e.Title)
		if existing, exists := m[lower]; exists {
			if existing != e.ID {
				m[lower] = "" // ambiguous
			}
		} else {
			m[lower] = e.ID
		}
		if e.Slug != "" {
			slugLower := strings.ToLower(e.Slug)
			if existing, exists := m[slugLower]; exists {
				if existing != e.ID {
					m[slugLower] = ""
				}
			} else {
				m[slugLower] = e.ID
			}
		}
	}
	return m
}

// suggestByTitle returns a suggested entity ID if the value matches a title/slug, or "".
func suggestByTitle(val string, titleMap map[string]string) string {
	if id := titleMap[strings.ToLower(val)]; id != "" {
		return id
	}
	return ""
}

func hintFor(message string) string {
	patterns := []struct {
		substr string
		hint   string
	}{
		{"code-map parse error", "fix YAML syntax in .c3/code-map.yaml"},
		{"missing required section", ""},
		{"empty required section", ""},
		{"empty required table", "add at least one data row below the table headers"},
		{"unknown entity reference", "verify the ID with 'c3x list'; check for typos"},
		{"unknown ref reference", "use a ref-* ID (e.g., ref-jwt); verify with 'c3x list'"},
		{"file does not exist", "create the file or fix the path"},
	}
	for _, p := range patterns {
		if !strings.Contains(message, p.substr) {
			continue
		}
		if p.hint != "" {
			return p.hint
		}
		if strings.Contains(message, "missing required section: ") {
			section := strings.TrimPrefix(message, "missing required section: ")
			return fmt.Sprintf("add a ## %s section with content", section)
		}
		if strings.Contains(message, "empty required section: ") {
			section := strings.TrimPrefix(message, "empty required section: ")
			return fmt.Sprintf("add content to the ## %s section", section)
		}
	}
	return ""
}

func countSeverities(issues []Issue) (errors, warnings int) {
	for _, issue := range issues {
		if issue.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}
	return
}

func formatCounts(errors, warnings int) string {
	var parts []string
	if errors > 0 {
		noun := "error"
		if errors > 1 {
			noun = "errors"
		}
		parts = append(parts, fmt.Sprintf("%d %s", errors, noun))
	}
	if warnings > 0 {
		noun := "warning"
		if warnings > 1 {
			noun = "warnings"
		}
		parts = append(parts, fmt.Sprintf("%d %s", warnings, noun))
	}
	return strings.Join(parts, ", ")
}

func checkRecipeSourcesStore(s *store.Store) []Issue {
	var issues []Issue
	recipes, _ := s.EntitiesByType("recipe")
	for _, r := range recipes {
		rels, _ := s.RelationshipsFrom(r.ID)
		for _, rel := range rels {
			if rel.RelType == "sources" {
				if _, err := s.GetEntity(rel.ToID); err != nil {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   r.ID,
						Message:  fmt.Sprintf("recipe references nonexistent entity: %s", rel.ToID),
					})
				}
			}
		}
	}
	return issues
}

// RunCheckV2 validates entities against the schema registry.
func RunCheckV2(opts CheckOptions, w io.Writer) error {
	var issues []Issue

	titleMap := buildTitleMapStore(opts.Store)

	entities, _ := opts.Store.AllEntities()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, entity := range entities {
		if entity.Type == "adr" && !opts.IncludeADR {
			continue
		}

		schemaSections := schema.ForType(entity.Type)
		if schemaSections == nil {
			continue
		}

		bodySections := markdown.ParseSections(entity.Body)

		// Build lookup of parsed sections by name
		sectionMap := make(map[string]markdown.Section)
		for _, s := range bodySections {
			if s.Name != "" {
				sectionMap[s.Name] = s
			}
		}

		for _, schemaDef := range schemaSections {
			if !schemaDef.Required {
				continue
			}

			bodySection, exists := sectionMap[schemaDef.Name]

			if !exists {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("missing required section: %s", schemaDef.Name),
				})
				continue
			}

			content := strings.TrimSpace(bodySection.Content)

			if schemaDef.ContentType == "table" {
				if content == "" {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("empty required table: %s", schemaDef.Name),
					})
					continue
				}
				table, err := markdown.ParseTable(content)
				if err != nil {
					continue
				}
				if len(table.Rows) == 0 {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("empty required table: %s (headers only, no data rows)", schemaDef.Name),
					})
					continue
				}

				for _, col := range schemaDef.Columns {
					issues = append(issues, validateColumn(col, table, entity, opts, titleMap)...)
				}
			} else {
				if content == "" {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("empty required section: %s", schemaDef.Name),
					})
				}
			}
		}

		// Layer 3: Check non-required table sections too
		for _, schemaDef := range schemaSections {
			if schemaDef.Required || schemaDef.ContentType != "table" {
				continue
			}
			bodySection, exists := sectionMap[schemaDef.Name]
			if !exists {
				continue
			}
			content := strings.TrimSpace(bodySection.Content)
			if content == "" {
				continue
			}
			table, err := markdown.ParseTable(content)
			if err != nil || len(table.Rows) == 0 {
				continue
			}
			for _, col := range schemaDef.Columns {
				issues = append(issues, validateColumn(col, table, entity, opts, titleMap)...)
			}
		}

		// Validate origin references for rules
		if entity.Type == "rule" {
			rels, _ := opts.Store.RelationshipsFrom(entity.ID)
			for _, r := range rels {
				if r.RelType == "origin" {
					if _, err := opts.Store.GetEntity(r.ToID); err != nil {
						issues = append(issues, Issue{
							Severity: "error",
							Entity:   entity.ID,
							Message:  fmt.Sprintf("origin reference %q not found", r.ToID),
							Hint:     "origin should reference an existing ref or ADR entity",
						})
					}
				}
			}
		}
	}

	// Code-map validation
	allCodeMap, err := opts.Store.AllCodeMap()
	if err == nil && len(allCodeMap) > 0 && opts.ProjectDir != "" {
		for entityID, patterns := range allCodeMap {
			entity, err := opts.Store.GetEntity(entityID)
			if err != nil {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entityID,
					Message:  fmt.Sprintf("code-map entity %s not found in store", entityID),
				})
				continue
			}
			if entity.Type != "component" && entity.Type != "ref" && entity.Type != "rule" {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entityID,
					Message:  fmt.Sprintf("code-map: %s is not a component or ref", entityID),
				})
			}
			for _, p := range patterns {
				if p == "" {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entityID,
						Message:  "empty path in code-map",
					})
				}
			}
		}
	}

	issues = append(issues, checkRecipeSourcesStore(opts.Store)...)

	// Scope cross-check
	refs, _ := opts.Store.EntitiesByType("ref")
	for _, ref := range refs {
		rels, _ := opts.Store.RelationshipsFrom(ref.ID)
		for _, r := range rels {
			if r.RelType != "scope" {
				continue
			}
			children, _ := opts.Store.Children(r.ToID)
			for _, child := range children {
				if child.Type != "component" {
					continue
				}
				// Check if child cites this ref
				childRels, _ := opts.Store.RelationshipsFrom(child.ID)
				cited := false
				for _, cr := range childRels {
					if cr.RelType == "uses" && cr.ToID == ref.ID {
						cited = true
						break
					}
				}
				if !cited {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   child.ID,
						Message:  fmt.Sprintf("ref %s scopes %s but %s does not cite it", ref.ID, r.ToID, child.ID),
					})
				}
			}
		}
	}

	result := CheckResult{
		Total:  len(entities),
		Issues: issues,
	}
	if result.Issues == nil {
		result.Issues = []Issue{}
	}

	// Output
	if opts.JSON {
		for i := range result.Issues {
			result.Issues[i].Hint = hintFor(result.Issues[i].Message)
		}
		return writeJSON(w, result)
	}

	// Text output
	errors, warnings := countSeverities(result.Issues)
	if len(result.Issues) == 0 {
		fmt.Fprintf(w, "Checked %d docs — all clear\n", result.Total)
		return nil
	}

	fmt.Fprintf(w, "Checked %d docs — %s\n\n", result.Total, formatCounts(errors, warnings))

	for _, issue := range result.Issues {
		entityLabel := issue.Entity
		if entityLabel == "" {
			entityLabel = "global"
		}
		icon := "!"
		if issue.Severity == "error" {
			icon = "x"
		}
		fmt.Fprintf(w, "  %s %s: %s\n", icon, entityLabel, issue.Message)
		if hint := hintFor(issue.Message); hint != "" {
			fmt.Fprintf(w, "    -> %s\n", hint)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "Legend: x = error (fix first)  ! = warning (incomplete, should fix)")

	return nil
}

// validateColumn checks typed column values in a table.
func validateColumn(col schema.ColumnDef, table *markdown.Table, entity *store.Entity, opts CheckOptions, titleMap map[string]string) []Issue {
	var issues []Issue
	switch col.Type {
	case "filepath":
		if opts.ProjectDir == "" {
			return nil
		}
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			absPath := filepath.Join(opts.ProjectDir, val)
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("file does not exist: %s", val),
				})
			}
		}
	case "entity_id":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			if _, err := opts.Store.GetEntity(val); err != nil {
				msg := fmt.Sprintf("unknown entity reference: %s", val)
				if suggested := suggestByTitle(val, titleMap); suggested != "" {
					msg = fmt.Sprintf("unknown entity reference: %s (did you mean %s?)", val, suggested)
				}
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  msg,
				})
			}
		}
	case "ref_id":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			if _, err := opts.Store.GetEntity(val); err != nil {
				msg := fmt.Sprintf("unknown ref reference: %s", val)
				if suggested := suggestByTitle(val, titleMap); suggested != "" {
					msg = fmt.Sprintf("unknown ref reference: %s (did you mean %s?)", val, suggested)
				}
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  msg,
				})
			}
		}
	}
	return issues
}
