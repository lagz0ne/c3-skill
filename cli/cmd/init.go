package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/templates"
)

// RunInit scaffolds a new .c3/ directory structure.
func RunInit(projectRoot string, w io.Writer) error {
	dotC3 := filepath.Join(projectRoot, ".c3")

	if info, err := os.Stat(dotC3); err == nil && info.IsDir() {
		return fmt.Errorf("error: .c3/ directory already exists")
	}

	today := time.Now().Format("20060102")
	projectName := filepath.Base(projectRoot)

	// Create directory structure
	for _, dir := range []string{dotC3, filepath.Join(dotC3, "refs"), filepath.Join(dotC3, "rules"), filepath.Join(dotC3, "adr")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error: creating %s: %w", dir, err)
		}
	}

	// config.yaml
	if err := os.WriteFile(filepath.Join(dotC3, "config.yaml"), []byte("# C3 configuration\n"), 0644); err != nil {
		return fmt.Errorf("error: writing config.yaml: %w", err)
	}

	// README.md from context template
	contextContent, err := templates.Read("context.md")
	if err != nil {
		return fmt.Errorf("error: reading context template: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dotC3, "README.md"), []byte(contextContent), 0644); err != nil {
		return fmt.Errorf("error: writing README.md: %w", err)
	}

	// ADR: adr-00000000-c3-adoption.md
	adrContent, err := templates.Render("adr-000.md", map[string]string{
		"${DATE}":    today,
		"${PROJECT}": projectName,
	})
	if err != nil {
		return fmt.Errorf("error: reading ADR template: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dotC3, "adr", "adr-00000000-c3-adoption.md"), []byte(adrContent), 0644); err != nil {
		return fmt.Errorf("error: writing ADR: %w", err)
	}

	fmt.Fprintln(w, "Created .c3/")
	fmt.Fprintln(w, "  ├── config.yaml")
	fmt.Fprintln(w, "  ├── README.md")
	fmt.Fprintln(w, "  ├── refs/")
	fmt.Fprintln(w, "  ├── rules/")
	fmt.Fprintln(w, "  └── adr/")
	fmt.Fprintln(w, "      └── adr-00000000-c3-adoption.md")

	return nil
}

// RunInitDB scaffolds a new .c3/ directory with a SQLite database instead of markdown files.
func RunInitDB(c3Dir string, projectName string, w io.Writer) error {
	if info, err := os.Stat(c3Dir); err == nil && info.IsDir() {
		return fmt.Errorf("error: %s already exists", c3Dir)
	}

	if err := os.MkdirAll(c3Dir, 0755); err != nil {
		return fmt.Errorf("error: creating %s: %w", c3Dir, err)
	}

	today := time.Now().Format("20060102")

	// Open store
	dbPath := filepath.Join(c3Dir, "c3.db")
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: creating database: %w", err)
	}
	defer s.Close()

	// Insert c3-0 context entity
	contextBody, err := templates.Render("context.md", map[string]string{
		"${PROJECT}": projectName,
		"${GOAL}":    "",
		"${SUMMARY}": "",
	})
	if err != nil {
		return fmt.Errorf("error: rendering context template: %w", err)
	}
	if err := s.InsertEntity(&store.Entity{
		ID:       "c3-0",
		Type:     "system",
		Title:    projectName,
		Slug:     "",
		Goal:     "",
		Summary:  "",
		Body:     contextBody,
		Status:   "active",
		Metadata: "{}",
	}); err != nil {
		return fmt.Errorf("error: inserting context entity: %w", err)
	}

	// Insert adoption ADR entity
	adrBody, err := templates.Render("adr-000.md", map[string]string{
		"${DATE}":    today,
		"${PROJECT}": projectName,
	})
	if err != nil {
		return fmt.Errorf("error: rendering ADR template: %w", err)
	}
	if err := s.InsertEntity(&store.Entity{
		ID:       "adr-00000000-c3-adoption",
		Type:     "adr",
		Title:    "C3 Architecture Documentation Adoption",
		Slug:     "c3-adoption",
		Status:   "in-progress",
		Date:     today,
		Body:     adrBody,
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

	// config.yaml
	if err := os.WriteFile(filepath.Join(c3Dir, "config.yaml"), []byte("# C3 configuration\n"), 0644); err != nil {
		return fmt.Errorf("error: writing config.yaml: %w", err)
	}

	fmt.Fprintln(w, "Created .c3/")
	fmt.Fprintln(w, "  ├── config.yaml")
	fmt.Fprintln(w, "  └── c3.db (system: c3-0, adr: adr-00000000-c3-adoption)")

	return nil
}
