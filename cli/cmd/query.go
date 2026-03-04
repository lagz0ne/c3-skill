package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/blocks"
	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// QueryOptions holds parameters for the query command.
type QueryOptions struct {
	Graph *walker.C3Graph
	C3Dir string
	Args  []string
	JSON  bool
	Chain bool
}

// BlockSummary is the JSON representation of a block in query output.
type BlockSummary struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Purpose string `json:"purpose"`
	Filled  bool   `json:"filled"`
	Summary string `json:"summary,omitempty"`
	Content any    `json:"content,omitempty"`
}

// EntityResult is the JSON representation of a C3 entity with its blocks.
type EntityResult struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Type   string         `json:"type"`
	Blocks []BlockSummary `json:"blocks"`
}

// ChainResult is the JSON output for --chain mode.
type ChainResult struct {
	Component *EntityResult  `json:"component,omitempty"`
	Container *EntityResult  `json:"container,omitempty"`
	Context   *EntityResult  `json:"context,omitempty"`
	Refs      []EntityResult `json:"refs"`
}

// RunQuery dispatches to the appropriate query mode.
func RunQuery(opts QueryOptions, w io.Writer) error {
	if len(opts.Args) == 0 {
		if opts.Chain {
			return fmt.Errorf("--chain requires an entity ID or file path")
		}
		return runQueryCatalog(opts, w)
	}

	target := opts.Args[0]

	// Try entity ID first (graph lookup), fall back to file path resolution.
	// This avoids misclassifying dotted IDs (e.g. "ref-jwt.v2") as file paths.
	entity := opts.Graph.Get(target)
	if entity != nil {
		if opts.Chain {
			return runQueryChain(opts, entity, w)
		}
		if len(opts.Args) >= 2 {
			return runQuerySingleBlock(opts, entity, w)
		}
		return runQuerySnapshot(opts, entity, w)
	}

	// Not an entity ID — treat as file path
	return runQueryFilePath(opts, w)
}

// --- Catalog mode ---

func runQueryCatalog(opts QueryOptions, w io.Writer) error {
	entities := opts.Graph.All()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	var entries []EntityResult
	for _, e := range entities {
		docType := frontmatter.ClassifyDoc(e.Frontmatter)
		// ADRs are change records, not standing architecture — skip in catalog
		if docType == frontmatter.DocADR {
			continue
		}
		typeName := docType.String()
		blks := blocks.ExtractBlocks(e.Body, typeName)
		if blks == nil {
			continue
		}

		entry := EntityResult{
			ID:    e.ID,
			Title: e.Title,
			Type:  typeName,
		}
		for _, b := range blks {
			bs := blockToSummary(b)
			bs.Summary = summarizeBlock(b)
			entry.Blocks = append(entry.Blocks, bs)
		}
		entries = append(entries, entry)
	}

	if opts.JSON {
		return writeJSON(w, entries)
	}

	for _, entry := range entries {
		fmt.Fprintf(w, "%s (%s) — %s\n", entry.ID, entry.Type, entry.Title)
		for _, b := range entry.Blocks {
			icon := "○"
			if b.Filled {
				icon = "✓"
			}
			line := fmt.Sprintf("  %s %s", icon, b.Name)
			if b.Summary != "" {
				line += " — " + b.Summary
			}
			fmt.Fprintln(w, line)
		}
		fmt.Fprintln(w)
	}
	return nil
}

// --- Entity snapshot ---

func runQuerySnapshot(opts QueryOptions, entity *walker.C3Entity, w io.Writer) error {
	if opts.JSON {
		return writeJSON(w, entityToResult(entity))
	}

	typeName := frontmatter.ClassifyDoc(entity.Frontmatter).String()
	blks := blocks.ExtractBlocks(entity.Body, typeName)

	fmt.Fprintf(w, "%s (%s) — %s\n\n", entity.ID, typeName, entity.Title)
	for _, b := range blks {
		icon := "○"
		if b.Filled {
			icon = "✓"
		}
		fmt.Fprintf(w, "%s %s", icon, b.Name)
		if b.Purpose != "" {
			fmt.Fprintf(w, "  [%s]", b.Purpose)
		}
		fmt.Fprintln(w)
		if b.Filled {
			printBlockContent(w, b.Content, "  ")
		}
		fmt.Fprintln(w)
	}
	return nil
}

// --- Single block ---

func runQuerySingleBlock(opts QueryOptions, entity *walker.C3Entity, w io.Writer) error {
	sectionSlug := opts.Args[1]
	typeName := frontmatter.ClassifyDoc(entity.Frontmatter).String()

	b := blocks.ExtractBlock(entity.Body, typeName, sectionSlug)
	if b == nil {
		return fmt.Errorf("section %q not found for %s", sectionSlug, typeName)
	}

	if opts.JSON {
		return writeJSON(w, b.Content)
	}

	printBlockContent(w, b.Content, "")
	return nil
}

// --- File resolution ---

