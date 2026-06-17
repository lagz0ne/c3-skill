package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

type evalCase struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Query    string   `json:"query"`
	Expected []string `json:"expected"`
}

type caseResult struct {
	evalCase
	Rank int      `json:"rank"`
	Hits []string `json:"hits"`
}

type metrics struct {
	CaseCount int     `json:"case_count"`
	HitAt1    float64 `json:"hit_at_1"`
	HitAt3    float64 `json:"hit_at_3"`
	HitAt5    float64 `json:"hit_at_5"`
	MRR       float64 `json:"mrr"`
}

type evalReport struct {
	Mode    string             `json:"mode"`
	DB      string             `json:"db"`
	Limit   int                `json:"limit"`
	Metrics metrics            `json:"metrics"`
	ByKind  map[string]metrics `json:"by_kind"`
	Cases   []caseResult       `json:"cases"`
}

var cases = []evalCase{
	// --- original 8 ---
	{ID: "lookup-owner-paraphrase", Kind: "paraphrase", Query: "which architecture unit owns a source path", Expected: []string{"c3-114"}},
	{ID: "binary-no-toolchain-paraphrase", Kind: "paraphrase", Query: "users can run c3x without building from source", Expected: []string{"ref-cross-compiled-binary"}},
	{ID: "intent-router-paraphrase", Kind: "paraphrase", Query: "send the user's request to the right workflow", Expected: []string{"c3-201"}},
	{ID: "agent-output-paraphrase", Kind: "paraphrase", Query: "machine responses use one formatter in agent mode", Expected: []string{"rule-output-via-helpers"}},
	{ID: "frontmatter-keyword", Kind: "keyword", Query: "yaml frontmatter markdown body", Expected: []string{"ref-frontmatter-docs"}},
	{ID: "templates-keyword", Kind: "keyword", Query: "embedded doc templates", Expected: []string{"ref-embedded-templates"}},
	{ID: "check-validation-keyword", Kind: "keyword", Query: "validate structural integrity docs refs rules", Expected: []string{"c3-113"}},
	{ID: "cache-lifecycle-keyword", Kind: "keyword", Query: "import export sync repair cache", Expected: []string{"c3-119"}},

	// --- CLI core libs (c3-101..110) ---
	{ID: "frontmatter-getset-keyword", Kind: "keyword", Query: "get and set YAML frontmatter fields in .c3 markdown files", Expected: []string{"c3-101"}},
	{ID: "walker-discover-entities-paraphrase", Kind: "paraphrase", Query: "recursively scan the docs folder to enumerate every container, component, ref, rule, and decision record", Expected: []string{"c3-102"}},
	{ID: "codemap-glob-match-keyword", Kind: "keyword", Query: "parse code-map.yaml and match files to components and rules with glob patterns", Expected: []string{"c3-105"}},
	{ID: "store-persistent-records-paraphrase", Kind: "paraphrase", Query: "durable database layer holding nodes, versions, hashes, and changelog rows for the CLI", Expected: []string{"c3-107"}},
	{ID: "runtime-rollback-keyword", Kind: "keyword", Query: "dispatcher self-healing preflight that repairs seal drift and rolls back mutating commands on canonical export failure", Expected: []string{"c3-108"}},
	{ID: "npm-wrapper-download-paraphrase", Kind: "paraphrase", Query: "node package that fetches the correct prebuilt executable for your platform, checks its signature, stores it, then launches it", Expected: []string{"c3-109", "ref-cross-compiled-binary"}},
	{ID: "init-scaffold-paraphrase", Kind: "paraphrase", Query: "bootstrap a fresh architecture-docs folder for a brand-new project, creating the config file and a starter decision record", Expected: []string{"c3-110"}},

	// --- CLI commands (c3-111..120) ---
	{ID: "add-adr-rollback-keyword", Kind: "keyword", Query: "add adr rollback canonical export all-or-nothing partial entity", Expected: []string{"c3-111"}},
	{ID: "create-entity-numbering-paraphrase", Kind: "paraphrase", Query: "command that allocates the next sequential id for a fresh container or rule and links it into its parent table", Expected: []string{"c3-111"}},
	{ID: "topology-output-modes-keyword", Kind: "keyword", Query: "output full .c3 topology flat compact json", Expected: []string{"c3-112"}},
	{ID: "print-doc-tree-paraphrase", Kind: "paraphrase", Query: "enumerate every documented node in the tree as plain text or machine form before running other checks", Expected: []string{"c3-112"}},
	{ID: "codemap-scaffold-stubs-keyword", Kind: "keyword", Query: "scaffold update code-map.yaml empty stubs idempotent re-run", Expected: []string{"c3-115"}},
	{ID: "bootstrap-mapping-paraphrase", Kind: "paraphrase", Query: "generate placeholder glob entries for each entity into the mapping file without clobbering existing patterns", Expected: []string{"c3-115"}},
	{ID: "coverage-governance-keyword", Kind: "keyword", Query: "code-map coverage mapped excluded unmapped file counts ref rule governance percentages", Expected: []string{"c3-116"}},
	{ID: "how-much-mapped-paraphrase", Kind: "paraphrase", Query: "tell me what fraction of the source tree is accounted for and how well refs and rules are attributed", Expected: []string{"c3-116"}},
	{ID: "schema-sections-keyword", Kind: "keyword", Query: "schema command print section names required marker purpose text table columns", Expected: []string{"c3-117", "c3-113"}},
	{ID: "read-write-status-paraphrase", Kind: "paraphrase", Query: "commands to load a canonical doc, save edits to it, stamp a field, and report its current condition", Expected: []string{"c3-117"}},
	{ID: "impact-blast-radius-paraphrase", Kind: "paraphrase", Query: "figure out the downstream consequences and affected pieces when one part of the architecture changes", Expected: []string{"c3-118"}},
	{ID: "query-graph-diff-impact-keyword", Kind: "keyword", Query: "query graph diff impact architecture change consequences", Expected: []string{"c3-118"}},
	{ID: "prune-versions-keyword", Kind: "keyword", Query: "versions hash nodes prune marketplace command families", Expected: []string{"c3-120"}},
	{ID: "snapshot-cleanup-paraphrase", Kind: "paraphrase", Query: "inspect past revisions, compute content fingerprints, and trim old snapshots, plus pulling shared skill packages", Expected: []string{"c3-120"}},

	// --- Claude Skill (c3-201..218) ---
	{ID: "migrate-version-cache-keyword", Kind: "keyword", Query: "C3 version upgrade and cache recovery without trusting c3.db as submitted truth", Expected: []string{"c3-215"}},
	{ID: "rule-golden-antipattern-keyword", Kind: "keyword", Query: "enforceable coding rules with golden examples and anti-patterns", Expected: []string{"c3-217"}},
	{ID: "audit-pass-warn-fail-keyword", Kind: "keyword", Query: "structural validation and semantic drift review producing PASS WARN FAIL output", Expected: []string{"c3-213"}},
	{ID: "sweep-ripple-paraphrase", Kind: "paraphrase", Query: "trace how far a risky architecture edit will ripple through dependents before committing to it", Expected: []string{"c3-218"}},
	{ID: "change-decision-record-paraphrase", Kind: "paraphrase", Query: "require a written decision record as the work order before touching architecture-affecting code, then verify", Expected: []string{"c3-214"}},
	{ID: "onboard-bootstrap-paraphrase", Kind: "paraphrase", Query: "bring an existing codebase under C3 for the first time, bootstrap the docs, map out the structure, and leave it passing checks", Expected: []string{"c3-211"}},
	{ID: "opindex-shared-contract-paraphrase", Kind: "paraphrase", Query: "the common template every per-operation workflow doc conforms to, the stages, gates, and final checks each one must spell out", Expected: []string{"c3-210"}},

	// --- refs / rules ---
	{ID: "thin-npm-model-blob-keyword", Kind: "keyword", Query: "per-platform fat skill zip vs thin npm download semantic model SHA256", Expected: []string{"ref-cross-compiled-binary"}},
	{ID: "templates-in-binary-paraphrase", Kind: "paraphrase", Query: "new-doc skeletons ship compiled inside the executable so generating stubs needs no files on disk", Expected: []string{"ref-embedded-templates", "c3-103"}},
	{ID: "metadata-header-body-paraphrase", Kind: "paraphrase", Query: "each architecture doc splits a machine-parseable header of identity fields from the human prose underneath", Expected: []string{"ref-frontmatter-docs", "c3-101"}},
	{ID: "error-suggests-next-paraphrase", Kind: "paraphrase", Query: "a failed top-level command tells the user which recovery action to run next instead of a bare message", Expected: []string{"rule-dispatcher-error-hint"}},
	{ID: "single-output-writer-keyword", Kind: "keyword", Query: "WriteTableOutput ResolveFormat never call json.Marshal directly TOON agent mode", Expected: []string{"rule-output-via-helpers", "c3-108"}},
	{ID: "wrap-error-cause-keyword", Kind: "keyword", Query: "wrap returned error with %w preserve cause errors.Is unwrap chain not %v", Expected: []string{"rule-wrap-error-cause"}},
}

