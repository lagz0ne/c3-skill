package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// GraphOptions holds parameters for the graph command.
type GraphOptions struct {
	Store     *store.Store
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
	if _, err := opts.Store.GetEntity(opts.EntityID); err != nil {
		return fmt.Errorf("entity %q not found", opts.EntityID)
	}

	if opts.Direction != "" && opts.Direction != "forward" && opts.Direction != "reverse" {
		return fmt.Errorf("--direction must be 'forward' or 'reverse', got %q", opts.Direction)
	}

	entities := collectSubgraphStore(opts.Store, opts.EntityID, opts.Depth, opts.Direction)

	if opts.JSON {
		if err := graphJSONStore(entities, opts.Store, w); err != nil {
			return err
		}
		return nil
	}
	if opts.Format == "mermaid" {
		if err := graphMermaidStore(entities, opts.Store, w); err != nil {
			return err
		}
		writeAgentHints(w, cascadeHintsForID(opts.Store, opts.EntityID))
		return nil
	}
	if err := graphTextStore(entities, opts.Store, w); err != nil {
		return err
	}
	writeAgentHints(w, cascadeHintsForID(opts.Store, opts.EntityID))
	return nil
}

// collectSubgraphStore returns entities reachable within depth hops from rootID.
func collectSubgraphStore(s *store.Store, rootID string, depth int, direction string) []*store.Entity {
	visited := map[string]bool{rootID: true}
	var result []*store.Entity
	root, err := s.GetEntity(rootID)
	if err == nil {
		result = append(result, root)
	}

	frontier := []string{rootID}
	for d := 0; d < depth && len(frontier) > 0; d++ {
		var next []string
		for _, id := range frontier {
			neighbors := graphNeighborsStore(s, id, direction)
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

// graphNeighborsStore collects connected entities based on direction.
func graphNeighborsStore(s *store.Store, id string, direction string) []*store.Entity {
	seen := make(map[string]bool)
	var result []*store.Entity

	add := func(entity *store.Entity) {
		if entity != nil && !seen[entity.ID] {
			seen[entity.ID] = true
			result = append(result, entity)
		}
	}

	if direction == "forward" || direction == "" {
		// Children
		children, _ := s.Children(id)
		for _, c := range children {
			add(c)
		}
		// Outbound affects
		rels, _ := s.RelationshipsFrom(id)
		for _, r := range rels {
			if r.RelType == "affects" {
				if e, err := s.GetEntity(r.ToID); err == nil {
					add(e)
				}
			}
		}
		// For refs/rules: entities that cite this one
		entity, err := s.GetEntity(id)
		if err == nil && (entity.Type == "ref" || entity.Type == "rule") {
			citers, _ := s.CitedBy(id)
			for _, c := range citers {
				add(c)
			}
		}
	}

	if direction == "reverse" || direction == "" {
		// Inbound relationships
		inbound, _ := s.RelationshipsTo(id)
		for _, r := range inbound {
			if e, err := s.GetEntity(r.FromID); err == nil {
				add(e)
			}
		}
		// Parent
		entity, err := s.GetEntity(id)
		if err == nil && entity.ParentID != "" {
			if p, err := s.GetEntity(entity.ParentID); err == nil {
				add(p)
			}
		}
	}

	if direction == "" {
		rels, _ := s.RelationshipsFrom(id)
		for _, r := range rels {
			if r.RelType == "uses" || r.RelType == "scope" {
				if e, err := s.GetEntity(r.ToID); err == nil {
					add(e)
				}
			}
		}
	}

	return result
}

// graphTextStore emits the adjacency-list text format.
func graphTextStore(entities []*store.Entity, s *store.Store, w io.Writer) error {
	for i, e := range entities {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s (%s) — %s\n", e.ID, e.Type, e.Title)

		// Parent
		if e.ParentID != "" {
			fmt.Fprintf(w, "  parent: %s\n", e.ParentID)
		}

		// Children
		children, _ := s.Children(e.ID)
		if len(children) > 0 {
			sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })
			ids := make([]string, len(children))
			for i, c := range children {
				ids[i] = c.ID
			}
			fmt.Fprintf(w, "  children: %s\n", strings.Join(ids, ", "))
		}

		// Uses (refs cited by this entity)
		rels, _ := s.RelationshipsFrom(e.ID)
		var usesIDs []string
		var affectsIDs []string
		for _, r := range rels {
			if r.RelType == "uses" {
				usesIDs = append(usesIDs, r.ToID)
			}
			if r.RelType == "affects" {
				affectsIDs = append(affectsIDs, r.ToID)
			}
		}
		if len(usesIDs) > 0 {
			sort.Strings(usesIDs)
			fmt.Fprintf(w, "  uses: %s\n", strings.Join(usesIDs, ", "))
		}

		// Cited-by (entities that cite this ref or rule)
		if e.Type == "ref" || e.Type == "rule" {
			citers, _ := s.CitedBy(e.ID)
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
		if len(affectsIDs) > 0 {
			sort.Strings(affectsIDs)
			fmt.Fprintf(w, "  affects: %s\n", strings.Join(affectsIDs, ", "))
		}

		// Files from code-map
		if files, _ := s.CodeMapFor(e.ID); len(files) > 0 {
			sorted := append([]string(nil), files...)
			sort.Strings(sorted)
			fmt.Fprintf(w, "  files: %s\n", strings.Join(sorted, ", "))
		}
	}
	return nil
}

// graphJSONStore emits the subgraph as JSON.
func graphJSONStore(entities []*store.Entity, s *store.Store, w io.Writer) error {
	nodes := make([]graphNode, 0, len(entities))
	for _, e := range entities {
		node := graphNode{
			ID:    e.ID,
			Type:  e.Type,
			Title: e.Title,
		}

		if e.ParentID != "" {
			node.Parent = e.ParentID
		}

		children, _ := s.Children(e.ID)
		if len(children) > 0 {
			sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })
			for _, c := range children {
				node.Children = append(node.Children, c.ID)
			}
		}

		rels, _ := s.RelationshipsFrom(e.ID)
		for _, r := range rels {
			if r.RelType == "uses" {
				node.Refs = append(node.Refs, r.ToID)
			}
			if r.RelType == "affects" {
				node.Affects = append(node.Affects, r.ToID)
			}
		}
		if len(node.Refs) > 0 {
			sort.Strings(node.Refs)
		}
		if len(node.Affects) > 0 {
			sort.Strings(node.Affects)
		}

		if e.Type == "ref" || e.Type == "rule" {
			citers, _ := s.CitedBy(e.ID)
			if len(citers) > 0 {
				sort.Slice(citers, func(i, j int) bool { return citers[i].ID < citers[j].ID })
				for _, c := range citers {
					node.CitedBy = append(node.CitedBy, c.ID)
				}
			}
		}

		if files, _ := s.CodeMapFor(e.ID); len(files) > 0 {
			node.Files = append([]string(nil), files...)
			sort.Strings(node.Files)
		}

		nodes = append(nodes, node)
	}
	if isAgentMode() {
		return writeJSON(w, struct {
			Nodes []graphNode `json:"nodes"`
			Help  []HelpHint  `json:"help,omitempty"`
		}{
			Nodes: nodes,
			Help:  agentHints(cascadeReviewHints()),
		})
	}
	return writeJSON(w, nodes)
}

