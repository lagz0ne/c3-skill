package cmd

// LLM Entity Creation Eval Framework
//
// Tests the LLM authoring experience: can an LLM produce valid C3 entities
// using only the CLI's schema/template output, without multiple retry rounds?
//
// Test naming:
//   TestEval_*  — eval-framework tests (RED = gap exists, GREEN = gap fixed)
//
// Theory: LLMs burn multiple rounds because:
//   1. Schema output hides validation constraints (min words, min rows, enums, order, N.A format)
//   2. No template command exists — LLM must construct markdown from scratch
//   3. ADR schema says sections optional but creation demands all (lies)
//   4. No dry-run validation — each attempt either creates or explodes

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/templates"
)

// ---------------------------------------------------------------------------
// SECTION 1: Template → Validation Roundtrip
// The embedded templates should pass their own validation when placeholders
// are filled with minimal valid content. This proves templates can serve as
// one-shot LLM scaffolds.
// ---------------------------------------------------------------------------

func TestEval_TemplateCommandExists(t *testing.T) {
	// RED: RunTemplate function should exist as a CLI command
	// When implemented, `c3x template <type>` outputs a fillable scaffold
	var buf bytes.Buffer
	err := RunTemplate("component", &buf)
	if err != nil {
		t.Fatalf("RunTemplate should exist and succeed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("template output should not be empty")
	}
}

func TestEval_TemplateComponentPassesValidation(t *testing.T) {
	// RED: template output should pass validateBodyContent when placeholders
	// are replaced with minimal valid content
	var buf bytes.Buffer
	if err := RunTemplate("component", &buf); err != nil {
		t.Skipf("RunTemplate not available: %v", err)
	}

	body := buf.String()
	issues := validateBodyContent(body, "component")
	if len(issues) > 0 {
		t.Errorf("template with filled placeholders should pass validation, got %d issues:", len(issues))
		for _, issue := range issues {
			t.Errorf("  %s: %s (hint: %s)", issue.Severity, issue.Message, issue.Hint)
		}
	}
}

func TestEval_TemplateADRPassesValidation(t *testing.T) {
	var buf bytes.Buffer
	if err := RunTemplate("adr", &buf); err != nil {
		t.Skipf("RunTemplate not available: %v", err)
	}

	body := buf.String()
	issues := validateADRCreationBody(body)
	if len(issues) > 0 {
		t.Errorf("ADR template should pass creation validation, got %d issues:", len(issues))
		for _, issue := range issues {
			t.Errorf("  %s: %s", issue.Severity, issue.Message)
		}
	}
}

func TestEval_TemplateAllTypesAvailable(t *testing.T) {
	for _, entityType := range []string{"component", "container", "ref", "rule", "adr", "recipe"} {
		t.Run(entityType, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunTemplate(entityType, &buf); err != nil {
				t.Fatalf("RunTemplate(%q) failed: %v", entityType, err)
			}
			if buf.Len() == 0 {
				t.Fatalf("template for %q should not be empty", entityType)
			}
		})
	}
}

func TestEval_TemplateContainsConstraintAnnotations(t *testing.T) {
	// Templates should contain inline constraint comments so LLMs
	// know the rules WITHOUT needing to fail first.
	var buf bytes.Buffer
	if err := RunTemplate("component", &buf); err != nil {
		t.Skipf("RunTemplate not available: %v", err)
	}

	body := buf.String()

	// Must mention minimum word counts
	if !strings.Contains(body, "4 words") && !strings.Contains(body, "min 4") {
		t.Error("template should mention Goal minimum word count")
	}
	if !strings.Contains(body, "12 words") && !strings.Contains(body, "min 12") {
		t.Error("template should mention Purpose minimum word count")
	}

	// Must mention N.A format
	if !strings.Contains(body, "N.A - ") {
		t.Error("template should show N.A - <reason> format")
	}

	// Must mention placeholder ban
	if !strings.Contains(body, "TBD") || !strings.Contains(body, "TODO") {
		t.Error("template should warn about banned placeholder words")
	}
}

// ---------------------------------------------------------------------------
// SECTION 2: Schema Constraint Visibility
// Schema output should contain enough info for an LLM to produce valid content
// on the FIRST attempt. If constraints are invisible, the LLM must fail-and-retry.
// ---------------------------------------------------------------------------

func TestEval_SchemaShowsMinWordCount(t *testing.T) {
	// RED: schema output for component should mention word minimums
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "4") || !containsAny(out, "word", "min") {
		t.Error("schema should show Goal minimum word count (4 words)")
	}
}

