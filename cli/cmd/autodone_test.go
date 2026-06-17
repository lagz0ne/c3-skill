package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// Item 6b — AUTO-DONE: per-row After-cites + the one-way actualization latch.
//
// SETTLED (reverses "never auto-set done"). When all per-row After cites in the
// STRICT change-set of an `accepted` change doc resolve fresh (Item 5's check),
// the doc auto-flips `accepted->done`, one-way, loudly. Any stale/missing After
// cite -> no flip; the unresolved cites are reported. The latch reuses the SAME
// freshness machinery as Item 5 and NEVER judges whether the chosen After
// conditions were the right success criteria.

// seedAcceptedPRD inserts a prd change doc whose STRICT change-set (Requirements +
// Story Traces) carries the given Evidence handles, sets it to `accepted` via the
// sanctioned status writer, and returns the entity + its body. handle1/handle2 are
// the cite cells used as the per-row After cites.
func seedAcceptedPRD(t *testing.T, s *store.Store, handle1, handle2 string) (*store.Entity, string) {
	t.Helper()
	body := acceptedPRDBody(handle1, handle2)
	e := &store.Entity{ID: "prd-autodone", Type: "prd", Title: "Auth rollout", Slug: "auth-rollout", Status: "open", Metadata: "{}"}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("seed prd: %v", err)
	}
	// Status is edit-proof (Item 2): move open->accepted via the dedicated writer.
	if err := s.SetEntityStatus(e.ID, "accepted"); err != nil {
		t.Fatalf("set accepted: %v", err)
	}
	got, err := s.GetEntity(e.ID)
	if err != nil {
		t.Fatalf("reget prd: %v", err)
	}
	return got, body
}

func acceptedPRDBody(handle1, handle2 string) string {
	return "# Auth rollout\n\n## Goal\n\nRoll out authenticated API access.\n\n" +
		"## Requirements\n\n" +
		"| Requirement | Priority | Evidence |\n" +
		"| --- | --- | --- |\n" +
		"| Tokens validated at boundary | must | " + handle1 + " |\n\n" +
		"## Story Traces\n\n" +
		"| Story | Status | Evidence |\n" +
		"| --- | --- | --- |\n" +
		"| story-login | done | " + handle2 + " |\n"
}

