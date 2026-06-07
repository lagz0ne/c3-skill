package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
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

var citationHandleRE = regexp.MustCompile(`^([A-Za-z0-9_.:-]+)#n([0-9]+)@v([0-9]+):sha256:([a-f0-9]{64})\s+"(.*)"$`)

func validateADRCoverage(s *store.Store, body string, severity string) []Issue {
	return validateADRCoverageMode(s, body, severity, true)
}

func validateADRAuthoredCoverage(s *store.Store, body string, severity string) []Issue {
	return validateADRCoverageMode(s, body, severity, false)
}

func validateADRCoverageMode(s *store.Store, body string, severity string, includeMissing bool) []Issue {
	schemaCommand := adrSchemaHint()
	affected, issues := parseADRAffectedTopology(s, body, severity, schemaCommand)
	relatedRefs, refIssues := parseADRRelatedTable(s, body, "Compliance Refs", "Ref", "ref", severity, schemaCommand)
	issues = append(issues, refIssues...)
	relatedRules, ruleIssues := parseADRRelatedTable(s, body, "Compliance Rules", "Rule", "rule", severity, schemaCommand)
	issues = append(issues, ruleIssues...)

	if !includeMissing {
		return issues
	}
	expected := expectedADRCoverage(s, affected)
	issues = append(issues, missingADRCoverageIssues(expected.refs, relatedRefs, "ref", severity)...)
	issues = append(issues, missingADRCoverageIssues(expected.rules, relatedRules, "rule", severity)...)
	return issues
}

func parseADRAffectedTopology(s *store.Store, body string, severity string, schemaCommand string) ([]adrAffectedTarget, []Issue) {
	table, ok, issues := extractADRTable(body, "Affected Topology", severity, schemaCommand)
	if !ok {
		return nil, issues
	}
	if table == nil {
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
		issues = append(issues, validateADREvidence(s, "Affected Topology", entityID, strings.TrimSpace(row["Evidence"]), severity, false)...)
		targets = append(targets, adrAffectedTarget{ID: entityID, Type: targetType})
	}
	return targets, issues
}

func parseADRRelatedTable(s *store.Store, body, sectionName, colName, targetType, severity string, schemaCommand string) (map[string]bool, []Issue) {
	table, ok, issues := extractADRTable(body, sectionName, severity, schemaCommand)
	if !ok {
		return nil, issues
	}
	if table == nil {
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
				issues = append(issues, validateADREvidence(s, sectionName, targetID, strings.TrimSpace(row["Evidence"]), severity, true)...)
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
		issues = append(issues, validateADREvidence(s, sectionName, targetID, strings.TrimSpace(row["Evidence"]), severity, false)...)
		mentioned[targetID] = true
	}
	return mentioned, issues
}