func TestEval_SchemaShowsMinRowCount(t *testing.T) {
	// RED: schema output should mention minimum row counts for tables
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	expectations := map[string]int{
		"Parent Fit":        4,
		"Foundational Flow": 4,
		"Business Flow":     4,
		"Contract":          2,
		"Change Safety":     2,
	}
	for section, minRows := range expectations {
		if !strings.Contains(out, section) {
			t.Errorf("schema should mention section %q", section)
			continue
		}
		// The schema output should mention the minimum row count near the section name
		if !strings.Contains(out, "min") {
			t.Errorf("schema should mention minimum rows for %s (need %d)", section, minRows)
		}
	}
}

func TestEval_SchemaShowsSectionOrder(t *testing.T) {
	// RED: schema output should mention that section order matters for components
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !containsAny(out, "order", "sequence", "must appear in order") {
		t.Error("schema should mention that component sections must be in order")
	}
}

func TestEval_SchemaShowsPlaceholderBan(t *testing.T) {
	// RED: schema output should warn about rejected placeholder words
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !containsAny(out, "placeholder", "TBD", "TODO") {
		t.Error("schema should warn about banned placeholder words (TBD, TODO, maybe, optional, later)")
	}
}

func TestEval_SchemaShowsNAFormat(t *testing.T) {
	// Schema output should clearly show the N.A - <reason> format requirement.
	// The enum values in schema already include "N.A - <reason>" but the schema
	// text output doesn't render these clearly enough.
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "N.A - ") {
		t.Error("schema should show the N.A - <reason> format requirement")
	}
}

func TestEval_SchemaShowsEvidenceGroundingRequirement(t *testing.T) {
	// RED: schema should mention that Evidence/Reference columns need
	// grounded content (entity IDs, file paths, commands) — not prose
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !containsAny(out, "grounded", "entity id", "file path", "command") {
		t.Error("schema should mention evidence grounding requirement")
	}
}

// ---------------------------------------------------------------------------
// SECTION 3: ADR Schema Honesty
// ADR schema marks most sections Required: false, but validateADRCreationBody
// requires ALL sections. This mismatch causes LLMs to trust the schema, skip
// sections, and fail at creation time.
// ---------------------------------------------------------------------------

func TestEval_ADRSchemaReflectsCreationReality(t *testing.T) {
	// RED: Every ADR section that validateADRCreationBody demands
	// should be marked Required: true in the schema registry.
	adrSections := schema.ForType("adr")
	if adrSections == nil {
		t.Fatal("adr schema should exist")
	}

	for _, sec := range adrSections {
		if !sec.Required {
			t.Errorf("ADR section %q is Required:false in schema but required at creation time — schema lies to LLMs", sec.Name)
		}
	}
}

func TestEval_ADRSchemaShowsAllOrNothing(t *testing.T) {
	// RED: Schema output should make the all-or-nothing rule visible.
	// Currently only Goal shows "(required)" — the other 8 sections show nothing,
	// yet creation demands ALL of them. Schema must warn about this.
	var buf bytes.Buffer
	if err := RunSchema("adr", false, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !containsAny(out, "all-or-nothing", "all sections required at creation") {
		t.Error("ADR schema output should explicitly mention all-or-nothing creation requirement")
	}
}

// ---------------------------------------------------------------------------
// SECTION 4: Dry-Run Validation
// LLMs should be able to validate content without creating entities.
// This prevents the create-fail-cleanup cycle.
// ---------------------------------------------------------------------------

func TestEval_AddDryRunValidatesWithoutCreating(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nJust a goal.\n"
	err := RunAddDryRun("component", "test-dry", s, "c3-1", false,
		strings.NewReader(body), &buf)
	// Should return validation errors without creating the entity
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "Parent Fit") {
		t.Errorf("dry-run should report validation errors: %v", err)
	}

	// Entity should NOT exist
	if _, getErr := s.GetEntity("c3-102"); getErr == nil {
		t.Error("dry-run should not create an entity")
	}
}

func TestEval_AddDryRunReportsSuccessWithoutCreating(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := strictComponentBody("test-ok", "Documents valid component behavior for dry-run verification.")
	err := RunAddDryRun("component", "test-ok", s, "c3-1", false,
		strings.NewReader(body), &buf)
	if err != nil {
		t.Fatalf("dry-run with valid content should succeed: %v", err)
	}
	if !containsAny(buf.String(), "valid", "pass", "ok") {
		t.Error("dry-run success should report that content is valid")
	}

	// Entity should NOT exist even on success
	if _, getErr := s.GetEntity("c3-102"); getErr == nil {
		t.Error("dry-run should not create an entity even when valid")
	}
}

