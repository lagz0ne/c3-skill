package content

import (
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// NodeTree is a parsed content tree ready for DB insertion.
// Nodes are in parent-first order. ParentIndex[i] = index of parent node, or -1 for root.
type NodeTree struct {
	Nodes       []*store.Node
	ParentIndex []int
}

// ParseMarkdown parses markdown into a NodeTree for the given entity.
func ParseMarkdown(entityID, markdown string) *NodeTree {
	source := []byte(markdown)
	md := goldmark.New(goldmark.WithExtensions(extension.Table, extension.TaskList))
	doc := md.Parser().Parse(text.NewReader(source))

	tree := &NodeTree{}
	seqByParent := map[int]int{} // parentIndex -> next seq

	var walk func(n ast.Node, parentIdx int)
	walk = func(n ast.Node, parentIdx int) {
		// At document level, headings create sections: subsequent non-heading
		// blocks become children of the most recent heading.
		isDocLevel := n.Kind() == ast.KindDocument
		curHeadingIdx := -1 // tracks current heading at this level

		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			nodeType, ok := mapNodeType(c, source)
			if !ok {
				walk(c, parentIdx)
				continue
			}

			// At document level, non-heading blocks nest under current heading
			effectiveParent := parentIdx
			if isDocLevel && nodeType != "heading" && curHeadingIdx >= 0 {
				effectiveParent = curHeadingIdx
			}

			content := extractContent(c, source, nodeType)
			seq := seqByParent[effectiveParent]
			seqByParent[effectiveParent]++

			node := &store.Node{
				EntityID: entityID,
				Type:     nodeType,
				Level:    nodeLevel(c),
				Seq:      seq,
				Content:  content,
			}
			node.Hash = store.ComputeNodeHash(node.Content, node.Type)

			idx := len(tree.Nodes)
			tree.Nodes = append(tree.Nodes, node)
			tree.ParentIndex = append(tree.ParentIndex, effectiveParent)

			// Track current heading at document level
			if isDocLevel && nodeType == "heading" {
				curHeadingIdx = idx
			}

			// Recurse into children for container types
			if isContainer(nodeType) {
				walk(c, idx)
			}
		}
	}

	walk(doc, -1)
	return tree
}

// mapNodeType returns our node type for a goldmark AST node, or false if we skip it.
func mapNodeType(n ast.Node, source []byte) (string, bool) {
	switch n.Kind() {
	case ast.KindHeading:
		return "heading", true
	case ast.KindParagraph:
		return "paragraph", true
	case ast.KindList:
		list := n.(*ast.List)
		if list.IsOrdered() {
			return "ordered_list", true
		}
		if isChecklist(n, source) {
			return "checklist", true
		}
		return "list", true
	case ast.KindListItem:
		// Check if parent is a checklist
		if p := n.Parent(); p != nil && p.Kind() == ast.KindList {
			if isChecklist(p, source) {
				return "checklist_item", true
			}
		}
		return "list_item", true
	case ast.KindFencedCodeBlock:
		return "code_block", true
	case ast.KindBlockquote:
		return "blockquote", true
	}

	// Extension types
	switch n.Kind() {
	case east.KindTable:
		return "table", true
	case east.KindTableHeader:
		return "table_header", true
	case east.KindTableRow:
		return "table_row", true
	}

	return "", false
}

// isChecklist returns true if a list contains task checkbox items.
func isChecklist(n ast.Node, source []byte) bool {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindListItem {
			for ic := c.FirstChild(); ic != nil; ic = ic.NextSibling() {
				if hasTaskCheckbox(ic) {
					return true
				}
			}
		}
	}
	return false
}

// hasTaskCheckbox checks if a node contains a TaskCheckBox inline.
func hasTaskCheckbox(n ast.Node) bool {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == east.KindTaskCheckBox {
			return true
		}
	}
	return false
}

func extractContent(n ast.Node, source []byte, nodeType string) string {
	switch nodeType {
	case "code_block":
		return extractCodeBlock(n, source)
	case "table_header", "table_row":
		return extractTableRow(n, source)
	case "checklist_item":
		return extractChecklistItem(n, source)
	case "list", "ordered_list", "checklist", "table", "blockquote":
		// Container nodes have no direct content
		return ""
	case "list_item":
		return extractListItem(n, source)
	default:
		// heading, paragraph — get inline text
		return strings.TrimSpace(string(n.Text(source)))
	}
}

// extractCodeBlock gets language + code lines.
func extractCodeBlock(n ast.Node, source []byte) string {
	var sb strings.Builder
	if fcb, ok := n.(*ast.FencedCodeBlock); ok {
		lang := fcb.Language(source)
		if len(lang) > 0 {
			sb.Write(lang)
			sb.WriteByte('\n')
		}
	}
	lines := n.Lines()
	for i := range lines.Len() {
		line := lines.At(i)
		sb.Write(line.Value(source))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// extractTableRow gets pipe-separated cell text.
func extractTableRow(n ast.Node, source []byte) string {
	var cells []string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		cells = append(cells, strings.TrimSpace(string(c.Text(source))))
	}
	return strings.Join(cells, " | ")
}

func extractListItem(n ast.Node, source []byte) string {
	// A ListItem's first child is typically a Paragraph (or TextBlock in tight lists)
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindParagraph || c.Kind() == ast.KindTextBlock {
			return strings.TrimSpace(string(c.Text(source)))
		}
	}
	return strings.TrimSpace(string(n.Text(source)))
}

// extractChecklistItem extracts [x]/[ ] prefix + text from source.
func extractChecklistItem(n ast.Node, source []byte) string {
	// n.Text(source) already includes [x]/[ ] prefix from source bytes
	return strings.TrimSpace(string(n.Text(source)))
}

func nodeLevel(n ast.Node) int {
	if h, ok := n.(*ast.Heading); ok {
		return h.Level
	}
	return 0
}

func isContainer(nodeType string) bool {
	switch nodeType {
	case "list", "ordered_list", "checklist", "table", "blockquote", "list_item":
		return true
	}
	return false
}
