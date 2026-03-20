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
	Content string // full markdown with frontmatter
}

// RunWrite replaces an entity's content with validation.
// Parses the input as frontmatter + body, validates required sections,
// then updates the entity in the store.
func RunWrite(opts WriteOptions, w io.Writer) error {
	if opts.ID == "" {
		return fmt.Errorf("error: usage: c3x write <entity-id> < content.md\nhint: pipe markdown content via stdin")
	}

	existing, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("error: entity %q not found", opts.ID)
	}

	// Parse the incoming content
	fm, body := frontmatter.ParseFrontmatter(opts.Content)

	// Update structured fields from frontmatter if provided
	if fm != nil {
		if fm.Goal != "" {
			existing.Goal = fm.Goal
		}
		if fm.Summary != "" {
			existing.Summary = fm.Summary
		}
		if fm.Description != "" {
			existing.Description = fm.Description
		}
		if fm.Status != "" {
			existing.Status = fm.Status
		}
		if fm.Boundary != "" {
			existing.Boundary = fm.Boundary
		}
		if fm.Category != "" {
			existing.Category = fm.Category
		}
		if fm.Title != "" {
			existing.Title = fm.Title
		}
		if fm.Date != "" {
			existing.Date = fm.Date
		}
	}

	// Use body from content (if frontmatter was present, body is after ---)
	// If no frontmatter delimiters, treat entire content as body
	if fm != nil {
		existing.Body = body
	} else {
		existing.Body = opts.Content
	}

	// Validate content against schema before accepting
	issues := validateContent(existing)
	if len(issues) > 0 {
		var sb strings.Builder
		for _, issue := range issues {
			fmt.Fprintf(&sb, "  %s: %s\n", issue.severity, issue.message)
			if issue.hint != "" {
				fmt.Fprintf(&sb, "    hint: %s\n", issue.hint)
			}
		}
		return fmt.Errorf("error: content validation failed for %s\n%s", opts.ID, sb.String())
	}

	if err := opts.Store.UpdateEntity(existing); err != nil {
		return fmt.Errorf("error: updating entity: %w", err)
	}

	// Sync relationships from frontmatter
	if fm != nil {
		if err := syncRelationships(opts.Store, opts.ID, fm); err != nil {
			fmt.Fprintf(w, "warning: relationship sync: %v\n", err)
		}
	}

	fmt.Fprintf(w, "Updated %s (%s)\n", opts.ID, existing.Type)
	return nil
}

type validationIssue struct {
	severity string // "error" or "warning"
	message  string
	hint     string
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

	// Check goal: if body has ## Goal content, frontmatter goal should match
	if e.Goal == "" {
		if bodyGoal, ok := sectionMap["Goal"]; ok && bodyGoal != "" {
			// Auto-promote: extract first line of body Goal to entity.Goal
			line := strings.SplitN(bodyGoal, "\n", 2)[0]
			e.Goal = strings.TrimSpace(line)
		}
	}

	return issues
}

// syncRelationships updates uses/affects/scope relationships from frontmatter.
func syncRelationships(s *store.Store, entityID string, fm *frontmatter.Frontmatter) error {
	// Sync uses
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
