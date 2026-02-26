package markdown

import (
	"testing"
)

// =============================================================================
// ParseSections: split markdown body into named sections by ## headers
// =============================================================================

func TestParseSections_BasicSplit(t *testing.T) {
	body := `
# Top Title

## Goal

Handle authentication.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | tokens | c3-102 |

## Code References

| File | Purpose |
|------|---------|
| src/auth.ts | Auth module |
`

	sections := ParseSections(body)

	if len(sections) < 3 {
		t.Fatalf("expected at least 3 sections, got %d", len(sections))
	}

	goal := findSection(sections, "Goal")
	if goal == nil {
		t.Fatal("expected Goal section")
	}
	if goal.Content != "Handle authentication." {
		t.Errorf("Goal content = %q, want %q", goal.Content, "Handle authentication.")
	}

	deps := findSection(sections, "Dependencies")
	if deps == nil {
		t.Fatal("expected Dependencies section")
	}
	if deps.Content == "" {
		t.Error("Dependencies section should have content")
	}
}

func TestParseSections_PreservesContentBeforeFirstHeading(t *testing.T) {
	body := `
# Top Title

Some intro text.

## Goal

The goal.
`
	sections := ParseSections(body)

	// Content before first ## should be captured as preamble (empty name)
	if len(sections) == 0 {
		t.Fatal("expected at least 1 section")
	}

	goal := findSection(sections, "Goal")
	if goal == nil {
		t.Fatal("expected Goal section")
	}
}

func TestParseSections_EmptySection(t *testing.T) {
	body := `
## Goal

## Dependencies

Some deps.
`
	sections := ParseSections(body)

	goal := findSection(sections, "Goal")
	if goal == nil {
		t.Fatal("expected Goal section")
	}
	if goal.Content != "" {
		t.Errorf("empty Goal section content = %q, want empty", goal.Content)
	}
}

func TestParseSections_IgnoresH3Subsections(t *testing.T) {
	body := `
## Goal

Main goal.

### Details

Some details under Goal.

## Dependencies

Deps here.
`
	sections := ParseSections(body)

	goal := findSection(sections, "Goal")
	if goal == nil {
		t.Fatal("expected Goal section")
	}
	// H3 content should stay inside Goal section
	if !containsStr(goal.Content, "Details") {
		t.Error("H3 content should be part of parent ## section")
	}
	if !containsStr(goal.Content, "Some details under Goal.") {
		t.Error("H3 body should be part of parent ## section")
	}
}

func TestParseSections_PreservesHTMLComments(t *testing.T) {
	body := `
## Goal

The goal.

<!-- Some comment to preserve -->

## Dependencies

Deps.
`
	sections := ParseSections(body)

	goal := findSection(sections, "Goal")
	if goal == nil {
		t.Fatal("expected Goal section")
	}
	if !containsStr(goal.Content, "<!-- Some comment to preserve -->") {
		t.Error("HTML comments should be preserved in section content")
	}
}

func TestParseSections_FencedCodeBlockWithHashHeaders(t *testing.T) {
	body := `
## Goal

The goal.

## Examples

Some code:

` + "```go" + `
// ## This is NOT a section header
func main() {
    fmt.Println("## also not a header")
}
` + "```" + `

## Dependencies

Deps.
`
	sections := ParseSections(body)

	examples := findSection(sections, "Examples")
	if examples == nil {
		t.Fatal("expected Examples section")
	}
	// The ## inside fenced code should NOT cause a section split
	if !containsStr(examples.Content, "This is NOT a section header") {
		t.Error("fenced code block content should stay inside Examples section")
	}

	deps := findSection(sections, "Dependencies")
	if deps == nil {
		t.Fatal("Dependencies section should exist after fenced code block")
	}
	if deps.Content != "Deps." {
		t.Errorf("Dependencies content = %q, want %q", deps.Content, "Deps.")
	}
}

func TestParseTable_EscapedPipes(t *testing.T) {
	tableStr := `| Pattern | Example |
|---------|---------|
| OR operator | a \| b |
| Simple | just text |`

	table, err := ParseTable(tableStr)
	if err != nil {
		t.Fatal(err)
	}

	if len(table.Rows) != 2 {
		t.Fatalf("row count = %d, want 2", len(table.Rows))
	}
	// Escaped pipe should be preserved in cell content
	if table.Rows[0]["Example"] != `a \| b` && table.Rows[0]["Example"] != "a | b" {
		t.Errorf("escaped pipe row = %q", table.Rows[0]["Example"])
	}
}

