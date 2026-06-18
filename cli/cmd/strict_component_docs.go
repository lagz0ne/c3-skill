package cmd

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

var (
	// Placeholder = a non-answer left in a strict section. Match clear markers only:
	// bare natural words (later, optional, maybe) and the domain word "TODO" (a TODO
	// app, a TODO list) are NOT placeholders — only "TODO:" as an explicit marker is.
	placeholderPattern = regexp.MustCompile(`(?i)\bTBD\b|\bFIXME\b|\bTODO:|\bif applicable\b|\bsee above\b|\bas needed\b`)
	entityRefPattern   = regexp.MustCompile(`\b(c3-[0-9]+|ref-[a-z0-9-]+|rule-[a-z0-9-]+|adr-[0-9]{8}-[a-z0-9-]+|recipe-[a-z0-9-]+)\b`)
	evidencePattern    = regexp.MustCompile(`(?i)(\b(c3x|go test|bunx|npm|pnpm|yarn|cargo|pytest|make|bash)\b|[./][A-Za-z0-9_./*-]+|\b[A-Za-z0-9_-]+\.(go|md|ts|tsx|js|jsx|py|rs|yaml|yml|json)\b|\b(c3-[0-9]+|ref-[a-z0-9-]+|rule-[a-z0-9-]+|adr-[0-9]{8}-[a-z0-9-]+|recipe-[a-z0-9-]+)\b)`)
)

// strictRules is the structural validation contract for an entity type, derived
// entirely from its canvas/schema definition — the single source of shape. No
// per-type hardcoded maps: whatever a definition declares (required sections in
// order, table columns, min rows, enum values) is what gets enforced.
type strictRules struct {
	sectionOrder []string                       // required sections, in definition order
	tableHeaders map[string][]string            // table section -> column names
	minRows      map[string]int                 // table section -> required rows
	columnEnums  map[string]map[string][]string // section -> column -> allowed values (sans N.A sentinel)
	columnTypes  map[string]map[string]string   // section -> column -> declared type (reference/evidence/...)
}

// deriveStrictRules projects a definition's []SectionDef into the structural
// rules the strict validator enforces. This is the seam that makes strict
// validation generic over any canvas definition rather than bound to component.
func deriveStrictRules(defs []schema.SectionDef) strictRules {
	rules := strictRules{
		tableHeaders: map[string][]string{},
		minRows:      map[string]int{},
		columnEnums:  map[string]map[string][]string{},
		columnTypes:  map[string]map[string]string{},
	}
	for _, section := range defs {
		if !section.Required {
			continue
		}
		// FREE sections are narrative: excluded from canvas-shape, MinWords,
		// typed-column, and discharge checks. Only STRICT sections are enforced
		// by the strict validator.
		if section.Free {
			continue
		}
		rules.sectionOrder = append(rules.sectionOrder, section.Name)
		if section.ContentType != "table" {
			continue
		}
		headers := make([]string, 0, len(section.Columns))
		for _, col := range section.Columns {
			headers = append(headers, col.Name)
			if rules.columnTypes[section.Name] == nil {
				rules.columnTypes[section.Name] = map[string]string{}
			}
			rules.columnTypes[section.Name][col.Name] = col.Type
			if col.Type != "enum" {
				continue
			}
			var allowed []string
			for _, v := range col.Values {
				// The N.A escape is handled generically by strictEnumAllowed;
				// keep it out of the declared set so messages stay precise.
				if strings.HasPrefix(v, "N.A") {
					continue
				}
				allowed = append(allowed, v)
			}
			if len(allowed) > 0 {
				if rules.columnEnums[section.Name] == nil {
					rules.columnEnums[section.Name] = map[string][]string{}
				}
				rules.columnEnums[section.Name][col.Name] = allowed
			}
		}
		rules.tableHeaders[section.Name] = headers
		rules.minRows[section.Name] = section.MinRows
	}
	return rules
}

// validateStrictComponentDoc validates a component body against the component
// definition. The component-specific entry point is preserved; the engine
// underneath (validateStrictDoc) is generic over any definition.
func validateStrictComponentDoc(body string, severity string) []Issue {
	return validateStrictDoc(schema.ForType("component"), body, severity)
}

