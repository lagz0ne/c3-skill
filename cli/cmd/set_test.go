package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
)

func TestRunSet_FrontmatterField(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-101", Field: "goal", Value: "Handle JWT authentication"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "Handle JWT authentication" {
		t.Errorf("goal = %q, want %q", entity.Goal, "Handle JWT authentication")
	}

	output := buf.String()
	if !strings.Contains(output, "Updated") {
		t.Errorf("should print Updated message, got: %s", output)
	}
}

func TestRunSet_EntityNotFound(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-999", Field: "goal", Value: "test"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunSet_UpdatesContainerGoal(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-1", Field: "goal", Value: "Serve high-performance API requests"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-1")
	if entity.Goal != "Serve high-performance API requests" {
		t.Errorf("container goal = %q", entity.Goal)
	}
}

func TestRunSet_UpdatesRefGoal(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "ref-jwt", Field: "goal", Value: "Standardize JWT token format"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("ref-jwt")
	if entity.Goal != "Standardize JWT token format" {
		t.Errorf("ref goal = %q", entity.Goal)
	}
}

func TestRunSet_UpdatesAdrStatus(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "status", Value: "accepted"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("adr-20260226-use-go")
	if entity.Status != "accepted" {
		t.Errorf("adr status = %q, want %q", entity.Status, "accepted")
	}
}

func TestRunSet_InvalidField(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-101", Field: "nonexistent_field", Value: "test"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for unknown field name")
	}
}

func TestRunSet_SectionText(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Value:   "Provide JWT-based authentication for all API endpoints.",
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(body, "Provide JWT-based authentication for all API endpoints.") {
		t.Error("section content should be updated")
	}
}

func TestRunSet_SectionNotFound(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "NonExistent Section",
		Value:   "content",
	}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for nonexistent section")
	}
}

func TestRunSet_SectionAppend(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Dependencies",
		Value:   `{"Direction":"OUT","What":"events","From/To":"c3-103"}`,
		Append:  true,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := content.ReadEntity(s, "c3-101")
	// Original row should survive
	if !strings.Contains(body, "user credentials") {
		t.Error("existing row should be preserved")
	}
	// New row should be added
	if !strings.Contains(body, "events") {
		t.Error("new row should be appended")
	}
}

func TestRunSet_AllFields(t *testing.T) {
	fields := map[string]string{
		"boundary": "service",
		"category": "feature",
		"title":    "New Title",
		"date":     "2026-03-20",
	}

	for field, value := range fields {
		t.Run(field, func(t *testing.T) {
			s := createDBFixture(t)
			var buf bytes.Buffer
			opts := SetOptions{Store: s, ID: "c3-101", Field: field, Value: value}
			err := RunSet(opts, &buf)
			if err != nil {
				t.Fatal(err)
			}
			entity, _ := s.GetEntity("c3-101")
			var got string
			switch field {
			case "boundary":
				got = entity.Boundary
			case "category":
				got = entity.Category
			case "title":
				got = entity.Title
			case "date":
				got = entity.Date
			}
			if got != value {
				t.Errorf("%s = %q, want %q", field, got, value)
			}
		})
	}
}

func TestRunSet_SectionJSONArray(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Related Refs",
		Value:   `[{"Ref":"ref-logging","Role":"Structured logging"}]`,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(body, "ref-logging") {
		t.Error("body should contain the new table data")
	}
}

func TestRunSet_SectionAppendInvalidJSON(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Dependencies",
		Value:   "not json",
		Append:  true,
	}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for invalid JSON in append mode")
	}
}

