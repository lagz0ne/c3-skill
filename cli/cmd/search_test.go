package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestBDD_SearchHybridReturnsFTSAndGraphContext(t *testing.T) {
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)

	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store:  s,
		Query:  "pool wait p95 latency",
		Hybrid: true,
		JSON:   true,
		Limit:  3,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	if len(out.Results) == 0 {
		t.Fatal("expected at least one search result")
	}
	top := out.Results[0]
	if top.ID != "research-note-20260605-api-latency" {
		t.Fatalf("top result = %q, want research-note-20260605-api-latency; all results: %+v", top.ID, out.Results)
	}
	requireStringSliceContains(t, top.MatchSources, "content_fts")
	requireStringSliceContains(t, top.MatchSources, "graph:affects:c3-101")
	requireStringSliceContains(t, top.MatchSources, "graph:uses:ref-latency-budget")
	requireStringSliceContains(t, top.MatchSources, "graph:uses:rule-trace-context")
	requireStringSliceContains(t, top.MatchSources, "code-map:src/api/handlers/latency.go")
	if !strings.Contains(top.Snippet, "p95") || !strings.Contains(top.Snippet, "pool wait") {
		t.Fatalf("snippet should mention p95 and pool wait, got %q", top.Snippet)
	}
	if top.Context.Component.ID != "c3-101" || top.Context.Ref.ID != "ref-latency-budget" || top.Context.Rule.ID != "rule-trace-context" {
		t.Fatalf("context missing component/ref/rule: %+v", top.Context)
	}
	if top.Context.Path != "src/api/handlers/latency.go" {
		t.Fatalf("context path = %q", top.Context.Path)
	}

	buf.Reset()
	err = RunSearch(SearchOptions{
		Store:  s,
		Query:  "traceparent request_id",
		Hybrid: true,
		JSON:   true,
		Limit:  3,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid linked-context search JSON: %v\n%s", err, buf.String())
	}
	if len(out.Results) == 0 || out.Results[0].ID != "research-note-20260605-api-latency" {
		t.Fatalf("hybrid search should retrieve research note via linked rule/ref context, got: %+v", out.Results)
	}
	requireStringSliceContains(t, out.Results[0].MatchSources, "graph:uses:rule-trace-context")
}

func TestSearch_DefaultNoSemanticIndexStillReturnsKeywordGraph(t *testing.T) {
	t.Setenv("C3_SEMANTIC_CACHE_DIR", t.TempDir())
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)

	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store:  s,
		Query:  "pool wait p95 latency",
		Hybrid: true,
		JSON:   true,
		Limit:  3,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	if len(out.Results) == 0 {
		t.Fatal("expected keyword/graph results without semantic model")
	}
	if out.Results[0].ID != "research-note-20260605-api-latency" {
		t.Fatalf("top result = %q, want research-note-20260605-api-latency", out.Results[0].ID)
	}
	for _, result := range out.Results {
		if hasSource(result.MatchSources, "semantic") {
			t.Fatalf("unexpected semantic source without index/model: %+v", result)
		}
	}
}

func TestFuseSemanticRows_ReciprocalRankFusionPromotesAgreement(t *testing.T) {
	rows := []SearchResultRow{
		{ID: "keyword-only", Type: "component", Title: "Keyword Only", MatchSources: []string{"content_fts"}},
		{ID: "also-keyword", Type: "component", Title: "Also Keyword", MatchSources: []string{"entity_fts"}},
		{ID: "agrees", Type: "component", Title: "Agrees", MatchSources: []string{"entity_fts"}},
	}
	semantic := []store.SearchResult{
		{ID: "agrees", Type: "component", Title: "Agrees", Snippet: "semantic agreement"},
		{ID: "semantic-only", Type: "ref", Title: "Semantic Only", Snippet: "semantic only"},
	}

	got := fuseSemanticRows(rows, semantic, 4)
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}
	if got[0].ID != "agrees" {
		t.Fatalf("top = %q, want agrees; all: %+v", got[0].ID, got)
	}
	requireStringSliceContains(t, got[0].MatchSources, "semantic")
	if !containsSearchID(got, "semantic-only") {
		t.Fatalf("semantic-only hit should be retained by fusion: %+v", got)
	}
}

func seedHybridSearchFixture(t *testing.T, s *store.Store) {
	t.Helper()
	mustInsertEntity(t, s, &store.Entity{
		ID: "rule-trace-context", Type: "rule", Title: "Trace Context Propagation", Slug: "trace-context",
		Goal: "Every outbound API call carries traceparent and request_id.", Status: "active", Metadata: "{}",
	})
	mustInsertEntity(t, s, &store.Entity{
		ID: "ref-latency-budget", Type: "ref", Title: "Latency Budget", Slug: "latency-budget",
		Goal: "Keep API p95 under 250 ms before checkout release.", Status: "active", Metadata: "{}",
	})
	mustInsertEntity(t, s, &store.Entity{
		ID: "research-note-20260605-api-latency", Type: "research-note", Title: "API Latency Investigation", Slug: "api-latency",
		Goal: "Investigate checkout API latency pool wait regression.", Status: "active", Metadata: "{}",
	})
	auth, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	auth.Title = "api-latency-gateway"
	auth.Goal = "Own API request routing and latency instrumentation."
	if err := s.UpdateEntity(auth); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []*store.Relationship{
		{FromID: "c3-101", ToID: "ref-latency-budget", RelType: "uses"},
		{FromID: "c3-101", ToID: "rule-trace-context", RelType: "uses"},
		{FromID: "research-note-20260605-api-latency", ToID: "c3-101", RelType: "affects"},
		{FromID: "research-note-20260605-api-latency", ToID: "c3-101", RelType: "sources"},
		{FromID: "research-note-20260605-api-latency", ToID: "ref-latency-budget", RelType: "uses"},
		{FromID: "research-note-20260605-api-latency", ToID: "rule-trace-context", RelType: "uses"},
	} {
		if err := s.AddRelationship(rel); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetCodeMap("c3-101", []string{"src/api/**/*.go", "src/api/handlers/latency.go"}); err != nil {
		t.Fatal(err)
	}

	body := `# API Latency Investigation

## Summary

Checkout API p95 increased from 180 ms to 420 ms after the connection-pool change. Span evidence points to DB pool wait, not JWT validation.

## Findings

| Finding | Evidence |
|---|---|
| Pool wait increased during checkout summary fan-out while token validation stayed below 15 ms. | c3-101 |
`
	if err := content.WriteEntity(s, "research-note-20260605-api-latency", body); err != nil {
		t.Fatal(err)
	}
}

func mustInsertEntity(t *testing.T, s *store.Store, e *store.Entity) {
	t.Helper()
	if err := s.InsertEntity(e); err != nil {
		t.Fatal(err)
	}
}

func requireStringSliceContains(t *testing.T, got []string, want string) {
	t.Helper()
	for _, value := range got {
		if value == want {
			return
		}
	}
	t.Fatalf("slice missing %q: %v", want, got)
}

func containsSearchID(rows []SearchResultRow, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}
	return false
}
