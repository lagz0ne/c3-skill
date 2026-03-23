package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// AddOptions holds parameters for the richer add command.
type AddOptions struct {
	EntityType string
	Slug       string
	Store      *store.Store
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
		return RunAdd(opts.EntityType, opts.Slug, opts.Store, opts.Container, opts.Feature, w)
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
	n, err := nextContainerNum(opts.Store)
	if err != nil {
		return fmt.Errorf("error: computing container number: %w", err)
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

	entity := &store.Entity{
		ID:       fmt.Sprintf("c3-%d", n),
		Type:     "container",
		Title:    opts.Slug,
		Slug:     opts.Slug,
		ParentID: "c3-0",
		Boundary: boundary,
		Goal:     opts.Goal,
		Summary:  opts.Summary,
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}

	if err := opts.Store.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting container: %w", err)
	}

	fmt.Fprintf(w, "Created: container (id: c3-%d)\n", n)
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

	if _, err := opts.Store.GetEntity(opts.Container); err != nil {
		return fmt.Errorf("error: container '%s' not found", opts.Container)
	}

	componentID, err := nextComponentID(opts.Store, containerNum, opts.Feature)
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

	entity := &store.Entity{
		ID:       componentID,
		Type:     "component",
		Title:    opts.Slug,
		Slug:     opts.Slug,
		Category: category,
		ParentID: opts.Container,
		Goal:     opts.Goal,
		Summary:  opts.Summary,
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}

	if err := opts.Store.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting component: %w", err)
	}

	fmt.Fprintf(w, "Created: component (id: %s)\n", componentID)
	return nil
}

func addRichRef(opts AddOptions, w io.Writer) error {
	id := fmt.Sprintf("ref-%s", opts.Slug)
	if _, err := opts.Store.GetEntity(id); err == nil {
		return fmt.Errorf("error: ref-%s already exists", opts.Slug)
	}

	content := buildDocument(
		map[string]string{
			"id":    id,
			"title": opts.Slug,
			"goal":  opts.Goal,
			"scope": "[]",
		},
		opts.Slug,
		"ref",
		opts.Goal,
	)

	entity := &store.Entity{
		ID:       id,
		Type:     "ref",
		Title:    opts.Slug,
		Slug:     opts.Slug,
		Goal:     opts.Goal,
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}
	if err := opts.Store.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting ref: %w", err)
	}

	fmt.Fprintf(w, "Created: ref (id: ref-%s)\n", opts.Slug)
	return nil
}

func addRichAdr(opts AddOptions, w io.Writer) error {
	now := time.Now()
	adrID := fmt.Sprintf("adr-%s-%s", now.Format("20060102"), opts.Slug)
	if _, err := opts.Store.GetEntity(adrID); err == nil {
		return fmt.Errorf("error: %s already exists", adrID)
	}

	today := now.Format("2006-01-02")
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

	entity := &store.Entity{
		ID:       adrID,
		Type:     "adr",
		Title:    opts.Slug,
		Slug:     opts.Slug,
		Status:   "proposed",
		Date:     today,
		Goal:     opts.Goal,
		Body:     content,
		Metadata: "{}",
	}
	if err := opts.Store.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting ADR: %w", err)
	}

	// Add affects relationship to c3-0
	_ = opts.Store.AddRelationship(&store.Relationship{
		FromID:  adrID,
		ToID:    "c3-0",
		RelType: "affects",
	})

	fmt.Fprintf(w, "Created: adr (id: %s)\n", adrID)
	return nil
}

func addRichRecipe(opts AddOptions, w io.Writer) error {
	id := fmt.Sprintf("recipe-%s", opts.Slug)
	if _, err := opts.Store.GetEntity(id); err == nil {
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

	entity := &store.Entity{
		ID:       id,
		Type:     "recipe",
		Title:    opts.Slug,
		Slug:     opts.Slug,
		Goal:     opts.Goal,
		Body:     content,
		Status:   "active",
		Metadata: "{}",
	}
	if err := opts.Store.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting recipe: %w", err)
	}

	fmt.Fprintf(w, "Created: recipe (id: %s)\n", id)
	return nil
}

func buildDocument(fmFields map[string]string, title, entityType, goal string) string {
	var b strings.Builder

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
