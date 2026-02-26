package templates

import (
	"embed"
	"strings"
)

//go:embed all:*.md
var fs embed.FS

// Read returns the content of an embedded template file.
func Read(name string) (string, error) {
	data, err := fs.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Render reads a template and replaces placeholders.
// Replacements is a map of "${KEY}" -> "value".
func Render(name string, replacements map[string]string) (string, error) {
	content, err := Read(name)
	if err != nil {
		return "", err
	}
	for k, v := range replacements {
		content = strings.ReplaceAll(content, k, v)
	}
	return content, nil
}