func TestAutoDone_AllAfterCitesFreshFlipsAcceptedToDone(t *testing.T) {
	s := createRichDBFixture(t)
	h1 := testCitationForEntity(t, s, "c3-1")
	h2 := testCitationForEntity(t, s, "c3-101")
	entity, body := seedAcceptedPRD(t, s, h1, h2)

	// The flip is gated behind --fix: drive the commit path.
	flipped, unresolved := autoDoneLatch(s, "", entity, body, true)
	if !flipped {
		t.Fatalf("expected accepted doc with all-fresh After cites to auto-flip to done; unresolved=%+v", unresolved)
	}
	if len(unresolved) != 0 {
		t.Fatalf("all-fresh latch should report no unresolved cites, got %+v", unresolved)
	}
	got, err := s.GetEntity(entity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "done" {
		t.Fatalf("status should be persisted as done, got %q", got.Status)
	}
}

// TestAutoDone_PlainCheckDoesNotFlip — the flip is gated behind --fix. A plain
// `check` (commit==false / opts.Fix==false) must REPORT readiness but leave the
// status unchanged: it never mutates the DB or rewrites sealed markdown on a read.
func TestAutoDone_PlainCheckDoesNotFlip(t *testing.T) {
	s := createRichDBFixture(t)
	h1 := testCitationForEntity(t, s, "c3-1")
	h2 := testCitationForEntity(t, s, "c3-101")
	entity, body := seedAcceptedPRD(t, s, h1, h2)

	// Direct latch, no commit: ready==true but status NOT flipped.
	ready, unresolved := autoDoneLatch(s, "", entity, body, false)
	if !ready {
		t.Fatalf("plain (non-commit) latch should report readiness when all After cites resolve fresh; unresolved=%+v", unresolved)
	}
	if entity.Status != "accepted" {
		t.Fatalf("plain latch must not mutate the passed entity status, got %q", entity.Status)
	}
	got, err := s.GetEntity(entity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "accepted" {
		t.Fatalf("plain check must leave status accepted (no DB mutation), got %q", got.Status)
	}

	// Full plain check path: status stays accepted AND a readiness info is reported.
	// Persist the body to the store's node tree so RunCheckV2 reads the same
	// change-set the latch sees (status is edit-proof, so this stays accepted).
	if err := content.WriteEntity(s, entity.ID, body); err != nil {
		t.Fatalf("persist prd body: %v", err)
	}
	var buf bytes.Buffer
	if err := RunCheckV2(CheckOptions{Store: s, IncludeADR: true, JSON: true, Only: []string{entity.ID}}, &buf); err != nil {
		// A plain check may still FAIL on other discharge issues; that is fine. The
		// invariant under test is only that status is not flipped and readiness is
		// surfaced — both checked below regardless of the check verdict.
		_ = err
	}
	after, err := s.GetEntity(entity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Status != "accepted" {
		t.Fatalf("plain RunCheckV2 must not flip status; got %q", after.Status)
	}
	out := buf.String()
	if !strings.Contains(out, "ready to auto-done") {
		t.Fatalf("plain check should report auto-done readiness, got:\n%s", out)
	}
}

func TestAutoDone_StaleAfterCiteDoesNotFlip(t *testing.T) {
	s := createRichDBFixture(t)
	h1 := testCitationForEntity(t, s, "c3-1")
	h2 := staleHashHandle(t, testCitationForEntity(t, s, "c3-101"))
	entity, body := seedAcceptedPRD(t, s, h1, h2)

	flipped, unresolved := autoDoneLatch(s, "", entity, body, true)
	if flipped {
		t.Fatalf("a stale After cite must NOT flip the latch")
	}
	if len(unresolved) == 0 {
		t.Fatalf("stale After cite should be reported as unresolved, got none")
	}
	got, err := s.GetEntity(entity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "accepted" {
		t.Fatalf("status must stay accepted on a stale cite, got %q", got.Status)
	}
}

func TestAutoDone_OneWayLatchNeverFlipsBackward(t *testing.T) {
	s := createRichDBFixture(t)
	h1 := testCitationForEntity(t, s, "c3-1")
	h2 := testCitationForEntity(t, s, "c3-101")

	// An `open` doc with fresh After cites is NOT auto-advanced: the latch only
	// fires on `accepted`.
	openEntity := &store.Entity{ID: "prd-open", Type: "prd", Title: "Open doc", Slug: "open-doc", Status: "open", Metadata: "{}"}
	if err := s.InsertEntity(openEntity); err != nil {
		t.Fatal(err)
	}
	if flipped, _ := autoDoneLatch(s, "", openEntity, acceptedPRDBody(h1, h2), true); flipped {
		t.Fatalf("an open doc must never be auto-advanced by the latch")
	}
	if got, _ := s.GetEntity(openEntity.ID); got.Status != "open" {
		t.Fatalf("open doc status must stay open, got %q", got.Status)
	}

	// A `done` doc with fresh After cites is never flipped backward (or forward):
	// the latch never traverses done->accepted or re-flips a terminal doc.
	doneEntity := &store.Entity{ID: "prd-done", Type: "prd", Title: "Done doc", Slug: "done-doc", Status: "open", Metadata: "{}"}
	if err := s.InsertEntity(doneEntity); err != nil {
		t.Fatal(err)
	}
	for _, st := range []string{"accepted", "done"} {
		if err := s.SetEntityStatus(doneEntity.ID, st); err != nil {
			t.Fatalf("seed %s: %v", st, err)
		}
	}
	got, _ := s.GetEntity(doneEntity.ID)
	if flipped, _ := autoDoneLatch(s, "", got, acceptedPRDBody(h1, h2), true); flipped {
		t.Fatalf("a done (terminal) doc must never be re-flipped by the latch")
	}
	if after, _ := s.GetEntity(doneEntity.ID); after.Status != "done" {
		t.Fatalf("done doc must stay done, got %q", after.Status)
	}
}

func TestAutoDone_UsesSameFreshnessCheckAsItem5(t *testing.T) {
	s := createRichDBFixture(t)
	// Build a handle that Item 5's validateCitationColumnValue itself flags stale
	// (stale version). The latch MUST reach the same verdict via the same
	// machinery — proof it does not roll its own freshness logic.
	staleVer := staleVersionHandle(t, testCitationForEntity(t, s, "c3-1"))
	if issues := validateCitationColumnValue(staleVer, mustEntity(t, s, "c3-1"), citeOpts(s)); len(issues) == 0 {
		t.Fatalf("precondition: Item 5 should flag the stale-version handle")
	}
	fresh := testCitationForEntity(t, s, "c3-101")
	entity, body := seedAcceptedPRD(t, s, staleVer, fresh)

	flipped, unresolved := autoDoneLatch(s, "", entity, body, true)
	if flipped {
		t.Fatalf("latch must agree with Item 5: a handle Item 5 flags stale must block the flip")
	}
	if len(unresolved) == 0 {
		t.Fatalf("the Item-5-stale handle should surface as unresolved")
	}
}

// NEGATIVE — the line. The latch checks only the mechanical "do all After cites
// resolve fresh?"; it must NOT judge whether the After conditions are the *right*
// success criteria. A doc whose After cites point at arbitrary-but-fresh evidence
// still auto-flips: the tool never second-guesses the chosen conditions.
func TestAutoDone_DoesNotJudgeSuccessCriteria(t *testing.T) {
	s := createRichDBFixture(t)
	// Both After cites point at a real, fresh, but topically unrelated ref. The
	// latch never opines on relevance — fresh is fresh.
	h := testCitationForEntity(t, s, "ref-error-handling")
	entity, body := seedAcceptedPRD(t, s, h, h)

	flipped, _ := autoDoneLatch(s, "", entity, body, true)
	if !flipped {
		t.Fatalf("latch must flip on fresh-but-arbitrary After cites; it never judges success criteria")
	}
	if !strings.HasPrefix(entity.ID, "prd-") {
		t.Fatalf("sanity: expected a prd change doc")
	}
}
