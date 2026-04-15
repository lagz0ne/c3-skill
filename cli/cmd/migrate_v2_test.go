package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestWriteMigrateWriteFailureGivesFixLoop(t *testing.T) {
	var buf bytes.Buffer
	writeMigrateWriteFailure(&buf, "c3-101", 3, errors.New("disk full"))

	requireAll(t, buf.String(),
		"BLOCKED: migration write failed at c3-101 after 3 successful write(s)",
		"C3 stopped before canonical export",
		"c3x cache clear",
		"c3x import --force",
		"c3x migrate --continue",
		"c3x check --include-adr && c3x verify",
	)
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

func TestRunMigrateV2_AggregatesStrictComponentBlockers(t *testing.T) {
	s := newMigrateV2Store(t)
	for _, e := range []*store.Entity{
		{ID: "c3-101", Type: "component", Title: "chat", Slug: "chat", Category: "feature", ParentID: "c3-1", Goal: "Render optional chat workspace behavior.", Status: "active", Metadata: "{}"},
		{ID: "c3-102", Type: "component", Title: "tasks", Slug: "tasks", Category: "feature", ParentID: "c3-1", Goal: "Track todo coordination behavior.", Status: "active", Metadata: "{}"},
	} {
		if err := s.InsertEntity(e); err != nil {
			t.Fatal(err)
		}
	}
	if err := content.WriteEntity(s, "c3-101", "# chat\n\n## Goal\n\nRender optional chat workspace behavior.\n"); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-102", "# tasks\n\n## Goal\n\nTrack todo coordination behavior.\n"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunMigrateV2(MigrateV2Options{Store: s}, &buf)
	if err == nil {
		t.Fatal("expected migration blocker error")
	}
	out := buf.String()
	requireAll(t, out,
		"0 migrated, 2 blocked",
		"BLOCKED: 2 component(s)",
		"C3 made no migration writes",
		"writesMade: false",
		"c3-101 chat",
		"matched: optional",
		"c3-102 tasks",
		"matched: todo",
		"common rewrite: optional->secondary",
		"c3x cache clear",
		"c3x import --force",
		"c3x migrate --continue",
		"c3x check --include-adr && c3x verify",
	)
	if !strings.Contains(err.Error(), "migrate blocked: 2 component(s)") {
		t.Fatalf("unexpected error: %v", err)
	}
	body, readErr := content.ReadEntity(s, "c3-101")
	if readErr != nil {
		t.Fatal(readErr)
	}
	if strings.Contains(body, "## Parent Fit") {
		t.Fatalf("blocked migrate must not write partial strict docs, got:\n%s", body)
	}
}

func TestRunMigrateV2_JSONBlockerReport(t *testing.T) {
	s := newMigrateV2Store(t)
	for _, e := range []*store.Entity{
		{ID: "c3-101", Type: "component", Title: "chat", Slug: "chat", Category: "feature", ParentID: "c3-1", Goal: "Render optional chat workspace behavior.", Status: "active", Metadata: "{}"},
		{ID: "c3-102", Type: "component", Title: "tasks", Slug: "tasks", Category: "feature", ParentID: "c3-1", Goal: "Track TODO coordination behavior.", Status: "active", Metadata: "{}"},
	} {
		if err := s.InsertEntity(e); err != nil {
			t.Fatal(err)
		}
		if err := content.WriteEntity(s, e.ID, "# "+e.Title+"\n\n## Goal\n\n"+e.Goal+"\n"); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	err := RunMigrateV2(MigrateV2Options{Store: s, DryRun: true, JSON: true}, &buf)
	if err == nil {
		t.Fatal("expected migration blocker error")
	}
	var report MigrateReport
	if unmarshalErr := json.Unmarshal(buf.Bytes(), &report); unmarshalErr != nil {
		t.Fatalf("unmarshal report: %v\n%s", unmarshalErr, buf.String())
	}
	if report.Status != "blocked" || report.WritesMade || report.Blocked != 2 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if len(report.Blockers) != 2 {
		t.Fatalf("expected two blockers: %#v", report.Blockers)
	}
	if report.Blockers[0].Issues[0].Matched != "optional" {
		t.Fatalf("expected matched optional, got %#v", report.Blockers[0].Issues[0])
	}
	if report.Blockers[1].Issues[0].Matched != "TODO" {
		t.Fatalf("expected matched TODO, got %#v", report.Blockers[1].Issues[0])
	}
	requireAll(t, strings.Join(report.Next, "\n"),
		"c3x migrate repair-plan",
		"c3x cache clear",
		"c3x migrate --continue",
	)
}

func TestRunMigrateRepairPlanGivesSafeLoop(t *testing.T) {
	s := newMigrateV2Store(t)
	e := &store.Entity{ID: "c3-101", Type: "component", Title: "chat", Slug: "chat", Category: "feature", ParentID: "c3-1", Goal: "Render optional chat workspace behavior.", Status: "active", Metadata: "{}"}
	if err := s.InsertEntity(e); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, e.ID, "# chat\n\n## Goal\n\nRender optional chat workspace behavior.\n"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunMigrateRepairPlan(s, &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(),
		"BLOCKED: 1 component(s)",
		"writesMade: false",
		"c3x migrate repair <id> --section <section>",
		"matched: optional",
		"c3x cache clear",
		"c3x import --force",
		"c3x migrate --continue",
	)
}

func TestRunMigrateRepairSectionOnlyRepairsCurrentBlockerSection(t *testing.T) {
	s := newMigrateV2Store(t)
	e := &store.Entity{ID: "c3-101", Type: "component", Title: "chat", Slug: "chat", Category: "feature", ParentID: "c3-1", Goal: "Render optional chat workspace behavior.", Status: "active", Metadata: "{}"}
	if err := s.InsertEntity(e); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, e.ID, "# chat\n\n## Goal\n\nRender optional chat workspace behavior.\n"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunMigrateRepairSection(s, e.ID, "Purpose", "Own chat behavior under the API container with explicit boundaries and verification.", &buf)
	if err == nil {
		t.Fatal("expected non-blocker section repair to fail")
	}
	requireAll(t, err.Error(), "not listed as a migration blocker")

	buf.Reset()
	err = RunMigrateRepairSection(s, e.ID, "Goal", "Render secondary chat workspace behavior under the API container.", &buf)
	if err != nil {
		t.Fatalf("repair Goal: %v\n%s", err, buf.String())
	}
	requireAll(t, buf.String(), "Updated c3-101 section \"Goal\"")

	blockers, err := collectMigrateBlockers(s, []*store.Entity{e})
	if err != nil {
		t.Fatal(err)
	}
	if len(blockers) != 0 {
		t.Fatalf("expected blocker fixed, got %#v", blockers)
	}
}
