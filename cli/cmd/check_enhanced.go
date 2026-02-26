package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// CheckOptions holds parameters for the enhanced check command.
type CheckOptions struct {
	Graph      *walker.C3Graph
	Docs       []frontmatter.ParsedDoc
	JSON       bool
	ProjectDir string
}

// RunCheckV2 validates entities against the schema registry.
func RunCheckV2(opts CheckOptions, w io.Writer) error {
	var issues []Issue

	// Build sorted entity list for deterministic output
	entities := opts.Graph.All()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, entity := range entities {
		docType := frontmatter.ClassifyDoc(entity.Frontmatter)
		typeName := docType.String()

		schemaSections, ok := schemaRegistry[typeName]
		if !ok {
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

	// Layer 3: Bidirectional consistency — if entity refs[] lists a ref,
	// that ref's "Cited By" table should contain the entity.
	for _, entity := range entities {
		if len(entity.Frontmatter.Refs) == 0 {
			continue
		}
		for _, refID := range entity.Frontmatter.Refs {
			refEntity := opts.Graph.Get(refID)
			if refEntity == nil {
				continue
			}
			// Check if ref's "Cited By" table lists this entity
			citedByTable, err := markdown.ExtractTableFromSection(refEntity.Body, "Cited By")
			if err != nil || citedByTable == nil {
				// No Cited By section or not parseable — flag inconsistency
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("refs %s but %s has no Cited By entry for %s", refID, refID, entity.ID),
				})
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   refID,
					Message:  fmt.Sprintf("missing Cited By entry for %s (which refs this)", entity.ID),
				})
				continue
			}

			// Check if any row in Cited By contains this entity's ID
			found := false
			for _, row := range citedByTable.Rows {
				for _, val := range row {
					if strings.Contains(val, entity.ID) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("refs %s but %s has no Cited By entry for %s", refID, refID, entity.ID),
				})
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   refID,
					Message:  fmt.Sprintf("missing Cited By entry for %s (which refs this)", entity.ID),
				})
			}
		}
	}

	result := CheckResult{
		Total:  len(opts.Docs),
		Issues: issues,
	}
	if result.Issues == nil {
		result.Issues = []Issue{}
	}

	// Output
	if opts.JSON {
		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
	}

	// Text output
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
	}

	return nil
}
