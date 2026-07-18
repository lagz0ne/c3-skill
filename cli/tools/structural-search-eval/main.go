package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	fixtureSchema    = "structural-retrieval-fixture.v1"
	candidateSchema  = "structural-retrieval-candidate.v1"
	reportSchema     = "structural-retrieval-report.v1"
	gateSchema       = "structural-retrieval-gate.v1"
	familyWrongLayer = "wrong_layer_structural_owner"
	familyRoute      = "behavioral_route_regression"

	wallFreeze                   = "AG-BENCHMARK-FREEZE"
	wallOneVariable              = "AG-ONE-VARIABLE"
	wallFalseStructure           = "AG-FALSE-STRUCTURE"
	wallBlastRouteRecall         = "AG-BLAST-REGRESSION-ROUTE-RECALL"
	wallBlastRouteMRR            = "AG-BLAST-REGRESSION-ROUTE-MRR"
	wallWrongLayerMRR            = "WRONG-LAYER-MRR-NO-REGRESSION"
	wallStructuralRecallDelta    = "STRUCTURAL-OWNER-RECALL-DELTA"
	wallStructuralOwnerPrecision = "CW-STRUCTURAL-OWNER-PRECISION"
	wallContextBytesPerCase      = "CW-CONTEXT-BYTES-PER-CASE"
)

