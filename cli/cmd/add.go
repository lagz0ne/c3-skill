package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/numbering"
	"github.com/lagz0ne/c3-design/cli/internal/templates"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
	"github.com/lagz0ne/c3-design/cli/internal/wiring"
)

var validSlug = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// RunAdd creates a new C3 entity.
func RunAdd(entityType, slug, c3Dir string, graph *walker.C3Graph, container string, feature bool, w io.Writer) error {
	if entityType == "" || slug == "" {
		return fmt.Errorf("error: usage: c3 add <type> <slug>\nhint: types: container, component, ref, adr, recipe")
	}

	if !validSlug.MatchString(slug) {
		return fmt.Errorf("error: invalid slug '%s'\nhint: use kebab-case (e.g. auth-provider, rate-limiting)", slug)
	}

	switch entityType {
	case "container":
		return addContainer(slug, c3Dir, graph, w)
	case "component":
		return addComponent(slug, c3Dir, graph, container, feature, w)
	case "ref":
		return addRef(slug, c3Dir, w)
	case "adr":
		return addAdr(slug, c3Dir, w)
	case "recipe":
		return addRecipe(slug, c3Dir, w)
	default:
		return fmt.Errorf("error: unknown entity type '%s'\nhint: types: container, component, ref, adr, recipe", entityType)
	}
}

func addContainer(slug, c3Dir string, graph *walker.C3Graph, w io.Writer) error {
	n := numbering.NextContainerId(graph)
	dirName := fmt.Sprintf("c3-%d-%s", n, slug)
	dirPath := filepath.Join(c3Dir, dirName)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("error: creating directory: %w", err)
	}

	content, err := templates.Render("container.md", map[string]string{
		"${N}":              strconv.Itoa(n),
		"${CONTAINER_NAME}": slug,
		"${BOUNDARY}":       "service",
		"${GOAL}":           "",
		"${SUMMARY}":        "",
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	readmePath := filepath.Join(dirPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing README.md: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(c3Dir), readmePath)
	fmt.Fprintf(w, "Created: %s (id: c3-%d)\n", rel, n)
	return nil
}

func addComponent(slug, c3Dir string, graph *walker.C3Graph, containerArg string, feature bool, w io.Writer) error {
	if containerArg == "" {
		return fmt.Errorf("error: --container <id> is required for component\nhint: c3 add component auth-provider --container c3-1")
	}

	containerMatch := regexp.MustCompile(`^c3-(\d+)$`).FindStringSubmatch(containerArg)
	if containerMatch == nil {
		return fmt.Errorf("error: invalid container id '%s'\nhint: use format c3-N, e.g. c3-1, c3-3", containerArg)
	}
	containerNum, _ := strconv.Atoi(containerMatch[1])

	containerEntity := graph.Get(containerArg)
	if containerEntity == nil {
		return fmt.Errorf("error: container '%s' not found", containerArg)
	}

	componentID, err := numbering.NextComponentId(graph, containerNum, feature)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	category := "foundation"
	if feature {
		category = "feature"
	}

	nn := componentID[len(fmt.Sprintf("c3-%d", containerNum)):]
	fileName := fmt.Sprintf("%s-%s.md", componentID, slug)
	containerDir := filepath.Join(c3Dir, filepath.Dir(containerEntity.Path))
	filePath := filepath.Join(containerDir, fileName)

	content, err := templates.Render("component.md", map[string]string{
		"${N}${NN}":         componentID[len("c3-"):],
		"${N}":              strconv.Itoa(containerNum),
		"${NN}":             nn,
		"${COMPONENT_NAME}": slug,
		"${CATEGORY}":       category,
		"${GOAL}":           "",
		"${SUMMARY}":        "",
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing component: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(c3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, componentID)

	// Update container table
	containerReadme := filepath.Join(containerDir, "README.md")
	if _, err := os.Stat(containerReadme); err == nil {
		if wiring.AddComponentToContainerTable(containerReadme, componentID, slug, category, "") {
			relReadme, _ := filepath.Rel(filepath.Dir(c3Dir), containerReadme)
			fmt.Fprintf(w, "Updated: %s (component list)\n", relReadme)
		}
	}

	return nil
}

// addSubdirEntity creates a simple entity (ref, recipe) in a subdirectory of .c3/.
func addSubdirEntity(slug, c3Dir, subDir, prefix, templateName string, templateVars map[string]string, w io.Writer) error {
	dir := filepath.Join(c3Dir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error: creating %s/: %w", subDir, err)
	}

	id := prefix + slug
	filePath := filepath.Join(dir, id+".md")

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("error: %s already exists", id)
	}

	content, err := templates.Render(templateName, templateVars)
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing %s: %w", subDir, err)
	}

	rel, _ := filepath.Rel(filepath.Dir(c3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, id)
	return nil
}

func addRef(slug, c3Dir string, w io.Writer) error {
	return addSubdirEntity(slug, c3Dir, "refs", "ref-", "ref.md", map[string]string{
		"${SLUG}":  slug,
		"${TITLE}": slug,
		"${GOAL}":  "",
	}, w)
}

func addRecipe(slug, c3Dir string, w io.Writer) error {
	return addSubdirEntity(slug, c3Dir, "recipes", "recipe-", "recipe.md", map[string]string{
		"${SLUG}": slug,
	}, w)
}

func addAdr(slug, c3Dir string, w io.Writer) error {
	adrDir := filepath.Join(c3Dir, "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		return fmt.Errorf("error: creating adr/: %w", err)
	}

	adrID := numbering.NextAdrId(slug)
	fileName := fmt.Sprintf("%s.md", adrID)
	filePath := filepath.Join(adrDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("error: %s already exists", adrID)
	}

	today := time.Now().Format("20060102")
	content, err := templates.Render("adr-000.md", map[string]string{
		"adr-00000000-c3-adoption":               adrID,
		"C3 Architecture Documentation Adoption":  slug,
		"Adopt C3 methodology for ${PROJECT}.\n":  "",
		"${DATE}":    today,
		"${PROJECT}": "",
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing ADR: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(c3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, adrID)
	return nil
}
