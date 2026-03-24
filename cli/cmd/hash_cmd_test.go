package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunHash(t *testing.T) {
	s := createDBFixture(t)
	// Set a known root_merkle on the entity
	e, _ := s.GetEntity("c3-101")
	e.RootMerkle = "deadbeef1234"
	s.UpdateEntity(e)

	var buf bytes.Buffer
	err := RunHash(HashOptions{Store: s, EntityID: "c3-101"}, &buf)
	if err != nil {
		t.Fatalf("RunHash: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "deadbeef1234" {
		t.Errorf("expected 'deadbeef1234', got %q", out)
	}
}

func TestRunHash_Recompute(t *testing.T) {
	s := createDBFixture(t)
	seedNodes(t, s)

	nodes, _ := s.NodesForEntity("c3-101")
	expected := store.HashNodes(nodes)

	e, _ := s.GetEntity("c3-101")
	e.RootMerkle = expected
	s.UpdateEntity(e)

	var buf bytes.Buffer
	err := RunHash(HashOptions{Store: s, EntityID: "c3-101", Recompute: true}, &buf)
	if err != nil {
		t.Fatalf("RunHash recompute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK (hashes match), got:\n%s", out)
	}
}

func TestRunHash_Recompute_Drift(t *testing.T) {
	s := createDBFixture(t)
	seedNodes(t, s)

	// Set entity merkle to a wrong value
	e, _ := s.GetEntity("c3-101")
	e.RootMerkle = "stale_hash"
	s.UpdateEntity(e)

	var buf bytes.Buffer
	err := RunHash(HashOptions{Store: s, EntityID: "c3-101", Recompute: true}, &buf)
	if err != nil {
		t.Fatalf("RunHash recompute drift: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRIFT") {
		t.Errorf("expected DRIFT warning, got:\n%s", out)
	}
	if !strings.Contains(out, "stale_hash") {
		t.Errorf("expected stored hash in output, got:\n%s", out)
	}
}

func TestRunHash_MissingEntity(t *testing.T) {
	s := createDBFixture(t)

	var buf bytes.Buffer
	err := RunHash(HashOptions{Store: s, EntityID: "nonexistent"}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}
