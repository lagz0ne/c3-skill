package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ---------------------------------------------------------------------------
// T4.1 — de-hardcode MinWords + section-name hints
// ---------------------------------------------------------------------------

// TestValidateTextSubstance_HonorsCanvasMinWords proves the thin-text check is
// driven by SectionDef.MinWords, not the old hardcoded 4/12 literals. A canvas
// declaring min_words:8 rejects a 6-word value; min_words:0 never trips.
func TestValidateTextSubstance_HonorsCanvasMinWords(t *testing.T) {
	sixWords := "one two three four five six"

	// min_words:8 must reject a 6-word value.
	issues := validateTextSubstance("Goal", sixWords, 8, "error")
	if !hasIssue(issues, "Goal") {
		t.Fatalf("expected MinWords=8 to reject 6-word Goal, got %#v", issues)
	}

	// min_words:6 must accept the same 6-word value (boundary, not thin).
	issues = validateTextSubstance("Goal", sixWords, 6, "error")
	if hasIssue(issues, "too thin") {
		t.Fatalf("expected 6 words to satisfy MinWords=6, got %#v", issues)
	}

	// min_words:0 means "no thin check" — any non-empty value passes.
	issues = validateTextSubstance("Goal", "tiny", 0, "error")
	if hasIssue(issues, "too thin") {
		t.Fatalf("expected MinWords=0 to skip thin check, got %#v", issues)
	}
}

// TestStrictIssue_HintNamesActualSections proves the hint names the actual
// doc-type section set rather than the hardcoded component section literal.
func TestStrictIssue_HintNamesActualSections(t *testing.T) {
	// prd sections: Goal (free), Requirements, Story Traces.
	prdDefs := schema.ForType("prd")
	prdHint := strictHintFor(prdDefs)
	if strings.Contains(prdHint, "Parent Fit") || strings.Contains(prdHint, "Foundational Flow") {
		t.Fatalf("hint named component sections for prd: %q", prdHint)
	}
	if !strings.Contains(prdHint, "Requirements") {
		t.Fatalf("expected prd hint to name actual section Requirements, got %q", prdHint)
	}

	// component keeps naming its own sections.
	compHint := strictHintFor(schema.ForType("component"))
	if !strings.Contains(compHint, "Parent Fit") {
		t.Fatalf("expected component hint to still name Parent Fit, got %q", compHint)
	}
}

// ---------------------------------------------------------------------------
// T4.2 — format + field/type check + touch-nothing FAIL on the STRICT change-set
// ---------------------------------------------------------------------------

// prdChangeDocBody returns a shape-valid prd body whose STRICT change-set rows
// can be mutated by callers. Goal is FREE narrative; Requirements + Story Traces
// are the STRICT change-set blocks.
func prdChangeDocBody(req, priority, evidence string) string {
	return "# Sample PRD\n\n" +
		"## Goal\n\n" +
		"Ship a reviewer-ready product requirements document for the sample feature.\n\n" +
		"## Requirements\n\n" +
		"| Requirement | Priority | Evidence |\n|---|---|---|\n" +
		"| " + req + " | " + priority + " | " + evidence + " |\n\n" +
		"## Story Traces\n\n" +
		"| Story | Status | Evidence |\n|---|---|---|\n" +
		"| story owns requirement delivery | done | ./docs/story.md |\n"
}

func TestChangeSet_FormatCheckMatchesCanvasShape(t *testing.T) {
	prdDefs := schema.ForType("prd")
	// Mis-shaped: rename the Priority column header so the change-set does NOT
	// match the canvas-declared shape.
	body := strings.Replace(
		prdChangeDocBody("deliver the export feature", "must", "./docs/export.md"),
		"| Requirement | Priority | Evidence |",
		"| Requirement | Importance | Evidence |", 1)

	issues := validateStrictDoc(prdDefs, body, "error")
	if !hasIssue(issues, "wrong table columns in Requirements") {
		t.Fatalf("expected change-set format check to FAIL for mis-shaped columns, got %#v", issues)
	}
}

func TestChangeSet_FieldTypeCheck(t *testing.T) {
	prdDefs := schema.ForType("prd")
	// Priority is a declared enum {must,should,could,wont}. A wrong-typed value
	// must FAIL the field/type check.
	body := prdChangeDocBody("deliver the export feature", "urgent", "./docs/export.md")

	issues := validateStrictDoc(prdDefs, body, "error")
	if !hasIssue(issues, "invalid enum value in Requirements row 1 column Priority") {
		t.Fatalf("expected change-set field/type check to FAIL for wrong-typed Priority, got %#v", issues)
	}
}

