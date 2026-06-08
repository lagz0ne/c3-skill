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
	{
		ID:       "lookup-owner-paraphrase",
		Kind:     "paraphrase",
		Query:    "which architecture unit owns a source path",
		Expected: []string{"c3-114"},
	},
	{
		ID:       "binary-no-toolchain-paraphrase",
		Kind:     "paraphrase",
		Query:    "users can run c3x without building from source",
		Expected: []string{"ref-cross-compiled-binary"},
	},
	{
		ID:       "intent-router-paraphrase",
		Kind:     "paraphrase",
		Query:    "send the user's request to the right workflow",
		Expected: []string{"c3-201"},
	},
	{
		ID:       "agent-output-paraphrase",
		Kind:     "paraphrase",
		Query:    "machine responses use one formatter in agent mode",
		Expected: []string{"rule-output-via-helpers"},
	},
	{
		ID:       "frontmatter-keyword",
		Kind:     "keyword",
		Query:    "yaml frontmatter markdown body",
		Expected: []string{"ref-frontmatter-docs"},
	},
	{
		ID:       "templates-keyword",
		Kind:     "keyword",
		Query:    "embedded doc templates",
		Expected: []string{"ref-embedded-templates"},
	},
	{
		ID:       "check-validation-keyword",
		Kind:     "keyword",
		Query:    "validate structural integrity docs refs rules",
		Expected: []string{"c3-113"},
	},
	{
		ID:       "cache-lifecycle-keyword",
		Kind:     "keyword",
		Query:    "import export sync repair cache",
		Expected: []string{"c3-119"},
	},
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
