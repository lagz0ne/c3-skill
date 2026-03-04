package index

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// StructuralIndex is the precomputed index of a C3 graph + code-map.
type StructuralIndex struct {
	Hash     string                 `json:"hash"`
	Entities map[string]EntityEntry `json:"entities"`
	Files    map[string]FileEntry   `json:"files"`
	Refs     map[string]RefEntry    `json:"refs"`
}

// EntityEntry holds the projected data for a single C3 entity.
type EntityEntry struct {
	Title           string          `json:"title"`
	Type            string          `json:"type"`
	Container       string          `json:"container,omitempty"`
	Context         string          `json:"context,omitempty"`
	Refs            []string        `json:"refs,omitempty"`
	CitedBy         []string        `json:"cited_by,omitempty"`
	ReverseDeps     []string        `json:"reverse_deps,omitempty"`
	CodePaths       []string        `json:"code_paths,omitempty"`
	ConstraintsFrom []string        `json:"constraints_from,omitempty"`
	BlockFill       map[string]bool `json:"block_fill,omitempty"`
}

// FileEntry maps a code-map pattern to the entities and refs it covers.
type FileEntry struct {
	Entities []string `json:"entities"`
	Refs     []string `json:"refs,omitempty"`
}

// RefEntry holds which entities cite a ref and its scope.
type RefEntry struct {
	Citers []string `json:"citers"`
	Scope  []string `json:"scope,omitempty"`
}

// Build projects a C3 graph and code-map into a StructuralIndex.
func Build(graph *walker.C3Graph, cm codemap.CodeMap, c3Dir string) *StructuralIndex {
	idx := &StructuralIndex{
		Entities: make(map[string]EntityEntry),
		Files:    make(map[string]FileEntry),
		Refs:     make(map[string]RefEntry),
	}

	all := graph.All()

	for _, e := range all {
		docType := frontmatter.ClassifyDoc(e.Frontmatter)
		if docType == frontmatter.DocUnknown {
			continue
		}
		typeName := docType.String()

		entry := EntityEntry{
			Title: e.Title,
			Type:  typeName,
		}

		// Parent chain: container + context
		switch docType {
		case frontmatter.DocComponent:
			entry.Container = e.Frontmatter.Parent
			if container := graph.Get(e.Frontmatter.Parent); container != nil {
				entry.Context = container.Frontmatter.Parent
			}
		case frontmatter.DocContainer:
			entry.Context = e.Frontmatter.Parent
		}

		// Refs
		for _, ref := range graph.RefsFor(e.ID) {
			entry.Refs = append(entry.Refs, ref.ID)
		}

		// Reverse deps
		for _, rev := range graph.Reverse(e.ID) {
			entry.ReverseDeps = append(entry.ReverseDeps, rev.ID)
		}
		sort.Strings(entry.ReverseDeps)

		// Code paths from code-map
		if patterns, ok := cm[e.ID]; ok {
			entry.CodePaths = patterns
		}

		// Constraints from: parent chain + refs
		var constraints []string
		if entry.Context != "" {
			constraints = append(constraints, entry.Context)
		}
		if entry.Container != "" {
			constraints = append(constraints, entry.Container)
		}
		constraints = append(constraints, entry.Refs...)
		if len(constraints) > 0 {
			entry.ConstraintsFrom = constraints
		}

		entry.BlockFill = extractBlockFill(e.Body, typeName)

		idx.Entities[e.ID] = entry

		// Build ref entries for ref-type entities
		if docType == frontmatter.DocRef {
			citers := graph.CitedBy(e.ID)
			var citerIDs []string
			for _, c := range citers {
				citerIDs = append(citerIDs, c.ID)
			}
			sort.Strings(citerIDs)
			idx.Refs[e.ID] = RefEntry{
				Citers: citerIDs,
				Scope:  e.Frontmatter.Scope,
			}
		}
	}

	// File map: invert code-map (pattern → entity IDs + their refs)
	for id, patterns := range cm {
		if strings.HasPrefix(id, "_") {
			continue
		}
		if graph.Get(id) == nil {
			continue
		}
		entityRefs := idx.Entities[id].Refs

		for _, pattern := range patterns {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}
			fe := idx.Files[pattern]
			fe.Entities = appendUnique(fe.Entities, id)
			for _, r := range entityRefs {
				fe.Refs = appendUnique(fe.Refs, r)
			}
			idx.Files[pattern] = fe
		}
	}
	// Sort file entry slices for determinism
	for k, fe := range idx.Files {
		sort.Strings(fe.Entities)
		sort.Strings(fe.Refs)
		idx.Files[k] = fe
	}

	idx.Hash = computeHash(idx)
	return idx
}

