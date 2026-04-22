package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunCheck_EmptyRequiredSection(t *testing.T) {
	s := createRichDBFixture(t)

	// Update c3-110 to have empty Goal section
	content.WriteEntity(s, "c3-110", strings.Replace(strictComponentBody("users", "Manage user account behavior for authenticated API requests."), "Manage user account behavior for authenticated API requests.", "", 1))

	var buf bytes.Buffer
	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err == nil {
		t.Fatal("expected check failure")
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
	content.WriteEntity(s, "c3-110", strings.Replace(strictComponentBody("users", "Manage user account behavior for authenticated API requests."), "| credentials | IN | Accept credential material for validation only. | API request boundary | ref-jwt |\n| identity result | OUT | Provide accepted identity or explicit rejection. | Downstream user workflow | c3-110 |", "", 1))
	var buf bytes.Buffer

	opts := CheckOptions{Store: s, JSON: false}
	if err := RunCheckV2(opts, &buf); err == nil {
		t.Fatal("expected check failure")
	}

	output := buf.String()
	if !strings.Contains(output, "not enough rows in Contract") {
		t.Errorf("should warn about thin Contract table, got: %s", output)
	}
}

func TestRunCheck_MissingRequiredSection_Ref(t *testing.T) {
	s := createRichDBFixture(t)

	// Add an incomplete ref missing required sections
	s.InsertEntity(&store.Entity{
		ID: "ref-incomplete", Type: "ref", Title: "Incomplete Ref", Slug: "incomplete",
		Goal: "Some pattern", Status: "active", Metadata: "{}",
	})
	content.WriteEntity(s, "ref-incomplete", "# Incomplete Ref\n\n## Goal\n\nSome pattern.\n")

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
	s := createRichDBFixture(t)

	// Update c3-1 to reference c3-999 in Components table
	content.WriteEntity(s, "c3-1", "# api\n\n## Goal\n\nServe API requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| c3-999 | ghost | feature | active | Missing component |\n\n## Responsibilities\n\nServe API requests.\n")

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
	s := createRichDBFixture(t)

	// c3-1 Components table references "api" instead of "c3-1"
	content.WriteEntity(s, "c3-1", "# api\n\n## Goal\n\nServe API requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| api | ghost | feature | active | Ambiguous title reference |\n\n## Responsibilities\n\nServe API requests.\n")

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
	content.WriteEntity(s, "c3-110", strings.Replace(strictComponentBody("users", "Manage user account behavior for authenticated API requests."), "Manage user account behavior for authenticated API requests.", "", 1))
	var buf bytes.Buffer

	opts := CheckOptions{Store: s, JSON: true}
	if err := RunCheckV2(opts, &buf); err == nil {
		t.Fatal("expected check failure")
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
		Status: "active", Metadata: "{}",
	})
	content.WriteEntity(s, "c3-0", "# Test\n\n## Goal\n\nTest.\n\n## Containers\n\n| ID | Name | Purpose |\n|----|------|---------|\n| | core | Core |\n\n## Abstract Constraints\n\nKeep it simple.\n")

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
	s := createRichDBFixture(t)
	// ref-jwt scopes c3-1. c3-101 cites ref-jwt. c3-110 does NOT.
	s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "ref-error-handling", RelType: "uses"})

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

func TestRunCheck_LayerDisconnectMissingComponentInContainer(t *testing.T) {
	s := createRichDBFixture(t)
	content.WriteEntity(s, "c3-1", "# api\n\n## Goal\n\nServe API requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| c3-101 | auth | foundation | active | Authentication |\n\n## Responsibilities\n\nServe API requests.\n")

	var buf bytes.Buffer
	if err := RunCheckV2(CheckOptions{Store: s, JSON: false}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"layer disconnect",
		"c3-110",
		"missing from c3-1 Components table",
	)
}

