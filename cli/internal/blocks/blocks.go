package blocks

import (
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// Block represents an extracted section from a C3 entity body.
type Block struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Purpose string      `json:"purpose"`
	Filled  bool        `json:"filled"`
	Content any `json:"content"`
}

// ExtractBlocks parses all schema-defined sections from a markdown body.
// Returns one Block per section in the schema, whether present in the body or not.
func ExtractBlocks(body, entityType string) []Block {
	schemaSections := schema.ForType(entityType)
	if schemaSections == nil {
		return nil
	}

	bodySections := markdown.ParseSections(body)
	sectionMap := make(map[string]markdown.Section)
	for _, s := range bodySections {
		if s.Name != "" {
			sectionMap[s.Name] = s
		}
	}

	blocks := make([]Block, 0, len(schemaSections))
	for _, def := range schemaSections {
		b := Block{
			Name:    def.Name,
			Type:    def.ContentType,
			Purpose: def.Purpose,
		}

		bodySection, exists := sectionMap[def.Name]
		if !exists {
			// Section not in body — unfilled, nil content
			blocks = append(blocks, b)
			continue
		}

		content := strings.TrimSpace(bodySection.Content)

		if def.ContentType == "table" {
			b.Content, b.Filled = extractTable(content)
		} else {
			// text (includes code blocks embedded in text sections)
			if content != "" {
				b.Content = content
				b.Filled = true
			}
		}

		blocks = append(blocks, b)
	}

	return blocks
}

// ExtractBlock returns a single named block, or nil if the section is not in the schema.
// Accepts exact name or slug form (lowercase hyphenated).
func ExtractBlock(body, entityType, sectionName string) *Block {
	schemaSections := schema.ForType(entityType)
	if schemaSections == nil {
		return nil
	}

	// Find the schema section by exact name or slug
	var def *schema.SectionDef
	for i := range schemaSections {
		if schemaSections[i].Name == sectionName || slugify(schemaSections[i].Name) == sectionName {
			def = &schemaSections[i]
			break
		}
	}
	if def == nil {
		return nil
	}

	bodySections := markdown.ParseSections(body)
	var bodySection *markdown.Section
	for i := range bodySections {
		if bodySections[i].Name == def.Name {
			bodySection = &bodySections[i]
			break
		}
	}

	b := &Block{
		Name:    def.Name,
		Type:    def.ContentType,
		Purpose: def.Purpose,
	}

	if bodySection == nil {
		return b
	}

	content := strings.TrimSpace(bodySection.Content)

	if def.ContentType == "table" {
		b.Content, b.Filled = extractTable(content)
	} else {
		if content != "" {
			b.Content = content
			b.Filled = true
		}
	}

	return b
}

// extractTable parses table content and returns (rows, filled).
func extractTable(content string) (any, bool) {
	if content == "" {
		return []map[string]string{}, false
	}
	table, err := markdown.ParseTable(content)
	if err != nil {
		return []map[string]string{}, false
	}
	if len(table.Rows) == 0 {
		return []map[string]string{}, false
	}
	return table.Rows, true
}

// slugify converts "Related Refs" → "related-refs".
func slugify(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
