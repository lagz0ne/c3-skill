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

	content := strictComponentBody("auth", "Updated authentication goal for API requests.")

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "Updated authentication goal for API requests." {
		t.Errorf("goal = %q", entity.Goal)
	}
}

// A bad uses[] target must fail the whole write — syncRelationships removes
// existing edges before re-adding, so a typo silently drops relationships
// when the error is downgraded to a warning.
func TestRunWrite_FailsOnRelationshipSyncError(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	body := strictComponentBody("auth", "Updated authentication goal for API requests.")
	content := "---\nid: c3-101\ntitle: auth\nuses:\n  - ref-does-not-exist\n---\n\n" + body

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err == nil {
		t.Fatal("expected write to fail when uses[] references unknown entity")
	}
	if !strings.Contains(err.Error(), "ref-does-not-exist") {
		t.Errorf("error must name the bad target, got: %v", err)
	}
}

// A read|write round-trip must clear fields the user removed from frontmatter.
// Otherwise old DB values resurface on the next read/export.
func TestRunWrite_ClearsRemovedFrontmatterFields(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Seed c3-101 with status + boundary set
	e, _ := s.GetEntity("c3-101")
	e.Status = "provisioned"
	e.Boundary = "API edge"
	if err := s.UpdateEntity(e); err != nil {
		t.Fatal(err)
	}

	// Write back a body with no status / boundary in frontmatter
	body := strictComponentBody("auth", "Updated authentication goal for API requests.")
	content := "---\nid: c3-101\ntitle: auth\n---\n\n" + body

	if err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf); err != nil {
		t.Fatal(err)
	}

	e, _ = s.GetEntity("c3-101")
	if e.Status != "" {
		t.Errorf("status must be cleared when removed from frontmatter, got %q", e.Status)
	}
	if e.Boundary != "" {
		t.Errorf("boundary must be cleared when removed from frontmatter, got %q", e.Boundary)
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

Has goal but no Parent Fit section.
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err == nil {
		t.Error("expected error for missing Parent Fit section")
	}
	if !strings.Contains(err.Error(), "Parent Fit") {
		t.Errorf("error should mention Parent Fit: %v", err)
	}
}

func TestRunWrite_RejectsEmptySection(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	content := strings.Replace(strictComponentBody("auth", "Updated authentication goal for API requests."), "Updated authentication goal for API requests.", "", 1)

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err == nil {
		t.Error("expected error for empty Goal section")
	}
}

func TestRunWrite_AutoPromotesGoal(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// No goal in frontmatter, but body has ## Goal
	content := strictComponentBody("auth", "Handle JWT-based authentication for API requests.")

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "Handle JWT-based authentication for API requests." {
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
	mdContent := strictComponentBody("auth", "New goal text for authentication behavior.")

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

func TestRunWrite_Section_ValidatesResultingBody(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Content: "",
	}, &buf)
	if err == nil {
		t.Fatal("expected section write to enforce full component validation")
	}
}

// === NODE TREE VERIFICATION ===

func TestRunWrite_FullPopulatesNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	mdContent := strictComponentBody("auth", "Updated goal via full write for authentication behavior.")
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
		Content: "Node-tree verified authentication goal.",
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify node tree is updated
	rendered, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Node-tree verified authentication goal.") {
		t.Errorf("node tree should contain updated section, got: %s", rendered)
	}

	// Verify node tree body matches
	bodyFromTree, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(bodyFromTree, "Node-tree verified authentication goal.") {
		t.Error("node tree should reflect updated content")
	}
}

func TestRunWrite_VersionIncrementsOnWrite(t *testing.T) {
	s := createRichDBFixture(t)

	entity, _ := s.GetEntity("c3-101")
	v0 := entity.Version

	var buf bytes.Buffer
	mdContent := strictComponentBody("auth", "Version test for authentication behavior.")
	RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: mdContent}, &buf)

	entity, _ = s.GetEntity("c3-101")
	if entity.Version <= v0 {
		t.Errorf("version should increment: was %d, now %d", v0, entity.Version)
	}
}

// === TRUNCATION ===

func TestRunRead_AgentTruncatesBody(t *testing.T) {
	s := createRichDBFixture(t)

	// Write a body longer than 1500 chars
	longBody := "# auth\n\n## Goal\n\nHandle authentication.\n\n## Dependencies\n\n" + strings.Repeat("Lorem ipsum dolor sit amet. ", 100)
	content.WriteEntity(s, "c3-101", longBody)

	t.Setenv("C3X_MODE", "agent")
	var buf bytes.Buffer
	err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("agent read should not emit JSON:\n%s", out)
	}
	if !strings.Contains(out, "body_truncated: true") {
		t.Error("body should be truncated in agent mode")
	}
	if !strings.Contains(out, "body_total_chars:") {
		t.Error("body_total_chars should be set")
	}
	if strings.Contains(out, strings.Repeat("Lorem ipsum dolor sit amet. ", 80)) {
		t.Error("agent read body should be truncated")
	}
}

func TestRunRead_AgentFullBypassesTruncation(t *testing.T) {
	s := createRichDBFixture(t)

	longBody := "# auth\n\n## Goal\n\nHandle authentication.\n\n## Dependencies\n\n" + strings.Repeat("Lorem ipsum dolor sit amet. ", 100)
	content.WriteEntity(s, "c3-101", longBody)

	t.Setenv("C3X_MODE", "agent")
	var buf bytes.Buffer
	err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true, Full: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("agent read --full should not emit JSON:\n%s", out)
	}
	if strings.Contains(out, "body_truncated: true") {
		t.Error("body should not be truncated with --full")
	}
	if !strings.Contains(out, strings.Repeat("Lorem ipsum dolor sit amet. ", 80)) {
		t.Error("full body should include long body content")
	}
}

func TestRunRead_NonAgentNoTruncation(t *testing.T) {
	s := createRichDBFixture(t)

	longBody := "# auth\n\n## Goal\n\nHandle authentication.\n\n## Dependencies\n\n" + strings.Repeat("Lorem ipsum dolor sit amet. ", 100)
	content.WriteEntity(s, "c3-101", longBody)

	t.Setenv("C3X_MODE", "")
	var buf bytes.Buffer
	err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result ReadResult
	json.Unmarshal(buf.Bytes(), &result)
	if result.BodyTruncated {
		t.Error("body should not be truncated in non-agent mode")
	}
}

func TestRunRead_ShortBodyNotTruncated(t *testing.T) {
	s := createDBFixture(t)
	content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nShort auth body.\n")

	t.Setenv("C3X_MODE", "agent")
	var buf bytes.Buffer
	err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result ReadResult
	json.Unmarshal(buf.Bytes(), &result)
	if result.BodyTruncated {
		t.Error("short body should not be truncated")
	}
}
