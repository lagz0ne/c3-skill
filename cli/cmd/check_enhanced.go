package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
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
	Graph         *walker.C3Graph
	Docs          []frontmatter.ParsedDoc
	JSON          bool
	ProjectDir    string
	C3Dir         string // path to .c3/ directory (may differ from ProjectDir/.c3/ with --c3-dir)
	ParseWarnings []walker.ParseWarning
	IncludeADR    bool
}

func hintFor(message string) string {
	patterns := []struct {
		substr string
		hint   string
	}{
		{"broken YAML frontmatter", "check for unquoted colons in values"},
		{"code-map parse error", "fix YAML syntax in .c3/code-map.yaml"},
		{"missing required section", ""},   // dynamic — handled below
		{"empty required section", ""},     // dynamic — handled below
		{"empty required table", "add at least one data row below the table headers"},
		{"unknown entity reference", "verify the ID with 'c3x list'; check for typos"},
		{"file does not exist", "create the file or fix the path"},
		{"not found in C3 graph", "remove from code-map.yaml, or create the entity doc"},
		{"not a component or ref", "only components and refs belong in code-map"},
		{"empty path in code-map", "remove the empty entry or add a file pattern"},
		{"absolute path not allowed", "use a relative path from the project root"},
		{"escapes project root", "use a path within the project"},
		{"is a directory", "point to source files, not directories"},
		{"no files match pattern", "fix the glob pattern or create matching files"},
		{"note references nonexistent entity", "remove the note or update its sources"},
		{"recipe references nonexistent entity", "fix the source ID or remove it from the recipe"},
	}
	for _, p := range patterns {
		if !strings.Contains(message, p.substr) {
			continue
		}
		if p.hint != "" {
			return p.hint
		}
		// Dynamic hints: extract section name from "missing required section: X" or "empty required section: X"
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

// Strips #fragment anchors before lookup.
func validateSourceRefs(entityLabel string, sources []string, graph *walker.C3Graph, noun string) []Issue {
	var issues []Issue
	for _, src := range sources {
		entityID := frontmatter.StripAnchor(src)
		if graph.Get(entityID) == nil {
			issues = append(issues, Issue{
				Severity: "warning",
				Entity:   entityLabel,
				Message:  fmt.Sprintf("%s references nonexistent entity: %s", noun, entityID),
			})
		}
	}
	return issues
}

func checkNotes(c3Dir string, graph *walker.C3Graph) []Issue {
	notesDir := filepath.Join(c3Dir, "_index", "notes")
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return nil // directory doesn't exist — skip silently
	}

	var issues []Issue
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(notesDir, entry.Name()))
		if err != nil {
			continue
		}
		sources := parseNoteSources(string(data))
		issues = append(issues, validateSourceRefs(entry.Name(), sources, graph, "note")...)
	}
	return issues
}

func parseNoteSources(content string) []string {
	if !strings.HasPrefix(content, "---\n") {
		return nil
	}
	end := strings.Index(content[4:], "\n---\n")
	if end == -1 {
		// Handle EOF edge case: frontmatter ends with \n--- at end of string
		if strings.HasSuffix(content[4:], "\n---") {
			end = len(content[4:]) - 4 // length minus "\n---"
		} else {
			return nil
		}
	}
	yamlBlock := content[4 : 4+end]

	var fm struct {
		Sources []string `yaml:"sources"`
	}
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil
	}
	return fm.Sources
}

func checkRecipeSources(graph *walker.C3Graph) []Issue {
	var issues []Issue
	for _, r := range graph.ByType(frontmatter.DocRecipe) {
		issues = append(issues, validateSourceRefs(r.ID, r.Frontmatter.Sources, graph, "recipe")...)
	}
	return issues
}

