package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
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
	// StrictCodemap promotes the change-unit codemap introspection (external binding
	// verification) from WARN to error. Off by default: an unresolved external
	// binding is reported, not gated.
	StrictCodemap bool
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
		{"layer disconnect", "open a change doc (c3 add adr) that amends the parent table top-down; rebuild only proves storage, not layer integration"},
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

func checkProjectCanvases(c3Dir string) []Issue {
	if c3Dir == "" {
		return nil
	}
	if _, err := schema.LoadProjectCanvases(c3Dir); err != nil {
		return []Issue{{
			Severity: "error",
			Entity:   "canvas",
			Message:  err.Error(),
			Hint:     "run c3x canvas list and repair the invalid canvas file",
		}}
	}
	return nil
}

func checkProjectCanvasesForTargets(c3Dir string, targets []string) []Issue {
	if c3Dir == "" || len(targets) == 0 {
		return nil
	}
	for _, target := range targets {
		target = strings.TrimSpace(filepath.ToSlash(target))
		if target == "" {
			continue
		}
		if strings.HasPrefix(target, "canvases/") || strings.HasPrefix(target, ".c3/canvases/") {
			return checkProjectCanvases(c3Dir)
		}
		if strings.HasSuffix(target, ".md") {
			continue
		}
		path := filepath.Join(c3Dir, schema.CanvasesDir, target+".md")
		if _, err := os.Stat(path); err == nil {
			return checkProjectCanvases(c3Dir)
		}
	}
	return nil
}