// =============================================================================
// ParseTable: parse a markdown table into structured rows
// =============================================================================

func TestParseTable_Basic(t *testing.T) {
	tableStr := `| Direction | What | From/To |
|-----------|------|---------|
| IN | auth tokens | c3-102 |
| OUT | sessions | c3-103 |`

	table, err := ParseTable(tableStr)
	if err != nil {
		t.Fatal(err)
	}

	if len(table.Headers) != 3 {
		t.Fatalf("headers count = %d, want 3", len(table.Headers))
	}
	if table.Headers[0] != "Direction" {
		t.Errorf("header[0] = %q, want %q", table.Headers[0], "Direction")
	}

	if len(table.Rows) != 2 {
		t.Fatalf("row count = %d, want 2", len(table.Rows))
	}
	if table.Rows[0]["Direction"] != "IN" {
		t.Errorf("row[0][Direction] = %q, want %q", table.Rows[0]["Direction"], "IN")
	}
	if table.Rows[0]["What"] != "auth tokens" {
		t.Errorf("row[0][What] = %q, want %q", table.Rows[0]["What"], "auth tokens")
	}
	if table.Rows[1]["From/To"] != "c3-103" {
		t.Errorf("row[1][From/To] = %q, want %q", table.Rows[1]["From/To"], "c3-103")
	}
}

func TestParseTable_EmptyCells(t *testing.T) {
	tableStr := `| ID | Name | Status |
|----|------|--------|
| c3-101 | auth | |
| c3-102 | | active |`

	table, err := ParseTable(tableStr)
	if err != nil {
		t.Fatal(err)
	}

	if table.Rows[0]["Status"] != "" {
		t.Errorf("empty cell should be empty string, got %q", table.Rows[0]["Status"])
	}
	if table.Rows[1]["Name"] != "" {
		t.Errorf("empty cell should be empty string, got %q", table.Rows[1]["Name"])
	}
}

func TestParseTable_ExtraWhitespace(t *testing.T) {
	tableStr := `|  Direction  |  What  |  From/To  |
|-------------|--------|-----------|
|  IN  |  tokens  |  c3-102  |`

	table, err := ParseTable(tableStr)
	if err != nil {
		t.Fatal(err)
	}

	if table.Rows[0]["Direction"] != "IN" {
		t.Errorf("should trim whitespace, got %q", table.Rows[0]["Direction"])
	}
	if table.Rows[0]["What"] != "tokens" {
		t.Errorf("should trim whitespace, got %q", table.Rows[0]["What"])
	}
}

func TestParseTable_HeaderOnly(t *testing.T) {
	tableStr := `| File | Purpose |
|------|---------|`

	table, err := ParseTable(tableStr)
	if err != nil {
		t.Fatal(err)
	}

	if len(table.Headers) != 2 {
		t.Fatalf("headers count = %d, want 2", len(table.Headers))
	}
	if len(table.Rows) != 0 {
		t.Errorf("empty table should have 0 rows, got %d", len(table.Rows))
	}
}

func TestParseTable_CommentRows(t *testing.T) {
	// Template tables often have HTML comment rows as hints
	tableStr := `| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
<!-- Category: foundation (01-09) | feature (10+) -->
| c3-101 | auth | foundation | active | Auth |`

	table, err := ParseTable(tableStr)
	if err != nil {
		t.Fatal(err)
	}

	// Comment rows should be skipped, only real data rows parsed
	if len(table.Rows) != 1 {
		t.Fatalf("should skip comment rows, got %d rows", len(table.Rows))
	}
	if table.Rows[0]["ID"] != "c3-101" {
		t.Errorf("data row ID = %q, want %q", table.Rows[0]["ID"], "c3-101")
	}
}

func TestParseTable_InvalidInput(t *testing.T) {
	_, err := ParseTable("not a table")
	if err == nil {
		t.Error("expected error for non-table input")
	}
}

func TestParseTable_ColumnCountMismatch(t *testing.T) {
	tableStr := `| A | B | C |
|---|---|---|
| 1 | 2 |`

	_, err := ParseTable(tableStr)
	if err == nil {
		t.Error("expected error for column count mismatch")
	}
}

// =============================================================================
// WriteTable: structured data → markdown table string
// =============================================================================