// RunCheckV2 validates entities against the schema registry.
func RunCheckV2(opts CheckOptions, w io.Writer) error {
	var issues []Issue

	for _, pw := range opts.ParseWarnings {
		issues = append(issues, Issue{
			Severity: "error",
			Entity:   pw.Path,
			Message:  "broken YAML frontmatter: file has --- delimiters but failed to parse (check for unquoted colons in values)",
		})
	}

	// Build sorted entity list for deterministic output
	entities := opts.Graph.All()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, entity := range entities {
		docType := frontmatter.ClassifyDoc(entity.Frontmatter)
		if docType == frontmatter.DocADR && !opts.IncludeADR {
			continue
		}
		typeName := docType.String()

		schemaSections := schema.ForType(typeName)
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

			// Layer 2: Missing required section
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
				// Layer 2: Empty required table (headers present but no data rows)
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

				// Layer 3: Typed column validation
				for _, col := range schemaDef.Columns {
					switch col.Type {
					case "filepath":
						if opts.ProjectDir == "" {
							continue
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
							if opts.Graph.Get(val) == nil {
								issues = append(issues, Issue{
									Severity: "warning",
									Entity:   entity.ID,
									Message:  fmt.Sprintf("unknown entity reference: %s", val),
								})
							}
						}
					}
				}
			} else {
				// Layer 2: Empty required text section
				if content == "" {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("empty required section: %s", schemaDef.Name),
					})
				}
			}
		}

		// Layer 3: Check non-required table sections too for typed validation
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
				switch col.Type {
				case "entity_id":
					for _, row := range table.Rows {
						val := strings.TrimSpace(row[col.Name])
						if val == "" {
							continue
						}
						if opts.Graph.Get(val) == nil {
							issues = append(issues, Issue{
								Severity: "warning",
								Entity:   entity.ID,
								Message:  fmt.Sprintf("unknown entity reference: %s", val),
							})
						}
					}
				case "filepath":
					if opts.ProjectDir == "" {
						continue
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
				}
			}
		}
	}

	// Code-map validation
	c3Dir := opts.C3Dir
	if c3Dir == "" && opts.ProjectDir != "" {
		c3Dir = filepath.Join(opts.ProjectDir, ".c3")
	}
	if c3Dir != "" {
		cmPath := filepath.Join(c3Dir, "code-map.yaml")
		cm, err := codemap.ParseCodeMap(cmPath)
		if err != nil {
			issues = append(issues, Issue{
				Severity: "error",
				Entity:   "",
				Message:  fmt.Sprintf("code-map parse error: %v", err),
			})
		} else if len(cm) > 0 {
			knownEntities := make(map[string]string)
			for _, e := range entities {
				knownEntities[e.ID] = frontmatter.ClassifyDoc(e.Frontmatter).String()
			}
			for _, ci := range codemap.Validate(cm, knownEntities, opts.ProjectDir) {
				issues = append(issues, Issue{
					Severity: ci.Severity,
					Entity:   ci.Entity,
					Message:  ci.Message,
				})
			}
		}
	}

	if c3Dir != "" {
		issues = append(issues, checkNotes(c3Dir, opts.Graph)...)
	}
	issues = append(issues, checkRecipeSources(opts.Graph)...)

	result := CheckResult{
		Total:  len(opts.Docs),
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

	// Text output — summary header
	errors, warnings := countSeverities(result.Issues)
	if len(result.Issues) == 0 {
		fmt.Fprintf(w, "Checked %d docs — all clear\n", result.Total)
		return nil
	}
	fmt.Fprintf(w, "Checked %d docs — %s\n\n", result.Total, formatCounts(errors, warnings))

	// Issue lines with hints
	for _, issue := range result.Issues {
		entityLabel := issue.Entity
		if entityLabel == "" {
			entityLabel = "global"
		}
		icon := "!"
		if issue.Severity == "error" {
			icon = "✗"
		}
		fmt.Fprintf(w, "  %s %s: %s\n", icon, entityLabel, issue.Message)
		if hint := hintFor(issue.Message); hint != "" {
			fmt.Fprintf(w, "    → %s\n", hint)
		}
		fmt.Fprintln(w)
	}

	// Legend
	fmt.Fprintln(w, "Legend: ✗ = error (fix first)  ! = warning (incomplete, should fix)")

	return nil
}