// TestChangeDoc_TouchNothingRejected proves an empty / all-N.A affected set is
// rejected — closes the empty-`expected` early return in coverage.
func TestChangeDoc_TouchNothingRejected(t *testing.T) {
	s := createRichDBFixture(t)
	body := adrTouchNothingBody()

	issues := validateADRCoverage(s, body, "error")
	if !hasIssue(issues, "touches nothing") {
		t.Fatalf("expected touch-nothing change doc to be rejected, got %#v", issues)
	}
}

// TestChangeDoc_DischargeWiredForPrd proves a generic-shape-valid prd that
// touches nothing FAILS the discharge check (prd/atomic no longer toothless).
func TestChangeDoc_DischargeWiredForPrd(t *testing.T) {
	s := createRichDBFixture(t)
	if err := s.InsertEntity(&store.Entity{
		ID: "prd-20260601-empty", Type: "prd", Title: "Empty PRD", Slug: "empty",
		Status: "open", Date: "20260601", Metadata: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	// A prd body with a STRICT change-set whose every requirement row is N.A —
	// shape-valid but touches nothing.
	body := "# Empty PRD\n\n" +
		"## Goal\n\nA prd that intentionally touches nothing.\n\n" +
		"## Requirements\n\n" +
		"| Requirement | Priority | Evidence |\n|---|---|---|\n" +
		"| N.A - nothing required | N.A - nothing | N.A - nothing |\n\n" +
		"## Story Traces\n\n" +
		"| Story | Status | Evidence |\n|---|---|---|\n" +
		"| N.A - no story | N.A - no story | N.A - no story |\n"
	if err := content.WriteEntity(s, "prd-20260601-empty", body); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunCheckV2(CheckOptions{Store: s, IncludeADR: true, Only: []string{"prd-20260601-empty"}}, &buf)
	out := buf.String()
	if err == nil && !strings.Contains(out, "touches nothing") {
		t.Fatalf("expected toothless prd to FAIL discharge, got err=%v out=%q", err, out)
	}
}

// NEGATIVE — an N.A - <reason> reason passes shape; c3x never judges validity.
func TestChangeDoc_DoesNotJudgeNAReasonValidity(t *testing.T) {
	prdDefs := schema.ForType("prd")
	// A row whose cells are N.A - lol — well-formed N.A reasons. Shape passes;
	// the silly reason is NOT flagged.
	body := "# Sample PRD\n\n" +
		"## Goal\n\nShip a reviewer-ready product requirements document.\n\n" +
		"## Requirements\n\n" +
		"| Requirement | Priority | Evidence |\n|---|---|---|\n" +
		"| deliver the export feature | must | ./docs/export.md |\n" +
		"| N.A - lol | N.A - lol | N.A - lol |\n\n" +
		"## Story Traces\n\n" +
		"| Story | Status | Evidence |\n|---|---|---|\n" +
		"| story owns requirement delivery | done | ./docs/story.md |\n"

	issues := validateStrictDoc(prdDefs, body, "error")
	for _, issue := range issues {
		if strings.Contains(issue.Message, "lol") {
			t.Fatalf("c3x judged an N.A reason's validity: %q", issue.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// T4.3 — top-down completeness + content-guided coverage + ceremony cliff
// ---------------------------------------------------------------------------

// TestChangeSet_TopDownCompletenessRequired proves a change-set naming a
// component (or container) but omitting its higher-level rows is flagged; adding
// those rows (delta or N.A - reason) clears it. Descent system->container->
// component is enforced by walking ParentID (<=3 deep).
func TestChangeSet_TopDownCompletenessRequired(t *testing.T) {
	s := createRichDBFixture(t)

	// Component c3-101 (parent c3-1, grandparent c3-0) named WITHOUT its higher
	// rows -> incomplete top-down.
	missing := adrAffectedTopologyBody([]adrTopoRow{
		{Entity: "c3-101", Type: "component", Why: "auth behavior changes"},
	})
	issues := topDownCompletenessIssues(s, missing, "error")
	if !hasIssue(issues, "c3-1") || !hasIssue(issues, "c3-0") {
		t.Fatalf("expected missing higher-level (container/system) rows to be flagged, got %#v", issues)
	}

	// Add the higher rows (one as a delta, one as N.A - reason) -> cleared.
	complete := adrAffectedTopologyBody([]adrTopoRow{
		{Entity: "c3-0", Type: "system", Why: "N.A - system unchanged"},
		{Entity: "c3-1", Type: "container", Why: "container boundary touched"},
		{Entity: "c3-101", Type: "component", Why: "auth behavior changes"},
	})
	issues = topDownCompletenessIssues(s, complete, "error")
	if len(issues) != 0 {
		t.Fatalf("expected complete top-down set to clear, got %#v", issues)
	}
}

// TestChangeSet_ContentGuidedCoverageFromAffectedSet proves the coverage walk
// reads the affected set from the change-set (parents always; descendant cliff
// when a high node is named).
func TestChangeSet_ContentGuidedCoverageFromAffectedSet(t *testing.T) {
	s := createRichDBFixture(t)

	// Naming the system c3-0 must seed coverage that follows the change-set's
	// declared affected set down to its descendants (the cliff).
	affected := []adrAffectedTarget{{ID: "c3-0", Type: "system"}}
	cov := expectedADRCoverage(s, affected)
	if len(cov.refs) == 0 {
		t.Fatalf("expected content-guided coverage to follow the change-set affected set from c3-0, got empty")
	}
}

// TestChangeDoc_CeremonyCliffPreserved locks the down-walk obligation: marking a
// system affected still owes a row per descendant; relieved only by N.A - reason.
func TestChangeDoc_CeremonyCliffPreserved(t *testing.T) {
	s := createRichDBFixture(t)
	// Marking system c3-0 affected, but mentioning no compliance refs, must
	// surface a missing-coverage issue for the ref scoped via the descendant
	// container c3-1 (ref-jwt is scoped to c3-1).
	body := adrAffectedTopologyBody([]adrTopoRow{
		{Entity: "c3-0", Type: "system", Why: "system-wide change"},
	}) + "\n## Compliance Refs\n\n" +
		"| Ref | Why required | Evidence | Action |\n|---|---|---|---|\n" +
		"| N.A - none | N.A - none | N.A - none | N.A - none |\n\n" +
		"## Compliance Rules\n\n" +
		"| Rule | Why required | Evidence | Action |\n|---|---|---|---|\n" +
		"| N.A - none | N.A - none | N.A - none | N.A - none |\n"

	issues := validateADRCoverage(s, body, "warning")
	if !hasIssue(issues, "ref-jwt") {
		t.Fatalf("expected ceremony-cliff down walk to still owe ref-jwt for affected system c3-0, got %#v", issues)
	}
}

// NEGATIVE — a delta whose wording is "wrong" is NOT flagged; only presence/shape.
func TestChangeDoc_DoesNotJudgeDeltaCorrectness(t *testing.T) {
	prdDefs := schema.ForType("prd")
	// A requirement with arbitrary-but-present wording; shape is valid.
	body := prdChangeDocBody("this requirement is probably wrong but well-formed", "must", "./docs/export.md")

	issues := validateStrictDoc(prdDefs, body, "error")
	for _, issue := range issues {
		if strings.Contains(issue.Message, "wrong") {
			t.Fatalf("c3x judged delta correctness: %q", issue.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// helpers for change-doc topology bodies
// ---------------------------------------------------------------------------

type adrTopoRow struct {
	Entity string
	Type   string
	Why    string
}

func adrAffectedTopologyBody(rows []adrTopoRow) string {
	b := "# Sample ADR\n\n## Affected Topology\n\n" +
		"| Entity | Type | Why affected | Evidence | Governance review |\n" +
		"|---|---|---|---|---|\n"
	for _, r := range rows {
		b += "| " + r.Entity + " | " + r.Type + " | " + r.Why + " | N.A - test | N.A - test |\n"
	}
	return b
}

func adrTouchNothingBody() string {
	return "# Empty ADR\n\n## Affected Topology\n\n" +
		"| Entity | Type | Why affected | Evidence | Governance review |\n" +
		"|---|---|---|---|---|\n" +
		"| N.A - nothing | N.A - nothing | N.A - nothing | N.A - nothing | N.A - nothing |\n"
}
