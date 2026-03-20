package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/templates"
)

var validSlug = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// AddResult is the JSON output from add commands.
type AddResult struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// RunAdd creates a new C3 entity in the store.
func RunAdd(entityType, slug, c3Dir string, s *store.Store, container string, feature bool, w io.Writer) error {
	if entityType == "" || slug == "" {
		return fmt.Errorf("error: usage: c3 add <type> <slug>\nhint: types: container, component, ref, rule, adr, recipe")
	}

	if !validSlug.MatchString(slug) {
		return fmt.Errorf("error: invalid slug '%s'\nhint: use kebab-case (e.g. auth-provider, rate-limiting)", slug)
	}

	switch entityType {
	case "container":
		return addContainer(slug, c3Dir, s, w)
	case "component":
		return addComponent(slug, c3Dir, s, container, feature, w)
	case "ref":
		return addRef(slug, s, w)
	case "rule":
		return addRule(slug, s, w)
	case "adr":
		return addAdr(slug, s, w)
	case "recipe":
		return addRecipe(slug, s, w)
	default:
		return fmt.Errorf("error: unknown entity type '%s'\nhint: types: container, component, ref, rule, adr, recipe", entityType)
	}
}

// nextContainerNum returns the next available container number by querying the store.
func nextContainerNum(s *store.Store) (int, error) {
	containers, err := s.EntitiesByType("container")
	if err != nil {
		return 0, err
	}
	max := 0
	for _, c := range containers {
		numStr := ""
		if len(c.ID) > 3 && c.ID[:3] == "c3-" {
			numStr = c.ID[3:]
		}
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}

// nextComponentID returns the next available component ID for a container.
func nextComponentID(s *store.Store, containerNum int, feature bool) (string, error) {
	prefix := fmt.Sprintf("c3-%d", containerNum)
	components, err := s.EntitiesByType("component")
	if err != nil {
		return "", err
	}

	var nums []int
	for _, c := range components {
		if len(c.ID) > len(prefix) && c.ID[:len(prefix)] == prefix {
			numStr := c.ID[len(prefix):]
			n, err := strconv.Atoi(numStr)
			if err != nil {
				continue
			}
			nums = append(nums, n)
		}
	}

	if feature {
		max := 9
		for _, n := range nums {
			if n >= 10 && n > max {
				max = n
			}
		}
		next := max + 1
		return fmt.Sprintf("c3-%d%02d", containerNum, next), nil
	}

	// Foundation: 01-09
	max := 0
	for _, n := range nums {
		if n >= 1 && n <= 9 && n > max {
			max = n
		}
	}
	next := max + 1
	if next > 9 {
		return "", fmt.Errorf("container c3-%d has no more foundation slots (01-09 full)", containerNum)
	}
	return fmt.Sprintf("c3-%d%02d", containerNum, next), nil
}

func addContainer(slug, c3Dir string, s *store.Store, w io.Writer) error {
	n, err := nextContainerNum(s)
	if err != nil {
		return fmt.Errorf("error: computing container number: %w", err)
	}

	id := fmt.Sprintf("c3-%d", n)
	entity := &store.Entity{
		ID:       id,
		Type:     "container",
		Title:    slug,
		Slug:     slug,
		ParentID: "c3-0",
		Boundary: "service",
		Goal:     "",
		Summary:  "",
		Status:   "active",
		Metadata: "{}",
	}

	// Generate body from template
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
	entity.Body = content

	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting container: %w", err)
	}

	// Also create the directory and file for backward compatibility
	dirName := fmt.Sprintf("c3-%d-%s", n, slug)
	dirPath := filepath.Join(c3Dir, dirName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("error: creating directory: %w", err)
	}
	readmePath := filepath.Join(dirPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing README.md: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(c3Dir), readmePath)
	fmt.Fprintf(w, "Created: %s (id: c3-%d)\n", rel, n)
	return nil
}

