package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ListOptions holds parameters for the list command.
type ListOptions struct {
	Store        *store.Store
	JSON         bool
	Flat         bool
	Compact      bool
	C3Dir        string
	IncludeADR   bool
	JSONExplicit bool
}

// ListResult wraps list output with a totalCount envelope.
type ListResult struct {
	TotalCount int         `json:"totalCount"`
	Entities   interface{} `json:"entities"`
}

// compactEntity is a minimal entity representation for compact/TOON output.
type compactEntity struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Parent string `json:"parent,omitempty"`
	Status string `json:"status,omitempty"`
}

// RunList outputs the topology of entities from the store.
func RunList(opts ListOptions, w io.Writer) error {
	if opts.Flat {
		return listFlat(opts.Store, opts.IncludeADR, w)
	}
	format := ResolveFormat(opts.JSONExplicit, isAgentMode())
	if opts.JSON || format == FormatTOON {
		return listStructured(opts, format, w)
	}
	return listTopology(opts.Store, opts.Compact, opts.IncludeADR, w)
}

func listStructured(opts ListOptions, format OutputFormat, w io.Writer) error {
	entities, err := opts.Store.AllEntities()
	if err != nil {
		return err
	}
	if !opts.IncludeADR {
		filtered := entities[:0]
		for _, e := range entities {
			if e.Type != "adr" {
				filtered = append(filtered, e)
			}
		}
		entities = filtered
	}
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	// Agent mode defaults to compact (AXI principle: minimal default schema)
	compact := opts.Compact || (isAgentMode() && !opts.JSONExplicit)

	hints := []HelpHint{
		{Command: "c3x read <id>", Description: "read entity content"},
		{Command: "c3x check", Description: "validate consistency"},
		{Command: "c3x graph <id> --format mermaid", Description: "visualize relationships"},
	}

	if compact {
		var result []compactEntity
		for _, e := range entities {
			result = append(result, compactEntity{
				ID: e.ID, Type: e.Type, Title: e.Title,
				Parent: e.ParentID, Status: e.Status,
			})
		}

		if format == FormatTOON {
			fmt.Fprintf(w, "totalCount: %d\n", len(result))
			return WriteTableOutput(w, "entities", result, []string{"id", "type", "title", "parent", "status"}, format, hints)
		}
		if opts.JSONExplicit {
			return writeJSON(w, ListResult{TotalCount: len(result), Entities: result})
		}
		// Legacy JSON path (JSON=true but not explicit) — plain array
		return writeJSON(w, result)
	}

	// Full JSON (non-compact)
	type jsonEntity struct {
		ID            string                 `json:"id"`
		Type          string                 `json:"type"`
		Title         string                 `json:"title"`
		Relationships []string               `json:"relationships"`
		Frontmatter   map[string]interface{} `json:"frontmatter"`
		Files         []string               `json:"files,omitempty"`
	}

	var data []jsonEntity
	for _, e := range entities {
		fm := make(map[string]interface{})
		if e.Goal != "" {
			fm["goal"] = e.Goal
		}
		if e.Category != "" {
			fm["category"] = e.Category
		}
		if e.ParentID != "" {
			fm["parent"] = e.ParentID
		}
		if e.Status != "" {
			fm["status"] = e.Status
		}
		if e.Boundary != "" {
			fm["boundary"] = e.Boundary
		}
		rels, _ := opts.Store.RelationshipsFrom(e.ID)
		var relationships []string
		relsByType := make(map[string][]string)
		for _, r := range rels {
			relationships = append(relationships, r.ToID)
			relsByType[r.RelType] = append(relsByType[r.RelType], r.ToID)
		}
		for _, rt := range []string{"uses", "affects", "scope", "sources"} {
			if ids := relsByType[rt]; len(ids) > 0 {
				fm[rt] = ids
			}
		}

		var files []string
		if f, _ := opts.Store.CodeMapFor(e.ID); len(f) > 0 {
			files = append([]string(nil), f...)
			sort.Strings(files)
		}

		data = append(data, jsonEntity{
			ID:            e.ID,
			Type:          e.Type,
			Title:         e.Title,
			Relationships: relationships,
			Frontmatter:   fm,
			Files:         files,
		})
	}

	if opts.JSONExplicit {
		return writeJSON(w, ListResult{TotalCount: len(data), Entities: data})
	}
	// Legacy JSON path — plain array
	return writeJSON(w, data)
}

