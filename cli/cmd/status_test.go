package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunStatus_BasicDashboard(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer
	err := RunStatus(StatusOptions{Store: s}, &buf)
	if err != nil {
		t.Fatalf("RunStatus error: %v", err)
	}

	out := buf.String()
	// Should contain project name from c3-0 system entity
	if !strings.Contains(out, "TestProject") {
		t.Errorf("expected project name 'TestProject' in output, got:\n%s", out)
	}
	// Should mention containers
	if !strings.Contains(out, "container") {
		t.Errorf("expected 'container' in output, got:\n%s", out)
	}
	// Should mention components
	if !strings.Contains(out, "component") {
		t.Errorf("expected 'component' in output, got:\n%s", out)
	}
	// Should mention refs
	if !strings.Contains(out, "ref") {
		t.Errorf("expected 'ref' in output, got:\n%s", out)
	}
	// Should show total count (fixture has 9 entities)
	if !strings.Contains(out, "9") {
		t.Errorf("expected total '9' in output, got:\n%s", out)
	}
}

func TestRunStatus_JSON(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer
	err := RunStatus(StatusOptions{Store: s, JSONExplicit: true}, &buf)
	if err != nil {
		t.Fatalf("RunStatus error: %v", err)
	}

	var result StatusResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse error: %v\nraw: %s", err, buf.String())
	}

	if result.Project != "TestProject" {
		t.Errorf("expected project 'TestProject', got %q", result.Project)
	}
	if result.Entities["container"] != 2 {
		t.Errorf("expected 2 containers, got %d", result.Entities["container"])
	}
	if result.Entities["component"] != 3 {
		t.Errorf("expected 3 components, got %d", result.Entities["component"])
	}
	if result.Entities["ref"] != 2 {
		t.Errorf("expected 2 refs, got %d", result.Entities["ref"])
	}
	if result.TotalCount != 9 {
		t.Errorf("expected total 9, got %d", result.TotalCount)
	}
	if result.PendingADRs != 1 {
		t.Errorf("expected 1 pending ADR, got %d", result.PendingADRs)
	}
}

func TestRunStatus_TOON(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	s := createRichDBFixture(t)
	var buf bytes.Buffer
	// Agent mode, no explicit --json -> TOON format
	err := RunStatus(StatusOptions{Store: s, JSONExplicit: false}, &buf)
	if err != nil {
		t.Fatalf("RunStatus error: %v", err)
	}

	out := buf.String()
	// TOON format uses key: value pairs
	if !strings.Contains(out, "project") {
		t.Errorf("expected 'project' key in TOON output, got:\n%s", out)
	}
	if !strings.Contains(out, "TestProject") {
		t.Errorf("expected 'TestProject' in TOON output, got:\n%s", out)
	}
	// Should include help hints in agent mode
	if !strings.Contains(out, "help[") {
		t.Errorf("expected help hints in agent TOON output, got:\n%s", out)
	}
}

func TestRunStatus_PendingADRCount(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer
	err := RunStatus(StatusOptions{Store: s, JSONExplicit: true}, &buf)
	if err != nil {
		t.Fatalf("RunStatus error: %v", err)
	}

	var result StatusResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	// The fixture has 1 ADR with status "proposed" — that counts as pending
	if result.PendingADRs != 1 {
		t.Errorf("expected 1 pending ADR (proposed), got %d", result.PendingADRs)
	}

	// ADR should still appear in entity counts
	if result.Entities["adr"] != 1 {
		t.Errorf("expected 1 adr in entities, got %d", result.Entities["adr"])
	}
}

func TestRunStatus_HelpHintsAppended(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	s := createRichDBFixture(t)
	var buf bytes.Buffer
	err := RunStatus(StatusOptions{Store: s, JSONExplicit: false}, &buf)
	if err != nil {
		t.Fatalf("RunStatus error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3x list --compact") {
		t.Errorf("expected 'c3x list --compact' in help hints, got:\n%s", out)
	}
	if !strings.Contains(out, "c3x check") {
		t.Errorf("expected 'c3x check' in help hints, got:\n%s", out)
	}
	if !strings.Contains(out, "c3x read <id>") {
		t.Errorf("expected 'c3x read <id>' in help hints, got:\n%s", out)
	}
}
