package wiring

import (
	"fmt"
	"os"
	"regexp"
)

var tablePattern = regexp.MustCompile(`(?s)(\| ID \| Name \| Category \| Status \| Goal Contribution \|[\s\S]*?)(\n\n|\n##|\n---|\z)`)

// AddComponentToContainerTable inserts a component row into a container's markdown table.
// Returns true if the table was found and modified.
func AddComponentToContainerTable(containerReadmePath, componentId, name, category, goal string) bool {
	data, err := os.ReadFile(containerReadmePath)
	if err != nil {
		return false
	}
	content := string(data)

	loc := tablePattern.FindStringSubmatchIndex(content)
	if loc == nil {
		return false
	}

	// loc[2:4] = group 1 (table content), loc[4:6] = group 2 (terminator)
	tableEnd := loc[3]   // end of group 1
	termStart := loc[4]  // start of group 2

	newRow := fmt.Sprintf("| %s | %s | %s | active | %s |\n", componentId, name, category, goal)

	result := content[:tableEnd] + newRow + content[termStart:]

	if err := os.WriteFile(containerReadmePath, []byte(result), 0644); err != nil {
		return false
	}
	return true
}
