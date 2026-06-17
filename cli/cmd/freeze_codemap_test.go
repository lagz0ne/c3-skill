package cmd

import (
	"strings"
	"testing"
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
