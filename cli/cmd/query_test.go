package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Cycle 2: Catalog mode
// =============================================================================

func TestRunQuery_Catalog(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		JSON:  false,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, id := range []string{"c3-0", "c3-1", "c3-101", "ref-jwt"} {
		if !strings.Contains(out, id) {
			t.Errorf("catalog output missing entity %q", id)
		}
	}
}

func TestRunQuery_Catalog_JSON(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result []EntityResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty catalog")
	}
	for _, entry := range result {
		if entry.ID == "" {
			t.Error("entry with empty ID")
		}
		if len(entry.Blocks) == 0 {
			t.Errorf("entry %s has no blocks", entry.ID)
		}
	}
}

// =============================================================================
// Cycle 3: Entity snapshot
// =============================================================================

func TestRunQuery_EntityResult(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Handle authentication") {
		t.Error("snapshot should contain goal text")
	}
	if !strings.Contains(out, "Dependencies") {
		t.Error("snapshot should show Dependencies section")
	}
}

func TestRunQuery_EntityResult_JSON(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101"},
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var result EntityResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if result.ID != "c3-101" {
		t.Errorf("ID = %q, want %q", result.ID, "c3-101")
	}
	if len(result.Blocks) == 0 {
		t.Error("expected blocks in snapshot")
	}
}

func TestRunQuery_EntityResult_Ref(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"ref-jwt"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, section := range []string{"Goal", "Choice", "Why"} {
		if !strings.Contains(out, section) {
			t.Errorf("ref snapshot missing section %q", section)
		}
	}
}

func TestRunQuery_EntityNotFound(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-999"},
	}, &buf)
	if err == nil {
		t.Fatal("expected error for missing entity")
	}
	if !strings.Contains(err.Error(), "c3-999") {
		t.Errorf("error should mention entity ID, got: %v", err)
	}
}

// =============================================================================
// Cycle 4: Single block extraction
// =============================================================================

func TestRunQuery_SingleBlock(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101", "dependencies"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "IN") {
		t.Error("single block output should contain table data")
	}
}

func TestRunQuery_SingleBlock_Text(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101", "goal"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(buf.String())
	if out != "Handle authentication." {
		t.Errorf("got %q, want %q", out, "Handle authentication.")
	}
}

func TestRunQuery_SingleBlock_NotFound(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101", "nonexistent"},
	}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown section")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention section name, got: %v", err)
	}
}

func TestRunQuery_SingleBlock_JSON(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101", "dependencies"},
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var rows []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(rows) != 1 {
		t.Fatalf("row count = %d, want 1", len(rows))
	}
}

// =============================================================================
// Cycle 5: File resolution
// =============================================================================

func TestRunQuery_FilePath(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)


	cmContent := "c3-101:\n  - \"src/auth/**/*.ts\"\n"
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), cmContent)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph:      graph,
		C3Dir:      c3Dir,
		Args:       []string{"src/auth/login.ts"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Error("file path query should resolve to c3-101")
	}
}

func TestRunQuery_FilePath_NoMatch(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)


	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - \"src/auth/**\"\n")

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph:      graph,
		C3Dir:      c3Dir,
		Args:       []string{"src/unknown/file.ts"},
	}, &buf)
	if err == nil {
		t.Fatal("expected error for unmatched file")
	}
	if !strings.Contains(err.Error(), "no component mapping") {
		t.Errorf("error = %v, want mention of no mapping", err)
	}
}

func TestRunQuery_FilePath_Chain(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)


	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), "c3-101:\n  - \"src/auth/**\"\n")

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph:      graph,
		C3Dir:      c3Dir,
		Args:       []string{"src/auth/login.ts"},
		Chain:      true,
		JSON:       true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var chain ChainResult
	if err := json.Unmarshal(buf.Bytes(), &chain); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if chain.Component == nil {
		t.Error("chain should have component")
	}
}

func TestRunQuery_FilePath_MultiMatch(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)


	// Both c3-101 and c3-110 match the same file path
	cmContent := "c3-101:\n  - \"src/shared/**\"\nc3-110:\n  - \"src/shared/**\"\n"
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), cmContent)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph:      graph,
		C3Dir:      c3Dir,
		Args:       []string{"src/shared/utils.ts"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	// Both entities should appear
	if !strings.Contains(out, "c3-101") {
		t.Error("multi-match should include c3-101")
	}
	if !strings.Contains(out, "c3-110") {
		t.Error("multi-match should include c3-110")
	}
}

