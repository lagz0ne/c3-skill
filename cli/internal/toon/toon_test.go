package toon

import (
	"testing"
)

func TestMarshalTable_BasicEntities(t *testing.T) {
	type entity struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Title string `json:"title"`
	}
	items := []entity{
		{ID: "c3-0", Type: "system", Title: "TestProject"},
		{ID: "c3-1", Type: "container", Title: "api"},
		{ID: "c3-101", Type: "component", Title: "auth"},
	}

	out, err := MarshalTable("entities", items, []string{"id", "type", "title"})
	if err != nil {
		t.Fatal(err)
	}

	// Header with count and field names
	if want := "entities[3]{id,type,title}:\n"; !contains(out, want) {
		t.Errorf("header mismatch\nwant: %q\ngot:\n%s", want, out)
	}
	// Rows indented with 2 spaces
	if want := "  c3-0,system,TestProject\n"; !contains(out, want) {
		t.Errorf("row mismatch\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "  c3-101,component,auth\n"; !contains(out, want) {
		t.Errorf("row mismatch\nwant: %q\ngot:\n%s", want, out)
	}
}

func TestMarshalTable_EmptySlice(t *testing.T) {
	type entity struct {
		ID string `json:"id"`
	}
	out, err := MarshalTable("items", []entity{}, []string{"id"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "items[0]{id}:\n"; out != want {
		t.Errorf("empty table\nwant: %q\ngot: %q", want, out)
	}
}

func TestMarshalTable_QuotesSpecialChars(t *testing.T) {
	type item struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	items := []item{
		{ID: "c3-1", Title: "API, Gateway"},       // contains comma
		{ID: "c3-2", Title: "Web: Frontend"},       // contains colon
		{ID: "c3-3", Title: `Has "quotes"`},        // contains quotes
		{ID: "c3-4", Title: "Normal title"},         // no quoting needed
	}
	out, err := MarshalTable("items", items, []string{"id", "title"})
	if err != nil {
		t.Fatal(err)
	}
	if want := `  c3-1,"API, Gateway"` + "\n"; !contains(out, want) {
		t.Errorf("comma not quoted\nwant: %q\ngot:\n%s", want, out)
	}
	if want := `  c3-2,"Web: Frontend"` + "\n"; !contains(out, want) {
		t.Errorf("colon not quoted\nwant: %q\ngot:\n%s", want, out)
	}
	if want := `  c3-3,"Has \"quotes\""` + "\n"; !contains(out, want) {
		t.Errorf("quotes not escaped\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "  c3-4,Normal title\n"; !contains(out, want) {
		t.Errorf("normal not quoted\nwant: %q\ngot:\n%s", want, out)
	}
}

func TestMarshalTable_OmitsEmptyOptionalFields(t *testing.T) {
	type item struct {
		ID     string `json:"id"`
		Parent string `json:"parent,omitempty"`
		Status string `json:"status"`
	}
	items := []item{
		{ID: "c3-0", Parent: "", Status: "active"},
		{ID: "c3-1", Parent: "c3-0", Status: "active"},
	}
	out, err := MarshalTable("items", items, []string{"id", "parent", "status"})
	if err != nil {
		t.Fatal(err)
	}
	// Empty parent should render as empty field between commas
	if want := "  c3-0,,active\n"; !contains(out, want) {
		t.Errorf("empty field mismatch\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "  c3-1,c3-0,active\n"; !contains(out, want) {
		t.Errorf("filled field mismatch\nwant: %q\ngot:\n%s", want, out)
	}
}

func TestMarshalObject_FlatStruct(t *testing.T) {
	type status struct {
		Project    string `json:"project"`
		TotalCount int    `json:"totalCount"`
		Warnings   int    `json:"warnings"`
	}
	v := status{Project: "MyProject", TotalCount: 14, Warnings: 3}
	out, err := MarshalObject(v)
	if err != nil {
		t.Fatal(err)
	}
	if want := "project: MyProject\n"; !contains(out, want) {
		t.Errorf("string field\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "totalCount: 14\n"; !contains(out, want) {
		t.Errorf("int field\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "warnings: 3\n"; !contains(out, want) {
		t.Errorf("int field\nwant: %q\ngot:\n%s", want, out)
	}
}

func TestMarshalObject_NestedMap(t *testing.T) {
	type status struct {
		Project  string         `json:"project"`
		Entities map[string]int `json:"entities"`
	}
	v := status{
		Project:  "Test",
		Entities: map[string]int{"container": 2, "component": 5},
	}
	out, err := MarshalObject(v)
	if err != nil {
		t.Fatal(err)
	}
	if want := "project: Test\n"; !contains(out, want) {
		t.Errorf("project\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "entities:\n"; !contains(out, want) {
		t.Errorf("entities header\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "  component: 5\n"; !contains(out, want) {
		t.Errorf("nested map\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "  container: 2\n"; !contains(out, want) {
		t.Errorf("nested map\nwant: %q\ngot:\n%s", want, out)
	}
}

func TestMarshalObject_OmitsZeroValues(t *testing.T) {
	type status struct {
		Project  string   `json:"project"`
		Coverage *float64 `json:"coverage_pct,omitempty"`
		Warnings int      `json:"warnings"`
	}
	v := status{Project: "Test", Coverage: nil, Warnings: 0}
	out, err := MarshalObject(v)
	if err != nil {
		t.Fatal(err)
	}
	if contains(out, "coverage_pct") {
		t.Errorf("nil pointer should be omitted\ngot:\n%s", out)
	}
	// Zero int with no omitempty should still appear
	if want := "warnings: 0\n"; !contains(out, want) {
		t.Errorf("zero int without omitempty\nwant: %q\ngot:\n%s", want, out)
	}
}

func TestMarshalValue_Primitives(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{"hello", "hello"},
		{"", `""`},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{nil, "null"},
	}
	for _, tt := range tests {
		got := MarshalValue(tt.input)
		if got != tt.want {
			t.Errorf("MarshalValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", false},
		{"Hello World", false},
		{"has,comma", true},
		{"has:colon", true},
		{`has"quote`, true},
		{"has\nnewline", true},
		{"true", true},
		{"false", true},
		{"null", true},
		{"42", true},
		{"-3.14", true},
		{"", true},
		{" leading", true},
		{"trailing ", true},
	}
	for _, tt := range tests {
		got := NeedsQuoting(tt.input)
		if got != tt.want {
			t.Errorf("NeedsQuoting(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestMarshalTable_NumberFields(t *testing.T) {
	type item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	items := []item{
		{Name: "alpha", Count: 10},
		{Name: "beta", Count: 0},
	}
	out, err := MarshalTable("items", items, []string{"name", "count"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "  alpha,10\n"; !contains(out, want) {
		t.Errorf("number field\nwant: %q\ngot:\n%s", want, out)
	}
	if want := "  beta,0\n"; !contains(out, want) {
		t.Errorf("zero field\nwant: %q\ngot:\n%s", want, out)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
