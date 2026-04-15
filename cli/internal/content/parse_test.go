package content

import (
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestParse_Headings(t *testing.T) {
	tree := ParseMarkdown("comp-1", "## Goal\n\n## Dependencies\n")
	if len(tree.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(tree.Nodes))
	}
	for i, n := range tree.Nodes {
		if n.Type != "heading" {
			t.Errorf("node %d: expected type heading, got %s", i, n.Type)
		}
		if n.Level != 2 {
			t.Errorf("node %d: expected level 2, got %d", i, n.Level)
		}
		if tree.ParentIndex[i] != -1 {
			t.Errorf("node %d: expected root (parent -1), got %d", i, tree.ParentIndex[i])
		}
		if n.Seq != i {
			t.Errorf("node %d: expected seq %d, got %d", i, i, n.Seq)
		}
	}
	if tree.Nodes[0].Content != "Goal" {
		t.Errorf("node 0 content: expected 'Goal', got %q", tree.Nodes[0].Content)
	}
	if tree.Nodes[1].Content != "Dependencies" {
		t.Errorf("node 1 content: expected 'Dependencies', got %q", tree.Nodes[1].Content)
	}
}

func TestParse_HeadingWithParagraph(t *testing.T) {
	tree := ParseMarkdown("comp-1", "## Goal\n\nAuthenticate requests\n")
	if len(tree.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(tree.Nodes))
	}
	h := tree.Nodes[0]
	if h.Type != "heading" || h.Content != "Goal" {
		t.Errorf("heading: type=%s content=%q", h.Type, h.Content)
	}
	p := tree.Nodes[1]
	if p.Type != "paragraph" || p.Content != "Authenticate requests" {
		t.Errorf("paragraph: type=%s content=%q", p.Type, p.Content)
	}
	// paragraph is child of heading
	if tree.ParentIndex[1] != 0 {
		t.Errorf("paragraph parent: expected 0, got %d", tree.ParentIndex[1])
	}
	if p.Seq != 0 {
		t.Errorf("paragraph seq: expected 0, got %d", p.Seq)
	}
}

func TestParse_List(t *testing.T) {
	tree := ParseMarkdown("comp-1", "- item 1\n- item 2\n")
	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes (list + 2 items), got %d", len(tree.Nodes))
	}
	if tree.Nodes[0].Type != "list" {
		t.Errorf("expected list, got %s", tree.Nodes[0].Type)
	}
	for i := 1; i <= 2; i++ {
		n := tree.Nodes[i]
		if n.Type != "list_item" {
			t.Errorf("node %d: expected list_item, got %s", i, n.Type)
		}
		if tree.ParentIndex[i] != 0 {
			t.Errorf("node %d: expected parent 0, got %d", i, tree.ParentIndex[i])
		}
		if n.Seq != i-1 {
			t.Errorf("node %d: expected seq %d, got %d", i, i-1, n.Seq)
		}
	}
	if tree.Nodes[1].Content != "item 1" {
		t.Errorf("item 1 content: %q", tree.Nodes[1].Content)
	}
	if tree.Nodes[2].Content != "item 2" {
		t.Errorf("item 2 content: %q", tree.Nodes[2].Content)
	}
}

func TestParse_OrderedList(t *testing.T) {
	tree := ParseMarkdown("comp-1", "1. first\n2. second\n")
	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(tree.Nodes))
	}
	if tree.Nodes[0].Type != "ordered_list" {
		t.Errorf("expected ordered_list, got %s", tree.Nodes[0].Type)
	}
	for i := 1; i <= 2; i++ {
		if tree.Nodes[i].Type != "list_item" {
			t.Errorf("node %d: expected list_item, got %s", i, tree.Nodes[i].Type)
		}
	}
	if tree.Nodes[1].Content != "first" {
		t.Errorf("item 1: %q", tree.Nodes[1].Content)
	}
}