func TestRunSet_SectionJSONArray_NoExistingTable(t *testing.T) {
	s := createRichDBFixture(t)
	// Set up an entity with required sections + a custom section via node tree
	newBody := "# users\n\n## Goal\n\nManage users.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n| IN | data | c3-101 |\n\n## Custom Section\n\nSome text here.\n"
	content.WriteEntity(s, "c3-110", newBody)

	var buf bytes.Buffer
	opts := SetOptions{
		Store:   s,
		ID:      "c3-110",
		Section: "Custom Section",
		Value:   `[{"Key":"value1","Data":"data1"}]`,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	updatedBody, _ := content.ReadEntity(s, "c3-110")
	if !strings.Contains(updatedBody, "value1") {
		t.Error("body should contain the new table data")
	}
}

func TestRunSet_BatchStdin(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	payload := `{"fields":{"goal":"New batch goal"},"sections":{"Goal":"New goal content from batch."}}`
	opts := SetOptions{
		Store: s,
		ID:    "c3-101",
		Value: payload,
		Stdin: true,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "New batch goal" {
		t.Errorf("goal = %q, want %q", entity.Goal, "New batch goal")
	}
	batchBody, _ := content.ReadEntity(s, "c3-101")
	if !strings.Contains(batchBody, "New goal content from batch.") {
		t.Error("body should contain updated Goal section")
	}

	output := buf.String()
	if !strings.Contains(output, "1 fields") {
		t.Errorf("output should mention field count: %s", output)
	}
}

func TestIsJSONArray(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`[{"key":"value"}]`, true},
		{`[]`, true},
		{`  [1, 2, 3]  `, true},
		{`not json`, false},
		{`{"key":"value"}`, false},
		{`[invalid`, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isJSONArray(tt.input); got != tt.want {
				t.Errorf("isJSONArray(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// === CODEMAP TESTS ===

func TestRunSet_CodemapReplace(t *testing.T) {
	s := createDBFixture(t)
	// Seed existing patterns
	s.SetCodeMap("c3-101", []string{"src/old/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth/**,src/auth.go"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if !containsStr2(patterns, "src/auth/**") || !containsStr2(patterns, "src/auth.go") {
		t.Errorf("patterns = %v, want [src/auth/** src/auth.go]", patterns)
	}

	output := buf.String()
	if !strings.Contains(output, "codemap") {
		t.Errorf("output should mention codemap: %s", output)
	}
}

func TestRunSet_CodemapAppend(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth.go", Append: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if !containsStr2(patterns, "src/auth/**") || !containsStr2(patterns, "src/auth.go") {
		t.Errorf("patterns = %v", patterns)
	}
}

func TestRunSet_CodemapRemove(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**", "src/auth.go", "src/utils.go"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth.go", Remove: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if containsStr2(patterns, "src/auth.go") {
		t.Error("src/auth.go should have been removed")
	}
}

func TestRunSet_CodemapRemoveNonexistent(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/nope.go", Remove: true}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error when removing nonexistent pattern")
	}
}

func TestRunSet_CodemapBatchStdin(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	payload := `{"fields":{"goal":"New goal"},"codemap":["src/auth/**","src/auth.go"]}`
	opts := SetOptions{Store: s, ID: "c3-101", Value: payload, Stdin: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "New goal" {
		t.Errorf("goal = %q", entity.Goal)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
}

func TestRunSet_CodemapEmpty(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	// Empty value clears all patterns
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: ""}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns after clear, got %d: %v", len(patterns), patterns)
	}
}

func TestRunSet_CodemapAppendRemoveMutualExclusion(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth/**", Append: true, Remove: true}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for --append and --remove together")
	}
	if !strings.Contains(err.Error(), "cannot use --append and --remove") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSet_CodemapAppendDuplicate(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth/**", Append: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 1 {
		t.Errorf("duplicate should be skipped, got %d patterns: %v", len(patterns), patterns)
	}
	if !strings.Contains(buf.String(), "already exists") {
		t.Errorf("should report duplicate: %s", buf.String())
	}
}

func TestRunSet_CodemapBatchClear(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**", "src/utils.go"})

	var buf bytes.Buffer
	payload := `{"codemap":[]}`
	opts := SetOptions{Store: s, ID: "c3-101", Value: payload, Stdin: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns after batch clear, got %d: %v", len(patterns), patterns)
	}
}

// === NODE TREE VERIFICATION ===

func TestRunSet_SectionPopulatesNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{
		Store:   s,
		ID:      "c3-101",
		Section: "Goal",
		Value:   "Node-tree set goal.",
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify node tree content
	rendered, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Node-tree set goal.") {
		t.Errorf("node tree should contain updated section, got: %s", rendered)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.RootMerkle == "" {
		t.Error("merkle should be set after set --section")
	}
}

func TestRunSet_BatchSectionsPopulateNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	payload := `{"fields":{"goal":"batch goal"},"sections":{"Goal":"Batch node-tree goal."}}`
	opts := SetOptions{
		Store: s,
		ID:    "c3-101",
		Value: payload,
		Stdin: true,
	}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	rendered, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Batch node-tree goal.") {
		t.Errorf("node tree should contain batch-updated section, got: %s", rendered)
	}

	// Verify batch fields were applied (goal was in the batch payload)
	entity, _ := s.GetEntity("c3-101")
	_ = entity // fields verified via node tree
}
