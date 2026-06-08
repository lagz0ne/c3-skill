package schema

import "strings"

// SectionDef defines a known section for an entity type.
type SectionDef struct {
	Name        string      `json:"name" yaml:"name"`
	ContentType string      `json:"content_type" yaml:"content_type"`
	Required    bool        `json:"required" yaml:"required"`
	Purpose     string      `json:"purpose,omitempty" yaml:"purpose,omitempty"`
	Fill        string      `json:"fill,omitempty" yaml:"fill,omitempty"`
	Failure     string      `json:"failure,omitempty" yaml:"failure,omitempty"`
	Columns     []ColumnDef `json:"columns,omitempty" yaml:"columns,omitempty"`
	MinWords    int         `json:"min_words,omitempty" yaml:"min_words,omitempty"`
	MinRows     int         `json:"min_rows,omitempty" yaml:"min_rows,omitempty"`
}

// ColumnDef defines a typed column within a table section.
type ColumnDef struct {
	Name   string   `json:"name" yaml:"name"`
	Type   string   `json:"type" yaml:"type"`
	Values []string `json:"values,omitempty" yaml:"values,omitempty"`
}

// RejectRules is the rejection contract surfaced before drafting an entity body.
// Bullets are individual reject conditions; Workorder is the prose framing that
// follows the bullets in text output.
type RejectRules struct {
	Bullets   []string `json:"bullets" yaml:"bullets"`
	Workorder string   `json:"workorder" yaml:"workorder"`
}

// RejectFor returns the rejection contract for an entity type, or zero value if
// the resolved canvas has none.
func RejectFor(entityType string) RejectRules {
	def, ok := DefinitionFor(entityType)
	if !ok {
		return RejectRules{}
	}
	return def.Reject
}

func titleFromID(id string) string {
	parts := strings.Split(id, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func CanonicalDefinitionID(entityType string) string {
	return canonicalDefinitionID(entityType)
}

func BuiltInDefinitionIDs() []string {
	return builtInDefinitionIDs()
}

func DefinitionFor(entityType string) (Canvas, bool) {
	id := canonicalDefinitionID(entityType)
	def, ok := builtInDefinitions[id]
	if !ok {
		return Canvas{}, false
	}
	return def, true
}

// ForType returns section definitions for an entity type, or nil if unknown.
func ForType(entityType string) []SectionDef {
	def, ok := DefinitionFor(entityType)
	if !ok {
		return nil
	}
	return def.Sections
}

// PurposeOf returns the purpose string for a section within an entity type.
func PurposeOf(entityType, sectionName string) string {
	def, ok := DefinitionFor(entityType)
	if !ok {
		return ""
	}
	for _, s := range def.Sections {
		if s.Name == sectionName {
			return s.Purpose
		}
	}
	return ""
}