// graphMermaidStore emits the subgraph as a Mermaid flowchart.
func graphMermaidStore(entities []*store.Entity, s *store.Store, w io.Writer) error {
	fmt.Fprintln(w, "graph TD")

	// Collect entities by container for subgraph grouping
	containerChildren := make(map[string][]*store.Entity)
	entitySet := make(map[string]bool)
	var containers []*store.Entity
	var standaloneNodes []*store.Entity

	for _, e := range entities {
		entitySet[e.ID] = true
	}

	for _, e := range entities {
		switch e.Type {
		case "container":
			containers = append(containers, e)
		case "component":
			if entitySet[e.ParentID] {
				containerChildren[e.ParentID] = append(containerChildren[e.ParentID], e)
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

	// Emit standalone nodes
	sort.Slice(standaloneNodes, func(i, j int) bool { return standaloneNodes[i].ID < standaloneNodes[j].ID })
	for _, e := range standaloneNodes {
		mID := mermaidSanitize(e.ID)
		switch e.Type {
		case "ref":
			fmt.Fprintf(w, "  %s([\"%s\"])\n", mID, mermaidEscape(e.Title))
		case "rule":
			fmt.Fprintf(w, "  %s{{\"%s\"}}\n", mID, mermaidEscape(e.Title))
		default:
			fmt.Fprintf(w, "  %s[\"%s\"]\n", mID, mermaidEscape(e.Title))
		}
	}

	// Emit edges
	for _, e := range entities {
		srcID := mermaidSanitize(e.ID)

		// Parent-child edges (only if not component inside subgraph)
		if e.Type != "component" && e.ParentID != "" && entitySet[e.ParentID] {
			fmt.Fprintf(w, "  %s --> %s\n", mermaidSanitize(e.ParentID), srcID)
		}

		// Ref citations (dashed)
		rels, _ := s.RelationshipsFrom(e.ID)
		for _, r := range rels {
			if r.RelType == "uses" && entitySet[r.ToID] {
				fmt.Fprintf(w, "  %s -.->|cites| %s\n", srcID, mermaidSanitize(r.ToID))
			}
			if r.RelType == "affects" && entitySet[r.ToID] {
				fmt.Fprintf(w, "  %s -->|affects| %s\n", srcID, mermaidSanitize(r.ToID))
			}
		}
	}

	return nil
}

// mermaidSanitize converts a C3 ID to a valid Mermaid node ID.
func mermaidSanitize(id string) string {
	return id
}

// mermaidEscape escapes double quotes in Mermaid labels.
func mermaidEscape(s string) string {
	return strings.ReplaceAll(s, "\"", "#quot;")
}
