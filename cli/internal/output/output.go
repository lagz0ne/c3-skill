package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

// RenderTopology renders a tree view of containers with their components.
func RenderTopology(graph *walker.C3Graph) string {
	var lines []string

	containers := graph.ByType(frontmatter.DocContainer)
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].ID < containers[j].ID
	})

	for _, container := range containers {
		goal := ""
		if container.Frontmatter.Goal != "" {
			goal = " — " + truncate(container.Frontmatter.Goal, 60)
		}
		lines = append(lines, fmt.Sprintf("%s-%s (container)%s", container.ID, container.Slug, goal))

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
				numStr := strings.TrimPrefix(comp.ID, strings.TrimRight(comp.Frontmatter.Parent, "0123456789"))
				// Try to determine from ID: if trailing number <= 9, foundation
				if n, err := strconv.Atoi(numStr); err == nil && n <= 9 {
					category = "foundation"
				} else {
					category = "feature"
				}
			}
			compGoal := ""
			if comp.Frontmatter.Goal != "" {
				compGoal = " — " + truncate(comp.Frontmatter.Goal, 60)
			}
			refs := graph.RefsFor(comp.ID)
			refIDs := make([]string, len(refs))
			for j, r := range refs {
				refIDs[j] = r.ID
			}
			suffix := ""
			if len(refIDs) > 0 {
				suffix = " → ref: " + strings.Join(refIDs, ", ")
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
			citers := graph.CitedBy(ref.ID)
			citerIDs := make([]string, len(citers))
			for j, c := range citers {
				citerIDs[j] = c.ID
			}
			suffix := ""
			if len(citerIDs) > 0 {
				suffix = " → used by: " + strings.Join(citerIDs, ", ")
			}
			lines = append(lines, fmt.Sprintf("  %s%s%s", ref.ID, refGoal, suffix))
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

	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// RenderFlatList renders a flat tab-separated list sorted by path.
func RenderFlatList(graph *walker.C3Graph) string {
	all := graph.All()
	sort.Slice(all, func(i, j int) bool {
		return all[i].Path < all[j].Path
	})

	var lines []string
	for _, e := range all {
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s", e.ID, e.Type, e.Path))
	}
	return strings.Join(lines, "\n")
}

// RenderJSON marshals data to indented JSON.
func RenderJSON(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(b)
}