func validateStrictDoc(defs []schema.SectionDef, body string, severity string) []Issue {
	rules := deriveStrictRules(defs)
	hint := strictHintFor(defs)
	// minWords per declared text section — drives the thin-text check from the
	// canvas (SectionDef.MinWords), not hardcoded literals.
	minWords := map[string]int{}
	for _, def := range defs {
		if def.ContentType == "text" {
			minWords[def.Name] = def.MinWords
		}
	}
	sections := markdown.ParseSections(body)
	sectionMap := make(map[string]markdown.Section)
	seen := make(map[string]int)
	var orderedNames []string
	for _, section := range sections {
		if section.Name == "" {
			continue
		}
		seen[section.Name]++
		sectionMap[section.Name] = section
		orderedNames = append(orderedNames, section.Name)
	}

	var issues []Issue
	for _, name := range rules.sectionOrder {
		if seen[name] > 1 {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("duplicate required section: %s", name), hint))
		}
		section, exists := sectionMap[name]
		if !exists {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("missing required section: %s", name), hint))
			continue
		}
		if strings.TrimSpace(section.Content) == "" {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("empty required section: %s", name), hint))
			continue
		}
	}
	issues = append(issues, validateRequiredSectionOrder(orderedNames, rules.sectionOrder, hint, severity)...)

	// Text-section thin checks run over every declared STRICT text section, keyed
	// by the canvas-declared MinWords. FREE sections are excluded by
	// deriveStrictRules (they are not in sectionOrder).
	for _, name := range rules.sectionOrder {
		section, ok := sectionMap[name]
		if !ok {
			continue
		}
		if _, isText := minWords[name]; !isText {
			continue
		}
		issues = append(issues, validateTextSubstance(name, section.Content, minWords[name], severity)...)
	}

	for sectionName, headers := range rules.tableHeaders {
		section, ok := sectionMap[sectionName]
		if !ok {
			continue
		}
		table, err := markdown.ParseTable(section.Content)
		if err != nil {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid required table: %s", sectionName), hint))
			continue
		}
		if !slices.Equal(table.Headers, headers) {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("wrong table columns in %s: expected %s", sectionName, strings.Join(headers, ", ")), hint))
		}
		if len(table.Rows) < rules.minRows[sectionName] {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("not enough rows in %s: need at least %d", sectionName, rules.minRows[sectionName]), hint))
		}
		issues = append(issues, validateStrictTableCells(sectionName, table, rules.columnEnums, hint, severity)...)
		issues = append(issues, validateStrictTableSemantics(sectionName, table, rules.columnTypes[sectionName], rules.sectionOrder, hint, severity)...)
	}

	return issues
}

func validateRequiredSectionOrder(orderedNames []string, sectionOrder []string, hint, severity string) []Issue {
	var issues []Issue
	lastIndex := -1
	lastName := ""
	for _, required := range sectionOrder {
		index := slices.Index(orderedNames, required)
		if index < 0 {
			continue
		}
		if index < lastIndex {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("sections out of order: expected %s before %s", lastName, required), hint))
			continue
		}
		lastIndex = index
		lastName = required
	}
	return issues
}

// strictHintFor derives the reviewer-ready hint from the actual STRICT section
// set of a definition, so a prd/atomic doc never sees the component section
// literal.
func strictHintFor(defs []schema.SectionDef) string {
	var names []string
	for _, def := range defs {
		if !def.Required || def.Free {
			continue
		}
		names = append(names, def.Name)
	}
	if len(names) == 0 {
		return "write reviewer-ready docs that match the canvas shape"
	}
	return "write reviewer-ready docs: " + strings.Join(names, ", ")
}

func strictIssue(severity, message, hint string) Issue {
	return Issue{
		Severity: severity,
		Message:  message,
		Hint:     hint,
	}
}

func validateTextSubstance(sectionName, text string, minWords int, severity string) []Issue {
	var issues []Issue
	trimmed := strings.TrimSpace(text)
	hint := "write reviewer-ready prose in " + sectionName
	if placeholderPattern.MatchString(trimmed) {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("placeholder language in %s", sectionName), hint))
	}
	if strings.Contains(trimmed, "N.A") && !strings.Contains(trimmed, "N.A - ") {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid N.A in %s: use N.A - <reason>", sectionName), hint))
	}
	if minWords > 0 && len(strings.Fields(trimmed)) < minWords {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("%s is too thin: write at least %d words", sectionName, minWords), hint))
	}
	return issues
}

