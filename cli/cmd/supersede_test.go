package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// relExists reports whether a from->to relationship of relType is in the store.
func relExists(t *testing.T, s *store.Store, fromID, toID, relType string) bool {
	t.Helper()
	rels, err := s.RelationshipsFrom(fromID)
	if err != nil {
		t.Fatalf("RelationshipsFrom(%s): %v", fromID, err)
	}
	for _, r := range rels {
		if r.ToID == toID && r.RelType == relType {
			return true
		}
	}
	return false
}

// seedSupersedeFixture seeds two change docs: an existing terminal `done` doc
// (the one to be superseded) and a fresh `open` successor doc.
func seedSupersedeFixture(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	entities := []*store.Entity{
		{ID: "c3-0", Type: "system", Title: "TestProject", Status: "active", Metadata: "{}"},
		{ID: "adr-old", Type: "adr", Title: "Old decision", Slug: "old", Status: "done", Date: "20260101", Metadata: "{}"},
		{ID: "adr-new", Type: "adr", Title: "New decision", Slug: "new", Status: "open", Date: "20260601", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed entity %s: %v", e.ID, err)
		}
	}
	return s
}

// T3.2 — supersede command flips old->superseded + writes the backlink.
func TestSupersede_FlipsOldToSupersededAndBacklinks(t *testing.T) {
	s := seedSupersedeFixture(t)

	var buf bytes.Buffer
	if err := RunSupersede(SupersedeOptions{Store: s, NewID: "adr-new", OldID: "adr-old"}, &buf); err != nil {
		t.Fatalf("RunSupersede: %v", err)
	}

	old, err := s.GetEntity("adr-old")
	if err != nil {
		t.Fatalf("get adr-old: %v", err)
	}
	if old.Status != "superseded" {
		t.Fatalf("expected adr-old to be flipped to superseded, got %q", old.Status)
	}
	if !relExists(t, s, "adr-new", "adr-old", "supersedes") {
		t.Fatalf("expected backlink adr-new --supersedes--> adr-old")
	}
}

// T3.2 — superseding a non-terminal (open/accepted) target is rejected.
func TestSupersede_RejectsNonTerminalTarget(t *testing.T) {
	s := seedSupersedeFixture(t)
	if err := s.SetEntityStatus("adr-old", "accepted"); err != nil {
		t.Fatalf("set adr-old accepted: %v", err)
	}

	var buf bytes.Buffer
	err := RunSupersede(SupersedeOptions{Store: s, NewID: "adr-new", OldID: "adr-old"}, &buf)
	if err == nil {
		t.Fatalf("expected error superseding a non-terminal target, got nil")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Fatalf("expected error to mention terminal, got %q", err.Error())
	}

	old, _ := s.GetEntity("adr-old")
	if old.Status != "accepted" {
		t.Fatalf("expected adr-old status unchanged after rejection, got %q", old.Status)
	}
	if relExists(t, s, "adr-new", "adr-old", "supersedes") {
		t.Fatalf("expected no backlink after a rejected supersede")
	}
}