func TestRunCheck_LayerDisconnectStaleComponentInContainer(t *testing.T) {
	s := createRichDBFixture(t)
	content.WriteEntity(s, "c3-1", "# api\n\n## Goal\n\nServe API requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n| c3-101 | auth | foundation | active | Authentication |\n| c3-201 | renderer | feature | active | Wrong parent |\n\n## Responsibilities\n\nServe API requests.\n")

	var buf bytes.Buffer
	if err := RunCheckV2(CheckOptions{Store: s, JSON: false}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"layer disconnect",
		"c3-201",
		"listed in c3-1 Components table but parent is c3-2",
	)
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
		{"code-map parse error: yaml: unmarshal error", "check code-map entries with 'c3x list'"},
		{"layer disconnect: child component c3-110 has parent c3-1 but is missing from c3-1 Components table", "update parent table or fix the child parent field; rebuild only proves storage, not layer integration"},
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
	s := createRichDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "recipe-auth", Type: "recipe", Title: "Auth Flow", Slug: "auth",
		Status: "active", Metadata: "{}",
	})
	content.WriteEntity(s, "recipe-auth", "# Auth Flow\n\n## Goal\n\nTrace auth.\n")
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

func TestValidateColumn_EntityID(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	table := &markdown.Table{
		Headers: []string{"Direction", "What", "From/To"},
		Rows: []map[string]string{
			{"Direction": "IN", "What": "data", "From/To": "c3-999"},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "From/To", Type: "entity_id"}

	issues := validateColumn(col, table, entity, CheckOptions{Store: s}, titleMap)
	if len(issues) == 0 {
		t.Error("should report unknown entity reference")
	}
}

func TestValidateColumn_EntityID_Valid(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	table := &markdown.Table{
		Headers: []string{"Direction", "What", "From/To"},
		Rows: []map[string]string{
			{"Direction": "IN", "What": "data", "From/To": "c3-110"},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "From/To", Type: "entity_id"}

	issues := validateColumn(col, table, entity, CheckOptions{Store: s}, titleMap)
	if len(issues) != 0 {
		t.Errorf("valid entity reference should produce no issues, got %d", len(issues))
	}
}

func TestValidateColumn_RefID(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	table := &markdown.Table{
		Headers: []string{"Ref", "Role"},
		Rows: []map[string]string{
			{"Ref": "ref-nonexistent", "Role": "test"},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "Ref", Type: "ref_id"}

	issues := validateColumn(col, table, entity, CheckOptions{Store: s}, titleMap)
	if len(issues) == 0 {
		t.Error("should report unknown ref reference")
	}
}

func TestValidateColumn_RefID_Valid(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	table := &markdown.Table{
		Headers: []string{"Ref", "Role"},
		Rows: []map[string]string{
			{"Ref": "ref-jwt", "Role": "auth"},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "Ref", Type: "ref_id"}

	issues := validateColumn(col, table, entity, CheckOptions{Store: s}, titleMap)
	if len(issues) != 0 {
		t.Errorf("valid ref reference should produce no issues, got %d", len(issues))
	}
}

func TestValidateColumn_Filepath(t *testing.T) {
	s := createRichDBFixture(t)
	projectDir := t.TempDir()

	table := &markdown.Table{
		Headers: []string{"File", "Purpose"},
		Rows: []map[string]string{
			{"File": "nonexistent/file.ts", "Purpose": "test"},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "File", Type: "filepath"}

	issues := validateColumn(col, table, entity, CheckOptions{Store: s, ProjectDir: projectDir}, nil)
	if len(issues) == 0 {
		t.Error("should report file does not exist")
	}
}

func TestValidateColumn_Filepath_NoProjectDir(t *testing.T) {
	s := createRichDBFixture(t)

	table := &markdown.Table{
		Headers: []string{"File", "Purpose"},
		Rows: []map[string]string{
			{"File": "nonexistent/file.ts", "Purpose": "test"},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "File", Type: "filepath"}

	// Without ProjectDir, filepath validation is skipped
	issues := validateColumn(col, table, entity, CheckOptions{Store: s}, nil)
	if len(issues) != 0 {
		t.Errorf("should skip filepath validation without ProjectDir, got %d issues", len(issues))
	}
}

func TestValidateColumn_EmptyValues(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	table := &markdown.Table{
		Headers: []string{"Direction", "What", "From/To"},
		Rows: []map[string]string{
			{"Direction": "IN", "What": "data", "From/To": ""},
		},
	}
	entity := &store.Entity{ID: "c3-101"}
	col := schema.ColumnDef{Name: "From/To", Type: "entity_id"}

	issues := validateColumn(col, table, entity, CheckOptions{Store: s}, titleMap)
	if len(issues) != 0 {
		t.Errorf("empty values should be skipped, got %d issues", len(issues))
	}
}

func TestFormatCounts(t *testing.T) {
	tests := []struct {
		errors, warnings int
		want             string
	}{
		{1, 0, "1 error"},
		{2, 0, "2 errors"},
		{0, 1, "1 warning"},
		{0, 2, "2 warnings"},
		{1, 1, "1 error, 1 warning"},
		{3, 2, "3 errors, 2 warnings"},
	}
	for _, tt := range tests {
		got := formatCounts(tt.errors, tt.warnings)
		if got != tt.want {
			t.Errorf("formatCounts(%d, %d) = %q, want %q", tt.errors, tt.warnings, got, tt.want)
		}
	}
}

func TestCountSeverities(t *testing.T) {
	issues := []Issue{
		{Severity: "error"},
		{Severity: "warning"},
		{Severity: "error"},
		{Severity: "warning"},
		{Severity: "warning"},
	}
	e, w := countSeverities(issues)
	if e != 2 || w != 3 {
		t.Errorf("countSeverities = (%d, %d), want (2, 3)", e, w)
	}
}

func TestCountSeverities_Empty(t *testing.T) {
	e, w := countSeverities(nil)
	if e != 0 || w != 0 {
		t.Errorf("countSeverities(nil) = (%d, %d), want (0, 0)", e, w)
	}
}

func TestBuildTitleMapStore(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	// Should map title -> entity ID
	if id := titleMap["auth"]; id != "c3-101" {
		t.Errorf("titleMap[auth] = %q, want c3-101", id)
	}
	if id := titleMap["jwt authentication"]; id != "ref-jwt" {
		t.Errorf("titleMap[jwt authentication] = %q, want ref-jwt", id)
	}
}

func TestSuggestByTitle(t *testing.T) {
	s := createRichDBFixture(t)
	titleMap := buildTitleMapStore(s)

	if id := suggestByTitle("auth", titleMap); id != "c3-101" {
		t.Errorf("suggestByTitle(auth) = %q, want c3-101", id)
	}
	if id := suggestByTitle("nonexistent", titleMap); id != "" {
		t.Errorf("suggestByTitle(nonexistent) = %q, want empty", id)
	}
}


// TestRunCheck_RuleFilter verifies --rule expands to citer entities.
func TestRunCheck_RuleFilter(t *testing.T) {
	s := createRichDBFixture(t)
	// Add a rule cited by c3-101.
	if err := s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging",
		Slug: "logging", Status: "active", Metadata: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	content.WriteEntity(s, "rule-logging", "# Structured Logging\n\n## Goal\n\nConsistent logs.\n\n## Choice\n\nJSON.\n\n## Why\n\nMachine-parseable.\n")
	if err := s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "rule-logging", RelType: "uses"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunCheckV2(CheckOptions{Store: s, JSON: true, Rules: []string{"rule-logging"}}, &buf)
	// Errors are possible but every issue in the output must target c3-101.
	_ = err
	var result CheckResult
	if jerr := json.Unmarshal(buf.Bytes(), &result); jerr != nil {
		t.Fatalf("invalid JSON: %v\n%s", jerr, buf.String())
	}
	for _, issue := range result.Issues {
		if issue.Entity != "" && issue.Entity != "c3-101" {
			t.Errorf("--rule filter leaked to non-citer %s: %+v", issue.Entity, issue)
		}
	}
}

// TestRunCheck_RuleFilterNoCiters errors loudly when the rule has no citers.
func TestRunCheck_RuleFilterNoCiters(t *testing.T) {
	s := createRichDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "rule-unused", Type: "rule", Title: "Unused",
		Slug: "unused", Status: "active", Metadata: "{}",
	})
	var buf bytes.Buffer
	err := RunCheckV2(CheckOptions{Store: s, JSON: true, Rules: []string{"rule-unused"}}, &buf)
	if err == nil {
		t.Fatal("expected error for rule with no citers")
	}
	if !strings.Contains(err.Error(), "no citers") {
		t.Errorf("expected 'no citers' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "c3x wire") {
		t.Errorf("expected actionable hint referencing 'c3x wire', got: %v", err)
	}
}

