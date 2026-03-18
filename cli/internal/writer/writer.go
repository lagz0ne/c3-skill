package writer

import (
	"fmt"
	"os"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// Allowed field sets.
var stringFields = map[string]bool{
	"goal": true, "summary": true, "status": true,
	"boundary": true, "category": true, "title": true, "date": true,
}

var arrayFields = map[string]bool{
	"uses": true, "refs": true, "affects": true, "scope": true, "sources": true,
}

// SetField reads a markdown file, updates a single string frontmatter field,
// and writes back preserving the body exactly.
func SetField(path string, field string, value string) error {
	if !stringFields[field] {
		return fmt.Errorf("unknown string field: %q", field)
	}

	fm, body, err := readAndParse(path)
	if err != nil {
		return err
	}

	setStringField(fm, field, value)
	return WriteBack(path, fm, body)
}

// AddToArrayField adds a value to an array frontmatter field (no duplicates).
func AddToArrayField(path string, field string, value string) error {
	if !arrayFields[field] {
		return fmt.Errorf("unknown array field: %q", field)
	}

	fm, body, err := readAndParse(path)
	if err != nil {
		return err
	}

	arr := getArrayField(fm, field)
	for _, v := range arr {
		if v == value {
			return WriteBack(path, fm, body) // already present
		}
	}
	arr = append(arr, value)
	setArrayField(fm, field, arr)
	return WriteBack(path, fm, body)
}

// RemoveFromArrayField removes a value from an array frontmatter field.
func RemoveFromArrayField(path string, field string, value string) error {
	if !arrayFields[field] {
		return fmt.Errorf("unknown array field: %q", field)
	}

	fm, body, err := readAndParse(path)
	if err != nil {
		return err
	}

	arr := getArrayField(fm, field)
	var filtered []string
	for _, v := range arr {
		if v != value {
			filtered = append(filtered, v)
		}
	}
	setArrayField(fm, field, filtered)
	return WriteBack(path, fm, body)
}

// readAndParse reads a file and returns parsed frontmatter + body.
func readAndParse(path string) (*frontmatter.Frontmatter, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	fm, body := frontmatter.ParseFrontmatter(string(data))
	if fm == nil {
		return nil, "", fmt.Errorf("no valid frontmatter in %s", path)
	}
	return fm, body, nil
}

// WriteBack serializes frontmatter + body back to file.
func WriteBack(path string, fm *frontmatter.Frontmatter, body string) error {
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}

	content := "---\n" + string(yamlBytes) + "---\n" + body
	return os.WriteFile(path, []byte(content), 0644)
}

func setStringField(fm *frontmatter.Frontmatter, field, value string) {
	switch field {
	case "goal":
		fm.Goal = value
	case "summary":
		fm.Summary = value
	case "status":
		fm.Status = value
	case "boundary":
		fm.Boundary = value
	case "category":
		fm.Category = value
	case "title":
		fm.Title = value
	case "date":
		fm.Date = value
	}
}

func getArrayField(fm *frontmatter.Frontmatter, field string) []string {
	switch field {
	case "uses", "refs":
		return fm.Refs
	case "affects":
		return fm.Affects
	case "scope":
		return fm.Scope
	case "sources":
		return fm.Sources
	}
	return nil
}

func setArrayField(fm *frontmatter.Frontmatter, field string, arr []string) {
	switch field {
	case "uses", "refs":
		fm.Refs = arr
	case "affects":
		fm.Affects = arr
	case "scope":
		fm.Scope = arr
	case "sources":
		fm.Sources = arr
	}
}