func validateStrictTableCells(sectionName string, table *markdown.Table, columnEnums map[string]map[string][]string, hint, severity string) []Issue {
	var issues []Issue
	for rowIndex, row := range table.Rows {
		naCells := 0
		for _, header := range table.Headers {
			value := strings.TrimSpace(row[header])
			cell := fmt.Sprintf("%s row %d column %s", sectionName, rowIndex+1, header)
			if value == "" {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("blank field in %s", cell), hint))
				continue
			}
			if isNAReason(value) {
				naCells++
			}
			if placeholderPattern.MatchString(value) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("placeholder language in %s", cell), hint))
			}
			if strings.Contains(value, "N.A") && !strings.HasPrefix(value, "N.A - ") {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid N.A in %s: use N.A - <reason>", cell), hint))
			}
			if allowed, ok := columnEnums[sectionName][header]; ok && !strictEnumAllowed(value, allowed) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid enum value in %s: expected one of %s or N.A - <reason>", cell, strings.Join(allowed, ", ")), hint))
			}
		}
		if naCells == len(table.Headers) {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("row cannot be entirely N.A in %s row %d", sectionName, rowIndex+1), hint))
		}
	}
	return issues
}

// validateStrictTableSemantics enforces cell grounding by COLUMN TYPE, read from
// the definition (columnTypes) — not by hardcoded section/column names. Any
// definition that types a column "reference" or "evidence" gets the same
// grounding rules, so the semantics are generic over any canvas definition.
func validateStrictTableSemantics(sectionName string, table *markdown.Table, columnTypes map[string]string, sectionOrder []string, hint, severity string) []Issue {
	var issues []Issue
	repeated := map[string]string{}
	groundedReferences := 0
	hasReferenceColumn := false
	for _, t := range columnTypes {
		if t == "reference" {
			hasReferenceColumn = true
			break
		}
	}

	for rowIndex, row := range table.Rows {
		for _, header := range table.Headers {
			value := strings.TrimSpace(row[header])
			cell := fmt.Sprintf("%s row %d column %s", sectionName, rowIndex+1, header)

			if columnTypes[header] == "reference" {
				if !isGroundedReference(value) {
					issues = append(issues, strictIssue(severity, fmt.Sprintf("ungrounded reference in %s: use an entity id or N.A - <reason>", cell), hint))
				} else if !isNAReason(value) {
					groundedReferences++
				}
			}
			if columnTypes[header] == "evidence" && !isGroundedEvidence(value) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("ungrounded evidence in %s: name a command, file path, or entity id", cell), hint))
			}
			if sectionName == "Derived Materials" && header == "Must derive from" && !mentionsComponentSection(value, sectionOrder) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("ungrounded derivation in %s: cite strict component sections", cell), hint))
			}
			if isSemanticDetailColumn(header) && !isNAReason(value) {
				normalized := normalizeSemanticText(value)
				if previous, ok := repeated[normalized]; ok {
					issues = append(issues, strictIssue(severity, fmt.Sprintf("repeated boilerplate in %s: duplicates %s", cell, previous), hint))
				} else if len(strings.Fields(normalized)) >= 4 {
					repeated[normalized] = cell
				}
			}
		}
	}

	if hasReferenceColumn && len(table.Rows) > 0 && groundedReferences == 0 {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("%s needs at least one grounded reference", sectionName), hint))
	}
	return issues
}

func strictEnumAllowed(value string, allowed []string) bool {
	if isNAReason(value) {
		return true
	}
	return slices.Contains(allowed, value)
}

func isNAReason(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), "N.A - ")
}

func isGroundedReference(value string) bool {
	value = strings.TrimSpace(value)
	return isNAReason(value) || entityRefPattern.MatchString(value)
}

func isGroundedEvidence(value string) bool {
	value = strings.TrimSpace(value)
	return isNAReason(value) || evidencePattern.MatchString(value)
}

func mentionsComponentSection(value string, sectionOrder []string) bool {
	for _, section := range sectionOrder {
		if strings.Contains(value, section) {
			return true
		}
	}
	return false
}

func isSemanticDetailColumn(header string) bool {
	switch header {
	case "Detail", "Governs", "Soft Cap", "Current Load", "Escalation", "Contract", "Detection", "Required Verification", "Evidence":
		return true
	default:
		return false
	}
}

func normalizeSemanticText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
