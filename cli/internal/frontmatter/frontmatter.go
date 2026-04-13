package frontmatter

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// DocType represents the type of a C3 document.
type DocType int

const (
	DocUnknown DocType = iota
	DocContext
	DocContainer
	DocComponent
	DocRef
	DocADR
	DocRecipe
	DocRule
)

func (d DocType) String() string {
	switch d {
	case DocContext:
		return "context"
	case DocContainer:
		return "container"
	case DocComponent:
		return "component"
	case DocRef:
		return "ref"
	case DocADR:
		return "adr"
	case DocRecipe:
		return "recipe"
	case DocRule:
		return "rule"
	default:
		return "unknown"
	}
}

// Frontmatter holds the YAML frontmatter of a C3 document.
type Frontmatter struct {
	ID          string   `yaml:"id"`
	Seal        string   `yaml:"c3-seal,omitempty"`
	Title       string   `yaml:"title,omitempty"`
	Type        string   `yaml:"type,omitempty"`
	Category    string   `yaml:"category,omitempty"`
	Parent      string   `yaml:"parent,omitempty"`
	Goal        string   `yaml:"goal,omitempty"`
	Summary     string   `yaml:"summary,omitempty"`
	Boundary    string   `yaml:"boundary,omitempty"`
	Status      string   `yaml:"status,omitempty"`
	Date        string   `yaml:"date,omitempty"`
	Affects     []string `yaml:"affects,omitempty"`
	Refs        []string `yaml:"uses,omitempty"`
	Scope       []string `yaml:"scope,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Sources     []string `yaml:"sources,omitempty"`
	Origin      []string `yaml:"origin,omitempty"`
	// Extra holds any additional fields not in the known schema.
	Extra map[string]interface{} `yaml:",inline"`
}

// ParsedDoc represents a parsed C3 markdown document.
type ParsedDoc struct {
	Frontmatter *Frontmatter
	Body        string
	Path        string // relative to .c3/
}

// ParseFrontmatter extracts YAML frontmatter from markdown content.
// Returns nil frontmatter if not present or invalid.
func ParseFrontmatter(content string) (*Frontmatter, string) {
	if !strings.HasPrefix(content, "---\n") {
		return nil, content
	}

	end := strings.Index(content[4:], "\n---\n")
	var body string
	if end == -1 {
		// Handle EOF edge case: frontmatter ends with \n--- at end of string
		if strings.HasSuffix(content[4:], "\n---") {
			end = len(content[4:]) - 4
			body = ""
		} else {
			return nil, content
		}
	} else {
		body = content[4+end+5:]
	}

	yamlStr := content[4 : 4+end]

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &raw); err != nil {
		return nil, body
	}

	// Strip null values (YAML parses empty values like "goal: " as nil)
	for k, v := range raw {
		if v == nil {
			delete(raw, k)
		}
	}

	// Backward compat: merge "refs" into "uses" (canonical field)
	if refsVal, hasRefs := raw["refs"]; hasRefs {
		if usesVal, hasUses := raw["uses"]; hasUses {
			// Both present: merge refs into uses, dedup
			usesSlice := toStringSlice(usesVal)
			refsSlice := toStringSlice(refsVal)
			seen := make(map[string]bool, len(usesSlice))
			for _, v := range usesSlice {
				seen[v] = true
			}
			for _, v := range refsSlice {
				if !seen[v] {
					usesSlice = append(usesSlice, v)
				}
			}
			raw["uses"] = usesSlice
		} else {
			// Only refs: rename to uses
			raw["uses"] = refsVal
		}
		delete(raw, "refs")
	}

	// Check required field: id
	idVal, ok := raw["id"]
	if !ok {
		return nil, body
	}
	idStr, ok := idVal.(string)
	if !ok {
		return nil, body
	}
	if idStr == "" {
		return nil, body
	}

	// Re-marshal cleaned map and unmarshal into struct
	cleaned, err := yaml.Marshal(raw)
	if err != nil {
		return nil, body
	}

	var fm Frontmatter
	if err := yaml.Unmarshal(cleaned, &fm); err != nil {
		return nil, body
	}

	return &fm, body
}

// ClassifyDoc determines the DocType of a document from its frontmatter.
func ClassifyDoc(fm *Frontmatter) DocType {
	if fm.ID == "c3-0" {
		return DocContext
	}
	if fm.Type == "container" {
		return DocContainer
	}
	if fm.Type == "component" {
		return DocComponent
	}
	if fm.Type == "adr" || strings.HasPrefix(fm.ID, "adr-") {
		return DocADR
	}
	if fm.Type == "recipe" || strings.HasPrefix(fm.ID, "recipe-") {
		return DocRecipe
	}
	if fm.Type == "rule" || strings.HasPrefix(fm.ID, "rule-") {
		return DocRule
	}
	if strings.HasPrefix(fm.ID, "ref-") {
		return DocRef
	}
	return DocUnknown
}

// toStringSlice converts an interface{} (expected []interface{} of strings) to []string.
func toStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []interface{}:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return val
	}
	return nil
}

// StripAnchor removes the #fragment suffix from a source reference.
func StripAnchor(src string) string {
	if idx := strings.Index(src, "#"); idx > 0 {
		return src[:idx]
	}
	return src
}

// DeriveRelationships extracts all entity IDs this document references.
func DeriveRelationships(fm *Frontmatter) []string {
	var rels []string
	if fm.Parent != "" {
		rels = append(rels, fm.Parent)
	}
	rels = append(rels, fm.Affects...)
	rels = append(rels, fm.Refs...)
	rels = append(rels, fm.Scope...)
	for _, src := range fm.Sources {
		rels = append(rels, StripAnchor(src))
	}
	rels = append(rels, fm.Origin...)
	if rels == nil {
		rels = []string{}
	}
	return rels
}