func listFlat(s *store.Store, includeADR bool, w io.Writer) error {
	entities, err := s.AllEntities()
	if err != nil {
		return err
	}
	if !includeADR {
		filtered := entities[:0]
		for _, e := range entities {
			if e.Type != "adr" {
				filtered = append(filtered, e)
			}
		}
		entities = filtered
	}
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, e := range entities {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Type, e.ID)
	}
	return nil
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

func listTopology(s *store.Store, compact bool, includeADR bool, w io.Writer) error {
	containers, _ := s.EntitiesByType("container")
	components, _ := s.EntitiesByType("component")
	refs, _ := s.EntitiesByType("ref")
	adrs, _ := s.EntitiesByType("adr")
	recipes, _ := s.EntitiesByType("recipe")
	rules, _ := s.EntitiesByType("rule")

	// System header from context doc
	contexts, _ := s.EntitiesByType("system")
	if len(contexts) > 0 {
		ctx := contexts[0]
		header := ctx.Title
		if ctx.Goal != "" {
			header += " — " + ctx.Goal
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
	if len(rules) > 0 {
		summaryParts = append(summaryParts, plural(len(rules), "rule"))
	}
	fmt.Fprintf(w, "%s\n\n", strings.Join(summaryParts, " · "))

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].ID < containers[j].ID
	})

	for _, container := range containers {
		line := fmt.Sprintf("%s-%s (container)", container.ID, container.Slug)
		if container.Goal != "" {
			line += " — " + container.Goal
		}
		fmt.Fprintln(w, line)

		comps, _ := s.Children(container.ID)
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

			category := comp.Category
			if category == "" {
				category = "foundation"
			}

			badge := ""
			if comp.Status == "provisioning" {
				badge = " [provisioning]"
			}

			line := fmt.Sprintf("%s%s-%s (%s)%s", prefix, comp.ID, comp.Slug, category, badge)
			if comp.Goal != "" {
				line += " — " + comp.Goal
			}
			fmt.Fprintln(w, line)

			if !compact {
				// Files from codemap
				if files, _ := s.CodeMapFor(comp.ID); len(files) > 0 {
					sorted := append([]string(nil), files...)
					sort.Strings(sorted)
					fmt.Fprintf(w, "%s  files: %s\n", indent, strings.Join(sorted, ", "))
				}

				// Refs used
				refsUsed, _ := s.RefsFor(comp.ID)
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
			if ref.Goal != "" {
				line += " — " + ref.Goal
			}
			fmt.Fprintln(w, line)

			// Citing components + aggregate file coverage
			citers, _ := s.CitedBy(ref.ID)
			var compCiters []*store.Entity
			for _, c := range citers {
				if c.Type == "component" {
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
					if f, _ := s.CodeMapFor(c.ID); len(f) > 0 {
						for _, file := range f {
							if !fileSet[file] {
								fileSet[file] = true
								fileList = append(fileList, file)
							}
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

	// Coding Rules
	if len(rules) > 0 {
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].ID < rules[j].ID
		})
		fmt.Fprintln(w, "Coding Rules:")
		for _, rule := range rules {
			line := fmt.Sprintf("  %s", rule.ID)
			if rule.Goal != "" {
				line += " — " + rule.Goal
			}
			fmt.Fprintln(w, line)

			citers, _ := s.CitedBy(rule.ID)
			var compCiters []*store.Entity
			for _, c := range citers {
				if c.Type == "component" {
					compCiters = append(compCiters, c)
				}
			}
			sort.Slice(compCiters, func(i, j int) bool {
				return compCiters[i].ID < compCiters[j].ID
			})

			if len(compCiters) > 0 && !compact {
				var citerIDs []string
				for _, c := range compCiters {
					citerIDs = append(citerIDs, c.ID)
				}
				fmt.Fprintf(w, "    enforced on: %s\n", strings.Join(citerIDs, ", "))
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
			line := fmt.Sprintf("  %s", r.ID)
			if r.Goal != "" {
				line += " — " + r.Goal
			}
			fmt.Fprintln(w, line)

			// Sources from relationships
			if !compact {
				rels, _ := s.RelationshipsFrom(r.ID)
				var sourceIDs []string
				for _, rel := range rels {
					if rel.RelType == "sources" {
						sourceIDs = append(sourceIDs, rel.ToID)
					}
				}
				if len(sourceIDs) > 0 {
					fmt.Fprintf(w, "    sources: %s\n", strings.Join(sourceIDs, ", "))
				}
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}
