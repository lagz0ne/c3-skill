package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// seedChangeDoc inserts a change-doc entity with the given status into the store
// and returns its ID. Change docs carry the new canonical status set
// {open, accepted, done, superseded}.
func seedChangeDoc(t *testing.T, s *store.Store, status string) string {
	t.Helper()
	e := &store.Entity{
		ID:       "cd-1",
		Type:     "prd",
		Title:    "Test Change Doc",
		Slug:     "test-change-doc",
		Status:   status,
		Date:     "20260615",
		Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("seed change doc (status=%s): %v", status, err)
	}
	return e.ID
}

// TestRunSet_AllowsOpenToAccepted — open->accepted via the status command succeeds.
func TestRunSet_AllowsOpenToAccepted(t *testing.T) {
	s := createDBFixture(t)
	id := seedChangeDoc(t, s, "open")
	var buf bytes.Buffer

	if err := RunSet(SetOptions{Store: s, ID: id, Field: "status", Value: "accepted"}, &buf); err != nil {
		t.Fatalf("open -> accepted should succeed: %v", err)
	}

	entity, _ := s.GetEntity(id)
	if entity.Status != "accepted" {
		t.Errorf("status = %q after open -> accepted, want %q", entity.Status, "accepted")
	}
}

// TestRunSet_AllowsAcceptedToDone — accepted->done via the status command succeeds.
func TestRunSet_AllowsAcceptedToDone(t *testing.T) {
	s := createDBFixture(t)
	id := seedChangeDoc(t, s, "accepted")
	var buf bytes.Buffer

	if err := RunSet(SetOptions{Store: s, ID: id, Field: "status", Value: "done"}, &buf); err != nil {
		t.Fatalf("accepted -> done should succeed: %v", err)
	}

	entity, _ := s.GetEntity(id)
	if entity.Status != "done" {
		t.Errorf("status = %q after accepted -> done, want %q", entity.Status, "done")
	}
}

// TestRunSet_BlocksIllegalStatusJump — illegal jumps are rejected, the error names
// the legal next state, and the stored status is unchanged after rejection.
func TestRunSet_BlocksIllegalStatusJump(t *testing.T) {
	cases := []struct {
		name      string
		from      string
		to        string
		wantInErr string // a legal next state the error should name
	}{
		{"open->done", "open", "done", "accepted"},
		{"done->open", "done", "open", "superseded"},
		{"accepted->open", "accepted", "open", "done"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := createDBFixture(t)
			id := seedChangeDoc(t, s, tc.from)
			var buf bytes.Buffer

			err := RunSet(SetOptions{Store: s, ID: id, Field: "status", Value: tc.to}, &buf)
			if err == nil {
				t.Fatalf("expected error for %s->%s transition", tc.from, tc.to)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Errorf("error should name legal next state %q: %v", tc.wantInErr, err)
			}

			entity, _ := s.GetEntity(id)
			if entity.Status != tc.from {
				t.Errorf("status should remain %q after blocked %s->%s, got %q", tc.from, tc.from, tc.to, entity.Status)
			}
		})
	}
}

// TestRunSet_DoesNotJudgeReadiness (NEGATIVE / the LINE) — the manual status command
// rejects only the illegal jump. A manual accepted->done is permitted regardless of
// whether the work is "really done"; there is no "are you sure / not really done" gate.
func TestRunSet_DoesNotJudgeReadiness(t *testing.T) {
	s := createDBFixture(t)
	id := seedChangeDoc(t, s, "accepted")
	var buf bytes.Buffer

	err := RunSet(SetOptions{Store: s, ID: id, Field: "status", Value: "done"}, &buf)
	if err != nil {
		t.Fatalf("manual accepted -> done must succeed (takes author's word): %v", err)
	}

	out := buf.String()
	for _, gate := range []string{"are you sure", "really done", "not really", "not done", "confirm"} {
		if strings.Contains(strings.ToLower(out), gate) {
			t.Errorf("manual accepted -> done must not emit a readiness gate %q, got: %s", gate, out)
		}
	}

	entity, _ := s.GetEntity(id)
	if entity.Status != "done" {
		t.Errorf("status = %q after manual accepted -> done, want %q", entity.Status, "done")
	}
}

// TestStatusMap_ProvisionedFlaggedLossy — mapping the legacy ADR `provisioned` status
// onto the new canonical set surfaces a lossy signal (recorded for item 8), not a
// silent coercion.
func TestStatusMap_ProvisionedFlaggedLossy(t *testing.T) {
	mapped, lossy := mapADRStatus("provisioned")
	if !lossy {
		t.Errorf("mapping provisioned should be flagged lossy, got lossy=false (mapped=%q)", mapped)
	}
	if mapped != "done" {
		t.Errorf("provisioned should map onto %q, got %q", "done", mapped)
	}

	// A non-lossy legacy mapping must NOT be flagged: proposed->open is a clean fold.
	if m, l := mapADRStatus("proposed"); l {
		t.Errorf("mapping proposed should not be lossy, got lossy=true (mapped=%q)", m)
	}
}