func addComponent(slug, c3Dir string, s *store.Store, containerArg string, feature bool, w io.Writer) error {
	if containerArg == "" {
		return fmt.Errorf("error: --container <id> is required for component\nhint: c3 add component auth-provider --container c3-1")
	}

	containerMatch := regexp.MustCompile(`^c3-(\d+)$`).FindStringSubmatch(containerArg)
	if containerMatch == nil {
		return fmt.Errorf("error: invalid container id '%s'\nhint: use format c3-N, e.g. c3-1, c3-3", containerArg)
	}
	containerNum, _ := strconv.Atoi(containerMatch[1])

	containerEntity, err := s.GetEntity(containerArg)
	if err != nil {
		return fmt.Errorf("error: container '%s' not found", containerArg)
	}

	componentID, err := nextComponentID(s, containerNum, feature)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	category := "foundation"
	if feature {
		category = "feature"
	}

	nn := componentID[len(fmt.Sprintf("c3-%d", containerNum)):]
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

	entity := &store.Entity{
		ID:       componentID,
		Type:     "component",
		Title:    slug,
		Slug:     slug,
		Category: category,
		ParentID: containerArg,
		Goal:     "",
		Summary:  "",
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}

	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting component: %w", err)
	}

	// Also create the file for backward compatibility
	// Find the container directory by slug
	containerDir := filepath.Join(c3Dir, fmt.Sprintf("c3-%d-%s", containerNum, containerEntity.Slug))
	fileName := fmt.Sprintf("%s-%s.md", componentID, slug)
	filePath := filepath.Join(containerDir, fileName)

	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return fmt.Errorf("error: creating directory: %w", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error: writing component: %w", err)
	}

	rel, _ := filepath.Rel(filepath.Dir(c3Dir), filePath)
	fmt.Fprintf(w, "Created: %s (id: %s)\n", rel, componentID)

	return nil
}

func addRef(slug string, s *store.Store, w io.Writer) error {
	id := "ref-" + slug
	// Check if already exists
	if _, err := s.GetEntity(id); err == nil {
		return fmt.Errorf("error: %s already exists", id)
	}

	content, err := templates.Render("ref.md", map[string]string{
		"${SLUG}":  slug,
		"${TITLE}": slug,
		"${GOAL}":  "",
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	entity := &store.Entity{
		ID:       id,
		Type:     "ref",
		Title:    slug,
		Slug:     slug,
		Goal:     "",
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}
	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting ref: %w", err)
	}

	fmt.Fprintf(w, "Created: ref (id: %s)\n", id)
	return nil
}

func addRule(slug string, s *store.Store, w io.Writer) error {
	id := "rule-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return fmt.Errorf("error: %s already exists", id)
	}

	content, err := templates.Render("rule.md", map[string]string{
		"${SLUG}":  slug,
		"${TITLE}": slug,
		"${GOAL}":  "",
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	entity := &store.Entity{
		ID:       id,
		Type:     "rule",
		Title:    slug,
		Slug:     slug,
		Goal:     "",
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}
	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting rule: %w", err)
	}

	fmt.Fprintf(w, "Created: rule (id: %s)\n", id)
	return nil
}

func addRecipe(slug string, s *store.Store, w io.Writer) error {
	id := "recipe-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return fmt.Errorf("error: %s already exists", id)
	}

	content, err := templates.Render("recipe.md", map[string]string{
		"${SLUG}": slug,
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	entity := &store.Entity{
		ID:       id,
		Type:     "recipe",
		Title:    slug,
		Slug:     slug,
		Goal:     "",
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}
	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting recipe: %w", err)
	}

	fmt.Fprintf(w, "Created: recipe (id: %s)\n", id)
	return nil
}

func addAdr(slug string, s *store.Store, w io.Writer) error {
	adrID := nextAdrID(slug)
	if _, err := s.GetEntity(adrID); err == nil {
		return fmt.Errorf("error: %s already exists", adrID)
	}

	today := time.Now().Format("2006-01-02")
	content, err := templates.Render("adr.md", map[string]string{
		"${ID}":    adrID,
		"${TITLE}": slug,
		"${DATE}":  today,
	})
	if err != nil {
		return fmt.Errorf("error: rendering template: %w", err)
	}

	entity := &store.Entity{
		ID:       adrID,
		Type:     "adr",
		Title:    slug,
		Slug:     slug,
		Status:   "proposed",
		Date:     today,
		Body:     content,
		Metadata: "{}",
	}
	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting ADR: %w", err)
	}

	fmt.Fprintf(w, "Created: adr (id: %s)\n", adrID)
	return nil
}

// nextAdrID generates an ADR ID with today's date and the given slug.
func nextAdrID(slug string) string {
	date := time.Now().Format("20060102")
	return fmt.Sprintf("adr-%s-%s", date, slug)
}
