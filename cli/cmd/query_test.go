package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunQuery_FTS(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "auth"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in output, got:\n%s", out)
	}
	// Should have numbered output
	if !strings.Contains(out, "1.") {
		t.Errorf("expected numbered list, got:\n%s", out)
	}
}

func TestRunQuery_WithTypeFilter(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "auth", TypeFilter: "ref"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	// ref-jwt mentions "auth tokens" — should match
	if !strings.Contains(out, "ref-jwt") {
		t.Errorf("expected ref-jwt in type-filtered output, got:\n%s", out)
	}
	// Should NOT contain component c3-101
	if strings.Contains(out, "c3-101") {
		t.Errorf("type filter should exclude c3-101, got:\n%s", out)
	}
}

func TestRunQuery_NoResults(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "zzzznonexistent"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	if !strings.Contains(buf.String(), "No results.") {
		t.Errorf("expected 'No results.' message, got:\n%s", buf.String())
	}
}

func TestRunQuery_JSON(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "auth", JSON: true}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	var results []store.SearchResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}

func TestRunQuery_EmptyQuery(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: ""}, &buf)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestRunQuery_WithLimit(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "auth", Limit: 1}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	// Should have at most 1 result
	if strings.Contains(out, "2.") {
		t.Errorf("expected at most 1 result with limit=1, got:\n%s", out)
	}
}

func TestRunQuery_FindsNodeContent(t *testing.T) {
	s := createDBFixture(t)
	// Insert a node with content that doesn't appear in any entity metadata.
	// "c3-101" entity exists but has no mention of "microservice" anywhere.
	n := &store.Node{
		EntityID: "c3-101",
		Type:     "paragraph",
		Seq:      1,
		Content:  "This component follows a microservice architecture pattern",
	}
	if _, err := s.InsertNode(n); err != nil {
		t.Fatalf("insert node: %v", err)
	}

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "microservice"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 from content search, got:\n%s", out)
	}
}

func TestRunQuery_DedupsResults(t *testing.T) {
	s := createDBFixture(t)
	// "auth" matches c3-101 entity metadata (title). Also add a node with "auth".
	n := &store.Node{
		EntityID: "c3-101",
		Type:     "paragraph",
		Seq:      1,
		Content:  "Handles auth token validation",
	}
	if _, err := s.InsertNode(n); err != nil {
		t.Fatalf("insert node: %v", err)
	}

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "auth"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	// c3-101 should appear exactly once (deduped).
	count := strings.Count(out, "c3-101")
	if count != 1 {
		t.Errorf("expected c3-101 exactly once, found %d times in:\n%s", count, out)
	}
}
