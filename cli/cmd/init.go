package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// RunInitDB scaffolds a new .c3/ directory with a SQLite database.
func RunInitDB(c3Dir string, projectName string, w io.Writer) error {
	if info, err := os.Stat(c3Dir); err == nil && info.IsDir() {
		return fmt.Errorf("error: %s already exists", c3Dir)
	}

	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		return fmt.Errorf("error: creating %s: %w", c3Dir, err)
	}
	defs, err := schema.AllDefinitions("")
	if err != nil {
		return fmt.Errorf("error: loading built-in canvas definitions: %w", err)
	}
	writtenDefs, err := MaterializeDefinitions(c3Dir, defs)
	if err != nil {
		return fmt.Errorf("error: materializing canvas definitions: %w", err)
	}

	// Open store
	dbPath := filepath.Join(c3Dir, "c3.db")
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: creating database: %w", err)
	}
	defer s.Close()

	// Insert c3-0 context entity
	if err := s.InsertEntity(&store.Entity{
		ID:       "c3-0",
		Type:     "system",
		Title:    projectName,
		Slug:     "",
		Goal:     "",
		Status:   "active",
		Metadata: "{}",
	}); err != nil {
		return fmt.Errorf("error: inserting context entity: %w", err)
	}

	// Insert adoption ADR entity
	if err := s.InsertEntity(&store.Entity{
		ID:       "adr-00000000-c3-adoption",
		Type:     "adr",
		Title:    "C3 Architecture Documentation Adoption",
		Slug:     "c3-adoption",
		Status:   "proposed",
		// No Date: the genesis id is the adr-00000000 sentinel, so the exported file
		// must be adr-00000000-c3-adoption.md (matching the id commands key off), not
		// a date-stamped name that read/check can't resolve.
		Metadata: "{}",
	}); err != nil {
		return fmt.Errorf("error: inserting ADR entity: %w", err)
	}

	// ADR affects c3-0
	if err := s.AddRelationship(&store.Relationship{
		FromID:  "adr-00000000-c3-adoption",
		ToID:    "c3-0",
		RelType: "affects",
	}); err != nil {
		return fmt.Errorf("error: adding ADR relationship: %w", err)
	}

	fmt.Fprintln(w, "Created .c3/")
	fmt.Fprintln(w, "  └── c3.db (system: c3-0, adr: adr-00000000-c3-adoption)")
	if len(writtenDefs) > 0 {
		fmt.Fprintf(w, "  └── canvases/ (%d definitions)\n", len(writtenDefs))
	}

	return nil
}
