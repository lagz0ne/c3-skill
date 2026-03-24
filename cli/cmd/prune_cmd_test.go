package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrune(t *testing.T) {
	s := createDBFixture(t)
	// Create 5 versions
	for i := 0; i < 5; i++ {
		s.CreateVersion("c3-101", "content", "merkle")
	}

	var buf bytes.Buffer
	err := RunPrune(PruneOptions{Store: s, EntityID: "c3-101", Keep: 2}, &buf)
	if err != nil {
		t.Fatalf("RunPrune: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Pruned 3") {
		t.Errorf("expected 'Pruned 3', got:\n%s", out)
	}

	// Verify only 2 remain
	versions, _ := s.ListVersions("c3-101")
	if len(versions) != 2 {
		t.Errorf("expected 2 versions remaining, got %d", len(versions))
	}
}

func TestRunPrune_KeepsMarked(t *testing.T) {
	s := createDBFixture(t)
	// Create 5 versions, mark version 1 with a commit hash
	for i := 0; i < 5; i++ {
		s.CreateVersion("c3-101", "content", "merkle")
	}
	s.MarkVersion("c3-101", 1, "abc123")

	var buf bytes.Buffer
	err := RunPrune(PruneOptions{Store: s, EntityID: "c3-101", Keep: 2}, &buf)
	if err != nil {
		t.Fatalf("RunPrune: %v", err)
	}

	// Version 1 should survive (marked), plus the last 2
	versions, _ := s.ListVersions("c3-101")
	hasMarked := false
	for _, v := range versions {
		if v.Version == 1 {
			hasMarked = true
		}
	}
	if !hasMarked {
		t.Error("expected marked version 1 to survive pruning")
	}
}

func TestRunPrune_MissingEntity(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunPrune(PruneOptions{Store: s, EntityID: "nonexistent", Keep: 2}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}
