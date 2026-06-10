package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunCodemap_ScaffoldsEntries(t *testing.T) {
	s := createRichDBFixture(t)

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{Store: s, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result CodemapResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected explicit JSON output: %v\nbody: %s", err, buf.String())
	}

	if len(result.Added) == 0 {
		t.Error("should have added code-map entries for components/refs")
	}

	for _, id := range result.Added {
		patterns, err := s.CodeMapFor(id)
		if err != nil {
			t.Fatalf("CodeMapFor(%s): %v", id, err)
		}
		if len(patterns) > 0 {
			t.Errorf("scaffolded entry %s should have empty patterns, got %v", id, patterns)
		}
	}
}

func TestRunCodemap_DefaultOutputTOON(t *testing.T) {
	s := createRichDBFixture(t)

	var buf bytes.Buffer
	if err := RunCodemap(CodemapOptions{Store: s}, &buf); err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(buf.String())
	if strings.HasPrefix(out, "{") {
		t.Fatalf("default codemap output must not be JSON:\n%s", out)
	}
	if !strings.Contains(out, "added:") || !strings.Contains(out, "existing:") {
		t.Fatalf("default codemap output should be TOON object output, got:\n%s", out)
	}
}

func TestRunCodemap_PreservesExisting(t *testing.T) {
	s := createRichDBFixture(t)
	// Pre-set code-map entry for c3-101
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{Store: s, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result CodemapResult
	json.Unmarshal(buf.Bytes(), &result)

	if !containsStr2(result.Existing, "c3-101") {
		t.Error("c3-101 should be in existing list")
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 1 || patterns[0] != "src/auth/**" {
		t.Errorf("expected preserved pattern, got %v", patterns)
	}
}

func TestRunCodemap_HumanOutput(t *testing.T) {
	s := createRichDBFixture(t)

	t.Setenv("HUMAN", "1")

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{Store: s, JSON: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "codemap scaffolded") {
		t.Error("human output should contain 'codemap scaffolded'")
	}
	if !strings.Contains(output, "added:") {
		t.Error("human output should contain 'added:'")
	}
}

func TestRunCodemap_WithExcludes(t *testing.T) {
	s := createRichDBFixture(t)
	s.AddExclude("**/*.test.ts")

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{Store: s, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	excludes, _ := s.Excludes()
	if len(excludes) != 1 || excludes[0] != "**/*.test.ts" {
		t.Errorf("expected exclude pattern in store, got %v", excludes)
	}
}

func TestRunCodemap_WithRules(t *testing.T) {
	s := createRichDBFixture(t)
	// Add a rule entity
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging",
		Slug: "logging", Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{Store: s, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result CodemapResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected explicit JSON output: %v\nbody: %s", err, buf.String())
	}
	if !containsStr2(result.Added, "rule-logging") {
		t.Error("rule-logging should be in added list")
	}
}
