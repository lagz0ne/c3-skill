package store

import (
	"database/sql"
	"errors"
	"testing"
)

func TestSaveEvalMatch_RoundTripsLatestManifest(t *testing.T) {
	s := createTestStore(t)
	record := EvalMatchRecord{
		Fact:          "c3-101",
		Claim:         "claim",
		FactRoot:      "fact-root",
		EvalSpecHash:  "spec-hash",
		ExternalState: "external",
		Verdict:       "holds",
		Evidence:      []string{"contains all 2"},
		Units: []EvalMatchUnit{
			{Kind: "code_block", Key: "doc.md#code_block[0]", Digest: "abc123", Bytes: 42},
			{Kind: "command_line", Key: "grep#command_line[0]", Digest: "def456", Bytes: 7},
		},
		CacheUnits: []EvalMatchUnit{
			{Kind: "command", Key: "grep token file", Digest: "cmd123", Bytes: 15},
			{Kind: "command_input", Key: "file", Digest: "input123", Bytes: 99},
		},
	}
	if err := s.SaveEvalMatch(record); err != nil {
		t.Fatal(err)
	}

	got, err := s.EvalMatch("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if got.Fact != record.Fact || got.FactRoot != record.FactRoot || got.EvalSpecHash != record.EvalSpecHash || got.ExternalState != record.ExternalState {
		t.Fatalf("round-trip mismatch: got %+v want %+v", got, record)
	}
	if len(got.Evidence) != 1 || got.Evidence[0] != "contains all 2" {
		t.Fatalf("evidence mismatch: %+v", got.Evidence)
	}
	if len(got.Units) != 2 || got.Units[0].Kind != "code_block" || got.Units[1].Key != "grep#command_line[0]" {
		t.Fatalf("units mismatch: %+v", got.Units)
	}
	if len(got.CacheUnits) != 2 || got.CacheUnits[0].Kind != "command" || got.CacheUnits[1].Digest != "input123" {
		t.Fatalf("cache units mismatch: %+v", got.CacheUnits)
	}
}

func TestSaveEvalMatch_ReplacesUnits(t *testing.T) {
	s := createTestStore(t)
	first := EvalMatchRecord{
		Fact:          "c3-101",
		ExternalState: "old",
		Verdict:       "holds",
		Units: []EvalMatchUnit{
			{Kind: "line", Key: "a#line[0]", Digest: "old-a", Bytes: 1},
			{Kind: "line", Key: "a#line[1]", Digest: "old-b", Bytes: 1},
		},
		CacheUnits: []EvalMatchUnit{
			{Kind: "command_input", Key: "old", Digest: "old-cache", Bytes: 1},
			{Kind: "command_input", Key: "stale", Digest: "stale-cache", Bytes: 1},
		},
	}
	if err := s.SaveEvalMatch(first); err != nil {
		t.Fatal(err)
	}
	second := EvalMatchRecord{
		Fact:          "c3-101",
		ExternalState: "new",
		Verdict:       "drift",
		Evidence:      []string{"missing token"},
		Units:         []EvalMatchUnit{{Kind: "line", Key: "a#line[0]", Digest: "new-a", Bytes: 3}},
		CacheUnits:    []EvalMatchUnit{{Kind: "command_input", Key: "new", Digest: "new-cache", Bytes: 3}},
	}
	if err := s.SaveEvalMatch(second); err != nil {
		t.Fatal(err)
	}

	got, err := s.EvalMatch("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExternalState != "new" || got.Verdict != "drift" {
		t.Fatalf("record was not replaced: %+v", got)
	}
	if len(got.Units) != 1 || got.Units[0].Digest != "new-a" {
		t.Fatalf("units were not replaced: %+v", got.Units)
	}
	if len(got.CacheUnits) != 1 || got.CacheUnits[0].Digest != "new-cache" {
		t.Fatalf("cache units were not replaced: %+v", got.CacheUnits)
	}
}

func TestEvalMatch_NotFound(t *testing.T) {
	s := createTestStore(t)
	_, err := s.EvalMatch("missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("EvalMatch missing err = %v, want sql.ErrNoRows", err)
	}
}