func main() {
	var dbPath string
	var limit int
	semantic := true
	var noSemantic bool
	flag.StringVar(&dbPath, "db", "", "path to .c3/c3.db")
	flag.IntVar(&limit, "k", 5, "ranking cutoff")
	flag.BoolVar(&semantic, "semantic", true, "enable local ONNX semantic fusion (default true)")
	flag.BoolVar(&noSemantic, "no-semantic", false, "disable semantic fusion for keyword/graph baseline")
	flag.Parse()
	if noSemantic {
		semantic = false
	}
	if semantic {
		seedEvalVersionEnv()
	}

	if limit <= 0 {
		fail("k must be positive")
	}
	if dbPath == "" {
		var err error
		dbPath, err = defaultDBPath()
		if err != nil {
			fail(err.Error())
		}
	}

	s, err := store.Open(dbPath)
	if err != nil {
		fail(fmt.Sprintf("open db: %v", err))
	}
	defer s.Close()

	report, err := runEval(s, dbPath, limit, semantic)
	if err != nil {
		fail(err.Error())
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fail(fmt.Sprintf("encode report: %v", err))
	}
}

func defaultDBPath() (string, error) {
	candidates := []string{
		filepath.Join("..", ".c3", "c3.db"),
		filepath.Join(".c3", "c3.db"),
		filepath.Join("..", "..", ".c3", "c3.db"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find .c3/c3.db; pass --db")
}

func seedEvalVersionEnv() {
	if strings.TrimSpace(os.Getenv("C3X_VERSION")) != "" {
		return
	}
	candidates := []string{
		filepath.Join("..", "skills", "c3", "bin", "VERSION"),
		filepath.Join("..", "..", "skills", "c3", "bin", "VERSION"),
		filepath.Join("skills", "c3", "bin", "VERSION"),
	}
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		if version := strings.TrimSpace(string(data)); version != "" {
			os.Setenv("C3X_VERSION", version)
			return
		}
	}
}

func runEval(s *store.Store, dbPath string, limit int, semantic bool) (evalReport, error) {
	// The CLI resolves C3X_MODE=agent to TOON. The eval needs JSON from RunSearch.
	os.Unsetenv("C3X_MODE")

	results := make([]caseResult, 0, len(cases))
	for _, tc := range cases {
		hits, err := searchIDs(s, tc.Query, limit, semantic)
		if err != nil {
			return evalReport{}, fmt.Errorf("%s: %w", tc.ID, err)
		}
		results = append(results, caseResult{
			evalCase: tc,
			Rank:     rankOf(hits, tc.Expected),
			Hits:     hits,
		})
	}

	byKindCases := make(map[string][]caseResult)
	for _, result := range results {
		byKindCases[result.Kind] = append(byKindCases[result.Kind], result)
	}
	byKind := make(map[string]metrics, len(byKindCases))
	for kind, kindCases := range byKindCases {
		byKind[kind] = summarize(kindCases)
	}

	mode := "keyword_graph"
	if semantic {
		mode = "keyword_graph_onnx_semantic"
	}
	return evalReport{
		Mode:    mode,
		DB:      dbPath,
		Limit:   limit,
		Metrics: summarize(results),
		ByKind:  byKind,
		Cases:   results,
	}, nil
}

func searchIDs(s *store.Store, query string, limit int, semantic bool) ([]string, error) {
	var buf bytes.Buffer
	if err := cmd.RunSearch(cmd.SearchOptions{
		Store:      s,
		Query:      query,
		Hybrid:     true,
		Semantic:   semantic,
		NoSemantic: !semantic,
		JSON:       true,
		Limit:      limit,
	}, &buf); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	var out cmd.SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("parse search JSON: %w\n%s", err, buf.String())
	}
	ids := make([]string, 0, len(out.Results))
	for _, row := range out.Results {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func rankOf(hits, expected []string) int {
	for i, hit := range hits {
		for _, want := range expected {
			if hit == want {
				return i + 1
			}
		}
	}
	return 0
}

func summarize(results []caseResult) metrics {
	if len(results) == 0 {
		return metrics{}
	}
	var hit1, hit3, hit5 int
	var rr float64
	for _, result := range results {
		switch {
		case result.Rank == 1:
			hit1++
			hit3++
			hit5++
		case result.Rank > 1 && result.Rank <= 3:
			hit3++
			hit5++
		case result.Rank > 3 && result.Rank <= 5:
			hit5++
		}
		if result.Rank > 0 {
			rr += 1 / float64(result.Rank)
		}
	}
	n := float64(len(results))
	return metrics{
		CaseCount: len(results),
		HitAt1:    float64(hit1) / n,
		HitAt3:    float64(hit3) / n,
		HitAt5:    float64(hit5) / n,
		MRR:       rr / n,
	}
}

func fail(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "unknown error"
	}
	fmt.Fprintf(os.Stderr, "error: %s\n", message)
	os.Exit(1)
}
