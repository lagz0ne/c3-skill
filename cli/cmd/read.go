package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ReadOptions holds parameters for the read command.
type ReadOptions struct {
	Store   *store.Store
	ID      string
	JSON    bool
	Section string
	Full    bool
	Cite    bool
}

// ReadResult is the JSON output for read.
type ReadResult struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Goal           string     `json:"goal,omitempty"`
	Status         string     `json:"status,omitempty"`
	Category       string     `json:"category,omitempty"`
	ParentID       string     `json:"parent,omitempty"`
	Boundary       string     `json:"boundary,omitempty"`
	Date           string     `json:"date,omitempty"`
	Uses           []string   `json:"uses,omitempty"`
	Affects        []string   `json:"affects,omitempty"`
	Scope          []string   `json:"scope,omitempty"`
	Body           string     `json:"body"`
	BodyTruncated  bool       `json:"body_truncated,omitempty"`
	BodyTotalChars int        `json:"body_total_chars,omitempty"`
	Citation       string     `json:"citation,omitempty"`
	Help           []HelpHint `json:"help,omitempty"`
}

// ReadSectionResult is the JSON output for read --section.
type ReadSectionResult struct {
	Section   string   `json:"section"`
	Content   string   `json:"content"`
	Citations []string `json:"citations,omitempty"`
}

// RunRead outputs the full content of a single entity.
func RunRead(opts ReadOptions, w io.Writer) error {
	if opts.ID == "" {
		return fmt.Errorf("error: usage: c3x read <entity-id>\nhint: c3x list to see all entities")
	}

	entity, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("error: entity %q not found\nhint: run c3x search %q or c3x list --flat to find the current id", opts.ID, opts.ID)
	}

	// Read body from node tree.
	body, err := content.ReadEntity(opts.Store, entity.ID)
	if err != nil {
		body = ""
	}

	if opts.Section != "" {
		sections := markdown.ParseSections(body)
		for _, s := range sections {
			if s.Name == opts.Section {
				citations := []string(nil)
				if opts.Cite {
					var citeErr error
					citations, citeErr = sectionCitations(opts.Store, entity, opts.Section)
					if citeErr != nil {
						return citeErr
					}
				}
				if opts.JSON {
					return writeJSON(w, struct {
						ReadSectionResult
						Help []HelpHint `json:"help,omitempty"`
					}{
						ReadSectionResult: ReadSectionResult{Section: s.Name, Content: strings.TrimSpace(s.Content), Citations: citations},
						Help:              agentHints(cascadeHintsForEntity(entity)),
					})
				}
				fmt.Fprintln(w, strings.TrimSpace(s.Content))
				if opts.Cite {
					writeCitations(w, citations)
				}
				writeAgentHints(w, cascadeHintsForEntity(entity))
				return nil
			}
		}
		return fmt.Errorf("error: section %q not found in %s\nhint: available sections: %s",
			opts.Section, opts.ID, readAvailableSections(body))
	}

	if opts.JSON {
		result := ReadResult{
			ID:       entity.ID,
			Type:     entity.Type,
			Title:    entity.Title,
			Goal:     entity.Goal,
			Status:   entity.Status,
			Category: entity.Category,
			ParentID: entity.ParentID,
			Boundary: entity.Boundary,
			Date:     entity.Date,
			Body:     body,
			Help:     agentHints(cascadeHintsForEntity(entity)),
		}
		if opts.Cite {
			citation, citeErr := entityCitation(entity)
			if citeErr != nil {
				return citeErr
			}
			result.Citation = citation
		}

		rels, _ := opts.Store.RelationshipsFrom(entity.ID)
		for _, r := range rels {
			switch r.RelType {
			case "uses":
				result.Uses = append(result.Uses, r.ToID)
			case "affects":
				result.Affects = append(result.Affects, r.ToID)
			case "scope":
				result.Scope = append(result.Scope, r.ToID)
			}
		}

		// Truncate body in agent mode unless --full
		if isAgentMode() && !opts.Full && len(result.Body) > defaultTruncateLen {
			result.BodyTotalChars = len(result.Body)
			result.Body = result.Body[:defaultTruncateLen]
			result.BodyTruncated = true
		}

		return writeJSON(w, result)
	}

	// Default: output as markdown (same format as export)
	fmt.Fprint(w, buildExportContent(opts.Store, entity))
	if opts.Cite {
		citation, citeErr := entityCitation(entity)
		if citeErr != nil {
			return citeErr
		}
		fmt.Fprintf(w, "\ncitation: %s\n", citation)
	}
	writeAgentHints(w, cascadeHintsForEntity(entity))
	return nil
}

func entityCitation(entity *store.Entity) (string, error) {
	if entity.Version <= 0 || entity.RootMerkle == "" {
		return "", fmt.Errorf("error: %s has no versioned content hash for citation\nhint: run c3x repair, then rerun c3x read %s --cite", entity.ID, entity.ID)
	}
	return fmt.Sprintf("%s@v%d:sha256:%s", entity.ID, entity.Version, entity.RootMerkle), nil
}

