package cmd

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
)

var (
	componentSectionOrder = []string{
		"Goal",
		"Parent Fit",
		"Purpose",
		"Foundational Flow",
		"Business Flow",
		"Governance",
		"Contract",
		"Change Safety",
		"Derived Materials",
	}
	strictTableHeaders = map[string][]string{
		"Parent Fit":        {"Field", "Value"},
		"Foundational Flow": {"Aspect", "Detail", "Reference"},
		"Business Flow":     {"Aspect", "Detail", "Reference"},
		"Governance":        {"Reference", "Type", "Governs", "Precedence", "Notes"},
		"Contract":          {"Surface", "Direction", "Contract", "Boundary", "Evidence"},
		"Change Safety":     {"Risk", "Trigger", "Detection", "Required Verification"},
		"Derived Materials": {"Material", "Must derive from", "Allowed variance", "Evidence"},
	}
	strictMinRows = map[string]int{
		"Parent Fit":        4,
		"Foundational Flow": 4,
		"Business Flow":     4,
		"Governance":        1,
		"Contract":          2,
		"Change Safety":     2,
		"Derived Materials": 1,
	}
	strictColumnEnums = map[string]map[string][]string{
		"Governance": {
			"Type": {"ref", "rule", "adr", "spec", "policy", "example"},
		},
		"Contract": {
			"Direction": {"IN", "OUT", "IN/OUT"},
		},
	}
	placeholderPattern = regexp.MustCompile(`(?i)\b(TBD|TODO|maybe|optional|later|if applicable)\b`)
	entityRefPattern   = regexp.MustCompile(`\b(c3-[0-9]+|ref-[a-z0-9-]+|rule-[a-z0-9-]+|adr-[0-9]{8}-[a-z0-9-]+|recipe-[a-z0-9-]+)\b`)
	evidencePattern    = regexp.MustCompile(`(?i)(\b(c3x|go test|bunx|npm|pnpm|yarn|cargo|pytest|make|bash)\b|[./][A-Za-z0-9_./*-]+|\b[A-Za-z0-9_-]+\.(go|md|ts|tsx|js|jsx|py|rs|yaml|yml|json)\b|\b(c3-[0-9]+|ref-[a-z0-9-]+|rule-[a-z0-9-]+|adr-[0-9]{8}-[a-z0-9-]+|recipe-[a-z0-9-]+)\b)`)
)

func validateStrictComponentDoc(body string, severity string) []Issue {
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
	for i, name := range componentSectionOrder {
		if seen[name] > 1 {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("duplicate required section: %s", name)))
		}
		section, exists := sectionMap[name]
		if !exists {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("missing required section: %s", name)))
			continue
		}
		if strings.TrimSpace(section.Content) == "" {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("empty required section: %s", name)))
			continue
		}
		if i < len(orderedNames) && orderedNames[i] != name {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("component sections out of order: expected %s before %s", name, orderedNames[i])))
		}
	}

	if goal, ok := sectionMap["Goal"]; ok {
		issues = append(issues, validateTextSubstance("Goal", goal.Content, severity)...)
	}
	if purpose, ok := sectionMap["Purpose"]; ok {
		issues = append(issues, validateTextSubstance("Purpose", purpose.Content, severity)...)
	}

	for sectionName, headers := range strictTableHeaders {
		section, ok := sectionMap[sectionName]
		if !ok {
			continue
		}
		table, err := markdown.ParseTable(section.Content)
		if err != nil {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid required table: %s", sectionName)))
			continue
		}
		if !slices.Equal(table.Headers, headers) {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("wrong table columns in %s: expected %s", sectionName, strings.Join(headers, ", "))))
		}
		if len(table.Rows) < strictMinRows[sectionName] {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("not enough rows in %s: need at least %d", sectionName, strictMinRows[sectionName])))
		}
		issues = append(issues, validateStrictTableCells(sectionName, table, severity)...)
		issues = append(issues, validateStrictTableSemantics(sectionName, table, severity)...)
	}

	return issues
}

func strictIssue(severity, message string) Issue {
	return Issue{
		Severity: severity,
		Message:  message,
		Hint:     "write reviewer-ready component docs: Parent Fit, Purpose, Foundational Flow, Business Flow, Governance, Contract, Change Safety, Derived Materials",
	}
}

