package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/writer"
)

// RunWire creates a bidirectional cite relationship between source and target.
// Three sides: (1) source frontmatter refs[], (2) source "Related Refs" table, (3) target "Cited By" table.
func RunWire(c3Dir, sourceID, relationType, targetID string, w io.Writer) error {
	if relationType != "cite" {
		return fmt.Errorf("unsupported relation type %q (only 'cite' supported)", relationType)
	}

	// Resolve file paths
	srcPath, err := findEntityFile(c3Dir, sourceID)
	if err != nil {
		return err
	}
	tgtPath, err := findEntityFile(c3Dir, targetID)
	if err != nil {
		return err
	}

	// Side 1: Add target to source's frontmatter refs[]
	if err := writer.AddToArrayField(srcPath, "refs", targetID); err != nil {
		return fmt.Errorf("side 1 (refs): %w", err)
	}

	// Side 2: Add row to source's "Related Refs" table
	if err := addTableRowIfAbsent(srcPath, "Related Refs", "Ref", targetID, map[string]string{
		"Ref":  targetID,
		"Role": "",
	}); err != nil {
		return fmt.Errorf("side 2 (Related Refs): %w", err)
	}

	// Side 3: Add row to target's "Cited By" table
	if err := addTableRowIfAbsent(tgtPath, "Cited By", "Component", sourceID, map[string]string{
		"Component": sourceID,
		"Usage":     "",
	}); err != nil {
		return fmt.Errorf("side 3 (Cited By): %w", err)
	}

	fmt.Fprintf(w, "Wired %s -[cite]-> %s\n", sourceID, targetID)
	return nil
}

// RunUnwire removes a cite relationship from all three sides.
func RunUnwire(c3Dir, sourceID, relationType, targetID string, w io.Writer) error {
	if relationType != "cite" {
		return fmt.Errorf("unsupported relation type %q (only 'cite' supported)", relationType)
	}

	srcPath, err := findEntityFile(c3Dir, sourceID)
	if err != nil {
		return err
	}
	tgtPath, err := findEntityFile(c3Dir, targetID)
	if err != nil {
		return err
	}

	// Side 1: Remove target from source's frontmatter refs[]
	if err := writer.RemoveFromArrayField(srcPath, "refs", targetID); err != nil {
		return fmt.Errorf("side 1 (refs): %w", err)
	}

	// Side 2: Remove row from source's "Related Refs" table where Ref matches target
	if err := removeTableRow(srcPath, "Related Refs", "Ref", targetID); err != nil {
		return fmt.Errorf("side 2 (Related Refs): %w", err)
	}

	// Side 3: Remove row from target's "Cited By" table where Component matches source
	if err := removeTableRow(tgtPath, "Cited By", "Component", sourceID); err != nil {
		return fmt.Errorf("side 3 (Cited By): %w", err)
	}

	fmt.Fprintf(w, "Unwired %s -[cite]-> %s\n", sourceID, targetID)
	return nil
}

// addTableRowIfAbsent adds a row to a section's table if no row with matchCol==matchVal exists.
func addTableRowIfAbsent(filePath, sectionName, matchCol, matchVal string, row map[string]string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fm, body := frontmatter.ParseFrontmatter(string(data))
	if fm == nil {
		return fmt.Errorf("no frontmatter in %s", filePath)
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

	return writeEntityFile(filePath, fm, newBody)
}

// removeTableRow removes rows from a section's table where matchCol==matchVal.
func removeTableRow(filePath, sectionName, matchCol, matchVal string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fm, body := frontmatter.ParseFrontmatter(string(data))
	if fm == nil {
		return fmt.Errorf("no frontmatter in %s", filePath)
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

	return writeEntityFile(filePath, fm, newBody)
}