type structuralFact struct {
	FactID   string `json:"fact_id"`
	Subject  string `json:"subject"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
	Polarity string `json:"polarity"`
	Anchor   string `json:"anchor"`
}

type structuralTriple struct {
	Subject  string `json:"subject"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

type document struct {
	DocumentID string           `json:"document_id"`
	SourceKind string           `json:"source_kind"`
	Text       string           `json:"text"`
	UTF8Bytes  int              `json:"utf8_bytes"`
	Facts      []structuralFact `json:"facts"`
}

type fixture struct {
	Schema                     string             `json:"$schema"`
	FixtureVersion             string             `json:"fixture_version"`
	CaseID                     string             `json:"case_id"`
	Family                     string             `json:"family"`
	PublicStatus               string             `json:"public_status"`
	Query                      string             `json:"query"`
	KValues                    []int              `json:"k_values"`
	RequiredOwnerFactIDs       []string           `json:"required_owner_fact_ids"`
	BehavioralRouteFactIDs     []string           `json:"behavioral_route_fact_ids"`
	AllowedStructuralFactIDs   []string           `json:"allowed_structural_fact_ids"`
	ForbiddenStructuralTriples []structuralTriple `json:"forbidden_structural_triples"`
	Documents                  []document         `json:"documents"`
}

type rankingConfig struct {
	Mode     string `json:"mode"`
	TieBreak string `json:"tie_break"`
}

type changedVariable struct {
	Key            string `json:"key"`
	BaselineValue  string `json:"baseline_value"`
	CandidateValue string `json:"candidate_value"`
}

type candidateManifest struct {
	Schema                 string             `json:"$schema"`
	CandidateID            string             `json:"candidate_id"`
	ParentCandidateID      string             `json:"parent_candidate_id"`
	ApproachFamily         string             `json:"approach_family"`
	BaselineManifestSHA256 string             `json:"baseline_manifest_sha256"`
	FixtureSHA256          string             `json:"fixture_sha256"`
	ScorerSHA256           string             `json:"scorer_sha256"`
	ImplementationCommit   string             `json:"implementation_commit"`
	InvariantConfigSHA256  string             `json:"invariant_config_sha256"`
	ChangedVariable        *changedVariable   `json:"changed_variable"`
	Ranking                rankingConfig      `json:"ranking"`
	AssertedFactIDs        []string           `json:"asserted_fact_ids"`
	AssertedTriples        []structuralTriple `json:"asserted_triples"`
	FrozenBaseline         metrics            `json:"frozen_baseline"`
	loadedSHA256           string
}

type caseMetrics struct {
	CaseID                 string   `json:"case_id"`
	Family                 string   `json:"family"`
	RankedDocumentIDs      []string `json:"ranked_document_ids"`
	RetrievedFactIDsAt5    []string `json:"retrieved_fact_ids_at_5"`
	OwnerRecallAt1         float64  `json:"owner_recall_at_1"`
	OwnerRecallAt3         float64  `json:"owner_recall_at_3"`
	OwnerRecallAt5         float64  `json:"owner_recall_at_5"`
	MRR                    float64  `json:"mrr"`
	StructuralPrecisionAt5 float64  `json:"structural_owner_precision_at_5"`
	FalseStructuralClaims  int      `json:"false_structural_claim_count"`
	ContextBytesAt5        int      `json:"context_bytes_at_5"`
}

type metrics struct {
	OwnerRecallAt1               float64        `json:"owner_recall_at_1"`
	OwnerRecallAt3               float64        `json:"owner_recall_at_3"`
	OwnerRecallAt5               float64        `json:"owner_recall_at_5"`
	MRR                          float64        `json:"mrr"`
	WrongLayerMRR                float64        `json:"wrong_layer_mrr"`
	StructuralOwnerPrecision     float64        `json:"structural_owner_precision"`
	FalseStructuralClaimCount    int            `json:"false_structural_claim_count"`
	BehavioralRouteRecallAt5     float64        `json:"behavioral_route_recall_at_5"`
	BehavioralRouteMRR           float64        `json:"behavioral_route_mrr"`
	ContextBytesAt5              int            `json:"context_bytes_at_5"`
	RetrievedContextBytesPerCase map[string]int `json:"retrieved_context_bytes_per_case"`
}

type evalReport struct {
	Schema         string            `json:"$schema"`
	Candidate      candidateManifest `json:"candidate"`
	FixtureSHA256  string            `json:"fixture_sha256"`
	ScorerSHA256   string            `json:"scorer_sha256"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	Metrics        metrics           `json:"metrics"`
	Cases          []caseMetrics     `json:"cases"`
}

type gateVerdict struct {
	Schema      string   `json:"$schema"`
	Verdict     string   `json:"verdict"`
	Keep        bool     `json:"keep"`
	FailedWalls []string `json:"failed_walls"`
	RecallDelta float64  `json:"microbench_structural_owner_recall_delta"`
	Notes       []string `json:"notes"`
}

type rankedDocument struct {
	Document document
	Score    int
	Index    int
}

func main() {
	os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr))
}

func execute(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: structural-search-eval <run|gate> [flags]")
		return 2
	}
	var err error
	switch args[0] {
	case "run":
		err = runCommand(args[1:], stdout)
	case "gate":
		err = gateCommand(args[1:], stdout)
	default:
		err = fmt.Errorf("unknown command %q", args[0])
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var fixturesPath, candidatePath, outPath string
	flags.StringVar(&fixturesPath, "fixtures", "", "frozen JSONL fixture path")
	flags.StringVar(&candidatePath, "candidate", "", "candidate manifest path")
	flags.StringVar(&outPath, "out", "", "result path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if fixturesPath == "" || candidatePath == "" || outPath == "" {
		return errors.New("run requires --fixtures, --candidate, and --out")
	}

	fixtures, err := loadFixtures(fixturesPath)
	if err != nil {
		return err
	}
	manifest, err := loadCandidate(candidatePath)
	if err != nil {
		return err
	}
	fixtureHash, err := hashFile(fixturesPath)
	if err != nil {
		return err
	}
	scorerPath, err := scorerSourcePath()
	if err != nil {
		return err
	}
	scorerHash, err := hashFile(scorerPath)
	if err != nil {
		return err
	}
	if fixtureHash != manifest.FixtureSHA256 {
		return fmt.Errorf("%s: fixture hash got %s, want %s", wallFreeze, fixtureHash, manifest.FixtureSHA256)
	}
	if scorerHash != manifest.ScorerSHA256 {
		return fmt.Errorf("%s: scorer hash got %s, want %s", wallFreeze, scorerHash, manifest.ScorerSHA256)
	}
	if got := invariantConfigHash(manifest); got != manifest.InvariantConfigSHA256 {
		return fmt.Errorf("%s: invariant config hash got %s, want %s", wallOneVariable, got, manifest.InvariantConfigSHA256)
	}

	report, err := runCandidate(fixtures, manifest)
	if err != nil {
		return err
	}
	report.FixtureSHA256 = fixtureHash
	report.ScorerSHA256 = scorerHash
	report.ManifestSHA256 = manifest.loadedSHA256
	if manifest.ParentCandidateID == "" {
		if err := compareFrozenBaseline(manifest.FrozenBaseline, report.Metrics); err != nil {
			return fmt.Errorf("frozen baseline mismatch: %w", err)
		}
	}
	if err := writeJSONFile(outPath, report); err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(map[string]any{"result": outPath, "manifest_sha256": report.ManifestSHA256})
}

func gateCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("gate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var baselinePath, candidatePath, freezePath string
	flags.StringVar(&baselinePath, "baseline", "", "baseline report path")
	flags.StringVar(&candidatePath, "candidate", "", "candidate report path")
	flags.StringVar(&freezePath, "freeze", "", "frozen baseline manifest path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if baselinePath == "" || candidatePath == "" || freezePath == "" {
		return errors.New("gate requires --baseline, --candidate, and --freeze")
	}
	baseline, err := loadReport(baselinePath)
	if err != nil {
		return err
	}
	candidate, err := loadReport(candidatePath)
	if err != nil {
		return err
	}
	freeze, err := loadCandidate(freezePath)
	if err != nil {
		return err
	}
	verdict := evaluateGate(baseline, candidate, freeze)
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(verdict); err != nil {
		return err
	}
	if !verdict.Keep {
		return fmt.Errorf("discard: %s", strings.Join(verdict.FailedWalls, ", "))
	}
	return nil
}

func loadFixtures(path string) ([]fixture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open fixtures: %w", err)
	}
	defer file.Close()

	var fixtures []fixture
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var item fixture
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, fmt.Errorf("fixture line %d: %w", line, err)
		}
		if err := validateFixture(item); err != nil {
			return nil, fmt.Errorf("fixture line %d: %w", line, err)
		}
		fixtures = append(fixtures, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(fixtures) != 12 {
		return nil, fmt.Errorf("fixture count %d, want 12", len(fixtures))
	}
	return fixtures, nil
}

func validateFixture(item fixture) error {
	if item.Schema != fixtureSchema || item.FixtureVersion != "v1" || item.PublicStatus != "synthetic_public" {
		return errors.New("fixture must use v1 synthetic-public schema")
	}
	if item.CaseID == "" || item.Query == "" {
		return errors.New("case_id and query are required")
	}
	if item.Family != familyWrongLayer && item.Family != familyRoute {
		return fmt.Errorf("unknown family %q", item.Family)
	}
	if fmt.Sprint(item.KValues) != "[1 3 5]" {
		return errors.New("k_values must be [1,3,5]")
	}
	if len(item.RequiredOwnerFactIDs) == 0 || len(item.Documents) < 5 {
		return errors.New("required facts and at least five documents are required")
	}
	allowed := stringSet(item.AllowedStructuralFactIDs)
	seenFacts := map[string]bool{}
	seenDocuments := map[string]bool{}
	for _, doc := range item.Documents {
		if doc.DocumentID == "" || seenDocuments[doc.DocumentID] {
			return fmt.Errorf("document id %q is empty or duplicated", doc.DocumentID)
		}
		seenDocuments[doc.DocumentID] = true
		if doc.UTF8Bytes != len([]byte(doc.Text)) {
			return fmt.Errorf("%s utf8_bytes = %d, actual %d", doc.DocumentID, doc.UTF8Bytes, len([]byte(doc.Text)))
		}
		for _, fact := range doc.Facts {
			if fact.FactID == "" || fact.Polarity != "positive" || fact.Anchor == "" {
				return fmt.Errorf("%s has incomplete or non-positive fact", doc.DocumentID)
			}
			if !allowed[fact.FactID] {
				return fmt.Errorf("%s fact %s is not in the exact allowlist", doc.DocumentID, fact.FactID)
			}
			if seenFacts[fact.FactID] {
				return fmt.Errorf("fact %s is duplicated", fact.FactID)
			}
			seenFacts[fact.FactID] = true
		}
	}
	for _, id := range item.AllowedStructuralFactIDs {
		if !seenFacts[id] {
			return fmt.Errorf("allowed fact %s has no anchored document", id)
		}
	}
	for _, id := range append(append([]string{}, item.RequiredOwnerFactIDs...), item.BehavioralRouteFactIDs...) {
		if !allowed[id] || !seenFacts[id] {
			return fmt.Errorf("required fact %s is not anchored and allowed", id)
		}
	}
	return nil
}

func loadCandidate(path string) (candidateManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return candidateManifest{}, fmt.Errorf("read candidate: %w", err)
	}
	var manifest candidateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return candidateManifest{}, fmt.Errorf("decode candidate: %w", err)
	}
	if manifest.Schema != candidateSchema || manifest.CandidateID == "" || manifest.ApproachFamily == "" {
		return candidateManifest{}, errors.New("candidate must use v1 schema with id and approach family")
	}
	if !isSHA256(manifest.FixtureSHA256) || !isSHA256(manifest.ScorerSHA256) || !isSHA256(manifest.InvariantConfigSHA256) {
		return candidateManifest{}, errors.New("candidate freeze hashes must be lowercase SHA-256")
	}
	if len(manifest.ImplementationCommit) != 40 {
		return candidateManifest{}, errors.New("implementation_commit must be 40 hex characters")
	}
	sum := sha256.Sum256(data)
	manifest.loadedSHA256 = hex.EncodeToString(sum[:])
	return manifest, nil
}

func loadReport(path string) (evalReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return evalReport{}, fmt.Errorf("read report: %w", err)
	}
	var report evalReport
	if err := json.Unmarshal(data, &report); err != nil {
		return evalReport{}, fmt.Errorf("decode report: %w", err)
	}
	if report.Schema != reportSchema {
		return evalReport{}, fmt.Errorf("report schema %q, want %q", report.Schema, reportSchema)
	}
	return report, nil
}

func runCandidate(fixtures []fixture, manifest candidateManifest) (evalReport, error) {
	if manifest.Ranking.TieBreak != "fixture_order" {
		return evalReport{}, fmt.Errorf("unsupported tie break %q", manifest.Ranking.TieBreak)
	}
	if manifest.Ranking.Mode != "lexical" && manifest.Ranking.Mode != "entity_dump" && manifest.Ranking.Mode != "structural_first" {
		return evalReport{}, fmt.Errorf("unsupported ranking mode %q", manifest.Ranking.Mode)
	}

	cases := make([]caseMetrics, 0, len(fixtures))
	for _, item := range fixtures {
		ranked := rankDocuments(item, manifest.Ranking.Mode)
		cases = append(cases, scoreCase(item, ranked, manifest))
	}
	return evalReport{
		Schema:         reportSchema,
		Candidate:      manifest,
		ManifestSHA256: manifest.loadedSHA256,
		Metrics:        summarize(cases),
		Cases:          cases,
	}, nil
}

func rankDocuments(item fixture, mode string) []rankedDocument {
	ranked := make([]rankedDocument, 0, len(item.Documents))
	queryTokens := tokenize(item.Query)
	for index, doc := range item.Documents {
		score := lexicalOverlap(queryTokens, tokenize(doc.Text))
		if mode == "entity_dump" || mode == "structural_first" {
			if len(doc.Facts) > 0 {
				score += 1000
			}
		}
		ranked = append(ranked, rankedDocument{Document: doc, Score: score, Index: index})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return ranked[i].Index < ranked[j].Index
	})
	return ranked
}

func scoreCase(item fixture, ranked []rankedDocument, manifest candidateManifest) caseMetrics {
	required := item.RequiredOwnerFactIDs
	if item.Family == familyRoute && len(item.BehavioralRouteFactIDs) > 0 {
		required = item.BehavioralRouteFactIDs
	}
	recallAt := func(k int) float64 {
		found := factsInTopK(ranked, k)
		hits := 0
		for _, id := range required {
			if found[id] {
				hits++
			}
		}
		return float64(hits) / float64(len(required))
	}
	firstRank := 0
	for index, rankedDoc := range ranked {
		facts := factIDs(rankedDoc.Document.Facts)
		if intersects(facts, stringSet(required)) {
			firstRank = index + 1
			break
		}
	}
	mrr := 0.0
	if firstRank > 0 {
		mrr = 1 / float64(firstRank)
	}

	retrieved := factsInTopK(ranked, 5)
	requiredSet := stringSet(required)
	requiredHits := 0
	for id := range retrieved {
		if requiredSet[id] {
			requiredHits++
		}
	}
	precision := 0.0
	if len(retrieved) > 0 {
		precision = float64(requiredHits) / float64(len(retrieved))
	}

	allowed := stringSet(item.AllowedStructuralFactIDs)
	allowedTriples := map[structuralTriple]bool{}
	for _, doc := range item.Documents {
		for _, fact := range doc.Facts {
			allowedTriples[structuralTriple{Subject: fact.Subject, Relation: fact.Relation, Object: fact.Object}] = true
		}
	}
	falseClaims := 0
	for _, id := range manifest.AssertedFactIDs {
		if !allowed[id] {
			falseClaims++
		}
	}
	for _, asserted := range manifest.AssertedTriples {
		isFalse := !allowedTriples[asserted]
		for _, forbidden := range item.ForbiddenStructuralTriples {
			if asserted == forbidden {
				isFalse = true
			}
		}
		if isFalse {
			falseClaims++
		}
	}

	contextBytes := 0
	documentIDs := make([]string, 0, len(ranked))
	countedContext := map[string]bool{}
	for index, rankedDoc := range ranked {
		documentIDs = append(documentIDs, rankedDoc.Document.DocumentID)
		if index < 5 && !countedContext[rankedDoc.Document.DocumentID] {
			contextBytes += len([]byte(rankedDoc.Document.Text))
			countedContext[rankedDoc.Document.DocumentID] = true
		}
	}
	return caseMetrics{
		CaseID:                 item.CaseID,
		Family:                 item.Family,
		RankedDocumentIDs:      documentIDs,
		RetrievedFactIDsAt5:    sortedKeys(retrieved),
		OwnerRecallAt1:         recallAt(1),
		OwnerRecallAt3:         recallAt(3),
		OwnerRecallAt5:         recallAt(5),
		MRR:                    mrr,
		StructuralPrecisionAt5: precision,
		FalseStructuralClaims:  falseClaims,
		ContextBytesAt5:        contextBytes,
	}
}

func summarize(cases []caseMetrics) metrics {
	result := metrics{RetrievedContextBytesPerCase: map[string]int{}}
	wrongLayerCount := 0
	routeCount := 0
	for _, item := range cases {
		result.MRR += item.MRR
		result.FalseStructuralClaimCount += item.FalseStructuralClaims
		result.ContextBytesAt5 += item.ContextBytesAt5
		result.RetrievedContextBytesPerCase[item.CaseID] = item.ContextBytesAt5
		if item.Family == familyWrongLayer {
			wrongLayerCount++
			result.OwnerRecallAt1 += item.OwnerRecallAt1
			result.OwnerRecallAt3 += item.OwnerRecallAt3
			result.OwnerRecallAt5 += item.OwnerRecallAt5
			result.WrongLayerMRR += item.MRR
			result.StructuralOwnerPrecision += item.StructuralPrecisionAt5
		} else if item.Family == familyRoute {
			routeCount++
			result.BehavioralRouteRecallAt5 += item.OwnerRecallAt5
			result.BehavioralRouteMRR += item.MRR
		}
	}
	if len(cases) > 0 {
		result.MRR /= float64(len(cases))
	}
	if wrongLayerCount > 0 {
		count := float64(wrongLayerCount)
		result.OwnerRecallAt1 /= count
		result.OwnerRecallAt3 /= count
		result.OwnerRecallAt5 /= count
		result.WrongLayerMRR /= count
		result.StructuralOwnerPrecision /= count
	}
	if routeCount > 0 {
		count := float64(routeCount)
		result.BehavioralRouteRecallAt5 /= count
		result.BehavioralRouteMRR /= count
	}
	return result
}

func evaluateGate(baseline, candidate evalReport, freeze candidateManifest) gateVerdict {
	failed := []string{}
	notes := []string{}
	addFailure := func(wall, note string) {
		if !contains(failed, wall) {
			failed = append(failed, wall)
			notes = append(notes, note)
		}
	}

	if baseline.FixtureSHA256 != "" && (baseline.FixtureSHA256 != freeze.FixtureSHA256 || candidate.FixtureSHA256 != freeze.FixtureSHA256) {
		addFailure(wallFreeze, "fixture hash differs from the frozen baseline")
	}
	if baseline.ScorerSHA256 != "" && (baseline.ScorerSHA256 != freeze.ScorerSHA256 || candidate.ScorerSHA256 != freeze.ScorerSHA256) {
		addFailure(wallFreeze, "scorer hash differs from the frozen baseline")
	}
	if baseline.ManifestSHA256 != "" && candidate.Candidate.BaselineManifestSHA256 != baseline.ManifestSHA256 {
		addFailure(wallFreeze, "candidate does not cite the frozen baseline manifest hash")
	}
	if err := validateOneVariable(baseline.Candidate, candidate.Candidate); err != nil {
		addFailure(wallOneVariable, err.Error())
	}

	delta := candidate.Metrics.OwnerRecallAt5 - baseline.Metrics.OwnerRecallAt5
	if delta < 0.20-1e-12 {
		addFailure(wallStructuralRecallDelta, fmt.Sprintf("recall delta %.6f is below 0.20", delta))
	}
	if candidate.Metrics.StructuralOwnerPrecision < 0.80-1e-12 {
		addFailure(wallStructuralOwnerPrecision, fmt.Sprintf("precision %.6f is below 0.80", candidate.Metrics.StructuralOwnerPrecision))
	}
	if candidate.Metrics.FalseStructuralClaimCount != 0 || candidate.Metrics.FalseStructuralClaimCount > baseline.Metrics.FalseStructuralClaimCount {
		addFailure(wallFalseStructure, "candidate emitted a false structural claim")
	}
	if candidate.Metrics.BehavioralRouteRecallAt5+1e-12 < baseline.Metrics.BehavioralRouteRecallAt5 {
		addFailure(wallBlastRouteRecall, "behavioral route recall@5 regressed")
	}
	if candidate.Metrics.BehavioralRouteMRR+1e-12 < baseline.Metrics.BehavioralRouteMRR {
		addFailure(wallBlastRouteMRR, "behavioral route MRR regressed")
	}
	if candidate.Metrics.WrongLayerMRR+1e-12 < baseline.Metrics.WrongLayerMRR {
		addFailure(wallWrongLayerMRR, "wrong-layer MRR regressed")
	}
	for caseID, baselineBytes := range baseline.Metrics.RetrievedContextBytesPerCase {
		candidateBytes, ok := candidate.Metrics.RetrievedContextBytesPerCase[caseID]
		if !ok || float64(candidateBytes) > float64(baselineBytes)*1.05+1e-12 {
			addFailure(wallContextBytesPerCase, fmt.Sprintf("%s context bytes exceed the per-case limit", caseID))
		}
	}
	sort.Strings(failed)
	keep := len(failed) == 0
	verdict := "discard"
	if keep {
		verdict = "keep"
	}
	return gateVerdict{Schema: gateSchema, Verdict: verdict, Keep: keep, FailedWalls: failed, RecallDelta: delta, Notes: notes}
}

func validateOneVariable(baseline, candidate candidateManifest) error {
	change := candidate.ChangedVariable
	if change == nil || change.Key != "ranking.mode" {
		return errors.New("candidate must register exactly ranking.mode")
	}
	if change.BaselineValue != baseline.Ranking.Mode || change.CandidateValue != candidate.Ranking.Mode || change.BaselineValue == change.CandidateValue {
		return errors.New("changed_variable values do not match the baseline and candidate")
	}
	if candidate.ParentCandidateID != baseline.CandidateID || candidate.ApproachFamily != baseline.ApproachFamily {
		return errors.New("candidate parent or approach family changed")
	}
	if candidate.Ranking.TieBreak != baseline.Ranking.TieBreak || !equalStrings(candidate.AssertedFactIDs, baseline.AssertedFactIDs) || !equalTriples(candidate.AssertedTriples, baseline.AssertedTriples) {
		return errors.New("configuration changed outside ranking.mode")
	}
	if candidate.InvariantConfigSHA256 != baseline.InvariantConfigSHA256 || invariantConfigHash(candidate) != baseline.InvariantConfigSHA256 {
		return errors.New("invariant configuration hash changed")
	}
	return nil
}

func compareFrozenBaseline(want, got metrics) error {
	wantJSON, _ := json.Marshal(want)
	gotJSON, _ := json.Marshal(got)
	if !bytes.Equal(wantJSON, gotJSON) {
		return fmt.Errorf("got %s, want %s", gotJSON, wantJSON)
	}
	return nil
}

func invariantConfigHash(manifest candidateManifest) string {
	value := struct {
		ApproachFamily  string             `json:"approach_family"`
		TieBreak        string             `json:"tie_break"`
		AssertedFactIDs []string           `json:"asserted_fact_ids"`
		AssertedTriples []structuralTriple `json:"asserted_triples"`
	}{manifest.ApproachFamily, manifest.Ranking.TieBreak, manifest.AssertedFactIDs, manifest.AssertedTriples}
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func scorerSourcePath() (string, error) {
	candidates := []string{
		filepath.Join("tools", "structural-search-eval", "main.go"),
		filepath.Join("cli", "tools", "structural-search-eval", "main.go"),
	}
	if cwd, err := os.Getwd(); err == nil && filepath.Base(cwd) == "structural-search-eval" {
		candidates = append(candidates, "main.go")
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", errors.New("cannot locate structural-search-eval/main.go for scorer freeze")
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func isSHA256(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func tokenize(value string) map[string]bool {
	parts := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	return stringSet(parts)
}

func lexicalOverlap(left, right map[string]bool) int {
	count := 0
	for token := range left {
		if right[token] {
			count++
		}
	}
	return count
}

func factsInTopK(ranked []rankedDocument, k int) map[string]bool {
	result := map[string]bool{}
	seenDocuments := map[string]bool{}
	if k > len(ranked) {
		k = len(ranked)
	}
	for _, item := range ranked[:k] {
		if seenDocuments[item.Document.DocumentID] {
			continue
		}
		seenDocuments[item.Document.DocumentID] = true
		for _, fact := range item.Document.Facts {
			result[fact.FactID] = true
		}
	}
	return result
}

func factIDs(facts []structuralFact) []string {
	result := make([]string, 0, len(facts))
	for _, fact := range facts {
		result = append(result, fact.FactID)
	}
	return result
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func intersects(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

func sortedKeys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func equalStrings(left, right []string) bool {
	return fmt.Sprint(left) == fmt.Sprint(right)
}

func equalTriples(left, right []structuralTriple) bool {
	return fmt.Sprint(left) == fmt.Sprint(right)
}
