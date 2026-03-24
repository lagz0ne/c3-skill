package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
)

// === READ ===

func TestRunRead_Markdown(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-101"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "---") {
		t.Error("should contain frontmatter delimiters")
	}
	if !strings.Contains(output, "id: c3-101") {
		t.Error("should contain entity ID")
	}
	if !strings.Contains(output, "## Goal") {
		t.Error("should contain body sections")
	}
}

func TestRunRead_JSON(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result ReadResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if result.ID != "c3-101" {
		t.Errorf("ID = %q", result.ID)
	}
	if result.Body == "" {
		t.Error("body should not be empty")
	}
}

func TestRunRead_NotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-999"}, &buf)
	if err == nil {
		t.Error("expected error for nonexistent entity")
	}
}

func TestRunRead_NoID(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: ""}, &buf)
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestRunRead_WithRelationships(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result ReadResult
	json.Unmarshal(buf.Bytes(), &result)
	if len(result.Uses) == 0 {
		t.Error("c3-101 should have uses relationships")
	}
}

// === READ --section ===

func TestRunRead_Section(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-101", Section: "Goal"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Handle authentication") {
		t.Errorf("should contain Goal content, got: %s", output)
	}
	// Should NOT contain other sections
	if strings.Contains(output, "## Dependencies") {
		t.Error("should not contain other sections")
	}
}

func TestRunRead_Section_JSON(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-101", Section: "Goal", JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result ReadSectionResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if result.Section != "Goal" {
		t.Errorf("section = %q", result.Section)
	}
	if result.Content == "" {
		t.Error("content should not be empty")
	}
}

func TestRunRead_Section_NotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunRead(ReadOptions{Store: s, ID: "c3-101", Section: "Nonexistent"}, &buf)
	if err == nil {
		t.Error("expected error for nonexistent section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

// === WRITE ===

func TestRunWrite_ValidContent(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	content := `---
id: c3-101
title: auth
goal: Updated authentication goal
---

# auth

## Goal

Updated authentication goal.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | credentials | c3-110 |
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "Updated authentication goal" {
		t.Errorf("goal = %q", entity.Goal)
	}
}

func TestRunWrite_RejectsMissingSections(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	content := `---
id: c3-101
title: auth
---

# auth

## Goal

Has goal but no Dependencies section.
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err == nil {
		t.Error("expected error for missing Dependencies section")
	}
	if !strings.Contains(err.Error(), "Dependencies") {
		t.Errorf("error should mention Dependencies: %v", err)
	}
}

func TestRunWrite_RejectsEmptySection(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	content := `---
id: c3-101
---

# auth

## Goal

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-110 |
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err == nil {
		t.Error("expected error for empty Goal section")
	}
}

func TestRunWrite_AutoPromotesGoal(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// No goal in frontmatter, but body has ## Goal
	content := `---
id: c3-101
title: auth
---

# auth

## Goal

Handle JWT-based authentication.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | credentials | c3-110 |
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "Handle JWT-based authentication." {
		t.Errorf("goal should be auto-promoted, got %q", entity.Goal)
	}
}

func TestRunWrite_NotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{Store: s, ID: "c3-999", Content: "test"}, &buf)
	if err == nil {
		t.Error("expected error for nonexistent entity")
	}
}

func TestRunWrite_NoID(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{Store: s, ID: "", Content: "test"}, &buf)
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestRunWrite_NoFrontmatter(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Plain markdown without frontmatter — treated as body replacement
	mdContent := `# auth

## Goal

New goal text.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-110 |
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: mdContent}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(body, "New goal text") {
		t.Error("body should be updated")
	}
}

// === WRITE --section ===

func TestRunWrite_Section(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Content: "Completely rewritten authentication goal.",
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(body, "Completely rewritten authentication goal.") {
		t.Error("section should be updated in body")
	}
	if !strings.Contains(buf.String(), "section") {
		t.Error("output should mention section update")
	}
}

func TestRunWrite_Section_AutoPromotesGoal(t *testing.T) {
	s := createRichDBFixture(t)
	// Clear the frontmatter goal
	entity, _ := s.GetEntity("c3-101")
	entity.Goal = ""
	s.UpdateEntity(entity)

	var buf bytes.Buffer
	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Content: "Auto-promoted goal from section write.",
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := s.GetEntity("c3-101")
	if updated.Goal != "Auto-promoted goal from section write." {
		t.Errorf("goal should be auto-promoted, got %q", updated.Goal)
	}
}

func TestRunWrite_Section_NotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Nonexistent Section",
		Content: "test",
	}, &buf)
	if err == nil {
		t.Error("expected error for nonexistent section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunWrite_Section_NoValidation(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Writing an empty Goal section should succeed (no section-level validation)
	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Content: "", // empty goal — allowed in section mode
	}, &buf)
	if err != nil {
		t.Errorf("section write should succeed without full validation: %v", err)
	}
}

// === NODE TREE VERIFICATION ===

func TestRunWrite_FullPopulatesNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	mdContent := `---
id: c3-101
title: auth
goal: Updated goal
---

# auth

## Goal

Updated goal via full write.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | credentials | c3-110 |
`
	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: mdContent}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify nodes exist in DB
	nodes, err := s.NodesForEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) == 0 {
		t.Fatal("node tree should be populated after write")
	}

	// Verify rendered content matches node tree
	rendered := content.RenderMarkdown(nodes)
	bodyFromTree, _ := content.ReadEntity(s, "c3-101")
	if rendered != bodyFromTree {
		t.Errorf("rendered nodes should match node tree\nrendered: %q\nbody: %q", rendered, bodyFromTree)
	}

	// Verify merkle is set
	entity, _ := s.GetEntity("c3-101")
	if entity.RootMerkle == "" {
		t.Error("root_merkle should be set after write")
	}
}

func TestRunWrite_SectionPopulatesNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Content: "Node-tree-verified goal.",
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify node tree is updated
	rendered, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Node-tree-verified goal.") {
		t.Errorf("node tree should contain updated section, got: %s", rendered)
	}

	// Verify node tree body matches
	bodyFromTree, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(bodyFromTree, "Node-tree-verified goal.") {
		t.Error("node tree should reflect updated content")
	}
}

func TestRunWrite_VersionIncrementsOnWrite(t *testing.T) {
	s := createRichDBFixture(t)

	entity, _ := s.GetEntity("c3-101")
	v0 := entity.Version

	var buf bytes.Buffer
	mdContent := `---
id: c3-101
title: auth
---

# auth

## Goal

Version test.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-110 |
`
	RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: mdContent}, &buf)

	entity, _ = s.GetEntity("c3-101")
	if entity.Version <= v0 {
		t.Errorf("version should increment: was %d, now %d", v0, entity.Version)
	}
}
