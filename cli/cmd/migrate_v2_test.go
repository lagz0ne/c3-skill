package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// newMigrateV2Store creates a store with entities that already have node trees.
func newMigrateV2Store(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	entities := []*store.Entity{
		{ID: "c3-0", Type: "system", Title: "Sys", Slug: "sys", Status: "active", Metadata: "{}"},
		{ID: "c3-1", Type: "container", Title: "API", Slug: "api", ParentID: "c3-0", Status: "active", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("insert %s: %v", e.ID, err)
		}
	}
	return s
}

func TestRunMigrateV2_SkipsAlreadyMigrated(t *testing.T) {
	s := newMigrateV2Store(t)

	// Pre-populate nodes for c3-0
	if err := content.WriteEntity(s, "c3-0", "# Sys\n\n## Goal\n\nTop level.\n"); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	var buf bytes.Buffer
	err := RunMigrateV2(MigrateV2Options{Store: s, DryRun: false}, &buf)
	if err != nil {
		t.Fatalf("RunMigrateV2: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "skipped") {
		t.Errorf("expected 'skipped' in output, got:\n%s", out)
	}
}

func TestRunMigrateV2_DryRun(t *testing.T) {
	s := newMigrateV2Store(t)

	var buf bytes.Buffer
	err := RunMigrateV2(MigrateV2Options{Store: s, DryRun: true}, &buf)
	if err != nil {
		t.Fatalf("RunMigrateV2 dry-run: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected 'dry-run' in output, got:\n%s", out)
	}
}

func TestRunMigrateV2_EmptyStore(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	// Entity with no content
	if err := s.InsertEntity(&store.Entity{
		ID: "c3-empty", Type: "component", Title: "Empty", Slug: "empty",
		Status: "active", Metadata: "{}",
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var buf bytes.Buffer
	err = RunMigrateV2(MigrateV2Options{Store: s, DryRun: false}, &buf)
	if err != nil {
		t.Fatalf("RunMigrateV2: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "migrated 0") {
		t.Errorf("expected 'migrated 0', got:\n%s", out)
	}
	if !strings.Contains(out, "skipped 1") {
		t.Errorf("expected 'skipped 1', got:\n%s", out)
	}
}