func TestWriteTable_Basic(t *testing.T) {
	table := &Table{
		Headers: []string{"File", "Purpose"},
		Rows: []map[string]string{
			{"File": "src/auth.ts", "Purpose": "Auth module"},
			{"File": "src/db.ts", "Purpose": "Database layer"},
		},
	}

	result := WriteTable(table)

	if !containsStr(result, "| File | Purpose |") {
		t.Errorf("should have header row, got:\n%s", result)
	}
	if !containsStr(result, "| src/auth.ts | Auth module |") {
		t.Errorf("should have data row, got:\n%s", result)
	}
	if !containsStr(result, "| src/db.ts | Database layer |") {
		t.Errorf("should have data row, got:\n%s", result)
	}
}

func TestWriteTable_EmptyRows(t *testing.T) {
	table := &Table{
		Headers: []string{"File", "Purpose"},
		Rows:    []map[string]string{},
	}

	result := WriteTable(table)

	if !containsStr(result, "| File | Purpose |") {
		t.Errorf("should have header row even with no data, got:\n%s", result)
	}
	// Should have header + separator, no data rows
	if containsStr(result, "| |") {
		t.Error("should not have empty data rows")
	}
}

func TestWriteTable_PreservesColumnOrder(t *testing.T) {
	table := &Table{
		Headers: []string{"Direction", "What", "From/To"},
		Rows: []map[string]string{
			{"Direction": "IN", "What": "tokens", "From/To": "c3-102"},
		},
	}

	result := WriteTable(table)

	// Parse it back and verify order matches
	parsed, err := ParseTable(result)
	if err != nil {
		t.Fatal(err)
	}
	for i, h := range table.Headers {
		if parsed.Headers[i] != h {
			t.Errorf("header[%d] = %q, want %q", i, parsed.Headers[i], h)
		}
	}
}

// =============================================================================
// Roundtrip: ParseTable → WriteTable → ParseTable
// =============================================================================

func TestTableRoundtrip(t *testing.T) {
	original := `| Direction | What | From/To |
|-----------|------|---------|
| IN | auth tokens | c3-102 |
| OUT | sessions | c3-103 |`

	table, err := ParseTable(original)
	if err != nil {
		t.Fatal(err)
	}

	written := WriteTable(table)
	reparsed, err := ParseTable(written)
	if err != nil {
		t.Fatal(err)
	}

	if len(reparsed.Rows) != len(table.Rows) {
		t.Fatalf("roundtrip row count: %d vs %d", len(reparsed.Rows), len(table.Rows))
	}
	for i, row := range table.Rows {
		for k, v := range row {
			if reparsed.Rows[i][k] != v {
				t.Errorf("roundtrip row[%d][%s] = %q, want %q", i, k, reparsed.Rows[i][k], v)
			}
		}
	}
}

// =============================================================================
// ReplaceSection: update a specific ## section in markdown
// =============================================================================

func TestReplaceSection_UpdateExisting(t *testing.T) {
	body := `
## Goal

Old goal text.

## Dependencies

Some deps.
`
	result, err := ReplaceSection(body, "Goal", "New goal text.")
	if err != nil {
		t.Fatal(err)
	}

	sections := ParseSections(result)
	goal := findSection(sections, "Goal")
	if goal == nil {
		t.Fatal("Goal section should exist after replace")
	}
	if goal.Content != "New goal text." {
		t.Errorf("Goal content = %q, want %q", goal.Content, "New goal text.")
	}

	// Dependencies should be untouched
	deps := findSection(sections, "Dependencies")
	if deps == nil {
		t.Fatal("Dependencies section should survive")
	}
	if deps.Content != "Some deps." {
		t.Errorf("Dependencies content = %q, want %q", deps.Content, "Some deps.")
	}
}

func TestReplaceSection_SectionNotFound(t *testing.T) {
	body := `
## Goal

The goal.
`
	_, err := ReplaceSection(body, "NonExistent", "content")
	if err == nil {
		t.Error("expected error when section doesn't exist")
	}
}

func TestReplaceSection_PreservesOrder(t *testing.T) {
	body := `
## First

Content A.

## Second

Content B.

## Third

Content C.
`
	result, err := ReplaceSection(body, "Second", "Updated B.")
	if err != nil {
		t.Fatal(err)
	}

	sections := ParseSections(result)
	if len(sections) < 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}

	// Verify ordering: First, Second, Third
	names := sectionNames(sections)
	if names[0] != "First" || names[1] != "Second" || names[2] != "Third" {
		t.Errorf("section order = %v, want [First Second Third]", names)
	}
}

