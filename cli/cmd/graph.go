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

// GraphOptions holds parameters for the graph command.
type GraphOptions struct {
	Graph     *walker.C3Graph
	EntityID  string
	Depth     int
	Direction string // "forward" or "reverse"
	Format    string // "" (text) or "mermaid"
	JSON      bool
	C3Dir     string
}

// graphNode is a single entity in the subgraph output.
type graphNode struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Parent   string   `json:"parent,omitempty"`
	Children []string `json:"children,omitempty"`
	Refs     []string `json:"refs,omitempty"`
	CitedBy  []string `json:"cited_by,omitempty"`
	Affects  []string `json:"affects,omitempty"`
	Files    []string `json:"files,omitempty"`
}

// RunGraph emits a subgraph rooted at the given entity.
func RunGraph(opts GraphOptions, w io.Writer) error {
	root := opts.Graph.Get(opts.EntityID)
	if root == nil {
		return fmt.Errorf("entity %q not found", opts.EntityID)
	}

	if opts.Direction != "" && opts.Direction != "forward" && opts.Direction != "reverse" {
		return fmt.Errorf("--direction must be 'forward' or 'reverse', got %q", opts.Direction)
	}

	// Load code-map for file paths
	var cm codemap.CodeMap
	if opts.C3Dir != "" {
		var err error
		cm, err = codemap.ParseCodeMap(filepath.Join(opts.C3Dir, "code-map.yaml"))
		if err != nil {
			cm = codemap.CodeMap{}
		}
	}

	// Collect subgraph: root + transitive reachable entities
	entities := collectSubgraph(opts.Graph, opts.EntityID, opts.Depth, opts.Direction)

	if opts.JSON {
		return graphJSON(entities, opts.Graph, cm, w)
	}
	if opts.Format == "mermaid" {
		return graphMermaid(entities, opts.Graph, cm, w)
	}
	return graphText(entities, opts.Graph, cm, w)
}