func TestParse_Checklist(t *testing.T) {
	tree := ParseMarkdown("comp-1", "- [x] done\n- [ ] todo\n")
	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(tree.Nodes))
	}
	if tree.Nodes[0].Type != "checklist" {
		t.Errorf("expected checklist, got %s", tree.Nodes[0].Type)
	}
	for i := 1; i <= 2; i++ {
		if tree.Nodes[i].Type != "checklist_item" {
			t.Errorf("node %d: expected checklist_item, got %s", i, tree.Nodes[i].Type)
		}
		if tree.ParentIndex[i] != 0 {
			t.Errorf("node %d: parent expected 0, got %d", i, tree.ParentIndex[i])
		}
	}
	if tree.Nodes[1].Content != "[x] done" {
		t.Errorf("item 1 content: %q", tree.Nodes[1].Content)
	}
	if tree.Nodes[2].Content != "[ ] todo" {
		t.Errorf("item 2 content: %q", tree.Nodes[2].Content)
	}
}

func TestParse_Table(t *testing.T) {
	md := "| Name | Type |\n| --- | --- |\n| auth | service |\n| db | store |\n"
	tree := ParseMarkdown("comp-1", md)
	if len(tree.Nodes) < 4 {
		t.Fatalf("expected at least 4 nodes (table + header + 2 rows), got %d", len(tree.Nodes))
	}
	if tree.Nodes[0].Type != "table" {
		t.Errorf("expected table, got %s", tree.Nodes[0].Type)
	}
	if tree.Nodes[1].Type != "table_header" {
		t.Errorf("expected table_header, got %s", tree.Nodes[1].Type)
	}
	if tree.ParentIndex[1] != 0 {
		t.Errorf("header parent: expected 0, got %d", tree.ParentIndex[1])
	}
	for i := 2; i < len(tree.Nodes); i++ {
		if tree.Nodes[i].Type != "table_row" {
			t.Errorf("node %d: expected table_row, got %s", i, tree.Nodes[i].Type)
		}
		if tree.ParentIndex[i] != 0 {
			t.Errorf("node %d: parent expected 0, got %d", i, tree.ParentIndex[i])
		}
	}
	// Header content should contain the header cells
	if tree.Nodes[1].Content != "Name | Type" {
		t.Errorf("header content: %q", tree.Nodes[1].Content)
	}
	if tree.Nodes[2].Content != "auth | service" {
		t.Errorf("row 1 content: %q", tree.Nodes[2].Content)
	}
}

func TestParse_CodeBlock(t *testing.T) {
	md := "```go\nfunc main() {}\n```\n"
	tree := ParseMarkdown("comp-1", md)
	if len(tree.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree.Nodes))
	}
	n := tree.Nodes[0]
	if n.Type != "code_block" {
		t.Errorf("expected code_block, got %s", n.Type)
	}
	if n.Content != "go\nfunc main() {}" {
		t.Errorf("code content: %q", n.Content)
	}
	if tree.ParentIndex[0] != -1 {
		t.Errorf("expected root, got parent %d", tree.ParentIndex[0])
	}
}

func TestParse_Blockquote(t *testing.T) {
	tree := ParseMarkdown("comp-1", "> quoted text\n")
	if len(tree.Nodes) != 2 {
		t.Fatalf("expected 2 nodes (blockquote + child paragraph), got %d", len(tree.Nodes))
	}
	if tree.Nodes[0].Type != "blockquote" {
		t.Errorf("expected blockquote, got %s", tree.Nodes[0].Type)
	}
	if tree.Nodes[1].Type != "paragraph" {
		t.Errorf("expected paragraph child, got %s", tree.Nodes[1].Type)
	}
	if tree.ParentIndex[1] != 0 {
		t.Errorf("child parent: expected 0, got %d", tree.ParentIndex[1])
	}
	if tree.Nodes[1].Content != "quoted text" {
		t.Errorf("content: %q", tree.Nodes[1].Content)
	}
}