// T3.2 — a supersede that would form a cycle (A supersedes B; B supersedes A) is rejected.
func TestSupersede_RejectsCycle(t *testing.T) {
	s := seedSupersedeFixture(t)

	// adr-new supersedes adr-old (legal: old is terminal).
	var buf bytes.Buffer
	if err := RunSupersede(SupersedeOptions{Store: s, NewID: "adr-new", OldID: "adr-old"}, &buf); err != nil {
		t.Fatalf("first RunSupersede: %v", err)
	}
	// adr-new is now to-be-superseded by adr-old; make adr-new terminal so the
	// only rejection cause is the cycle, not a non-terminal target.
	if err := s.SetEntityStatus("adr-new", "superseded"); err != nil {
		t.Fatalf("set adr-new superseded: %v", err)
	}

	// adr-old supersedes adr-new would close the cycle adr-new->adr-old->adr-new.
	err := RunSupersede(SupersedeOptions{Store: s, NewID: "adr-old", OldID: "adr-new"}, &buf)
	if err == nil {
		t.Fatalf("expected error for a supersede cycle, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected error to mention cycle, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "hint:") || !strings.Contains(err.Error(), "c3x graph adr-new --direction forward") {
		t.Fatalf("cycle error should include actionable hint, got %q", err.Error())
	}
}

func TestSupersede_UsageAndMissingIDsIncludeHints(t *testing.T) {
	s := seedSupersedeFixture(t)
	var buf bytes.Buffer

	err := RunSupersede(SupersedeOptions{Store: s}, &buf)
	if err == nil {
		t.Fatal("expected usage error")
	}
	requireAll(t, err.Error(), "usage", "hint:", "c3x list --include-adr")

	err = RunSupersede(SupersedeOptions{Store: s, NewID: "adr-missing", OldID: "adr-old"}, &buf)
	if err == nil {
		t.Fatal("expected missing successor error")
	}
	requireAll(t, err.Error(), "successor", "not found", "hint:", "c3x search")

	err = RunSupersede(SupersedeOptions{Store: s, NewID: "adr-new", OldID: "adr-missing"}, &buf)
	if err == nil {
		t.Fatal("expected missing target error")
	}
	requireAll(t, err.Error(), "target", "not found", "hint:", "c3x search")
}

// T3.2 — the newly-superseded doc is now terminal/immutable per Item 2's edit-proof rule.
func TestSupersede_SupersededDocIsImmutable(t *testing.T) {
	s := seedSupersedeFixture(t)

	var buf bytes.Buffer
	if err := RunSupersede(SupersedeOptions{Store: s, NewID: "adr-new", OldID: "adr-old"}, &buf); err != nil {
		t.Fatalf("RunSupersede: %v", err)
	}

	// Item 1: superseded is terminal — no legal next state via the status command.
	err := RunSet(SetOptions{Store: s, ID: "adr-old", Field: "status", Value: "open"}, &buf)
	if err == nil {
		t.Fatalf("expected a superseded doc to reject a status change via the manual command")
	}

	old, _ := s.GetEntity("adr-old")
	if old.Status != "superseded" {
		t.Fatalf("expected superseded doc to stay superseded, got %q", old.Status)
	}
}

// T3.1 — the supersede rel round-trips through a body write (write-side wiring).
//
// The wiring under test (syncRelationships) is entity-type-agnostic and the
// store accepts any rel_type; a `ref` body keeps the test focused on the rel
// round-trip rather than change-doc section validation.
func TestSupersede_BacklinkRoundTripsThroughWrite(t *testing.T) {
	s := seedSupersedeFixture(t)
	if err := s.InsertEntity(&store.Entity{ID: "ref-new", Type: "ref", Title: "New ref", Slug: "new-ref", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatalf("seed ref-new: %v", err)
	}

	content := `---
id: ref-new
type: ref
title: New ref
supersedes: [adr-old]
---

# New ref

## Goal

Standardize the replacement pattern.

## Choice

Adopt the new approach from adr-new.

## Why

The old decision was retired; this pattern carries it forward.
`
	var buf bytes.Buffer
	if err := RunWrite(WriteOptions{Store: s, ID: "ref-new", Content: content}, &buf); err != nil {
		t.Fatalf("RunWrite: %v", err)
	}

	if !relExists(t, s, "ref-new", "adr-old", "supersedes") {
		t.Fatalf("expected supersedes rel present in store after RunWrite")
	}
}

// T3.1 — the both-places guard: the supersede rel must SURVIVE an import rebuild.
// A naive single-place (write-only) wiring leaves this RED.
func TestSupersede_BacklinkRoundTripsThroughImport(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	for _, sub := range []string{c3Dir, filepath.Join(c3Dir, "adr")} {
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}

	writeFile(t, filepath.Join(c3Dir, "README.md"), `---
id: c3-0
title: TestProject
---

# TestProject

## Goal

System goal.
`)
	writeFile(t, filepath.Join(c3Dir, "adr", "adr-old.md"), `---
id: adr-old
type: adr
title: Old decision
date: 2026-01-01T00:00:00Z
status: done
---

# Old decision

## Context

The old decision.
`)
	writeFile(t, filepath.Join(c3Dir, "adr", "adr-new.md"), `---
id: adr-new
type: adr
title: New decision
date: 2026-06-01T00:00:00Z
supersedes: [adr-old]
---

# New decision

## Context

Replaces the old decision.
`)

	var buf bytes.Buffer
	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true, SkipBackup: true}, &buf); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open rebuilt db: %v", err)
	}
	defer s.Close()

	if !relExists(t, s, "adr-new", "adr-old", "supersedes") {
		t.Fatalf("supersede rel missing after import rebuild")
	}
}

// NEGATIVE (the LINE) — supersede performs the mechanical flip+link+guards only;
// it does NOT judge whether the successor is a legitimate replacement.
func TestSupersede_DoesNotJudgeSuccessorLegitimacy(t *testing.T) {
	s := seedSupersedeFixture(t)
	// An arbitrary, semantically-unrelated successor still supersedes a terminal
	// target: the tool never second-guesses the choice.
	if err := s.UpdateEntity(&store.Entity{ID: "adr-new", Type: "adr", Title: "Totally unrelated thing", Slug: "new", Status: "open", Date: "20260601", Metadata: "{}"}); err != nil {
		t.Fatalf("update adr-new: %v", err)
	}

	var buf bytes.Buffer
	if err := RunSupersede(SupersedeOptions{Store: s, NewID: "adr-new", OldID: "adr-old"}, &buf); err != nil {
		t.Fatalf("expected mechanical supersede to succeed regardless of successor semantics, got %v", err)
	}
	if !relExists(t, s, "adr-new", "adr-old", "supersedes") {
		t.Fatalf("expected backlink for an arbitrary-but-mechanically-valid supersede")
	}
}
