package markdown

import (
	"fmt"
	"strings"
)

// Section represents a named section parsed from ## headers in markdown.
type Section struct {
	Name    string
	Content string
}

// Table represents a parsed markdown table.
type Table struct {
	Headers []string
	Rows    []map[string]string
}

// ParseSections splits a markdown body into sections by ## headers.
// Content before the first ## is captured as a preamble section with empty Name.
// Only ## (h2) headers create new sections; ### and deeper stay inside their parent.
// ## inside fenced code blocks are ignored.
func ParseSections(body string) []Section {
	lines := strings.Split(body, "\n")
	var sections []Section
	var currentName string
	var currentLines []string
	inFence := false
	started := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
		}

		if !inFence && strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			// Save previous section
			if started {
				sections = append(sections, Section{
					Name:    currentName,
					Content: trimContent(currentLines),
				})
			}
			currentName = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			currentLines = nil
			started = true
		} else {
			if !started {
				// Preamble content before first ##
				currentLines = append(currentLines, line)
				started = true
				currentName = ""
			} else {
				currentLines = append(currentLines, line)
			}
		}
	}

	// Save last section
	if started {
		sections = append(sections, Section{
			Name:    currentName,
			Content: trimContent(currentLines),
		})
	}

	return sections
}

// trimContent joins lines and trims leading/trailing whitespace.
func trimContent(lines []string) string {
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// ParseTable parses a markdown table string into a Table struct.
func ParseTable(markdown string) (*Table, error) {
	lines := strings.Split(strings.TrimSpace(markdown), "\n")

	// Filter out comment rows and empty lines, find header/separator/data
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		filtered = append(filtered, line)
	}

	if len(filtered) < 2 {
		return nil, fmt.Errorf("not a valid markdown table: need at least header and separator rows")
	}

	// Parse header row
	headers := parseCells(filtered[0])
	if len(headers) == 0 {
		return nil, fmt.Errorf("not a valid markdown table: no headers found")
	}

	// Verify separator row
	sepCells := parseCells(filtered[1])
	isSeparator := true
	for _, c := range sepCells {
		stripped := strings.Trim(c, "- :")
		if stripped != "" {
			isSeparator = false
			break
		}
	}
	if !isSeparator {
		return nil, fmt.Errorf("not a valid markdown table: second row is not a separator")
	}

	// Parse data rows
	var rows []map[string]string
	for _, line := range filtered[2:] {
		cells := parseCells(line)
		if len(cells) != len(headers) {
			return nil, fmt.Errorf("column count mismatch: header has %d columns, row has %d", len(headers), len(cells))
		}
		row := make(map[string]string, len(headers))
		for i, h := range headers {
			row[h] = cells[i]
		}
		rows = append(rows, row)
	}

	if rows == nil {
		rows = []map[string]string{}
	}

	return &Table{Headers: headers, Rows: rows}, nil
}

// parseCells splits a markdown table row into cell values, handling escaped pipes.
func parseCells(line string) []string {
	line = strings.TrimSpace(line)

	// Strip leading and trailing pipe
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}

	// Split by unescaped pipes
	var cells []string
	var current strings.Builder
	for i := 0; i < len(line); i++ {
		if line[i] == '\\' && i+1 < len(line) && line[i+1] == '|' {
			current.WriteString(`\|`)
			i++ // skip the pipe
		} else if line[i] == '|' {
			cells = append(cells, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(line[i])
		}
	}
	cells = append(cells, strings.TrimSpace(current.String()))

	return cells
}

// WriteTable converts a Table struct back to a markdown table string.
func WriteTable(t *Table) string {
	var sb strings.Builder

	// Header row
	sb.WriteString("| ")
	sb.WriteString(strings.Join(t.Headers, " | "))
	sb.WriteString(" |")
	sb.WriteString("\n")

	// Separator row
	sb.WriteString("|")
	for range t.Headers {
		sb.WriteString("------|")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range t.Rows {
		sb.WriteString("| ")
		vals := make([]string, len(t.Headers))
		for i, h := range t.Headers {
			vals[i] = row[h]
		}
		sb.WriteString(strings.Join(vals, " | "))
		sb.WriteString(" |")
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// ReplaceSection replaces the content of a named ## section in the markdown body.
func ReplaceSection(body string, name string, newContent string) (string, error) {
	sections := ParseSections(body)

	found := false
	for i := range sections {
		if sections[i].Name == name {
			sections[i].Content = newContent
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("section %q not found", name)
	}

	return reassemble(body, sections), nil
}

// reassemble reconstructs a markdown body from sections, preserving
// the leading whitespace style of the original.
func reassemble(originalBody string, sections []Section) string {
	var sb strings.Builder

	// Check if original body starts with a newline
	prefix := ""
	if strings.HasPrefix(originalBody, "\n") {
		prefix = "\n"
	}
	sb.WriteString(prefix)

	for i, s := range sections {
		if s.Name == "" {
			// Preamble
			if s.Content != "" {
				sb.WriteString(s.Content)
				sb.WriteString("\n")
			}
		} else {
			sb.WriteString("## ")
			sb.WriteString(s.Name)
			sb.WriteString("\n")
			if s.Content != "" {
				sb.WriteString("\n")
				sb.WriteString(s.Content)
				sb.WriteString("\n")
			} else {
				sb.WriteString("\n")
			}
		}
		// Add separator between sections (but not after the last)
		if i < len(sections)-1 && !(s.Name == "" && s.Content == "") {
			// Only add blank line if next section is named (not joining preamble with section)
			if sections[i+1].Name != "" || sections[i+1].Content != "" {
				// empty line not needed if we just wrote one
			}
		}
	}

	result := sb.String()
	// Ensure trailing newline
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// SetTableInSection replaces the table in a named section with a new table.
func SetTableInSection(body string, sectionName string, table *Table) (string, error) {
	newContent := WriteTable(table)
	return ReplaceSection(body, sectionName, newContent)
}

// AppendTableRow appends a row to the table in a named section.
func AppendTableRow(body string, sectionName string, row map[string]string) (string, error) {
	table, err := ExtractTableFromSection(body, sectionName)
	if err != nil {
		return "", err
	}
	if table == nil {
		return "", fmt.Errorf("no table found in section %q", sectionName)
	}

	table.Rows = append(table.Rows, row)
	return SetTableInSection(body, sectionName, table)
}

// ExtractTableFromSection extracts and parses the table from a named section.
func ExtractTableFromSection(body string, sectionName string) (*Table, error) {
	sections := ParseSections(body)

	var section *Section
	for i := range sections {
		if sections[i].Name == sectionName {
			section = &sections[i]
			break
		}
	}

	if section == nil {
		return nil, fmt.Errorf("section %q not found", sectionName)
	}

	// Find table lines in section content
	lines := strings.Split(section.Content, "\n")
	var tableLines []string
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, "<!--") {
			inTable = true
			tableLines = append(tableLines, line)
		} else if inTable && trimmed == "" {
			// End of table
			break
		} else if inTable {
			break
		}
	}

	if len(tableLines) == 0 {
		return nil, nil
	}

	return ParseTable(strings.Join(tableLines, "\n"))
}
