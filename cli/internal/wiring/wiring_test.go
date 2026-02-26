package wiring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddComponentToContainerTable(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		compID    string
		compName  string
		category  string
		goal      string
		wantRow   string
		wantFound bool
	}{
		{
			name: "append to existing table",
			input: `---
id: c3-1
title: API Gateway
type: container
---
# API Gateway

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|--------------------|
| c3-101 | Auth | foundation | active | Authentication |

## Notes
`,
			compID:    "c3-110",
			compName:  "Users",
			category:  "feature",
			goal:      "User management",
			wantRow:   "| c3-110 | Users | feature | active | User management |",
			wantFound: true,
		},
		{
			name: "table followed by heading",
			input: `| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|--------------------|
| c3-101 | Auth | foundation | active | Auth |

## Next Section
`,
			compID:    "c3-102",
			compName:  "Logging",
			category:  "foundation",
			goal:      "Structured logging",
			wantRow:   "| c3-102 | Logging | foundation | active | Structured logging |",
			wantFound: true,
		},
		{
			name: "no table present",
			input: `---
id: c3-1
title: API Gateway
---
No table here.
`,
			compID:    "c3-110",
			compName:  "Users",
			category:  "feature",
			goal:      "User management",
			wantFound: false,
		},
		{
			name: "table at end of file",
			input: `| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|--------------------|
| c3-101 | Auth | foundation | active | Auth |
`,
			compID:    "c3-102",
			compName:  "Cache",
			category:  "foundation",
			goal:      "Caching layer",
			wantRow:   "| c3-102 | Cache | foundation | active | Caching layer |",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			fp := filepath.Join(tmp, "README.md")
			if err := os.WriteFile(fp, []byte(tt.input), 0644); err != nil {
				t.Fatal(err)
			}

			found := AddComponentToContainerTable(fp, tt.compID, tt.compName, tt.category, tt.goal)

			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}

			if !tt.wantFound {
				return
			}

			content, err := os.ReadFile(fp)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(content), tt.wantRow) {
				t.Errorf("output does not contain expected row %q\ngot:\n%s", tt.wantRow, string(content))
			}
		})
	}
}

func TestAddComponentToContainerTable_PreservesStructure(t *testing.T) {
	input := `---
id: c3-1
title: API Gateway
type: container
---
# API Gateway

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|--------------------|
| c3-101 | Auth | foundation | active | Authentication |

## Architecture Notes

Some notes here.
`

	tmp := t.TempDir()
	fp := filepath.Join(tmp, "README.md")
	if err := os.WriteFile(fp, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	AddComponentToContainerTable(fp, "c3-110", "Users", "feature", "User management")

	content, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// Should still have the Architecture Notes section
	if !strings.Contains(result, "## Architecture Notes") {
		t.Error("Architecture Notes section was lost")
	}

	// New row should appear before the next section
	authIdx := strings.Index(result, "| c3-101 |")
	usersIdx := strings.Index(result, "| c3-110 |")
	notesIdx := strings.Index(result, "## Architecture Notes")

	if authIdx >= usersIdx {
		t.Error("new row should come after existing row")
	}
	if usersIdx >= notesIdx {
		t.Error("new row should come before Architecture Notes section")
	}
}