// collectSubgraph returns the root entity plus all entities reachable within depth hops.
// Direction "forward" uses Forward(), "reverse" uses Reverse(), default collects all neighbors.
func collectSubgraph(graph *walker.C3Graph, rootID string, depth int, direction string) []*walker.C3Entity {
	visited := map[string]bool{rootID: true}
	var result []*walker.C3Entity
	root := graph.Get(rootID)
	if root != nil {
		result = append(result, root)
	}

	frontier := []string{rootID}
	for d := 0; d < depth && len(frontier) > 0; d++ {
		var next []string
		for _, id := range frontier {
			neighbors := graphNeighbors(graph, id, direction)
			for _, e := range neighbors {
				if !visited[e.ID] {
					visited[e.ID] = true
					result = append(result, e)
					next = append(next, e.ID)
				}
			}
		}
		frontier = next
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// graphNeighbors collects connected entities based on direction.
// "forward": only Forward() edges (impact analysis — children, affects, cited-by for refs)
// "reverse": only Reverse() edges (what points to me)
// default: all neighbors (Forward + Reverse + explicit references from frontmatter)
func graphNeighbors(graph *walker.C3Graph, id string, direction string) []*walker.C3Entity {
	if direction == "forward" {
		return graph.Forward(id)
	}
	if direction == "reverse" {
		return graph.Reverse(id)
	}

	// Default: collect all connected entities
	seen := make(map[string]bool)
	var result []*walker.C3Entity

	add := func(entities []*walker.C3Entity) {
		for _, e := range entities {
			if !seen[e.ID] {
				seen[e.ID] = true
				result = append(result, e)
			}
		}
	}

	add(graph.Forward(id))
	add(graph.Reverse(id))

	// Explicit frontmatter references (parent, refs, scope) that Forward/Reverse may miss
	entity := graph.Get(id)
	if entity != nil {
		if entity.Frontmatter.Parent != "" {
			if p := graph.Get(entity.Frontmatter.Parent); p != nil && !seen[p.ID] {
				seen[p.ID] = true
				result = append(result, p)
			}
		}
		for _, refID := range entity.Frontmatter.Refs {
			if r := graph.Get(refID); r != nil && !seen[r.ID] {
				seen[r.ID] = true
				result = append(result, r)
			}
		}
		for _, scopeID := range entity.Frontmatter.Scope {
			if s := graph.Get(scopeID); s != nil && !seen[s.ID] {
				seen[s.ID] = true
				result = append(result, s)
			}
		}
	}

	return result
}

// graphText emits the adjacency-list text format.
func graphText(entities []*walker.C3Entity, graph *walker.C3Graph, cm codemap.CodeMap, w io.Writer) error {
	for i, e := range entities {
		if i > 0 {
			fmt.Fprintln(w)
		}
		docType := frontmatter.ClassifyDoc(e.Frontmatter)
		fmt.Fprintf(w, "%s (%s) — %s\n", e.ID, docType.String(), e.Title)

		// Parent
		if e.Frontmatter.Parent != "" {
			fmt.Fprintf(w, "  parent: %s\n", e.Frontmatter.Parent)
		}

		// Children
		children := graph.Children(e.ID)
		if len(children) > 0 {
			sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })
			ids := make([]string, len(children))
			for i, c := range children {
				ids[i] = c.ID
			}
			fmt.Fprintf(w, "  children: %s\n", strings.Join(ids, ", "))
		}

		// Refs (cited by this entity)
		if len(e.Frontmatter.Refs) > 0 {
			sorted := append([]string(nil), e.Frontmatter.Refs...)
			sort.Strings(sorted)
			fmt.Fprintf(w, "  refs: %s\n", strings.Join(sorted, ", "))
		}

		// Cited-by (entities that cite this ref)
		if docType == frontmatter.DocRef {
			citers := graph.CitedBy(e.ID)
			if len(citers) > 0 {
				sort.Slice(citers, func(i, j int) bool { return citers[i].ID < citers[j].ID })
				ids := make([]string, len(citers))
				for i, c := range citers {
					ids[i] = c.ID
				}
				fmt.Fprintf(w, "  cited-by: %s\n", strings.Join(ids, ", "))
			}
		}

		// Affects
		if len(e.Frontmatter.Affects) > 0 {
			sorted := append([]string(nil), e.Frontmatter.Affects...)
			sort.Strings(sorted)
			fmt.Fprintf(w, "  affects: %s\n", strings.Join(sorted, ", "))
		}

		// Files from code-map
		if files := cm[e.ID]; len(files) > 0 {
			sorted := append([]string(nil), files...)
			sort.Strings(sorted)
			fmt.Fprintf(w, "  files: %s\n", strings.Join(sorted, ", "))
		}
	}
	return nil
}

// graphJSON emits the subgraph as JSON.
func graphJSON(entities []*walker.C3Entity, graph *walker.C3Graph, cm codemap.CodeMap, w io.Writer) error {
	nodes := make([]graphNode, 0, len(entities))
	for _, e := range entities {
		docType := frontmatter.ClassifyDoc(e.Frontmatter)
		node := graphNode{
			ID:    e.ID,
			Type:  docType.String(),
			Title: e.Title,
		}

		if e.Frontmatter.Parent != "" {
			node.Parent = e.Frontmatter.Parent
		}

		children := graph.Children(e.ID)
		if len(children) > 0 {
			sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })
			for _, c := range children {
				node.Children = append(node.Children, c.ID)
			}
		}

		if len(e.Frontmatter.Refs) > 0 {
			node.Refs = append([]string(nil), e.Frontmatter.Refs...)
			sort.Strings(node.Refs)
		}

		if docType == frontmatter.DocRef {
			citers := graph.CitedBy(e.ID)
			if len(citers) > 0 {
				sort.Slice(citers, func(i, j int) bool { return citers[i].ID < citers[j].ID })
				for _, c := range citers {
					node.CitedBy = append(node.CitedBy, c.ID)
				}
			}
		}

		if len(e.Frontmatter.Affects) > 0 {
			node.Affects = append([]string(nil), e.Frontmatter.Affects...)
			sort.Strings(node.Affects)
		}

		if files := cm[e.ID]; len(files) > 0 {
			node.Files = append([]string(nil), files...)
			sort.Strings(node.Files)
		}

		nodes = append(nodes, node)
	}
	return writeJSON(w, nodes)
}

