package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/numbering"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
	"github.com/lagz0ne/c3-design/cli/internal/wiring"
)

// AddOptions holds parameters for the richer add command.
type AddOptions struct {
	EntityType string
	Slug       string
	C3Dir      string
	Graph      *walker.C3Graph
	Container  string
	Feature    bool
	Goal       string
	Summary    string
	Boundary   string
}

func (o *AddOptions) hasContent() bool {
	return o.Goal != "" || o.Summary != "" || o.Boundary != ""
}

// RunAddRich creates a new entity with optional content pre-populated.
func RunAddRich(opts AddOptions, w io.Writer) error {
	if !opts.hasContent() {
		return RunAdd(opts.EntityType, opts.Slug, opts.C3Dir, opts.Graph, opts.Container, opts.Feature, w)
	}

	switch opts.EntityType {
	case "container":
		return addRichContainer(opts, w)
	case "component":
		return addRichComponent(opts, w)
	case "ref":
		return addRichRef(opts, w)
	case "adr":
		return addRichAdr(opts, w)
	case "recipe":
		return addRichRecipe(opts, w)
	default:
		return fmt.Errorf("error: unknown entity type '%s'\nhint: types: container, component, ref, adr, recipe", opts.EntityType)
	}
}

func addRichContainer(opts AddOptions, w io.Writer) error {
	n := numbering.NextContainerId(opts.Graph)
	dirName := fmt.Sprintf("c3-%d-%s", n, opts.Slug)
	dirPath := filepath.Join(opts.C3Dir, dirName)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("error: creating directory: %w", err)
	}

	boundary := opts.Boundary
	if boundary == "" {
		boundary = "service"
	}

	content := buildDocument(
		map[string]string{
			"id":       fmt.Sprintf("c3-%d", n),
			"title":    opts.Slug,
			"type":     "container",
			"boundary": boundary,
			"parent":   "c3-0",
			"goal":     opts.Goal,
			"summary":  opts.Summary,
		},
		opts.Slug,
		"container",
		opts.Goal,
	)

	readmePath := filepath.Join(dirPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing README.md: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(opts.C3Dir), readmePath)
	fmt.Fprintf(w, "Created: %s (id: c3-%d)\n", rel, n)
	return nil
}

func addRichComponent(opts AddOptions, w io.Writer) error {
	if opts.Container == "" {
		return fmt.Errorf("error: --container <id> is required for component\nhint: c3 add component auth-provider --container c3-1")
	}

	containerMatch := regexp.MustCompile(`^c3-(\d+)$`).FindStringSubmatch(opts.Container)
	if containerMatch == nil {
		return fmt.Errorf("error: invalid container id '%s'\nhint: use format c3-N, e.g. c3-1, c3-3", opts.Container)
	}
	containerNum, _ := strconv.Atoi(containerMatch[1])

	containerEntity := opts.Graph.Get(opts.Container)
	if containerEntity == nil {
		return fmt.Errorf("error: container '%s' not found", opts.Container)
	}

	componentID, err := numbering.NextComponentId(opts.Graph, containerNum, opts.Feature)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	category := "foundation"
	if opts.Feature {
		category = "feature"
	}

	content := buildDocument(
		map[string]string{
			"id":       componentID,
			"title":    opts.Slug,
			"type":     "component",
			"category": category,
			"parent":   opts.Container,
			"goal":     opts.Goal,
			"summary":  opts.Summary,
		},
		opts.Slug,
		"component",
		opts.Goal,
	)

	fileName := fmt.Sprintf("%s-%s.md", componentID, opts.Slug)
	containerDir := filepath.Join(opts.C3Dir, filepath.Dir(containerEntity.Path))
	filePath := filepath.Join(containerDir, fileName)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing component: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(opts.C3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, componentID)

	// Update container table
	containerReadme := filepath.Join(containerDir, "README.md")
	if _, err := os.Stat(containerReadme); err == nil {
		if wiring.AddComponentToContainerTable(containerReadme, componentID, opts.Slug, category, opts.Goal) {
			relReadme, _ := filepath.Rel(filepath.Dir(opts.C3Dir), containerReadme)
			fmt.Fprintf(w, "Updated: %s (component list)\n", relReadme)
		}
	}

	return nil
}

