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
	// Status is the declared legal-set for a change doc, from the canvas
	// frontmatter `status: [...]`. Empty for fact canvases. Its presence (not
	// any column) is what makes a canvas a change doc.
	Status []string `json:"status,omitempty"`
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
	if fm.Title != "" || fm.Goal != "" || fm.Summary != "" {
		return Canvas{}, fmt.Errorf("invalid canvas %s: frontmatter allows only id, type, description, status, and c3-seal", path)
	}
	// A scalar `status:` on a canvas is malformed — a canvas declares a status
	// legal-SET (a list), not a single value.
	if fm.Status != "" {
		return Canvas{}, fmt.Errorf("invalid canvas %s: status must be a list of states, e.g. status: [open, accepted, done, superseded]", path)
	}
	// A declared-but-empty or duplicate-state status list is malformed. A nil
	// StatusSet means the key was absent (fact canvas), which is fine.
	if fm.StatusSet != nil {
		if len(fm.StatusSet) == 0 {
			return Canvas{}, fmt.Errorf("invalid canvas %s: status declaration is empty; list at least one state", path)
		}
		seenState := map[string]bool{}
		for _, state := range fm.StatusSet {
			if strings.TrimSpace(state) == "" {
				return Canvas{}, fmt.Errorf("invalid canvas %s: status declaration has an empty state", path)
			}
			if seenState[state] {
				return Canvas{}, fmt.Errorf("invalid canvas %s: status declaration has duplicate state %q", path, state)
			}
			seenState[state] = true
		}
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
		Status: fm.StatusSet,
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
	strictTableBlocks := 0
	for _, section := range canvas.Sections {
		if strings.TrimSpace(section.Name) == "" {
			return fmt.Errorf("section missing name")
		}
		if seen[section.Name] {
			return fmt.Errorf("duplicate section %q", section.Name)
		}
		seen[section.Name] = true
		// FREE sections are narrative: their content is never shape-checked
		// (content_type, columns, typed-column primitives are all skipped). Only
		// the name/uniqueness guards above apply.
		if section.Free {
			continue
		}
		switch section.ContentType {
		case "text", "table":
		default:
			return fmt.Errorf("section %q has unknown content_type %q", section.Name, section.ContentType)
		}
		if section.ContentType == "table" {
			if len(section.Columns) == 0 {
				return fmt.Errorf("table section %q missing columns", section.Name)
			}
			strictTableBlocks++
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
	// A change doc (declares a status set) must carry at least one STRICT
	// change-set block — a non-FREE table section. Zero is invalid: a change doc
	// with only FREE reasoning sections has nothing to typed-column check.
	if len(canvas.Status) > 0 && strictTableBlocks == 0 {
		return fmt.Errorf("change doc %q must declare a STRICT change-set block (a non-free table section)", canvas.ID)
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
// the embedded seed.
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
