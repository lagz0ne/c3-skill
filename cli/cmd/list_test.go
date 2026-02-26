package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRunList_Topology(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(graph, false, false, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Containers: slug from "c3-1-api/README.md" -> "api"
	if !strings.Contains(output, "c3-1-api (container)") {
		t.Errorf("should list c3-1-api container, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-2-web (container)") {
		t.Errorf("should list c3-2-web container, got:\n%s", output)
	}

	// Components: slug from "c3-101-auth.md" -> "auth"
	if !strings.Contains(output, "c3-101-auth (foundation)") {
		t.Errorf("should list c3-101 auth component, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-110-users (feature)") {
		t.Errorf("should list c3-110 users component, got:\n%s", output)
	}

	// Should show ref references
	if !strings.Contains(output, "ref: ref-jwt") {
		t.Error("should show ref citation for c3-101")
	}

	// Should show cross-cutting refs
	if !strings.Contains(output, "Cross-cutting:") {
		t.Error("should have Cross-cutting section")
	}
	if !strings.Contains(output, "ref-jwt") {
		t.Error("should list ref-jwt")
	}

	// Should show ADRs
	if !strings.Contains(output, "ADRs:") {
		t.Error("should have ADRs section")
	}
	if !strings.Contains(output, "adr-20260226-use-go") {
		t.Error("should list ADR")
	}
	if !strings.Contains(output, "status: proposed") {
		t.Error("should show ADR status")
	}
}

func TestRunList_Flat(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(graph, false, true, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines, got %d: %s", len(lines), output)
	}

	// Each line should be tab-separated: id\ttype\tpath
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			t.Errorf("expected 3 tab-separated fields, got %d: %q", len(parts), line)
		}
	}
}

func TestRunList_JSON(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(graph, true, false, &buf); err != nil {
		t.Fatal(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(data) < 5 {
		t.Fatalf("expected at least 5 entities, got %d", len(data))
	}

	// Check required fields
	for _, entity := range data {
		if _, ok := entity["id"]; !ok {
			t.Error("entity missing 'id' field")
		}
		if _, ok := entity["type"]; !ok {
			t.Error("entity missing 'type' field")
		}
		if _, ok := entity["path"]; !ok {
			t.Error("entity missing 'path' field")
		}
	}
}

func TestRunList_TopologyGoalTruncation(t *testing.T) {
	dir := t.TempDir()
	c3Dir := dir + "/.c3"
	os.MkdirAll(c3Dir+"/c3-1-test", 0755)

	writeFile(t, c3Dir+"/README.md", `---
id: c3-0
title: Test
---
# Test
`)

	longGoal := strings.Repeat("x", 100)
	writeFile(t, c3Dir+"/c3-1-test/README.md", `---
id: c3-1
title: test
type: container
parent: c3-0
goal: `+longGoal+`
---
# test
Content here.
`)

	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer
	RunList(graph, false, false, &buf)

	output := buf.String()
	if strings.Contains(output, longGoal) {
		t.Error("long goal should be truncated")
	}
	if !strings.Contains(output, "...") {
		t.Error("truncated goal should end with ...")
	}
}
