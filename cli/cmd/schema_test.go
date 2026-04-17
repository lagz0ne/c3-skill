package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// =============================================================================
// RunSchema: expose known sections per entity type
// =============================================================================

func TestRunSchema_Component(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("component", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, section := range []string{"Goal", "Parent Fit", "Foundational Flow", "Business Flow", "Governance", "Contract", "Change Safety", "Derived Materials"} {
		if !strings.Contains(output, section) {
			t.Errorf("component schema should include %q, got: %s", section, output)
		}
	}
}

func TestRunSchema_Container(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("container", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, section := range []string{"Goal", "Components", "Responsibilities"} {
		if !strings.Contains(output, section) {
			t.Errorf("container schema should include %q, got: %s", section, output)
		}
	}
}

func TestRunSchema_Ref(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("ref", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, section := range []string{"Goal", "Choice", "Why"} {
		if !strings.Contains(output, section) {
			t.Errorf("ref schema should include %q, got: %s", section, output)
		}
	}
}

func TestRunSchema_ADRIncludesDecisionLedger(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("adr", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, want := range []string{
		"Goal",
		"Context",
		"Decision",
		"Work Breakdown",
		"Underlay C3 Changes",
		"Enforcement Surfaces",
		"Alternatives Considered",
		"Risks",
		"Verification",
		"Current behavior, user pain, constraints, and affected topology",
		"C3 CLI files, validators, commands, hints, help, schemas, templates, or tests",
		"Commands, validators, tests, docs, or runtime paths that enforce the decision",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("ADR schema should include %q, got: %s", want, output)
		}
	}
}

func TestRunSchema_Context(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("context", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, section := range []string{"Goal", "Containers", "Abstract Constraints"} {
		if !strings.Contains(output, section) {
			t.Errorf("context schema should include %q, got: %s", section, output)
		}
	}
}

func TestRunSchema_UnknownType(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("widget", false, &buf)
	if err == nil {
		t.Fatal("expected error for unknown entity type")
	}
}

func TestRunSchema_JSON_ADRUnderlayColumns(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("adr", true, &buf); err != nil {
		t.Fatal(err)
	}

	var schema SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if schema.Type != "adr" {
		t.Errorf("type = %q, want adr", schema.Type)
	}
	for _, s := range schema.Sections {
		if s.Name != "Underlay C3 Changes" {
			continue
		}
		if s.ContentType != "table" {
			t.Errorf("Underlay C3 Changes content_type = %q, want table", s.ContentType)
		}
		for _, col := range []string{"Underlay area", "Exact C3 change", "Verification evidence"} {
			if findColumn(s.Columns, col) == nil {
				t.Fatalf("Underlay C3 Changes should include column %q", col)
			}
		}
		if !strings.Contains(s.Purpose, "validators") {
			t.Fatalf("Underlay C3 Changes purpose should guide enforceable detail, got %q", s.Purpose)
		}
		return
	}
	t.Fatal("Underlay C3 Changes section not found in ADR schema")
}

func TestRunSchema_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := RunSchema("component", true, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var schema SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if schema.Type != "component" {
		t.Errorf("type = %q, want %q", schema.Type, "component")
	}
	if len(schema.Sections) == 0 {
		t.Error("sections should not be empty")
	}

	// Verify required sections are marked
	goalFound := false
	for _, s := range schema.Sections {
		if s.Name == "Goal" {
			goalFound = true
			if !s.Required {
				t.Error("Goal should be required")
			}
			if s.ContentType != "text" {
				t.Errorf("Goal content_type = %q, want %q", s.ContentType, "text")
			}
		}
	}
	if !goalFound {
		t.Error("Goal section not found in schema")
	}
}

func TestRunSchema_JSON_TableColumns(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("component", true, &buf); err != nil {
		t.Fatal(err)
	}

	var schema SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	// Find Contract section — should have typed columns
	for _, s := range schema.Sections {
		if s.Name == "Contract" {
			if s.ContentType != "table" {
				t.Errorf("Contract content_type = %q, want %q", s.ContentType, "table")
			}
			if len(s.Columns) != 5 {
				t.Fatalf("Contract columns count = %d, want 5", len(s.Columns))
			}
			// Verify column types (Layer 3)
			dirCol := findColumn(s.Columns, "Direction")
			if dirCol == nil {
				t.Fatal("Direction column not found")
			}
			if dirCol.Type != "enum" {
				t.Errorf("Direction column type = %q, want %q", dirCol.Type, "enum")
			}
			// Verify enum values are specified
			if len(dirCol.Values) == 0 {
				t.Error("Direction column should specify allowed enum values")
			}
			hasIN := false
			hasOUT := false
			for _, v := range dirCol.Values {
				if v == "IN" {
					hasIN = true
				}
				if v == "OUT" {
					hasOUT = true
				}
			}
			if !hasIN || !hasOUT {
				t.Errorf("Direction enum values should include IN and OUT, got %v", dirCol.Values)
			}

			evidenceCol := findColumn(s.Columns, "Evidence")
			if evidenceCol == nil {
				t.Fatal("Evidence column not found")
			}
			if evidenceCol.Type != "text" {
				t.Errorf("Evidence column type = %q, want %q", evidenceCol.Type, "text")
			}
			return
		}
	}
	t.Error("Contract section not found in schema")
}

func TestRunSchema_SectionOrder(t *testing.T) {
	// Schema sections should come back in a deterministic, template-defined order
	var buf1, buf2 bytes.Buffer
	if err := RunSchema("component", true, &buf1); err != nil {
		t.Fatal(err)
	}
	if err := RunSchema("component", true, &buf2); err != nil {
		t.Fatal(err)
	}

	var s1, s2 SchemaOutput
	json.Unmarshal(buf1.Bytes(), &s1)
	json.Unmarshal(buf2.Bytes(), &s2)

	if len(s1.Sections) != len(s2.Sections) {
		t.Fatalf("section count differs: %d vs %d", len(s1.Sections), len(s2.Sections))
	}
	for i := range s1.Sections {
		if s1.Sections[i].Name != s2.Sections[i].Name {
			t.Errorf("section order differs at index %d: %q vs %q", i, s1.Sections[i].Name, s2.Sections[i].Name)
		}
	}

	// Goal should come before Parent Fit (template order)
	goalIdx, parentFitIdx := -1, -1
	for i, s := range s1.Sections {
		if s.Name == "Goal" {
			goalIdx = i
		}
		if s.Name == "Parent Fit" {
			parentFitIdx = i
		}
	}
	if goalIdx == -1 || parentFitIdx == -1 {
		t.Fatal("Goal and Parent Fit sections must both exist")
	}
	if goalIdx >= parentFitIdx {
		t.Errorf("Goal (index %d) should come before Parent Fit (index %d)", goalIdx, parentFitIdx)
	}
}

func TestRunSchema_Component_NoCodeReferences(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("component", true, &buf); err != nil {
		t.Fatal(err)
	}

	var schema SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	for _, s := range schema.Sections {
		if s.Name == "Code References" {
			t.Error("component schema should not contain Code References section")
		}
	}
}

func TestRunSchema_Ref_NoCitedBy(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("ref", true, &buf); err != nil {
		t.Fatal(err)
	}

	var schema SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	for _, s := range schema.Sections {
		if s.Name == "Cited By" {
			t.Error("ref schema should not contain Cited By section")
		}
	}
}

func findColumn(cols []schema.ColumnDef, name string) *schema.ColumnDef {
	for i := range cols {
		if cols[i].Name == name {
			return &cols[i]
		}
	}
	return nil
}
