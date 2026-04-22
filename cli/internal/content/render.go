package content

import (
	"fmt"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// RenderMarkdown converts a flat node list back into a markdown string.
func RenderMarkdown(nodes []*store.Node) string {
	if len(nodes) == 0 {
		return ""
	}

	children := map[int64][]*store.Node{} // parentID → children
	var roots []*store.Node
	for _, n := range nodes {
		if n.ParentID.Valid {
			children[n.ParentID.Int64] = append(children[n.ParentID.Int64], n)
		} else {
			roots = append(roots, n)
		}
	}

	var b strings.Builder
	for _, n := range roots {
		renderNode(&b, n, children, 0)
	}
	// Trim at most one trailing newline to avoid double-blank at end.
	s := b.String()
	s = strings.TrimRight(s, "\n")
	if s != "" {
		s += "\n"
	}
	return s
}

func renderNode(b *strings.Builder, n *store.Node, children map[int64][]*store.Node, depth int) {
	switch n.Type {
	case "heading":
		b.WriteString(strings.Repeat("#", n.Level))
		b.WriteString(" ")
		b.WriteString(n.Content)
		b.WriteString("\n\n")
		for _, c := range children[n.ID] {
			renderNode(b, c, children, depth)
		}

	case "paragraph":
		b.WriteString(n.Content)
		b.WriteString("\n\n")

	case "list":
		for _, c := range children[n.ID] {
			if c.Type == "list_item" {
				indent := strings.Repeat("  ", depth)
				b.WriteString(indent)
				b.WriteString("- ")
				b.WriteString(c.Content)
				b.WriteString("\n")
				for _, sub := range children[c.ID] {
					renderNode(b, sub, children, depth+1)
				}
			}
		}
		b.WriteString("\n")

	case "ordered_list":
		for i, c := range children[n.ID] {
			if c.Type == "list_item" {
				indent := strings.Repeat("  ", depth)
				fmt.Fprintf(b, "%s%d. %s\n", indent, i+1, c.Content)
				for _, sub := range children[c.ID] {
					renderNode(b, sub, children, depth+1)
				}
			}
		}
		b.WriteString("\n")

	case "checklist":
		for _, c := range children[n.ID] {
			if c.Type == "checklist_item" {
				b.WriteString("- ")
				b.WriteString(c.Content)
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")

	case "table":
		kids := children[n.ID]
		for i, c := range kids {
			cols := strings.Split(c.Content, " | ")
			b.WriteString("| ")
			b.WriteString(strings.Join(cols, " | "))
			b.WriteString(" |\n")
			if i == 0 && c.Type == "table_header" {
				b.WriteString("| ")
				seps := make([]string, len(cols))
				for j := range seps {
					seps[j] = "---"
				}
				b.WriteString(strings.Join(seps, " | "))
				b.WriteString(" |\n")
			}
		}
		b.WriteString("\n")

	case "code_block":
		lang, code := parseCodeContent(n.Content)
		b.WriteString("```")
		b.WriteString(lang)
		b.WriteString("\n")
		b.WriteString(code)
		b.WriteString("\n```\n\n")

	case "blockquote":
		var inner strings.Builder
		for _, c := range children[n.ID] {
			renderNode(&inner, c, children, depth)
		}
		for _, line := range strings.Split(strings.TrimRight(inner.String(), "\n"), "\n") {
			b.WriteString("> ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")

	default:
		if n.Content != "" {
			b.WriteString(n.Content)
			b.WriteString("\n\n")
		}
	}
}

// parseCodeContent splits `lang\ncode` (lang may be empty). Legacy rows
// without a newline are treated as no-lang code.
func parseCodeContent(content string) (lang, code string) {
	idx := strings.IndexByte(content, '\n')
	if idx == -1 {
		return "", content
	}
	return content[:idx], content[idx+1:]
}
