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
		{Name: "Parent Fit", ContentType: "table", Required: true, Purpose: "How this component fits top-down into the parent container", Columns: []ColumnDef{
			{Name: "Field", Type: "text"},
			{Name: "Value", Type: "text"},
		}},
		{Name: "Purpose", ContentType: "text", Required: true, Purpose: "Concrete ownership and non-goals"},
		{Name: "Foundational Flow", ContentType: "table", Required: true, Purpose: "Preconditions, inputs, state, and shared dependencies", Columns: []ColumnDef{
			{Name: "Aspect", Type: "text"},
			{Name: "Detail", Type: "text"},
			{Name: "Reference", Type: "text"},
		}},
		{Name: "Business Flow", ContentType: "table", Required: true, Purpose: "Business outcome, primary path, alternates, and failure behavior", Columns: []ColumnDef{
			{Name: "Aspect", Type: "text"},
			{Name: "Detail", Type: "text"},
			{Name: "Reference", Type: "text"},
		}},
		{Name: "Governance", ContentType: "table", Required: true, Purpose: "Refs, rules, ADRs, specs, and precedence governing this component", Columns: []ColumnDef{
			{Name: "Reference", Type: "text"},
			{Name: "Type", Type: "enum", Values: []string{"ref", "rule", "adr", "spec", "policy", "example", "N.A - <reason>"}},
			{Name: "Governs", Type: "text"},
			{Name: "Precedence", Type: "text"},
			{Name: "Notes", Type: "text"},
		}},
		{Name: "Contract", ContentType: "table", Required: true, Purpose: "Behavior surfaces that downstream code/material must honor", Columns: []ColumnDef{
			{Name: "Surface", Type: "text"},
			{Name: "Direction", Type: "enum", Values: []string{"IN", "OUT", "IN/OUT", "N.A - <reason>"}},
			{Name: "Contract", Type: "text"},
			{Name: "Boundary", Type: "text"},
			{Name: "Evidence", Type: "text"},
		}},
		{Name: "Change Safety", ContentType: "table", Required: true, Purpose: "Risks, triggers, detection, and verification required before done", Columns: []ColumnDef{
			{Name: "Risk", Type: "text"},
			{Name: "Trigger", Type: "text"},
			{Name: "Detection", Type: "text"},
			{Name: "Required Verification", Type: "text"},
		}},
		{Name: "Derived Materials", ContentType: "table", Required: true, Purpose: "Code, config, tests, docs, prompts, or assets that must derive from this component", Columns: []ColumnDef{
			{Name: "Material", Type: "text"},
			{Name: "Must derive from", Type: "text"},
			{Name: "Allowed variance", Type: "text"},
			{Name: "Evidence", Type: "text"},
		}},
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
		{Name: "How", ContentType: "text", Required: false, Purpose: "Golden pattern — prescriptive examples and implementation guidance"},
	},
	"rule": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What standard this rule enforces"},
		{Name: "Rule", ContentType: "text", Required: true, Purpose: "One-line statement of what must be true"},
		{Name: "Golden Example", ContentType: "text", Required: true, Purpose: "Canonical code showing the correct pattern"},
		{Name: "Not This", ContentType: "table", Required: false, Purpose: "Anti-patterns with why they're wrong here", Columns: []ColumnDef{
			{Name: "Anti-Pattern", Type: "text"},
			{Name: "Correct", Type: "text"},
			{Name: "Why Wrong Here", Type: "text"},
		}},
		{Name: "Scope", ContentType: "text", Required: false, Purpose: "Where this rule applies and doesn't"},
		{Name: "Override", ContentType: "text", Required: false, Purpose: "How to deviate from this rule when justified"},
	},
	"adr": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "Decision context and objective"},
	},
	"recipe": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What cross-cutting concern this traces"},
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
