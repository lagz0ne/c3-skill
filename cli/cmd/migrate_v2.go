package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

type MigrateV2Options struct {
	Store  *store.Store
	DryRun bool
}

type migrateBlocker struct {
	ID     string
	Title  string
	Issues []Issue
}

func RunMigrateV2(opts MigrateV2Options, w io.Writer) error {
	entities, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("listing entities: %w", err)
	}

	blockers, err := collectMigrateBlockers(opts.Store, entities)
	if err != nil {
		return err
	}
	if len(blockers) > 0 {
		fmt.Fprintf(w, "\n0 migrated, %d blocked\n", len(blockers))
		writeMigrateBlockerPlan(w, blockers)
		return fmt.Errorf("migrate blocked: %d component(s) need repair", len(blockers))
	}

	migrated, hasNodes, empty, strictComponents := 0, 0, 0, 0
	var emptyIDs []string
	var dirtyIDs []string

	for _, e := range entities {
		nodes, err := opts.Store.NodesForEntity(e.ID)
		if err != nil {
			return fmt.Errorf("checking nodes for %s: %w", e.ID, err)
		}
		if len(nodes) > 0 {
			hasNodes++
			if e.Type == "component" {
				body, err := content.ReadEntity(opts.Store, e.ID)
				if err != nil {
					return fmt.Errorf("reading component body %s: %w", e.ID, err)
				}
				if len(validateStrictComponentDoc(body, "error")) > 0 {
					strictBody := strictMigrationComponentBody(opts.Store, e, body)
					if issues := validateStrictComponentDoc(strictBody, "error"); len(issues) > 0 {
						return formatValidationError(e.ID, issues)
					}
					if opts.DryRun {
						fmt.Fprintf(w, "  will strict-migrate component: %s (%s)\n", e.ID, e.Title)
					} else {
						if err := content.WriteEntity(opts.Store, e.ID, strictBody); err != nil {
							return fmt.Errorf("strict-migrating component %s: %w", e.ID, err)
						}
						fmt.Fprintf(w, "  strict-migrated component: %s\n", e.ID)
					}
					strictComponents++
				}
			}
			continue
		}

		body := opts.Store.LegacyBody(e.ID)
		if body == "" {
			if e.Type == "component" {
				body = strictMigrationComponentBody(opts.Store, e, "")
				if issues := validateStrictComponentDoc(body, "error"); len(issues) > 0 {
					return formatValidationError(e.ID, issues)
				}
				if opts.DryRun {
					fmt.Fprintf(w, "  will strict-migrate empty component: %s (%s)\n", e.ID, e.Title)
				} else {
					if err := content.WriteEntity(opts.Store, e.ID, body); err != nil {
						return fmt.Errorf("strict-migrating empty component %s: %w", e.ID, err)
					}
					fmt.Fprintf(w, "  strict-migrated empty component: %s\n", e.ID)
				}
				migrated++
				strictComponents++
				continue
			}
			if body = defaultMigrationBody(opts.Store, e); body != "" {
				if opts.DryRun {
					fmt.Fprintf(w, "  will recover empty %s: %s (%s)\n", e.Type, e.ID, e.Title)
				} else {
					if err := content.WriteEntity(opts.Store, e.ID, body); err != nil {
						return fmt.Errorf("recovering empty %s %s: %w", e.Type, e.ID, err)
					}
					fmt.Fprintf(w, "  recovered empty %s: %s\n", e.Type, e.ID)
				}
				migrated++
				continue
			}
			empty++
			emptyIDs = append(emptyIDs, e.ID)
			continue
		}

		if hasStaleFrontmatter(body) {
			dirtyIDs = append(dirtyIDs, e.ID)
		}

		if opts.DryRun {
			fmt.Fprintf(w, "  will migrate: %s (%s)\n", e.ID, e.Title)
			migrated++
			continue
		}

		if e.Type == "component" {
			body = strictMigrationComponentBody(opts.Store, e, body)
			if issues := validateStrictComponentDoc(body, "error"); len(issues) > 0 {
				return formatValidationError(e.ID, issues)
			}
			strictComponents++
		}

		if err := content.WriteEntity(opts.Store, e.ID, body); err != nil {
			writeMigrateWriteFailure(w, e.ID, migrated, err)
			return fmt.Errorf("migrate write failed at %s: %w", e.ID, err)
		}
		fmt.Fprintf(w, "  migrated: %s\n", e.ID)
		migrated++
	}

	fmt.Fprintln(w)
	if opts.DryRun {
		fmt.Fprintf(w, "dry-run: %d to migrate", migrated)
	} else {
		fmt.Fprintf(w, "%d migrated", migrated)
	}
	if hasNodes > 0 {
		fmt.Fprintf(w, ", %d already have nodes (ok)", hasNodes)
	}
	if strictComponents > 0 {
		fmt.Fprintf(w, ", %d strict component docs", strictComponents)
	}
	fmt.Fprintln(w)

	if len(dirtyIDs) > 0 {
		fmt.Fprintf(w, "\nWARNING: %d entities had stale frontmatter in body (auto-stripped during migration).\n", len(dirtyIDs))
		fmt.Fprintln(w, "Review and rewrite with accurate content:")
		for _, id := range dirtyIDs {
			fmt.Fprintf(w, "  c3x read %s        # review current content\n", id)
			fmt.Fprintf(w, "  c3x write %s       # pipe corrected markdown\n", id)
		}
	}

	if empty > 0 {
		fmt.Fprintf(w, "\n%d entities have no content yet:\n", empty)
		for _, id := range emptyIDs {
			fmt.Fprintf(w, "  c3x write %s\n", id)
		}
	}

	return nil
}