// RunCheckV2 validates entities against the schema registry.
func RunCheckV2(opts CheckOptions, w io.Writer) error {
	var issues []Issue

	titleMap := buildTitleMapStore(opts.Store)

	entities, _ := opts.Store.AllEntities()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	// `check --fix` heals membership by construction: reconcile every parent's table
	// from its children's parent: edges, so a disconnect left by any path (a direct
	// `c3 add`, a hand-edit) is repaired, not merely reported below.
	if opts.Fix {
		if err := healMembership(opts.Store, opts.C3Dir); err != nil {
			return err
		}
	}

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
		// Terminal change docs (prd/atomic and any non-ADR change doc) are frozen:
		// their content is historical, so the default check skips them. `--only`
		// naming the doc still inspects it. The ADR `--include-adr` default-flip is
		// deferred to migration; this is the generalized terminal-skip only.
		if entity.Type != "adr" && schema.IsChangeDoc(entity.Type) &&
			isChangeDocTerminal(entity) && !slices.Contains(opts.Only, entity.ID) {
			continue
		}
		if !targetMatcher.matches(entity) {
			continue
		}

		def, ok := schema.DefinitionForDir(opts.C3Dir, entity.Type)
		if !ok {
			continue
		}
		schemaSections := def.Sections

		// Read body from node tree.
		body, err := content.ReadEntity(opts.Store, entity.ID)
		if err != nil {
			body = ""
		}

		// AUTO-DONE latch: an `accepted` change doc whose per-row After cites all
		// resolve fresh is ready to actualize accepted->done, one-way. The flip is
		// gated behind --fix: a plain `check` (opts.Fix == false) only REPORTS
		// readiness and never mutates the DB or rewrites sealed markdown; only
		// `check --fix` performs the flip. On a committed flip the doc is now
		// terminal/frozen, so its discharge is not re-validated this pass.
		if schema.IsChangeDocDir(opts.C3Dir, entity.Type) && entity.Status == "accepted" {
			// External matching arm (double-V right side): verify the unit's affected
			// code bindings still resolve. Runs BEFORE the latch so it reports even
			// when the doc auto-dones this pass.
			codemapIssues := codemapIntrospection(opts.Store, opts.ProjectDir, body, opts.StrictCodemap)
			issues = append(issues, codemapIssues...)
			// Under --strict-codemap an unresolved external binding is a GATE on done:
			// the doc may not actualize accepted->done while a declared code binding does
			// not resolve, exactly as an unresolved internal After-cite blocks it. Without
			// --strict it is WARN-only and the flip proceeds.
			strictBlocked := opts.StrictCodemap && hasErrorSeverity(codemapIssues)
			if ready, _ := autoDoneLatch(opts.Store, opts.C3Dir, entity, body, opts.Fix && !strictBlocked); ready {
				switch {
				case strictBlocked:
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("%s is ready to auto-done but --strict-codemap blocks it until every declared code binding resolves", entity.ID),
					})
				case opts.Fix:
					issues = append(issues, Issue{
						Severity: "info",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("auto-done: all After cites resolved fresh; %s actualized accepted->done", entity.ID),
					})
					continue
				default:
					issues = append(issues, Issue{
						Severity: "info",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("%s ready to auto-done: all After cites resolve fresh; run 'c3x check --fix' to actualize accepted->done", entity.ID),
					})
				}
			}
		}

		bodySections := markdown.ParseSections(body)

		// Build lookup of parsed sections by name
		sectionMap := make(map[string]markdown.Section)
		for _, s := range bodySections {
			if s.Name != "" {
				sectionMap[s.Name] = s
			}
		}

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
					if isToolMaintainedTable(entity.Type, schemaDef.Name) {
						continue // membership table — the reconciler owns its rows
					}
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("empty required table: %s (headers only, no data rows)", schemaDef.Name),
					})
					continue
				}

				for _, col := range schemaDef.Columns {
					if !containsString(table.Headers, col.Name) {
						issues = append(issues, Issue{
							Severity: "warning",
							Entity:   entity.ID,
							Message:  fmt.Sprintf("missing required column %q in table: %s", col.Name, schemaDef.Name),
						})
						continue
					}
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
		if entity.Type == "adr" {
			for _, issue := range validateADRCoverage(opts.Store, body, "warning") {
				issue.Entity = entity.ID
				issues = append(issues, issue)
			}
		}
		// The STRICT change-set on a change doc is format + field/type checked,
		// and a touch-nothing change-set FAILS discharge. The component-only gate
		// is extended to change-doc canvases via the declared-status predicate,
		// scoped to the STRICT (non-FREE) sections by deriveStrictRules. ADR keeps
		// its dedicated validateADRCoverage discharge (above) and is not
		// double-checked here.
		isStrictChangeDoc := schema.IsChangeDoc(entity.Type) && entity.Type != "adr"
		if entity.Type == "component" || isStrictChangeDoc {
			for _, issue := range validateStrictDoc(schemaSections, body, "error") {
				issue.Entity = entity.ID
				issues = append(issues, issue)
			}
		}
		if isStrictChangeDoc && changeDocTouchesNothing(schemaSections, body) {
			issues = append(issues, Issue{
				Severity: "error",
				Entity:   entity.ID,
				Message:  "change doc touches nothing: every STRICT change-set row is entirely N.A",
				Hint:     "name at least one real change in the change-set, or this doc changes nothing",
			})
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
				if !containsString(table.Headers, col.Name) {
					continue
				}
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
		issues = append(issues, checkProjectCanvases(opts.C3Dir)...)
		issues = append(issues, checkRecipeSourcesStore(opts.Store)...)
		issues = append(issues, checkLayerDisconnectsStore(opts.Store)...)
		issues = append(issues, checkFactSealsOnDisk(opts.C3Dir)...)
	} else {
		issues = append(issues, checkProjectCanvasesForTargets(opts.C3Dir, opts.Only)...)
		issues = append(issues, filterIssuesByTargets(opts.Store, targetMatcher, checkRecipeSourcesStore(opts.Store))...)
		issues = append(issues, filterIssuesByTargets(opts.Store, targetMatcher, checkLayerDisconnectsStore(opts.Store))...)
		issues = append(issues, filterIssuesByTargets(opts.Store, targetMatcher, checkFactSealsOnDisk(opts.C3Dir))...)
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

// checkFactSealsOnDisk walks the on-disk .c3 tree and, for each doc whose content
// no longer matches its committed c3-seal, emits a mechanical WARN. This is the
// only seal verification on the check path; it mirrors the import-path seal check
// without removing it.
//
// Scope covers BOTH facts and change docs (adr/prd/atomic-design-change). Change
// docs mutate through the sanctioned status-writer path, which reseals on export;
// so an on-disk seal mismatch on a change doc means its body/status was
// hand-edited without resealing — tampering that must surface during routine
// `check`, not only at import. The WARN is purely mechanical — it reports a
// content/seal mismatch and never judges whether the edit was a legitimate reseal
// or a sneaky hand-edit (provenance is judgment), and it never escalates to a hard
// error/FAIL.
func checkFactSealsOnDisk(c3Dir string) []Issue {
	if c3Dir == "" {
		return nil
	}
	docs, err := walker.WalkC3Docs(c3Dir)
	if err != nil {
		return nil
	}
	var issues []Issue
	for _, doc := range docs {
		if doc.Frontmatter == nil {
			continue
		}
		// Canvases are user-owned definitions, not facts; their seals are governed by
		// structural validation + the embedded-seal test, not this fact-seal-on-disk
		// check (which otherwise flags a benign renderCanvasDoc-vs-verify seal format
		// difference on status-bearing seed canvases at fresh init).
		if doc.Frontmatter.Type == "canvas" {
			continue
		}
		actual, expected := verifyParsedDocSeal(doc)
		if actual == "" || expected == "" {
			continue
		}
		if actual == expected {
			continue
		}
		issues = append(issues, Issue{
			Severity: "warning",
			Entity:   doc.Frontmatter.ID,
			Message:  fmt.Sprintf("seal mismatch: %s content was hand-edited since its last seal (have %s, recomputed %s)", doc.Frontmatter.ID, sealPrefix(actual), sealPrefix(expected)),
			Hint:     "reseal through the sanctioned path (e.g. 'c3x repair') if the edit is intended, or revert the hand-edit",
		})
	}
	return issues
}

func sealPrefix(seal string) string {
	if len(seal) > 12 {
		return seal[:12]
	}
	return seal
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
	case "reference":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" || isNAReason(val) {
				continue
			}
			// Grounded if any token resolves to a real entity — ANY id shape,
			// builtin or custom-type. An edge column wires by resolution, so a
			// hardcoded builtin-id pattern would falsely flag a cite to a custom-
			// canvas fact (a requirement/objective/design-token id) as ungrounded.
			grounded := false
			for _, tok := range referenceTokens(val) {
				if _, err := opts.Store.GetEntity(tok); err == nil {
					grounded = true
					break
				}
			}
			refs := entityRefPattern.FindAllString(val, -1)
			if !grounded && len(refs) == 0 {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("ungrounded reference in %s: %s", col.Name, val),
					Hint:     "use an entity id or N.A - <reason>",
				})
				continue
			}
			for _, ref := range refs {
				if _, err := opts.Store.GetEntity(ref); err != nil {
					msg := fmt.Sprintf("unknown entity reference: %s", ref)
					if suggested := suggestByTitle(ref, titleMap); suggested != "" {
						msg = fmt.Sprintf("unknown entity reference: %s (did you mean %s?)", ref, suggested)
					}
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  msg,
					})
				}
			}
		}
	case "evidence":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" || isNAReason(val) {
				continue
			}
			if !isGroundedEvidence(val) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("ungrounded evidence in %s: %s", col.Name, val),
					Hint:     "name a command, file path, or entity id",
				})
			}
		}
	case "cite":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" || isNAReason(val) {
				continue
			}
			issues = append(issues, validateCitationColumnValue(val, entity, opts)...)
		}
	case "check":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" || isNAReason(val) {
				continue
			}
			if placeholderPattern.MatchString(val) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("placeholder check result in %s: %s", col.Name, val),
					Hint:     "record an observed command result or N.A - <reason>",
				})
			}
		}
	case "enum":
		for _, row := range table.Rows {
			val := strings.TrimSpace(row[col.Name])
			if val == "" {
				continue
			}
			if !enumValueAllowed(val, col.Values) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   entity.ID,
					Message:  fmt.Sprintf("invalid enum value in %s: %s", col.Name, val),
					Hint:     fmt.Sprintf("expected one of: %s", strings.Join(col.Values, ", ")),
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
	default:
		if strings.HasPrefix(col.Type, "edge<") {
			for _, row := range table.Rows {
				val := strings.TrimSpace(row[col.Name])
				if val == "" || isNAReason(val) {
					continue
				}
				if placeholderPattern.MatchString(val) {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   entity.ID,
						Message:  fmt.Sprintf("placeholder edge value in %s: %s", col.Name, val),
						Hint:     "record a concrete edge target or N.A - <reason>",
					})
				}
			}
		}
	}
	return issues
}

