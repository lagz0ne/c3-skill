package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

// =============================================================================
// Richer add: accept content at creation time via AddOptions
// =============================================================================

func TestRunAddRich_ContainerWithGoal(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "container",
		Slug:       "payments",
		C3Dir:      c3Dir,
		Graph:      graph,
		Goal:       "Process payments securely",
		Summary:    "Handles all payment processing",
		Boundary:   "service",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify frontmatter has goal populated
	dirPath := filepath.Join(c3Dir, "c3-3-payments")
	content, err := os.ReadFile(filepath.Join(dirPath, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	fm, body := frontmatter.ParseFrontmatter(string(content))
	if fm == nil {
		t.Fatal("frontmatter should exist")
	}
	if fm.Goal != "Process payments securely" {
		t.Errorf("goal = %q, want %q", fm.Goal, "Process payments securely")
	}
	if fm.Summary != "Handles all payment processing" {
		t.Errorf("summary = %q", fm.Summary)
	}
	if fm.Boundary != "service" {
		t.Errorf("boundary = %q, want %q", fm.Boundary, "service")
	}

	// Body should contain Goal section populated from opts
	if !strings.Contains(body, "## Goal") {
		t.Error("body should contain Goal section")
	}
	if !strings.Contains(body, "Process payments securely") {
		t.Error("body Goal section should contain the goal text")
	}
	// Body should contain Components table (empty scaffold)
	if !strings.Contains(body, "## Components") {
		t.Error("body should contain Components table section")
	}
}

func TestRunAddRich_ComponentWithGoal(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "component",
		Slug:       "rate-limiter",
		C3Dir:      c3Dir,
		Graph:      graph,
		Container:  "c3-1",
		Feature:    false,
		Goal:       "Throttle API requests",
		Summary:    "Token bucket rate limiting",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Find the created file (c3-102-rate-limiter.md)
	filePath := filepath.Join(c3Dir, "c3-1-api", "c3-102-rate-limiter.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("component file not created: %v", err)
	}

	fm, body := frontmatter.ParseFrontmatter(string(content))
	if fm.Goal != "Throttle API requests" {
		t.Errorf("goal = %q, want %q", fm.Goal, "Throttle API requests")
	}
	if fm.Summary != "Token bucket rate limiting" {
		t.Errorf("summary = %q", fm.Summary)
	}

	// Body should contain scaffolded sections
	if !strings.Contains(body, "## Goal") {
		t.Error("body should contain Goal section")
	}
	if !strings.Contains(body, "Throttle API requests") {
		t.Error("body Goal section should contain the goal text")
	}
	if !strings.Contains(body, "## Dependencies") {
		t.Error("body should contain Dependencies table scaffold")
	}
	if !strings.Contains(body, "## Code References") {
		t.Error("body should contain Code References table scaffold")
	}

	// Container's Components table should now list the new component
	containerContent, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile container: %v", err)
	}
	if !strings.Contains(string(containerContent), "rate-limiter") {
		t.Error("container Components table should include the new component")
	}
}

func TestRunAddRich_RefWithGoal(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "ref",
		Slug:       "error-handling",
		C3Dir:      c3Dir,
		Graph:      graph,
		Goal:       "Consistent error responses across all services",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	filePath := filepath.Join(c3Dir, "refs", "ref-error-handling.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	fm, body := frontmatter.ParseFrontmatter(string(content))
	if fm.Goal != "Consistent error responses across all services" {
		t.Errorf("goal = %q", fm.Goal)
	}
	// Ref body should contain required sections
	if !strings.Contains(body, "## Goal") {
		t.Error("ref body should contain Goal section")
	}
	if !strings.Contains(body, "## Choice") {
		t.Error("ref body should contain Choice section scaffold")
	}
	if !strings.Contains(body, "## Why") {
		t.Error("ref body should contain Why section scaffold")
	}
	if !strings.Contains(body, "## Cited By") {
		t.Error("ref body should contain Cited By table scaffold")
	}
}

func TestRunAddRich_AdrWithGoal(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "adr",
		Slug:       "use-grpc",
		C3Dir:      c3Dir,
		Graph:      graph,
		Goal:       "Migrate inter-service communication to gRPC",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Find the created ADR
	entries, err := os.ReadDir(filepath.Join(c3Dir, "adr"))
	if err != nil {
		t.Fatalf("ReadDir adr: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "use-grpc") {
			content, err := os.ReadFile(filepath.Join(c3Dir, "adr", e.Name()))
			if err != nil {
				t.Fatalf("ReadFile %s: %v", e.Name(), err)
			}
			fm, body := frontmatter.ParseFrontmatter(string(content))
			if fm == nil {
				t.Fatal("ADR frontmatter should be parseable")
			}
			// Check goal in frontmatter or body
			if strings.Contains(body, "Migrate inter-service") || fm.Goal == "Migrate inter-service communication to gRPC" {
				found = true
			}
		}
	}
	if !found {
		t.Error("ADR should contain the provided goal")
	}
}

func TestRunAddRich_FallsBackToLegacy(t *testing.T) {
	// When no content flags are provided, should work exactly like current add
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "container",
		Slug:       "worker",
		C3Dir:      c3Dir,
		Graph:      graph,
		// No Goal, Summary, etc.
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	dirPath := filepath.Join(c3Dir, "c3-3-worker")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("container should be created even without content flags")
	}
}
