package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunImpact(t *testing.T) {
	s := createDBFixture(t)
	// ref-jwt is used by c3-101, so impact of ref-jwt should show c3-101
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: "ref-jwt", Depth: 3}, &buf)
	if err != nil {
		t.Fatalf("RunImpact: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in impact of ref-jwt, got:\n%s", out)
	}
	if !strings.Contains(out, "Impact of ref-jwt") {
		t.Errorf("expected header line, got:\n%s", out)
	}
}

func TestRunImpact_NoAffected(t *testing.T) {
	s := createDBFixture(t)
	// c3-110 (users) has no inbound "uses" relationships, so impact should be empty
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: "c3-110", Depth: 3}, &buf)
	if err != nil {
		t.Fatalf("RunImpact: %v", err)
	}
	if !strings.Contains(buf.String(), "No affected entities found.") {
		t.Errorf("expected 'No affected entities found.', got:\n%s", buf.String())
	}
}

func TestRunImpact_JSON(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: "ref-jwt", Depth: 3, JSON: true}, &buf)
	if err != nil {
		t.Fatalf("RunImpact JSON: %v", err)
	}
	var results []store.ImpactResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(results) == 0 {
		t.Error("expected at least one impact result")
	}
}

func TestRunImpact_EmptyEntityID(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: ""}, &buf)
	if err == nil {
		t.Fatal("expected error for empty entity ID")
	}
}
