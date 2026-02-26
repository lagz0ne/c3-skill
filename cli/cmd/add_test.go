package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAdd_Container(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("container", "payments", c3Dir, graph, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd container failed: %v", err)
	}

	// Container 3 should be created (1 and 2 already exist)
	dirPath := filepath.Join(c3Dir, "c3-3-payments")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("expected c3-3-payments directory to exist")
	}

	readmePath := filepath.Join(dirPath, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "id: c3-3") {
		t.Error("container README should have id c3-3")
	}
	if !strings.Contains(s, "type: container") {
		t.Error("container README should have type: container")
	}

	output := buf.String()
	if !strings.Contains(output, "Created:") {
		t.Error("should print Created message")
	}
	if !strings.Contains(output, "c3-3") {
		t.Errorf("output should mention c3-3: %s", output)
	}
}

func TestRunAdd_Component(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("component", "logging", c3Dir, graph, "c3-1", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd component failed: %v", err)
	}

	// c3-101 already exists (auth), so next foundation should be c3-102
	filePath := filepath.Join(c3Dir, "c3-1-api", "c3-102-logging.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("component file not created: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "id: c3-102") {
		t.Errorf("component should have id c3-102, content: %s", s)
	}
	if !strings.Contains(s, "type: component") {
		t.Error("component should have type: component")
	}
	if !strings.Contains(s, "category: foundation") {
		t.Error("component should be foundation category")
	}
}

func TestRunAdd_ComponentFeature(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("component", "checkout", c3Dir, graph, "c3-1", true, &buf)
	if err != nil {
		t.Fatalf("RunAdd feature component failed: %v", err)
	}

	// c3-110 already exists (users), so next feature should be c3-111
	filePath := filepath.Join(c3Dir, "c3-1-api", "c3-111-checkout.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("feature component file not created: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "id: c3-111") {
		t.Errorf("component should have id c3-111, content: %s", s)
	}
	if !strings.Contains(s, "category: feature") {
		t.Error("component should be feature category")
	}
}

func TestRunAdd_ComponentWiresContainerTable(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	RunAdd("component", "logging", c3Dir, graph, "c3-1", false, &buf)

	containerReadme, err := os.ReadFile(filepath.Join(c3Dir, "c3-1-api", "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	s := string(containerReadme)
	if !strings.Contains(s, "c3-102") {
		t.Error("container README should be updated with new component")
	}
	if !strings.Contains(s, "logging") {
		t.Error("container README should mention component name")
	}

	output := buf.String()
	if !strings.Contains(output, "Updated:") {
		t.Error("should print Updated message for container table")
	}
}

func TestRunAdd_ComponentMissingContainer(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("component", "orphan", c3Dir, graph, "", false, &buf)
	if err == nil {
		t.Fatal("expected error when --container is missing")
	}
	if !strings.Contains(err.Error(), "--container") {
		t.Errorf("error should mention --container: %v", err)
	}
}

func TestRunAdd_ComponentContainerNotFound(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("component", "orphan", c3Dir, graph, "c3-99", false, &buf)
	if err == nil {
		t.Fatal("expected error when container doesn't exist")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunAdd_Ref(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("ref", "rate-limiting", c3Dir, graph, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd ref failed: %v", err)
	}

	filePath := filepath.Join(c3Dir, "refs", "ref-rate-limiting.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "id: ref-rate-limiting") {
		t.Error("ref should have id ref-rate-limiting")
	}

	output := buf.String()
	if !strings.Contains(output, "ref-rate-limiting") {
		t.Errorf("output should mention ref id: %s", output)
	}
}

func TestRunAdd_RefDuplicate(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	// ref-jwt already exists in fixture
	err := RunAdd("ref", "jwt", c3Dir, graph, "", false, &buf)
	if err == nil {
		t.Fatal("expected error when ref already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists': %v", err)
	}
}

func TestRunAdd_Adr(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("adr", "oauth-support", c3Dir, graph, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd adr failed: %v", err)
	}

	// Check that a file was created in adr/
	entries, _ := os.ReadDir(filepath.Join(c3Dir, "adr"))
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "oauth-support") {
			found = true

			content, _ := os.ReadFile(filepath.Join(c3Dir, "adr", e.Name()))
			s := string(content)
			if !strings.Contains(s, "oauth-support") {
				t.Error("ADR should contain slug in id")
			}
			if strings.Contains(s, "${DATE}") {
				t.Error("ADR should have ${DATE} replaced")
			}
		}
	}
	if !found {
		t.Error("expected ADR file with oauth-support slug")
	}
}

func TestRunAdd_InvalidSlug(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("container", "Invalid_Slug", c3Dir, graph, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
	if !strings.Contains(err.Error(), "invalid slug") {
		t.Errorf("error should mention 'invalid slug': %v", err)
	}
}

func TestRunAdd_UnknownType(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("widget", "test", c3Dir, graph, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for unknown entity type")
	}
	if !strings.Contains(err.Error(), "unknown entity type") {
		t.Errorf("error should mention 'unknown entity type': %v", err)
	}
}

func TestRunAdd_MissingArgs(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunAdd("", "", c3Dir, graph, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error should mention usage: %v", err)
	}
}

func TestRunAdd_SequentialContainers(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	// Add container 3
	RunAdd("container", "payments", c3Dir, graph, "", false, &buf)

	// Reload graph to pick up the new container
	graph = loadGraph(t, c3Dir)
	buf.Reset()

	// Add container 4
	err := RunAdd("container", "worker", c3Dir, graph, "", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(c3Dir, "c3-4-worker")); os.IsNotExist(err) {
		t.Error("expected c3-4-worker directory")
	}
}
