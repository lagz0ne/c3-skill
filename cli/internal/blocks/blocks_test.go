package blocks

import (
	"testing"
)

func TestExtractBlocks_TextSection(t *testing.T) {
	body := "## Goal\n\nHandle authentication."
	blocks := ExtractBlocks(body, "component")

	goal := findBlock(blocks, "Goal")
	if goal == nil {
		t.Fatal("expected Goal block")
	}
	if goal.Type != "text" {
		t.Errorf("Goal.Type = %q, want %q", goal.Type, "text")
	}
	if !goal.Filled {
		t.Error("Goal should be filled")
	}
	text, ok := goal.Content.(string)
	if !ok {
		t.Fatalf("Goal.Content type = %T, want string", goal.Content)
	}
	if text != "Handle authentication." {
		t.Errorf("Goal.Content = %q, want %q", text, "Handle authentication.")
	}
}

func TestExtractBlocks_TableSection(t *testing.T) {
	body := `## Dependencies

| Direction | What | From/To |
|---|---|---|
| IN | creds | c3-110 |`

	blocks := ExtractBlocks(body, "component")

	deps := findBlock(blocks, "Dependencies")
	if deps == nil {
		t.Fatal("expected Dependencies block")
	}
	if deps.Type != "table" {
		t.Errorf("Dependencies.Type = %q, want %q", deps.Type, "table")
	}
	if !deps.Filled {
		t.Error("Dependencies should be filled")
	}
	rows, ok := deps.Content.([]map[string]string)
	if !ok {
		t.Fatalf("Dependencies.Content type = %T, want []map[string]string", deps.Content)
	}
	if len(rows) != 1 {
		t.Fatalf("row count = %d, want 1", len(rows))
	}
	if rows[0]["Direction"] != "IN" {
		t.Errorf("row[0][Direction] = %q, want %q", rows[0]["Direction"], "IN")
	}
	if rows[0]["What"] != "creds" {
		t.Errorf("row[0][What] = %q, want %q", rows[0]["What"], "creds")
	}
}

func TestExtractBlocks_EmptyTableSection(t *testing.T) {
	body := `## Dependencies

| Direction | What | From/To |
|---|---|---|
`

	blocks := ExtractBlocks(body, "component")

	deps := findBlock(blocks, "Dependencies")
	if deps == nil {
		t.Fatal("expected Dependencies block")
	}
	if deps.Filled {
		t.Error("empty table should not be filled")
	}
	rows, ok := deps.Content.([]map[string]string)
	if !ok {
		t.Fatalf("Content type = %T, want []map[string]string", deps.Content)
	}
	if len(rows) != 0 {
		t.Errorf("row count = %d, want 0", len(rows))
	}
}

func TestExtractBlocks_MissingSection(t *testing.T) {
	body := "## Goal\n\nSome goal."

	blocks := ExtractBlocks(body, "component")

	deps := findBlock(blocks, "Dependencies")
	if deps == nil {
		t.Fatal("expected Dependencies block (from schema, even if missing)")
	}
	if deps.Filled {
		t.Error("missing section should not be filled")
	}
	if deps.Content != nil {
		t.Errorf("missing section Content should be nil, got %v", deps.Content)
	}
}

func TestExtractBlocks_CodeSection(t *testing.T) {
	body := "## How\n\n```go\nfunc Example() {}\n```"

	blocks := ExtractBlocks(body, "ref")

	how := findBlock(blocks, "How")
	if how == nil {
		t.Fatal("expected How block")
	}
	if how.Type != "text" {
		t.Errorf("How.Type = %q, want %q", how.Type, "text")
	}
	if !how.Filled {
		t.Error("How should be filled")
	}
	// Code sections are text type in schema — content includes the code block
	text, ok := how.Content.(string)
	if !ok {
		t.Fatalf("How.Content type = %T, want string", how.Content)
	}
	if text == "" {
		t.Error("How.Content should not be empty")
	}
}

func TestExtractBlocks_PurposePopulated(t *testing.T) {
	body := "## Goal\n\nSome goal."

	blocks := ExtractBlocks(body, "component")

	for _, b := range blocks {
		if b.Purpose == "" {
			t.Errorf("block %q has no Purpose", b.Name)
		}
	}
}

func TestExtractBlock_SingleSection(t *testing.T) {
	body := "## Goal\n\nThe goal.\n\n## Dependencies\n\n| Direction | What | From/To |\n|---|---|---|\n| IN | x | c3-1 |"

	b := ExtractBlock(body, "component", "Goal")
	if b == nil {
		t.Fatal("expected Goal block")
	}
	if b.Name != "Goal" {
		t.Errorf("Name = %q, want %q", b.Name, "Goal")
	}

	text, ok := b.Content.(string)
	if !ok {
		t.Fatalf("Content type = %T, want string", b.Content)
	}
	if text != "The goal." {
		t.Errorf("Content = %q, want %q", text, "The goal.")
	}
}

func TestExtractBlock_Unknown(t *testing.T) {
	body := "## Goal\n\nThe goal."

	b := ExtractBlock(body, "component", "NonExistent")
	if b != nil {
		t.Errorf("expected nil for unknown section, got %+v", b)
	}
}

func TestExtractBlocks_AllSchemaTypes(t *testing.T) {
	// Verify we get blocks for all entity types
	for _, typ := range []string{"component", "container", "context", "ref", "adr"} {
		blocks := ExtractBlocks("## Goal\n\nTest.", typ)
		if len(blocks) == 0 {
			t.Errorf("expected blocks for type %q, got none", typ)
		}
	}
}

func TestExtractBlocks_UnknownType(t *testing.T) {
	blocks := ExtractBlocks("## Goal\n\nTest.", "bogus")
	if len(blocks) != 0 {
		t.Errorf("expected no blocks for unknown type, got %d", len(blocks))
	}
}

func TestExtractBlock_SlugMatching(t *testing.T) {
	body := "## Related Refs\n\n| Ref | Role |\n|---|---|\n| ref-jwt | Auth |"

	// Exact name match
	b := ExtractBlock(body, "component", "Related Refs")
	if b == nil {
		t.Fatal("expected Related Refs block with exact name")
	}

	// Slug match (lowercase + hyphenated)
	b = ExtractBlock(body, "component", "related-refs")
	if b == nil {
		t.Fatal("expected Related Refs block with slug match")
	}
}

// helpers

func findBlock(blocks []Block, name string) *Block {
	for i := range blocks {
		if blocks[i].Name == name {
			return &blocks[i]
		}
	}
	return nil
}