func sectionCitations(s *store.Store, entity *store.Entity, sectionName string) ([]string, error) {
	nodes, err := s.NodesForEntity(entity.ID)
	if err != nil {
		return nil, fmt.Errorf("error: read nodes for %s citations: %w", entity.ID, err)
	}
	if entity.Version <= 0 {
		return nil, fmt.Errorf("error: %s has no versioned content for citation\nhint: run c3x repair, then rerun c3x read %s --section %q --cite", entity.ID, entity.ID, sectionName)
	}
	heading := sectionHeading(nodes, sectionName)
	if heading == nil {
		return nil, fmt.Errorf("error: section %q not found in node tree for %s\nhint: run c3x read %s --full to inspect available sections", sectionName, entity.ID, entity.ID)
	}
	citations := make([]string, 0)
	collectSectionChildCitations(nodes, entity, heading.ID, heading.Level, &citations)
	collectSiblingSectionSpanCitations(nodes, entity, heading, &citations)
	if len(citations) == 0 {
		return nil, fmt.Errorf("error: section %q in %s has no citable body nodes\nhint: add body content to that section, then rerun c3x repair and c3x read %s --section %q --cite", sectionName, entity.ID, entity.ID, sectionName)
	}
	return citations, nil
}

func sectionHeading(nodes []*store.Node, sectionName string) *store.Node {
	for _, n := range nodes {
		if n.Type == "heading" && n.Level == 2 && n.Content == sectionName {
			return n
		}
	}
	return nil
}

func collectSiblingSectionSpanCitations(nodes []*store.Node, entity *store.Entity, heading *store.Node, citations *[]string) {
	roots := rootNodes(nodes)
	inSection := false
	for _, n := range roots {
		if n.ID == heading.ID {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		if n.Type == "heading" && n.Level > 0 && n.Level <= heading.Level {
			return
		}
		collectNodeAndDescendantCitations(nodes, entity, n, heading.Level, citations)
	}
}

func rootNodes(nodes []*store.Node) []*store.Node {
	roots := make([]*store.Node, 0)
	for _, n := range nodes {
		if !n.ParentID.Valid {
			roots = append(roots, n)
		}
	}
	return roots
}

func collectNodeAndDescendantCitations(nodes []*store.Node, entity *store.Entity, n *store.Node, sectionLevel int, citations *[]string) {
	if isCitableSectionNode(n, sectionLevel) {
		if snippet := citationSnippetForNode(n); snippet != "" {
			*citations = append(*citations, nodeCitation(entity, n, snippet))
		}
	}
	collectSectionChildCitations(nodes, entity, n.ID, sectionLevel, citations)
}

func collectSectionChildCitations(nodes []*store.Node, entity *store.Entity, parentID int64, sectionLevel int, citations *[]string) {
	for _, n := range nodes {
		if !n.ParentID.Valid || n.ParentID.Int64 != parentID {
			continue
		}
		if isCitableSectionNode(n, sectionLevel) {
			if snippet := citationSnippetForNode(n); snippet != "" {
				*citations = append(*citations, nodeCitation(entity, n, snippet))
			}
		}
		collectSectionChildCitations(nodes, entity, n.ID, sectionLevel, citations)
	}
}

func isCitableNode(n *store.Node) bool {
	switch n.Type {
	case "paragraph", "list_item", "checklist_item", "table_header", "table_row", "code_block", "blockquote", "html_block":
		return strings.TrimSpace(n.Content) != ""
	default:
		return false
	}
}

func isCitableSectionNode(n *store.Node, sectionLevel int) bool {
	if isCitableNode(n) {
		return true
	}
	return n.Type == "heading" && n.Level > sectionLevel && strings.TrimSpace(n.Content) != ""
}

func nodeCitation(entity *store.Entity, n *store.Node, snippet string) string {
	return fmt.Sprintf("%s#n%d@v%d:sha256:%s %q", entity.ID, n.ID, entity.Version, n.Hash, snippet)
}

func citationSnippetForNode(n *store.Node) string {
	content := n.Content
	if n.Type == "code_block" {
		if _, code, ok := strings.Cut(content, "\n"); ok {
			content = code
		}
	}
	return citationSnippet(content)
}

func citationSnippet(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 160 {
			line = line[:160]
		}
		return line
	}
	return ""
}

func writeCitations(w io.Writer, citations []string) {
	if len(citations) == 1 {
		fmt.Fprintf(w, "\ncitation: %s\n", citations[0])
		return
	}
	fmt.Fprintln(w, "\ncitations:")
	for _, c := range citations {
		fmt.Fprintf(w, "  %s\n", c)
	}
}

func readAvailableSections(body string) string {
	sections := markdown.ParseSections(body)
	var names []string
	for _, s := range sections {
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}
	if len(names) == 0 {
		return "(none)"
	}
	return strings.Join(names, ", ")
}
