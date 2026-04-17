package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

var (
	validSlug   = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	reContainer = regexp.MustCompile(`^c3-(\d+)$`)
)

// AddResult is the structured output from add commands.
type AddResult struct {
	ID       string   `json:"id"`
	Type     string   `json:"type,omitempty"`
	Sections []string `json:"sections,omitempty"`
}

// RunAdd creates a new C3 entity with body content. Body is required via reader.
func RunAdd(entityType, slug string, s *store.Store, container string, feature bool, body io.Reader, w io.Writer) error {
	return RunAddFormatted(entityType, slug, s, container, feature, body, w, FormatHuman)
}

// RunAddFormatted creates a new C3 entity and writes either human or structured output.
func RunAddFormatted(entityType, slug string, s *store.Store, container string, feature bool, body io.Reader, w io.Writer, format OutputFormat) error {
	if entityType == "" || slug == "" {
		return fmt.Errorf("error: usage: c3x add <type> <slug> < body.md\nhint: types: container, component, ref, rule, adr, recipe")
	}

	if !validSlug.MatchString(slug) {
		return fmt.Errorf("error: invalid slug '%s'\nhint: use kebab-case (e.g. auth-provider, rate-limiting)", slug)
	}

	// Read body content
	bodyContent, err := readBody(body)
	if err != nil {
		return err
	}

	// Build entity
	entity, err := buildEntity(entityType, slug, s, container, feature)
	if err != nil {
		return err
	}

	// Validate body against schema BEFORE any DB writes
	issues := validateBodyContent(bodyContent, entityType)
	if entityType == "adr" {
		issues = append(issues, validateADRCreationBody(bodyContent)...)
	}
	if len(issues) > 0 {
		return formatValidationError(entityType+"-"+slug, issues)
	}

	// Insert entity
	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting %s: %w", entityType, err)
	}

	// Write content (nodes, merkle, version, goal sync)
	if err := content.WriteEntity(s, entity.ID, bodyContent); err != nil {
		// Compensate: remove the entity we just inserted
		s.DeleteEntity(entity.ID)
		return fmt.Errorf("error: writing content: %w", err)
	}

	result := AddResult{ID: entity.ID, Type: entityType}
	if sections := schema.ForType(entityType); sections != nil {
		for _, sec := range sections {
			result.Sections = append(result.Sections, sec.Name)
		}
	}
	hints := cascadeHintsForEntity(entity)
	if entity.Type == "component" {
		hints = newComponentTopDownHints(entity)
	}
	if format != FormatHuman {
		return WriteObjectOutput(w, result, format, hints)
	}

	fmt.Fprintf(w, "Created: %s %s (id: %s)\n", entityType, slug, entity.ID)
	if entity.Type == "component" {
		writeAgentHints(w, hints)
		return nil
	}
	writeAgentHints(w, hints)
	return nil
}

func readBody(r io.Reader) (string, error) {
	if r == nil {
		return "", fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("error: reading body: %w", err)
	}
	body := strings.TrimSpace(string(data))
	if body == "" {
		return "", fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}
	return body, nil
}

