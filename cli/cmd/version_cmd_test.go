package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	s := createDBFixture(t)
	s.CreateVersion("c3-101", "# auth\n\nHandle authentication.\n", "merkle1")

	var buf bytes.Buffer
	err := RunVersion(VersionOptions{Store: s, EntityID: "c3-101", Version: 1}, &buf)
	if err != nil {
		t.Fatalf("RunVersion: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Handle authentication") {
		t.Errorf("expected version content, got:\n%s", out)
	}
}

func TestRunVersion_NotFound(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunVersion(VersionOptions{Store: s, EntityID: "c3-101", Version: 99}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent version")
	}
}

func TestRunVersion_MissingEntity(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunVersion(VersionOptions{Store: s, EntityID: "nonexistent", Version: 1}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}