func TestReplaceSection_WithTable(t *testing.T) {
	body := `
## Goal

The goal.

## Code References

| File | Purpose |
|------|---------|
| old.ts | Old file |
`
	newTable := `| File | Purpose |
|------|---------|
| new.ts | New file |
| other.ts | Other file |`

	result, err := ReplaceSection(body, "Code References", newTable)
	if err != nil {
		t.Fatal(err)
	}

	if !containsStr(result, "new.ts") {
		t.Error("should contain new table content")
	}
	if containsStr(result, "old.ts") {
		t.Error("should not contain old table content")
	}
	// Goal should survive
	if !containsStr(result, "The goal.") {
		t.Error("Goal section should survive")
	}
}

// =============================================================================
// SetTableInSection: update a table within a named section
// =============================================================================

func TestSetTableInSection_ReplaceEntireTable(t *testing.T) {
	body := `
## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | old dep | c3-99 |

## Code References

| File | Purpose |
|------|---------|
`
	newTable := &Table{
		Headers: []string{"Direction", "What", "From/To"},
		Rows: []map[string]string{
			{"Direction": "IN", "What": "auth tokens", "From/To": "c3-102"},
			{"Direction": "OUT", "What": "sessions", "From/To": "c3-103"},
		},
	}

	result, err := SetTableInSection(body, "Dependencies", newTable)
	if err != nil {
		t.Fatal(err)
	}

	if containsStr(result, "old dep") {
		t.Error("old table content should be replaced")
	}
	if !containsStr(result, "auth tokens") {
		t.Error("new table content should be present")
	}
	if !containsStr(result, "sessions") {
		t.Error("new table content should be present")
	}

	// Code References should survive
	sections := ParseSections(result)
	cr := findSection(sections, "Code References")
	if cr == nil {
		t.Error("Code References section should survive")
	}
}

func TestSetTableInSection_AppendRow(t *testing.T) {
	body := `
## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | auth tokens | c3-102 |
`
	newRow := map[string]string{
		"Direction": "OUT",
		"What":      "events",
		"From/To":   "c3-103",
	}

	result, err := AppendTableRow(body, "Dependencies", newRow)
	if err != nil {
		t.Fatal(err)
	}

	// Original row should still be there
	if !containsStr(result, "auth tokens") {
		t.Error("existing row should be preserved")
	}
	// New row should be added
	if !containsStr(result, "events") {
		t.Error("new row should be appended")
	}
	if !containsStr(result, "c3-103") {
		t.Error("new row data should be present")
	}
}

func TestSetTableInSection_AppendToEmptyTable(t *testing.T) {
	body := `
## Code References

| File | Purpose |
|------|---------|
`
	newRow := map[string]string{
		"File":    "src/auth.ts",
		"Purpose": "Auth module",
	}

	result, err := AppendTableRow(body, "Code References", newRow)
	if err != nil {
		t.Fatal(err)
	}

	if !containsStr(result, "src/auth.ts") {
		t.Error("new row should be appended to empty table")
	}
}

// =============================================================================
// ExtractTableFromSection: get the parsed table from a named section
// =============================================================================

func TestExtractTableFromSection_Found(t *testing.T) {
	body := `
## Goal

The goal.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN | tokens | c3-102 |

## Other

More text.
`
	table, err := ExtractTableFromSection(body, "Dependencies")
	if err != nil {
		t.Fatal(err)
	}
	if table == nil {
		t.Fatal("expected table in Dependencies section")
	}
	if len(table.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(table.Rows))
	}
	if table.Rows[0]["Direction"] != "IN" {
		t.Errorf("row[0][Direction] = %q, want %q", table.Rows[0]["Direction"], "IN")
	}
}

func TestExtractTableFromSection_NoTable(t *testing.T) {
	body := `
## Goal

Just text, no table.
`
	table, err := ExtractTableFromSection(body, "Goal")
	if err != nil {
		t.Fatal(err)
	}
	if table != nil {
		t.Error("expected nil table for text-only section")
	}
}

func TestExtractTableFromSection_SectionNotFound(t *testing.T) {
	body := `
## Goal

The goal.
`
	_, err := ExtractTableFromSection(body, "NonExistent")
	if err == nil {
		t.Error("expected error for missing section")
	}
}

// =============================================================================
// helpers
// =============================================================================

func findSection(sections []Section, name string) *Section {
	for i := range sections {
		if sections[i].Name == name {
			return &sections[i]
		}
	}
	return nil
}

func sectionNames(sections []Section) []string {
	var names []string
	for _, s := range sections {
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}
	return names
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