func enumValueAllowed(value string, allowed []string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
		if strings.HasPrefix(candidate, "N.A") && isNAReason(value) {
			return true
		}
	}
	return false
}

// validateCitationColumnValue applies the same freshness/hash-intactness check the
// ADR Evidence path uses (validateADREvidence) to ANY cite column: version,
// node-hash, and snippet must still match the cited target. It only answers "is
// this handle still current/intact?" — it never judges whether the cited node is
// the right evidence for the claim.
func validateCitationColumnValue(raw string, entity *store.Entity, opts CheckOptions) []Issue {
	m := citationHandleRE.FindStringSubmatch(raw)
	if m == nil {
		return []Issue{{
			Severity: "warning",
			Entity:   entity.ID,
			Message:  fmt.Sprintf("invalid citation handle: %s", raw),
			Hint:     `expected <entity>#n<node>@v<version>:sha256:<nodeHash> "exact snippet" from c3x read --cite`,
		}}
	}
	citedEntity := m[1]
	nodeID, _ := strconv.ParseInt(m[2], 10, 64)
	version, _ := strconv.Atoi(m[3])
	hash := m[4]
	snippet := m[5]

	cited, err := opts.Store.GetEntity(citedEntity)
	if err != nil {
		return []Issue{{
			Severity: "warning",
			Entity:   entity.ID,
			Message:  fmt.Sprintf("citation references unknown entity: %s", citedEntity),
		}}
	}

	if cited.Version != version {
		return []Issue{{
			Severity: "warning",
			Entity:   entity.ID,
			Message:  fmt.Sprintf("citation to %s cites version %d, current version is %d", citedEntity, version, cited.Version),
			Hint:     fmt.Sprintf("refresh the handle with c3x read %s --cite", citedEntity),
		}}
	}

	if evidenceNodeMatches(opts.Store, citedEntity, nodeID, hash, snippet) {
		return nil
	}
	if node, err := opts.Store.GetNode(nodeID); err == nil && node.EntityID != citedEntity {
		return []Issue{{
			Severity: "warning",
			Entity:   entity.ID,
			Message:  fmt.Sprintf("citation to %s cites node %d from %s", citedEntity, nodeID, node.EntityID),
			Hint:     fmt.Sprintf("refresh the handle with c3x read %s --cite", citedEntity),
		}}
	}
	if snippet == "" {
		return []Issue{{
			Severity: "warning",
			Entity:   entity.ID,
			Message:  fmt.Sprintf("citation to %s has empty snippet", citedEntity),
			Hint:     "paste the exact quoted snippet emitted by c3x read --cite",
		}}
	}
	return []Issue{{
		Severity: "warning",
		Entity:   entity.ID,
		Message:  fmt.Sprintf("citation to %s has stale node hash or snippet", citedEntity),
		Hint:     fmt.Sprintf("refresh the handle with c3x read %s --cite", citedEntity),
	}}
}

// referenceTokens splits a reference-cell value into candidate entity-id tokens —
// reference columns name ids (one, or comma/space/pipe separated), so each is
// resolved directly, making grounding work for any id shape (builtin or custom-type).
func referenceTokens(val string) []string {
	raw := strings.FieldsFunc(val, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == '|' || r == ';' || r == '\n'
	})
	var out []string
	for _, f := range raw {
		f = strings.Trim(f, "`\"'()[]")
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}