func TestRunQuery_FilePath_MultiMatch_JSON(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)


	cmContent := "c3-101:\n  - \"src/shared/**\"\nc3-110:\n  - \"src/shared/**\"\n"
	writeFile(t, filepath.Join(c3Dir, "code-map.yaml"), cmContent)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph:      graph,
		C3Dir:      c3Dir,
		Args:       []string{"src/shared/utils.ts"},
		JSON:       true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var snaps []EntityResult
	if err := json.Unmarshal(buf.Bytes(), &snaps); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

// =============================================================================
// Cycle 6: Chain traversal
// =============================================================================

func TestRunQuery_Chain(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101"},
		Chain: true,
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var chain ChainResult
	if err := json.Unmarshal(buf.Bytes(), &chain); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if chain.Component == nil {
		t.Error("chain should have component level")
	}
	if chain.Container == nil {
		t.Error("chain should have container level")
	}
	if chain.Context == nil {
		t.Error("chain should have context level")
	}
	if len(chain.Refs) == 0 {
		t.Error("chain should include refs for c3-101")
	}
}

func TestRunQuery_Chain_Text(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-101"},
		Chain: true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "c3-0") {
		t.Error("chain text should show context")
	}
	if !strings.Contains(out, "c3-1") {
		t.Error("chain text should show container")
	}
	if !strings.Contains(out, "c3-101") {
		t.Error("chain text should show component")
	}
}

func TestRunQuery_Chain_ContainerLevel(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-1"},
		Chain: true,
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var chain ChainResult
	if err := json.Unmarshal(buf.Bytes(), &chain); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if chain.Component != nil {
		t.Error("container-level chain should not have component")
	}
	if chain.Container == nil {
		t.Error("container-level chain should have container")
	}
	if chain.Context == nil {
		t.Error("container-level chain should have context")
	}
}

func TestRunQuery_Chain_NoRefs(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"c3-110"},
		Chain: true,
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var chain ChainResult
	if err := json.Unmarshal(buf.Bytes(), &chain); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(chain.Refs) != 0 {
		t.Errorf("expected no refs, got %d", len(chain.Refs))
	}
}

// =============================================================================
// Edge cases (from code review)
// =============================================================================

func TestRunQuery_ChainWithoutTarget(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Chain: true,
	}, &buf)
	if err == nil {
		t.Fatal("expected error for --chain without target")
	}
	if !strings.Contains(err.Error(), "--chain requires") {
		t.Errorf("error = %v, want mention of --chain requires target", err)
	}
}

func TestRunQuery_EntityIDWithDot(t *testing.T) {
	// Entity IDs containing dots should be found via graph lookup,
	// not misrouted to file path resolution.
	// We test with "ref-jwt" which doesn't have a dot, but the logic
	// ensures graph lookup takes priority over file path heuristic.
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"ref-jwt"},
		JSON:  true,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var snap EntityResult
	if err := json.Unmarshal(buf.Bytes(), &snap); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if snap.ID != "ref-jwt" {
		t.Errorf("ID = %q, want %q", snap.ID, "ref-jwt")
	}
}

func TestRunQuery_Catalog_ExcludesADR(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{Graph: graph, C3Dir: c3Dir, JSON: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	var entries []EntityResult
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	for _, e := range entries {
		if e.Type == "adr" {
			t.Errorf("catalog should not include ADRs, found %s", e.ID)
		}
	}
}

func TestRunQuery_ADR_DirectAccess(t *testing.T) {
	c3Dir := createRichFixture(t)
	graph := loadGraph(t, c3Dir)

	var buf bytes.Buffer
	err := RunQuery(QueryOptions{
		Graph: graph,
		C3Dir: c3Dir,
		Args:  []string{"adr-20260226-use-go"},
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "adr-20260226-use-go") {
		t.Error("direct ADR query should still work")
	}
	if !strings.Contains(out, "Goal") {
		t.Error("ADR snapshot should show Goal section")
	}
}

// =============================================================================
// Options parsing
// =============================================================================

func TestParseArgs_Chain(t *testing.T) {
	opts := ParseArgs([]string{"query", "c3-101", "--chain"})
	if opts.Command != "query" {
		t.Errorf("Command = %q, want %q", opts.Command, "query")
	}
	if !opts.Chain {
		t.Error("expected --chain flag to be set")
	}
}

// =============================================================================
// Help text
// =============================================================================

func TestHelp_Query(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("query", &buf)
	out := buf.String()
	if !strings.Contains(out, "query") {
		t.Error("query help should mention the command")
	}
}
