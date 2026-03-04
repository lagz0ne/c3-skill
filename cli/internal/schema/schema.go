package schema

// SectionDef defines a known section for an entity type.
type SectionDef struct {
	Name        string      `json:"name"`
	ContentType string      `json:"content_type"`
	Required    bool        `json:"required"`
	Purpose     string      `json:"purpose,omitempty"`
	Columns     []ColumnDef `json:"columns,omitempty"`
}

// ColumnDef defines a typed column within a table section.
type ColumnDef struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []string `json:"values,omitempty"`
}

// Registry maps entity types to their ordered section definitions.
var Registry = map[string][]SectionDef{
	"component": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What this component exists to do"},
		{Name: "Dependencies", ContentType: "table", Required: true, Purpose: "Data flowing in and out", Columns: []ColumnDef{
			{Name: "Direction", Type: "enum", Values: []string{"IN", "OUT"}},
			{Name: "What", Type: "text"},
			{Name: "From/To", Type: "entity_id"},
		}},
		{Name: "Related Refs", ContentType: "table", Required: false, Purpose: "Cross-cutting concerns applied here", Columns: []ColumnDef{
			{Name: "Ref", Type: "ref_id"},
			{Name: "Role", Type: "text"},
		}},
		{Name: "Container Connection", ContentType: "text", Required: false, Purpose: "How this component fits in the container"},
	},
	"container": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What this container exists to do"},
		{Name: "Components", ContentType: "table", Required: true, Purpose: "Parts that compose this container", Columns: []ColumnDef{
			{Name: "ID", Type: "entity_id"},
			{Name: "Name", Type: "text"},
			{Name: "Category", Type: "text"},
			{Name: "Status", Type: "text"},
			{Name: "Goal Contribution", Type: "text"},
		}},
		{Name: "Responsibilities", ContentType: "text", Required: true, Purpose: "What this container is accountable for"},
		{Name: "Complexity Assessment", ContentType: "text", Required: false, Purpose: "Known complexity and risks"},
	},
	"context": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "System-level objective"},
		{Name: "Containers", ContentType: "table", Required: true, Purpose: "Top-level deployment units", Columns: []ColumnDef{
			{Name: "ID", Type: "entity_id"},
			{Name: "Name", Type: "text"},
			{Name: "Boundary", Type: "text"},
			{Name: "Status", Type: "text"},
			{Name: "Responsibilities", Type: "text"},
			{Name: "Goal Contribution", Type: "text"},
		}},
		{Name: "Abstract Constraints", ContentType: "table", Required: true, Purpose: "System-wide architectural rules", Columns: []ColumnDef{
			{Name: "Constraint", Type: "text"},
			{Name: "Rationale", Type: "text"},
			{Name: "Affected Containers", Type: "text"},
		}},
	},
	"ref": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What problem this ref addresses"},
		{Name: "Choice", ContentType: "text", Required: true, Purpose: "The selected approach"},
		{Name: "Why", ContentType: "text", Required: true, Purpose: "Rationale for this choice"},
		{Name: "How", ContentType: "text", Required: false, Purpose: "Implementation guidance"},
	},
	"adr": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "Decision context and objective"},
	},
}

// ForType returns section definitions for an entity type, or nil if unknown.
func ForType(entityType string) []SectionDef {
	return Registry[entityType]
}

// PurposeOf returns the purpose string for a section within an entity type.
func PurposeOf(entityType, sectionName string) string {
	for _, s := range Registry[entityType] {
		if s.Name == sectionName {
			return s.Purpose
		}
	}
	return ""
}
