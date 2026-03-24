package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

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

	if err := content.WriteEntity(s, "c3-0", "# Sys\n\n## Goal\n\nTop level.\n"); err != nil {
		t.Fatalf("WriteEntity: %v", err)
	}

	var buf bytes.Buffer
	if err := RunMigrateV2(MigrateV2Options{Store: s}, &buf); err != nil {
		t.Fatalf("RunMigrateV2: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "already have nodes") {
		t.Errorf("expected 'already have nodes', got:\n%s", out)
	}
}

func TestRunMigrateV2_DryRun(t *testing.T) {
	s := newMigrateV2Store(t)

	var buf bytes.Buffer
	if err := RunMigrateV2(MigrateV2Options{Store: s, DryRun: true}, &buf); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected 'dry-run', got:\n%s", buf.String())
	}
}

func TestRunMigrateV2_EmptyStore(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	s.InsertEntity(&store.Entity{
		ID: "c3-empty", Type: "component", Title: "Empty", Slug: "empty",
		Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	RunMigrateV2(MigrateV2Options{Store: s}, &buf)

	out := buf.String()
	if !strings.Contains(out, "0 migrated") {
		t.Errorf("expected '0 migrated', got:\n%s", out)
	}
	if !strings.Contains(out, "no content yet") {
		t.Errorf("expected guidance for empty entities, got:\n%s", out)
	}
	if !strings.Contains(out, "c3x write c3-empty") {
		t.Errorf("expected actionable command, got:\n%s", out)
	}
}
