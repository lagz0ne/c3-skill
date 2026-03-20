package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunCodemap_ScaffoldsEntries(t *testing.T) {
	s := createRichDBFixture(t)
	c3Dir := filepath.Join(t.TempDir(), ".c3")
	os.MkdirAll(c3Dir, 0755)

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{C3Dir: c3Dir, Store: s, JSON: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Default output (no HUMAN env) is JSON
	var result CodemapResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON output: %v\nbody: %s", err, buf.String())
	}

	if len(result.Added) == 0 {
		t.Error("should have added code-map entries for components/refs")
	}

	// Verify code-map.yaml was written
	cmPath := filepath.Join(c3Dir, "code-map.yaml")
	data, err := os.ReadFile(cmPath)
	if err != nil {
		t.Fatal("code-map.yaml should be created")
	}
	content := string(data)
	if !strings.Contains(content, "# Components") {
		t.Error("code-map should have Components section")
	}
	if !strings.Contains(content, "# Refs") {
		t.Error("code-map should have Refs section")
	}
}

func TestRunCodemap_PreservesExisting(t *testing.T) {
	s := createRichDBFixture(t)
	// Pre-set code-map entry for c3-101
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	c3Dir := filepath.Join(t.TempDir(), ".c3")
	os.MkdirAll(c3Dir, 0755)

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{C3Dir: c3Dir, Store: s, JSON: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result CodemapResult
	json.Unmarshal(buf.Bytes(), &result)

	if !containsStr2(result.Existing, "c3-101") {
		t.Error("c3-101 should be in existing list")
	}
}

func TestRunCodemap_HumanOutput(t *testing.T) {
	s := createRichDBFixture(t)
	c3Dir := filepath.Join(t.TempDir(), ".c3")
	os.MkdirAll(c3Dir, 0755)

	t.Setenv("HUMAN", "1")

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{C3Dir: c3Dir, Store: s, JSON: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "code-map:") {
		t.Error("human output should contain 'code-map:'")
	}
	if !strings.Contains(output, "added:") {
		t.Error("human output should contain 'added:'")
	}
}

func TestRunCodemap_WithExcludes(t *testing.T) {
	s := createRichDBFixture(t)
	s.AddExclude("**/*.test.ts")

	c3Dir := filepath.Join(t.TempDir(), ".c3")
	os.MkdirAll(c3Dir, 0755)

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{C3Dir: c3Dir, Store: s}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(c3Dir, "code-map.yaml"))
	if !strings.Contains(string(data), "_exclude") {
		t.Error("code-map should contain _exclude section")
	}
}

func TestRunCodemap_WithRules(t *testing.T) {
	s := createRichDBFixture(t)
	// Add a rule entity
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging",
		Slug: "logging", Status: "active", Metadata: "{}",
	})

	c3Dir := filepath.Join(t.TempDir(), ".c3")
	os.MkdirAll(c3Dir, 0755)

	var buf bytes.Buffer
	err := RunCodemap(CodemapOptions{C3Dir: c3Dir, Store: s}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(c3Dir, "code-map.yaml"))
	if !strings.Contains(string(data), "# Rules") {
		t.Error("code-map should have Rules section")
	}
}
