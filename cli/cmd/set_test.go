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