// ---------------------------------------------------------------------------
// SECTION 5: Error Actionability Baseline
// These tests document the EXISTING behavior (should pass).
// They prove that error batching and hints work, establishing
// a baseline for the eval framework.
// ---------------------------------------------------------------------------

func TestEval_Baseline_ErrorsBatched(t *testing.T) {
	// GREEN: Submit content with multiple issues → get ALL errors at once
	body := "## Goal\nHi.\n" // thin goal, missing 8 sections
	issues := validateBodyContent(body, "component")

	if len(issues) < 5 {
		t.Errorf("expected at least 5 batched issues for minimal component, got %d", len(issues))
	}
}

func TestEval_Baseline_EveryIssueHasHint(t *testing.T) {
	// GREEN: Every validation issue should have a non-empty hint
	body := "## Goal\nHi.\n"
	issues := validateBodyContent(body, "component")

	for _, issue := range issues {
		if issue.Hint == "" {
			t.Errorf("issue %q has no hint — LLM can't fix what it doesn't understand", issue.Message)
		}
	}
}

func TestEval_Baseline_ADRErrorsBatched(t *testing.T) {
	// GREEN: ADR with only Goal → should report ALL missing sections at once
	body := "## Goal\nAdopt OAuth.\n"
	issues := validateADRCreationBody(body)

	// Should get errors for Context, Decision, Work Breakdown, etc.
	if len(issues) < 8 {
		t.Errorf("expected 8+ issues for ADR with only Goal, got %d", len(issues))
	}
}

func TestEval_Baseline_PlaceholderDetection(t *testing.T) {
	// GREEN: Placeholder words should be detected
	body := strictComponentBody("auth", "Provide TBD authentication behavior for API requests.")
	issues := validateStrictComponentDoc(body, "error")

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "placeholder") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should detect placeholder 'TBD' in Purpose")
	}
}

func TestEval_Baseline_NADotFormatDetection(t *testing.T) {
	// GREEN: "N.A" without reason should be detected in table cells
	body := strings.Replace(
		strictComponentBody("auth", "Provide authentication behavior for API requests."),
		"| ref-jwt | ref |", "| N.A | ref |", 1,
	)
	issues := validateStrictComponentDoc(body, "error")

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "invalid N.A") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should detect N.A without reason in table cells")
	}
}

func TestEval_NASlashFormatDetection(t *testing.T) {
	// RED: LLMs commonly write "N/A" instead of "N.A - <reason>".
	// The validator should catch this common LLM mistake.
	// Currently it only checks for "N.A" string, missing "N/A" entirely.
	body := strings.Replace(
		strictComponentBody("auth", "Provide authentication behavior for API requests."),
		"| ref-jwt | ref |", "| N/A | ref |", 1,
	)
	issues := validateStrictComponentDoc(body, "error")

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "N.A") || strings.Contains(issue.Message, "N/A") || strings.Contains(issue.Message, "invalid") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should detect N/A (slash format) — LLMs commonly write this instead of N.A - <reason>")
	}
}

// ---------------------------------------------------------------------------
// SECTION 6: Embedded Template Quality
// The templates in cli/internal/templates/ should be usable as scaffolds.
// Test that they contain all required sections and correct table structures.
// ---------------------------------------------------------------------------

func TestEval_EmbeddedTemplateHasAllRequiredSections(t *testing.T) {
	types := map[string]string{
		"component": "component.md",
		"container": "container.md",
		"ref":       "ref.md",
		"rule":      "rule.md",
		"adr":       "adr.md",
		"recipe":    "recipe.md",
	}

	for entityType, filename := range types {
		t.Run(entityType, func(t *testing.T) {
			content, err := templates.Read(filename)
			if err != nil {
				t.Fatalf("failed to read template %s: %v", filename, err)
			}

			schemaSections := schema.ForType(entityType)
			for _, sec := range schemaSections {
				if sec.Required {
					if !strings.Contains(content, "## "+sec.Name) {
						t.Errorf("template %s missing required section ## %s", filename, sec.Name)
					}
				}
			}
		})
	}
}

func TestEval_EmbeddedADRTemplateHasTableHeaders(t *testing.T) {
	// ADR template tables should have correct column headers matching schema
	content, err := templates.Read("adr.md")
	if err != nil {
		t.Fatal(err)
	}

	adrSections := schema.ForType("adr")
	for _, sec := range adrSections {
		if sec.ContentType != "table" || len(sec.Columns) == 0 {
			continue
		}
		for _, col := range sec.Columns {
			if !strings.Contains(content, col.Name) {
				t.Errorf("ADR template missing column %q in section %s", col.Name, sec.Name)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
