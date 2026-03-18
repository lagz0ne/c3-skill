package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunList_Topology(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Architecture summary line
	if !strings.Contains(output, "containers") {
		t.Errorf("should show container count in summary, got:\n%s", output)
	}
	if !strings.Contains(output, "components") {
		t.Errorf("should show component count in summary, got:\n%s", output)
	}

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

	// Should show ref usage via "uses:"
	if !strings.Contains(output, "uses:") {
		t.Errorf("should show ref usage for c3-101, got:\n%s", output)
	}

	// Should show cross-cutting refs
	if !strings.Contains(output, "Cross-cutting:") {
		t.Error("should have Cross-cutting section")
	}
	if !strings.Contains(output, "ref-jwt") {
		t.Error("should list ref-jwt")
	}
}

func TestRunList_TopologyProvisioning(t *testing.T) {
	c3Dir := createFixture(t)

	// Add a provisioning component
	writeFile(t, c3Dir+"/c3-1-api/c3-120-payments.md", `---
id: c3-120
title: payments
type: component
category: feature
parent: c3-1
status: provisioning
goal: Process payments via Stripe
---

# payments
`)

	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "[provisioning]") {
		t.Errorf("should show [provisioning] badge, got:\n%s", output)
	}
}

func TestRunList_Flat(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, Flat: true, C3Dir: c3Dir}, &buf); err != nil {
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

func TestRunList_DefaultExcludesADR(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "ADR") {
		t.Errorf("default topology should not mention ADRs, got:\n%s", output)
	}
}

func TestRunList_IncludeADRShowsADR(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, C3Dir: c3Dir, IncludeADR: true}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ADR") {
		t.Errorf("--include-adr topology should mention ADRs, got:\n%s", output)
	}
}

func TestRunList_FlatExcludesADR(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, Flat: true, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "adr-") {
		t.Errorf("default flat should not include ADR lines, got:\n%s", output)
	}
}

func TestRunList_FlatIncludesADR(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, Flat: true, C3Dir: c3Dir, IncludeADR: true}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "adr-") {
		t.Errorf("--include-adr flat should include ADR lines, got:\n%s", output)
	}
}

func TestRunList_JSONExcludesADR(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, JSON: true, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, entity := range data {
		if entity["type"] == "adr" {
			t.Errorf("default JSON should not include ADR entities, found: %v", entity["id"])
		}
	}
}

func TestRunList_JSONIncludesADR(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, JSON: true, C3Dir: c3Dir, IncludeADR: true}, &buf); err != nil {
		t.Fatal(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	hasADR := false
	for _, entity := range data {
		if entity["type"] == "adr" {
			hasADR = true
			break
		}
	}
	if !hasADR {
		t.Error("--include-adr JSON should include ADR entities")
	}
}

func TestRunList_JSON(t *testing.T) {
	c3Dir := createFixture(t)
	graph := loadGraph(t, c3Dir)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Graph: graph, JSON: true, C3Dir: c3Dir}, &buf); err != nil {
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