func writeMigrateWriteFailure(w io.Writer, id string, migrated int, err error) {
	fmt.Fprintf(w, "\nBLOCKED: migration write failed at %s after %d successful write(s).\n", id, migrated)
	fmt.Fprintf(w, "Reason: %v\n", err)
	fmt.Fprintln(w, "Why: C3 stopped before canonical export, so submitted .c3/ markdown is not rewritten from a partial cache.")
	fmt.Fprintln(w, "Fix loop:")
	fmt.Fprintln(w, "  1. Fix the write/database error above.")
	fmt.Fprintln(w, "  2. Remove .c3/c3.db* and .c3/.c3.import.tmp.db*.")
	fmt.Fprintln(w, "  3. Run: c3x import --force")
	fmt.Fprintln(w, "  4. Run: c3x migrate")
	fmt.Fprintln(w, "  5. Run: c3x check --include-adr && c3x verify")
}

func collectMigrateBlockers(s *store.Store, entities []*store.Entity) ([]migrateBlocker, error) {
	var blockers []migrateBlocker
	for _, e := range entities {
		if e.Type != "component" {
			continue
		}
		nodes, err := s.NodesForEntity(e.ID)
		if err != nil {
			return nil, fmt.Errorf("preflight failed while checking nodes for %s: %w\nhint: remove .c3/c3.db* and .c3/.c3.import.tmp.db*, then run c3x import --force && c3x migrate", e.ID, err)
		}
		var body string
		if len(nodes) > 0 {
			body, err = content.ReadEntity(s, e.ID)
			if err != nil {
				return nil, fmt.Errorf("preflight failed while reading component body %s: %w\nhint: remove .c3/c3.db* and .c3/.c3.import.tmp.db*, then run c3x import --force && c3x migrate", e.ID, err)
			}
			if len(validateStrictComponentDoc(body, "error")) == 0 {
				continue
			}
		} else {
			body = s.LegacyBody(e.ID)
		}
		strictBody := strictMigrationComponentBody(s, e, body)
		if issues := validateStrictComponentDoc(strictBody, "error"); len(issues) > 0 {
			blockers = append(blockers, migrateBlocker{ID: e.ID, Title: e.Title, Issues: issues})
		}
	}
	return blockers, nil
}

func writeMigrateBlockerPlan(w io.Writer, blockers []migrateBlocker) {
	fmt.Fprintf(w, "\nBLOCKED: %d component(s) need repair before migration can finish.\n", len(blockers))
	fmt.Fprintln(w, "Why: strict component docs are all-or-nothing; C3 made no migration writes.")
	fmt.Fprintln(w, "Fix loop:")
	fmt.Fprintln(w, "  1. Edit the listed component docs in the copied/sandbox tree.")
	fmt.Fprintln(w, "  2. Remove .c3/c3.db* and .c3/.c3.import.tmp.db*.")
	fmt.Fprintln(w, "  3. Run: c3x import --force")
	fmt.Fprintln(w, "  4. Run: c3x migrate")
	fmt.Fprintln(w, "  5. Run: c3x check --include-adr && c3x verify")
	for _, blocker := range blockers {
		fmt.Fprintf(w, "\n%s %s\n", blocker.ID, blocker.Title)
		for _, issue := range blocker.Issues {
			fmt.Fprintf(w, "  - %s\n", issue.Message)
			fmt.Fprintf(w, "    fix: update %s so generated strict sections pass validation.\n", blocker.ID)
			if strings.Contains(issue.Message, "placeholder language") {
				fmt.Fprintln(w, "    common rewrite: optional->secondary, todo->task, later->future, TBD->specified decision.")
			}
		}
		fmt.Fprintf(w, "    inspect: c3x read %s --full\n", blocker.ID)
	}
}

