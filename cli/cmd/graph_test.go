package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGraph_TextDepth1(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-1",
		Depth:     1,
		Direction: "",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Root entity should appear
	if !strings.Contains(output, "c3-1 (container)") {
		t.Errorf("should contain root entity, got:\n%s", output)
	}

	// Direct children should appear (depth 1 forward from container)
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("should contain child c3-101, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-110 (component)") {
		t.Errorf("should contain child c3-110, got:\n%s", output)
	}

	// Should show parent relationship
	if !strings.Contains(output, "parent: c3-1") {
		t.Errorf("should show parent for components, got:\n%s", output)
	}
}

func TestRunGraph_TextDepth0(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-101",
		Depth:     0,
		Direction: "",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Only the root entity
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("should contain entity, got:\n%s", output)
	}

	// Should NOT contain other entities
	if strings.Contains(output, "c3-1 (container)") {
		t.Errorf("depth 0 should not include parent, got:\n%s", output)
	}
}

func TestRunGraph_TextReverse(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	// Reverse from c3-1: find entities that point TO c3-1 (children, ref-jwt scope)
	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-1",
		Depth:     1,
		Direction: "reverse",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Root
	if !strings.Contains(output, "c3-1 (container)") {
		t.Errorf("should contain root, got:\n%s", output)
	}

	// Reverse: children point to c3-1 via parent field
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("reverse should include c3-101 (has parent c3-1), got:\n%s", output)
	}

	// ref-jwt has scope: [c3-1], so it points to c3-1
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("reverse should include ref-jwt (has scope c3-1), got:\n%s", output)
	}
}

func TestRunGraph_RefForward(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "ref-jwt",
		Depth:     1,
		Direction: "",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Ref forward should include entities that cite it
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("should contain root ref, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-101") {
		t.Errorf("forward from ref should include citer c3-101, got:\n%s", output)
	}

	// Should show cited-by
	if !strings.Contains(output, "cited-by:") {
		t.Errorf("ref should show cited-by, got:\n%s", output)
	}
}

func TestRunGraph_JSON(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-1",
		Depth:     1,
		Direction: "",
		JSON:      true,
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var nodes []graphNode
	if err := json.Unmarshal(buf.Bytes(), &nodes); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(nodes) < 3 {
		t.Fatalf("expected at least 3 nodes (container + 2 components), got %d", len(nodes))
	}

	// Check root is present
	found := false
	for _, n := range nodes {
		if n.ID == "c3-1" {
			found = true
			if n.Type != "container" {
				t.Errorf("c3-1 type should be container, got %s", n.Type)
			}
			if len(n.Children) != 2 {
				t.Errorf("c3-1 should have 2 children, got %d", len(n.Children))
			}
		}
	}
	if !found {
		t.Error("root c3-1 not found in JSON output")
	}
}

func TestRunGraph_Mermaid(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-1",
		Depth:     1,
		Direction: "",
		Format:    "mermaid",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Should start with graph TD
	if !strings.HasPrefix(output, "graph TD\n") {
		t.Errorf("mermaid should start with 'graph TD', got:\n%s", output)
	}

	// Should have subgraph for the container
	if !strings.Contains(output, "subgraph c3-1") {
		t.Errorf("should have subgraph for c3-1, got:\n%s", output)
	}

	// Components should be inside subgraph
	if !strings.Contains(output, "c3-101") {
		t.Errorf("should contain c3-101 node, got:\n%s", output)
	}
}

func TestRunGraph_MermaidRefEdges(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	// Graph from c3-101 depth 1 (default=all neighbors): reaches ref-jwt via refs field
	err := RunGraph(GraphOptions{
		Graph:    graph,
		EntityID: "c3-101",
		Depth:    1,
		Format:   "mermaid",
		C3Dir:    c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// c3-101 cites ref-jwt — should show dashed arrow
	if !strings.Contains(output, "-.->|cites|") {
		t.Errorf("should have dashed cite edge, got:\n%s", output)
	}

	// ref-jwt should appear as a stadium-shaped node
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("should contain ref-jwt node, got:\n%s", output)
	}
}

func TestRunGraph_WithCodeMap(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)

	// Write a code-map.yaml
	cmContent := "c3-101:\n  - src/auth/**\nc3-110:\n  - src/users/**\n"
	os.WriteFile(filepath.Join(c3Dir, "code-map.yaml"), []byte(cmContent), 0644)

	var buf bytes.Buffer
	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-101",
		Depth:     0,
		Direction: "",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "files: src/auth/**") {
		t.Errorf("should show files from code-map, got:\n%s", output)
	}
}

func TestRunGraph_UnknownEntity(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "nonexistent",
		Depth:     1,
		Direction: "",
		C3Dir:     c3Dir,
	}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestRunGraph_InvalidDirection(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-1",
		Depth:     1,
		Direction: "sideways",
		C3Dir:     c3Dir,
	}, &buf)
	if err == nil {
		t.Fatal("expected error for invalid direction")
	}
}

func TestRunGraph_Depth2(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	// From context c3-0 depth 2: should reach containers (depth 1) and components (depth 2)
	err := RunGraph(GraphOptions{
		Graph:     graph,
		EntityID:  "c3-0",
		Depth:     2,
		Direction: "",
		C3Dir:     c3Dir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Context
	if !strings.Contains(output, "c3-0 (context)") {
		t.Errorf("should contain context, got:\n%s", output)
	}
	// Containers (depth 1)
	if !strings.Contains(output, "c3-1 (container)") {
		t.Errorf("should contain container at depth 1, got:\n%s", output)
	}
	// Components (depth 2)
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("should contain component at depth 2, got:\n%s", output)
	}
}
