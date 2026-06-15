package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// Item 2 — Status is EDIT-PROOF: a body write/import/repair never touches status.
//
// The SETTLED reframe: Pin 3 ("terminal = immutable") is delivered by making status a
// NON-body-writable field, not by guarding a write at 5 paths. These tests assert
// "body path leaves status unchanged" rather than "write rejected." Status is moved
// ONLY by the status command, supersede, the auto-done latch, and migration.

// setStatusDirect seeds an entity's status through the privileged store writer,
// bypassing the legal-jump table — these tests pin where status lives, not the
// transition rules (Item 1 already covers those).
func setStatusDirect(t *testing.T, s *store.Store, id, status string) {
	t.Helper()
	if err := s.SetEntityStatus(id, status); err != nil {
		t.Fatalf("seed status %s=%s: %v", id, status, err)
	}
}

// TestEditProof_SetBodyFieldDoesNotMoveStatus — a RunSet of a BODY field on a change
// doc does not co-write status; only the status field moves status.
func TestEditProof_SetBodyFieldDoesNotMoveStatus(t *testing.T) {
	s := createRichDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "accepted")
	var buf bytes.Buffer

	// Set a body/metadata field (title), not status.
	if err := RunSet(SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "title", Value: "Use Go (revised)"}, &buf); err != nil {
		t.Fatalf("set title: %v", err)
	}

	e, _ := s.GetEntity("adr-20260226-use-go")
	if e.Status != "accepted" {
		t.Errorf("body-field set moved status: got %q, want %q", e.Status, "accepted")
	}
	if e.Title != "Use Go (revised)" {
		t.Errorf("body-field set should still update title, got %q", e.Title)
	}
}

// TestEditProof_FullWriteDoesNotMoveStatus — a full RunWrite body edit (FM
// including/omitting/altering status) leaves the stored status unchanged.
func TestEditProof_FullWriteDoesNotMoveStatus(t *testing.T) {
	cases := []struct {
		name   string
		fmLine string // a status frontmatter line to inject (or "" to omit)
	}{
		{"omits status", ""},
		{"includes same status", "status: accepted\n"},
		{"attempts to alter status", "status: done\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := createRichDBFixture(t)
			setStatusDirect(t, s, "adr-20260226-use-go", "accepted")
			var buf bytes.Buffer

			content := "---\nid: adr-20260226-use-go\ntitle: Use Go for CLI\n" + tc.fmLine + "---\n\n" +
				fullADRBody("Adopt Go for the CLI implementation.")
			if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Content: content}, &buf); err != nil {
				t.Fatalf("full write: %v", err)
			}

			e, _ := s.GetEntity("adr-20260226-use-go")
			if e.Status != "accepted" {
				t.Errorf("full write (%s) moved status: got %q, want unchanged %q", tc.name, e.Status, "accepted")
			}
		})
	}
}

// TestEditProof_WriteSectionDoesNotMoveStatus — a section write leaves status
// unchanged (write-section never touches frontmatter status).
func TestEditProof_WriteSectionDoesNotMoveStatus(t *testing.T) {
	s := createRichDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "accepted")
	// Materialize a full body so a section exists to replace.
	var seed bytes.Buffer
	full := "---\nid: adr-20260226-use-go\ntitle: Use Go for CLI\n---\n\n" + fullADRBody("Adopt Go for the CLI implementation.")
	if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Content: full}, &seed); err != nil {
		t.Fatalf("seed body: %v", err)
	}

	var buf bytes.Buffer
	if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Section: "Goal", Content: "Adopt Go for the CLI everywhere."}, &buf); err != nil {
		t.Fatalf("write section: %v", err)
	}

	e, _ := s.GetEntity("adr-20260226-use-go")
	if e.Status != "accepted" {
		t.Errorf("section write moved status: got %q, want %q", e.Status, "accepted")
	}
}

// exportThenStore exports a store to a .c3/ tree and returns the c3Dir, so an
// import/repair rebuild can be exercised against on-disk canonical markdown.
func exportThenStore(t *testing.T, s *store.Store) string {
	t.Helper()
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	if err := RunExport(ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard); err != nil {
		t.Fatalf("export: %v", err)
	}
	return c3Dir
}

// TestEditProof_ImportRebuildPreservesStatus — an import rebuild from canonical
// markdown preserves a change doc's stored status (does not default it to "active").
func TestEditProof_ImportRebuildPreservesStatus(t *testing.T) {
	s := createDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "accepted")
	c3Dir := exportThenStore(t, s)

	var buf bytes.Buffer
	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true, SkipBackup: true}, &buf); err != nil {
		t.Fatalf("import: %v", err)
	}

	rs, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open rebuilt db: %v", err)
	}
	defer rs.Close()
	e, err := rs.GetEntity("adr-20260226-use-go")
	if err != nil {
		t.Fatalf("get rebuilt entity: %v", err)
	}
	if e.Status != "accepted" {
		t.Errorf("import rebuild lost change-doc status: got %q, want %q", e.Status, "accepted")
	}
}