// WriteMarkdown writes the structural index in human-readable markdown.
func WriteMarkdown(w io.Writer, idx *StructuralIndex) error {
	fmt.Fprintf(w, "# C3 Structural Index\n")
	fmt.Fprintf(w, "<!-- hash: %s -->\n\n", idx.Hash)

	// Sort entity IDs for deterministic output
	ids := make([]string, 0, len(idx.Entities))
	for id := range idx.Entities {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		e := idx.Entities[id]
		fmt.Fprintf(w, "## %s — %s (%s)\n", id, e.Title, e.Type)

		// Parent chain
		var chain []string
		if e.Container != "" {
			chain = append(chain, "container: "+e.Container)
		}
		if e.Context != "" {
			chain = append(chain, "context: "+e.Context)
		}
		if len(chain) > 0 {
			fmt.Fprintf(w, "%s\n", strings.Join(chain, " | "))
		}

		if len(e.Refs) > 0 {
			fmt.Fprintf(w, "refs: %s\n", strings.Join(e.Refs, ", "))
		}
		if len(e.ReverseDeps) > 0 {
			fmt.Fprintf(w, "reverse deps: %s\n", strings.Join(e.ReverseDeps, ", "))
		}
		if len(e.CodePaths) > 0 {
			fmt.Fprintf(w, "files: %s\n", strings.Join(e.CodePaths, ", "))
		}
		if len(e.ConstraintsFrom) > 0 {
			fmt.Fprintf(w, "constraints from: %s\n", strings.Join(e.ConstraintsFrom, ", "))
		}

		// Block fill
		if len(e.BlockFill) > 0 {
			var blockParts []string
			// Sort block names for determinism
			names := make([]string, 0, len(e.BlockFill))
			for name := range e.BlockFill {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				icon := "○"
				if e.BlockFill[name] {
					icon = "✓"
				}
				blockParts = append(blockParts, name+" "+icon)
			}
			fmt.Fprintf(w, "blocks: %s\n", strings.Join(blockParts, ", "))
		}

		fmt.Fprintln(w)
	}

	// File map
	if len(idx.Files) > 0 {
		fmt.Fprintln(w, "## File Map")
		patterns := make([]string, 0, len(idx.Files))
		for p := range idx.Files {
			patterns = append(patterns, p)
		}
		sort.Strings(patterns)

		for _, p := range patterns {
			fe := idx.Files[p]
			line := p + " → " + strings.Join(fe.Entities, ", ")
			if len(fe.Refs) > 0 {
				line += " | refs: " + strings.Join(fe.Refs, ", ")
			}
			fmt.Fprintln(w, line)
		}
		fmt.Fprintln(w)
	}

	// Ref map
	if len(idx.Refs) > 0 {
		fmt.Fprintln(w, "## Ref Map")
		refIDs := make([]string, 0, len(idx.Refs))
		for id := range idx.Refs {
			refIDs = append(refIDs, id)
		}
		sort.Strings(refIDs)

		for _, id := range refIDs {
			re := idx.Refs[id]
			line := id
			if len(re.Citers) > 0 {
				line += " cited by: " + strings.Join(re.Citers, ", ")
			}
			if len(re.Scope) > 0 {
				line += " | scope: " + strings.Join(re.Scope, ", ")
			}
			fmt.Fprintln(w, line)
		}
	}

	return nil
}

// WriteJSON writes the structural index as JSON.
func WriteJSON(w io.Writer, idx *StructuralIndex) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(idx)
}

// WriteTo writes the markdown index to .c3/_index/structural.md.
func WriteTo(c3Dir string, idx *StructuralIndex) error {
	dir := filepath.Join(c3Dir, "_index")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "structural.md"))
	if err != nil {
		return err
	}
	defer f.Close()
	return WriteMarkdown(f, idx)
}

// computeHash produces a deterministic SHA256 from sorted entity data.
func computeHash(idx *StructuralIndex) string {
	h := sha256.New()

	ids := make([]string, 0, len(idx.Entities))
	for id := range idx.Entities {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		e := idx.Entities[id]
		fmt.Fprintf(h, "%s|%s|%s|%s|%s|", id, e.Title, e.Type, e.Container, e.Context)
		fmt.Fprintf(h, "refs:%s|", strings.Join(e.Refs, ","))
		fmt.Fprintf(h, "code:%s|", strings.Join(e.CodePaths, ","))
		// Sort block names for deterministic hash
		blockNames := make([]string, 0, len(e.BlockFill))
		for name := range e.BlockFill {
			blockNames = append(blockNames, name)
		}
		sort.Strings(blockNames)
		for _, name := range blockNames {
			fmt.Fprintf(h, "block:%s=%v|", name, e.BlockFill[name])
		}
	}

	// Include file map
	patterns := make([]string, 0, len(idx.Files))
	for p := range idx.Files {
		patterns = append(patterns, p)
	}
	sort.Strings(patterns)
	for _, p := range patterns {
		fe := idx.Files[p]
		fmt.Fprintf(h, "file:%s→%s|", p, strings.Join(fe.Entities, ","))
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// extractBlockFill computes block fill status for an entity body using schema + markdown parsing.
func extractBlockFill(body, entityType string) map[string]bool {
	schemaSections := schema.ForType(entityType)
	if schemaSections == nil {
		return nil
	}

	bodySections := markdown.ParseSections(body)
	sectionMap := make(map[string]markdown.Section)
	for _, s := range bodySections {
		if s.Name != "" {
			sectionMap[s.Name] = s
		}
	}

	fill := make(map[string]bool, len(schemaSections))
	for _, def := range schemaSections {
		bodySection, exists := sectionMap[def.Name]
		if !exists {
			fill[def.Name] = false
			continue
		}
		content := strings.TrimSpace(bodySection.Content)
		if content == "" {
			fill[def.Name] = false
			continue
		}
		if def.ContentType == "table" {
			table, err := markdown.ParseTable(content)
			fill[def.Name] = err == nil && len(table.Rows) > 0
		} else {
			fill[def.Name] = true
		}
	}
	return fill
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
