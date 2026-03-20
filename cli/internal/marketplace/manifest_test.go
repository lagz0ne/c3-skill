package marketplace

import (
	"testing"
)

func TestParseManifest(t *testing.T) {
	yaml := `name: go-patterns
description: Opinionated Go patterns
tags: [go, backend]
compatibility:
  languages: [go]
  frameworks: [gin]
rules:
  - id: rule-error-handling
    title: Structured Error Handling
    category: reliability
    tags: [errors]
    summary: Wrap errors with context
  - id: rule-config-loading
    title: Config from Environment
    category: operations
    tags: [config]
    summary: Single config struct
`
	m, err := ParseManifest([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if m.Name != "go-patterns" {
		t.Errorf("Name = %q, want %q", m.Name, "go-patterns")
	}
	if len(m.Rules) != 2 {
		t.Fatalf("len(Rules) = %d, want 2", len(m.Rules))
	}
	if m.Rules[0].ID != "rule-error-handling" {
		t.Errorf("Rules[0].ID = %q, want %q", m.Rules[0].ID, "rule-error-handling")
	}
	if m.Compatibility.Languages[0] != "go" {
		t.Errorf("Languages[0] = %q, want %q", m.Compatibility.Languages[0], "go")
	}
}

func TestParseManifestValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"missing name", "description: foo\nrules:\n  - id: rule-x\n    summary: x\n", "name is required"},
		{"missing rules", "name: foo\n", "at least one rule"},
		{"rule without id", "name: foo\nrules:\n  - summary: x\n", "rule[0]: id is required"},
		{"rule without summary", "name: foo\nrules:\n  - id: rule-x\n", "rule[0]: summary is required"},
		{"bad rule id prefix", "name: foo\nrules:\n  - id: ref-x\n    summary: x\n", "must start with \"rule-\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(tt.yaml))
			if err == nil {
				t.Fatal("expected error")
			}
			if !containsStr(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
