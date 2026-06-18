package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// componentWithObligation is a minimal component body carrying one non-N.A Derived
// Materials row — i.e. a derivation obligation the switch must force an inspection of.
const componentWithObligation = `# auth

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| auth handler | this component's Contract | naming only | auth.go |
`

// gateFixture builds a project where c3-101 (a component with a derivation
// obligation) is touched by a codemap carrier that binds it to a real file. It
// returns the store, the c3Dir, the unit id, and the carrier's current material hash.
func gateFixture(t *testing.T) (*store.Store, string, string, string) {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	projectDir := t.TempDir()
	c3Dir := filepath.Join(projectDir, ".c3")
	if err := os.MkdirAll(c3Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A real file for the code-map territory to resolve to.
	if err := os.WriteFile(filepath.Join(projectDir, "auth.go"), []byte("package auth\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.InsertEntity(&store.Entity{ID: "c3-101", Type: "component", Title: "auth", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-101", componentWithObligation); err != nil {
		t.Fatal(err)
	}

	unit := "adr-20260618-x"
	dir := filepath.Join(c3Dir, "changes", unit)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	carrier := "---\ntarget: c3-101\n---\nauth.go\n"
	if err := os.WriteFile(filepath.Join(dir, "01.codemap.md"), []byte(carrier), 0o644); err != nil {
		t.Fatal(err)
	}
	return s, c3Dir, unit, changeset.MaterialHash(carrier)
}

func readCarriers(t *testing.T, c3Dir, unit string) []changeset.CodemapChange {
	t.Helper()
	cm, err := changeset.ReadCodemapDir(filepath.Join(c3Dir, "changes", unit))
	if err != nil {
		t.Fatal(err)
	}
	return cm
}

// The gate must REFUSE: c3-101 has a derivation obligation + territory, but the
// unit carries no inspection.
func TestInspectionGate_RefusesWithoutInspection(t *testing.T) {
	s, c3Dir, unit, _ := gateFixture(t)
	rejects, err := inspectionGate(s, c3Dir, unit, nil, readCarriers(t, c3Dir, unit))
	if err != nil {
		t.Fatal(err)
	}
	if len(rejects) == 0 {
		t.Fatal("expected a reject: obligation present, no inspection")
	}
	if !strings.Contains(strings.Join(rejects, "\n"), "no inspection") {
		t.Errorf("reject should name the missing inspection, got: %v", rejects)
	}
}

// The gate must PASS with a fresh, grounded, territory-citing inspection.
func TestInspectionGate_PassesWithGroundedInspection(t *testing.T) {
	s, c3Dir, unit, hash := gateFixture(t)
	insp := "---\ntarget: c3-101\ncovers:\n  - source: 01.codemap.md\n    hash: " + hash + "\n---\n" +
		"## Inspections\n\n| Obligation | Territory | Verdict | Evidence |\n| --- | --- | --- | --- |\n" +
		"| auth handler | auth.go | matches | `go build ./...`; auth.go:1 |\n"
	if err := os.WriteFile(filepath.Join(c3Dir, "changes", unit, "02.inspect.md"), []byte(insp), 0o644); err != nil {
		t.Fatal(err)
	}
	rejects, err := inspectionGate(s, c3Dir, unit, nil, readCarriers(t, c3Dir, unit))
	if err != nil {
		t.Fatal(err)
	}
	if len(rejects) != 0 {
		t.Fatalf("a fresh grounded inspection should pass, got: %v", rejects)
	}
}

// The gate must REFUSE a stale inspection — one whose recorded material hash no
// longer matches the current carrier (the doc material changed since inspecting).
func TestInspectionGate_RefusesStaleInspection(t *testing.T) {
	s, c3Dir, unit, _ := gateFixture(t)
	insp := "---\ntarget: c3-101\ncovers:\n  - source: 01.codemap.md\n    hash: sha256:STALE\n---\n" +
		"## Inspections\n\n| Obligation | Territory | Verdict | Evidence |\n| --- | --- | --- | --- |\n" +
		"| auth handler | auth.go | matches | `go build`; auth.go:1 |\n"
	if err := os.WriteFile(filepath.Join(c3Dir, "changes", unit, "02.inspect.md"), []byte(insp), 0o644); err != nil {
		t.Fatal(err)
	}
	rejects, err := inspectionGate(s, c3Dir, unit, nil, readCarriers(t, c3Dir, unit))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(rejects, "\n"), "stale") {
		t.Errorf("expected a stale reject, got: %v", rejects)
	}
}

// The gate must REFUSE an ungrounded rubber-stamp — a "matches" whose evidence
// names no file in the resolved territory.
func TestInspectionGate_RefusesEvidenceOutsideTerritory(t *testing.T) {
	s, c3Dir, unit, hash := gateFixture(t)
	insp := "---\ntarget: c3-101\ncovers:\n  - source: 01.codemap.md\n    hash: " + hash + "\n---\n" +
		"## Inspections\n\n| Obligation | Territory | Verdict | Evidence |\n| --- | --- | --- | --- |\n" +
		"| auth handler | somewhere | matches | looked at it, all good |\n"
	if err := os.WriteFile(filepath.Join(c3Dir, "changes", unit, "02.inspect.md"), []byte(insp), 0o644); err != nil {
		t.Fatal(err)
	}
	rejects, err := inspectionGate(s, c3Dir, unit, nil, readCarriers(t, c3Dir, unit))
	if err != nil {
		t.Fatal(err)
	}
	if len(rejects) == 0 {
		t.Fatal("expected a reject: evidence cites nothing in territory")
	}
}
