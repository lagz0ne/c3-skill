package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunVersions(t *testing.T) {
	s := createDBFixture(t)
	// Create some versions
	s.CreateVersion("c3-101", "content v1", "merkle1")
	s.CreateVersion("c3-101", "content v2", "merkle2")

	var buf bytes.Buffer
	err := RunVersions(VersionsOptions{Store: s, EntityID: "c3-101", JSON: false}, &buf)
	if err != nil {
		t.Fatalf("RunVersions: %v", err)
	}
	out := buf.String()
	// Should contain header
	if !strings.Contains(out, "VERSION") || !strings.Contains(out, "MERKLE") {
		t.Errorf("expected table header, got:\n%s", out)
	}
	// Should show both versions
	if !strings.Contains(out, "merkle1") || !strings.Contains(out, "merkle2") {
		t.Errorf("expected both merkle hashes, got:\n%s", out)
	}
}

func TestRunVersions_JSON(t *testing.T) {
	s := createDBFixture(t)
	s.CreateVersion("c3-101", "content v1", "merkle1")

	var buf bytes.Buffer
	err := RunVersions(VersionsOptions{Store: s, EntityID: "c3-101", JSON: true}, &buf)
	if err != nil {
		t.Fatalf("RunVersions JSON: %v", err)
	}
	var versions []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &versions); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(versions))
	}
}

func TestRunVersions_Empty(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunVersions(VersionsOptions{Store: s, EntityID: "c3-101", JSON: false}, &buf)
	if err != nil {
		t.Fatalf("RunVersions empty: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "No versions") {
		t.Errorf("expected 'No versions' message, got:\n%s", out)
	}
}

func TestRunVersions_MissingEntity(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunVersions(VersionsOptions{Store: s, EntityID: "nonexistent", JSON: false}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}
