package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// ListOptions holds parameters for the list command.
type ListOptions struct {
	Graph      *walker.C3Graph
	JSON       bool
	Flat       bool
	Compact    bool
	C3Dir      string
	IncludeADR bool
}

// RunList outputs the topology of a C3 graph.
func RunList(opts ListOptions, w io.Writer) error {
	// Load codemap for file paths (used by JSON and topology modes)
	var cm codemap.CodeMap
	if !opts.Compact || opts.JSON {
		var err error
		cm, err = codemap.ParseCodeMap(filepath.Join(opts.C3Dir, "code-map.yaml"))
		if err != nil {
			cm = codemap.CodeMap{}
		}
	}
	if opts.JSON {
		return listJSON(opts.Graph, opts.IncludeADR, cm, w)
	}
	if opts.Flat {
		return listFlat(opts.Graph, opts.IncludeADR, w)
	}
	return listTopology(opts.Graph, cm, opts.Compact, opts.IncludeADR, w)
}

func listJSON(graph *walker.C3Graph, includeADR bool, cm codemap.CodeMap, w io.Writer) error {
	entities := graph.All()
	if !includeADR {
		filtered := entities[:0]
		for _, e := range entities {
			if frontmatter.ClassifyDoc(e.Frontmatter) != frontmatter.DocADR {
				filtered = append(filtered, e)
			}
		}
		entities = filtered
	}
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
		Files         []string               `json:"files,omitempty"`
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
		if e.Frontmatter.Description != "" {
			fm["description"] = e.Frontmatter.Description
		}
		if len(e.Frontmatter.Sources) > 0 {
			fm["sources"] = e.Frontmatter.Sources
		}
		if len(e.Frontmatter.Refs) > 0 {
			fm["refs"] = e.Frontmatter.Refs
		}
		if len(e.Frontmatter.Affects) > 0 {
			fm["affects"] = e.Frontmatter.Affects
		}
		if len(e.Frontmatter.Scope) > 0 {
			fm["scope"] = e.Frontmatter.Scope
		}

		var files []string
		if f := cm[e.ID]; len(f) > 0 {
			files = append([]string(nil), f...)
			sort.Strings(files)
		}

		data = append(data, jsonEntity{
			ID:            e.ID,
			Type:          e.Type.String(),
			Title:         e.Title,
			Path:          e.Path,
			Relationships: e.Relationships,
			Frontmatter:   fm,
			Files:         files,
		})
	}

	return writeJSON(w, data)
}

func listFlat(graph *walker.C3Graph, includeADR bool, w io.Writer) error {
	entities := graph.All()
	if !includeADR {
		filtered := entities[:0]
		for _, e := range entities {
			if frontmatter.ClassifyDoc(e.Frontmatter) != frontmatter.DocADR {
				filtered = append(filtered, e)
			}
		}
		entities = filtered
	}
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Path < entities[j].Path
	})

	for _, e := range entities {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Type.String(), e.Path)
	}
	return nil
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

