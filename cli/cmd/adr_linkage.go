package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

type adrAffectedTarget struct {
	ID   string
	Type string
}

type adrCoverage struct {
	refs  map[string][]string
	rules map[string][]string
}

func validateADRCoverage(s *store.Store, body string, severity string) []Issue {
	affected, issues := parseADRAffectedTopology(s, body, severity)
	relatedRefs, refIssues := parseADRRelatedTable(s, body, "Compliance Refs", "Ref", "ref", severity)
	issues = append(issues, refIssues...)
	relatedRules, ruleIssues := parseADRRelatedTable(s, body, "Compliance Rules", "Rule", "rule", severity)
	issues = append(issues, ruleIssues...)

	expected := expectedADRCoverage(s, affected)
	issues = append(issues, missingADRCoverageIssues(expected.refs, relatedRefs, "ref", severity)...)
	issues = append(issues, missingADRCoverageIssues(expected.rules, relatedRules, "rule", severity)...)
	return issues
}

func parseADRAffectedTopology(s *store.Store, body string, severity string) ([]adrAffectedTarget, []Issue) {
	table, ok, issues := extractADRTable(body, "Affected Topology", severity)
	if !ok {
		return nil, issues
	}
	var targets []adrAffectedTarget
	for _, row := range table.Rows {
		entityID := strings.TrimSpace(row["Entity"])
		targetType := strings.TrimSpace(row["Type"])
		whyAffected := strings.TrimSpace(row["Why affected"])
		if isNARow(entityID) || isNARow(targetType) {
			continue
		}
		if entityID == "" || targetType == "" {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  "Affected Topology rows must include both Entity and Type, or use N.A - <reason>",
				Hint:     "fill the Entity and Type cells for each affected topology row",
			})
			continue
		}
		entity, err := s.GetEntity(entityID)
		if err != nil {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("Affected Topology references unknown entity: %s", entityID),
				Hint:     "use an existing c3-* ID, or change the row to N.A - <reason>",
			})
			continue
		}
		if entity.Type != targetType {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("Affected Topology type mismatch: %s is %s, not %s", entityID, entity.Type, targetType),
				Hint:     "align the Type column with the referenced entity kind",
			})
			continue
		}
		if whyAffected == "" || isNARow(whyAffected) {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("Affected Topology row for %s must explain why it is affected", entityID),
				Hint:     "fill the Why affected column with the concrete reason, or mark the entire row N.A - <reason>",
			})
			continue
		}
		targets = append(targets, adrAffectedTarget{ID: entityID, Type: targetType})
	}
	return targets, issues
}

func parseADRRelatedTable(s *store.Store, body, sectionName, colName, targetType, severity string) (map[string]bool, []Issue) {
	table, ok, issues := extractADRTable(body, sectionName, severity)
	if !ok {
		return nil, issues
	}
	mentioned := make(map[string]bool, len(table.Rows))
	for _, row := range table.Rows {
		targetID := strings.TrimSpace(row[colName])
		whyRequired := strings.TrimSpace(row["Why required"])
		action := strings.ToLower(strings.TrimSpace(row["Action"]))
		if isNARow(targetID) {
			continue
		}
		if targetID == "" {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("%s rows must include %s, or use N.A - <reason>", sectionName, colName),
				Hint:     fmt.Sprintf("fill the %s column for each %s row", colName, sectionName),
			})
			continue
		}
		if whyRequired == "" || isNARow(whyRequired) {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("%s row for %s must explain why compliance/review is required", sectionName, targetID),
				Hint:     "fill the Why required column with the compliance reason, or mark the entire row N.A - <reason>",
			})
			continue
		}
		entity, err := s.GetEntity(targetID)
		if err != nil {
			if strings.Contains(action, "create") {
				continue
			}
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("%s references unknown %s: %s", sectionName, targetType, targetID),
				Hint:     fmt.Sprintf("create %s first, or mark the Action as create-%s", targetID, targetType),
			})
			continue
		}
		if entity.Type != targetType {
			issues = append(issues, Issue{
				Severity: severity,
				Message:  fmt.Sprintf("%s type mismatch: %s is %s, not %s", sectionName, targetID, entity.Type, targetType),
				Hint:     fmt.Sprintf("move %s to the correct ADR section", targetID),
			})
			continue
		}
		mentioned[targetID] = true
	}
	return mentioned, issues
}

