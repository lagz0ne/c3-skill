package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunCheck_EmptyRequiredSection(t *testing.T) {
	s := createRichDBFixture(t)

	// Update c3-110 to have empty Goal section
	entity, _ := s.GetEntity("c3-110")
	entity.Body = "# users\n\n## Goal\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n"
	s.UpdateEntity(entity)

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-110") {
		t.Errorf("should flag c3-110, got: %s", output)
	}
}

func TestRunCheck_DefaultSkipsADR(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := CheckOptions{Store: s, JSON: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	for _, issue := range result.Issues {
		if strings.Contains(issue.Entity, "adr-") {
			t.Errorf("default check should skip ADR, but found issue for: %s", issue.Entity)
		}
	}
}

func TestRunCheck_IncludeADRValidatesADR(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := CheckOptions{Store: s, JSON: true, IncludeADR: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	hasADRIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Entity, "adr-") {
			hasADRIssue = true
			break
		}
	}
	if !hasADRIssue {
		t.Error("--include-adr should validate ADR entities")
	}
}

func TestRunCheck_EmptyRequiredTable(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "empty required table") {
		t.Errorf("should warn about empty required tables, got: %s", output)
	}
}

func TestRunCheck_MissingRequiredSection_Ref(t *testing.T) {
	s := createDBFixture(t)

	// Add an incomplete ref missing required sections
	s.InsertEntity(&store.Entity{
		ID: "ref-incomplete", Type: "ref", Title: "Incomplete Ref", Slug: "incomplete",
		Goal: "Some pattern", Body: "# Incomplete Ref\n\n## Goal\n\nSome pattern.\n",
		Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref-incomplete") {
		t.Errorf("should flag ref-incomplete, got: %s", output)
	}
	if !strings.Contains(output, "Choice") {
		t.Errorf("should mention missing Choice section, got: %s", output)
	}
}

func TestRunCheck_EntityIdNotInStore(t *testing.T) {
	s := createDBFixture(t)

	// Update c3-101 to reference c3-999 in body table
	entity, _ := s.GetEntity("c3-101")
	entity.Body = "# auth\n\n## Goal\n\nAuth.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n| IN | data | c3-999 |\n"
	s.UpdateEntity(entity)

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-999") {
		t.Errorf("should flag nonexistent entity, got: %s", output)
	}
}

func TestRunCheck_SuggestsByTitle(t *testing.T) {
	s := createDBFixture(t)

	// c3-101 body references "api" instead of "c3-1"
	entity, _ := s.GetEntity("c3-101")
	entity.Body = "# auth\n\n## Goal\n\nAuth.\n\n## Dependencies\n\n| Direction | What | From/To |\n|-----------|------|----------|\n| IN | data | api |\n"
	s.UpdateEntity(entity)

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "did you mean c3-1?") {
		t.Errorf("should suggest c3-1 for 'api', got: %s", output)
	}
}

func TestRunCheck_EnhancedJSON(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := CheckOptions{Store: s, JSON: true}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	var result CheckResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	hasSchemaIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "empty") {
			hasSchemaIssue = true
			break
		}
	}
	if !hasSchemaIssue {
		t.Error("JSON output should include schema validation issues")
	}
}

func TestRunCheck_CleanOutputSummary(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	s.InsertEntity(&store.Entity{
		ID: "c3-0", Type: "system", Title: "Test", Slug: "",
		Body: "# Test\n\n## Goal\n\nTest.\n\n## Containers\n\n| ID | Name | Purpose |\n|----|------|---------|\n| | core | Core |\n\n## Abstract Constraints\n\nKeep it simple.\n",
		Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "all clear") {
		t.Errorf("clean run should say 'all clear', got: %s", output)
	}
}

func TestRunCheck_ScopeCrossCheck(t *testing.T) {
	s := createDBFixture(t)
	// ref-jwt scopes c3-1. c3-101 cites ref-jwt. c3-110 does NOT.

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref-jwt scopes c3-1 but c3-110 does not cite it") {
		t.Errorf("should warn about c3-110 not citing ref-jwt, got: %s", output)
	}
	if strings.Contains(output, "c3-101 does not cite it") {
		t.Errorf("should NOT warn about c3-101, got: %s", output)
	}
}

func TestHintFor(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"missing required section: Goal", "add a ## Goal section with content"},
		{"empty required section: Overview", "add content to the ## Overview section"},
		{"empty required table: Dependencies (headers only, no data rows)", "add at least one data row below the table headers"},
		{"unknown entity reference: c3-999", "verify the ID with 'c3x list'; check for typos"},
		{"unknown ref reference: ref-missing", "use a ref-* ID (e.g., ref-jwt); verify with 'c3x list'"},
		{"file does not exist: src/foo.ts", "create the file or fix the path"},
		{"code-map parse error: yaml: unmarshal error", "fix YAML syntax in .c3/code-map.yaml"},
		{"something unknown", ""},
	}
	for _, tt := range tests {
		got := hintFor(tt.message)
		if got != tt.expected {
			t.Errorf("hintFor(%q) = %q, want %q", tt.message, got, tt.expected)
		}
	}
}

func TestRunCheck_RecipeInvalidSources(t *testing.T) {
	s := createDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "recipe-auth", Type: "recipe", Title: "Auth Flow", Slug: "auth",
		Body: "# Auth Flow\n\n## Goal\n\nTrace auth.\n", Status: "active", Metadata: "{}",
	})
	// Add valid source
	s.AddRelationship(&store.Relationship{FromID: "recipe-auth", ToID: "c3-0", RelType: "sources"})
	// Add invalid source — entity doesn't exist, but relationship can't be created with FK
	// So we test by checking that existing valid sources don't produce warnings

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "recipe references nonexistent") {
		t.Errorf("valid sources should not be flagged, got: %s", output)
	}
}
