package cmd

import "testing"

// The LINE — standing negative guards. These lock c3x to MECHANICAL signals only.
// They pass once their feature lands and go RED only if a later change over-reaches
// into judgment. This file holds the guards owned by the slices built so far; it is
// part of every subsequent slice's regression gate.

// TestLine_AutoDoneNeverJudgesSuccessCriteria (slice 6b). The auto-done latch flips
// on a purely mechanical "all After cites fresh"; it must NEVER judge whether the
// chosen After conditions were the *right* success criteria. The SETTLED auto-done
// reversal replaces the old "never auto-advance status" guard — a sanctioned
// mechanical auto-advance (accepted->done) is allowed; only judgment-driven
// advancement is forbidden.
func TestLine_AutoDoneNeverJudgesSuccessCriteria(t *testing.T) {
	s := createRichDBFixture(t)
	// After cites resolve to a real, fresh, but topically arbitrary target. If the
	// latch judged "are these the right success criteria?", it would refuse; the
	// LINE forbids exactly that — fresh is fresh, the flip fires.
	arbitrary := testCitationForEntity(t, s, "ref-error-handling")
	entity, body := seedAcceptedPRD(t, s, arbitrary, arbitrary)

	flipped, unresolved := autoDoneLatch(s, "", entity, body, true)
	if !flipped {
		t.Fatalf("LINE: the latch must flip on fresh-but-arbitrary After cites (mechanical only); unresolved=%+v", unresolved)
	}
	got, err := s.GetEntity(entity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "done" {
		t.Fatalf("LINE: a mechanical auto-advance to done is sanctioned, got status %q", got.Status)
	}
}