func TestParse_NestedList(t *testing.T) {
	md := "- parent\n  - child 1\n  - child 2\n"
	tree := ParseMarkdown("comp-1", md)
	// Current parser shape: list(0), list_item "parent"(1), nested child items under parent(2,3).
	if len(tree.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(tree.Nodes))
	}
	// Top-level list
	if tree.Nodes[0].Type != "list" {
		t.Errorf("node 0: expected list, got %s", tree.Nodes[0].Type)
	}
	// Parent item is child of top list
	if tree.Nodes[1].Type != "list_item" {
		t.Errorf("node 1: expected list_item, got %s", tree.Nodes[1].Type)
	}
	if tree.ParentIndex[1] != 0 {
		t.Errorf("node 1 parent: expected 0, got %d", tree.ParentIndex[1])
	}
	// Nested items are direct children of parent list_item.
	if tree.Nodes[2].Type != "list_item" || tree.ParentIndex[2] != 1 {
		t.Errorf("node 2: type=%s parent=%d", tree.Nodes[2].Type, tree.ParentIndex[2])
	}
	if tree.Nodes[2].Content != "child 1" {
		t.Errorf("node 2 content: %q", tree.Nodes[2].Content)
	}
	if tree.Nodes[3].Type != "list_item" || tree.ParentIndex[3] != 1 {
		t.Errorf("node 3: type=%s parent=%d", tree.Nodes[3].Type, tree.ParentIndex[3])
	}
	if tree.Nodes[3].Content != "child 2" {
		t.Errorf("node 3 content: %q", tree.Nodes[3].Content)
	}
}

func TestParse_ListItemDoesNotDuplicateParagraphChild(t *testing.T) {
	md := "- alpha\n- beta\n"
	tree := ParseMarkdown("comp-1", md)

	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes (list + 2 items), got %d", len(tree.Nodes))
	}
	for i, n := range tree.Nodes {
		if n.Type == "paragraph" {
			t.Fatalf("unexpected paragraph node at %d: %+v", i, n)
		}
	}
}

func TestParse_FullDocument(t *testing.T) {
	md := `## Goal

Authenticate incoming API requests using JWT tokens.

## Dependencies

| Name | Type | Contract |
| --- | --- | --- |
| auth-service | service | OAuth2 |
| user-store | store | SQL |
`
	tree := ParseMarkdown("comp-auth", md)

	// Expect:
	// 0: heading "Goal" (root, seq 0)
	// 1: paragraph (child of 0, seq 0)
	// 2: heading "Dependencies" (root, seq 1)
	// 3: table (child of 2, seq 0)
	// 4: table_header (child of 3, seq 0)
	// 5: table_row (child of 3, seq 1)
	// 6: table_row (child of 3, seq 2)
	if len(tree.Nodes) != 7 {
		t.Fatalf("expected 7 nodes, got %d", len(tree.Nodes))
	}

	checks := []struct {
		idx        int
		typ        string
		parentIdx  int
		seq        int
		contentHas string
	}{
		{0, "heading", -1, 0, "Goal"},
		{1, "paragraph", 0, 0, "Authenticate"},
		{2, "heading", -1, 1, "Dependencies"},
		{3, "table", 2, 0, ""},
		{4, "table_header", 3, 0, "Name"},
		{5, "table_row", 3, 1, "auth-service"},
		{6, "table_row", 3, 2, "user-store"},
	}
	for _, c := range checks {
		n := tree.Nodes[c.idx]
		if n.Type != c.typ {
			t.Errorf("node %d: expected type %s, got %s", c.idx, c.typ, n.Type)
		}
		if tree.ParentIndex[c.idx] != c.parentIdx {
			t.Errorf("node %d: expected parent %d, got %d", c.idx, c.parentIdx, tree.ParentIndex[c.idx])
		}
		if n.Seq != c.seq {
			t.Errorf("node %d: expected seq %d, got %d", c.idx, c.seq, n.Seq)
		}
		if c.contentHas != "" && !containsStr(n.Content, c.contentHas) {
			t.Errorf("node %d: content %q does not contain %q", c.idx, n.Content, c.contentHas)
		}
	}
}

func TestParse_HashesComputed(t *testing.T) {
	tree := ParseMarkdown("e1", "## Title\n\nBody text\n")
	for i, n := range tree.Nodes {
		if n.Hash == "" {
			t.Errorf("node %d has empty hash", i)
		}
		expected := store.ComputeNodeHash(n.Content, n.Type)
		if n.Hash != expected {
			t.Errorf("node %d: hash mismatch: got %s, want %s", i, n.Hash, expected)
		}
	}
}

func TestParse_EntityIDSet(t *testing.T) {
	tree := ParseMarkdown("my-entity", "## Heading\n\nParagraph\n")
	for i, n := range tree.Nodes {
		if n.EntityID != "my-entity" {
			t.Errorf("node %d: EntityID = %q, want 'my-entity'", i, n.EntityID)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
