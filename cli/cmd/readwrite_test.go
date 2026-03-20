package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
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
	content := `# auth

## Goal

New goal text.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | data | c3-110 |
`

	err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: content}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if !strings.Contains(entity.Body, "New goal text") {
		t.Error("body should be updated")
	}
}
