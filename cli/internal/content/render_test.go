package content

import (
	"database/sql"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// --- helpers ---

func rootNode(id int64, typ string, level, seq int, content string) *store.Node {
	return &store.Node{
		ID:       id,
		EntityID: "test",
		ParentID: sql.NullInt64{Valid: false},
		Type:     typ,
		Level:    level,
		Seq:      seq,
		Content:  content,
	}
}

func childNode(id, parentID int64, typ string, level, seq int, content string) *store.Node {
	return &store.Node{
		ID:       id,
		EntityID: "test",
		ParentID: sql.NullInt64{Int64: parentID, Valid: true},
		Type:     typ,
		Level:    level,
		Seq:      seq,
		Content:  content,
	}
}

// --- tests ---

func TestRender_Empty(t *testing.T) {
	if got := RenderMarkdown(nil); got != "" {
		t.Errorf("nil nodes: got %q, want empty", got)
	}
	if got := RenderMarkdown([]*store.Node{}); got != "" {
		t.Errorf("empty nodes: got %q, want empty", got)
	}
}

func TestRender_Heading(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "heading", 2, 0, "Goal"),
	}
	want := "## Goal\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_HeadingWithParagraph(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "heading", 2, 0, "Goal"),
		childNode(2, 1, "paragraph", 0, 0, "Authenticate requests"),
	}
	want := "## Goal\n\nAuthenticate requests\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_MultipleSections(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "heading", 2, 0, "Goal"),
		childNode(2, 1, "paragraph", 0, 0, "First section"),
		rootNode(3, "heading", 2, 1, "Design"),
		childNode(4, 3, "paragraph", 0, 0, "Second section"),
	}
	want := "## Goal\n\nFirst section\n\n## Design\n\nSecond section\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_List(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "list", 0, 0, ""),
		childNode(2, 1, "list_item", 0, 0, "item 1"),
		childNode(3, 1, "list_item", 0, 1, "item 2"),
	}
	want := "- item 1\n- item 2\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_OrderedList(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "ordered_list", 0, 0, ""),
		childNode(2, 1, "list_item", 0, 0, "first"),
		childNode(3, 1, "list_item", 0, 1, "second"),
	}
	want := "1. first\n2. second\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_Checklist(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "checklist", 0, 0, ""),
		childNode(2, 1, "checklist_item", 0, 0, "[x] done"),
		childNode(3, 1, "checklist_item", 0, 1, "[ ] todo"),
	}
	want := "- [x] done\n- [ ] todo\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_Table(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "table", 0, 0, ""),
		childNode(2, 1, "table_header", 0, 0, "Name | Age"),
		childNode(3, 1, "table_row", 0, 1, "Alice | 30"),
		childNode(4, 1, "table_row", 0, 2, "Bob | 25"),
	}
	want := "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 25 |\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_CodeBlock(t *testing.T) {
	// Convention: content is "lang\ncode" when a language is present.
	nodes := []*store.Node{
		rootNode(1, "code_block", 0, 0, "go\nfunc main() {}"),
	}
	want := "```go\nfunc main() {}\n```\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_CodeBlockNoLang(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "code_block", 0, 0, "echo hello"),
	}
	want := "```\necho hello\n```\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_Blockquote(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "blockquote", 0, 0, ""),
		childNode(2, 1, "paragraph", 0, 0, "quoted text"),
	}
	want := "> quoted text\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// TestRender_SiblingHeadingsHaveBlankLine guards against the regression
// where two top-level headings emitted without a blank line between them
// (e.g. an H1 with no body followed by H2 sections).
func TestRender_SiblingHeadingsHaveBlankLine(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "heading", 1, 0, "Title"),
		rootNode(2, "heading", 2, 1, "Goal"),
		childNode(3, 2, "paragraph", 0, 0, "Body."),
	}
	want := "# Title\n\n## Goal\n\nBody.\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// TestRender_BlockToHeading guards against a list/table/code_block being
// directly abutted to a following heading.
func TestRender_BlockToHeading(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "heading", 2, 0, "A"),
		childNode(2, 1, "list", 0, 0, ""),
		childNode(3, 2, "list_item", 0, 0, "x"),
		rootNode(4, "heading", 2, 1, "B"),
	}
	want := "## A\n\n- x\n\n## B\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRender_NestedList(t *testing.T) {
	nodes := []*store.Node{
		rootNode(1, "list", 0, 0, ""),
		childNode(2, 1, "list_item", 0, 0, "parent"),
		childNode(3, 2, "list", 0, 0, ""),
		childNode(4, 3, "list_item", 0, 0, "child"),
	}
	want := "- parent\n  - child\n"
	if got := RenderMarkdown(nodes); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