func strictMigrationComponentBody(s *store.Store, e *store.Entity, oldBody string) string {
	title := nonEmpty(e.Title, e.Slug, e.ID)
	goal := firstGoal(e.Goal, oldBody, fmt.Sprintf("Document %s behavior within its parent container.", title))
	parent := nonEmpty(e.ParentID, "N.A - no parent container recorded")
	fallbackRef := parent
	if isNAReason(fallbackRef) {
		fallbackRef = e.ID
	}
	ref := firstGovernanceReference(s, e.ID, fallbackRef)
	refType := governanceType(ref)
	if strings.HasPrefix(ref, "N.A - ") {
		refType = "N.A - no reference type recorded"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", title)
	fmt.Fprintf(&b, "## Goal\n\n%s\n\n", goal)
	fmt.Fprintf(&b, "## Parent Fit\n\n")
	b.WriteString("| Field | Value |\n|-------|-------|\n")
	fmt.Fprintf(&b, "| Parent | %s |\n", parent)
	fmt.Fprintf(&b, "| Role | Own %s behavior inside the parent container without taking over sibling responsibilities. |\n", title)
	fmt.Fprintf(&b, "| Boundary | Keep %s decisions inside this component and escalate container-wide policy to the parent. |\n", title)
	fmt.Fprintf(&b, "| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |\n\n")
	fmt.Fprintf(&b, "## Purpose\n\n")
	fmt.Fprintf(&b, "Provide durable agent-ready documentation for %s so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.\n\n", title)
	fmt.Fprintf(&b, "## Foundational Flow\n\n")
	b.WriteString("| Aspect | Detail | Reference |\n|--------|--------|-----------|\n")
	fmt.Fprintf(&b, "| Preconditions | Parent container context is loaded before %s behavior is changed. | %s |\n", title, ref)
	fmt.Fprintf(&b, "| Inputs | Accept only the files, commands, data, or calls that belong to %s ownership. | %s |\n", title, ref)
	fmt.Fprintf(&b, "| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | %s |\n", ref)
	fmt.Fprintf(&b, "| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | %s |\n\n", ref)
	fmt.Fprintf(&b, "## Business Flow\n\n")
	b.WriteString("| Aspect | Detail | Reference |\n|--------|--------|-----------|\n")
	fmt.Fprintf(&b, "| Actor / caller | Agent, command, or workflow asks %s to deliver its documented responsibility. | %s |\n", title, ref)
	fmt.Fprintf(&b, "| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | %s |\n", ref)
	fmt.Fprintf(&b, "| Alternate paths | When a request falls outside %s ownership, hand it to the parent or sibling component. | %s |\n", title, ref)
	fmt.Fprintf(&b, "| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | %s |\n\n", ref)
	fmt.Fprintf(&b, "## Governance\n\n")
	b.WriteString("| Reference | Type | Governs | Precedence | Notes |\n|-----------|------|---------|------------|-------|\n")
	fmt.Fprintf(&b, "| %s | %s | Governs %s behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |\n\n", ref, refType, title)
	fmt.Fprintf(&b, "## Contract\n\n")
	b.WriteString("| Surface | Direction | Contract | Boundary | Evidence |\n|---------|-----------|----------|----------|----------|\n")
	fmt.Fprintf(&b, "| %s input | IN | Callers must provide context that matches the component goal and parent fit. | %s boundary | c3x lookup plus targeted tests or review. |\n", title, parent)
	fmt.Fprintf(&b, "| %s output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | %s boundary | c3x check and project test suite. |\n\n", title, parent)
	fmt.Fprintf(&b, "## Change Safety\n\n")
	b.WriteString("| Risk | Trigger | Detection | Required Verification |\n|------|---------|-----------|-----------------------|\n")
	fmt.Fprintf(&b, "| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |\n")
	fmt.Fprintf(&b, "| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |\n\n")
	fmt.Fprintf(&b, "## Derived Materials\n\n")
	b.WriteString("| Material | Must derive from | Allowed variance | Evidence |\n|----------|------------------|------------------|----------|\n")
	fmt.Fprintf(&b, "| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |\n")
	return b.String()
}

func defaultMigrationBody(s *store.Store, e *store.Entity) string {
	title := nonEmpty(e.Title, e.Slug, e.ID)
	goal := nonEmpty(e.Goal, fmt.Sprintf("Document %s architecture decisions and responsibilities.", title))
	switch e.Type {
	case "system":
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n## Goal\n\n%s\n\n", title, goal)
		b.WriteString("## Containers\n\n| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |\n|----|------|----------|--------|------------------|-------------------|\n")
		for _, child := range mustChildren(s, e.ID) {
			if child.Type == "container" {
				fmt.Fprintf(&b, "| %s | %s | %s | %s | Own %s architecture area. | Supports the system goal through %s. |\n", child.ID, child.Title, nonEmpty(child.Boundary, "documented"), nonEmpty(child.Status, "active"), child.Title, child.Title)
			}
		}
		b.WriteString("\n## Abstract Constraints\n\n| Constraint | Rationale | Affected Containers |\n|------------|-----------|---------------------|\n| Use C3 CLI as canonical mutation path. | Preserves sealed docs, cache rebuilds, and verification evidence. | c3-1, c3-2 |\n")
		return b.String()
	case "container":
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n## Goal\n\n%s\n\n", title, goal)
		b.WriteString("## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n|----|------|----------|--------|-------------------|\n")
		for _, child := range mustChildren(s, e.ID) {
			if child.Type == "component" {
				fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", child.ID, child.Title, nonEmpty(child.Category, "foundation"), nonEmpty(child.Status, "active"), nonEmpty(child.Goal, fmt.Sprintf("Supports %s.", title)))
			}
		}
		fmt.Fprintf(&b, "\n## Responsibilities\n\nOwn %s responsibilities, keep child components aligned with the container goal, and preserve top-down C3 verification evidence.\n", title)
		return b.String()
	case "ref":
		return fmt.Sprintf("# %s\n\n## Goal\n\n%s\n\n## Choice\n\nKeep this reference as the cited source of truth for components that depend on %s.\n\n## Why\n\nCentralizing this guidance prevents component docs from duplicating shared policy and keeps future changes reviewable.\n", title, goal, title)
	case "adr":
		return fmt.Sprintf("# %s\n\n## Goal\n\n%s\n", title, goal)
	default:
		return ""
	}
}

func mustChildren(s *store.Store, id string) []*store.Entity {
	children, err := s.Children(id)
	if err != nil {
		return nil
	}
	return children
}

func firstGoal(entityGoal, body, fallback string) string {
	if strings.TrimSpace(entityGoal) != "" {
		return substantiveGoal(strings.TrimSpace(entityGoal), fallback)
	}
	for _, section := range markdown.ParseSections(body) {
		if section.Name == "Goal" && strings.TrimSpace(section.Content) != "" {
			return substantiveGoal(strings.TrimSpace(section.Content), fallback)
		}
	}
	return fallback
}

func substantiveGoal(goal, fallback string) string {
	if len(strings.Fields(goal)) >= 4 {
		return goal
	}
	return fallback
}

func firstGovernanceReference(s *store.Store, id string, fallback string) string {
	rels, err := s.RelationshipsFrom(id)
	if err != nil {
		return fallback
	}
	for _, rel := range rels {
		if strings.HasPrefix(rel.ToID, "ref-") || strings.HasPrefix(rel.ToID, "rule-") || strings.HasPrefix(rel.ToID, "adr-") {
			return rel.ToID
		}
	}
	return fallback
}

func governanceType(ref string) string {
	switch {
	case strings.HasPrefix(ref, "ref-"):
		return "ref"
	case strings.HasPrefix(ref, "rule-"):
		return "rule"
	case strings.HasPrefix(ref, "adr-"):
		return "adr"
	case strings.HasPrefix(ref, "c3-"):
		return "policy"
	default:
		return "policy"
	}
}

func nonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func hasStaleFrontmatter(body string) bool {
	if strings.HasPrefix(body, "---\n") {
		return true
	}
	lines := strings.SplitN(body, "\n", 5)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			return false
		}
		idx := strings.Index(l, ":")
		if idx > 0 && !strings.Contains(l[:idx], " ") {
			return true
		}
	}
	return false
}