func extractADRTable(body, sectionName, severity string) (*markdown.Table, bool, []Issue) {
	for _, section := range markdown.ParseSections(body) {
		if section.Name != sectionName {
			continue
		}
		table, err := markdown.ParseTable(strings.TrimSpace(section.Content))
		if err != nil {
			return nil, true, []Issue{{
				Severity: severity,
				Message:  fmt.Sprintf("invalid ADR table: %s", sectionName),
				Hint:     "use the exact table columns from c3x schema adr",
			}}
		}
		return table, true, nil
	}
	return nil, false, nil
}

func expectedADRCoverage(s *store.Store, affected []adrAffectedTarget) adrCoverage {
	coverage := adrCoverage{
		refs:  map[string][]string{},
		rules: map[string][]string{},
	}
	for _, target := range affected {
		collectADRCoverageForEntity(s, coverage, target.ID)
	}
	return coverage
}

func collectADRCoverageForEntity(s *store.Store, coverage adrCoverage, entityID string) {
	entity, err := s.GetEntity(entityID)
	if err != nil {
		return
	}
	switch entity.Type {
	case "system":
		children, _ := s.Children(entityID)
		for _, child := range children {
			if child.Type == "container" {
				collectADRCoverageForEntity(s, coverage, child.ID)
			}
		}
	case "container":
		collectScopedRefs(s, coverage.refs, entityID, fmt.Sprintf("scoped to %s", entityID))
		children, _ := s.Children(entityID)
		for _, child := range children {
			if child.Type == "component" || child.Type == "container" {
				collectADRCoverageForEntity(s, coverage, child.ID)
			}
		}
	case "component":
		if entity.ParentID != "" {
			collectScopedRefs(s, coverage.refs, entity.ParentID, fmt.Sprintf("scoped to %s via %s", entity.ParentID, entity.ID))
		}
		rels, _ := s.RelationshipsFrom(entityID)
		for _, rel := range rels {
			if rel.RelType != "uses" {
				continue
			}
			switch {
			case strings.HasPrefix(rel.ToID, "ref-"):
				coverage.refs[rel.ToID] = appendUniqueString(coverage.refs[rel.ToID], fmt.Sprintf("cited by %s", entityID))
			case strings.HasPrefix(rel.ToID, "rule-"):
				coverage.rules[rel.ToID] = appendUniqueString(coverage.rules[rel.ToID], fmt.Sprintf("cited by %s", entityID))
			}
		}
	}
}

func collectScopedRefs(s *store.Store, target map[string][]string, entityID, reason string) {
	rels, _ := s.RelationshipsTo(entityID)
	for _, rel := range rels {
		if rel.RelType != "scope" || !strings.HasPrefix(rel.FromID, "ref-") {
			continue
		}
		target[rel.FromID] = appendUniqueString(target[rel.FromID], reason)
	}
}

func missingADRCoverageIssues(expected map[string][]string, mentioned map[string]bool, targetType, severity string) []Issue {
	if len(expected) == 0 {
		return nil
	}
	var ids []string
	for id := range expected {
		if mentioned[id] {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var issues []Issue
	for _, id := range ids {
		issues = append(issues, Issue{
			Severity: severity,
			Message:  fmt.Sprintf("ADR missing compliance %s %s (%s)", targetType, id, strings.Join(expected[id], "; ")),
			Hint:     fmt.Sprintf("add %s to the ADR's compliance %ss with why it must be reviewed/complied with, or document why it is N.A", id, targetType),
		})
	}
	return issues
}

func isNARow(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || strings.HasPrefix(value, "N.A -")
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
