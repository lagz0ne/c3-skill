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

func TestRunQuery_SpecialCharsNoError(t *testing.T) {
	s := createDBFixture(t)
	cases := []struct {
		name  string
		query string
	}{
		{"comma", "auth, handler"},
		{"period", "auth."},
		{"parens", "auth(handler)"},
		{"semicolon", "auth; handler"},
		{"email-like", "user@example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunQuery(QueryOptions{Store: s, Query: tc.query}, &buf)
			if err != nil {
				t.Errorf("RunQuery(%q) returned error: %v", tc.query, err)
			}
		})
	}
}

func TestRunQuery_PureSymbolsNoResults(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: ",,,"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery(',,,') returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "No results.") {
		t.Errorf("expected 'No results.' for pure-symbol query, got:\n%s", buf.String())
	}
}

func TestRunQuery_NoResultsHints(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "zzzznonexistent", TypeFilter: "ref"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "No results.") {
		t.Errorf("expected 'No results.' message, got:\n%s", out)
	}
	if !strings.Contains(out, "remove --type ref filter") {
		t.Errorf("expected type filter hint, got:\n%s", out)
	}
}

func TestRunQuery_ShowsResultCount(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	// Limit to 1 — should show truncation warning since "auth" matches multiple entities.
	err := RunQuery(QueryOptions{Store: s, Query: "auth", Limit: 1}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "showing") || !strings.Contains(out, "limit reached") {
		t.Errorf("expected truncation indicator when limit reached, got:\n%s", out)
	}
}

func TestRunQuery_SuggestsOnNoResults(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	// "auht" is a typo for "auth" — should suggest "auth" entity.
	err := RunQuery(QueryOptions{Store: s, Query: "auht"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "auth") {
		t.Errorf("expected 'auth' suggestion for typo 'auht', got:\n%s", out)
	}
}

func TestRunQuery_ShowsRefinementHints(t *testing.T) {
	s := createDBFixture(t)
	// Seed extra entities of different types to get a mixed result set.
	extra := &store.Entity{
		ID: "ref-auth-pattern", Type: "ref", Title: "Auth Pattern",
		Slug: "auth-pattern", Goal: "Auth standardization", Status: "active", Metadata: "{}",
	}
	if err := s.InsertEntity(extra); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Store: s, Query: "auth"}, &buf)
	if err != nil {
		t.Fatalf("RunQuery: %v", err)
	}
	out := buf.String()
	// Results contain both components and refs — should suggest narrowing with --type.
	if !strings.Contains(out, "--type") {
		t.Errorf("expected --type refinement hint for mixed results, got:\n%s", out)
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