// graphMermaid emits the subgraph as a Mermaid flowchart.
func graphMermaid(entities []*walker.C3Entity, graph *walker.C3Graph, cm codemap.CodeMap, w io.Writer) error {
	fmt.Fprintln(w, "graph TD")

	// Collect entities by container for subgraph grouping
	containerChildren := make(map[string][]*walker.C3Entity)
	entitySet := make(map[string]bool)
	var containers []*walker.C3Entity
	var standaloneNodes []*walker.C3Entity

	for _, e := range entities {
		entitySet[e.ID] = true
	}

	for _, e := range entities {
		docType := frontmatter.ClassifyDoc(e.Frontmatter)
		switch docType {
		case frontmatter.DocContainer:
			containers = append(containers, e)
		case frontmatter.DocComponent:
			if entitySet[e.Frontmatter.Parent] {
				containerChildren[e.Frontmatter.Parent] = append(containerChildren[e.Frontmatter.Parent], e)
			} else {
				standaloneNodes = append(standaloneNodes, e)
			}
		default:
			standaloneNodes = append(standaloneNodes, e)
		}
	}

	// Emit containers as subgraphs
	sort.Slice(containers, func(i, j int) bool { return containers[i].ID < containers[j].ID })
	for _, c := range containers {
		mermaidID := mermaidSanitize(c.ID)
		fmt.Fprintf(w, "  subgraph %s[\"%s\"]\n", mermaidID, mermaidEscape(c.Title))
		children := containerChildren[c.ID]
		sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })
		for _, child := range children {
			childID := mermaidSanitize(child.ID)
			fmt.Fprintf(w, "    %s[\"%s\"]\n", childID, mermaidEscape(child.Title))
		}
		fmt.Fprintln(w, "  end")
	}

	// Emit standalone nodes (refs, context, orphaned components)
	sort.Slice(standaloneNodes, func(i, j int) bool { return standaloneNodes[i].ID < standaloneNodes[j].ID })
	for _, e := range standaloneNodes {
		mID := mermaidSanitize(e.ID)
		docType := frontmatter.ClassifyDoc(e.Frontmatter)
		if docType == frontmatter.DocRef {
			fmt.Fprintf(w, "  %s([\"%s\"])\n", mID, mermaidEscape(e.Title))
		} else {
			fmt.Fprintf(w, "  %s[\"%s\"]\n", mID, mermaidEscape(e.Title))
		}
	}

	// Emit edges
	for _, e := range entities {
		srcID := mermaidSanitize(e.ID)
		docType := frontmatter.ClassifyDoc(e.Frontmatter)

		// Parent-child edges (only if both in subgraph and not already shown via subgraph nesting)
		if docType != frontmatter.DocComponent && e.Frontmatter.Parent != "" && entitySet[e.Frontmatter.Parent] {
			fmt.Fprintf(w, "  %s --> %s\n", mermaidSanitize(e.Frontmatter.Parent), srcID)
		}

		// Ref citations (dashed)
		for _, refID := range e.Frontmatter.Refs {
			if entitySet[refID] {
				fmt.Fprintf(w, "  %s -.->|cites| %s\n", srcID, mermaidSanitize(refID))
			}
		}

		// Affects edges
		for _, affID := range e.Frontmatter.Affects {
			if entitySet[affID] {
				fmt.Fprintf(w, "  %s -->|affects| %s\n", srcID, mermaidSanitize(affID))
			}
		}
	}

	return nil
}

// mermaidSanitize converts a C3 ID to a valid Mermaid node ID.
// Mermaid allows hyphens in node IDs, so we just need to handle edge cases.
func mermaidSanitize(id string) string {
	return id
}

// mermaidEscape escapes double quotes in Mermaid labels.
func mermaidEscape(s string) string {
	return strings.ReplaceAll(s, "\"", "#quot;")
}
