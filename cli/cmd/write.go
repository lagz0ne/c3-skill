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

	// Auto-promote goal from body if writing the Goal section
	if opts.Section == "Goal" && existing.Goal == "" {
		line := strings.SplitN(strings.TrimSpace(opts.Content), "\n", 2)[0]
		existing.Goal = strings.TrimSpace(line)
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

type validationIssue struct {
	severity string
	message  string
	hint     string
}

func formatValidationError(id string, issues []validationIssue) error {
	var sb strings.Builder
	for _, issue := range issues {
		fmt.Fprintf(&sb, "  %s: %s\n", issue.severity, issue.message)
		if issue.hint != "" {
			fmt.Fprintf(&sb, "    hint: %s\n", issue.hint)
		}
	}
	return fmt.Errorf("error: content validation failed for %s\n%s", id, sb.String())
}

// validateContent checks required schema sections and goal presence.
func validateContent(e *store.Entity) []validationIssue {
	var issues []validationIssue

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

	for _, sec := range schemaSections {
		if !sec.Required {
			continue
		}

		content, exists := sectionMap[sec.Name]
		if !exists {
			issues = append(issues, validationIssue{
				severity: "error",
				message:  fmt.Sprintf("missing required section: %s", sec.Name),
				hint:     fmt.Sprintf("add ## %s — %s", sec.Name, sec.Purpose),
			})
			continue
		}

		if content == "" {
			issues = append(issues, validationIssue{
				severity: "error",
				message:  fmt.Sprintf("empty required section: %s", sec.Name),
				hint:     fmt.Sprintf("add content to ## %s — %s", sec.Name, sec.Purpose),
			})
		}
	}

	// Auto-promote goal from body if frontmatter goal is empty
	if e.Goal == "" {
		if bodyGoal, ok := sectionMap["Goal"]; ok && bodyGoal != "" {
			line := strings.SplitN(bodyGoal, "\n", 2)[0]
			e.Goal = strings.TrimSpace(line)
		}
	}

	return issues
}

// syncRelationships updates uses/affects/scope relationships from frontmatter.
func syncRelationships(s *store.Store, entityID string, fm *frontmatter.Frontmatter) error {
	for _, relType := range []struct {
		targets []string
		name    string
	}{
		{fm.Refs, "uses"},
		{fm.Affects, "affects"},
		{fm.Scope, "scope"},
		{fm.Sources, "sources"},
		{fm.Origin, "origin"},
	} {
		for _, target := range relType.targets {
			target = frontmatter.StripAnchor(target)
			if target == "" {
				continue
			}
			_ = s.AddRelationship(&store.Relationship{
				FromID:  entityID,
				ToID:    target,
				RelType: relType.name,
			})
		}
	}
	return nil
}