// TestEditProof_RepairRebuildPreservesStatus — a repair (rebuild + reseal) preserves
// a change doc's stored status. RunRepair also runs the verification suite; the
// fixture is not check-clean, so a suite error is expected and irrelevant here — the
// property under test is that the rebuild does not move status.
func TestEditProof_RepairRebuildPreservesStatus(t *testing.T) {
	s := createDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "accepted")
	c3Dir := exportThenStore(t, s)

	// Ignore the verification-suite error: the rebuild+reseal is what must preserve
	// status, not whether the fixture passes every check.
	_ = RunRepair(RepairOptions{C3Dir: c3Dir, IncludeADR: true}, io.Discard)

	rs, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open rebuilt db: %v", err)
	}
	defer rs.Close()
	e, err := rs.GetEntity("adr-20260226-use-go")
	if err != nil {
		t.Fatalf("get rebuilt entity: %v", err)
	}
	if e.Status != "accepted" {
		t.Errorf("repair rebuild lost change-doc status: got %q, want %q", e.Status, "accepted")
	}
}

// TestEditProof_TerminalDocImmutableViaBody — a done/superseded change doc cannot be
// un-frozen by any body write; status stays terminal.
func TestEditProof_TerminalDocImmutableViaBody(t *testing.T) {
	for _, terminal := range []string{"done", "superseded"} {
		t.Run(terminal, func(t *testing.T) {
			s := createRichDBFixture(t)
			setStatusDirect(t, s, "adr-20260226-use-go", terminal)
			var buf bytes.Buffer

			// A body write attempting to set a non-terminal status must not un-freeze.
			content := "---\nid: adr-20260226-use-go\ntitle: Use Go for CLI\nstatus: open\n---\n\n" +
				fullADRBody("Reopen the decision via a body edit.")
			if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Content: content}, &buf); err != nil {
				t.Fatalf("full write: %v", err)
			}

			e, _ := s.GetEntity("adr-20260226-use-go")
			if e.Status != terminal {
				t.Errorf("body write un-froze terminal doc: got %q, want %q", e.Status, terminal)
			}
		})
	}
}

// TestEditProof_OmissionDoesNotUnfreeze — a write whose FM OMITS status for a terminal
// doc does NOT demote it to ""/"active" on the next import round-trip.
func TestEditProof_OmissionDoesNotUnfreeze(t *testing.T) {
	s := createDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "done")

	// Body write omitting status — must leave the stored status terminal.
	var buf bytes.Buffer
	content := "---\nid: adr-20260226-use-go\ntitle: Use Go for CLI\n---\n\n" +
		fullADRBody("Edit the body without re-declaring status.")
	if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Content: content}, &buf); err != nil {
		t.Fatalf("full write: %v", err)
	}
	if e, _ := s.GetEntity("adr-20260226-use-go"); e.Status != "done" {
		t.Fatalf("status-omitting body write demoted terminal doc: got %q, want %q", e.Status, "done")
	}

	// And the export->import round-trip must keep it terminal (not default to active).
	c3Dir := exportThenStore(t, s)
	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true, SkipBackup: true}, io.Discard); err != nil {
		t.Fatalf("import: %v", err)
	}
	rs, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		t.Fatalf("open rebuilt db: %v", err)
	}
	defer rs.Close()
	e, _ := rs.GetEntity("adr-20260226-use-go")
	if e.Status != "done" {
		t.Errorf("import round-trip un-froze terminal doc: got %q, want %q", e.Status, "done")
	}
}

// TestEditProof_NonTerminalAndFactsStillBodyWritable — an open change doc and any fact
// remain body-writable (their content edits succeed); only status is untouched.
func TestEditProof_NonTerminalAndFactsStillBodyWritable(t *testing.T) {
	s := createRichDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "open")
	var buf bytes.Buffer

	// Open change doc: body content still writable.
	adrContent := "---\nid: adr-20260226-use-go\ntitle: Use Go for CLI\n---\n\n" +
		fullADRBody("Open change doc body remains editable.")
	if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Content: adrContent}, &buf); err != nil {
		t.Fatalf("open change-doc write: %v", err)
	}
	if e, _ := s.GetEntity("adr-20260226-use-go"); e.Status != "open" {
		t.Errorf("open change-doc status changed: got %q, want %q", e.Status, "open")
	}

	// Fact (component): body content still writable.
	factBody := strictComponentBody("auth", "Body of a fact stays freely editable.")
	if err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: factBody}, &buf); err != nil {
		t.Fatalf("fact write: %v", err)
	}
	if e, _ := s.GetEntity("c3-101"); e.Goal != "Body of a fact stays freely editable." {
		t.Errorf("fact body write did not take effect: Goal = %q", e.Goal)
	}
}

// TestEditProof_DoesNotJudgeWhetherShouldBeFrozen (NEGATIVE / the LINE) — the
// edit-proof rule enforces "status is not reachable from a body write"; it does NOT
// judge whether the doc SHOULD have been frozen. A body write on a terminal doc
// succeeds (content updates) and emits no readiness/should-freeze verdict — only the
// status stays put.
func TestEditProof_DoesNotJudgeWhetherShouldBeFrozen(t *testing.T) {
	s := createRichDBFixture(t)
	setStatusDirect(t, s, "adr-20260226-use-go", "done")
	var buf bytes.Buffer

	content := "---\nid: adr-20260226-use-go\ntitle: Use Go for CLI\n---\n\n" +
		fullADRBody("Body content of a terminal doc is still editable.")
	if err := RunWrite(WriteOptions{Store: s, ID: "adr-20260226-use-go", Content: content}, &buf); err != nil {
		t.Fatalf("body write on terminal doc must succeed (no freeze judgment): %v", err)
	}

	out := buf.String()
	for _, verdict := range []string{"should be frozen", "should not edit", "frozen", "immutable", "terminal"} {
		if bytes.Contains([]byte(out), []byte(verdict)) {
			t.Errorf("edit-proof rule must not judge whether doc should be frozen, got verdict %q in: %s", verdict, out)
		}
	}
}
