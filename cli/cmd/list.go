package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// RunList outputs the topology of a C3 graph.
func RunList(graph *walker.C3Graph, jsonOutput bool, flat bool, w io.Writer) error {
	if jsonOutput {
		return listJSON(graph, w)
	}
	if flat {
		return listFlat(graph, w)
	}
	return listTopology(graph, w)
}

func listJSON(graph *walker.C3Graph, w io.Writer) error {
	entities := graph.All()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	type jsonEntity struct {
		ID            string                 `json:"id"`
		Type          string                 `json:"type"`
		Title         string                 `json:"title"`
		Path          string                 `json:"path"`
		Relationships []string               `json:"relationships"`
		Frontmatter   map[string]interface{} `json:"frontmatter"`
	}

	var data []jsonEntity
	for _, e := range entities {
		fm := make(map[string]interface{})
		if e.Frontmatter.Goal != "" {
			fm["goal"] = e.Frontmatter.Goal
		}
		if e.Frontmatter.Category != "" {
			fm["category"] = e.Frontmatter.Category
		}
		if e.Frontmatter.Parent != "" {
			fm["parent"] = e.Frontmatter.Parent
		}
		if e.Frontmatter.Status != "" {
			fm["status"] = e.Frontmatter.Status
		}
		if e.Frontmatter.Boundary != "" {
			fm["boundary"] = e.Frontmatter.Boundary
		}

		data = append(data, jsonEntity{
			ID:            e.ID,
			Type:          e.Type.String(),
			Title:         e.Title,
			Path:          e.Path,
			Relationships: e.Relationships,
			Frontmatter:   fm,
		})
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(out))
	return nil
}

func listFlat(graph *walker.C3Graph, w io.Writer) error {
	entities := graph.All()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Path < entities[j].Path
	})

	for _, e := range entities {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Type.String(), e.Path)
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func listTopology(graph *walker.C3Graph, w io.Writer) error {
	var lines []string

	// Containers with their components
	containers := graph.ByType(frontmatter.DocContainer)
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].ID < containers[j].ID
	})

	for _, container := range containers {
		containerGoal := ""
		if container.Frontmatter.Goal != "" {
			containerGoal = " — " + truncate(container.Frontmatter.Goal, 60)
		}
		lines = append(lines, fmt.Sprintf("%s-%s (container)%s", container.ID, container.Slug, containerGoal))

		components := graph.Children(container.ID)
		sort.Slice(components, func(i, j int) bool {
			return components[i].ID < components[j].ID
		})

		for i, comp := range components {
			isLast := i == len(components)-1
			prefix := "├── "
			if isLast {
				prefix = "└── "
			}

			category := comp.Frontmatter.Category
			if category == "" {
				category = "foundation"
			}

			var refIDs []string
			for _, ref := range graph.RefsFor(comp.ID) {
				refIDs = append(refIDs, ref.ID)
			}
			suffix := ""
			if len(refIDs) > 0 {
				suffix = " → ref: " + strings.Join(refIDs, ", ")
			}

			compGoal := ""
			if comp.Frontmatter.Goal != "" {
				compGoal = " — " + truncate(comp.Frontmatter.Goal, 60)
			}

			lines = append(lines, fmt.Sprintf("%s%s-%s (%s)%s%s", prefix, comp.ID, comp.Slug, category, compGoal, suffix))
		}
		lines = append(lines, "")
	}

	// Cross-cutting refs
	refs := graph.ByType(frontmatter.DocRef)
	if len(refs) > 0 {
		sort.Slice(refs, func(i, j int) bool {
			return refs[i].ID < refs[j].ID
		})
		lines = append(lines, "Cross-cutting:")
		for _, ref := range refs {
			refGoal := ""
			if ref.Frontmatter.Goal != "" {
				refGoal = " — " + truncate(ref.Frontmatter.Goal, 60)
			}

			var citerIDs []string
			for _, c := range graph.CitedBy(ref.ID) {
				citerIDs = append(citerIDs, c.ID)
			}
			citerSuffix := ""
			if len(citerIDs) > 0 {
				citerSuffix = " → used by: " + strings.Join(citerIDs, ", ")
			}

			lines = append(lines, fmt.Sprintf("  %s%s%s", ref.ID, refGoal, citerSuffix))
		}
		lines = append(lines, "")
	}

	// ADRs
	adrs := graph.ByType(frontmatter.DocADR)
	if len(adrs) > 0 {
		sort.Slice(adrs, func(i, j int) bool {
			return adrs[i].ID < adrs[j].ID
		})
		lines = append(lines, "ADRs:")
		for _, adr := range adrs {
			status := adr.Frontmatter.Status
			if status == "" {
				status = "unknown"
			}
			lines = append(lines, fmt.Sprintf("  %s: %s → status: %s", adr.ID, adr.Title, status))
		}
		lines = append(lines, "")
	}

	fmt.Fprint(w, strings.TrimRight(strings.Join(lines, "\n"), "\n"))
	fmt.Fprintln(w)
	return nil
}
