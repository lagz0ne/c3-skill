package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunDiff_ShowsChanges(t *testing.T) {
	s := createDBFixture(t)
	// createDBFixture inserts entities which logs "add" changelog entries.
	// Those entries are unmarked, so diff should show them.
	var buf bytes.Buffer
	err := RunDiff(s, false, "", false, &buf)
	if err != nil {
		t.Fatalf("RunDiff: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "+ ADDED") {
		t.Errorf("expected ADDED entries, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-0") {
		t.Errorf("expected c3-0 in output, got:\n%s", out)
	}
}

func TestRunDiff_Mark(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunDiff(s, true, "abc123", false, &buf)
	if err != nil {
		t.Fatalf("RunDiff mark: %v", err)
	}
	if !strings.Contains(buf.String(), "abc123") {
		t.Errorf("expected confirmation with commit hash, got:\n%s", buf.String())
	}

	// Now diff should show no changes
	buf.Reset()
	err = RunDiff(s, false, "", false, &buf)
	if err != nil {
		t.Fatalf("RunDiff after mark: %v", err)
	}
	if !strings.Contains(buf.String(), "No uncommitted changes.") {
		t.Errorf("expected no changes after mark, got:\n%s", buf.String())
	}
}

func TestRunDiff_NoChanges(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()

	var buf bytes.Buffer
	err = RunDiff(s, false, "", false, &buf)
	if err != nil {
		t.Fatalf("RunDiff: %v", err)
	}
	if !strings.Contains(buf.String(), "No uncommitted changes.") {
		t.Errorf("expected 'No uncommitted changes.', got:\n%s", buf.String())
	}
}

func TestRunDiff_JSON(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunDiff(s, false, "", true, &buf)
	if err != nil {
		t.Fatalf("RunDiff JSON: %v", err)
	}
	var entries []*store.ChangeEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(entries) == 0 {
		t.Error("expected changelog entries in JSON output")
	}
}

func TestRunDiff_ModifyEntries(t *testing.T) {
	s := createDBFixture(t)
	// Mark initial adds
	s.MarkChangelog("initial")

	// Now modify an entity to create update changelog entries
	e, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatalf("get entity: %v", err)
	}
	e.Goal = "Updated auth goal"
	if err := s.UpdateEntity(e); err != nil {
		t.Fatalf("update entity: %v", err)
	}

	var buf bytes.Buffer
	err = RunDiff(s, false, "", false, &buf)
	if err != nil {
		t.Fatalf("RunDiff: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "~ MODIFIED") {
		t.Errorf("expected MODIFIED entry, got:\n%s", out)
	}
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in output, got:\n%s", out)
	}
}