func validateADRCreationBody(body string) []Issue {
	sections := markdown.ParseSections(body)
	sectionMap := make(map[string]string)
	for _, section := range sections {
		if section.Name != "" {
			sectionMap[section.Name] = strings.TrimSpace(section.Content)
		}
	}

	var issues []Issue
	for _, def := range schema.ForType("adr") {
		content, exists := sectionMap[def.Name]
		if !exists {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  fmt.Sprintf("missing required section: %s", def.Name),
				Hint:     fmt.Sprintf("add ## %s before running c3x add adr; ADR creation is all-or-nothing", def.Name),
			})
			continue
		}
		if content == "" {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  fmt.Sprintf("empty required section: %s", def.Name),
				Hint:     fmt.Sprintf("fill ## %s before running c3x add adr; use N.A - <reason> when not applicable", def.Name),
			})
			continue
		}
		if def.ContentType != "table" {
			continue
		}
		table, err := markdown.ParseTable(content)
		if err != nil {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  fmt.Sprintf("invalid required table: %s", def.Name),
				Hint:     fmt.Sprintf("use c3x schema adr for the %s table columns", def.Name),
			})
			continue
		}
		if len(table.Rows) == 0 {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  fmt.Sprintf("empty required table: %s", def.Name),
				Hint:     fmt.Sprintf("add at least one %s row; use N.A - <reason> when not applicable", def.Name),
			})
		}
		for _, col := range def.Columns {
			if !containsString(table.Headers, col.Name) {
				issues = append(issues, Issue{
					Severity: "error",
					Message:  fmt.Sprintf("missing required column %q in table: %s", col.Name, def.Name),
					Hint:     fmt.Sprintf("use c3x schema adr for the %s table columns", def.Name),
				})
			}
		}
	}
	return issues
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func buildEntity(entityType, slug string, s *store.Store, container string, feature bool) (*store.Entity, error) {
	switch entityType {
	case "container":
		return buildContainer(slug, s)
	case "component":
		return buildComponent(slug, s, container, feature)
	case "ref":
		return buildRef(slug, s)
	case "rule":
		return buildRule(slug, s)
	case "adr":
		return buildAdr(slug, s)
	case "recipe":
		return buildRecipe(slug, s)
	default:
		return nil, fmt.Errorf("error: unknown entity type '%s'\nhint: types: container, component, ref, rule, adr, recipe", entityType)
	}
}

func buildContainer(slug string, s *store.Store) (*store.Entity, error) {
	n, err := nextContainerNum(s)
	if err != nil {
		return nil, fmt.Errorf("error: computing container number: %w", err)
	}
	return &store.Entity{
		ID: fmt.Sprintf("c3-%d", n), Type: "container", Title: slug, Slug: slug,
		ParentID: "c3-0", Boundary: "service", Status: "active", Metadata: "{}",
	}, nil
}

func buildComponent(slug string, s *store.Store, containerArg string, feature bool) (*store.Entity, error) {
	if containerArg == "" {
		return nil, fmt.Errorf("error: --container <id> is required for component\nhint: c3x add component auth-provider --container c3-1")
	}
	containerMatch := reContainer.FindStringSubmatch(containerArg)
	if containerMatch == nil {
		return nil, fmt.Errorf("error: invalid container id '%s'\nhint: use format c3-N, e.g. c3-1, c3-3", containerArg)
	}
	containerNum, _ := strconv.Atoi(containerMatch[1])
	if _, err := s.GetEntity(containerArg); err != nil {
		return nil, fmt.Errorf("error: container '%s' not found", containerArg)
	}
	componentID, err := nextComponentID(s, containerNum, feature)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	category := "foundation"
	if feature {
		category = "feature"
	}
	return &store.Entity{
		ID: componentID, Type: "component", Title: slug, Slug: slug,
		Category: category, ParentID: containerArg, Status: "active", Metadata: "{}",
	}, nil
}

func buildRef(slug string, s *store.Store) (*store.Entity, error) {
	id := "ref-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return nil, fmt.Errorf("error: %s already exists", id)
	}
	return &store.Entity{
		ID: id, Type: "ref", Title: slug, Slug: slug, Status: "active", Metadata: "{}",
	}, nil
}

func buildRule(slug string, s *store.Store) (*store.Entity, error) {
	id := "rule-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return nil, fmt.Errorf("error: %s already exists", id)
	}
	return &store.Entity{
		ID: id, Type: "rule", Title: slug, Slug: slug, Status: "active", Metadata: "{}",
	}, nil
}

func buildAdr(slug string, s *store.Store) (*store.Entity, error) {
	now := time.Now()
	adrID := fmt.Sprintf("adr-%s-%s", now.Format("20060102"), slug)
	if _, err := s.GetEntity(adrID); err == nil {
		return nil, fmt.Errorf("error: %s already exists", adrID)
	}
	return &store.Entity{
		ID: adrID, Type: "adr", Title: slug, Slug: slug,
		Status: "proposed", Date: now.Format("2006-01-02"), Metadata: "{}",
	}, nil
}

func buildRecipe(slug string, s *store.Store) (*store.Entity, error) {
	id := "recipe-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return nil, fmt.Errorf("error: %s already exists", id)
	}
	return &store.Entity{
		ID: id, Type: "recipe", Title: slug, Slug: slug, Status: "active", Metadata: "{}",
	}, nil
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