func validateADREvidence(s *store.Store, sectionName, targetID, raw string, severity string, allowNA bool) []Issue {
	if raw == "" {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("%s row for %s must include Evidence citation", sectionName, targetID),
			Hint:     fmt.Sprintf("run c3x read %s --section <section> --cite and paste the matching handle, or use N.A - <reason> only when creating a new target", targetID),
		}}
	}
	if strings.HasPrefix(raw, "N.A -") {
		if allowNA {
			return nil
		}
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("%s row for %s must cite current C3 evidence, not N.A", sectionName, targetID),
			Hint:     fmt.Sprintf("run c3x read %s --section <section> --cite and paste the matching handle", targetID),
		}}
	}

	m := citationHandleRE.FindStringSubmatch(raw)
	if m == nil {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("%s row for %s has invalid Evidence citation", sectionName, targetID),
			Hint:     `expected <entity>#n<node>@v<version>:sha256:<nodeHash> "exact snippet" from c3x read --cite`,
		}}
	}

	citedEntity := m[1]
	nodeID, _ := strconv.ParseInt(m[2], 10, 64)
	version, _ := strconv.Atoi(m[3])
	hash := m[4]
	snippet := m[5]

	if citedEntity != targetID {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("Evidence for %s row %s cites %s, want %s", sectionName, targetID, citedEntity, targetID),
			Hint:     "use evidence generated from the row target, not a nearby document",
		}}
	}

	entity, err := s.GetEntity(citedEntity)
	if err != nil {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("Evidence for %s row %s cites unknown entity %s", sectionName, targetID, citedEntity),
			Hint:     "create the target first, or use N.A - <reason> with a create action",
		}}
	}
	if entity.Version != version {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("Evidence for %s row %s cites version %d, current version is %d", sectionName, targetID, version, entity.Version),
			Hint:     fmt.Sprintf("refresh the handle with c3x read %s --cite", targetID),
		}}
	}

	if evidenceNodeMatches(s, citedEntity, nodeID, hash, snippet) {
		return nil
	}
	if node, err := s.GetNode(nodeID); err == nil && node.EntityID != citedEntity {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("Evidence for %s row %s cites node %d from %s", sectionName, targetID, nodeID, node.EntityID),
			Hint:     fmt.Sprintf("refresh the handle with c3x read %s --cite", targetID),
		}}
	}
	if snippet == "" {
		return []Issue{{
			Severity: severity,
			Message:  fmt.Sprintf("Evidence for %s row %s has empty snippet", sectionName, targetID),
			Hint:     "paste the exact quoted snippet emitted by c3x read --cite",
		}}
	}
	return []Issue{{
		Severity: severity,
		Message:  fmt.Sprintf("Evidence for %s row %s has stale node hash or snippet", sectionName, targetID),
		Hint:     fmt.Sprintf("refresh the handle with c3x read %s --cite", targetID),
	}}
}

func evidenceNodeMatches(s *store.Store, entityID string, nodeID int64, hash, snippet string) bool {
	if snippet == "" {
		return false
	}
	if node, err := s.GetNode(nodeID); err == nil {
		if node.EntityID == entityID && node.Hash == hash && strings.Contains(node.Content, snippet) {
			return true
		}
	}
	nodes, err := s.NodesForEntity(entityID)
	if err != nil {
		return false
	}
	for _, node := range nodes {
		if node.Hash == hash && strings.Contains(node.Content, snippet) {
			return true
		}
	}
	return false
}

func extractADRTable(body, sectionName, severity string, schemaCommand string) (*markdown.Table, bool, []Issue) {
	for _, section := range markdown.ParseSections(body) {
		if section.Name != sectionName {
			continue
		}
		table, err := markdown.ParseTable(strings.TrimSpace(section.Content))
		if err != nil {
			return nil, true, []Issue{{
				Severity: severity,
				Message:  fmt.Sprintf("invalid ADR table: %s", sectionName),
				Hint:     fmt.Sprintf("use the exact table columns from %s", schemaCommand),
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
			switch citationType(s, rel.ToID) {
			case "ref":
				coverage.refs[rel.ToID] = appendUniqueString(coverage.refs[rel.ToID], fmt.Sprintf("cited by %s", entityID))
			case "rule":
				coverage.rules[rel.ToID] = appendUniqueString(coverage.rules[rel.ToID], fmt.Sprintf("cited by %s", entityID))
			}
		}
	}
}

// citationType classifies a citation target by its real entity type from the
// store, so linkage does not depend on the id-prefix naming convention. It
// falls back to the prefix only when the entity is absent (dangling citation),
// preserving prior behavior on malformed input.
func citationType(s *store.Store, id string) string {
	if e, err := s.GetEntity(id); err == nil {
		return e.Type
	}
	switch {
	case strings.HasPrefix(id, "ref-"):
		return "ref"
	case strings.HasPrefix(id, "rule-"):
		return "rule"
	case strings.HasPrefix(id, "adr-"):
		return "adr"
	case strings.HasPrefix(id, "recipe-"):
		return "recipe"
	default:
		return ""
	}
}

func collectScopedRefs(s *store.Store, target map[string][]string, entityID, reason string) {
	rels, _ := s.RelationshipsTo(entityID)
	for _, rel := range rels {
		if rel.RelType != "scope" || citationType(s, rel.FromID) != "ref" {
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

// isADRTerminal reports whether an ADR status is a terminal (historical) state.
// Terminal-state ADRs are exempt from check validation; their content is frozen.
func isADRTerminal(status string) bool {
	return status == "implemented" || status == "provisioned"
}