func listTopology(graph *walker.C3Graph, cm codemap.CodeMap, compact bool, includeADR bool, w io.Writer) error {
	containers := graph.ByType(frontmatter.DocContainer)
	components := graph.ByType(frontmatter.DocComponent)
	refs := graph.ByType(frontmatter.DocRef)
	adrs := graph.ByType(frontmatter.DocADR)
	recipes := graph.ByType(frontmatter.DocRecipe)

	// System header from context doc
	contexts := graph.ByType(frontmatter.DocContext)
	if len(contexts) > 0 {
		ctx := contexts[0]
		header := ctx.Title
		if ctx.Frontmatter.Goal != "" {
			header += " — " + ctx.Frontmatter.Goal
		} else if ctx.Frontmatter.Summary != "" {
			header += " — " + ctx.Frontmatter.Summary
		}
		fmt.Fprintln(w, header)
	}

	// Architecture summary
	summaryParts := []string{
		plural(len(containers), "container"),
		plural(len(components), "component"),
		plural(len(refs), "ref"),
	}
	if includeADR {
		summaryParts = append(summaryParts, plural(len(adrs), "ADR"))
	}
	if len(recipes) > 0 {
		summaryParts = append(summaryParts, plural(len(recipes), "recipe"))
	}
	fmt.Fprintf(w, "%s\n\n", strings.Join(summaryParts, " · "))

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].ID < containers[j].ID
	})

	for _, container := range containers {
		line := fmt.Sprintf("%s-%s (container)", container.ID, container.Slug)
		if container.Frontmatter.Goal != "" {
			line += " — " + container.Frontmatter.Goal
		}
		fmt.Fprintln(w, line)

		comps := graph.Children(container.ID)
		sort.Slice(comps, func(i, j int) bool {
			return comps[i].ID < comps[j].ID
		})

		for i, comp := range comps {
			isLast := i == len(comps)-1
			prefix := "├── "
			indent := "│   "
			if isLast {
				prefix = "└── "
				indent = "    "
			}

			category := comp.Frontmatter.Category
			if category == "" {
				category = "foundation"
			}

			badge := ""
			if comp.Frontmatter.Status == "provisioning" {
				badge = " [provisioning]"
			}

			line := fmt.Sprintf("%s%s-%s (%s)%s", prefix, comp.ID, comp.Slug, category, badge)
			if comp.Frontmatter.Goal != "" {
				line += " — " + comp.Frontmatter.Goal
			}
			fmt.Fprintln(w, line)

			if !compact {
				// Files from codemap
				if files := cm[comp.ID]; len(files) > 0 {
					sorted := append([]string(nil), files...)
					sort.Strings(sorted)
					fmt.Fprintf(w, "%s  files: %s\n", indent, strings.Join(sorted, ", "))
				}

				// Refs used
				refsUsed := graph.RefsFor(comp.ID)
				if len(refsUsed) > 0 {
					sort.Slice(refsUsed, func(a, b int) bool {
						return refsUsed[a].ID < refsUsed[b].ID
					})
					var refIDs []string
					for _, r := range refsUsed {
						refIDs = append(refIDs, r.ID)
					}
					fmt.Fprintf(w, "%s  uses:  %s\n", indent, strings.Join(refIDs, ", "))
				}
			}
		}
		fmt.Fprintln(w)
	}

	// Cross-cutting refs
	if len(refs) > 0 {
		sort.Slice(refs, func(i, j int) bool {
			return refs[i].ID < refs[j].ID
		})
		fmt.Fprintln(w, "Cross-cutting:")
		for _, ref := range refs {
			line := fmt.Sprintf("  %s", ref.ID)
			if ref.Frontmatter.Goal != "" {
				line += " — " + ref.Frontmatter.Goal
			}
			fmt.Fprintln(w, line)

			// Citing components + aggregate file coverage
			citers := graph.CitedBy(ref.ID)
			var compCiters []*walker.C3Entity
			for _, c := range citers {
				if c.Type == frontmatter.DocComponent {
					compCiters = append(compCiters, c)
				}
			}
			sort.Slice(compCiters, func(i, j int) bool {
				return compCiters[i].ID < compCiters[j].ID
			})

			if len(compCiters) > 0 && !compact {
				var citerIDs []string
				fileSet := map[string]bool{}
				var fileList []string
				for _, c := range compCiters {
					citerIDs = append(citerIDs, c.ID)
					for _, f := range cm[c.ID] {
						if !fileSet[f] {
							fileSet[f] = true
							fileList = append(fileList, f)
						}
					}
				}
				sort.Strings(fileList)
				fmt.Fprintf(w, "    via:   %s\n", strings.Join(citerIDs, ", "))
				if len(fileList) > 0 {
					fmt.Fprintf(w, "    files: %s\n", strings.Join(fileList, ", "))
				}
			}
		}
		fmt.Fprintln(w)
	}

	// Recipes
	if len(recipes) > 0 {
		sort.Slice(recipes, func(i, j int) bool {
			return recipes[i].ID < recipes[j].ID
		})
		fmt.Fprintln(w, "Recipes:")
		for _, r := range recipes {
			desc := r.Frontmatter.Description
			if desc == "" {
				desc = r.Frontmatter.Goal
			}
			line := fmt.Sprintf("  %s", r.ID)
			if desc != "" {
				line += " — " + desc
			}
			fmt.Fprintln(w, line)

			if len(r.Frontmatter.Sources) > 0 && !compact {
				fmt.Fprintf(w, "    sources: %s\n", strings.Join(r.Frontmatter.Sources, ", "))
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}
