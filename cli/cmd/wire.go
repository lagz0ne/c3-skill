package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// citeTarget returns the section name and column name for a citation target.
func citeTarget(targetID string) (sectionName, colName string) {
	if strings.HasPrefix(targetID, "rule-") {
		return "Compliance Rules", "Rule"
	}
	return "Compliance Refs", "Ref"
}

// RunWire creates a cite relationship between source and target.
func RunWire(s *store.Store, sourceID, relationType, targetID string, w io.Writer) error {
	if relationType == "" {
		relationType = "cite"
	}
	if relationType != "cite" {
		return fmt.Errorf("unsupported relation type %q (only 'cite' supported)", relationType)
	}

	// Validate both entities exist
	srcEntity, err := s.GetEntity(sourceID)
	if err != nil {
		return fmt.Errorf("entity %q not found", sourceID)
	}
	if _, err := s.GetEntity(targetID); err != nil {
		return fmt.Errorf("entity %q not found", targetID)
	}

	// Add relationship in the store
	if err := s.AddRelationship(&store.Relationship{
		FromID:  sourceID,
		ToID:    targetID,
		RelType: "uses",
	}); err != nil {
		return fmt.Errorf("adding relationship: %w", err)
	}

	sectionName, colName, row := citeRow(srcEntity, targetID)
	if err := addTableRowIfAbsentStore(s, srcEntity, sectionName, colName, targetID, row); err != nil {
		return fmt.Errorf("body table update: %w", err)
	}

	fmt.Fprintf(w, "Wired %s -[cite]-> %s\n", sourceID, targetID)
	return nil
}

// RunUnwire removes a cite relationship from both sides.
func RunUnwire(s *store.Store, sourceID, relationType, targetID string, w io.Writer) error {
	if relationType == "" {
		relationType = "cite"
	}
	if relationType != "cite" {
		return fmt.Errorf("unsupported relation type %q (only 'cite' supported)", relationType)
	}

	srcEntity, err := s.GetEntity(sourceID)
	if err != nil {
		return fmt.Errorf("entity %q not found", sourceID)
	}
	if _, err := s.GetEntity(targetID); err != nil {
		return fmt.Errorf("entity %q not found", targetID)
	}

	// Remove relationship from the store
	if err := s.RemoveRelationship(&store.Relationship{
		FromID:  sourceID,
		ToID:    targetID,
		RelType: "uses",
	}); err != nil {
		return fmt.Errorf("removing relationship: %w", err)
	}

	// Update body table
	sectionName, colName, _ := citeRow(srcEntity, targetID)
	if err := removeTableRowStore(s, srcEntity, sectionName, colName, targetID); err != nil {
		return fmt.Errorf("body table update: %w", err)
	}

	fmt.Fprintf(w, "Unwired %s -[cite]-> %s\n", sourceID, targetID)
	return nil
}

func citeRow(source *store.Entity, targetID string) (sectionName, colName string, row map[string]string) {
	if source != nil && source.Type == "component" {
		targetType := "ref"
		if strings.HasPrefix(targetID, "rule-") {
			targetType = "rule"
		}
		return "Governance", "Reference", map[string]string{
			"Reference":  targetID,
			"Type":       targetType,
			"Governs":    "Compliance target added by c3x wire; refine what must be reviewed or complied with before handoff.",
			"Precedence": "wired compliance target beats uncited local prose",
			"Notes":      "Added by c3x wire for explicit compliance review.",
		}
	}
	sectionName, colName = citeTarget(targetID)
	return sectionName, colName, map[string]string{
		colName:        targetID,
		"Why required": "Added by c3x wire; fill why this target must be reviewed or complied with.",
		"Action":       "review-and-refine",
	}
}

// addTableRowIfAbsentStore adds a row to a section's table if no row with matchCol==matchVal exists.
func addTableRowIfAbsentStore(s *store.Store, entity *store.Entity, sectionName, matchCol, matchVal string, row map[string]string) error {
	// Read current content from node tree.
	body, err := content.ReadEntity(s, entity.ID)
	if err != nil || body == "" {
		return nil
	}

	// Check if row already exists
	table, err := markdown.ExtractTableFromSection(body, sectionName)
	if err != nil {
		// Section not found — skip silently
		return nil
	}
	if table != nil {
		for _, r := range table.Rows {
			if r[matchCol] == matchVal {
				return nil // already present
			}
		}
	}

	newBody, err := markdown.AppendTableRow(body, sectionName, row)
	if err != nil {
		return nil // section/table not found — skip
	}

	return content.WriteEntity(s, entity.ID, newBody)
}

// removeTableRowStore removes rows from a section's table where matchCol==matchVal.
func removeTableRowStore(s *store.Store, entity *store.Entity, sectionName, matchCol, matchVal string) error {
	// Read current content from node tree.
	body, err := content.ReadEntity(s, entity.ID)
	if err != nil || body == "" {
		return nil
	}

	table, err := markdown.ExtractTableFromSection(body, sectionName)
	if err != nil || table == nil {
		return nil // section/table not found — idempotent
	}

	var filtered []map[string]string
	for _, r := range table.Rows {
		if r[matchCol] != matchVal {
			filtered = append(filtered, r)
		}
	}

	if filtered == nil {
		filtered = []map[string]string{}
	}
	table.Rows = filtered

	newBody, err := markdown.SetTableInSection(body, sectionName, table)
	if err != nil {
		return err
	}

	return content.WriteEntity(s, entity.ID, newBody)
}
