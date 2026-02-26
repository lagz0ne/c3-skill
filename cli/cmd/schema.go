package cmd

import (
	"encoding/json"
	"fmt"
	"io"
)

// SchemaOutput is the JSON structure returned by the schema command.
type SchemaOutput struct {
	Type     string       `json:"type"`
	Sections []SectionDef `json:"sections"`
}

// SectionDef defines a known section for an entity type.
type SectionDef struct {
	Name        string      `json:"name"`
	ContentType string      `json:"content_type"`
	Required    bool        `json:"required"`
	Columns     []ColumnDef `json:"columns,omitempty"`
}

// ColumnDef defines a typed column within a table section.
type ColumnDef struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []string `json:"values,omitempty"`
}

// schemaRegistry maps entity types to their ordered section definitions.
var schemaRegistry = map[string][]SectionDef{
	"component": {
		{Name: "Goal", ContentType: "text", Required: true},
		{Name: "Dependencies", ContentType: "table", Required: true, Columns: []ColumnDef{
			{Name: "Direction", Type: "enum", Values: []string{"IN", "OUT"}},
			{Name: "What", Type: "text"},
			{Name: "From/To", Type: "entity_id"},
		}},
		{Name: "Related Refs", ContentType: "table", Required: false, Columns: []ColumnDef{
			{Name: "Ref", Type: "ref_id"},
			{Name: "Role", Type: "text"},
		}},
		{Name: "Container Connection", ContentType: "text", Required: false},
	},
	"container": {
		{Name: "Goal", ContentType: "text", Required: true},
		{Name: "Components", ContentType: "table", Required: true, Columns: []ColumnDef{
			{Name: "ID", Type: "entity_id"},
			{Name: "Name", Type: "text"},
			{Name: "Category", Type: "text"},
			{Name: "Status", Type: "text"},
			{Name: "Goal Contribution", Type: "text"},
		}},
		{Name: "Responsibilities", ContentType: "text", Required: true},
		{Name: "Complexity Assessment", ContentType: "text", Required: false},
	},
	"context": {
		{Name: "Goal", ContentType: "text", Required: true},
		{Name: "Containers", ContentType: "table", Required: true, Columns: []ColumnDef{
			{Name: "ID", Type: "entity_id"},
			{Name: "Name", Type: "text"},
			{Name: "Boundary", Type: "text"},
			{Name: "Status", Type: "text"},
			{Name: "Responsibilities", Type: "text"},
			{Name: "Goal Contribution", Type: "text"},
		}},
		{Name: "Abstract Constraints", ContentType: "table", Required: true, Columns: []ColumnDef{
			{Name: "Constraint", Type: "text"},
			{Name: "Rationale", Type: "text"},
			{Name: "Affected Containers", Type: "text"},
		}},
	},
	"ref": {
		{Name: "Goal", ContentType: "text", Required: true},
		{Name: "Choice", ContentType: "text", Required: true},
		{Name: "Why", ContentType: "text", Required: true},
		{Name: "How", ContentType: "text", Required: false},
	},
	"adr": {
		{Name: "Goal", ContentType: "text", Required: true},
	},
}

// RunSchema outputs the section schema for a given entity type.
func RunSchema(entityType string, jsonOutput bool, w io.Writer) error {
	sections, ok := schemaRegistry[entityType]
	if !ok {
		return fmt.Errorf("unknown entity type: %q", entityType)
	}

	if jsonOutput {
		out := SchemaOutput{
			Type:     entityType,
			Sections: sections,
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	// Text output
	fmt.Fprintf(w, "Schema: %s\n", entityType)
	for _, s := range sections {
		req := ""
		if s.Required {
			req = " (required)"
		}
		fmt.Fprintf(w, "  %s [%s]%s\n", s.Name, s.ContentType, req)
		for _, col := range s.Columns {
			fmt.Fprintf(w, "    - %s (%s)\n", col.Name, col.Type)
		}
	}
	return nil
}
