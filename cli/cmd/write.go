package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// WriteOptions holds parameters for the write command.
type WriteOptions struct {
	Store   *store.Store
	ID      string
	Section string // if set, only replace this section's content
	Content string // full markdown (no --section) or section body (with --section)
}

// RunWrite replaces an entity's content (or a single section) with validation.
func RunWrite(opts WriteOptions, w io.Writer) error {
	if opts.ID == "" {
		return fmt.Errorf("error: usage: c3x write <entity-id> [--section <name>] < content\nhint: pipe content via stdin")
	}

	existing, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("error: entity %q not found", opts.ID)
	}

	if opts.Section != "" {
		return runWriteSection(existing, opts, w)
	}
	return runWriteFull(existing, opts, w)
}

// runWriteFull replaces the entire entity body + frontmatter fields.
func runWriteFull(existing *store.Entity, opts WriteOptions, w io.Writer) error {
	fm, body := frontmatter.ParseFrontmatter(opts.Content)

	applyFrontmatter(existing, fm)

	if fm != nil {
		existing.Body = body
	} else {
		existing.Body = opts.Content
	}

	promoteGoalIfEmpty(existing)

	issues := validateContent(existing)
	if len(issues) > 0 {
		return formatValidationError(opts.ID, issues)
	}

	if err := opts.Store.UpdateEntity(existing); err != nil {
		return fmt.Errorf("error: updating entity: %w", err)
	}

	if fm != nil {
		if err := syncRelationships(opts.Store, opts.ID, fm); err != nil {
			fmt.Fprintf(w, "warning: relationship sync: %v\n", err)
		}
	}

	fmt.Fprintf(w, "Updated %s (%s)\n", opts.ID, existing.Type)
	return nil
}

// runWriteSection replaces a single section's content, validates the result.
func runWriteSection(existing *store.Entity, opts WriteOptions, w io.Writer) error {
	newBody, err := markdown.ReplaceSection(existing.Body, opts.Section, opts.Content)
	if err != nil {
		return fmt.Errorf("error: section %q not found in %s\nhint: available sections: %s",
			opts.Section, opts.ID, availableSections(existing.Body))
	}

	existing.Body = newBody

	if opts.Section == "Goal" {
		promoteGoalIfEmpty(existing)
	}

	issues := validateContent(existing)
	if len(issues) > 0 {
		return formatValidationError(opts.ID, issues)
	}

	if err := opts.Store.UpdateEntity(existing); err != nil {
		return fmt.Errorf("error: updating entity: %w", err)
	}

	fmt.Fprintf(w, "Updated %s section %q\n", opts.ID, opts.Section)
	return nil
}

func availableSections(body string) string {
	sections := markdown.ParseSections(body)
	var names []string
	for _, s := range sections {
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}
	if len(names) == 0 {
		return "(none)"
	}
	return strings.Join(names, ", ")
}

func applyFrontmatter(e *store.Entity, fm *frontmatter.Frontmatter) {
	if fm == nil {
		return
	}
	if fm.Goal != "" {
		e.Goal = fm.Goal
	}
	if fm.Summary != "" {
		e.Summary = fm.Summary
	}
	if fm.Description != "" {
		e.Description = fm.Description
	}
	if fm.Status != "" {
		e.Status = fm.Status
	}
	if fm.Boundary != "" {
		e.Boundary = fm.Boundary
	}
	if fm.Category != "" {
		e.Category = fm.Category
	}
	if fm.Title != "" {
		e.Title = fm.Title
	}
	if fm.Date != "" {
		e.Date = fm.Date
	}
}

// promoteGoalIfEmpty extracts the first line of ## Goal body section
// into entity.Goal if the frontmatter goal is empty.
func promoteGoalIfEmpty(e *store.Entity) {
	if e.Goal != "" {
		return
	}
	sections := markdown.ParseSections(e.Body)
	for _, s := range sections {
		if s.Name == "Goal" {
			content := strings.TrimSpace(s.Content)
			if content != "" {
				e.Goal = strings.SplitN(content, "\n", 2)[0]
			}
			return
		}
	}
}

func formatValidationError(id string, issues []Issue) error {
	var sb strings.Builder
	for _, issue := range issues {
		fmt.Fprintf(&sb, "  %s: %s\n", issue.Severity, issue.Message)
		if issue.Hint != "" {
			fmt.Fprintf(&sb, "    hint: %s\n", issue.Hint)
		}
	}
	return fmt.Errorf("error: content validation failed for %s\n%s", id, sb.String())
}

// validateContent checks required schema sections. Pure validation — no mutations.
func validateContent(e *store.Entity) []Issue {
	schemaSections := schema.ForType(e.Type)
	if schemaSections == nil {
		return nil
	}

	sections := markdown.ParseSections(e.Body)
	sectionMap := make(map[string]string)
	for _, s := range sections {
		if s.Name != "" {
			sectionMap[s.Name] = strings.TrimSpace(s.Content)
		}
	}

	var issues []Issue
	for _, sec := range schemaSections {
		if !sec.Required {
			continue
		}

		content, exists := sectionMap[sec.Name]
		if !exists {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  fmt.Sprintf("missing required section: %s", sec.Name),
				Hint:     fmt.Sprintf("add ## %s — %s", sec.Name, sec.Purpose),
			})
			continue
		}

		if content == "" {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  fmt.Sprintf("empty required section: %s", sec.Name),
				Hint:     fmt.Sprintf("add content to ## %s — %s", sec.Name, sec.Purpose),
			})
		}
	}

	return issues
}

func syncRelationships(s *store.Store, entityID string, fm *frontmatter.Frontmatter) error {
	existing, _ := s.RelationshipsFrom(entityID)

	type relKey struct{ toID, relType string }
	desired := make(map[relKey]bool)
	for _, rt := range []struct {
		targets []string
		name    string
		strip   bool
	}{
		{fm.Refs, "uses", false},
		{fm.Affects, "affects", false},
		{fm.Scope, "scope", true},
		{fm.Sources, "sources", true},
		{fm.Origin, "origin", true},
	} {
		for _, target := range rt.targets {
			if rt.strip {
				target = frontmatter.StripAnchor(target)
			}
			if target == "" {
				continue
			}
			desired[relKey{target, rt.name}] = true
		}
	}

	var errs []string
	for _, r := range existing {
		if !desired[relKey{r.ToID, r.RelType}] {
			if err := s.RemoveRelationship(r); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	for key := range desired {
		if err := s.AddRelationship(&store.Relationship{
			FromID: entityID, ToID: key.toID, RelType: key.relType,
		}); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}
