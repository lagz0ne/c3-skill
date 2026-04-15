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

func TestRunMigrateV2_StrictMigratesLegacyComponentNodes(t *testing.T) {
	s := newMigrateV2Store(t)
	if err := s.InsertEntity(&store.Entity{
		ID: "c3-101", Type: "component", Title: "auth", Slug: "auth",
		Category: "foundation", ParentID: "c3-1", Goal: "Handle authentication behavior for API requests.",
		Status: "active", Metadata: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nHandle auth.\n\n## Dependencies\n\n| Direction | What | From/To |\n|---|---|---|\n| IN | credentials | c3-1 |\n"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunMigrateV2(MigrateV2Options{Store: s}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "strict-migrated component: c3-101") {
		t.Fatalf("expected strict migration output, got:\n%s", buf.String())
	}

	body, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	requireAll(t, body, "## Parent Fit", "## Governance", "## Derived Materials")
	if issues := validateStrictComponentDoc(body, "error"); len(issues) > 0 {
		t.Fatalf("strict migrated body failed validation: %#v\n%s", issues, body)
	}
}

func TestRunMigrateV2_EmptyStore(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	s.InsertEntity(&store.Entity{
		ID: "c3-101", Type: "component", Title: "Empty", Slug: "empty",
		Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	if err := RunMigrateV2(MigrateV2Options{Store: s}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "strict-migrated empty component: c3-101") {
		t.Errorf("expected strict empty component recovery, got:\n%s", out)
	}
	body, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if issues := validateStrictComponentDoc(body, "error"); len(issues) > 0 {
		t.Fatalf("empty component recovery should produce strict body: %#v\n%s", issues, body)
	}
}
