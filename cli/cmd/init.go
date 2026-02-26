package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

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
	for _, dir := range []string{dotC3, filepath.Join(dotC3, "refs"), filepath.Join(dotC3, "adr")} {
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
	fmt.Fprintln(w, "  └── adr/")
	fmt.Fprintln(w, "      └── adr-00000000-c3-adoption.md")

	return nil
}