func runQueryFilePath(opts QueryOptions, w io.Writer) error {
	filePath := opts.Args[0]
	cmPath := filepath.Join(opts.C3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err != nil {
		return fmt.Errorf("code-map parse error: %w", err)
	}

	ids := codemap.Match(cm, filePath)
	if len(ids) == 0 {
		return fmt.Errorf("no component mapping found for %s", filePath)
	}

	// Multiple matches: snapshot each entity
	if len(ids) > 1 && !opts.Chain {
		return runQueryMultiMatch(opts, ids, w)
	}

	entity := opts.Graph.Get(ids[0])
	if entity == nil {
		return fmt.Errorf("entity %s from code-map not found in graph", ids[0])
	}

	if opts.Chain {
		return runQueryChain(opts, entity, w)
	}
	return runQuerySnapshot(opts, entity, w)
}

// runQueryMultiMatch handles file paths that map to multiple entities.
func runQueryMultiMatch(opts QueryOptions, ids []string, w io.Writer) error {
	if opts.JSON {
		var snaps []EntityResult
		for _, id := range ids {
			entity := opts.Graph.Get(id)
			if entity == nil {
				continue
			}
			snaps = append(snaps, entityToResult(entity))
		}
		return writeJSON(w, snaps)
	}

	fmt.Fprintf(w, "file maps to %d entities: %s\n\n", len(ids), strings.Join(ids, ", "))
	for _, id := range ids {
		entity := opts.Graph.Get(id)
		if entity == nil {
			continue
		}
		if err := runQuerySnapshot(opts, entity, w); err != nil {
			return err
		}
	}
	return nil
}

// --- Chain traversal ---

func runQueryChain(opts QueryOptions, entity *walker.C3Entity, w io.Writer) error {
	chain := buildChain(opts.Graph, entity)

	if opts.JSON {
		return writeJSON(w, chain)
	}

	if chain.Context != nil {
		printChainLevel(w, "context", *chain.Context)
	}
	if chain.Container != nil {
		printChainLevel(w, "container", *chain.Container)
	}
	if chain.Component != nil {
		printChainLevel(w, "component", *chain.Component)
	}
	if len(chain.Refs) > 0 {
		fmt.Fprintln(w, "refs:")
		for _, ref := range chain.Refs {
			printChainLevel(w, "  ref", ref)
		}
	}
	return nil
}

func buildChain(graph *walker.C3Graph, entity *walker.C3Entity) ChainResult {
	result := ChainResult{Refs: []EntityResult{}}

	docType := frontmatter.ClassifyDoc(entity.Frontmatter)

	switch docType {
	case frontmatter.DocComponent:
		result.Component = entityToResultPtr(entity)
		if parent := graph.Get(entity.Frontmatter.Parent); parent != nil {
			result.Container = entityToResultPtr(parent)
			if ctx := graph.Get(parent.Frontmatter.Parent); ctx != nil {
				result.Context = entityToResultPtr(ctx)
			}
		}
	case frontmatter.DocContainer:
		result.Container = entityToResultPtr(entity)
		if ctx := graph.Get(entity.Frontmatter.Parent); ctx != nil {
			result.Context = entityToResultPtr(ctx)
		}
	case frontmatter.DocContext:
		result.Context = entityToResultPtr(entity)
	case frontmatter.DocRef:
		result.Context = entityToResultPtr(entity)
	}

	refs := graph.RefsFor(entity.ID)
	for _, ref := range refs {
		result.Refs = append(result.Refs, entityToResult(ref))
	}

	return result
}

func entityToResult(entity *walker.C3Entity) EntityResult {
	typeName := frontmatter.ClassifyDoc(entity.Frontmatter).String()
	blks := blocks.ExtractBlocks(entity.Body, typeName)

	result := EntityResult{
		ID:    entity.ID,
		Title: entity.Title,
		Type:  typeName,
	}
	for _, b := range blks {
		result.Blocks = append(result.Blocks, blockToSummary(b))
	}
	return result
}

func entityToResultPtr(entity *walker.C3Entity) *EntityResult {
	r := entityToResult(entity)
	return &r
}

func printChainLevel(w io.Writer, label string, level EntityResult) {
	fmt.Fprintf(w, "%s: %s (%s) — %s\n", label, level.ID, level.Type, level.Title)
	for _, b := range level.Blocks {
		if b.Filled {
			fmt.Fprintf(w, "  ✓ %s\n", b.Name)
			printBlockContent(w, b.Content, "    ")
		}
	}
	fmt.Fprintln(w)
}

// --- Helpers ---

func blockToSummary(b blocks.Block) BlockSummary {
	return BlockSummary{
		Name:    b.Name,
		Type:    b.Type,
		Purpose: b.Purpose,
		Filled:  b.Filled,
		Content: b.Content,
	}
}

func summarizeBlock(b blocks.Block) string {
	if !b.Filled {
		return ""
	}
	switch rows := b.Content.(type) {
	case []map[string]string:
		return fmt.Sprintf("%d rows", len(rows))
	case string:
		if len(rows) > 60 {
			return rows[:60] + "..."
		}
		return rows
	default:
		return ""
	}
}

func printBlockContent(w io.Writer, content any, indent string) {
	switch c := content.(type) {
	case string:
		for _, line := range strings.Split(c, "\n") {
			fmt.Fprintf(w, "%s%s\n", indent, line)
		}
	case []map[string]string:
		if len(c) == 0 {
			return
		}
		var keys []string
		for k := range c[0] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, row := range c {
			var parts []string
			for _, k := range keys {
				parts = append(parts, fmt.Sprintf("%s=%s", k, row[k]))
			}
			fmt.Fprintf(w, "%s%s\n", indent, strings.Join(parts, " | "))
		}
	}
}
