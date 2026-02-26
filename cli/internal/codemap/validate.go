package codemap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Issue represents a validation finding from code-map.
type Issue struct {
	Severity string
	Entity   string
	Message  string
}

// Validate checks a CodeMap against known entities and filesystem.
// entities maps entity ID → type string ("component", "container", "ref", etc.).
func Validate(cm CodeMap, entities map[string]string, projectDir string) []Issue {
	var issues []Issue

	for id, files := range cm {
		if typ, ok := entities[id]; !ok {
			issues = append(issues, Issue{
				Severity: "error",
				Entity:   id,
				Message:  fmt.Sprintf("ID %q not found in C3 graph", id),
			})
		} else if typ != "component" {
			issues = append(issues, Issue{
				Severity: "warning",
				Entity:   id,
				Message:  fmt.Sprintf("ID %q is not a component (type: %s)", id, typ),
			})
		}

		for _, f := range files {
			p := strings.TrimSpace(f)
			if p == "" {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   id,
					Message:  "empty path in code-map",
				})
				continue
			}

			if filepath.IsAbs(p) {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   id,
					Message:  fmt.Sprintf("absolute path not allowed in code-map: %s", p),
				})
				continue
			}

			if strings.HasPrefix(filepath.Clean(p), "..") {
				issues = append(issues, Issue{
					Severity: "warning",
					Entity:   id,
					Message:  fmt.Sprintf("path escapes project root: %s", p),
				})
				continue
			}

			if projectDir != "" {
				full := filepath.Join(projectDir, p)
				info, err := os.Stat(full)
				if err != nil {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   id,
						Message:  fmt.Sprintf("file %q does not exist", f),
					})
				} else if !info.Mode().IsRegular() {
					issues = append(issues, Issue{
						Severity: "warning",
						Entity:   id,
						Message:  fmt.Sprintf("path is a directory, not a file: %s", f),
					})
				}
			}
		}
	}

	return issues
}