func addRichRef(opts AddOptions, w io.Writer) error {
	refsDir := filepath.Join(opts.C3Dir, "refs")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		return fmt.Errorf("error: creating refs/: %w", err)
	}

	fileName := fmt.Sprintf("ref-%s.md", opts.Slug)
	filePath := filepath.Join(refsDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("error: ref-%s already exists", opts.Slug)
	}

	content := buildDocument(
		map[string]string{
			"id":    fmt.Sprintf("ref-%s", opts.Slug),
			"title": opts.Slug,
			"goal":  opts.Goal,
			"scope": "[]",
		},
		opts.Slug,
		"ref",
		opts.Goal,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing ref: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(opts.C3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: ref-%s)\n", rel, opts.Slug)
	return nil
}

func addRichAdr(opts AddOptions, w io.Writer) error {
	adrDir := filepath.Join(opts.C3Dir, "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		return fmt.Errorf("error: creating adr/: %w", err)
	}

	adrID := numbering.NextAdrId(opts.Slug)
	fileName := fmt.Sprintf("%s.md", adrID)
	filePath := filepath.Join(adrDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("error: %s already exists", adrID)
	}

	today := time.Now().Format("20060102")
	content := buildDocument(
		map[string]string{
			"id":      adrID,
			"title":   opts.Slug,
			"type":    "adr",
			"status":  "proposed",
			"date":    today,
			"affects": "[c3-0]",
			"goal":    opts.Goal,
		},
		opts.Slug,
		"adr",
		opts.Goal,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing ADR: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(opts.C3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, adrID)
	return nil
}

func addRichRecipe(opts AddOptions, w io.Writer) error {
	recipesDir := filepath.Join(opts.C3Dir, "recipes")
	if err := os.MkdirAll(recipesDir, 0755); err != nil {
		return fmt.Errorf("error: creating recipes/: %w", err)
	}

	id := fmt.Sprintf("recipe-%s", opts.Slug)
	filePath := filepath.Join(recipesDir, id+".md")

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("error: %s already exists", id)
	}

	content := buildDocument(
		map[string]string{
			"id":    id,
			"title": opts.Slug,
			"type":  "recipe",
			"goal":  opts.Goal,
		},
		opts.Slug,
		"recipe",
		opts.Goal,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing recipe: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(opts.C3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, id)
	return nil
}

func buildDocument(fmFields map[string]string, title, entityType, goal string) string {
	var b strings.Builder

	b.WriteString("---\n")
	orderedKeys := []string{"id", "title", "type", "category", "boundary", "parent", "goal", "summary", "status", "date", "affects", "scope"}
	for _, k := range orderedKeys {
		v, ok := fmFields[k]
		if !ok || v == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	b.WriteString("---\n\n")

	b.WriteString(fmt.Sprintf("# %s\n", title))

	sections := schema.ForType(entityType)
	if sections == nil {
		return b.String()
	}

	for _, sec := range sections {
		b.WriteString(fmt.Sprintf("\n## %s\n", sec.Name))

		if sec.Name == "Goal" && goal != "" {
			b.WriteString(fmt.Sprintf("\n%s\n", goal))
			continue
		}

		if sec.ContentType == "table" && len(sec.Columns) > 0 {
			b.WriteString("\n")
			// Header row
			headers := make([]string, len(sec.Columns))
			seps := make([]string, len(sec.Columns))
			for i, col := range sec.Columns {
				headers[i] = col.Name
				seps[i] = strings.Repeat("-", len(col.Name))
			}
			b.WriteString(fmt.Sprintf("| %s |\n", strings.Join(headers, " | ")))
			b.WriteString(fmt.Sprintf("| %s |\n", strings.Join(seps, " | ")))
		} else {
			b.WriteString("\n")
		}
	}

	return b.String()
}
