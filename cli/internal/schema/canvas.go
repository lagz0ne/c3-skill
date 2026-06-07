package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

type Canvas struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Domain      string       `json:"domain,omitempty"`
	Source      string       `json:"source,omitempty"`
	Sections    []SectionDef `json:"sections"`
	Reject      RejectRules  `json:"reject"`
}

type CanvasDocument struct {
	Domain    string       `yaml:"domain,omitempty"`
	Sections  []SectionDef `yaml:"sections"`
	RejectIf  []string     `yaml:"reject_if"`
	Workorder string       `yaml:"workorder"`
}

const CanvasesDir = "canvases"

var edgeTypeRE = regexp.MustCompile(`^edge<[^>\s]+>$`)

func CanvasFor(id string) (Canvas, bool) {
	return DefinitionFor(id)
}

func ResolveCanvas(c3Dir, id string) (Canvas, error) {
	if canvas, ok := DefinitionForDir(c3Dir, id); ok {
		return canvas, nil
	}
	return Canvas{}, fmt.Errorf("unknown canvas %q", id)
}

func AllCanvases(c3Dir string) ([]Canvas, error) {
	return AllDefinitions(c3Dir)
}

func AllDefinitions(c3Dir string) ([]Canvas, error) {
	defs := map[string]Canvas{}
	for _, id := range BuiltInDefinitionIDs() {
		def, ok := DefinitionFor(id)
		if !ok {
			continue
		}
		defs[def.ID] = def
	}
	projectCanvases, err := LoadProjectCanvases(c3Dir)
	if err != nil {
		return nil, err
	}
	for _, canvas := range projectCanvases {
		defs[canvas.ID] = canvas
	}
	out := make([]Canvas, 0, len(defs))
	for _, canvas := range defs {
		out = append(out, canvas)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func LoadProjectCanvases(c3Dir string) ([]Canvas, error) {
	if c3Dir == "" {
		return nil, nil
	}
	dir := filepath.Join(c3Dir, CanvasesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read canvases: %w", err)
	}
	var canvases []Canvas
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read canvas %s: %w", entry.Name(), err)
		}
		canvas, err := ParseCanvasDocument(filepath.ToSlash(filepath.Join(CanvasesDir, entry.Name())), string(data))
		if err != nil {
			return nil, err
		}
		if seen[canvas.ID] {
			return nil, fmt.Errorf("duplicate canvas id %q", canvas.ID)
		}
		seen[canvas.ID] = true
		canvases = append(canvases, canvas)
	}
	sort.Slice(canvases, func(i, j int) bool { return canvases[i].ID < canvases[j].ID })
	return canvases, nil
}

func ParseCanvasDocument(path, raw string) (Canvas, error) {
	fm, body := frontmatter.ParseFrontmatter(raw)
	if fm == nil {
		return Canvas{}, fmt.Errorf("invalid canvas %s: missing frontmatter", path)
	}
	if strings.TrimSpace(fm.ID) == "" {
		return Canvas{}, fmt.Errorf("invalid canvas %s: missing id", path)
	}
	if fm.Type != "canvas" {
		return Canvas{}, fmt.Errorf("invalid canvas %s: type must be canvas", path)
	}
	if strings.TrimSpace(fm.Description) == "" {
		return Canvas{}, fmt.Errorf("invalid canvas %s: missing description", path)
	}
	if fm.Title != "" || fm.Goal != "" || fm.Status != "" || fm.Summary != "" {
		return Canvas{}, fmt.Errorf("invalid canvas %s: frontmatter allows only id, type, description, and c3-seal", path)
	}
	for key := range fm.Extra {
		return Canvas{}, fmt.Errorf("invalid canvas %s: unknown frontmatter field %q", path, key)
	}
	var doc CanvasDocument
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		return Canvas{}, fmt.Errorf("invalid canvas %s: parse body: %w", path, err)
	}
	canvas := Canvas{
		ID:          fm.ID,
		Title:       titleFromID(fm.ID),
		Description: fm.Description,
		Domain:      doc.Domain,
		Source:      path,
		Sections:    doc.Sections,
		Reject: RejectRules{
			Bullets:   doc.RejectIf,
			Workorder: doc.Workorder,
		},
	}
	if err := ValidateCanvas(canvas); err != nil {
		return Canvas{}, fmt.Errorf("invalid canvas %s: %w", path, err)
	}
	return canvas, nil
}

func ValidateCanvas(canvas Canvas) error {
	if strings.TrimSpace(canvas.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(canvas.Description) == "" {
		return fmt.Errorf("missing description")
	}
	if len(canvas.Sections) == 0 {
		return fmt.Errorf("missing sections")
	}
	seen := map[string]bool{}
	for _, section := range canvas.Sections {
		if strings.TrimSpace(section.Name) == "" {
			return fmt.Errorf("section missing name")
		}
		if seen[section.Name] {
			return fmt.Errorf("duplicate section %q", section.Name)
		}
		seen[section.Name] = true
		switch section.ContentType {
		case "text", "table":
		default:
			return fmt.Errorf("section %q has unknown content_type %q", section.Name, section.ContentType)
		}
		if section.ContentType == "table" && len(section.Columns) == 0 {
			return fmt.Errorf("table section %q missing columns", section.Name)
		}
		for _, column := range section.Columns {
			if strings.TrimSpace(column.Name) == "" {
				return fmt.Errorf("section %q has column missing name", section.Name)
			}
			if !IsCanvasPrimitive(column.Type, column.Values) {
				return fmt.Errorf("section %q column %q has unsupported type %q", section.Name, column.Name, column.Type)
			}
		}
	}
	return nil
}

func IsCanvasPrimitive(columnType string, values []string) bool {
	switch strings.TrimSpace(columnType) {
	case "text", "date", "cite", "check", "entity_id", "reference", "evidence":
		return true
	case "enum":
		return len(values) > 0
	default:
		return edgeTypeRE.MatchString(columnType)
	}
}

// DefinitionForDir resolves an entity type's definition, preferring a
// project-local materialized definition at .c3/canvases/<type>.md (user-owned,
// frozen) and falling back to the embedded definition (DefinitionFor). This is
// the seam that makes definitions user-editable: an edit to the materialized
// file changes what validation enforces, while a fresh project still works off
// the embedded seed. (slice 8)
func DefinitionForDir(c3Dir, entityType string) (Canvas, bool) {
	if c3Dir != "" {
		for _, id := range definitionLookupIDs(entityType) {
			rel := filepath.ToSlash(filepath.Join(CanvasesDir, id+".md"))
			data, err := os.ReadFile(filepath.Join(c3Dir, CanvasesDir, id+".md"))
			if err != nil {
				continue
			}
			if canvas, err := ParseCanvasDocument(rel, string(data)); err == nil {
				return canvas, true
			}
		}
	}
	return DefinitionFor(entityType)
}

func definitionLookupIDs(entityType string) []string {
	entityType = strings.TrimSpace(entityType)
	canonical := CanonicalDefinitionID(entityType)
	if entityType == "" || entityType == canonical {
		return []string{canonical}
	}
	return []string{entityType, canonical}
}
