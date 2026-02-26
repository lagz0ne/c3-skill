package templates

import (
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []string{"context.md", "container.md", "component.md", "ref.md", "adr-000.md"}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			content, err := Read(name)
			if err != nil {
				t.Fatalf("Read(%q) error: %v", name, err)
			}
			if content == "" {
				t.Errorf("Read(%q) returned empty", name)
			}
			if !strings.HasPrefix(content, "---\n") {
				t.Errorf("Read(%q) should start with frontmatter delimiters", name)
			}
		})
	}
}

func TestRead_NotFound(t *testing.T) {
	_, err := Read("nonexistent.md")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestRender(t *testing.T) {
	content, err := Render("context.md", map[string]string{
		"${PROJECT}": "MyProject",
		"${GOAL}":    "Build great software",
		"${SUMMARY}": "A project",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "MyProject") {
		t.Error("expected PROJECT to be replaced")
	}
	if !strings.Contains(content, "Build great software") {
		t.Error("expected GOAL to be replaced")
	}
	if strings.Contains(content, "${PROJECT}") {
		t.Error("placeholder ${PROJECT} should be replaced")
	}
}