func validateTextSubstance(sectionName, text, severity string) []Issue {
	var issues []Issue
	trimmed := strings.TrimSpace(text)
	if placeholderPattern.MatchString(trimmed) {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("placeholder language in %s", sectionName)))
	}
	if strings.Contains(trimmed, "N.A") && !strings.Contains(trimmed, "N.A - ") {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid N.A in %s: use N.A - <reason>", sectionName)))
	}
	if sectionName == "Goal" && len(strings.Fields(trimmed)) < 4 {
		issues = append(issues, strictIssue(severity, "Goal is too thin: write one clear sentence"))
	}
	if sectionName == "Purpose" && len(strings.Fields(trimmed)) < 12 {
		issues = append(issues, strictIssue(severity, "Purpose is too thin: explain ownership and non-goals"))
	}
	return issues
}

func validateStrictTableCells(sectionName string, table *markdown.Table, severity string) []Issue {
	var issues []Issue
	for rowIndex, row := range table.Rows {
		naCells := 0
		for _, header := range table.Headers {
			value := strings.TrimSpace(row[header])
			cell := fmt.Sprintf("%s row %d column %s", sectionName, rowIndex+1, header)
			if value == "" {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("blank field in %s", cell)))
				continue
			}
			if isNAReason(value) {
				naCells++
			}
			if placeholderPattern.MatchString(value) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("placeholder language in %s", cell)))
			}
			if strings.Contains(value, "N.A") && !strings.HasPrefix(value, "N.A - ") {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid N.A in %s: use N.A - <reason>", cell)))
			}
			if allowed, ok := strictColumnEnums[sectionName][header]; ok && !strictEnumAllowed(value, allowed) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("invalid enum value in %s: expected one of %s or N.A - <reason>", cell, strings.Join(allowed, ", "))))
			}
		}
		if naCells == len(table.Headers) {
			issues = append(issues, strictIssue(severity, fmt.Sprintf("row cannot be entirely N.A in %s row %d", sectionName, rowIndex+1)))
		}
	}
	return issues
}

func validateStrictTableSemantics(sectionName string, table *markdown.Table, severity string) []Issue {
	var issues []Issue
	repeated := map[string]string{}
	groundedReferences := 0

	for rowIndex, row := range table.Rows {
		for _, header := range table.Headers {
			value := strings.TrimSpace(row[header])
			cell := fmt.Sprintf("%s row %d column %s", sectionName, rowIndex+1, header)

			if isReferenceColumn(sectionName, header) {
				if !isGroundedReference(value) {
					issues = append(issues, strictIssue(severity, fmt.Sprintf("ungrounded reference in %s: use an entity id or N.A - <reason>", cell)))
				} else if !isNAReason(value) {
					groundedReferences++
				}
			}
			if isEvidenceColumn(sectionName, header) && !isGroundedEvidence(value) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("ungrounded evidence in %s: name a command, file path, or entity id", cell)))
			}
			if sectionName == "Derived Materials" && header == "Must derive from" && !mentionsComponentSection(value) {
				issues = append(issues, strictIssue(severity, fmt.Sprintf("ungrounded derivation in %s: cite strict component sections", cell)))
			}
			if isSemanticDetailColumn(header) && !isNAReason(value) {
				normalized := normalizeSemanticText(value)
				if previous, ok := repeated[normalized]; ok {
					issues = append(issues, strictIssue(severity, fmt.Sprintf("repeated boilerplate in %s: duplicates %s", cell, previous)))
				} else if len(strings.Fields(normalized)) >= 4 {
					repeated[normalized] = cell
				}
			}
		}
	}

	if requiresGroundedReference(sectionName) && len(table.Rows) > 0 && groundedReferences == 0 {
		issues = append(issues, strictIssue(severity, fmt.Sprintf("%s needs at least one grounded reference", sectionName)))
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

func isReferenceColumn(sectionName, header string) bool {
	return header == "Reference" && (sectionName == "Foundational Flow" || sectionName == "Business Flow" || sectionName == "Governance")
}

func isEvidenceColumn(sectionName, header string) bool {
	return header == "Evidence" || header == "Required Verification"
}

func isGroundedReference(value string) bool {
	value = strings.TrimSpace(value)
	return isNAReason(value) || entityRefPattern.MatchString(value)
}

func isGroundedEvidence(value string) bool {
	value = strings.TrimSpace(value)
	return isNAReason(value) || evidencePattern.MatchString(value)
}

func requiresGroundedReference(sectionName string) bool {
	return sectionName == "Foundational Flow" || sectionName == "Business Flow" || sectionName == "Governance"
}

func mentionsComponentSection(value string) bool {
	for _, section := range componentSectionOrder {
		if strings.Contains(value, section) {
			return true
		}
	}
	return false
}

func isSemanticDetailColumn(header string) bool {
	switch header {
	case "Detail", "Governs", "Contract", "Detection", "Required Verification", "Evidence":
		return true
	default:
		return false
	}
}

func normalizeSemanticText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
