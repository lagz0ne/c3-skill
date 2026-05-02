package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// Issue represents a validation finding.
type Issue struct {
	Severity string `json:"severity"`
	Entity   string `json:"entity,omitempty"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
	Matched  string `json:"matched,omitempty"`
}

// CheckResult holds the validation output.
type CheckResult struct {
	Total  int        `json:"total"`
	Issues []Issue    `json:"issues"`
	Help   []HelpHint `json:"help,omitempty"`
}

// CheckOptions holds parameters for the enhanced check command.
type CheckOptions struct {
	Store      *store.Store
	JSON       bool
	ProjectDir string
	C3Dir      string
	IncludeADR bool
	Fix        bool
	Only       []string
	Rules      []string
}

// buildTitleMapStore creates a case-insensitive title/slug -> entity ID lookup.
func buildTitleMapStore(s *store.Store) map[string]string {
	m := make(map[string]string)
	entities, _ := s.AllEntities()
	for _, e := range entities {
		lower := strings.ToLower(e.Title)
		if existing, exists := m[lower]; exists {
			if existing != e.ID {
				m[lower] = "" // ambiguous
			}
		} else {
			m[lower] = e.ID
		}
		if e.Slug != "" {
			slugLower := strings.ToLower(e.Slug)
			if existing, exists := m[slugLower]; exists {
				if existing != e.ID {
					m[slugLower] = ""
				}
			} else {
				m[slugLower] = e.ID
			}
		}
	}
	return m
}

// suggestByTitle returns a suggested entity ID if the value matches a title/slug, or "".
func suggestByTitle(val string, titleMap map[string]string) string {
	if id := titleMap[strings.ToLower(val)]; id != "" {
		return id
	}
	return ""
}

// resolveRuleCiters errors if a rule has no citers — silently checking
// nothing would mask the misuse.
func resolveRuleCiters(s *store.Store, ruleIDs []string) ([]string, error) {
	seen := map[string]bool{}
	var ids []string
	for _, ruleID := range ruleIDs {
		if _, err := s.GetEntity(ruleID); err != nil {
			return nil, fmt.Errorf("--rule %s: %w", ruleID, err)
		}
		rels, err := s.RelationshipsTo(ruleID)
		if err != nil {
			return nil, fmt.Errorf("--rule %s citers: %w", ruleID, err)
		}
		found := false
		for _, r := range rels {
			if r.RelType != "uses" {
				continue
			}
			found = true
			if seen[r.FromID] {
				continue
			}
			seen[r.FromID] = true
			ids = append(ids, r.FromID)
		}
		if !found {
			return nil, fmt.Errorf("rule %s has no citers. Wire one with: c3x wire <component> %s\nOr check a different rule.", ruleID, ruleID)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func hintFor(message string) string {
	patterns := []struct {
		substr string
		hint   string
	}{
		{"code-map parse error", "check code-map entries with 'c3x list'"},
		{"missing required section", ""},
		{"empty required section", ""},
		{"empty required table", "add at least one data row below the table headers"},
		{"unknown entity reference", "verify the ID with 'c3x list'; check for typos"},
		{"unknown ref reference", "use a ref-* ID (e.g., ref-jwt); verify with 'c3x list'"},
		{"file does not exist", "create the file or fix the path"},
		{"layer disconnect", "update parent table or fix the child parent field; rebuild only proves storage, not layer integration"},
	}
	for _, p := range patterns {
		if !strings.Contains(message, p.substr) {
			continue
		}
		if p.hint != "" {
			return p.hint
		}
		if strings.Contains(message, "missing required section: ") {
			section := strings.TrimPrefix(message, "missing required section: ")
			return fmt.Sprintf("add a ## %s section with content", section)
		}
		if strings.Contains(message, "empty required section: ") {
			section := strings.TrimPrefix(message, "empty required section: ")
			return fmt.Sprintf("add content to the ## %s section", section)
		}
	}
	return ""
}

func countSeverities(issues []Issue) (errors, warnings int) {
	for _, issue := range issues {
		if issue.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}
	return
}

func formatCounts(errors, warnings int) string {
	var parts []string
	if errors > 0 {
		noun := "error"
		if errors > 1 {
			noun = "errors"
		}
		parts = append(parts, fmt.Sprintf("%d %s", errors, noun))
	}
	if warnings > 0 {
		noun := "warning"
		if warnings > 1 {
			noun = "warnings"
		}
		parts = append(parts, fmt.Sprintf("%d %s", warnings, noun))
	}
	return strings.Join(parts, ", ")
}

func checkRecipeSourcesStore(s *store.Store) []Issue {
	var issues []Issue
	recipes, _ := s.EntitiesByType("recipe")
	for _, r := range recipes {
		rels, _ := s.RelationshipsFrom(r.ID)
		for _, rel := range rels {
			if rel.RelType == "sources" {
				if _, err := s.GetEntity(rel.ToID); err != nil {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   r.ID,
						Message:  fmt.Sprintf("recipe references nonexistent entity: %s", rel.ToID),
					})
				}
			}
		}
	}
	return issues
}

// RunCheckV2 validates entities against the schema registry.
func RunCheckV2(opts CheckOptions, w io.Writer) error {
	var issues []Issue

	titleMap := buildTitleMapStore(opts.Store)

	entities, _ := opts.Store.AllEntities()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	if len(opts.Rules) > 0 {
		ruleCiters, err := resolveRuleCiters(opts.Store, opts.Rules)
		if err != nil {
			return err
		}
		opts.Only = append(opts.Only, ruleCiters...)
	}

	targetMatcher := newCheckTargetMatcher(entities, opts.Only)

	for _, entity := range entities {
		if entity.Type == "adr" {
			if !opts.IncludeADR {
				continue
			}
			if isADRTerminal(entity.Status) && !slices.Contains(opts.Only, entity.ID) {
				continue
			}
		}
		if !targetMatcher.matches(entity) {
			continue
		}

		schemaSections := schema.ForType(entity.Type)
		if schemaSections == nil {
			continue
		}

		// Read body from node tree.
		body, err := content.ReadEntity(opts.Store, entity.ID)
		if err != nil {
			body = ""
		}

		bodySections := markdown.ParseSections(body)

		// Build lookup of parsed sections by name
		sectionMap := make(map[string]markdown.Section)
		for _, s := range bodySections {
			if s.Name != "" {
				sectionMap[s.Name] = s
			}
		}

		if entity.Type != "component" {
			for _, schemaDef := range schemaSections {
				if !schemaDef.Required {
					continue
				}

				bodySection, exists := sectionMap[schemaDef.Name]

				if !exists {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("missing required section: %s", schemaDef.Name),
					})
					continue
				}

				content := strings.TrimSpace(bodySection.Content)

				if schemaDef.ContentType == "table" {
					if content == "" {
						issues = append(issues, Issue{
							Severity: "warning",
							Entity:   entity.ID,
							Message:  fmt.Sprintf("empty required table: %s", schemaDef.Name),
						})
						continue
					}
					table, err := markdown.ParseTable(content)
					if err != nil {
						continue
					}
					if len(table.Rows) == 0 {
						issues = append(issues, Issue{
							Severity: "warning",
							Entity:   entity.ID,
							Message:  fmt.Sprintf("empty required table: %s (headers only, no data rows)", schemaDef.Name),
						})
						continue
					}

					for _, col := range schemaDef.Columns {
						issues = append(issues, validateColumn(col, table, entity, opts, titleMap)...)
					}
				} else {
					if content == "" {
						issues = append(issues, Issue{
							Severity: "warning",
							Entity:   entity.ID,
							Message:  fmt.Sprintf("empty required section: %s", schemaDef.Name),
						})
					}
				}
			}
		}
		if entity.Type == "adr" {
			for _, issue := range validateADRCoverage(opts.Store, body, "warning") {
				issue.Entity = entity.ID
				issues = append(issues, issue)
			}
		}
		if entity.Type == "component" {
			for _, issue := range validateStrictComponentDoc(body, "error") {
				issue.Entity = entity.ID
				issues = append(issues, issue)
			}
		}

		// Layer 3: Check non-required table sections too
		for _, schemaDef := range schemaSections {
			if schemaDef.Required || schemaDef.ContentType != "table" {
				continue
			}
			bodySection, exists := sectionMap[schemaDef.Name]
			if !exists {
				continue
			}
			content := strings.TrimSpace(bodySection.Content)
			if content == "" {
				continue
			}
			table, err := markdown.ParseTable(content)
			if err != nil || len(table.Rows) == 0 {
				continue
			}
			for _, col := range schemaDef.Columns {
				issues = append(issues, validateColumn(col, table, entity, opts, titleMap)...)
			}
		}

		// Validate origin references for rules
		if entity.Type == "rule" {
			rels, _ := opts.Store.RelationshipsFrom(entity.ID)
			for _, r := range rels {
				if r.RelType == "origin" {
					if _, err := opts.Store.GetEntity(r.ToID); err != nil {
						issues = append(issues, Issue{
							Severity: "error",
							Entity:   entity.ID,
							Message:  fmt.Sprintf("origin reference %q not found", r.ToID),
							Hint:     "origin should reference an existing ref or ADR entity",
						})
					}
				}
			}
		}
	}

	// Code-map validation
	allCodeMap, err := opts.Store.AllCodeMap()
	if err == nil && len(allCodeMap) > 0 && opts.ProjectDir != "" {
		for entityID, patterns := range allCodeMap {
			entity, err := opts.Store.GetEntity(entityID)
			if err != nil {
				if len(opts.Only) > 0 {
					continue
				}
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entityID,
					Message:  fmt.Sprintf("code-map entity %s not found in store", entityID),
				})
				continue
			}
			if !targetMatcher.matches(entity) {
				continue
			}
			if entity.Type != "component" && entity.Type != "ref" && entity.Type != "rule" {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entityID,
					Message:  fmt.Sprintf("code-map: %s is not a component or ref", entityID),
				})
			}
			for _, p := range patterns {
				if p == "" {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entityID,
						Message:  "empty path in code-map",
					})
				}
			}
		}
	}

	if len(opts.Only) == 0 {
		issues = append(issues, checkRecipeSourcesStore(opts.Store)...)
		issues = append(issues, checkLayerDisconnectsStore(opts.Store)...)
	} else {
		issues = append(issues, filterIssuesByTargets(opts.Store, targetMatcher, checkRecipeSourcesStore(opts.Store))...)
		issues = append(issues, filterIssuesByTargets(opts.Store, targetMatcher, checkLayerDisconnectsStore(opts.Store))...)
	}

	// Scope cross-check
	refs, _ := opts.Store.EntitiesByType("ref")
	for _, ref := range refs {
		rels, _ := opts.Store.RelationshipsFrom(ref.ID)
		for _, r := range rels {
			if r.RelType != "scope" {
				continue
			}
			children, _ := opts.Store.Children(r.ToID)
			for _, child := range children {
				if child.Type != "component" {
					continue
				}
				if !targetMatcher.matches(child) {
					continue
				}
				// Check if child cites this ref
				childRels, _ := opts.Store.RelationshipsFrom(child.ID)
				cited := false
				for _, cr := range childRels {
					if cr.RelType == "uses" && cr.ToID == ref.ID {
						cited = true
						break
					}
				}
				if !cited {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   child.ID,
						Message:  fmt.Sprintf("ref %s scopes %s but %s does not cite it", ref.ID, r.ToID, child.ID),
					})
				}
			}
		}
	}

	result := CheckResult{
		Total:  len(entities),
		Issues: issues,
	}
	if result.Issues == nil {
		result.Issues = []Issue{}
	}

	// Output
	if opts.JSON {
		for i := range result.Issues {
			result.Issues[i].Hint = hintFor(result.Issues[i].Message)
		}
		result.Help = agentHints(cascadeReviewHints())
		if err := writeJSON(w, result); err != nil {
			return err
		}
		if errors, _ := countSeverities(result.Issues); errors > 0 {
			return fmt.Errorf("check failed: %d error(s)", errors)
		}
		return nil
	}

	// Text output
	errors, warnings := countSeverities(result.Issues)
	if len(result.Issues) == 0 {
		fmt.Fprintf(w, "Checked %d docs — all clear\n", result.Total)
		return nil
	}

	fmt.Fprintf(w, "Checked %d docs — %s\n\n", result.Total, formatCounts(errors, warnings))

	for _, issue := range result.Issues {
		entityLabel := issue.Entity
		if entityLabel == "" {
			entityLabel = "global"
		}
		icon := "!"
		if issue.Severity == "error" {
			icon = "x"
		}
		fmt.Fprintf(w, "  %s %s: %s\n", icon, entityLabel, issue.Message)
		if hint := hintFor(issue.Message); hint != "" {
			fmt.Fprintf(w, "    -> %s\n", hint)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "Legend: x = error (fix first)  ! = warning (incomplete, should fix)")
	writeAgentHints(w, cascadeReviewHints())

	if errors > 0 {
		return fmt.Errorf("check failed: %d error(s)", errors)
	}
	return nil
}

type checkTargetMatcher struct {
	targets    []string
	parentSlug map[string]string
}

func newCheckTargetMatcher(entities []*store.Entity, targets []string) checkTargetMatcher {
	m := checkTargetMatcher{targets: targets, parentSlug: map[string]string{}}
	for _, e := range entities {
		if e.Type == "container" {
			m.parentSlug[e.ID] = fmt.Sprintf("%s-%s", e.ID, e.Slug)
		}
	}
	return m
}

func (m checkTargetMatcher) matches(entity *store.Entity) bool {
	if len(m.targets) == 0 {
		return true
	}
	if entity == nil {
		return false
	}
	return verifyTargetMatchesDoc(m.targets, entity.ID, entityRelativePath(entity, m.parentSlug))
}

func filterIssuesByTargets(s *store.Store, matcher checkTargetMatcher, issues []Issue) []Issue {
	filtered := issues[:0]
	for _, issue := range issues {
		if issue.Entity == "" {
			continue
		}
		entity, err := s.GetEntity(issue.Entity)
		if err != nil {
			continue
		}
		if matcher.matches(entity) {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func checkLayerDisconnectsStore(s *store.Store) []Issue {
	entities, _ := s.AllEntities()
	entityByID := make(map[string]*store.Entity, len(entities))
	for _, entity := range entities {
		entityByID[entity.ID] = entity
	}

	var issues []Issue
	for _, parent := range entities {
		sectionName, childType := layerSection(parent.Type)
		if sectionName == "" {
			continue
		}

		body, err := content.ReadEntity(s, parent.ID)
		if err != nil {
			continue
		}
		tableIDs, ok := idsFromSectionTable(body, sectionName)
		if !ok {
			continue
		}

		children, _ := s.Children(parent.ID)
		for _, child := range children {
			if child.Type != childType {
				continue
			}
			if !tableIDs[child.ID] {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   parent.ID,
					Message:  fmt.Sprintf("layer disconnect: child %s %s has parent %s but is missing from %s %s table", child.Type, child.ID, parent.ID, parent.ID, sectionName),
				})
			}
		}

		for id := range tableIDs {
			if id == "" {
				continue
			}
			child := entityByID[id]
			if child == nil {
				continue
			}
			if child.Type != childType {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   parent.ID,
					Message:  fmt.Sprintf("layer disconnect: %s listed in %s %s table but type is %s, not %s", id, parent.ID, sectionName, child.Type, childType),
				})
				continue
			}
			if child.ParentID != parent.ID {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   parent.ID,
					Message:  fmt.Sprintf("layer disconnect: %s listed in %s %s table but parent is %s", id, parent.ID, sectionName, child.ParentID),
				})
			}
		}
	}
	return issues
}

func layerSection(entityType string) (sectionName string, childType string) {
	switch entityType {
	case "system", "context":
		return "Containers", "container"
	case "container":
		return "Components", "component"
	default:
		return "", ""
	}
}

func idsFromSectionTable(body, sectionName string) (map[string]bool, bool) {
	for _, section := range markdown.ParseSections(body) {
		if section.Name != sectionName {
			continue
		}
		table, err := markdown.ParseTable(section.Content)
		if err != nil {
			return nil, false
		}
		ids := make(map[string]bool, len(table.Rows))
		for _, row := range table.Rows {
			ids[strings.TrimSpace(row["ID"])] = true
		}
		return ids, true
	}
	return nil, false
}

// validateColumn checks typed column values in a table.
func validateColumn(col schema.ColumnDef, table *markdown.Table, entity *store.Entity, opts CheckOptions, titleMap map[string]string) []Issue {
	var issues []Issue
	switch col.Type {
	case "filepath":
		if opts.ProjectDir == "" {
			return nil
		}
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			absPath := filepath.Join(opts.ProjectDir, val)
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("file does not exist: %s", val),
				})
			}
		}
	case "entity_id":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			if _, err := opts.Store.GetEntity(val); err != nil {
				msg := fmt.Sprintf("unknown entity reference: %s", val)
				if suggested := suggestByTitle(val, titleMap); suggested != "" {
					msg = fmt.Sprintf("unknown entity reference: %s (did you mean %s?)", val, suggested)
				}
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  msg,
				})
			}
		}
	case "ref_id":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			if _, err := opts.Store.GetEntity(val); err != nil {
				msg := fmt.Sprintf("unknown ref reference: %s", val)
				if suggested := suggestByTitle(val, titleMap); suggested != "" {
					msg = fmt.Sprintf("unknown ref reference: %s (did you mean %s?)", val, suggested)
				}
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  msg,
				})
			}
		}
	}
	return issues
}
