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
	if strings.Contains(output, "Up Cap") {
		t.Errorf("component schema should not require Up Cap, got: %s", output)
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
		"Affected Topology",
		"Compliance Refs",
		"Compliance Rules",
		"Work Breakdown",
		"Underlay C3 Changes",
		"Enforcement Surfaces",
		"Alternatives Considered",
		"Risks",
		"Verification",
		"Current behavior, user pain, constraints, and affected topology",
		"C3 CLI files, validators, commands, hints, help, schemas, templates, or tests",
		"Commands, validators, tests, docs, or runtime paths that enforce the decision",
		"fill:",
		"rejected when:",
		"REJECT IF:",
		"Run c3x schema adr before drafting",
		"volatile Discovery Brief",
		"owner, governing material, stop condition",
		"Compliance rows must say why the ref/rule applies",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("ADR schema should include %q, got: %s", want, output)
		}
	}

	if strings.Contains(output, "Template:") {
		t.Errorf("ADR schema should not mention retired ADR templates, got: %s", output)
	}
	if strings.Contains(output, "if weak/missing:") {
		t.Errorf("ADR schema should NOT contain renamed label %q", "if weak/missing:")
	}
	if strings.Contains(output, "ADR rules:") {
		t.Errorf("ADR schema should NOT contain old %q header (merged into REJECT IF)", "ADR rules:")
	}
}

func TestRunSchema_ADRUsesCanvasDefinition(t *testing.T) {
	sections := schema.ForType("adr")
	def, ok := schema.DefinitionFor("adr")
	if !ok {
		t.Fatal("adr definition missing")
	}
	if len(sections) != len(def.Sections) {
		t.Fatalf("sections = %d, definition sections = %d", len(sections), len(def.Sections))
	}
	for i := range sections {
		if sections[i].Name != def.Sections[i].Name {
			t.Fatalf("section %d = %q, want %q", i, sections[i].Name, def.Sections[i].Name)
		}
	}
}

func TestRunSchema_ADR_RejectIfFiresOnVagueRows(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("adr", false, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	rejectIdx := strings.Index(output, "REJECT IF:")
	if rejectIdx < 0 {
		t.Fatal("REJECT IF: block not found")
	}
	sectionsIdx := strings.Index(output, "Goal [text]")
	if sectionsIdx < 0 {
		t.Fatal("section listing not found")
	}
	if rejectIdx >= sectionsIdx {
		t.Errorf("REJECT IF: block must precede section listing (rejectIdx=%d sectionsIdx=%d)", rejectIdx, sectionsIdx)
	}
}

func TestRunSchema_Ref_HasRejectIfBlock(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("ref", false, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	for _, want := range []string{
		"REJECT IF:",
		"rationale",
		"fill:",
		"rejected when:",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("ref schema should include %q, got: %s", want, output)
		}
	}
}

func TestRunSchema_Rule_HasRejectIfBlock(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("rule", false, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	for _, want := range []string{
		"REJECT IF:",
		"Golden Example",
		"literal code",
		"fill:",
		"rejected when:",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("rule schema should include %q, got: %s", want, output)
		}
	}
}

func TestRunSchema_Component_NoRejectIfBlock(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("component", false, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if strings.Contains(output, "REJECT IF:") {
		t.Errorf("component schema should NOT include REJECT IF: block (out of scope)")
	}
	if !strings.Contains(output, "Component rules:") {
		t.Errorf("component schema should still contain %q block", "Component rules:")
	}
}

func TestRunSchema_Container_NoRejectIfBlock(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("container", false, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if strings.Contains(output, "REJECT IF:") {
		t.Errorf("container schema should NOT include REJECT IF: block (out of scope)")
	}
}

func TestRunSchema_JSON_ADRRejectIf(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("adr", true, &buf); err != nil {
		t.Fatal(err)
	}
	var out SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.RejectIf) == 0 {
		t.Fatal("ADR JSON output should include reject_if bullets so LLMs in agent mode see the rejection contract")
	}
	if out.Workorder == "" {
		t.Error("ADR JSON output should include workorder prose")
	}
	if !strings.Contains(out.Workorder, "volatile Discovery Brief") {
		t.Fatalf("ADR JSON workorder should require a volatile Discovery Brief, got %q", out.Workorder)
	}
	if !strings.Contains(out.Workorder, "owner, governing material, stop condition") {
		t.Fatalf("ADR JSON workorder should name the brief fields, got %q", out.Workorder)
	}
}

func TestRunSchema_JSON_RefRejectIf(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("ref", true, &buf); err != nil {
		t.Fatal(err)
	}
	var out SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.RejectIf) == 0 {
		t.Fatal("ref JSON output should include reject_if bullets")
	}
}

func TestRunSchema_JSON_RuleRejectIf(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("rule", true, &buf); err != nil {
		t.Fatal(err)
	}
	var out SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.RejectIf) == 0 {
		t.Fatal("rule JSON output should include reject_if bullets")
	}
}

func TestRunSchema_JSON_ComponentNoRejectIf(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("component", true, &buf); err != nil {
		t.Fatal(err)
	}
	var out SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.RejectIf) != 0 {
		t.Errorf("component JSON output should not include reject_if (out of scope), got %v", out.RejectIf)
	}
}

func TestRunSchema_AgentTOONUsesCompactSchema(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	var buf bytes.Buffer
	if err := RunSchemaWithOptions(SchemaOptions{EntityType: "component", JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	requireAll(t, out,
		"type: component",
		"sections[9]",
		"kind: table",
		"min: rows>=1",
		"Reference:reference>uses(ref|rule)",
		"rules:",
		"empty cells use N.A - <reason>",
	)
	for _, noisy := range []string{"content_type:", "required:", "columns["} {
		if strings.Contains(out, noisy) {
			t.Fatalf("agent schema should use compact names, found %q:\n%s", noisy, out)
		}
	}
}

func TestRunSchema_JSON_ADRGuidanceFields(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSchema("adr", true, &buf); err != nil {
		t.Fatal(err)
	}

	var out SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	for _, s := range out.Sections {
		if s.Name != "Compliance Refs" {
			continue
		}
		if !strings.Contains(s.Fill, "why") {
			t.Fatalf("Compliance Refs fill guidance should explain what to write, got %q", s.Fill)
		}
		if !strings.Contains(strings.ToLower(s.Failure), "governing references") {
			t.Fatalf("Compliance Refs failure guidance should explain what goes wrong, got %q", s.Failure)
		}
		return
	}
	t.Fatal("Compliance Refs section not found in ADR schema")
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

	var out SchemaOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if out.Type != "adr" {
		t.Errorf("type = %q, want adr", out.Type)
	}
	for _, s := range out.Sections {
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

func TestRunSchema_ReadsProjectCanvasOverride(t *testing.T) {
	_, c3Dir := createDBFixtureWithC3Dir(t)
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "write", ID: "component", Body: strings.NewReader(projectComponentCanvasDoc())}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunSchemaWithOptions(SchemaOptions{EntityType: "component", C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"Schema: component",
		"Custom Project Section",
	)
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
			// Evidence is a semantic "evidence" column: validation (grounded
			// command/path/entity-id) keys off this type, not the column name.
			if evidenceCol.Type != "evidence" {
				t.Errorf("Evidence column type = %q, want %q", evidenceCol.Type, "evidence")
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
