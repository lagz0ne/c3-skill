package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestGuardCanonicalMutation_AllowsFrozenFactCodemapSet(t *testing.T) {
	s := createDBFixture(t)
	c3Dir := t.TempDir()

	err := GuardCanonicalMutation(s, c3Dir, Options{
		Command: "set",
		Args:    []string{"c3-101", "codemap", "src/auth/**"},
	})
	if err != nil {
		t.Fatalf("set codemap on a frozen fact should be allowed, got: %v", err)
	}
}

func TestGuardCanonicalMutation_AllowsFrozenFactCodemapSetWithFieldFlag(t *testing.T) {
	s := createDBFixture(t)
	c3Dir := t.TempDir()

	err := GuardCanonicalMutation(s, c3Dir, Options{
		Command: "set",
		Args:    []string{"c3-101", "src/auth/**"},
		Field:   "codemap",
	})
	if err != nil {
		t.Fatalf("set --field codemap on a frozen fact should be allowed, got: %v", err)
	}
}

func TestGuardCanonicalMutation_RefusesFrozenFactNonCodemapSet(t *testing.T) {
	s := createDBFixture(t)
	c3Dir := t.TempDir()

	// A real frozen fact has authored content (born sealed via `add`). Seed it so
	// the refusal tests the actual freeze, not the creation window.
	if err := content.WriteEntity(s, "c3-101", "## Goal\n\nAuthenticate requests.\n"); err != nil {
		t.Fatalf("seed c3-101 body: %v", err)
	}

	err := GuardCanonicalMutation(s, c3Dir, Options{
		Command: "set",
		Args:    []string{"c3-101", "goal", "x"},
	})
	if err == nil {
		t.Fatal("set goal on a frozen fact must be refused")
	}
	if !strings.Contains(err.Error(), "facts are frozen") {
		t.Fatalf("expected frozen-fact error, got: %v", err)
	}
}

// The creation window: a frozen-TYPE fact that was never authored (no body nodes
// AND Version 0 — e.g. the init-seeded system c3-0) accepts its FIRST body `write`;
// once it carries a body the freeze engages. The window is write-only — set/
// delete stay frozen even while the body is empty, since they touch frontmatter /
// existence, which can be authored independently of the body.
func TestGuardCanonicalMutation_CreationWindow(t *testing.T) {
	s := createDBFixture(t)
	c3Dir := t.TempDir()

	// A bodyless, never-versioned frozen-type fact models a fresh init's c3-0.
	if err := s.InsertEntity(&store.Entity{ID: "c3-9", Type: "system", Title: "Fresh", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatalf("insert bodyless system fact: %v", err)
	}

	// `write` opens the window — the first body authoring is allowed.
	if err := GuardCanonicalMutation(s, c3Dir, Options{Command: "write", Args: []string{"c3-9"}}); err != nil {
		t.Fatalf("first write to a bodyless fact must be allowed (creation window), got: %v", err)
	}

	// set / delete stay frozen even in the window (write-only carve-out).
	frozenCases := map[string][]string{
		"set":    {"c3-9", "goal", "x"},
		"delete": {"c3-9"},
	}
	for command, args := range frozenCases {
		if err := GuardCanonicalMutation(s, c3Dir, Options{Command: command, Args: args}); err == nil {
			t.Fatalf("%s on a bodyless fact must stay frozen (window is write-only)", command)
		}
	}

	// Author a body → the window shuts; a subsequent direct write is refused.
	if err := content.WriteEntity(s, "c3-9", "## Goal\n\nRun the thing.\n"); err != nil {
		t.Fatalf("seed body: %v", err)
	}
	if err := GuardCanonicalMutation(s, c3Dir, Options{Command: "write", Args: []string{"c3-9"}}); err == nil {
		t.Fatal("once a fact carries a body, direct write must be frozen")
	}
}
