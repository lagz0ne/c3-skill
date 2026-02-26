package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// Issue represents a validation finding.
type Issue struct {
	Severity string `json:"severity"`
	Entity   string `json:"entity,omitempty"`
	Message  string `json:"message"`
}

// CheckResult holds the validation output.
type CheckResult struct {
	Total  int     `json:"total"`
	Issues []Issue `json:"issues"`
}

// RunCheck validates a C3 graph and raw docs, reporting issues.
// docs is used for duplicate ID detection (the graph deduplicates by map key).
func RunCheck(graph *walker.C3Graph, docs []frontmatter.ParsedDoc, jsonOutput bool, w io.Writer) error {
	result := CheckResult{
		Total:  len(docs),
		Issues: []Issue{},
	}

	// Duplicate ID check from raw docs (before graph dedup)
	seenIDs := make(map[string]bool)
	for _, doc := range docs {
		if doc.Frontmatter == nil {
			continue
		}
		id := doc.Frontmatter.ID
		if seenIDs[id] {
			result.Issues = append(result.Issues, Issue{
				Severity: "error",
				Entity:   id,
				Message:  "duplicate ID",
			})
		}
		seenIDs[id] = true
	}

	entities := graph.All()
	// Sort for deterministic output
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, entity := range entities {
		// Broken relationships: references to non-existent entities
		for _, relID := range entity.Relationships {
			if graph.Get(relID) == nil {
				result.Issues = append(result.Issues, Issue{
					Severity: "error",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("broken link to '%s'", relID),
				})
			}
		}

		// Missing parent container for components
		if entity.Type == frontmatter.DocComponent {
			parentID := entity.Frontmatter.Parent
			if parentID == "" {
				result.Issues = append(result.Issues, Issue{
					Severity: "error",
					Entity:   entity.ID,
					Message:  "missing parent container",
				})
			} else if graph.Get(parentID) == nil {
				result.Issues = append(result.Issues, Issue{
					Severity: "error",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("parent container '%s' not found", parentID),
				})
			}
		}

		// Empty content check
		if strings.TrimSpace(entity.Body) == "" {
			result.Issues = append(result.Issues, Issue{
				Severity: "warning",
				Entity:   entity.ID,
				Message:  "empty content body",
			})
		}

		// ID/filename mismatch check
		basename := strings.TrimSuffix(filepath.Base(entity.Path), ".md")
		if basename != "README" {
			if !strings.HasPrefix(basename, entity.ID) {
				result.Issues = append(result.Issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("ID/filename mismatch: id='%s', file='%s.md'", entity.ID, basename),
				})
			}
		}
	}

	// Orphan check: entities with no incoming relationships (except context and containers)
	for _, entity := range entities {
		if entity.Type == frontmatter.DocContext || entity.Type == frontmatter.DocContainer {
			continue
		}
		incoming := graph.Reverse(entity.ID)
		if len(incoming) == 0 && entity.Frontmatter.Parent == "" {
			result.Issues = append(result.Issues, Issue{
				Severity: "warning",
				Entity:   entity.ID,
				Message:  "orphan: no incoming relationships",
			})
		}
	}

	// Output
	if jsonOutput {
		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
	}

	errors := 0
	warnings := 0
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}

	if errors == 0 && warnings == 0 {
		fmt.Fprintf(w, "✓ %d entities, 0 issues\n", result.Total)
	} else {
		fmt.Fprintf(w, "%d entities, %d errors, %d warnings\n", result.Total, errors, warnings)
		for _, issue := range result.Issues {
			icon := "✗"
			if issue.Severity != "error" {
				icon = "!"
			}
			entity := issue.Entity
			if entity == "" {
				entity = "global"
			}
			fmt.Fprintf(w, "  %s %s: %s\n", icon, entity, issue.Message)
		}
	}

	return nil
}
