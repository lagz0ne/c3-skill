package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

type searchSemanticProvider struct {
	calls int
}

func (p *searchSemanticProvider) Embed(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error) {
	if strings.TrimSpace(text) == "" {
		return nil, false, nil
	}
	p.calls++
	vec := make([]float32, 384)
	vec[0] = 1
	return vec, true, nil
}

func TestSearch_DefaultEnsuresSemanticIndexAndReusesIt(t *testing.T) {
	provider := &searchSemanticProvider{}
	restore := store.SetSemanticProviderForTest(provider)
	defer restore()
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)

	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store: s,
		Query: "pool wait p95 latency",
		JSON:  true,
		Limit: 5,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	var research *SearchResultRow
	for i := range out.Results {
		if out.Results[i].ID == "research-note-20260605-api-latency" {
			research = &out.Results[i]
			break
		}
	}
	if research == nil {
		t.Fatalf("research note missing from default hybrid search: %+v", out.Results)
	}
	requireStringSliceContains(t, research.MatchSources, "semantic")
	count, err := s.SemanticIndexCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("default search should build semantic vectors on a fresh index")
	}
	firstCalls := provider.calls
	if firstCalls <= 1 {
		t.Fatalf("first default search should embed entities plus query, calls = %d", firstCalls)
	}

	buf.Reset()
	err = RunSearch(SearchOptions{
		Store: s,
		Query: "pool wait p95 latency",
		JSON:  true,
		Limit: 5,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if provider.calls != firstCalls+1 {
		t.Fatalf("repeat search should reuse entity vectors and embed only the query; calls = %d, want %d", provider.calls, firstCalls+1)
	}
}

func TestBDD_SearchHybridReturnsFTSAndGraphContext(t *testing.T) {
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)

	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "pool wait p95 latency",
		Hybrid:     true,
		NoSemantic: true,
		JSON:       true,
		Limit:      3,
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
	if !strings.Contains(top.Snippet, "p95") || !strings.Contains(top.Snippet, "pool wait") {
		t.Fatalf("snippet should mention p95 and pool wait, got %q", top.Snippet)
	}
	if top.Context.Component.ID != "c3-101" || top.Context.Ref.ID != "ref-latency-budget" || top.Context.Rule.ID != "rule-trace-context" {
		t.Fatalf("context missing component/ref/rule: %+v", top.Context)
	}

	buf.Reset()
	err = RunSearch(SearchOptions{
		Store:      s,
		Query:      "traceparent request_id",
		Hybrid:     true,
		NoSemantic: true,
		JSON:       true,
		Limit:      3,
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

func TestRunSearch_EnrichesRouteFromGraphAndEvalBindings(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	projectDir := filepath.Dir(c3Dir)
	if err := os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "auth", "login.ts"), []byte("export function login() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "auth lifecycle frontend backend",
		NoSemantic: true,
		JSON:       true,
		Limit:      5,
		ProjectDir: projectDir,
		C3Dir:      c3Dir,
	}, &buf); err != nil {
		t.Fatal(err)
	}

	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	var auth *SearchResultRow
	for i := range out.Results {
		if out.Results[i].ID == "c3-101" {
			auth = &out.Results[i]
			break
		}
	}
	if auth == nil {
		t.Fatalf("expected c3-101 search hit, got %+v", out.Results)
	}
	requireStringSliceContains(t, auth.Route.Facts, "c3-101")
	requireStringSliceContains(t, auth.Route.Facts, "ref-jwt")
	requireStringSliceContains(t, auth.Route.Anchors, "src/auth/*.ts")
	requireStringSliceContains(t, auth.Route.Lanes, "auth")
	if auth.Route.Hash == "" {
		t.Fatalf("route hash should be populated: %+v", auth.Route)
	}
	if strings.Contains(auth.Route.HashBasis, "full file") || strings.Contains(auth.Route.HashBasis, "line number") {
		t.Fatalf("route hash basis should stay stable, got %q", auth.Route.HashBasis)
	}
}

func TestRunSearch_PackReturnsDiverseCitedEvidence(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	projectDir := filepath.Dir(c3Dir)
	if err := os.MkdirAll(filepath.Join(projectDir, "src", "auth"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "auth", "login.ts"), []byte("export function login() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bindCode(t, c3Dir, "c3-101", "src/auth/*.ts")

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store: s, Query: "auth lifecycle frontend backend", NoSemantic: true,
		JSON: true, Limit: 3, Pack: true, ProjectDir: projectDir, C3Dir: c3Dir,
	}, &buf); err != nil {
		t.Fatal(err)
	}

	var out SearchPackOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search pack JSON: %v\n%s", err, buf.String())
	}
	if len(out.Evidence) < 2 {
		t.Fatalf("pack should preserve several candidates, got %+v", out.Evidence)
	}
	for _, row := range out.Evidence {
		if row.Citation == "" || row.RecordClass == "" || row.Evidence == "" {
			t.Fatalf("pack row lacks citation, class, or evidence: %+v", row)
		}
	}
	var auth *SearchPackEvidence
	for i := range out.Evidence {
		if out.Evidence[i].ID == "c3-101" {
			auth = &out.Evidence[i]
		}
	}
	if auth == nil {
		t.Fatalf("expected auth component evidence, got %+v", out.Evidence)
	}
	if auth.RecordClass != "fact" || auth.State != "" {
		t.Fatalf("fact classification must be deterministic without invented state: %+v", auth)
	}
	requireStringSliceContains(t, auth.SourceAnchors, "src/auth/*.ts")
}

func TestRunSearch_PackAgentOutputIsCompactTOON(t *testing.T) {
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store: s, Query: "pool wait p95 latency", NoSemantic: true, Limit: 3, Pack: true,
	}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	requireAll(t, out, "evidence[", "{id,type,class,state,cite,why,lanes,anchors,record_claims,evidence}", "research-note-20260605-api-latency", "p95")
	if strings.Contains(out, "body:") || strings.Contains(out, "help[") || len(out) > 1800 {
		t.Fatalf("pack output should stay compact:\n%s", out)
	}
}

func TestCurrentBehaviorClaimsRanksMatchingSourceBackedRows(t *testing.T) {
	s := createDBFixture(t)
	mustInsertEntity(t, s, &store.Entity{
		ID: "c3-session", Type: "component", Title: "Session", Slug: "session",
		Goal: "Session lifecycle.", Status: "active", Metadata: "{}",
	})
	markdown := `# Session

## Current Behavior

| Mechanism | Class | Current claim | Source | Failure or absence |
|---|---|---|---|---|
| Logout | lifecycle | Logout expires only the presented browser cookie. | src/logout.ts:10 blob:abc | Other copied credentials remain valid. |
| Lookup | authorization | Request authentication reloads the user from the database. | src/auth.ts:20 blob:def | Missing users are rejected. |
| Styling | composition | Buttons share one variant function. | src/ui.ts:5 blob:ghi | N/A |
`
	if err := content.WriteEntity(s, "c3-session", markdown); err != nil {
		t.Fatal(err)
	}

	got, err := currentBehaviorClaims(s, "c3-session", "logout copied credential", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("claims = %#v, want two", got)
	}
	requireAll(t, got[0], "Logout", "[lifecycle]", "expires only", "copied credentials", "src/logout.ts:10")
	if strings.Contains(strings.Join(got, "\n"), "Buttons share") {
		t.Fatalf("unrelated row should not displace a better match: %#v", got)
	}
}

func TestRunSearch_AgentTOONOmitsGenericHelpHints(t *testing.T) {
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "pool wait",
		NoSemantic: true,
		JSON:       true,
		Limit:      3,
	}, &buf); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(buf.String(), "help[") {
		t.Fatalf("agent search success output should omit generic help hints:\n%s", buf.String())
	}
}

func TestRunSearch_AgentTOONUsesCompactSearchTable(t *testing.T) {
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "pool wait p95 latency",
		NoSemantic: true,
		Limit:      3,
	}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	requireAll(t, out,
		"results[3]{id,title,why,ctx,route,s}:",
		"research-note-20260605-api-latency",
		"p95",
		"DB pool",
		"body",
		"meta",
		"comp=c3-101",
		"ref=ref-latency-budget",
		"rule=rule-trace-context",
		"graph=",
		"lanes=",
		"hash=",
	)
	for _, noisy := range []string{"query:", "match_sources:", "context:", "component:\n", "title: Trace Context Propagation", "affects/", "uses/", "semantic", "help["} {
		if strings.Contains(out, noisy) {
			t.Fatalf("compact search output should not contain %q:\n%s", noisy, out)
		}
	}
	if len(out) > 1500 {
		t.Fatalf("compact search output too large: %d bytes\n%s", len(out), out)
	}
}

func TestCompactSearchRows_OmitsTitleDuplicateSnippet(t *testing.T) {
	rows := compactSearchRows([]SearchResultRow{
		{
			ID:           "c3-108",
			Title:        "eval-engine",
			Snippet:      "# eval-engine",
			MatchSources: []string{"content_fts", "entity_fts", "semantic"},
		},
		{
			ID:      "ref-eval-determinism",
			Title:   "eval-determinism",
			Snippet: "# Eval Determinism",
		},
	})
	if len(rows) != 2 {
		t.Fatalf("len = %d", len(rows))
	}
	for _, row := range rows {
		if row.Snippet != "" {
			t.Fatalf("duplicate title snippet should be omitted, got %+v", rows)
		}
	}
	if rows[0].Why != "body+meta+sem" {
		t.Fatalf("why = %q, want compact source labels", rows[0].Why)
	}
}

func TestRunSearch_AgentDefaultLimitIsFive(t *testing.T) {
	s := createDBFixture(t)
	for i := 0; i < 8; i++ {
		id := fmt.Sprintf("ref-agent-limit-%d", i)
		mustInsertEntity(t, s, &store.Entity{
			ID: id, Type: "ref", Title: fmt.Sprintf("Agent Limit %d", i), Slug: id,
			Goal: "zzagentlimit default search limit fixture.", Status: "active", Metadata: "{}",
		})
		if err := content.WriteEntity(s, id, "# Agent Limit\n\n## Goal\n\nzzagentlimit default search limit fixture.\n"); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "zzagentlimit",
		NoSemantic: true,
	}, &buf); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "results[5]{") {
		t.Fatalf("agent default limit should be 5 results:\n%s", buf.String())
	}
}

func TestCompactSearchSnippetBoundsLongRows(t *testing.T) {
	long := strings.Repeat("token ", 80)
	got := compactSearchSnippet(long)
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("truncated snippet should end with ellipsis marker: %q", got)
	}
	if len([]rune(got)) > compactSearchSnippetMax+3 {
		t.Fatalf("snippet len = %d, want <= %d: %q", len([]rune(got)), compactSearchSnippetMax+3, got)
	}
}

func TestSearch_HyphenatedNaturalLanguageQueryDoesNotReachFTSSyntax(t *testing.T) {
	s := createDBFixture(t)
	mustInsertEntity(t, s, &store.Entity{
		ID: "ref-realtime-sync", Type: "ref", Title: "Realtime Sync", Slug: "realtime-sync",
		Goal: "Real time sync keeps websocket clients current without polling.", Status: "active", Metadata: "{}",
	})
	if err := content.WriteEntity(s, "ref-realtime-sync", `# Realtime Sync

## Goal

Real time sync keeps websocket clients current without polling.
`); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "real-time sync",
		NoSemantic: true,
		JSON:       true,
		Limit:      5,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	if len(out.Results) == 0 {
		t.Fatalf("expected result for hyphenated natural-language query, got none: %s", buf.String())
	}
	if out.Results[0].ID != "ref-realtime-sync" {
		t.Fatalf("top result = %q, want ref-realtime-sync; results: %+v", out.Results[0].ID, out.Results)
	}
}

func TestSearch_DefaultUnavailableSemanticModelDegradesToKeywordGraph(t *testing.T) {
	t.Setenv("C3_SEMANTIC_CACHE_DIR", t.TempDir())
	t.Setenv("C3_SEMANTIC_OFFLINE", "1")
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

func TestSearch_NoSemanticSkipsAutoEnsure(t *testing.T) {
	provider := &searchSemanticProvider{}
	restore := store.SetSemanticProviderForTest(provider)
	defer restore()
	s := createDBFixture(t)
	seedHybridSearchFixture(t, s)

	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store:      s,
		Query:      "pool wait p95 latency",
		NoSemantic: true,
		JSON:       true,
		Limit:      3,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if provider.calls != 0 {
		t.Fatalf("--no-semantic should skip entity and query embedding; calls = %d", provider.calls)
	}
	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	for _, result := range out.Results {
		if hasSource(result.MatchSources, "semantic") {
			t.Fatalf("unexpected semantic source with --no-semantic: %+v", result)
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

func TestSearchSafetyProof_FailsClosedWithoutCertificate(t *testing.T) {
	if (SearchSafetyProof{EntityID: "ctx", Classification: "safe_context"}).Valid() {
		t.Fatal("partial safety proof must not validate")
	}
	if (SearchSafetyProof{EntityID: "ctx", Classification: "safe_context", EvidenceKind: "typed_context_relationship", EvidenceID: "edge-1", SourceHash: "hash"}).Valid() == false {
		t.Fatal("complete generic safety proof should validate")
	}
	conflict := SearchSafetyProof{EntityID: "ctx", Classification: "safe_context", EvidenceKind: "typed_context_relationship", EvidenceID: "edge-1", SourceHash: "hash", Conflict: true}
	if conflict.Valid() {
		t.Fatal("conflicting safety proof must fail closed")
	}
}

func TestBoundRouteWitness_RequiresExactBinding(t *testing.T) {
	witness := BoundRouteWitness{EntityID: "owner", MatchSource: "graph:affects:anchor", GraphFromID: "owner", GraphRelType: "affects", GraphToID: "anchor", DirectFTSEntityMissID: "owner", DirectFTSContentMissIDs: []string{"node-1"}}
	if !witness.Bound() {
		t.Fatal("exact route witness should be bound")
	}
	witness.GraphFromID = "other"
	if witness.Bound() {
		t.Fatal("unbound graph endpoint must not earn route credit")
	}
	witness = BoundRouteWitness{EntityID: "owner", MatchSource: "graph:affects:anchor", GraphFromID: "owner", GraphRelType: "affects", GraphToID: "anchor", DirectFTSEntityHitIDs: []string{"owner"}, DirectFTSEntityMissID: "owner", DirectFTSContentMissIDs: []string{"node-1"}}
	if witness.Bound() {
		t.Fatal("a direct entity hit cannot be a route miss witness")
	}
}

func TestRouteSnapshotRejectsPostEnrichmentMismatch(t *testing.T) {
	snapshot := RouteEnrichmentSnapshot{EntityID: "owner", Expected: RouteEnrichment{Facts: []string{"owner"}, Hash: "hash"}}
	if err := validateRouteSnapshot(snapshot, RouteEnrichment{Facts: []string{"other"}, Hash: "hash"}); err == nil {
		t.Fatal("route mismatch must be rejected")
	}
}

func TestSearchProvenanceDoesNotLaunderPublicRows(t *testing.T) {
	row := SearchResultRow{ID: "owner", Type: "component", Title: "Owner", MatchSources: []string{"entity_fts"}}
	encoded, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "safe_context") || strings.Contains(string(encoded), "DirectFTS") {
		t.Fatalf("internal provenance leaked into public row: %s", encoded)
	}
}

func TestRunSearch_DefaultProjectionIsByteCompatible(t *testing.T) {
	s := createDBFixture(t)
	base := SearchOptions{Store: s, Query: "auth", NoSemantic: true, JSON: true, Limit: 5}
	var implicit, explicit bytes.Buffer
	if err := RunSearch(base, &implicit); err != nil {
		t.Fatal(err)
	}
	base.StructuralProjection = false
	if err := RunSearch(base, &explicit); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(implicit.Bytes(), explicit.Bytes()) {
		t.Fatalf("default search output changed when candidate flag is false:\nimplicit=%s\nexplicit=%s", implicit.String(), explicit.String())
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(implicit.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if _, ok := envelope["route_witnesses"]; ok {
		t.Fatal("default search must omit candidate route witnesses")
	}
}

func TestRunSearch_StructuralProjectionUsesCandidateAdapter(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store: s, Query: "auth", NoSemantic: true, JSON: true, Limit: 5,
		StructuralProjection: true,
	}, &buf); err != nil {
		t.Fatal(err)
	}
	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	var owner *SearchResultRow
	for i := range out.Results {
		if out.Results[i].ID == "c3-1" {
			owner = &out.Results[i]
		}
		if out.Results[i].ID == "c3-101" {
			t.Fatalf("candidate projection should replace direct auth hit, rows=%+v", out.Results)
		}
	}
	if owner == nil {
		t.Fatalf("candidate projection should emit the immediate owner: %+v", out.Results)
	}
	if len(owner.MatchSources) != 1 || owner.MatchSources[0] != "containment_parent:c3-101" {
		t.Fatalf("generated owner must not inherit direct child provenance: %+v", owner)
	}
}

func TestRunSearch_StructuralProjectionNoTargetFailsClosed(t *testing.T) {
	s := createDBFixture(t)
	mustInsertEntity(t, s, &store.Entity{ID: "orphan", Type: "ref", Title: "orphan", Slug: "orphan", Status: "active", Metadata: "{}"})
	if err := content.WriteEntity(s, "orphan", "# orphan\n\n## Goal\n\nAn orphan record.\n"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := RunSearch(SearchOptions{
		Store: s, Query: "orphan", NoSemantic: true, JSON: true, Limit: 5,
		StructuralProjection: true,
	}, &buf)
	if err == nil || !strings.Contains(err.Error(), "no structurally witnessed owner target") {
		t.Fatalf("candidate no-target response must fail closed, err=%v output=%s", err, buf.String())
	}
}

func TestRunSearch_CaptureProvenanceSerializesBoundRouteWitness(t *testing.T) {
	s := createDBFixture(t)
	mustInsertEntity(t, s, &store.Entity{ID: "route-anchor", Type: "ref", Title: "route-anchor", Slug: "route-anchor", Status: "active", Metadata: "{}"})
	mustInsertEntity(t, s, &store.Entity{ID: "route-owner", Type: "component", Title: "route-owner", Slug: "route-owner", Status: "active", Metadata: "{}"})
	if err := content.WriteEntity(s, "route-anchor", "# route-anchor\n\n## Goal\n\nThe route anchor.\n"); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "route-owner", "# route-owner\n\n## Goal\n\nThe route owner.\n"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddRelationship(&store.Relationship{FromID: "route-owner", ToID: "route-anchor", RelType: "uses"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{
		Store: s, Query: "route-anchor", NoSemantic: true, JSON: true, Limit: 5,
		CaptureProvenance: true,
	}, &buf); err != nil {
		t.Fatal(err)
	}
	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid candidate envelope: %v\n%s", err, buf.String())
	}
	if len(out.RouteWitnesses) != 1 {
		t.Fatalf("expected one bound route witness, got %+v; rows=%+v", out.RouteWitnesses, out.Results)
	}
	witness := out.RouteWitnesses[0]
	if witness.EntityID != "route-owner" || witness.GraphFromID != "route-owner" || witness.GraphToID != "route-anchor" || witness.GraphRelType != "uses" {
		t.Fatalf("route witness graph binding is incomplete: %+v", witness)
	}
	if witness.MatchSource != "graph:uses:route-anchor" || witness.DirectFTSEntityMissID != "route-owner" || len(witness.DirectFTSContentMissIDs) == 0 {
		t.Fatalf("route witness direct/source binding is incomplete: %+v", witness)
	}
	if witness.RouteFieldValues.Hash == "" {
		t.Fatalf("route witness should carry typed route fields: %+v", witness)
	}
	encoded := buf.String()
	if strings.Contains(encoded, "oracle") || strings.Contains(encoded, "markdown") || strings.Contains(encoded, "The route owner") {
		t.Fatalf("candidate provenance leaked oracle/raw content: %s", encoded)
	}
}

func TestStructuralOwnerProjection_RecoversImmediateParentAndGeneratesOwnRow(t *testing.T) {
	s := createDBFixture(t)
	rows := []SearchResultRow{{ID: "c3-101", Type: "component", Title: "child", Snippet: "child snippet", MatchSources: []string{"content_fts"}}}
	provenance := map[string]*searchProvenance{
		"c3-101": {OriginalDirect: true, Direct: store.DirectFTSProbe{QueryToken: "query"}},
	}

	got, gotProvenance, err := projectStructuralOwnerWitnessRows(s, "auth", "", rows, provenance)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "c3-1" {
		t.Fatalf("projected rows = %+v, want generated c3-1 parent", got)
	}
	if got[0].Type != "container" || got[0].Title != "api" || got[0].Snippet == "child snippet" {
		t.Fatalf("generated parent must use its own entity fields/snippet: %+v", got[0])
	}
	if len(got[0].MatchSources) != 1 || got[0].MatchSources[0] != "containment_parent:c3-101" {
		t.Fatalf("generated parent sources = %v, want containment-only source", got[0].MatchSources)
	}
	if gotProvenance["c3-1"] == nil || gotProvenance["c3-1"].OriginalDirect {
		t.Fatalf("generated parent must not inherit direct provenance: %+v", gotProvenance["c3-1"])
	}
}

func TestStructuralOwnerProjection_ExistingParentIsNotLaunderedAndNoCascade(t *testing.T) {
	s := createDBFixture(t)
	rows := []SearchResultRow{
		{ID: "c3-101", Type: "component", Title: "child", Snippet: "child snippet", MatchSources: []string{"content_fts"}},
		{ID: "c3-1", Type: "container", Title: "exact parent row", Snippet: "exact parent snippet", MatchSources: []string{"entity_fts"}, Context: SearchContext{Ref: EntityRef{ID: "ref"}}, Route: RouteEnrichment{Facts: []string{"c3-1"}}},
	}
	provenance := map[string]*searchProvenance{
		"c3-101": {OriginalDirect: true, Direct: store.DirectFTSProbe{QueryToken: "query"}},
		"c3-1":   {OriginalDirect: true, Direct: store.DirectFTSProbe{QueryToken: "query"}},
	}

	got, _, err := projectStructuralOwnerWitnessRows(s, "auth", "", rows, provenance)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "c3-1" {
		t.Fatalf("rows = %+v, want only existing parent", got)
	}
	if got[0].Title != "exact parent row" || got[0].Snippet != "exact parent snippet" || got[0].Context.Ref.ID != "ref" || len(got[0].Route.Facts) != 1 {
		t.Fatalf("existing parent fields were not preserved: %+v", got[0])
	}
	requireStringSliceContains(t, got[0].MatchSources, "entity_fts")
	requireStringSliceContains(t, got[0].MatchSources, "containment_parent:c3-101")
	if hasSource(got[0].MatchSources, "containment_parent:c3-1") || got[0].ID == "c3-0" {
		t.Fatalf("projection must be one-hop and must not cascade: %+v", got)
	}
}

func TestStructuralOwnerProjection_PreservesParentlessPeerAndFilteredMismatch(t *testing.T) {
	s := createDBFixture(t)
	mustInsertEntity(t, s, &store.Entity{ID: "peer", Type: "component", Title: "Peer", Slug: "peer", Status: "active", Metadata: "{}"})
	rows := []SearchResultRow{
		{ID: "c3-101", Type: "component", Title: "child", MatchSources: []string{"entity_fts"}},
		{ID: "peer", Type: "component", Title: "Peer", MatchSources: []string{"semantic"}},
	}
	provenance := map[string]*searchProvenance{
		"c3-101": {OriginalDirect: true},
		"peer":   {OriginalDirect: true, Safety: SearchSafetyProof{EntityID: "peer", Classification: "safe_context", EvidenceKind: "typed_context", EvidenceID: "edge-peer", SourceHash: "hash"}},
	}
	got, _, err := projectStructuralOwnerWitnessRows(s, "auth", "", rows, provenance)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSearchID(got, "c3-1") || !containsSearchID(got, "peer") {
		t.Fatalf("owner and parentless peer must both survive: %+v", got)
	}

	filtered := map[string]*searchProvenance{
		"c3-101": {OriginalDirect: true, Safety: SearchSafetyProof{EntityID: "c3-101", Classification: "safe_context", EvidenceKind: "typed_context", EvidenceID: "edge-child", SourceHash: "hash"}},
	}
	got, _, err = projectStructuralOwnerWitnessRows(s, "auth", "component", rows[:1], filtered)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "c3-101" {
		t.Fatalf("filtered parent type mismatch must preserve child: %+v", got)
	}
}

func TestStructuralOwnerProjection_NoTargetSafetyAndSemanticOnlyFailClosed(t *testing.T) {
	s := createDBFixture(t)
	rows := []SearchResultRow{{ID: "semantic", Type: "ref", Title: "semantic", MatchSources: []string{"semantic"}}}
	provenance := map[string]*searchProvenance{"semantic": {Safety: SearchSafetyProof{EntityID: "semantic", Classification: "safe_context", EvidenceKind: "typed_context", EvidenceID: "edge", SourceHash: "hash"}}}
	// Semantic-only rows cannot manufacture an owner, but a complete safety
	// certificate lets them remain as neutral context in a no-target response.
	mustInsertEntity(t, s, &store.Entity{ID: "semantic", Type: "ref", Title: "semantic", Slug: "semantic", Status: "active", Metadata: "{}"})
	got, _, err := projectStructuralOwnerWitnessRows(s, "query", "", rows, provenance)
	if err != nil || len(got) != 1 || got[0].ID != "semantic" {
		t.Fatalf("safe semantic-only context should be retained, rows=%+v err=%v", got, err)
	}
	provenance["semantic"].Safety = SearchSafetyProof{}
	if _, _, err := projectStructuralOwnerWitnessRows(s, "query", "", rows, provenance); err == nil {
		t.Fatal("no-target response without a complete safety proof must fail closed")
	}
}

func TestStructuralOwnerProjection_MissingCycleAndLookupErrorFailClosed(t *testing.T) {
	s := createDBFixture(t)
	missingRows := []SearchResultRow{{ID: "c3-101", Type: "component", Title: "Missing"}}
	missingProv := map[string]*searchProvenance{"c3-101": {OriginalDirect: true, HasParentWitness: true, Parent: store.ImmediateParentWitness{ChildID: "c3-101", ParentID: "missing", RequestedTypeFilter: "", FilterMode: store.ParentFilterDefault, ParentRead: "missing", CycleChecked: true}, Safety: SearchSafetyProof{EntityID: "c3-101", Classification: "safe_context", EvidenceKind: "typed_context", EvidenceID: "edge", SourceHash: "hash"}}}
	got, _, err := projectStructuralOwnerWitnessRows(s, "query", "", missingRows, missingProv)
	if err != nil || len(got) != 1 || got[0].ID != "c3-101" {
		t.Fatalf("missing parent should stay an explicitly safe row, rows=%+v err=%v", got, err)
	}

	child, err := s.GetEntity("c3-101")
	if err != nil {
		t.Fatal(err)
	}
	child.ParentID = "c3-1"
	if err := s.UpdateEntity(child); err != nil {
		t.Fatal(err)
	}
	parent, err := s.GetEntity("c3-1")
	if err != nil {
		t.Fatal(err)
	}
	parent.ParentID = "c3-101"
	if err := s.UpdateEntity(parent); err != nil {
		t.Fatal(err)
	}
	cycleProv := map[string]*searchProvenance{"c3-101": {OriginalDirect: true, Safety: SearchSafetyProof{EntityID: "c3-101", Classification: "safe_context", EvidenceKind: "typed_context", EvidenceID: "edge", SourceHash: "hash"}}}
	got, _, err = projectStructuralOwnerWitnessRows(s, "query", "", []SearchResultRow{{ID: "c3-101", Type: "component"}}, cycleProv)
	if err != nil || len(got) != 1 || got[0].ID != "c3-101" {
		t.Fatalf("cycle must not project a parent, rows=%+v err=%v", got, err)
	}

	errorProv := map[string]*searchProvenance{"c3-101": {OriginalDirect: true, HasParentWitness: true, Parent: store.ImmediateParentWitness{ChildID: "c3-101", RequestedTypeFilter: "", FilterMode: store.ParentFilterDefault, ParentRead: "error"}}}
	if _, _, err := projectStructuralOwnerWitnessRows(s, "query", "", []SearchResultRow{{ID: "c3-101"}}, errorProv); err == nil {
		t.Fatal("parent lookup error must be a normal validation error")
	}
}

func TestStructuralOwnerProjection_BoundRouteOnlyAndStableDedup(t *testing.T) {
	s := createDBFixture(t)
	route := &BoundRouteWitness{EntityID: "route-owner", EntityContentIDs: []string{"node-1"}, MatchSource: "graph:affects:anchor", GraphFromID: "route-owner", GraphRelType: "affects", GraphToID: "anchor", DirectFTSEntityMissID: "route-owner", DirectFTSContentMissIDs: []string{"node-1"}}
	routeRow := SearchResultRow{ID: "route-owner", Type: "component", Title: "Route owner", MatchSources: []string{"graph:affects:anchor"}}
	validProv := map[string]*searchProvenance{"route-owner": {Route: route}}
	mustInsertEntity(t, s, &store.Entity{ID: "route-owner", Type: "component", Title: "Route owner", Slug: "route-owner", Status: "active", Metadata: "{}"})
	got, _, err := projectStructuralOwnerWitnessRows(s, "query", "", []SearchResultRow{routeRow}, validProv)
	if err != nil || len(got) != 1 || got[0].ID != "route-owner" {
		t.Fatalf("bound route owner should survive: rows=%+v err=%v", got, err)
	}
	invalid := &BoundRouteWitness{EntityID: "route-owner", EntityContentIDs: []string{"node-1"}, MatchSource: "graph:affects:anchor", GraphFromID: "other", GraphRelType: "affects", GraphToID: "anchor", DirectFTSEntityMissID: "route-owner", DirectFTSContentMissIDs: []string{"node-1"}}
	if got, _, err := projectStructuralOwnerWitnessRows(s, "query", "", []SearchResultRow{routeRow}, map[string]*searchProvenance{"route-owner": {Route: invalid}}); err != nil || len(got) != 0 {
		t.Fatalf("unbound graph row should be omitted, rows=%+v err=%v", got, err)
	}

	rows := []SearchResultRow{{ID: "c3-101", Type: "component", MatchSources: []string{"entity_fts"}}, {ID: "c3-110", Type: "component", MatchSources: []string{"entity_fts"}}}
	prov := map[string]*searchProvenance{"c3-101": {OriginalDirect: true}, "c3-110": {OriginalDirect: true}}
	got, _, err = projectStructuralOwnerWitnessRows(s, "query", "", rows, prov)
	if err != nil || len(got) != 1 || got[0].ID != "c3-1" {
		t.Fatalf("same parent targets should deduplicate stably: rows=%+v err=%v", got, err)
	}
	requireStringSliceContains(t, got[0].MatchSources, "containment_parent:c3-101")
	requireStringSliceContains(t, got[0].MatchSources, "containment_parent:c3-110")
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
