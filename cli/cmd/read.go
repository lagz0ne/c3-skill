package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ReadOptions holds parameters for the read command.
type ReadOptions struct {
	Store   *store.Store
	ID      string
	JSON    bool
	Section string
}

// ReadResult is the JSON output for read.
type ReadResult struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Goal        string   `json:"goal,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Category    string   `json:"category,omitempty"`
	ParentID    string   `json:"parent,omitempty"`
	Boundary    string   `json:"boundary,omitempty"`
	Date        string   `json:"date,omitempty"`
	Uses        []string `json:"uses,omitempty"`
	Affects     []string `json:"affects,omitempty"`
	Scope       []string `json:"scope,omitempty"`
	Body        string   `json:"body"`
}

// ReadSectionResult is the JSON output for read --section.
type ReadSectionResult struct {
	Section string `json:"section"`
	Content string `json:"content"`
}

// RunRead outputs the full content of a single entity.
func RunRead(opts ReadOptions, w io.Writer) error {
	if opts.ID == "" {
		return fmt.Errorf("error: usage: c3x read <entity-id>\nhint: c3x list to see all entities")
	}

	entity, err := opts.Store.GetEntity(opts.ID)
	if err != nil {
		return fmt.Errorf("error: entity %q not found", opts.ID)
	}

	if opts.Section != "" {
		sections := markdown.ParseSections(entity.Body)
		for _, s := range sections {
			if s.Name == opts.Section {
				if opts.JSON {
					return writeJSON(w, ReadSectionResult{Section: s.Name, Content: strings.TrimSpace(s.Content)})
				}
				fmt.Fprintln(w, strings.TrimSpace(s.Content))
				return nil
			}
		}
		return fmt.Errorf("error: section %q not found in %s\nhint: available sections: %s",
			opts.Section, opts.ID, readAvailableSections(entity.Body))
	}

	if opts.JSON {
		result := ReadResult{
			ID:          entity.ID,
			Type:        entity.Type,
			Title:       entity.Title,
			Goal:        entity.Goal,
			Summary:     entity.Summary,
			Description: entity.Description,
			Status:      entity.Status,
			Category:    entity.Category,
			ParentID:    entity.ParentID,
			Boundary:    entity.Boundary,
			Date:        entity.Date,
			Body:        entity.Body,
		}

		rels, _ := opts.Store.RelationshipsFrom(entity.ID)
		for _, r := range rels {
			switch r.RelType {
			case "uses":
				result.Uses = append(result.Uses, r.ToID)
			case "affects":
				result.Affects = append(result.Affects, r.ToID)
			case "scope":
				result.Scope = append(result.Scope, r.ToID)
			}
		}

		return writeJSON(w, result)
	}

	// Default: output as markdown (same format as export)
	fmt.Fprint(w, buildExportContent(opts.Store, entity))
	return nil
}

func readAvailableSections(body string) string {
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
