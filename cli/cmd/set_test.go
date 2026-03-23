package cmd

import (
	"bytes"
	"strings"
	"testing"
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

	entity, _ := s.GetEntity("c3-101")
	if !strings.Contains(entity.Body, "Provide JWT-based authentication for all API endpoints.") {
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

	entity, _ := s.GetEntity("c3-101")
	// Original row should survive
	if !strings.Contains(entity.Body, "user credentials") {
		t.Error("existing row should be preserved")
	}
	// New row should be added
	if !strings.Contains(entity.Body, "events") {
		t.Error("new row should be appended")
	}
}

func TestRunSet_AllFields(t *testing.T) {
	fields := map[string]string{
		"summary":     "New summary",
		"boundary":    "service",
		"category":    "feature",
		"title":       "New Title",
		"date":        "2026-03-20",
		"description": "Test desc",
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
			case "summary":
				got = entity.Summary
			case "boundary":
				got = entity.Boundary
			case "category":
				got = entity.Category
			case "title":
				got = entity.Title
			case "date":
				got = entity.Date
			case "description":
				got = entity.Description
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

	entity, _ := s.GetEntity("c3-101")
	if !strings.Contains(entity.Body, "ref-logging") {
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
	// Set up an entity with required sections + a custom section
	entity, _ := s.GetEntity("c3-110")
	entity.Body = "# users\n\n## Goal\n\nManage users.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n| IN | data | c3-101 |\n\n## Custom Section\n\nSome text here.\n"
	s.UpdateEntity(entity)

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

	updated, _ := s.GetEntity("c3-110")
	if !strings.Contains(updated.Body, "value1") {
		t.Error("body should contain the new table data")
	}
}

func TestRunSet_BatchStdin(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	payload := `{"fields":{"goal":"New batch goal","summary":"Batch summary"},"sections":{"Goal":"New goal content from batch."}}`
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
	if entity.Summary != "Batch summary" {
		t.Errorf("summary = %q, want %q", entity.Summary, "Batch summary")
	}
	if !strings.Contains(entity.Body, "New goal content from batch.") {
		t.Error("body should contain updated Goal section")
	}

	output := buf.String()
	if !strings.Contains(output, "2 fields") {
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
