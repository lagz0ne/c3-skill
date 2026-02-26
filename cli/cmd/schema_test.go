package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
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
	for _, section := range []string{"Goal", "Dependencies", "Code References"} {
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

	// Find Dependencies section — should have typed columns
	for _, s := range schema.Sections {
		if s.Name == "Dependencies" {
			if s.ContentType != "table" {
				t.Errorf("Dependencies content_type = %q, want %q", s.ContentType, "table")
			}
			if len(s.Columns) != 3 {
				t.Fatalf("Dependencies columns count = %d, want 3", len(s.Columns))
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

			fromCol := findColumn(s.Columns, "From/To")
			if fromCol == nil {
				t.Fatal("From/To column not found")
			}
			if fromCol.Type != "entity_id" {
				t.Errorf("From/To column type = %q, want %q", fromCol.Type, "entity_id")
			}
			return
		}
	}
	t.Error("Dependencies section not found in schema")
}

func TestRunSchema_JSON_CodeRefColumns(t *testing.T) {
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
			fileCol := findColumn(s.Columns, "File")
			if fileCol == nil {
				t.Fatal("File column not found")
			}
			if fileCol.Type != "filepath" {
				t.Errorf("File column type = %q, want %q", fileCol.Type, "filepath")
			}
			return
		}
	}
	t.Error("Code References section not found in schema")
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

	// Goal should come before Dependencies (template order)
	goalIdx, depsIdx := -1, -1
	for i, s := range s1.Sections {
		if s.Name == "Goal" {
			goalIdx = i
		}
		if s.Name == "Dependencies" {
			depsIdx = i
		}
	}
	if goalIdx == -1 || depsIdx == -1 {
		t.Fatal("Goal and Dependencies sections must both exist")
	}
	if goalIdx >= depsIdx {
		t.Errorf("Goal (index %d) should come before Dependencies (index %d)", goalIdx, depsIdx)
	}
}

func findColumn(cols []ColumnDef, name string) *ColumnDef {
	for i := range cols {
		if cols[i].Name == name {
			return &cols[i]
		}
	}
	return nil
}
