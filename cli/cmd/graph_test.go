package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunGraph_TextDepth1(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-1", Depth: 1}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-1 (container)") {
		t.Errorf("should contain root entity, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("should contain child c3-101, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-110 (component)") {
		t.Errorf("should contain child c3-110, got:\n%s", output)
	}
}

func TestRunGraph_TextDepth0(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-101", Depth: 0}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("should contain entity, got:\n%s", output)
	}
	if strings.Contains(output, "c3-1 (container)") {
		t.Errorf("depth 0 should not include parent, got:\n%s", output)
	}
}

func TestRunGraph_TextReverse(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-1", Depth: 1, Direction: "reverse"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-1 (container)") {
		t.Errorf("should contain root, got:\n%s", output)
	}
	// Children point to c3-1 via parent
	if !strings.Contains(output, "c3-101") {
		t.Errorf("reverse should include c3-101, got:\n%s", output)
	}
	// ref-jwt has scope -> c3-1
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("reverse should include ref-jwt, got:\n%s", output)
	}
}

func TestRunGraph_RefForward(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "ref-jwt", Depth: 1}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("should contain root ref, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-101") {
		t.Errorf("forward from ref should include citer c3-101, got:\n%s", output)
	}
	if !strings.Contains(output, "cited-by:") {
		t.Errorf("ref should show cited-by, got:\n%s", output)
	}
}

func TestRunGraph_JSON(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-1", Depth: 1, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var nodes []graphNode
	if err := json.Unmarshal(buf.Bytes(), &nodes); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(nodes) < 3 {
		t.Fatalf("expected at least 3 nodes, got %d", len(nodes))
	}

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
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-1", Depth: 1, Format: "mermaid"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "graph TD\n") {
		t.Errorf("mermaid should start with 'graph TD', got:\n%s", output)
	}
	if !strings.Contains(output, "subgraph c3-1") {
		t.Errorf("should have subgraph for c3-1, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-101") {
		t.Errorf("should contain c3-101 node, got:\n%s", output)
	}
}

func TestRunGraph_MermaidRefEdges(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-101", Depth: 1, Format: "mermaid"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "-.->|cites|") {
		t.Errorf("should have dashed cite edge, got:\n%s", output)
	}
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("should contain ref-jwt node, got:\n%s", output)
	}
}

func TestRunGraph_WithCodeMap(t *testing.T) {
	s := createDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-101", Depth: 0}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "files: src/auth/**") {
		t.Errorf("should show files from code-map, got:\n%s", output)
	}
}

func TestRunGraph_UnknownEntity(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "nonexistent", Depth: 1}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestRunGraph_InvalidDirection(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-1", Depth: 1, Direction: "sideways"}, &buf)
	if err == nil {
		t.Fatal("expected error for invalid direction")
	}
}

func TestRunGraph_Depth2(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunGraph(GraphOptions{Store: s, EntityID: "c3-0", Depth: 2}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-0 (system)") {
		t.Errorf("should contain context, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-1 (container)") {
		t.Errorf("should contain container at depth 1, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-101 (component)") {
		t.Errorf("should contain component at depth 2, got:\n%s", output)
	}
}

func TestGraphMermaidRuleShape(t *testing.T) {
	s := createDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging", Slug: "logging",
		Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	RunGraph(GraphOptions{Store: s, EntityID: "rule-logging", Depth: 1, Format: "mermaid"}, &buf)
	output := buf.String()
	if !strings.Contains(output, "{{") {
		t.Error("mermaid should render rules with hexagon shape {{}}")
	}
}
