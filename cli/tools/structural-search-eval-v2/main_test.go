package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func skipUnlessReleaseFixtures(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			t.Skipf("optional release fixture unavailable: %s", path)
		}
	}
}

func skipUnlessBubblewrap(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bubblewrap unavailable")
	}
}

func TestStrictRuntimeOutputAcceptsOneObjectAndEOFOnly(t *testing.T) {
	want := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: "one"}}}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeArmResponse(bytes.NewReader(append(data, '\n')), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("response mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestStrictRuntimeOutputRejectsUnknownFieldPrefixSuffixSecondObjectAndLog(t *testing.T) {
	valid := `{"$schema":"structural-retrieval-arm-response.v2","cases":[]}`
	for name, input := range map[string]string{
		"unknown":       `{"$schema":"structural-retrieval-arm-response.v2","cases":[],"score":1}`,
		"prefix":        "log\n" + valid,
		"suffix":        valid + "log",
		"second_object": valid + valid,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeArmResponse(strings.NewReader(input), 1<<20); err == nil {
				t.Fatal("malformed arm output was accepted")
			}
		})
	}
}

func TestRuntimeUsesRealRunSearchAndControllerRedactsOracle(t *testing.T) {
	fixture := sampleFixture()
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{fixture.Oracle.RequiredOwnerFactIDs[0], fixture.Oracle.ForbiddenFactIDs[0]} {
		if bytes.Contains(encoded, []byte(secret)) {
			t.Fatalf("redacted arm input contains oracle value %q", secret)
		}
	}
	root := t.TempDir()
	got, err := runArm(req, filepath.Join(root, "db", "c3.db"), filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Database.Integrity != "ok" || len(got.Cases) != 1 {
		t.Fatalf("unexpected arm response: %#v", got)
	}
	if len(got.Cases[0].Rows) == 0 {
		t.Fatal("real RunSearch returned no rows for the seeded corpus")
	}
	if got.Cases[0].Rows[0].ID != "c3-100" {
		t.Fatalf("top real RunSearch row = %q, want c3-100", got.Cases[0].Rows[0].ID)
	}
}

func TestRuntimeSeedsStoreContentAndRelationships(t *testing.T) {
	fixture := sampleRouteFixture()
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	got, err := runArm(req, filepath.Join(root, "db", "c3.db"), filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Database.EntityCount != len(fixture.Corpus.Entities) || got.Database.RelationshipCount != len(fixture.Corpus.Relationships) {
		t.Fatalf("database counts = entities %d relationships %d", got.Database.EntityCount, got.Database.RelationshipCount)
	}
	want := "graph:uses:c3-201"
	found := false
	for _, row := range got.Cases[0].Rows {
		if row.ID == "c3-200" && containsString(row.MatchSources, want) {
			found = true
		}
	}
	if !found {
		t.Fatalf("real relationship expansion did not return c3-200 with %q: %#v", want, got.Cases[0].Rows)
	}
}

func TestRuntimeTopologicallySeedsParentsBeforeChildren(t *testing.T) {
	fixture := sampleFixture()
	fixture.Corpus.Entities[0].ParentID = "c3-101"
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	got, err := runArm(req, filepath.Join(root, "db", "c3.db"), filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Database.EntityCount != len(fixture.Corpus.Entities) {
		t.Fatalf("topological seed count=%d", got.Database.EntityCount)
	}
}

func TestPreCandidateSixArmLogicalAndOutputGolden(t *testing.T) {
	root := repoRoot(t)
	fixturesPath := filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json")
	fixtures, _, err := loadFixtures(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil {
		t.Fatal(err)
	}
	type golden struct {
		name, requestMode, resultSHA256, logicalSHA256 string
		fixtures                                       []fixtureCase
		entities, relationships, sqliteRows, bytes     int
	}
	plans := make([]golden, 0, len(fixtures)+2)
	resultHashes := []string{
		"9388bd3a161a20af90471756305454bb41282954671120035ba025b5135c788c",
		"b6b2e502a9be67034ab668390840fffdc69ac753400eb549b397eb200c4c0716",
		"a9dc856ec207216b133404ba747ff60d7e9f761c670c64e63bc37b1173f6922a",
		"cf68dc9db226672c3e3d248646b525afa086abf554efd83a65547ea8433c116e",
	}
	logicalHashes := []string{
		"09264f191cea4f03a828329a59d72ff25bf56b91952fa512ac46d5bfc050e185",
		"0471b7457d5a6e889a0891c57347fb8425b4b869a8f7721ce9ccd37e7b672d54",
		"f3c9344a2099aba91938ca01bbb80ddcea20afe87b18d2602bd171c0455821a6",
		"b2915fde1609989cc554058f4da9237e538e2705c3bcf73dbe46d46bf84e4757",
	}
	entityCounts := []int{3, 3, 2, 3}
	relationshipCounts := []int{0, 0, 1, 1}
	sqliteRows := []int{9, 9, 7, 10}
	logicalBytes := []int{7514, 7553, 6918, 7570}
	for i, fixture := range fixtures {
		plans = append(plans, golden{
			name: fixture.CaseID, requestMode: corpusIsolated, fixtures: []fixtureCase{fixture},
			resultSHA256: resultHashes[i], logicalSHA256: logicalHashes[i], entities: entityCounts[i],
			relationships: relationshipCounts[i], sqliteRows: sqliteRows[i], bytes: logicalBytes[i],
		})
	}
	plans = append(plans, golden{
		name: "combined", requestMode: corpusCombined, fixtures: fixtures,
		resultSHA256:  "41fdad29ffb0ed7f482c3649dc0448c3b9aaebe5a5aa9d169e15d91d53e1c69e",
		logicalSHA256: "09dbf7790e86a98ba42bcd74b1924a27be9043d18e4f344bbc9671222f09dba1",
		entities:      11, relationships: 2, sqliteRows: 35, bytes: 12798,
	})
	scaled := append([]fixtureCase(nil), fixtures...)
	for i := range scaled {
		cfg := bench.Scale
		cfg.Seed += i * 1000
		scaled[i].Corpus = generateScaleCorpus(scaled[i].Corpus, cfg)
	}
	plans = append(plans, golden{
		name: "scale", requestMode: corpusCombined, fixtures: scaled,
		resultSHA256:  "41fdad29ffb0ed7f482c3649dc0448c3b9aaebe5a5aa9d169e15d91d53e1c69e",
		logicalSHA256: "057e448b731c0d72295c46057a714afe1f688d89b12dbe677afb1020ec8758ba",
		entities:      22, relationships: 2, sqliteRows: 68, bytes: 19936,
	})
	for _, plan := range plans {
		plan := plan
		t.Run(plan.name, func(t *testing.T) {
			req, err := buildArmRequest(plan.fixtures, plan.requestMode, bench.SemanticMode, bench)
			if err != nil {
				t.Fatal(err)
			}
			armRoot := t.TempDir()
			dbPath := filepath.Join(armRoot, "db", "c3.db")
			projectDir := filepath.Join(armRoot, "project")
			c3Dir := filepath.Join(armRoot, "c3")
			response, err := executeArm(req, dbPath, projectDir, c3Dir)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			raw = append(raw, '\n')
			if got := sha(raw); got != plan.resultSHA256 {
				t.Fatalf("raw result SHA-256=%s want %s", got, plan.resultSHA256)
			}
			result, err := inspectArmResult(req, response, dbPath)
			if err != nil {
				t.Fatal(err)
			}
			dump := result.Database
			if dump.LogicalSHA256 != plan.logicalSHA256 || dump.EntityCount != plan.entities || dump.RelationshipCount != plan.relationships || dump.SQLiteRowCount != plan.sqliteRows || dump.LogicalBytes != plan.bytes {
				t.Fatalf("logical golden changed: dump=%#v want_sha=%s entities=%d relationships=%d rows=%d bytes=%d", dump, plan.logicalSHA256, plan.entities, plan.relationships, plan.sqliteRows, plan.bytes)
			}
			for name, path := range map[string]string{"project": projectDir, "c3": c3Dir} {
				got, err := directoryTreeSHA256(path)
				if err != nil {
					t.Fatal(err)
				}
				if got != "eeb78c091bad369214da171f0ec0a62f644b9664c8f4fed6f09c2ed29fb79632" {
					t.Fatalf("%s tree SHA-256=%s", name, got)
				}
			}
		})
	}
}

func TestExecuteArmRollsBackAllSeedWritesWhenRelationshipFails(t *testing.T) {
	req := armRequest{
		Schema: armRequestSchema, SemanticMode: semanticDisabled,
		Corpus: corpusInput{
			Entities: []entityInput{
				{ID: "rollback-owner", Type: "component", Title: "Rollback Owner", Slug: "rollback-owner", Status: "active", Metadata: "{}", Markdown: "# Rollback Owner\n\nOwned content."},
				{ID: "rollback-dependency", Type: "component", Title: "Rollback Dependency", Slug: "rollback-dependency", Status: "active", Metadata: "{}", Markdown: "# Rollback Dependency\n\nDependency content."},
			},
			Relationships: []relationshipInput{
				{FromID: "rollback-owner", ToID: "rollback-dependency", RelType: "uses"},
				{FromID: "rollback-owner", ToID: "missing-target", RelType: "uses"},
			},
		},
	}
	root := t.TempDir()
	dbPath := filepath.Join(root, "db", "c3.db")
	_, err := executeArm(req, dbPath, filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err == nil || !strings.Contains(err.Error(), "add relationship") || !strings.Contains(strings.ToLower(err.Error()), "foreign key") {
		t.Fatalf("late invalid relationship did not cause the expected failure: %v", err)
	}
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, table := range []string{"entities", "nodes", "versions", "relationships", "entities_fts", "content_fts"} {
		var count int
		if err := s.DB().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("%s retained %d rows after failed seed transaction", table, count)
		}
	}
	var integrity string
	if err := s.DB().QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("integrity after rollback=%q err=%v", integrity, err)
	}
}

func TestRelationshipRouteRequiresExpectedTop5AndFrozenLogicalRelationshipWitness(t *testing.T) {
	fixture := sampleRouteFixture()
	result := armCaseResult{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-200", MatchSources: []string{"graph:uses:c3-201"}}}}
	dump := logicalDump{Relationships: []logicalRelationship{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}}}
	if !relationshipRouteHit(fixture, result, dump) {
		t.Fatal("valid top-5 row plus frozen witness was rejected")
	}
	dump.Relationships = nil
	if relationshipRouteHit(fixture, result, dump) {
		t.Fatal("MatchSources alone was accepted without the controller-owned witness")
	}
	result.Rows = []cmd.SearchResultRow{{ID: "other", MatchSources: []string{"graph:uses:c3-201"}}}
	dump.Relationships = []logicalRelationship{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}}
	if relationshipRouteHit(fixture, result, dump) {
		t.Fatal("witness alone was accepted without the expected top-5 entity")
	}
}

func TestMatchSourcesIsCorroborationNotProducerProof(t *testing.T) {
	fixture := sampleRouteFixture()
	result := armCaseResult{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-200", MatchSources: []string{"graph:uses:c3-201"}}}}
	if relationshipRouteHit(fixture, result, logicalDump{}) {
		t.Fatal("MatchSources was treated as producer proof")
	}
}

func TestExpansionSpecificOwnerHasNoDirectLexicalMatchAndDirectFTSProbesMiss(t *testing.T) {
	fixture := sampleRouteFixture()
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	got, err := runArm(req, filepath.Join(root, "db", "c3.db"), filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err != nil {
		t.Fatal(err)
	}
	probes := got.DirectProbes[fixture.CaseID]
	if containsString(probes.EntityFTSIDs, fixture.Oracle.RelationshipWitness.FromID) || containsString(probes.ContentFTSIDs, fixture.Oracle.RelationshipWitness.FromID) {
		t.Fatalf("expansion-specific owner was directly retrievable: %#v", probes)
	}
	if err := validateExpansionSpecificFixture(fixture); err != nil {
		t.Fatal(err)
	}
}

func TestControllerRecomputesMetricsFromRowsAndRejectsSuppliedSummary(t *testing.T) {
	fixture := sampleFixture()
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-100"}}}}}
	dump := logicalDump{Integrity: "ok", Entities: []logicalEntity{{ID: "c3-100"}}}
	report, err := scoreArmResponseWithProbes([]fixtureCase{fixture}, response, dump, corpusCombined, map[string]directProbes{fixture.CaseID: {}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.OwnerRecallAt5 != 1 || report.Metrics.StructuralOwnerPrecision != 1 {
		t.Fatalf("recomputed metrics = %#v", report.Metrics)
	}
	forged := report
	forged.Metrics.OwnerRecallAt5 = 0
	if err := verifyReport([]fixtureCase{fixture}, response, dump, corpusCombined, forged); err == nil {
		t.Fatal("forged report summary was accepted")
	}
}

func TestForbiddenStructuralRetrievalProxyDoesNotClaimAnswerCorrectness(t *testing.T) {
	fixture := sampleFixture()
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-101"}}}}}
	report, err := scoreArmResponse([]fixtureCase{fixture}, response, logicalDump{Integrity: "ok"}, corpusCombined)
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.ForbiddenStructuralRetrievalCount != 1 {
		t.Fatalf("forbidden retrieval count = %d, want 1", report.Metrics.ForbiddenStructuralRetrievalCount)
	}
	data, _ := json.Marshal(report)
	if bytes.Contains(data, []byte("blocking_false_structural_claim")) {
		t.Fatal("retrieval proxy was labeled as answer correctness")
	}
}

func TestCanonicalRowBytesIncludeSparseRowsAndParentSnippet(t *testing.T) {
	base := []cmd.SearchResultRow{{ID: "x", Title: "X", Snippet: "one"}}
	withParent := []cmd.SearchResultRow{{ID: "x", Title: "X", Snippet: "one parent:c3-9"}}
	if canonicalRowBytes(base) == 0 {
		t.Fatal("sparse result was padded or treated as zero")
	}
	if canonicalRowBytes(withParent) <= canonicalRowBytes(base) {
		t.Fatal("parent snippet bytes were not counted")
	}
}

func TestContextThresholdBlocksAdmissionButNotAdapterImplementation(t *testing.T) {
	skipUnlessReleaseFixtures(t, filepath.Join(repoRoot(t), ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"))
	bench := sampleBenchmark()
	bench.ContextThresholdAuthority = nil
	if err := validateAdapterImplementationConfig(bench); err != nil {
		t.Fatalf("adapter implementation was blocked by threshold: %v", err)
	}
	if err := validateBaselineAdmission(bench, provenance{}); err == nil || !strings.Contains(err.Error(), "re-ratified") {
		t.Fatalf("baseline admission did not fail on threshold, got %v", err)
	}
	bench.ContextThresholdAuthority = &thresholdAuthority{CheckinRef: "checkins.jsonl#seq=5", CheckinSHA256: "2e52b66573dc67b8950420c2e6e232a4899c1aaac3f915608f634bffdcd704e7", RecordHash: "66a927cbe7a5b78bf8641c662e952e3ed250ec7cdacae81749f8c3fde362ecac", DefinitionSHA256: "cfb85b044b6b4a3d5d694a9fe009ed4d597bc3006c44b02f59ed48a5e12e8429"}
	p := validProvenance()
	if err := validateBaselineAdmission(bench, p); err == nil || !strings.Contains(err.Error(), "external authority record") {
		t.Fatalf("missing external threshold authority was accepted, got %v", err)
	}
	line := checkinLine(t, filepath.Join(repoRoot(t), ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"), 5)
	p.ContextThresholdAuthorityRecord = append([]byte(nil), line...)
	p.ContextThresholdAuthorityRecord[len(p.ContextThresholdAuthorityRecord)-2] ^= 1
	if err := validateBaselineAdmission(bench, p); err == nil || !strings.Contains(err.Error(), "authority") {
		t.Fatalf("forged threshold authority line was accepted, got %v", err)
	}
	p.ContextThresholdAuthorityRecord = line
	if err := validateBaselineAdmission(bench, p); err != nil {
		t.Fatalf("exact external threshold authority was rejected: %v", err)
	}
}

func TestCombinedCorpusExposesCrossCaseInterference(t *testing.T) {
	first := sampleFixture()
	second := sampleFixture()
	second.CaseID = "second"
	second.Query = "shared query"
	second.Corpus.Entities[0].ID = "c3-300"
	second.Corpus.Entities[0].Title = "Shared Query Decoy"
	second.Corpus.Entities[0].Slug = "shared-query-decoy"
	second.Oracle.FactBindings = map[string][]string{"c3-300": {"fact-second"}}
	second.Oracle.RequiredOwnerFactIDs = []string{"fact-second"}
	second.Oracle.ForbiddenFactIDs = []string{"forbidden-second"}
	isolated, err := buildArmRequest([]fixtureCase{first}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	combined, err := buildArmRequest([]fixtureCase{first, second}, corpusCombined, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	if len(combined.Corpus.Entities) <= len(isolated.Corpus.Entities) {
		t.Fatal("combined corpus did not include cross-case entities")
	}
}

func TestScaleCorpusUsesFrozenOracleBlindGenerator(t *testing.T) {
	fixture := sampleFixture()
	bench := sampleBenchmark()
	a := generateScaleCorpus(fixture.Corpus, bench.Scale)
	fixture.Oracle.RequiredOwnerFactIDs = []string{"changed-secret"}
	b := generateScaleCorpus(fixture.Corpus, bench.Scale)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("scale generator changed after oracle-only mutation")
	}
	if len(a.Entities) <= len(fixture.Corpus.Entities) {
		t.Fatal("scale generator added no decoys")
	}
}

func TestCanonicalLogicalDumpStableWhenRawSQLiteBytesDiffer(t *testing.T) {
	fixture := sampleFixture()
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	var hashes []string
	var raw [][]byte
	for i := 0; i < 2; i++ {
		root := t.TempDir()
		dbPath := filepath.Join(root, "db", "c3.db")
		got, err := runArm(req, dbPath, filepath.Join(root, "project"), filepath.Join(root, "c3"))
		if err != nil {
			t.Fatal(err)
		}
		if err := setRawDiagnosticMarker(dbPath, byte(i+1)); err != nil {
			t.Fatal(err)
		}
		hashes = append(hashes, got.Database.LogicalSHA256)
		data, err := os.ReadFile(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		raw = append(raw, data)
	}
	if hashes[0] != hashes[1] {
		t.Fatalf("logical hashes differ: %v", hashes)
	}
	if len(raw[0]) == 0 || len(raw[1]) == 0 {
		t.Fatal("raw databases were not retained for diagnosis")
	}
	if bytes.Equal(raw[0], raw[1]) {
		t.Fatal("test did not deliberately produce different raw SQLite bytes")
	}
}

func TestStrictANDParentDecoySuppressesORFallbackAndRouteMatchSourceMustSurvive(t *testing.T) {
	fixture := sampleRouteFixture()
	fixture.CaseID = "strict-and"
	fixture.Query = "approval notification"
	fixture.Corpus.Entities[1].Markdown = "# Notification Relay\n\nApproval notification relay events."
	fixture.Corpus.Entities = append(fixture.Corpus.Entities, entityInput{ID: "c3-202", Type: "component", Title: "Approval Notification", Slug: "approval-notification", Goal: "approval notification", Status: "active", Metadata: `{}`, Markdown: "# Approval Notification\n\nApproval notification."})
	if err := validateRouteFixture(fixture); err != nil {
		t.Fatal(err)
	}
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	got, err := runArm(req, filepath.Join(root, "db", "c3.db"), filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cases) != 1 || !relationshipRouteHit(fixture, got.Cases[0], got.Database) {
		t.Fatal("route source did not survive strict-AND decoy scoring")
	}
}

func TestRelationshipRouteAndRouteEnrichmentAreScoredSeparately(t *testing.T) {
	fixture := sampleRouteFixture()
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-200", MatchSources: []string{"graph:uses:c3-201"}}}}}}
	dump := logicalDump{Integrity: "ok", Relationships: []logicalRelationship{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}}}
	report, err := scoreArmResponseWithProbes([]fixtureCase{fixture}, response, dump, corpusCombined, map[string]directProbes{fixture.CaseID: {}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.RelationshipRouteRecallAt5 != 1 {
		t.Fatal("relationship route was not scored")
	}
	if report.Metrics.RouteCoverage["facts"] != 0 {
		t.Fatal("MatchSources was counted as Route field coverage")
	}
}

func TestRelationshipExpansionRequiresControllerObservedDirectFTSMiss(t *testing.T) {
	fixture := sampleRouteFixture()
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-200", MatchSources: []string{"graph:uses:c3-201"}}}}}}
	dump := logicalDump{Relationships: []logicalRelationship{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}}}
	probes := map[string]directProbes{fixture.CaseID: {EntityFTSIDs: []string{"c3-200"}}}
	report, err := scoreArmResponseWithProbes([]fixtureCase{fixture}, response, dump, corpusCombined, probes)
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.RelationshipRouteRecallAt5 != 0 || report.Cases[0].ExpansionSpecificDirectMiss {
		t.Fatalf("route credit survived a direct FTS hit: %#v", report.Cases[0])
	}
}

func TestScorerUsesOnlyFrozenRequiredRouteFields(t *testing.T) {
	fixture := sampleRouteFixture()
	fixture.Oracle.RequiredRouteFields = []string{"graph", "hash"}
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-200", MatchSources: []string{"graph:uses:c3-201"}, Route: cmd.RouteEnrichment{Graph: []string{"uses:c3-201"}, Hash: strings.Repeat("a", 64)}}}}}}
	dump := logicalDump{Relationships: []logicalRelationship{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}}}
	report, err := scoreArmResponseWithProbes([]fixtureCase{fixture}, response, dump, corpusCombined, map[string]directProbes{fixture.CaseID: {}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.Cases[0].RouteCoverage, map[string]float64{"graph": 1, "hash": 1}) {
		t.Fatalf("scorer ignored frozen route field set: %#v", report.Cases[0].RouteCoverage)
	}
}

func TestOwnerRecallAveragesWrongLayerCasesOnly(t *testing.T) {
	wrong := sampleFixture()
	route := sampleRouteFixture()
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: wrong.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-102"}}}, {CaseID: route.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-200", MatchSources: []string{"graph:uses:c3-201"}}}}}}
	dump := logicalDump{Relationships: []logicalRelationship{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}}}
	report, err := scoreArmResponseWithProbes([]fixtureCase{wrong, route}, response, dump, corpusCombined, map[string]directProbes{route.CaseID: {}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.OwnerRecallAt5 != 0 {
		t.Fatalf("route owner leaked into wrong-layer recall: %v", report.Metrics.OwnerRecallAt5)
	}
	if report.Metrics.RelationshipRouteRecallAt5 != 1 {
		t.Fatal("separate route metric lost")
	}
}

func TestBroadDumpIsScorerOnlyAndFailsPrecisionOrForbiddenRetrieval(t *testing.T) {
	fixture := sampleFixture()
	rows := []cmd.SearchResultRow{{ID: "c3-100"}, {ID: "c3-101"}, {ID: "c3-102"}}
	report, err := scoreArmResponse([]fixtureCase{fixture}, armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: rows}}}, logicalDump{Integrity: "ok"}, corpusCombined)
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.StructuralOwnerPrecision >= 0.8 && report.Metrics.ForbiddenStructuralRetrievalCount == 0 {
		t.Fatalf("broad dump passed both walls: %#v", report.Metrics)
	}
}

func TestOracleTranscriptWouldPassEveryRetrievalGateIfExternallyRatified(t *testing.T) {
	root := repoRoot(t)
	fixtures, _, err := loadFixtures(filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	bench, err := loadBenchmark(filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	baselineResponse := armResponse{Schema: armResponseSchema}
	idealResponse := armResponse{Schema: armResponseSchema}
	dump := logicalDump{Integrity: "ok"}
	for _, fixture := range fixtures {
		if fixture.Family == familyWrongLayer {
			baselineID := entityForFacts(fixture, fixture.Oracle.AllowedExtraFactIDs)
			idealID := entityForFacts(fixture, fixture.Oracle.RequiredOwnerFactIDs)
			if baselineID == "" || idealID == "" {
				t.Fatalf("fixture %s lacks satisfiable fact binding", fixture.CaseID)
			}
			baselineResponse.Cases = append(baselineResponse.Cases, armCaseResult{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: baselineID, Snippet: strings.Repeat("x", 200)}}})
			idealResponse.Cases = append(idealResponse.Cases, armCaseResult{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: idealID, Snippet: "x"}}})
			continue
		}
		w := fixture.Oracle.RelationshipWitness
		if w == nil {
			t.Fatalf("route fixture %s lacks witness", fixture.CaseID)
		}
		dump.Relationships = append(dump.Relationships, logicalRelationship{FromID: w.FromID, ToID: w.ToID, RelType: w.RelType})
		route := cmd.RouteEnrichment{Facts: []string{"fact"}, Graph: []string{"graph"}, Anchors: []string{"anchor"}, Lanes: []string{"lane"}, Hash: strings.Repeat("a", 64)}
		baselineResponse.Cases = append(baselineResponse.Cases, armCaseResult{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: w.ExpectedEntityID, MatchSources: []string{w.ExpectedMatchSource}, Snippet: strings.Repeat("x", 200), Route: route}}})
		idealResponse.Cases = append(idealResponse.Cases, armCaseResult{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: w.ExpectedEntityID, MatchSources: []string{w.ExpectedMatchSource}, Snippet: "x", Route: route}}})
	}
	probes := map[string]directProbes{}
	for _, fixture := range fixtures {
		if fixture.Oracle.RelationshipWitness != nil {
			probes[fixture.CaseID] = directProbes{}
		}
	}
	baseline, err := scoreArmResponseWithProbes(fixtures, baselineResponse, dump, corpusCombined, probes)
	if err != nil {
		t.Fatal(err)
	}
	ideal, err := scoreArmResponseWithProbes(fixtures, idealResponse, dump, corpusCombined, probes)
	if err != nil {
		t.Fatal(err)
	}
	verdict := evaluateGate(baseline, ideal, bench)
	if verdict.Keep || !verdict.WouldPassIfRatified {
		t.Fatalf("oracle transcript did not stay conditional on external ratification: %#v", verdict)
	}
}

func entityForFacts(f fixtureCase, facts []string) string {
	wanted := stringSet(facts)
	for _, entity := range f.Corpus.Entities {
		for _, fact := range f.Oracle.FactBindings[entity.ID] {
			if wanted[fact] {
				return entity.ID
			}
		}
	}
	return ""
}

func TestForgedReportMetricsProvenanceOrHistoryCannotKeep(t *testing.T) {
	fixture := sampleFixture()
	response := armResponse{Schema: armResponseSchema, Cases: []armCaseResult{{CaseID: fixture.CaseID, Rows: []cmd.SearchResultRow{{ID: "c3-100"}}}}}
	dump := logicalDump{Integrity: "ok", Entities: []logicalEntity{{ID: "c3-100"}}}
	report, err := scoreArmResponse([]fixtureCase{fixture}, response, dump, corpusCombined)
	if err != nil {
		t.Fatal(err)
	}
	report.Metrics.OwnerRecallAt5 = 9
	if err := verifyReport([]fixtureCase{fixture}, response, dump, corpusCombined, report); err == nil {
		t.Fatal("forged metric was accepted")
	}
	badProv := validProvenance()
	badProv.RuntimeSHA256 = strings.Repeat("0", 64)
	if err := validateProvenance(badProv); err == nil {
		t.Fatal("zero provenance hash was accepted")
	}
	record := validHistoryRecord()
	record.ResultSHA256 = strings.Repeat("b", 64)
	if err := verifyHistory([]historyRecord{record}); err == nil {
		t.Fatal("tampered history row was accepted")
	}
}

func TestHistoryIsHashChainedAndRequiresFullBudgets(t *testing.T) {
	first := validHistoryRecord()
	first.RecordHash = hashHistoryRecord(first)
	second := first
	second.ExperimentID = "exp-2"
	second.PrevHash = first.RecordHash
	second.RecordHash = hashHistoryRecord(second)
	if err := verifyHistory([]historyRecord{first, second}); err != nil {
		t.Fatal(err)
	}
	second.Budgets.MaxRSSBytes = 0
	second.RecordHash = hashHistoryRecord(second)
	if err := verifyHistory([]historyRecord{first, second}); err == nil {
		t.Fatal("history with an empty budget was accepted")
	}
}

func TestReportAndHistoryWritesAreExclusiveCreateAndAppendOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	if err := writeExclusive(path, []byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(path, []byte("two")); err == nil {
		t.Fatal("write-once report was overwritten")
	}
	history := filepath.Join(dir, "history.jsonl")
	first := validHistoryRecord()
	if err := appendHistory(history, first); err != nil {
		t.Fatal(err)
	}
	second := first
	second.ExperimentID = "exp-2"
	if err := appendHistory(history, second); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(history)
	if bytes.Count(data, []byte("\n")) != 2 {
		t.Fatalf("history row count = %d, want 2", bytes.Count(data, []byte("\n")))
	}
}

func TestRuntimeIsConfinedToControllerTempAndHasAllowlistedEnv(t *testing.T) {
	root := t.TempDir()
	spec, err := newConfinementSpec("/bin/true", root, testControllerBudgetLimits())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateConfinementSpec(spec); err != nil {
		t.Fatal(err)
	}
	if spec.Environment["HOME"] != "/home" || spec.Environment["TZ"] != "UTC" {
		t.Fatalf("environment is not frozen: %#v", spec.Environment)
	}
	if containsString(spec.ReadOnlyBinds, filepath.Dir(repoRoot(t))) {
		t.Fatal("repository was exposed to the runtime")
	}
}

func TestConfinementWallDerivesFromBudgetAuthorityAndExceedsCPUStartupMargin(t *testing.T) {
	limits := testControllerBudgetLimits()
	spec, err := newConfinementSpec("/bin/true", t.TempDir(), limits)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(spec.Command, "\x00")
	if !strings.Contains(joined, "RuntimeMaxSec=60s") {
		t.Fatalf("confinement wall did not derive from authority: %q", joined)
	}
	if strings.Contains(joined, "RuntimeMaxSec=10s") {
		t.Fatalf("legacy duplicate wall remains in command: %q", joined)
	}

	tight := limits
	tight.WallTimeMillis = 15_000
	if _, err := newConfinementSpec("/bin/true", t.TempDir(), tight); err == nil || !strings.Contains(err.Error(), "CPU plus startup margin") {
		t.Fatalf("wall equal to CPU plus startup margin was not rejected: %v", err)
	}
	tightSpec := spec
	tightSpec.BudgetLimits = tight
	for i, arg := range tightSpec.Command {
		if arg == "RuntimeMaxSec=60s" {
			tightSpec.Command[i] = "RuntimeMaxSec=15s"
		}
	}
	if err := validateConfinementSpec(tightSpec); err == nil || !strings.Contains(err.Error(), "CPU plus startup margin") {
		t.Fatalf("confinement validation did not prove wall above CPU plus startup margin: %v", err)
	}
	unregistered := limits
	unregistered.WallTimeMillis = 61_000
	if _, err := newConfinementSpec("/bin/true", t.TempDir(), unregistered); err == nil || !strings.Contains(err.Error(), "registered wall") {
		t.Fatalf("self-consistent but unregistered authority wall was accepted: %v", err)
	}
}

func TestControllerTimeoutDerivesFromAggregateArmWallAndOverhead(t *testing.T) {
	limits := testControllerBudgetLimits()
	got, err := controllerTimeoutFor(6, limits)
	if err != nil {
		t.Fatal(err)
	}
	if got != 600*time.Second {
		t.Fatalf("controller timeout = %s, want 600s", got)
	}
	if err := validateControllerTimeout(599*time.Second, 6, limits); err == nil {
		t.Fatal("controller timeout below aggregate arm wall plus overhead was accepted")
	}
	if err := validateControllerTimeout(got, 6, limits); err != nil {
		t.Fatalf("derived controller timeout rejected: %v", err)
	}
}

func TestRuntimeCannotReadFixtureOracleReportHistoryRepoOrNetwork(t *testing.T) {
	skipUnlessBubblewrap(t)
	if err := proveConfinementBackend(context.Background(), t.TempDir(), testControllerBudgetLimits()); err != nil {
		t.Fatal(err)
	}
}

func TestConfinedRealArmBinaryExecutesRunSearchWithControllerOwnedInspection(t *testing.T) {
	skipUnlessBubblewrap(t)
	runtimePath := buildV2Runtime(t)
	fixture := sampleRouteFixture()
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	got, err := runConfinedArm(ctx, runtimePath, req, t.TempDir(), testControllerBudgetLimits())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cases) != 1 || len(got.Cases[0].Rows) == 0 {
		t.Fatalf("confined real arm returned no raw rows: %#v", got.Cases)
	}
	if got.Database.Integrity != "ok" || got.Database.RelationshipCount != 1 {
		t.Fatalf("controller inspection failed: %#v", got.Database)
	}
	if containsString(got.DirectProbes[fixture.CaseID].EntityFTSIDs, fixture.Oracle.RelationshipWitness.FromID) {
		t.Fatal("controller probe unexpectedly found expansion-only owner directly")
	}
}

// Historical v2 execution fixture retained for read-only compatibility review.
// Normal controller entry points reject this authority before calling it.
func historicalV2ControllerEmitsSeparateRealIsolatedCombinedAndScaleRuns(t *testing.T) {
	runtimePath := buildV2Runtime(t)
	root := repoRoot(t)
	fixturesPath := filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(root, "cli/tools/structural-search-eval-v2/main.go")
	fixtures, _, err := loadFixtures(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	authorityPath := writeControllerAuthority(t, runtimePath, root, fixturesPath, scorerPath, benchmarkPath)
	workRoot := filepath.Join(t.TempDir(), "work")
	retainedRoot, outputDir := inspectableControllerOutputDir(t)
	command := exec.Command(runtimePath, "--controller", "--runtime", runtimePath, "--fixtures", fixturesPath, "--benchmark", benchmarkPath, "--work-root", workRoot, "--authority", authorityPath, "--source-root", root, "--scorer-source", scorerPath, "--output-dir", outputDir)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()
	if runErr != nil {
		t.Fatalf("run built controller: %v; retained_output=%s; refs_stdout=%s; durable_stderr=%s", runErr, retainedRoot, stdout.Bytes(), stderr.Bytes())
	}
	var output controllerOutput
	if err := decodeStrictBytes(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Runs) != len(fixtures)+2 {
		t.Fatalf("controller run count=%d want %d", len(output.Runs), len(fixtures)+2)
	}
	modes := map[string]int{}
	for _, run := range output.Runs {
		modes[run.Mode]++
		if !validSHA256(run.ResultSHA256) || !validSHA256(run.ReportSHA256) || !validSHA256(run.HistoryRecordHash) {
			t.Fatalf("controller emitted an unhashed ref: %#v", run)
		}
		if !actualBudgetComplete(run.ActualBudget) {
			t.Fatalf("controller emitted incomplete actuals: %#v", run.ActualBudget)
		}
	}
	if modes[corpusIsolated] != len(fixtures) || modes[corpusCombined] != 1 || modes["scale"] != 1 {
		t.Fatalf("mode split=%#v", modes)
	}
	if output.Admitted || output.Admission != "diagnostic_unadmitted" || !validSHA256(output.HistorySHA256) {
		t.Fatalf("diagnostic controller output was admitted or unhashed: %#v", output)
	}
}

func historicalV2ControllerRejectsStrippedRuntimeBeforeCreatingWorkOrOutput(t *testing.T) {
	runtimePath := buildV2Runtime(t)
	root := repoRoot(t)
	fixturesPath := filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(root, "cli/tools/structural-search-eval-v2/main.go")
	authorityPath := writeControllerAuthority(t, runtimePath, root, fixturesPath, scorerPath, benchmarkPath)
	stripped := filepath.Join(t.TempDir(), "stripped-runtime")
	strip := exec.Command("strip", "-o", stripped, runtimePath)
	if data, err := strip.CombinedOutput(); err != nil {
		t.Fatalf("strip runtime: %v: %s", err, data)
	}
	workRoot := filepath.Join(t.TempDir(), "work-must-not-exist")
	outputDir := filepath.Join(t.TempDir(), "output-must-not-exist")
	command := exec.Command(runtimePath, "--controller", "--runtime", stripped, "--fixtures", fixturesPath, "--benchmark", benchmarkPath, "--work-root", workRoot, "--authority", authorityPath, "--source-root", root, "--scorer-source", scorerPath, "--output-dir", outputDir)
	if data, err := command.CombinedOutput(); err == nil || !bytes.Contains(data, []byte("runtime SHA-256 mismatch before spawn")) {
		t.Fatalf("stripped runtime was not rejected by pre-spawn authority: err=%v output=%s", err, data)
	}
	for _, path := range []string{workRoot, outputDir} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("preflight failure created %s: %v", path, err)
		}
	}
}

func historicalV2SecondRunFailurePreservesFirstRunAndAppendsFailureEvidence(t *testing.T) {
	runtimePath := buildV2Runtime(t)
	root := repoRoot(t)
	fixturesPath := filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(root, "cli/tools/structural-search-eval-v2/main.go")
	fixtures, _, err := loadFixtures(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) < 2 {
		t.Fatal("failure replay requires at least two frozen fixtures")
	}
	authorityPath := writeControllerAuthority(t, runtimePath, root, fixturesPath, scorerPath, benchmarkPath)
	workRoot := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(workRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	blockedSecondRoot := filepath.Join(workRoot, "isolated-"+safePathComponent(fixtures[1].CaseID))
	if err := os.WriteFile(blockedSecondRoot, []byte("force second-run mkdir failure"), 0o600); err != nil {
		t.Fatal(err)
	}
	retainedRoot, outputDir := inspectableControllerOutputDir(t)
	command := exec.Command(runtimePath, "--controller", "--runtime", runtimePath, "--fixtures", fixturesPath, "--benchmark", benchmarkPath, "--work-root", workRoot, "--authority", authorityPath, "--source-root", root, "--scorer-source", scorerPath, "--output-dir", outputDir)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err == nil {
		t.Fatal("forced second-run failure exited successfully")
	}
	var output controllerOutput
	if err := decodeStrictBytes(stdout.Bytes(), &output); err != nil {
		t.Fatalf("failure stdout is not a strict durable-output ref: %v; retained_output=%s stdout=%s stderr=%s", err, retainedRoot, stdout.Bytes(), stderr.Bytes())
	}
	if len(output.Runs) != 1 || output.Failure == nil {
		t.Fatalf("first success or failure ref was lost; retained_output=%s stdout=%s stderr=%s output=%#v", retainedRoot, stdout.Bytes(), stderr.Bytes(), output)
	}
	if output.Failure.ErrorClass != "runtime_or_inspection" || !validSHA256(output.Failure.ErrorSHA256) || !validSHA256(output.Failure.HistoryRecordHash) {
		t.Fatalf("failure ref is incomplete: %#v", output.Failure)
	}
	for _, ref := range []struct{ path, hash string }{
		{output.Runs[0].ResultPath, output.Runs[0].ResultSHA256},
		{output.Runs[0].ReportPath, output.Runs[0].ReportSHA256},
		{output.Failure.EvidencePath, output.Failure.EvidenceSHA256},
	} {
		got, err := fileSHA256(filepath.Join(outputDir, filepath.FromSlash(ref.path)))
		if err != nil || got != ref.hash {
			t.Fatalf("durable evidence %s missing or changed: got=%s want=%s err=%v", ref.path, got, ref.hash, err)
		}
	}
	for _, snapshot := range []string{"000.json", "001.json", "002.json"} {
		if _, err := os.Stat(filepath.Join(outputDir, "controller-output", snapshot)); err != nil {
			t.Fatalf("append-only output snapshot %s is missing: %v", snapshot, err)
		}
	}
	records, err := readHistory(filepath.Join(outputDir, output.HistoryPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Status != "invalid" || records[1].Status != "crash" || records[1].PrevHash != records[0].RecordHash {
		t.Fatalf("success plus failure history was not preserved: %#v", records)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("durable failure evidence")) || !bytes.Contains(stderr.Bytes(), []byte(output.Failure.HistoryRecordHash)) || !bytes.Contains(stderr.Bytes(), []byte(outputDir)) || !bytes.Contains(stderr.Bytes(), []byte("controller-output/002.json")) {
		t.Fatalf("controller error did not point to durable failure evidence; retained_output=%s stderr=%s", retainedRoot, stderr.Bytes())
	}
}

func historicalV2LiveControllerRuntimeSetupFailureLeavesInspectableEvidence(t *testing.T) {
	runtimePath := buildV2Runtime(t)
	root := repoRoot(t)
	fixturesPath := filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(root, "cli/tools/structural-search-eval-v2/main.go")
	fixtures, _, err := loadFixtures(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	authorityPath := writeControllerAuthority(t, runtimePath, root, fixturesPath, scorerPath, benchmarkPath)
	workRoot := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(workRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	blockedFirstRoot := filepath.Join(workRoot, "isolated-"+safePathComponent(fixtures[0].CaseID))
	if err := os.WriteFile(blockedFirstRoot, []byte("force first runtime setup failure"), 0o600); err != nil {
		t.Fatal(err)
	}
	retainedRoot, outputDir := inspectableControllerOutputDir(t)
	command := exec.Command(runtimePath, "--controller", "--runtime", runtimePath, "--fixtures", fixturesPath, "--benchmark", benchmarkPath, "--work-root", workRoot, "--authority", authorityPath, "--source-root", root, "--scorer-source", scorerPath, "--output-dir", outputDir)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err == nil {
		t.Fatal("forced first-run runtime setup failure exited successfully")
	}
	var output controllerOutput
	if err := decodeStrictBytes(stdout.Bytes(), &output); err != nil {
		t.Fatalf("failure stdout is not refs-only JSON; retained_output=%s stdout=%s stderr=%s err=%v", retainedRoot, stdout.Bytes(), stderr.Bytes(), err)
	}
	if len(output.Runs) != 0 || output.Failure == nil || output.Failure.ErrorClass != "runtime_or_inspection" {
		t.Fatalf("runtime failure envelope is incomplete; retained_output=%s output=%#v", retainedRoot, output)
	}
	if !bytes.Contains(stderr.Bytes(), []byte(outputDir)) || !bytes.Contains(stderr.Bytes(), []byte("controller-output/001.json")) || !bytes.Contains(stderr.Bytes(), []byte(output.Failure.HistoryRecordHash)) {
		t.Fatalf("stderr lacks an absolute durable reference; retained_output=%s stderr=%s", retainedRoot, stderr.Bytes())
	}
	for _, ref := range []struct{ path, hash string }{
		{output.Failure.EvidencePath, output.Failure.EvidenceSHA256},
		{output.HistoryPath, output.HistorySHA256},
		{"controller-output/000.json", ""},
		{"controller-output/001.json", ""},
	} {
		path := filepath.Join(outputDir, filepath.FromSlash(ref.path))
		got, err := fileSHA256(path)
		if err != nil || ref.hash != "" && got != ref.hash {
			t.Fatalf("durable runtime failure evidence is missing; retained_output=%s path=%s got=%s want=%s err=%v", retainedRoot, path, got, ref.hash, err)
		}
	}
	records, err := readHistory(filepath.Join(outputDir, output.HistoryPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Status != "crash" || records[0].ErrorClass != "runtime_or_inspection" || records[0].PrevHash != "GENESIS" {
		t.Fatalf("runtime failure history is not truthful; retained_output=%s history=%#v", retainedRoot, records)
	}
}

func inspectableControllerOutputDir(t *testing.T) (string, string) {
	t.Helper()
	if retained := strings.TrimSpace(os.Getenv("C3_V2_RETAIN_OUTPUT_ROOT")); retained != "" {
		root, output, err := createExclusiveRetainedControllerOutputDir(retained)
		if err != nil {
			t.Fatal(err)
		}
		return root, output
	}
	root, err := os.MkdirTemp("", "c3-v2-live-evidence-")
	if err != nil {
		t.Fatal(err)
	}
	// Passing tests clean up. Failed tests deliberately retain this directory so
	// the failure message's absolute path remains inspectable after t.TempDir is gone.
	t.Cleanup(func() {
		if !t.Failed() {
			_ = os.RemoveAll(root)
		}
	})
	return root, filepath.Join(root, "output")
}

func createExclusiveRetainedControllerOutputDir(root string) (string, string, error) {
	if !filepath.IsAbs(root) {
		return "", "", errors.New("retained controller output root must be absolute")
	}
	if _, err := os.Lstat(root); err == nil {
		return "", "", errors.New("retained controller output root already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("inspect retained controller output root: %w", err)
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		return "", "", fmt.Errorf("create retained controller output root: %w", err)
	}
	return root, filepath.Join(root, "output"), nil
}

func TestInspectableControllerOutputDirCanRetainAtExclusiveExternalRoot(t *testing.T) {
	parent := t.TempDir()
	retained := filepath.Join(parent, "retained")
	var gotRoot, gotOutput string
	t.Run("create", func(t *testing.T) {
		t.Setenv("C3_V2_RETAIN_OUTPUT_ROOT", retained)
		gotRoot, gotOutput = inspectableControllerOutputDir(t)
	})
	if gotRoot != retained || gotOutput != filepath.Join(retained, "output") {
		t.Fatalf("retained paths=(%q,%q)", gotRoot, gotOutput)
	}
	info, err := os.Stat(retained)
	if err != nil {
		t.Fatalf("retained root was cleaned at subtest return: %v", err)
	}
	if !info.IsDir() || info.Mode().Perm() != 0o700 {
		t.Fatalf("retained root mode=%v", info.Mode())
	}
	if _, _, err := createExclusiveRetainedControllerOutputDir("relative"); err == nil {
		t.Fatal("relative retained root was accepted")
	}
	if _, _, err := createExclusiveRetainedControllerOutputDir(retained); err == nil {
		t.Fatal("existing retained root was accepted")
	}
	target := filepath.Join(parent, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(parent, "symlink")
	if err := os.Symlink(target, symlink); err != nil {
		t.Fatal(err)
	}
	if _, _, err := createExclusiveRetainedControllerOutputDir(symlink); err == nil {
		t.Fatal("symlink retained root was accepted")
	}
}

func TestMaliciousConfinedArmCannotReadControllerFilesOrNetworkAndCannotForgeOutput(t *testing.T) {
	skipUnlessBubblewrap(t)
	root := t.TempDir()
	secret := filepath.Join(root, "oracle.json")
	report := filepath.Join(root, "report.json")
	history := filepath.Join(root, "history.jsonl")
	for _, path := range []string{secret, report, history} {
		if err := os.WriteFile(path, []byte("controller-only"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	armRoot := filepath.Join(root, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(root, "malicious-arm.sh")
	body := "#!/bin/sh\ncat >/dev/null\nif test -e " + shellQuote(secret) + " || test -e " + shellQuote(report) + " || test -e " + shellQuote(history) + " || /usr/bin/curl -fsS --connect-timeout 1 http://1.1.1.1 >/dev/null 2>&1; then\n  printf 'log\\n'\n  exit 0\nfi\nprintf '%s\\n' '{\"$schema\":\"structural-retrieval-arm-response.v2\",\"cases\":[]}'\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := invokeConfinedArm(ctx, script, armRequest{Schema: armRequestSchema, SemanticMode: semanticDisabled}, armRoot, testControllerBudgetLimits()); err != nil {
		t.Fatalf("malicious isolation probe failed: %v", err)
	}
	forged := filepath.Join(root, "forged-arm.sh")
	forgedBody := "#!/bin/sh\ncat >/dev/null\nprintf 'log\\n%s\\n' '{\"$schema\":\"structural-retrieval-arm-response.v2\",\"cases\":[]}'\n"
	if err := os.WriteFile(forged, []byte(forgedBody), 0o700); err != nil {
		t.Fatal(err)
	}
	secondRoot := filepath.Join(root, "arm-forged")
	if err := os.Mkdir(secondRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := invokeConfinedArm(ctx, forged, armRequest{Schema: armRequestSchema, SemanticMode: semanticDisabled}, secondRoot, testControllerBudgetLimits()); err == nil {
		t.Fatal("forged log prefix was accepted from a confined arm")
	}
}

func buildV2Runtime(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "structural-search-eval-v2")
	command := exec.Command("go", "build", "-trimpath", "-o", out, "./tools/structural-search-eval-v2")
	command.Dir = filepath.Join(repoRoot(t), "cli")
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build v2 runtime: %v: %s", err, data)
	}
	return out
}

func writeControllerAuthority(t *testing.T, runtimePath, root, fixturesPath, scorerPath, benchmarkPath string) string {
	t.Helper()
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil {
		t.Fatal(err)
	}
	capsule, err := captureSourceCapsule(root)
	if err != nil {
		t.Fatal(err)
	}
	runtimeHash, err := fileSHA256(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	fixtureHash, err := fileSHA256(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	scorerHash, err := fileSHA256(scorerPath)
	if err != nil {
		t.Fatal(err)
	}
	environmentHash, err := environmentSHA256(root)
	if err != nil {
		t.Fatal(err)
	}
	moduleHash, err := moduleGraphSHA256(root)
	if err != nil {
		t.Fatal(err)
	}
	limits := testControllerBudgetLimits()
	actionEnvelope := "diagnostic-only real structural retrieval replay; no product writes"
	p := provenance{ExperimentID: "real-adapter-v2-diagnostic", ArmID: "diagnostic", Commit: capsule.HeadCommit, Tree: capsule.HeadTree, SourceCapsuleSHA256: canonicalSHA256(capsule), DiffSHA256: capsule.DirtyPatchSHA256, FixtureSHA256: fixtureHash, ScorerSHA256: scorerHash, ControllerSHA256: runtimeHash, RuntimeSHA256: runtimeHash, EnvironmentSHA256: environmentHash, ModuleGraphSHA256: moduleHash, BudgetSHA256: canonicalSHA256(limits), ActionEnvelopeSHA256: shaString(actionEnvelope), SemanticMode: bench.SemanticMode, ContextThresholdAuthoritySHA256: hashThresholdAuthority(bench.ContextThresholdAuthority)}
	line := checkinLine(t, filepath.Join(root, ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"), 5)
	authority := controllerAuthority{Schema: controllerAuthoritySchema, Expected: p, SourceCapsule: capsule, BuildReplay: buildReplayManifest{ControllerSHA256: runtimeHash, RuntimeSHA256: runtimeHash, RebuiltControllerSHA256: runtimeHash, RebuiltRuntimeSHA256: runtimeHash, SourceCapsuleRebuildVerified: true, BundleVerified: false}, BudgetLimits: limits, ActionEnvelope: actionEnvelope, CanonicalRowBytesDefinition: canonicalRowBytesDefinition, ContextThresholdAuthorityRecord: string(line)}
	data, err := json.Marshal(authority)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "controller-authority.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func testControllerBudgetLimits() resourceBudget {
	return resourceBudget{WallTimeMillis: registeredWallTimeMillis, CPUTimeMillis: 10_000, MaxRSSBytes: 536_870_912, ProcessCount: 16, SQLiteRowCount: 1_000_000, LogicalDumpBytes: 64 << 20, StdoutBytes: 16 << 20, StderrBytes: 1 << 20, CaseCount: 100}
}
func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

func TestConfinementCapabilityFailureInvalidatesArm(t *testing.T) {
	spec, err := newConfinementSpec("/missing/runtime", t.TempDir(), testControllerBudgetLimits())
	if err == nil || spec.RuntimePath != "" {
		t.Fatalf("missing runtime did not fail closed: spec=%#v err=%v", spec, err)
	}
}

func TestBaselineBMatchesInspectedDirtyBuildInputClosureOnly(t *testing.T) {
	capsule := validSourceCapsule()
	if err := validateSourceCapsule(capsule); err != nil {
		t.Fatal(err)
	}
	capsule.Inputs = append(capsule.Inputs, sourceInput{Path: ".okra/private.json", SHA256: strings.Repeat("a", 64), Origin: "working-tree"})
	if err := validateSourceCapsule(capsule); err == nil {
		t.Fatal("private unrelated source path was accepted")
	}
}

func TestBundleRebuildMatchesControllerAndRuntimeHashes(t *testing.T) {
	manifest := buildReplayManifest{ControllerSHA256: strings.Repeat("a", 64), RuntimeSHA256: strings.Repeat("b", 64), RebuiltControllerSHA256: strings.Repeat("a", 64), RebuiltRuntimeSHA256: strings.Repeat("b", 64), SourceCapsuleRebuildVerified: true, BundleVerified: false}
	if err := validateBuildReplay(manifest); err != nil {
		t.Fatal(err)
	}
	manifest.RebuiltRuntimeSHA256 = strings.Repeat("c", 64)
	if err := validateBuildReplay(manifest); err == nil {
		t.Fatal("mismatched rebuild was accepted")
	}
}

func TestControllerRejectsUnverifiedCommitDiffBeforeBuild(t *testing.T) {
	p := validProvenance()
	p.DiffSHA256 = ""
	if err := validateProvenance(p); err == nil {
		t.Fatal("missing diff hash was accepted")
	}
	p = validProvenance()
	p.SourceCapsuleSHA256 = strings.Repeat("0", 64)
	if err := validateProvenance(p); err == nil {
		t.Fatal("zero source capsule hash was accepted")
	}
}

func TestSourceCapsuleRequiresCorrectSemanticHashAndBuildClosure(t *testing.T) {
	capsule := validSourceCapsule()
	capsule.Inputs = append(capsule.Inputs, sourceInput{Path: "cli/internal/store/semantic.go", SHA256: "7b79f5d218fb422654174c5d651a55b0ac2b3fb1f38f9bb048f03492afc34883", Origin: "head"})
	capsule.RepositoryBuildInputCount++
	if err := validateSourceCapsule(capsule); err != nil {
		t.Fatal(err)
	}
	capsule.Inputs[len(capsule.Inputs)-1].SHA256 = "7b79f5d218fb422654175c5d651a55b0ac2b3fb1f38f9bb048f03492afc34883"
	if err := validateSourceCapsule(capsule); err == nil {
		t.Fatal("known semantic.go hash typo was accepted")
	}
}

func TestCandidateRejectsNewImportInitEnvFileNetworkProcessStdoutAndFixtureConstant(t *testing.T) {
	base := `package store
import "database/sql"
const schemaSQL = "CREATE VIRTUAL TABLE entities_fts USING fts5(title, goal)"
func migrateSchema() error { return nil }
`
	badBodies := []string{
		`import ("database/sql"; "os")`,
		`func init(){}`,
		`os.Getenv("X")`,
		`os.ReadFile("fixture")`,
		`net.Dial("tcp", "x")`,
		`exec.Command("sh")`,
		`fmt.Println("answer")`,
		`const gold = "expected answer"`,
	}
	for _, bad := range badBodies {
		candidate := strings.Replace(base, `import "database/sql"`, bad, 1)
		if err := validateCandidateSource(base, candidate); err == nil {
			t.Fatalf("candidate source accepted forbidden construct %q", bad)
		}
	}
}

func TestCandidateASTAndNormalizedSQLMatchRegisteredParentKeywordBehavior(t *testing.T) {
	base := `package store
const schemaSQL = "CREATE VIRTUAL TABLE entities_fts USING fts5(title, goal)"
func migrateSchema() error { return nil }
`
	candidate := `package store
const schemaSQL = "CREATE VIRTUAL TABLE entities_fts USING fts5(title, goal, parent_id)"
func migrateSchema() error { return migrateParentKeywordFTS() }
func migrateParentKeywordFTS() error { return nil }
`
	if err := validateCandidateSource(base, candidate); err != nil {
		t.Fatal(err)
	}
	candidate = strings.Replace(candidate, "parent_id", "parent_title", 1)
	if err := validateCandidateSource(base, candidate); err == nil {
		t.Fatal("unregistered SQL column was accepted")
	}
}

func TestDefaultNumericParentNaturalLanguageQueryDoesNotGetCreditFromExplicitIDLift(t *testing.T) {
	results := familyResults{DefaultIDNaturalLanguageLift: 0, ExplicitIDLift: 1, SemanticIDLift: 1}
	if familyDecision(results) != decisionDiscard {
		t.Fatalf("family decision = %q, want discard", familyDecision(results))
	}
}

func TestNoSemanticResultTransfersAsCandidateOnlyToDefaultHybrid(t *testing.T) {
	transfer := transferRecord{SourceSemanticMode: semanticDisabled, DestinationSemanticMode: semanticDefault, NumericComparison: false, Status: "candidate"}
	if err := validateTransfer(transfer); err != nil {
		t.Fatal(err)
	}
	transfer.NumericComparison = true
	if err := validateTransfer(transfer); err == nil {
		t.Fatal("numeric cross-DKR comparison was accepted")
	}
}

func TestWriterCannotReadFreshGenericHoldoutAndAccessLogIsHashChained(t *testing.T) {
	policy := holdoutPolicy{ContentVisibleToWriter: false, ContentVisibleToCandidate: false, AllowedPurposes: []string{"post-selection-confirmation", "audit"}, MaxScoreReads: 1}
	if err := validateHoldoutPolicy(policy); err != nil {
		t.Fatal(err)
	}
	policy.ContentVisibleToWriter = true
	if err := validateHoldoutPolicy(policy); err == nil {
		t.Fatal("writer-visible holdout was accepted")
	}
}

func TestFreshCreateUsesExpectedFTSSchemaAndTriggers(t *testing.T) {
	fixture := sampleFixture()
	req, err := buildArmRequest([]fixtureCase{fixture}, corpusIsolated, semanticDisabled, sampleBenchmark())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	got, err := runArm(req, filepath.Join(root, "db", "c3.db"), filepath.Join(root, "project"), filepath.Join(root, "c3"))
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got.Database.SchemaSQL, "\n")
	for _, want := range []string{"entities_fts", "entities_ai", "entities_ad", "entities_au"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("fresh logical schema omits %q", want)
		}
	}
}

func TestLegacyTwoColumnMigrationRebuildsOnceAndIsIdempotent(t *testing.T) {
	base := legacySchemaState{FTSColumns: []string{"title", "goal"}, LogicalRowsSHA256: strings.Repeat("a", 64)}
	migrated, err := simulateRegisteredMigration(base)
	if err != nil {
		t.Fatal(err)
	}
	again, err := simulateRegisteredMigration(migrated)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(migrated, again) || !reflect.DeepEqual(migrated.FTSColumns, []string{"title", "goal", "parent_id"}) {
		t.Fatalf("migration is not idempotent: first=%#v second=%#v", migrated, again)
	}
	unknown := base
	unknown.FTSColumns = []string{"title", "goal", "other"}
	if _, err := simulateRegisteredMigration(unknown); err == nil {
		t.Fatal("unknown FTS schema was silently rewritten")
	}
}

func TestResourceBudgetOverflowIsCrashOrInvalidNeverScore(t *testing.T) {
	limit := resourceBudget{WallTimeMillis: 100, CPUTimeMillis: 100, MaxRSSBytes: 1024, ProcessCount: 1, StdoutBytes: 100, StderrBytes: 100, CaseCount: 1, SQLiteRowCount: 10, LogicalDumpBytes: 1000}
	actual := limit
	actual.StdoutBytes++
	if status := classifyBudget(limit, actual); status != "crash" {
		t.Fatalf("budget overflow status = %q, want crash", status)
	}
}

func TestProtocolV6ScoringRegionUsesExactGoTokenOffsets(t *testing.T) {
	root := repoRoot(t)
	got, err := scoringRegionSHA256(filepath.Join(root, "cli/tools/structural-search-eval-v2/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	const want = "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb"
	if got != want {
		t.Fatalf("scoring region sha256 = %s, want %s", got, want)
	}
}

func TestProtocolV6DualSourceCapsulesAllowOnlyRegisteredRuntimePath(t *testing.T) {
	clean := shaString("")
	baseline := sourceCapsule{
		Schema: sourceCapsuleSchema, HeadCommit: strings.Repeat("a", 40), HeadTree: strings.Repeat("b", 40),
		RepositoryBuildInputCount: 3, DirtyPatchSHA256: clean,
		Inputs: []sourceInput{
			{Path: "cli/cmd/search.go", SHA256: strings.Repeat("1", 64), Origin: "head"},
			{Path: "cli/go.mod", SHA256: strings.Repeat("2", 64), Origin: "head"},
			{Path: "cli/tools/structural-search-eval-v2/main.go", SHA256: strings.Repeat("3", 64), Origin: "head"},
		},
	}
	candidate := baseline
	candidate.HeadCommit = strings.Repeat("c", 40)
	candidate.HeadTree = strings.Repeat("d", 40)
	candidate.Inputs = append([]sourceInput(nil), baseline.Inputs...)
	candidate.Inputs[0].SHA256 = strings.Repeat("4", 64)
	if err := validateDualSourceCapsules(baseline, candidate, []string{"cli/cmd/search.go"}); err != nil {
		t.Fatal(err)
	}
	candidate.Inputs[1].SHA256 = strings.Repeat("5", 64)
	if err := validateDualSourceCapsules(baseline, candidate, []string{"cli/cmd/search.go"}); err == nil {
		t.Fatal("unregistered build-input drift was accepted")
	}
}

func TestProtocolV6CanonicalCandidateDeltaIsStableAndSensitive(t *testing.T) {
	delta := candidateDelta{
		Variable:       "direct_hit_containment_owner_substitution",
		BaselineCommit: strings.Repeat("a", 40), BaselineTree: strings.Repeat("b", 40),
		CandidateCommit: strings.Repeat("c", 40), CandidateTree: strings.Repeat("d", 40),
		DiffSHA256: strings.Repeat("1", 64), NameStatusSHA256: strings.Repeat("2", 64),
		NameStatus: []string{"M\tcli/cmd/search.go"}, AllowedPaths: []string{"cli/cmd/search.go"},
		BeforeBlobSHA256: map[string]string{"cli/cmd/search.go": strings.Repeat("3", 64)},
		AfterBlobSHA256:  map[string]string{"cli/cmd/search.go": strings.Repeat("4", 64)},
		BundleSHA256:     strings.Repeat("5", 64), BundleHeadsSHA256: strings.Repeat("6", 64),
	}
	first := canonicalCandidateDeltaSHA256(delta)
	second := canonicalCandidateDeltaSHA256(delta)
	if first != second || !validSHA256(first) {
		t.Fatalf("candidate delta hash is unstable: first=%s second=%s", first, second)
	}
	delta.AfterBlobSHA256["cli/cmd/search.go"] = strings.Repeat("7", 64)
	if canonicalCandidateDeltaSHA256(delta) == first {
		t.Fatal("candidate delta hash ignored an after-blob change")
	}
}

func TestProtocolV6CandidateDeltaIdentityMustMatchSelectedSourceCapsules(t *testing.T) {
	baselineCommit := strings.Repeat("a", 40)
	baselineTree := strings.Repeat("b", 40)
	candidateCommit := strings.Repeat("c", 40)
	candidateTree := strings.Repeat("d", 40)
	authority := controllerAuthorityV3{
		Mode: "candidate",
		Expected: provenance{
			ControllerCommit: baselineCommit,
			ControllerTree:   baselineTree,
			Commit:           candidateCommit,
			Tree:             candidateTree,
		},
		ControllerSourceCapsule: sourceCapsule{HeadCommit: baselineCommit, HeadTree: baselineTree},
		RuntimeSourceCapsule:    sourceCapsule{HeadCommit: candidateCommit, HeadTree: candidateTree},
		CandidateDelta: &candidateDelta{
			BaselineCommit:  baselineCommit,
			BaselineTree:    baselineTree,
			CandidateCommit: candidateCommit,
			CandidateTree:   candidateTree,
		},
	}
	if err := validateCandidateDeltaAuthorityBinding(authority); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*candidateDelta){
		"baseline commit":  func(delta *candidateDelta) { delta.BaselineCommit = strings.Repeat("1", 40) },
		"baseline tree":    func(delta *candidateDelta) { delta.BaselineTree = strings.Repeat("2", 40) },
		"candidate commit": func(delta *candidateDelta) { delta.CandidateCommit = strings.Repeat("3", 40) },
		"candidate tree":   func(delta *candidateDelta) { delta.CandidateTree = strings.Repeat("4", 40) },
	} {
		t.Run(name, func(t *testing.T) {
			tampered := authority
			delta := *authority.CandidateDelta
			mutate(&delta)
			tampered.CandidateDelta = &delta
			if err := validateCandidateDeltaAuthorityBinding(tampered); err == nil {
				t.Fatal("unrelated candidate delta identity was accepted")
			}
		})
	}
}

func TestProtocolV6GitProofIgnoresAmbientRepositoryRedirects(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "init", "-q")
	runGitTest(t, repo, "config", "user.name", "C3 Eval")
	runGitTest(t, repo, "config", "user.email", "c3-eval@invalid")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("frozen\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "add", "tracked.txt")
	runGitTest(t, repo, "commit", "-q", "-m", "baseline")

	hostile := filepath.Join(t.TempDir(), "ambient-git-dir")
	for key, value := range map[string]string{
		"GIT_DIR":                          hostile,
		"GIT_WORK_TREE":                    hostile,
		"GIT_COMMON_DIR":                   hostile,
		"GIT_OBJECT_DIRECTORY":             hostile,
		"GIT_INDEX_FILE":                   filepath.Join(hostile, "index"),
		"GIT_SHALLOW_FILE":                 filepath.Join(hostile, "shallow"),
		"GIT_NAMESPACE":                    "hostile",
		"GIT_CEILING_DIRECTORIES":          repo,
		"GIT_DISCOVERY_ACROSS_FILESYSTEM":  "1",
		"GIT_CONFIG_COUNT":                 "1",
		"GIT_CONFIG_KEY_0":                 "core.bare",
		"GIT_CONFIG_VALUE_0":               "true",
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": hostile,
	} {
		t.Setenv(key, value)
	}
	got, err := gitCommandBytes(repo, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != repo {
		t.Fatalf("git proof used ambient repository: got %q want %q", strings.TrimSpace(string(got)), repo)
	}
}

func TestProtocolV6CandidateGitDeltaRequiresCleanDirectChildAndExactBundle(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, "cli/cmd"), 0o755); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "init", "-q")
	runGitTest(t, repo, "config", "user.name", "C3 Eval")
	runGitTest(t, repo, "config", "user.email", "c3-eval@invalid")
	searchPath := filepath.Join(repo, "cli/cmd/search.go")
	if err := os.WriteFile(searchPath, []byte("package cmd\nfunc searchVersion() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "add", "cli/cmd/search.go")
	runGitTest(t, repo, "commit", "-q", "-m", "baseline")
	baselineCommit := strings.TrimSpace(runGitTest(t, repo, "rev-parse", "HEAD"))
	baselineTree := strings.TrimSpace(runGitTest(t, repo, "rev-parse", "HEAD^{tree}"))
	before := mustFileHash(t, searchPath)
	if err := os.WriteFile(searchPath, []byte("package cmd\nfunc searchVersion() int { return 2 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "add", "cli/cmd/search.go")
	runGitTest(t, repo, "commit", "-q", "-m", "candidate")
	candidateCommit := strings.TrimSpace(runGitTest(t, repo, "rev-parse", "HEAD"))
	candidateTree := strings.TrimSpace(runGitTest(t, repo, "rev-parse", "HEAD^{tree}"))
	after := mustFileHash(t, searchPath)
	ref := "refs/c3-eval/commit-pool/protocol-v6-test"
	runGitTest(t, repo, "update-ref", ref, candidateCommit)
	bundle := filepath.Join(t.TempDir(), "candidate.bundle")
	runGitTest(t, repo, "bundle", "create", bundle, ref)
	diffBytes := []byte(runGitTest(t, repo, "diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--no-renames", baselineCommit+".."+candidateCommit))
	nameStatusBytes := []byte(runGitTest(t, repo, "diff", "--name-status", "--no-renames", baselineCommit+".."+candidateCommit))
	delta := candidateDelta{
		Variable:       "direct_hit_containment_owner_substitution",
		BaselineCommit: baselineCommit, BaselineTree: baselineTree, CandidateCommit: candidateCommit, CandidateTree: candidateTree,
		DiffSHA256: shaBytes(diffBytes), NameStatusSHA256: shaBytes(nameStatusBytes),
		NameStatus: []string{"M\tcli/cmd/search.go"}, AllowedPaths: []string{"cli/cmd/search.go"},
		BeforeBlobSHA256: map[string]string{"cli/cmd/search.go": before}, AfterBlobSHA256: map[string]string{"cli/cmd/search.go": after},
		BundleSHA256: mustFileHash(t, bundle), BundleHeadsSHA256: shaBytes([]byte(runGitTest(t, repo, "bundle", "list-heads", bundle))),
	}
	if err := verifyCandidateDeltaGit(repo, bundle, delta); err != nil {
		t.Fatal(err)
	}
	extraBlob := delta
	extraBlob.BeforeBlobSHA256 = map[string]string{
		"cli/cmd/search.go": delta.BeforeBlobSHA256["cli/cmd/search.go"],
		"cli/cmd/extra.go":  strings.Repeat("8", 64),
	}
	if err := verifyCandidateDeltaGit(repo, bundle, extraBlob); err == nil {
		t.Fatal("candidate delta accepted an extra before-blob key")
	}
	runGitTest(t, repo, "update-ref", "refs/c3-eval/commit-pool/protocol-v6-baseline-test", baselineCommit)
	multiHeadBundle := filepath.Join(t.TempDir(), "multi-head.bundle")
	runGitTest(t, repo, "bundle", "create", multiHeadBundle, ref, "refs/c3-eval/commit-pool/protocol-v6-baseline-test")
	multiHead := delta
	multiHead.BundleSHA256 = mustFileHash(t, multiHeadBundle)
	multiHead.BundleHeadsSHA256 = shaBytes([]byte(runGitTest(t, repo, "bundle", "list-heads", multiHeadBundle)))
	if err := verifyCandidateDeltaGit(repo, multiHeadBundle, multiHead); err == nil {
		t.Fatal("candidate delta accepted a multi-head bundle")
	}
	delta.BaselineCommit = candidateCommit
	if err := verifyCandidateDeltaGit(repo, bundle, delta); err == nil {
		t.Fatal("non-child/self-baselining candidate was accepted")
	}
}

func TestProtocolV6AuthorityRejectsUnknownFields(t *testing.T) {
	input := `{"$schema":"structural-retrieval-controller-authority.v3","unknown":true}`
	if _, err := decodeControllerAuthorityV3(strings.NewReader(input)); err == nil {
		t.Fatal("v3 authority accepted an unknown field")
	}
}

func TestProtocolV6AuthorityPrivacyRejectsAbsolutePathsAndCredentialLikeValues(t *testing.T) {
	for _, value := range []string{"inspect /" + "home/user/repo", "token gh" + "p_abcdefghijklmnopqrstuvwxyz"} {
		authority := controllerAuthorityV3{Schema: controllerAuthorityV3Schema, Mode: "candidate", ActionEnvelope: value}
		if err := verifyPortableAuthorityPrivacy(authority); err == nil {
			t.Fatalf("authority privacy accepted %q", value)
		}
	}
}

func TestProtocolV6BuildReplayBindsDistinctIndependentRebuilds(t *testing.T) {
	replay := buildReplayManifest{
		ControllerSHA256: strings.Repeat("a", 64), RuntimeSHA256: strings.Repeat("b", 64),
		RebuiltControllerSHA256: strings.Repeat("a", 64), RebuiltRuntimeSHA256: strings.Repeat("b", 64),
		ControllerCapsuleRebuildVerified: true, RuntimeCapsuleRebuildVerified: true, BundleVerified: true,
	}
	if err := validateBuildReplay(replay); err != nil {
		t.Fatal(err)
	}
	replay.RebuiltRuntimeSHA256 = replay.RebuiltControllerSHA256
	if err := validateBuildReplay(replay); err == nil {
		t.Fatal("one rebuilt binary was accepted for distinct controller and runtime")
	}
}

func TestProtocolV6ProvenanceRequiresControllerRuntimeDeltaAndBundle(t *testing.T) {
	p := validProvenance()
	p.ControllerCommit = strings.Repeat("1", 40)
	p.ControllerTree = strings.Repeat("2", 40)
	p.ControllerSourceCapsuleSHA256 = strings.Repeat("3", 64)
	p.CandidateDeltaSHA256 = strings.Repeat("4", 64)
	p.BundleSHA256 = strings.Repeat("5", 64)
	p.BenchmarkSHA256 = strings.Repeat("6", 64)
	if err := validateProvenance(p); err != nil {
		t.Fatal(err)
	}
	p.BundleSHA256 = ""
	if err := validateProvenance(p); err == nil {
		t.Fatal("v3 provenance without bundle binding was accepted")
	}
}

func TestProtocolV6ControllerSelectsDualRootsBeforeAnyOutput(t *testing.T) {
	root := t.TempDir()
	authorityPath := filepath.Join(root, "authority.json")
	if err := os.WriteFile(authorityPath, []byte(`{"$schema":"structural-retrieval-controller-authority.v3","mode":"candidate"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	workRoot := filepath.Join(root, "work-must-not-exist")
	outputRoot := filepath.Join(root, "output-must-not-exist")
	var stdout bytes.Buffer
	err := runControllerCLI([]string{
		"--runtime", filepath.Join(root, "runtime"),
		"--fixtures", filepath.Join(root, "fixtures.jsonl"),
		"--benchmark", filepath.Join(root, "benchmark.json"),
		"--work-root", workRoot,
		"--authority", authorityPath,
		"--controller-source-root", filepath.Join(root, "B"),
		"--runtime-source-root", filepath.Join(root, "C"),
		"--bundle", filepath.Join(root, "candidate.bundle"),
		"--scorer-source", filepath.Join(root, "B/main.go"),
		"--output-dir", outputRoot,
	}, &stdout)
	if err == nil || strings.Contains(err.Error(), "--source-root is required") {
		t.Fatalf("controller did not select v3 dual roots: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("preflight failure wrote stdout: %q", stdout.String())
	}
	for _, path := range []string{workRoot, outputRoot} {
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("preflight failure created %s: %v", path, statErr)
		}
	}
}

func TestProtocolV6AuthorityAcceptsFrozenBControllerAndDistinctCRuntime(t *testing.T) {
	skipUnlessReleaseFixtures(t, filepath.Join(repoRoot(t), ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"))
	root := repoRoot(t)
	base := filepath.Join(t.TempDir(), "B")
	paths, err := discoverRepositoryBuildInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths,
		"research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
		"research/eval/structural-retrieval/benchmark.v2.json",
		"cli/tools/structural-search-eval-v2/main_test.go",
	)
	for _, relative := range paths {
		from := filepath.Join(root, filepath.FromSlash(relative))
		to := filepath.Join(base, filepath.FromSlash(relative))
		data, err := os.ReadFile(from)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGitTest(t, base, "init", "-q")
	runGitTest(t, base, "config", "user.name", "C3 Eval")
	runGitTest(t, base, "config", "user.email", "c3-eval@invalid")
	runGitTest(t, base, "add", ".")
	runGitTest(t, base, "commit", "-q", "-m", "portable baseline B")
	baselineCommit := strings.TrimSpace(runGitTest(t, base, "rev-parse", "HEAD"))
	baselineTree := strings.TrimSpace(runGitTest(t, base, "rev-parse", "HEAD^{tree}"))
	candidate := filepath.Join(t.TempDir(), "C")
	runGitTest(t, filepath.Dir(candidate), "clone", "-q", base, candidate)
	runGitTest(t, candidate, "config", "user.name", "C3 Eval")
	runGitTest(t, candidate, "config", "user.email", "c3-eval@invalid")
	searchPath := filepath.Join(candidate, "cli/cmd/search.go")
	searchSource, err := os.ReadFile(searchPath)
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(searchSource, []byte("defaultSearchLimit      = 20"), []byte("defaultSearchLimit      = 21"), 1)
	if bytes.Equal(changed, searchSource) {
		t.Fatal("protocol test candidate did not change search.go")
	}
	if err := os.WriteFile(searchPath, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, candidate, "add", "cli/cmd/search.go")
	runGitTest(t, candidate, "commit", "-q", "-m", "protocol test candidate C")
	candidateCommit := strings.TrimSpace(runGitTest(t, candidate, "rev-parse", "HEAD"))
	candidateTree := strings.TrimSpace(runGitTest(t, candidate, "rev-parse", "HEAD^{tree}"))
	ref := "refs/c3-eval/commit-pool/protocol-v6-dual-source-test"
	runGitTest(t, candidate, "update-ref", ref, candidateCommit)
	bundle := filepath.Join(t.TempDir(), "candidate.bundle")
	runGitTest(t, candidate, "bundle", "create", bundle, ref)
	rangeArg := baselineCommit + ".." + candidateCommit
	diffBytes := []byte(runGitTest(t, candidate, "diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--no-renames", rangeArg))
	nameStatusBytes := []byte(runGitTest(t, candidate, "diff", "--name-status", "--no-renames", rangeArg))
	delta := candidateDelta{
		Variable:       "direct_hit_containment_owner_substitution",
		BaselineCommit: baselineCommit, BaselineTree: baselineTree, CandidateCommit: candidateCommit, CandidateTree: candidateTree,
		DiffSHA256: shaBytes(diffBytes), NameStatusSHA256: shaBytes(nameStatusBytes),
		NameStatus: []string{"M\tcli/cmd/search.go"}, AllowedPaths: []string{"cli/cmd/search.go"},
		BeforeBlobSHA256: map[string]string{"cli/cmd/search.go": mustGitFileSHA256(t, candidate, baselineCommit, "cli/cmd/search.go")},
		AfterBlobSHA256:  map[string]string{"cli/cmd/search.go": mustGitFileSHA256(t, candidate, candidateCommit, "cli/cmd/search.go")},
		BundleSHA256:     mustFileHash(t, bundle), BundleHeadsSHA256: shaBytes([]byte(runGitTest(t, candidate, "bundle", "list-heads", bundle))),
	}
	controllerBinary := filepath.Join(t.TempDir(), "controller")
	runtimeBinary := filepath.Join(t.TempDir(), "runtime")
	if err := buildFrozenRuntime(base, controllerBinary); err != nil {
		t.Fatal(err)
	}
	if err := buildFrozenRuntime(candidate, runtimeBinary); err != nil {
		t.Fatal(err)
	}
	controllerHash := mustFileHash(t, controllerBinary)
	runtimeHash := mustFileHash(t, runtimeBinary)
	if controllerHash == runtimeHash {
		t.Fatal("protocol test did not produce a distinct candidate runtime")
	}
	controllerCapsule, err := captureSourceCapsule(base)
	if err != nil {
		t.Fatal(err)
	}
	runtimeCapsule, err := captureSourceCapsule(candidate)
	if err != nil {
		t.Fatal(err)
	}
	fixturesPath := filepath.Join(base, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(base, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(base, "cli/tools/structural-search-eval-v2/main.go")
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil {
		t.Fatal(err)
	}
	limits := testControllerBudgetLimits()
	actionEnvelope := "protocol-v6 test-only dual-source preflight; no retrieval score"
	line := checkinLine(t, filepath.Join(root, ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"), 5)
	environmentHash, err := environmentSHA256(base)
	if err != nil {
		t.Fatal(err)
	}
	moduleHash, err := moduleGraphSHA256(base)
	if err != nil {
		t.Fatal(err)
	}
	authority := controllerAuthorityV3{
		Schema: controllerAuthorityV3Schema, Mode: "candidate",
		Expected: provenance{
			ExperimentID: "protocol-v6-test", ArmID: "diagnostic",
			Commit: candidateCommit, Tree: candidateTree, SourceCapsuleSHA256: canonicalSHA256(runtimeCapsule), DiffSHA256: delta.DiffSHA256,
			ControllerCommit: baselineCommit, ControllerTree: baselineTree, ControllerSourceCapsuleSHA256: canonicalSHA256(controllerCapsule),
			CandidateDeltaSHA256: canonicalCandidateDeltaSHA256(delta), BundleSHA256: delta.BundleSHA256,
			FixtureSHA256: mustFileHash(t, fixturesPath), BenchmarkSHA256: mustFileHash(t, benchmarkPath), ScorerSHA256: mustFileHash(t, scorerPath),
			ControllerSHA256: controllerHash, RuntimeSHA256: runtimeHash, EnvironmentSHA256: environmentHash, ModuleGraphSHA256: moduleHash,
			BudgetSHA256: canonicalSHA256(limits), ActionEnvelopeSHA256: shaString(actionEnvelope), SemanticMode: bench.SemanticMode,
			ContextThresholdAuthoritySHA256: hashThresholdAuthority(bench.ContextThresholdAuthority),
		},
		ControllerSourceCapsule: controllerCapsule, RuntimeSourceCapsule: runtimeCapsule, CandidateDelta: &delta,
		BuildReplay: buildReplayManifest{
			ControllerSHA256: controllerHash, RuntimeSHA256: runtimeHash, RebuiltControllerSHA256: controllerHash, RebuiltRuntimeSHA256: runtimeHash,
			ControllerCapsuleRebuildVerified: true, RuntimeCapsuleRebuildVerified: true, BundleVerified: true,
		},
		BudgetLimits: limits, ActionEnvelope: actionEnvelope, CanonicalRowBytesDefinition: canonicalRowBytesDefinition, ContextThresholdAuthorityRecord: string(line),
	}
	if err := verifyControllerAuthorityV3(authority, runtimeBinary, controllerBinary, base, candidate, bundle, fixturesPath, benchmarkPath, scorerPath, bench); err != nil {
		t.Fatal(err)
	}
	tampered := authority
	tampered.BuildReplay.RebuiltRuntimeSHA256 = controllerHash
	if err := verifyControllerAuthorityV3(tampered, runtimeBinary, controllerBinary, base, candidate, bundle, fixturesPath, benchmarkPath, scorerPath, bench); err == nil {
		t.Fatal("candidate runtime accepted one rebuilt binary for both roots")
	}

	baselineRef := "refs/c3-eval/commit-pool/protocol-v7-baseline-test"
	runGitTest(t, base, "update-ref", baselineRef, baselineCommit)
	baselineBundle := filepath.Join(t.TempDir(), "baseline.bundle")
	runGitTest(t, base, "bundle", "create", baselineBundle, baselineRef)
	policyBytes := []byte(`{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private ` + `sentinel"]}`)
	policyPath := filepath.Join(t.TempDir(), "privacy-policy.json")
	if err := os.WriteFile(policyPath, policyBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes))
	if err != nil {
		t.Fatal(err)
	}
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		t.Fatal(err)
	}
	goHash, err := goExecutableSHA256()
	if err != nil {
		t.Fatal(err)
	}
	goVerifyHash, err := goModVerifySHA256(base)
	if err != nil {
		t.Fatal(err)
	}
	baselineExpected := authority.Expected
	baselineExpected.Commit = baselineCommit
	baselineExpected.Tree = baselineTree
	baselineExpected.ControllerCommit = baselineCommit
	baselineExpected.ControllerTree = baselineTree
	baselineExpected.SourceCapsuleSHA256 = canonicalSHA256(controllerCapsule)
	baselineExpected.ControllerSourceCapsuleSHA256 = canonicalSHA256(controllerCapsule)
	baselineExpected.DiffSHA256 = shaString("")
	baselineExpected.RuntimeSHA256 = controllerHash
	baselineExpected.CandidateDeltaSHA256 = ""
	baselineExpected.BundleSHA256 = mustFileHash(t, baselineBundle)
	baselineExpected.SourceBundleHeadsSHA256 = shaBytes([]byte(runGitTest(t, base, "bundle", "list-heads", baselineBundle)))
	baselineExpected.ProtocolTestSHA256 = mustFileHash(t, filepath.Join(root, "cli/tools/structural-search-eval-v2/main_test.go"))
	baselineExpected.PrivacyPolicySHA256 = policy.SHA256
	baselineExpected.PrivacyTermCount = len(policy.DenyTerms)
	baselineExpected.PrivacyDetectorSHA256 = scanner.detector.DefinitionSHA256
	baselineExpected.GoExecutableSHA256 = goHash
	baselineExpected.GoModVerifySHA256 = goVerifyHash
	baselineExpected.ScanCapsSHA256 = canonicalSHA256(protocolV7ScanCaps())
	v4 := controllerAuthorityV4{
		Schema: controllerAuthorityV4Schema, Mode: "baseline", Expected: baselineExpected,
		ControllerSourceCapsule: controllerCapsule, RuntimeSourceCapsule: controllerCapsule,
		BuildReplay: buildReplayManifest{
			ControllerSHA256: controllerHash, RuntimeSHA256: controllerHash,
			RebuiltControllerSHA256: controllerHash, RebuiltRuntimeSHA256: controllerHash,
			ControllerCapsuleRebuildVerified: true, RuntimeCapsuleRebuildVerified: true, BundleVerified: true,
		},
		BudgetLimits: limits, ScanCaps: protocolV7ScanCaps(), PrivacyPolicySHA256: policy.SHA256,
		PrivacyTermCount: len(policy.DenyTerms), PrivacyDetectorSHA256: scanner.detector.DefinitionSHA256,
		GoExecutableSHA256: goHash, GoModVerifySHA256: goVerifyHash,
		SourceBundleHeadsSHA256: baselineExpected.SourceBundleHeadsSHA256,
		ProtocolTestSHA256:      baselineExpected.ProtocolTestSHA256,
		ActionEnvelope:          actionEnvelope, CanonicalRowBytesDefinition: canonicalRowBytesDefinition,
		ContextThresholdAuthorityRecord: string(line),
	}
	v4Bytes, err := json.Marshal(v4)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyControllerAuthorityV4(v4, v4Bytes, controllerBinary, controllerBinary, base, base, baselineBundle, policyPath, fixturesPath, benchmarkPath, scorerPath, bench, policy, scanner, parentBaselineFiles{}); err != nil {
		t.Fatalf("%v; clean scans before failure=%#v", err, scanner.entries)
	}
}

func TestProtocolV7PrivacyPolicyIsCanonicalBoundedAndExact(t *testing.T) {
	raw := `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private target","target.example"]}`
	policy, err := decodePrivacyPolicy(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if policy.SHA256 != shaBytes([]byte(raw)) || !reflect.DeepEqual(policy.DenyTerms, []string{"private target", "target.example"}) {
		t.Fatalf("decoded privacy policy is not byte-bound: %#v", policy)
	}
	for name, bad := range map[string]string{
		"terminal LF": raw + "\n",
		"uppercase":   `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["Private target"]}`,
		"unsorted":    `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["target.example","private target"]}`,
		"duplicate":   `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private target","private target"]}`,
		"short":       `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["abc"]}`,
		"whitespace":  `{ "$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private target"]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodePrivacyPolicy(strings.NewReader(bad)); err == nil {
				t.Fatal("non-canonical privacy policy was accepted")
			}
		})
	}
}

func TestProtocolV7GenericPrivacyDetectorVectorsAndClaimBoundary(t *testing.T) {
	detector, err := newGenericPrivacyDetector()
	if err != nil {
		t.Fatal(err)
	}
	if detector.Version != "structural-retrieval-generic-privacy-detector.v1" || !validSHA256(detector.DefinitionSHA256) {
		t.Fatalf("detector identity is not frozen: %#v", detector)
	}
	positives := []string{
		"/ho" + "me/example/work",
		"/ro" + "ot/private",
		`C:\` + `Users\Example\repo`,
		"-----BEGIN PRIVATE " + "KEY-----",
		"gh" + "p_abcdefghijklmnopqrstuvwxyz123456",
		"github" + "_pat_abcdefghijklmnopqrstuvwxyz",
		"s" + "k-ant-abcdefghijklmnopqrstuvwxyz1234",
		"AK" + "IAABCDEFGHIJKLMNOP",
		"xo" + "xb-1234567890-abcdef",
		"Authorization: " + "Bearer abcdefghijklmnop",
		"api_" + "key=abcdefghijklmnop",
		`token = "abcdefghijklmnop";`,
		`password: "abcdefghijklmnop"`,
		`'secret': 'abcdefghijklmnop'`,
		`"api_key": "abcdefghijklmnop",`,
		`  "token": "abc-def-12345678"  `,
		`"secret": "密密密密密密密密"`,
	}
	for _, input := range positives {
		if detector.Match([]byte(input)) == "" {
			t.Fatalf("positive privacy vector did not match: %q", input)
		}
	}
	negatives := []string{
		"/template/path", "/homework/example", `C:\UserGuide\repo`, "-----BEGIN PUBLIC KEY-----",
		"gh" + "p_example", "sk-short", "AKIAEXAMPLE", "Authorization: Basic abcdefghijklmnop",
		"token_count=12", "api_key_name=generic",
		"token = strings.TrimPrefix(token, prefix)", "token = normalizeProjectPath(token)",
		"secret = config.Value", "password = make([]byte, 32)", "secret=config/value",
		"token=config.Value", "prefix api_key=abcdefghijklmnop",
		`"secret': "abcdefghijklmnop"`, `'secret": 'abcdefghijklmnop'`, `"secret": "abc\"defghijklmnop"`,
	}
	for _, input := range negatives {
		if got := detector.Match([]byte(input)); got != "" {
			t.Fatalf("negative privacy vector matched %q: %s", input, got)
		}
	}
}

func TestProtocolV7FrozenBuildInputsAreGenericPrivacyClean(t *testing.T) {
	detector, err := newGenericPrivacyDetector()
	if err != nil {
		t.Fatal(err)
	}
	root := repoRoot(t)
	paths, err := discoverRepositoryBuildInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths,
		"research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
		"research/eval/structural-retrieval/benchmark.v2.json",
		"cli/tools/structural-search-eval-v2/main_test.go",
	)
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if match := detector.Match(data); match != "" {
			t.Fatalf("frozen generic input %s matched detector %s", path, match)
		}
	}
}

func TestProtocolV7PrivacyScannerUsesEphemeralTermsWithoutLeakingMatches(t *testing.T) {
	raw := `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private target"]}`
	policy, err := decodePrivacyPolicy(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		t.Fatal(err)
	}
	if err := scanner.Scan("result", "results/clean.json", []byte(`{"value":"generic"}`)); err != nil {
		t.Fatal(err)
	}
	err = scanner.Scan("report", "reports/late.json", []byte(`{"value":"PRIVATE TARGET"}`))
	if err == nil || strings.Contains(strings.ToLower(err.Error()), "private target") || strings.Contains(err.Error(), "reports/late.json") {
		t.Fatalf("privacy failure was absent or leaked match details: %v", err)
	}
}

func TestProtocolV7GovernedValidatorPrefixReplaysExactHashChain(t *testing.T) {
	payload := map[string]any{
		"event": "finish", "worker_id": "validator-baseline-protocol-v7", "role": "independent baseline validator",
		"status": "accepted", "effect_claim": false,
	}
	payloadHash := canonicalSHA256(payload)
	record := map[string]any{
		"seq": 1, "recorded_at": "2026-07-17T16:00:00Z", "prev_hash": "GENESIS",
		"payload_sha256": payloadHash, "payload": payload,
	}
	recordHash := canonicalSHA256(record)
	record["record_hash"] = recordHash
	line, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	got, err := verifyWorkerRecordPrefix(bytes.NewReader(append(line, '\n')), 1, recordHash, payloadHash)
	if err != nil {
		t.Fatal(err)
	}
	if canonicalSHA256(got) != payloadHash {
		t.Fatalf("validator payload changed: %#v", got)
	}
	record["prev_hash"] = strings.Repeat("0", 64)
	tampered, _ := json.Marshal(record)
	if _, err := verifyWorkerRecordPrefix(bytes.NewReader(append(tampered, '\n')), 1, recordHash, payloadHash); err == nil {
		t.Fatal("tampered validator chain was accepted")
	}
}

func TestProtocolV7StrictJSONRejectsDuplicateKeysAtEveryDepth(t *testing.T) {
	for _, input := range []string{
		`{"$schema":"one","$schema":"two"}`,
		`{"outer":{"value":1,"value":2}}`,
		`[{"value":1,"value":2}]`,
	} {
		var decoded any
		if err := decodeStrictBytes([]byte(input), &decoded); err == nil {
			t.Fatalf("duplicate JSON key was accepted: %s", input)
		}
	}
}

func TestProtocolV7ResolvedPathsRejectSymlinkEscapeAndRootAlias(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte("generic\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "inside.json")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if pathWithin(root, link) {
		t.Fatal("symlink escaping the frozen root was accepted")
	}
	alias := filepath.Join(t.TempDir(), "root-alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	if !pathsOverlap(root, alias) {
		t.Fatal("symlink aliases were treated as disjoint roots")
	}
}

func TestProtocolV7ScansFullHistoricalRepositoryPaths(t *testing.T) {
	repo := t.TempDir()
	runGitTest(t, repo, "init", "-q")
	runGitTest(t, repo, "config", "user.name", "C3 Eval")
	runGitTest(t, repo, "config", "user.email", "c3-eval@invalid")
	nested := filepath.Join(repo, "client", "private-area", "record.txt")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("generic\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "add", ".")
	runGitTest(t, repo, "commit", "-q", "-m", "generic history")
	commit := strings.TrimSpace(runGitTest(t, repo, "rev-parse", "HEAD"))
	artifacts, contents, err := fullReachablePathArtifacts(repo, commit, protocolV7ScanCaps())
	if err != nil {
		t.Fatal(err)
	}
	policy, err := decodePrivacyPolicy(strings.NewReader(`{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["client/` + `private-area"]}`))
	if err != nil {
		t.Fatal(err)
	}
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		t.Fatal(err)
	}
	matched := false
	for path, artifact := range artifacts {
		if err := scanner.Scan(artifact.Role, path, contents[path]); err != nil {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatal("deny term spanning repository path components was not detected")
	}
}

func TestProtocolV7HermeticGoEnvironmentDropsAmbientOverlays(t *testing.T) {
	for key, value := range map[string]string{
		"GOENV": "/host/goenv", "GOWORK": "/host/go.work", "GOFLAGS": "-overlay=/host/overlay.json",
		"GOTOOLCHAIN": "hostile", "GOPROXY": "https://host.invalid", "GOPRIVATE": "private.invalid", "CC": "/host/cc",
	} {
		t.Setenv(key, value)
	}
	env := hermeticGoEnvironment("/generic/home", "/generic/modcache", "/generic/gocache")
	joined := "\n" + strings.Join(env, "\n") + "\n"
	for _, exact := range []string{"GOENV=off", "GOWORK=off", "GOFLAGS=", "GOTOOLCHAIN=local", "CGO_ENABLED=0", "GOPROXY=off", "GOSUMDB=off"} {
		if !strings.Contains(joined, "\n"+exact+"\n") {
			t.Fatalf("hermetic Go environment lacks %q: %v", exact, env)
		}
	}
	for _, forbidden := range []string{"/host/goenv", "/host/go.work", "/host/overlay.json", "hostile", "host.invalid", "private.invalid", "/host/cc"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("ambient Go configuration survived: %q in %v", forbidden, env)
		}
	}
}

func TestProtocolV7EnvironmentIdentityNormalizesOnlyVolatilePaths(t *testing.T) {
	goEnv := protocolV7TestGoEnvironment()
	allowed := protocolV7TestAllowedEnvironment()
	goEnvJSON := protocolV7MarshalGoEnvironment(t, goEnv)
	want, err := protocolEnvironmentIdentitySHA256(goEnvJSON, strings.Repeat("a", 64), allowed)
	if err != nil {
		t.Fatal(err)
	}
	legacy := canonicalSHA256(map[string]any{"go_env": string(goEnvJSON), "go_executable_sha256": strings.Repeat("a", 64), "allowed_environment": allowed})

	goEnv["GOROOT"] = "/another/go-root"
	goEnv["GOMODCACHE"] = "/another/module-cache"
	allowed = protocolV7ReplaceEnvironmentValues(t, allowed, map[string]string{
		"PATH":       "/another/volatile/path",
		"HOME":       "/another/private/home",
		"GOMODCACHE": goEnv["GOMODCACHE"],
		"GOCACHE":    "/another/build-cache",
	})
	changedGoEnvJSON := protocolV7MarshalGoEnvironment(t, goEnv)
	got, err := protocolEnvironmentIdentitySHA256(changedGoEnvJSON, strings.Repeat("a", 64), allowed)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("volatile path spellings changed environment identity: got %s want %s", got, want)
	}
	changedLegacy := canonicalSHA256(map[string]any{"go_env": string(changedGoEnvJSON), "go_executable_sha256": strings.Repeat("a", 64), "allowed_environment": allowed})
	if changedLegacy == legacy {
		t.Fatal("path-only changes unexpectedly retained the same legacy identity")
	}
}

func TestProtocolV7EnvironmentIdentityRetainsStableGoEnvironment(t *testing.T) {
	stableGoEnvironment := []string{"GOOS", "GOARCH", "GOVERSION", "CGO_ENABLED", "GOENV", "GOWORK", "GOFLAGS", "GOTOOLCHAIN"}
	for _, key := range stableGoEnvironment {
		t.Run(key, func(t *testing.T) {
			baseGoEnv := protocolV7TestGoEnvironment()
			baseAllowed := protocolV7TestAllowedEnvironment()
			base, err := protocolEnvironmentIdentitySHA256(protocolV7MarshalGoEnvironment(t, baseGoEnv), strings.Repeat("a", 64), baseAllowed)
			if err != nil {
				t.Fatal(err)
			}
			baseGoEnv[key] += "-changed"
			if protocolV7AllowedEnvironmentHasKey(baseAllowed, key) {
				baseAllowed = protocolV7ReplaceEnvironmentValues(t, baseAllowed, map[string]string{key: baseGoEnv[key]})
			}
			changed, err := protocolEnvironmentIdentitySHA256(protocolV7MarshalGoEnvironment(t, baseGoEnv), strings.Repeat("a", 64), baseAllowed)
			if err != nil {
				t.Fatal(err)
			}
			if changed == base {
				t.Fatalf("stable Go environment field %s did not change identity", key)
			}
		})
	}
}

func TestProtocolV7EnvironmentIdentityRetainsEveryNonPathAllowedValue(t *testing.T) {
	nonPathAllowed := []string{
		"LC_ALL", "LANG", "TZ", "GOENV", "GOWORK", "GOFLAGS", "GOTOOLCHAIN", "CGO_ENABLED",
		"GOPROXY", "GOSUMDB", "GONOSUMDB", "GOPRIVATE",
	}
	for _, key := range nonPathAllowed {
		t.Run(key, func(t *testing.T) {
			baseGoEnv := protocolV7TestGoEnvironment()
			baseAllowed := protocolV7TestAllowedEnvironment()
			base, err := protocolEnvironmentIdentitySHA256(protocolV7MarshalGoEnvironment(t, baseGoEnv), strings.Repeat("a", 64), baseAllowed)
			if err != nil {
				t.Fatal(err)
			}
			values := map[string]string{key: "changed-" + key}
			if _, duplicatedByGoEnv := baseGoEnv[key]; duplicatedByGoEnv {
				baseGoEnv[key] = values[key]
			}
			baseAllowed = protocolV7ReplaceEnvironmentValues(t, baseAllowed, values)
			changed, err := protocolEnvironmentIdentitySHA256(protocolV7MarshalGoEnvironment(t, baseGoEnv), strings.Repeat("a", 64), baseAllowed)
			if err != nil {
				t.Fatal(err)
			}
			if changed == base {
				t.Fatalf("allowed environment field %s did not change identity", key)
			}
		})
	}
}

func TestProtocolV7EnvironmentIdentityRetainsGoExecutableHash(t *testing.T) {
	goEnv := protocolV7MarshalGoEnvironment(t, protocolV7TestGoEnvironment())
	allowed := protocolV7TestAllowedEnvironment()
	first, err := protocolEnvironmentIdentitySHA256(goEnv, strings.Repeat("a", 64), allowed)
	if err != nil {
		t.Fatal(err)
	}
	second, err := protocolEnvironmentIdentitySHA256(goEnv, strings.Repeat("b", 64), allowed)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("Go executable hash did not change environment identity")
	}
}

func TestProtocolV7EnvironmentIdentityFailsClosed(t *testing.T) {
	validGoEnv := protocolV7MarshalGoEnvironment(t, protocolV7TestGoEnvironment())
	validAllowed := protocolV7TestAllowedEnvironment()
	validHash := strings.Repeat("a", 64)

	missingGoEnv := protocolV7TestGoEnvironment()
	delete(missingGoEnv, "GOOS")
	extraGoEnv := protocolV7TestGoEnvironment()
	extraGoEnv["UNEXPECTED"] = "value"
	emptyGoEnv := protocolV7TestGoEnvironment()
	emptyGoEnv["GOOS"] = ""

	cases := map[string]struct {
		goEnv   []byte
		goHash  string
		allowed []string
	}{
		"missing_go_environment_field": {goEnv: protocolV7MarshalGoEnvironment(t, missingGoEnv), goHash: validHash, allowed: validAllowed},
		"extra_go_environment_field":   {goEnv: protocolV7MarshalGoEnvironment(t, extraGoEnv), goHash: validHash, allowed: validAllowed},
		"empty_required_go_field":      {goEnv: protocolV7MarshalGoEnvironment(t, emptyGoEnv), goHash: validHash, allowed: validAllowed},
		"duplicate_go_environment_key": {goEnv: []byte(`{"GOOS":"linux","GOOS":"darwin","GOARCH":"amd64","GOVERSION":"go1.test","GOROOT":"/go","CGO_ENABLED":"0","GOENV":"off","GOWORK":"off","GOFLAGS":"","GOTOOLCHAIN":"local","GOMODCACHE":"/mod"}`), goHash: validHash, allowed: validAllowed},
		"malformed_go_environment":     {goEnv: []byte(`{"GOOS":1}`), goHash: validHash, allowed: validAllowed},
		"invalid_go_executable_hash":   {goEnv: validGoEnv, goHash: "not-a-hash", allowed: validAllowed},
		"missing_allowed_key":          {goEnv: validGoEnv, goHash: validHash, allowed: validAllowed[1:]},
		"extra_allowed_key":            {goEnv: validGoEnv, goHash: validHash, allowed: append(append([]string(nil), validAllowed...), "EXTRA=value")},
		"duplicate_allowed_key":        {goEnv: validGoEnv, goHash: validHash, allowed: append(append([]string(nil), validAllowed...), validAllowed[0])},
		"malformed_allowed_key":        {goEnv: validGoEnv, goHash: validHash, allowed: append(append([]string(nil), validAllowed[:1]...), append([]string{"malformed"}, validAllowed[2:]...)...)},
		"empty_required_allowed_value": {goEnv: validGoEnv, goHash: validHash, allowed: protocolV7ReplaceEnvironmentValues(t, validAllowed, map[string]string{"PATH": ""})},
		"inconsistent_duplicate_value": {goEnv: validGoEnv, goHash: validHash, allowed: protocolV7ReplaceEnvironmentValues(t, validAllowed, map[string]string{"GOENV": "host-goenv"})},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := protocolEnvironmentIdentitySHA256(tc.goEnv, tc.goHash, tc.allowed); err == nil {
				t.Fatal("invalid environment identity input was accepted")
			}
		})
	}
}

func protocolV7TestGoEnvironment() map[string]string {
	return map[string]string{
		"GOOS": "linux", "GOARCH": "amd64", "GOVERSION": "go1.test", "GOROOT": "/go-root",
		"CGO_ENABLED": "0", "GOENV": "", "GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local",
		"GOMODCACHE": "/module-cache",
	}
}

func protocolV7TestAllowedEnvironment() []string {
	return []string{
		"PATH=/volatile/path", "HOME=/private/home", "LC_ALL=C", "LANG=C", "TZ=UTC", "GOENV=off", "GOWORK=off",
		"GOFLAGS=", "GOTOOLCHAIN=local", "CGO_ENABLED=0", "GOPROXY=off", "GOSUMDB=off", "GONOSUMDB=",
		"GOPRIVATE=", "GOMODCACHE=/module-cache", "GOCACHE=/build-cache",
	}
}

func protocolV7MarshalGoEnvironment(t *testing.T, environment map[string]string) []byte {
	t.Helper()
	data, err := json.Marshal(environment)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func protocolV7AllowedEnvironmentHasKey(environment []string, want string) bool {
	for _, assignment := range environment {
		key, _, ok := strings.Cut(assignment, "=")
		if ok && key == want {
			return true
		}
	}
	return false
}

func protocolV7ReplaceEnvironmentValues(t *testing.T, environment []string, replacements map[string]string) []string {
	t.Helper()
	result := append([]string(nil), environment...)
	seen := make(map[string]bool, len(replacements))
	for index, assignment := range result {
		key, _, ok := strings.Cut(assignment, "=")
		if !ok {
			t.Fatalf("malformed test environment assignment %q", assignment)
		}
		if value, replace := replacements[key]; replace {
			result[index] = key + "=" + value
			seen[key] = true
		}
	}
	for key := range replacements {
		if !seen[key] {
			t.Fatalf("test environment does not contain %q", key)
		}
	}
	return result
}

func TestProtocolV7PreflightDeadlineKillsCommandProcessGroup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := withProtocolCommandContext(ctx, func() error {
		return protocolCommand("sh", "-c", "sleep 30 & wait").Run()
	})
	if err == nil || !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("deadline did not terminate the command group: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("deadline termination was too slow: %s", elapsed)
	}
}

func TestProtocolV7NormalControllerRejectsLegacyAuthorityBeforeEffects(t *testing.T) {
	root := t.TempDir()
	authorityPath := filepath.Join(root, "authority.json")
	if err := os.WriteFile(authorityPath, []byte(`{"$schema":"structural-retrieval-controller-authority.v3","mode":"baseline"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	workRoot := filepath.Join(root, "work")
	outputRoot := filepath.Join(root, "output")
	var stdout bytes.Buffer
	err := runControllerCLI([]string{
		"--runtime", filepath.Join(root, "runtime"), "--fixtures", filepath.Join(root, "fixtures.jsonl"),
		"--benchmark", filepath.Join(root, "benchmark.json"), "--work-root", workRoot, "--authority", authorityPath,
		"--controller-source-root", filepath.Join(root, "B"), "--runtime-source-root", filepath.Join(root, "C"),
		"--bundle", filepath.Join(root, "source.bundle"), "--scorer-source", filepath.Join(root, "main.go"), "--output-dir", outputRoot,
	}, &stdout)
	if err == nil || !strings.Contains(err.Error(), "protocol-v7 required") {
		t.Fatalf("legacy authority was not operationally disabled: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("legacy rejection wrote stdout: %q", stdout.String())
	}
	for _, path := range []string{workRoot, outputRoot} {
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("legacy rejection created %s: %v", path, statErr)
		}
	}
}

func TestProtocolV7PrivacyTransactionPublishesNothingAfterLateHit(t *testing.T) {
	raw := `{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private target"]}`
	policy, err := decodePrivacyPolicy(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		t.Fatal(err)
	}
	outputRoot := filepath.Join(t.TempDir(), "must-not-exist")
	tx := newPrivacyTransaction(outputRoot, scanner)
	if err := tx.Add("result", "results/first.json", []byte(`{"value":"generic"}`)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Add("report", "reports/late.json", []byte(`{"value":"private target"}`)); err == nil {
		t.Fatal("late privacy hit was accepted")
	}
	var stdout bytes.Buffer
	if err := tx.Publish(&stdout, []byte(`{"$schema":"structural-retrieval-controller-output.v4"}`)); err == nil {
		t.Fatal("privacy-tainted transaction published")
	}
	if stdout.Len() != 0 {
		t.Fatalf("privacy-tainted transaction wrote stdout: %q", stdout.String())
	}
	if _, err := os.Stat(outputRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("privacy-tainted transaction created output: %v", err)
	}
}

type protocolV7ShortWriter struct{}

func (protocolV7ShortWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	return len(data) - 1, nil
}

func TestProtocolV7ShortStdoutWriteRollsBackPublishedTree(t *testing.T) {
	policy, err := decodePrivacyPolicy(strings.NewReader(`{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private target"]}`))
	if err != nil {
		t.Fatal(err)
	}
	scanner, err := newPrivacyScanner(policy)
	if err != nil {
		t.Fatal(err)
	}
	outputRoot := filepath.Join(t.TempDir(), "must-be-rolled-back")
	tx := newPrivacyTransaction(outputRoot, scanner)
	if err := tx.Add("result", "results/one.json", []byte(`{"value":"generic"}`)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Publish(protocolV7ShortWriter{}, []byte(`{"$schema":"structural-retrieval-controller-output.v4"}`)); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short stdout write was not rejected: %v", err)
	}
	if _, err := os.Stat(outputRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("short stdout write left a published tree: %v", err)
	}
}

func TestProtocolV7RealBaselineControllerPublishesOnePrivacyCleanTransaction(t *testing.T) {
	skipUnlessReleaseFixtures(t, filepath.Join(repoRoot(t), ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"))
	root := repoRoot(t)
	base := filepath.Join(t.TempDir(), "B")
	paths, err := discoverRepositoryBuildInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths,
		"research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
		"research/eval/structural-retrieval/benchmark.v2.json",
		"cli/tools/structural-search-eval-v2/main_test.go",
	)
	for _, relative := range paths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(base, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGitTest(t, base, "init", "-q")
	runGitTest(t, base, "config", "user.name", "C3 Eval")
	runGitTest(t, base, "config", "user.email", "c3-eval@invalid")
	runGitTest(t, base, "add", ".")
	runGitTest(t, base, "commit", "-q", "-m", "portable protocol-v7 baseline")
	commit := strings.TrimSpace(runGitTest(t, base, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(runGitTest(t, base, "rev-parse", "HEAD^{tree}"))
	ref := "refs/c3-eval/commit-pool/protocol-v7-live-baseline"
	runGitTest(t, base, "update-ref", ref, commit)
	bundle := filepath.Join(t.TempDir(), "baseline.bundle")
	runGitTest(t, base, "bundle", "create", bundle, ref)
	runtimePath := filepath.Join(t.TempDir(), "controller-runtime")
	if err := buildFrozenRuntime(base, runtimePath); err != nil {
		t.Fatal(err)
	}
	runtimeHash := mustFileHash(t, runtimePath)
	capsule, err := captureSourceCapsule(base)
	if err != nil {
		t.Fatal(err)
	}
	fixturesPath := filepath.Join(base, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(base, "research/eval/structural-retrieval/benchmark.v2.json")
	scorerPath := filepath.Join(base, "cli/tools/structural-search-eval-v2/main.go")
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil {
		t.Fatal(err)
	}
	policyBytes := []byte(`{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private ` + `sentinel"]}`)
	policyPath := filepath.Join(t.TempDir(), "privacy-policy.json")
	if err := os.WriteFile(policyPath, policyBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes))
	if err != nil {
		t.Fatal(err)
	}
	detector, err := newGenericPrivacyDetector()
	if err != nil {
		t.Fatal(err)
	}
	environmentHash, err := environmentSHA256(base)
	if err != nil {
		t.Fatal(err)
	}
	moduleHash, err := moduleGraphSHA256(base)
	if err != nil {
		t.Fatal(err)
	}
	goHash, err := goExecutableSHA256()
	if err != nil {
		t.Fatal(err)
	}
	goVerifyHash, err := goModVerifySHA256(base)
	if err != nil {
		t.Fatal(err)
	}
	limits := testControllerBudgetLimits()
	actionEnvelope := "protocol-v7 generic baseline transaction; no product writes"
	line := checkinLine(t, filepath.Join(root, ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"), 5)
	expected := provenance{
		ExperimentID: "protocol-v7-live-baseline", ArmID: "baseline", Commit: commit, Tree: tree,
		ControllerCommit: commit, ControllerTree: tree, SourceCapsuleSHA256: canonicalSHA256(capsule),
		ControllerSourceCapsuleSHA256: canonicalSHA256(capsule), DiffSHA256: shaString(""),
		FixtureSHA256: mustFileHash(t, fixturesPath), BenchmarkSHA256: mustFileHash(t, benchmarkPath), ScorerSHA256: mustFileHash(t, scorerPath),
		ControllerSHA256: runtimeHash, RuntimeSHA256: runtimeHash, EnvironmentSHA256: environmentHash, ModuleGraphSHA256: moduleHash,
		BudgetSHA256: canonicalSHA256(limits), ActionEnvelopeSHA256: shaString(actionEnvelope), SemanticMode: bench.SemanticMode,
		ContextThresholdAuthoritySHA256: hashThresholdAuthority(bench.ContextThresholdAuthority), BundleSHA256: mustFileHash(t, bundle),
		PrivacyPolicySHA256: policy.SHA256, PrivacyTermCount: len(policy.DenyTerms), PrivacyDetectorSHA256: detector.DefinitionSHA256,
		GoExecutableSHA256: goHash, GoModVerifySHA256: goVerifyHash, ScanCapsSHA256: canonicalSHA256(protocolV7ScanCaps()),
		SourceBundleHeadsSHA256: shaBytes([]byte(runGitTest(t, base, "bundle", "list-heads", bundle))),
		ProtocolTestSHA256:      mustFileHash(t, filepath.Join(root, "cli/tools/structural-search-eval-v2/main_test.go")),
	}
	authority := controllerAuthorityV4{
		Schema: controllerAuthorityV4Schema, Mode: "baseline", Expected: expected,
		ControllerSourceCapsule: capsule, RuntimeSourceCapsule: capsule,
		BuildReplay: buildReplayManifest{
			ControllerSHA256: runtimeHash, RuntimeSHA256: runtimeHash, RebuiltControllerSHA256: runtimeHash, RebuiltRuntimeSHA256: runtimeHash,
			ControllerCapsuleRebuildVerified: true, RuntimeCapsuleRebuildVerified: true, BundleVerified: true,
		},
		BudgetLimits: limits, ScanCaps: protocolV7ScanCaps(), PrivacyPolicySHA256: policy.SHA256, PrivacyTermCount: len(policy.DenyTerms),
		PrivacyDetectorSHA256: detector.DefinitionSHA256, GoExecutableSHA256: goHash, GoModVerifySHA256: goVerifyHash,
		SourceBundleHeadsSHA256: expected.SourceBundleHeadsSHA256, ActionEnvelope: actionEnvelope,
		ProtocolTestSHA256:          expected.ProtocolTestSHA256,
		CanonicalRowBytesDefinition: canonicalRowBytesDefinition, ContextThresholdAuthorityRecord: string(line),
	}
	authorityBytes, err := json.Marshal(authority)
	if err != nil {
		t.Fatal(err)
	}
	authorityPath := filepath.Join(t.TempDir(), "authority.json")
	if err := os.WriteFile(authorityPath, authorityBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	workRoot := filepath.Join(t.TempDir(), "ephemeral-work-must-remain-absent")
	outputRoot := filepath.Join(t.TempDir(), "output")
	command := exec.Command(runtimePath, "--controller", "--runtime", runtimePath, "--fixtures", fixturesPath, "--benchmark", benchmarkPath,
		"--work-root", workRoot, "--authority", authorityPath, "--controller-source-root", base, "--runtime-source-root", base,
		"--bundle", bundle, "--privacy-policy", policyPath, "--scorer-source", scorerPath, "--output-dir", outputRoot)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("protocol-v7 controller failed: %v: %s", err, stderr.Bytes())
	}
	var output controllerOutput
	if err := decodeStrictBytes(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	fixtures, _, err := loadFixtures(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	if output.Schema != "structural-retrieval-controller-output.v4" || len(output.Runs) != len(fixtures)+2 || !validSHA256(output.PrivacyManifestSHA256) || !validSHA256(output.OrderedRunManifestSHA256) {
		t.Fatalf("protocol-v7 output is incomplete: %#v", output)
	}
	if _, err := os.Stat(filepath.Join(outputRoot, output.PrivacyManifestPath)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(workRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("protocol-v7 retained caller work root: %v", err)
	}

	parentAuthorityPath := filepath.Join(outputRoot, "controller-authority.v4.json")
	parentOutputPath := filepath.Join(outputRoot, "controller-output.v4.json")
	if err := os.WriteFile(parentAuthorityPath, authorityBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(parentOutputPath, stdout.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	history, err := readHistory(filepath.Join(outputRoot, output.HistoryPath))
	if err != nil {
		t.Fatal(err)
	}
	binding := parentBaselineBinding{
		AuthoritySHA256: mustFileHash(t, parentAuthorityPath), OutputSHA256: mustFileHash(t, parentOutputPath),
		OrderedRunSHA256: output.OrderedRunManifestSHA256, RunCount: len(output.Runs), HistorySHA256: output.HistorySHA256,
		HistoryTailRecordHash: history[len(history)-1].RecordHash, PrivacyManifestSHA256: output.PrivacyManifestSHA256,
		ValidatorRecordRef: "workers/validator-baseline-protocol-v7/progress.jsonl#seq=1",
	}
	acceptance := baselineAcceptance{
		Schema: "structural-retrieval-baseline-acceptance.v1", Verdict: "accepted",
		AuthoritySHA256: binding.AuthoritySHA256, OutputSHA256: binding.OutputSHA256,
		OrderedRunManifestSHA256: binding.OrderedRunSHA256, RunCount: binding.RunCount,
		HistorySHA256: binding.HistorySHA256, HistoryTailRecordHash: binding.HistoryTailRecordHash,
		PrivacyManifestSHA256:     binding.PrivacyManifestSHA256,
		ValidatedSourceMainSHA256: mustFileHash(t, scorerPath),
		ValidatedSourceTestSHA256: authority.ProtocolTestSHA256,
	}
	acceptanceBytes, err := json.Marshal(acceptance)
	if err != nil {
		t.Fatal(err)
	}
	var acceptanceMap map[string]any
	if err := json.Unmarshal(acceptanceBytes, &acceptanceMap); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{
		"event": "finish", "worker_id": "validator-baseline-protocol-v7", "role": "independent baseline validator",
		"status": "accepted", "effect_claim": false, "baseline_acceptance": acceptanceMap,
	}
	payloadHash := canonicalSHA256(payload)
	record := map[string]any{
		"seq": 1, "recorded_at": "2026-07-17T17:00:00Z", "prev_hash": "GENESIS",
		"payload_sha256": payloadHash, "payload": payload,
	}
	recordHash := canonicalSHA256(record)
	record["record_hash"] = recordHash
	binding.ValidatorPayloadSHA256 = payloadHash
	binding.ValidatorRecordHash = recordHash
	validatorStore := t.TempDir()
	validatorPath := filepath.Join(validatorStore, "workers/validator-baseline-protocol-v7/progress.jsonl")
	if err := os.MkdirAll(filepath.Dir(validatorPath), 0o700); err != nil {
		t.Fatal(err)
	}
	recordBytes, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validatorPath, append(recordBytes, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	candidateAuthority := authority
	candidateAuthority.Mode = "candidate"
	candidateAuthority.CandidateDelta = &candidateDelta{BaselineCommit: commit, BaselineTree: tree}
	candidateAuthority.ParentBaseline = &binding
	candidateAuthority.Expected.ParentBaselineAuthoritySHA256 = binding.AuthoritySHA256
	candidateAuthority.Expected.ParentBaselineOutputSHA256 = binding.OutputSHA256
	candidateAuthority.Expected.ParentBaselineOrderedRunSHA256 = binding.OrderedRunSHA256
	candidateAuthority.Expected.ParentBaselineRunCount = binding.RunCount
	candidateAuthority.Expected.ParentBaselineHistorySHA256 = binding.HistorySHA256
	candidateAuthority.Expected.ParentBaselineHistoryTailSHA256 = binding.HistoryTailRecordHash
	candidateAuthority.Expected.ParentBaselinePrivacySHA256 = binding.PrivacyManifestSHA256
	candidateAuthority.Expected.ParentBaselineValidatorHash = binding.ValidatorRecordHash
	candidateAuthority.Expected.ParentBaselineValidatorPayload = binding.ValidatorPayloadSHA256
	parentScanner, err := newPrivacyScanner(policy)
	if err != nil {
		t.Fatal(err)
	}
	parentFiles := parentBaselineFiles{Root: outputRoot, Authority: parentAuthorityPath, Output: parentOutputPath, ValidatorStore: validatorStore}
	if err := verifyParentBaseline(candidateAuthority, parentFiles, base, base, fixturesPath, benchmarkPath, scorerPath, bench, parentScanner); err != nil {
		t.Fatal(err)
	}
	tampered := candidateAuthority
	tamperedBinding := *tampered.ParentBaseline
	tamperedBinding.OutputSHA256 = strings.Repeat("f", 64)
	tampered.ParentBaseline = &tamperedBinding
	if err := verifyParentBaseline(tampered, parentFiles, base, base, fixturesPath, benchmarkPath, scorerPath, bench, parentScanner); err == nil {
		t.Fatal("candidate accepted a tampered parent output binding")
	}
}

func runGitTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = append(os.Environ(), "LC_ALL=C", "TZ=UTC", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_NO_REPLACE_OBJECTS=1", "GIT_ALTERNATE_OBJECT_DIRECTORIES=")
	output, err := command.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			t.Fatalf("git %v: %v: %s", args, err, exit.Stderr)
		}
		t.Fatalf("git %v: %v", args, err)
	}
	return string(output)
}

func mustFileHash(t *testing.T, path string) string {
	t.Helper()
	hash, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func mustGitFileSHA256(t *testing.T, repo, commit, path string) string {
	t.Helper()
	data := []byte(runGitTest(t, repo, "show", commit+":"+path))
	return shaBytes(data)
}

func shaBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestFixtureAndBenchmarkStrictJSONAndFreezeHashes(t *testing.T) {
	root := repoRoot(t)
	fixturesPath := filepath.Join(root, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
	benchmarkPath := filepath.Join(root, "research/eval/structural-retrieval/benchmark.v2.json")
	fixtures, fixturesHash, err := loadFixtures(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	benchmark, err := loadBenchmark(benchmarkPath)
	if err != nil {
		t.Fatal(err)
	}
	if benchmark.FixtureSHA256 != fixturesHash || benchmark.FixtureCount != len(fixtures) {
		t.Fatalf("fixture freeze mismatch: benchmark=%#v count=%d hash=%s", benchmark, len(fixtures), fixturesHash)
	}
	bad := strings.Replace(string(mustRead(t, benchmarkPath)), "\n}", ",\n  \"unknown\": true\n}", 1)
	if _, err := decodeBenchmark(strings.NewReader(bad)); err == nil {
		t.Fatal("unknown benchmark field was accepted")
	}
}

func sampleFixture() fixtureCase {
	return fixtureCase{
		Schema: fixtureSchema, CaseID: "wrong-layer", Family: familyWrongLayer, Query: "sync ledger owner",
		Corpus: corpusInput{Entities: []entityInput{
			{ID: "c3-100", Type: "component", Title: "Sync Ledger Owner", Slug: "sync-ledger-owner", Goal: "Owns the synchronization ledger", Status: "active", Metadata: `{}`, Markdown: "# Sync Ledger Owner\n\n## Goal\n\nOwns the synchronization ledger."},
			{ID: "c3-101", Type: "component", Title: "Ledger Display", Slug: "ledger-display", Goal: "Displays ledger status", Status: "active", Metadata: `{}`, Markdown: "# Ledger Display\n\nSync ledger display."},
			{ID: "c3-102", Type: "component", Title: "Unrelated Queue", Slug: "unrelated-queue", Goal: "Queues reports", Status: "active", Metadata: `{}`, Markdown: "# Queue\n\nReports."},
		}},
		Oracle: oracleSpec{
			RequiredOwnerFactIDs: []string{"fact-owner"}, AllowedExtraFactIDs: []string{"fact-neutral"}, ForbiddenFactIDs: []string{"fact-forbidden"},
			FactBindings: map[string][]string{"c3-100": {"fact-owner"}, "c3-101": {"fact-forbidden"}, "c3-102": {"fact-neutral"}},
		},
	}
}

func sampleRouteFixture() fixtureCase {
	return fixtureCase{
		Schema: fixtureSchema, CaseID: "route-expansion", Family: familyRoute, Query: "notification relay",
		Corpus: corpusInput{
			Entities: []entityInput{
				{ID: "c3-200", Type: "component", Title: "Dispatch Coordinator", Slug: "dispatch-coordinator", Goal: "Coordinates outbound work", Status: "active", Metadata: `{}`, Markdown: "# Dispatch Coordinator\n\nCoordinates outbound work."},
				{ID: "c3-201", Type: "component", Title: "Notification Relay", Slug: "notification-relay", Goal: "Relays notification events", Status: "active", Metadata: `{}`, Markdown: "# Notification Relay\n\nNotification relay events."},
			},
			Relationships: []relationshipInput{{FromID: "c3-200", ToID: "c3-201", RelType: "uses"}},
		},
		Oracle: oracleSpec{
			RequiredOwnerFactIDs: []string{"fact-route-owner"}, AllowedExtraFactIDs: []string{"fact-anchor"}, ForbiddenFactIDs: []string{"fact-route-forbidden"},
			FactBindings:        map[string][]string{"c3-200": {"fact-route-owner"}, "c3-201": {"fact-anchor"}},
			RelationshipWitness: &relationshipWitness{ExpectedEntityID: "c3-200", FromID: "c3-200", ToID: "c3-201", RelType: "uses", ExpectedMatchSource: "graph:uses:c3-201", RequireDirectFTSMiss: true},
			RequiredRouteFields: []string{"facts", "graph", "lanes", "hash"},
		},
	}
}

func sampleBenchmark() benchmarkConfig {
	return benchmarkConfig{
		Schema: benchmarkSchema, FixtureCount: 1, K: 5, SemanticMode: semanticDisabled,
		Scale:                     scaleConfig{Seed: 17, Multiplier: 2, Tokens: []string{"amber", "cobalt", "fern", "harbor"}, MaxRelationshipDegree: 2},
		Thresholds:                thresholds{OwnerRecallAt5Delta: 0.2, StructuralOwnerPrecision: 0.8, CanonicalRowBytesRatio: 1.05},
		ContextThresholdAuthority: &thresholdAuthority{CheckinRef: "checkins.jsonl#seq=5", CheckinSHA256: "2e52b66573dc67b8950420c2e6e232a4899c1aaac3f915608f634bffdcd704e7", RecordHash: "66a927cbe7a5b78bf8641c662e952e3ed250ec7cdacae81749f8c3fde362ecac", DefinitionSHA256: "cfb85b044b6b4a3d5d694a9fe009ed4d597bc3006c44b02f59ed48a5e12e8429"},
	}
}

func validProvenance() provenance {
	h := strings.Repeat("a", 64)
	p := provenance{ExperimentID: "exp-1", ArmID: "B", Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40), SourceCapsuleSHA256: h, DiffSHA256: h, FixtureSHA256: h, ScorerSHA256: h, ControllerSHA256: h, RuntimeSHA256: h, LogicalDumpSHA256: h, EnvironmentSHA256: h, ModuleGraphSHA256: h, BudgetSHA256: h, ActionEnvelopeSHA256: h, CorpusMode: corpusCombined, SemanticMode: semanticDisabled, ProjectDirSHA256: h, C3DirSHA256: h}
	p.ContextThresholdAuthoritySHA256 = hashThresholdAuthority(sampleBenchmark().ContextThresholdAuthority)
	return p
}

func validHistoryRecord() historyRecord {
	h := strings.Repeat("a", 64)
	r := historyRecord{Schema: historySchema, ExperimentID: "exp-1", ArmID: "B", ParentKeep: "GENESIS", ChangedVariable: "baseline", ChangedPaths: []string{"baseline"}, ResultPath: "results/exp-1.json", ResultSHA256: h, Provenance: validProvenance(), Budgets: resourceBudget{WallTimeMillis: 1, CPUTimeMillis: 1, MaxRSSBytes: 1, ProcessCount: 1, SQLiteRowCount: 1, LogicalDumpBytes: 1, StdoutBytes: 1, StderrBytes: 1, CaseCount: 1}, Status: "keep", Reason: "baseline", Evidence: []string{"report"}, PrevHash: "GENESIS"}
	r.RecordHash = hashHistoryRecord(r)
	return r
}

func validSourceCapsule() sourceCapsule {
	return sourceCapsule{Schema: sourceCapsuleSchema, HeadCommit: strings.Repeat("a", 40), HeadTree: strings.Repeat("b", 40), RepositoryBuildInputCount: 1, Inputs: []sourceInput{{Path: "cli/cmd/search.go", SHA256: strings.Repeat("c", 64), Origin: "working-tree"}}, DirtyPatchSHA256: "def9ef26b435525e0ba8b9dcb704bc55ce1c51eb93f51e12ee372d05a72036af"}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", ".."))
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func checkinLine(t *testing.T, path string, seq int) []byte {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(mustRead(t, path)), []byte("\n"))
	if seq < 1 || seq > len(lines) {
		t.Fatalf("checkin seq %d is absent", seq)
	}
	return append(append([]byte(nil), lines[seq-1]...), '\n')
}

func sha(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestMainDeadlineIsBounded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 3*time.Second {
		t.Fatal("test controller has no bounded deadline")
	}
}

func TestArmWallExcludesControllerSideMonitorDrain(t *testing.T) {
	started := time.Unix(100, 0)
	waitedAt := started.Add(25 * time.Millisecond)
	joinedAt := waitedAt.Add(200 * time.Millisecond)
	accounting := cgroupAccountingResult{
		Final:       cgroupStats{CPUUsageMicros: 2_500, MemoryPeak: 4096, PIDsPeak: 3},
		Diagnostics: cgroupAccountingDiagnostics{SuccessfulCPUReads: 1},
	}
	actual, diagnostics, err := finalizeArmAccounting(started, waitedAt, joinedAt, accounting, 10, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if actual.WallTimeMillis != 25 {
		t.Fatalf("arm wall = %dms, want command.Wait boundary 25ms", actual.WallTimeMillis)
	}
	if diagnostics.MonitorJoinTailMillis != 200 {
		t.Fatalf("monitor join tail = %dms, want 200ms", diagnostics.MonitorJoinTailMillis)
	}
}

func TestReadCgroupStatsFromResolvedRootDoesNotExecSystemctl(t *testing.T) {
	reader := &recordingCgroupReader{files: map[string]string{
		"/resolved/cpu.stat":    "usage_usec 2500\n",
		"/resolved/memory.peak": "4096\n",
		"/resolved/pids.peak":   "3\n",
	}}
	stats, cpuRead, err := readCgroupStatsFromResolvedRoot(context.Background(), "/resolved", reader)
	if err != nil {
		t.Fatal(err)
	}
	if !cpuRead || stats.CPUUsageMicros != 2500 || stats.MemoryPeak != 4096 || stats.PIDsPeak != 3 {
		t.Fatalf("direct stats = %#v cpuRead=%v", stats, cpuRead)
	}
	if got, want := reader.paths, []string{"/resolved/cpu.stat", "/resolved/memory.peak", "/resolved/pids.peak"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("direct read paths = %v, want %v", got, want)
	}
}

func TestCgroupSamplerResolvesControlGroupOnce(t *testing.T) {
	resolver := &countingCgroupResolver{resolved: resolvedControlGroup{Root: "/resolved", Attempts: 2, Latency: 3 * time.Millisecond}}
	reader := &recordingCgroupReader{files: map[string]string{
		"/resolved/cpu.stat":    "usage_usec 1000\n",
		"/resolved/memory.peak": "2048\n",
		"/resolved/pids.peak":   "2\n",
	}}
	sampler, err := newCgroupSampler(context.Background(), "unit", resolver, reader, testCgroupSamplerConfig())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sampler.sampleOnce(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	if _, err := sampler.sampleOnce(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	if resolver.calls != 1 {
		t.Fatalf("ControlGroup resolve calls = %d, want 1", resolver.calls)
	}
}

func TestCPUActualIsFinalPostExitReadNotSampledPeak(t *testing.T) {
	started := time.Unix(100, 0)
	accounting := cgroupAccountingResult{
		PeriodicPeak: cgroupStats{CPUUsageMicros: 99_000, MemoryPeak: 2048, PIDsPeak: 2},
		Final:        cgroupStats{CPUUsageMicros: 2_500, MemoryPeak: 4096, PIDsPeak: 3},
		Diagnostics:  cgroupAccountingDiagnostics{SuccessfulCPUReads: 2},
	}
	actual, _, err := finalizeArmAccounting(started, started.Add(time.Millisecond), started.Add(2*time.Millisecond), accounting, 1, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if actual.CPUTimeMillis != 3 {
		t.Fatalf("CPU actual = %dms, want final post-exit 3ms", actual.CPUTimeMillis)
	}
}

func TestZeroSuccessfulCPUReadsIsInvalidNotZeroMillis(t *testing.T) {
	started := time.Unix(100, 0)
	_, _, err := finalizeArmAccounting(started, started.Add(time.Millisecond), started.Add(2*time.Millisecond), cgroupAccountingResult{
		Final:       cgroupStats{MemoryPeak: 4096, PIDsPeak: 2},
		Diagnostics: cgroupAccountingDiagnostics{SuccessfulCPUReads: 0},
	}, 1, 0, 1)
	if err == nil || !strings.Contains(err.Error(), "zero successful CPU reads") {
		t.Fatalf("missing CPU read was not invalid: %v", err)
	}
}

func TestCgroupSamplingHasBoundedPerSampleDeadline(t *testing.T) {
	reader := &deadlineCheckingCgroupReader{}
	config := testCgroupSamplerConfig()
	stats, cpuRead, elapsed, err := sampleCgroupStats(context.Background(), "/resolved", reader, config.SampleTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if !cpuRead || stats.CPUUsageMicros != 1000 || elapsed < 0 {
		t.Fatalf("sample = %#v cpuRead=%v elapsed=%v", stats, cpuRead, elapsed)
	}
	if reader.calls != 3 || reader.maxRemaining <= 0 || reader.maxRemaining > config.SampleTimeout {
		t.Fatalf("sample deadline calls=%d max_remaining=%v timeout=%v", reader.calls, reader.maxRemaining, config.SampleTimeout)
	}
}

func TestDurableReportRecordsWallAtWaitAndMonitorDrainSeparately(t *testing.T) {
	run := controllerRun{
		ActualBudget: resourceBudget{WallTimeMillis: 25, CPUTimeMillis: 3},
		AccountingDiagnostics: cgroupAccountingDiagnostics{
			SampleAttempts: 4, SuccessfulCPUReads: 4,
			CommandWaitWallMillis: 25, MonitorJoinTailMillis: 200,
		},
	}
	envelope := durableReport{Schema: durableReportSchema}
	populateDurableAccounting(&envelope, run)
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw["accounting_diagnostics"], []byte(`"command_wait_wall_millis":25`)) || !bytes.Contains(raw["accounting_diagnostics"], []byte(`"monitor_join_tail_millis":200`)) {
		t.Fatalf("durable diagnostics omit separate wall/tail: %s", data)
	}
	if bytes.Contains(raw["actual_budget"], []byte("monitor_join")) || bytes.Contains(raw["actual_budget"], []byte("sample_attempts")) {
		t.Fatalf("sampler diagnostics leaked into resourceBudget: %s", raw["actual_budget"])
	}
}

type recordingCgroupReader struct {
	files map[string]string
	paths []string
}

func (r *recordingCgroupReader) ReadFile(_ context.Context, path string) ([]byte, error) {
	r.paths = append(r.paths, path)
	value, ok := r.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(value), nil
}

type deadlineCheckingCgroupReader struct {
	calls        int
	maxRemaining time.Duration
}

func (r *deadlineCheckingCgroupReader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	r.calls++
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil, errors.New("cgroup read has no deadline")
	}
	remaining := time.Until(deadline)
	if remaining > r.maxRemaining {
		r.maxRemaining = remaining
	}
	switch filepath.Base(path) {
	case "cpu.stat":
		return []byte("usage_usec 1000\n"), nil
	case "memory.peak":
		return []byte("2048\n"), nil
	case "pids.peak":
		return []byte("2\n"), nil
	default:
		return nil, os.ErrNotExist
	}
}

type countingCgroupResolver struct {
	calls    int
	resolved resolvedControlGroup
	err      error
}

func (r *countingCgroupResolver) Resolve(context.Context, string) (resolvedControlGroup, error) {
	r.calls++
	return r.resolved, r.err
}

func testCgroupSamplerConfig() cgroupSamplerConfig {
	return cgroupSamplerConfig{
		ResolveDeadline:       2 * time.Second,
		ResolveAttemptTimeout: 250 * time.Millisecond,
		ResolveRetryInterval:  10 * time.Millisecond,
		SampleInterval:        10 * time.Millisecond,
		SampleTimeout:         50 * time.Millisecond,
		JoinTimeout:           100 * time.Millisecond,
	}
}

func TestCompletionMarkerRejectsWrongNoncePartialSymlinkNonRegularAndOversize(t *testing.T) {
	status := completionStatus{RC: 0}
	for name, prepare := range map[string]func(t *testing.T, endpoint completionEndpoint){
		"wrong_nonce": func(t *testing.T, endpoint completionEndpoint) {
			writeTestCompletionMarker(t, endpoint.Path, "nonce=wrong\nrc=0\n", 0o600)
		},
		"partial": func(t *testing.T, endpoint completionEndpoint) {
			writeTestCompletionMarker(t, endpoint.Path, "nonce="+endpoint.Nonce+"\n", 0o600)
		},
		"symlink": func(t *testing.T, endpoint completionEndpoint) {
			target := filepath.Join(endpoint.Dir, "target")
			writeTestCompletionMarker(t, target, encodeCompletionMarker(endpoint.Nonce, status), 0o600)
			if err := os.Symlink(target, endpoint.Path); err != nil {
				t.Fatal(err)
			}
		},
		"non_regular": func(t *testing.T, endpoint completionEndpoint) {
			if err := os.Mkdir(endpoint.Path, 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"oversize": func(t *testing.T, endpoint completionEndpoint) {
			writeTestCompletionMarker(t, endpoint.Path, strings.Repeat("x", completionMarkerMaxBytes+1), 0o600)
		},
		"wrong_mode": func(t *testing.T, endpoint completionEndpoint) {
			writeTestCompletionMarker(t, endpoint.Path, encodeCompletionMarker(endpoint.Nonce, status), 0o644)
		},
	} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			armRoot := filepath.Join(parent, "arm")
			if err := os.Mkdir(armRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanupCompletionEndpoint(endpoint)
			prepare(t, endpoint)
			if _, err := readCompletionMarker(endpoint); err == nil {
				t.Fatal("invalid completion marker was accepted")
			}
		})
	}
	t.Run("wrong_owner", func(t *testing.T) {
		parent := t.TempDir()
		armRoot := filepath.Join(parent, "arm")
		if err := os.Mkdir(armRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanupCompletionEndpoint(endpoint)
		writeTestCompletionMarker(t, endpoint.Path, encodeCompletionMarker(endpoint.Nonce, status), 0o600)
		endpoint.OwnerUID++
		if _, err := readCompletionMarker(endpoint); err == nil || !strings.Contains(err.Error(), "owner") {
			t.Fatalf("wrong owner was not rejected: %v", err)
		}
	})
}

func TestCompletionMarkerPublishIsAtomicAndStaleMarkersAreRejected(t *testing.T) {
	parent := t.TempDir()
	armRoot := filepath.Join(parent, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupCompletionEndpoint(endpoint)
	if err := publishCompletionMarker(endpoint, completionStatus{RC: 0}); err != nil {
		t.Fatal(err)
	}
	marker, err := readCompletionMarker(endpoint)
	if err != nil || marker.Status.RC != 0 {
		t.Fatalf("published marker = %#v err=%v", marker, err)
	}
	entries, err := os.ReadDir(endpoint.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(endpoint.Path) {
		t.Fatalf("publication left non-atomic artifacts: %v", entries)
	}

	staleParent := t.TempDir()
	staleArm := filepath.Join(staleParent, "arm")
	if err := os.Mkdir(staleArm, 0o700); err != nil {
		t.Fatal(err)
	}
	stale, err := newCompletionEndpoint(staleParent, staleArm, func() time.Time { return time.Now().Add(time.Hour) })
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupCompletionEndpoint(stale)
	writeTestCompletionMarker(t, stale.Path, encodeCompletionMarker(stale.Nonce, completionStatus{RC: 0}), 0o600)
	if _, err := readCompletionMarker(stale); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale marker was not rejected: %v", err)
	}
}

func TestConfinedRuntimeCannotDiscoverOrCreateCompletionMarker(t *testing.T) {
	skipUnlessBubblewrap(t)
	parent := t.TempDir()
	armRoot := filepath.Join(parent, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupCompletionEndpoint(endpoint)
	spec, err := newConfinementSpec("/bin/true", armRoot, testControllerBudgetLimits())
	if err != nil {
		t.Fatal(err)
	}
	command, err := completionProtocolCommand(spec, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	bwrapAt := indexString(command, spec.BwrapPath)
	if bwrapAt < 0 {
		t.Fatal("bwrap command is absent")
	}
	confinedArgs := strings.Join(command[bwrapAt:], "\x00")
	if strings.Contains(confinedArgs, endpoint.Dir) || strings.Contains(confinedArgs, endpoint.Nonce) || pathsOverlap(endpoint.Dir, armRoot) {
		t.Fatal("completion capability is exposed inside bwrap")
	}
	info, err := os.Stat(endpoint.Dir)
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("completion directory mode = %v err=%v", info.Mode().Perm(), err)
	}
}

func TestCommandWaitRunsExactlyOnceAndAlwaysJoinsBeforeReturn(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	handle := startCommandWait(func() error {
		calls.Add(1)
		<-release
		return nil
	}, time.Now)
	if handle.Completed() {
		t.Fatal("wait completed before release")
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	first, err := handle.Join(ctx)
	if err != nil || first.Err != nil {
		t.Fatalf("join failed: result=%#v err=%v", first, err)
	}
	if _, err := handle.Join(ctx); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || !handle.Completed() {
		t.Fatalf("Wait calls=%d completed=%v", calls.Load(), handle.Completed())
	}
}

func TestWaitBeforeMarkerAndMarkerBeforeTargetExitFailClosed(t *testing.T) {
	targetExit := time.Unix(100, 20)
	if err := validateCompletionOrdering(targetExit, targetExit.Add(time.Second), targetExit.Add(-time.Second)); err == nil {
		t.Fatal("Wait before marker was accepted")
	}
	if err := validateCompletionOrdering(targetExit, targetExit.Add(-time.Second), targetExit.Add(time.Second)); err == nil {
		t.Fatal("marker before target exit was accepted")
	}
	if err := validateCompletionOrdering(targetExit, targetExit, targetExit.Add(time.Second)); err != nil {
		t.Fatalf("valid ordering rejected: %v", err)
	}
}

func TestCompletionProtocolPreservesStdoutAndStderrBytesForSuccessAndFailure(t *testing.T) {
	for _, waitErr := range []error{nil, errors.New("exit 7")} {
		stdout := &limitedBuffer{Limit: 100}
		stderr := &limitedBuffer{Limit: 100}
		_, _ = stdout.Write([]byte("raw stdout\x00\n"))
		_, _ = stderr.Write([]byte("raw stderr\x00\n"))
		handle := startCommandWait(func() error { return waitErr }, time.Now)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if _, err := handle.Join(ctx); err != nil {
			cancel()
			t.Fatal(err)
		}
		gotOut, gotErr, err := snapshotCompletedOutput(handle, stdout, stderr)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		if string(gotOut) != "raw stdout\x00\n" || string(gotErr) != "raw stderr\x00\n" {
			t.Fatalf("coordination changed output: %q %q", gotOut, gotErr)
		}
	}
}

func TestCompletionStatusReconcilesZeroNonzeroSignalAndRuntimeMaxKill(t *testing.T) {
	valid := []completionStatus{{RC: 0}, {RC: 7}, {RC: 143}}
	for _, status := range valid {
		if err := reconcileCompletionStatus(status, status); err != nil {
			t.Fatalf("matching status rejected: %#v: %v", status, err)
		}
	}
	if err := reconcileCompletionStatus(completionStatus{RC: 0}, completionStatus{RC: 7}); err == nil {
		t.Fatal("mismatched exit was accepted")
	}
	if err := reconcileCompletionStatus(completionStatus{RC: -1}, completionStatus{RC: 137}); err == nil {
		t.Fatal("RuntimeMax kill without marker was accepted")
	}
}

func TestCancellationBeforeMarkerAndDuringFinalReadReturnsBoundedAndCleansUnit(t *testing.T) {
	for _, duringRead := range []bool{false, true} {
		t.Run(strconv.FormatBool(duringRead), func(t *testing.T) {
			parent := t.TempDir()
			armRoot := filepath.Join(parent, "arm")
			if err := os.Mkdir(armRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
			if err != nil {
				t.Fatal(err)
			}
			if duringRead {
				readCtx, readCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				sampler := &cgroupSampler{root: "/resolved", reader: blockingCgroupReader{}, config: testCgroupSamplerConfig()}
				if _, readErr := takeCompletionAccountingSnapshot(readCtx, sampler); readErr == nil || !errors.Is(readErr, context.DeadlineExceeded) {
					readCancel()
					t.Fatalf("canceled final read was not bounded: %v", readErr)
				}
				readCancel()
			}
			release := make(chan struct{})
			handle := startCommandWait(func() error { <-release; return context.Canceled }, time.Now)
			stopped := false
			err = abortAndJoinCompletionPhased(endpoint, handle, time.Second, time.Second, func(context.Context) error {
				stopped = true
				close(release)
				return nil
			}, func(context.Context) error {
				t.Fatal("hard stop ran after normal join")
				return nil
			}, cleanupCompletionEndpoint)
			if err != nil || !stopped || !handle.Completed() {
				t.Fatalf("bounded cancellation failed: stopped=%v completed=%v err=%v", stopped, handle.Completed(), err)
			}
			if _, err := os.Stat(endpoint.Dir); !os.IsNotExist(err) {
				t.Fatalf("completion directory survived cancellation: %v", err)
			}
		})
	}
}

func TestFinalReadUsesOnceResolvedCgroupIdentityWithoutReresolution(t *testing.T) {
	resolver := &countingCgroupResolver{resolved: resolvedControlGroup{Root: "/resolved", Attempts: 1}}
	reader := &recordingCgroupReader{files: validCgroupFiles("/resolved", 2500, 4096, 3)}
	sampler, err := newCgroupSampler(context.Background(), "unit", resolver, reader, testCgroupSamplerConfig())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := takeCompletionAccountingSnapshot(context.Background(), sampler); err != nil {
		t.Fatal(err)
	}
	if resolver.calls != 1 {
		t.Fatalf("cgroup identity resolved %d times", resolver.calls)
	}
}

func TestPostMarkerRetentionCannotChangeAccountedCPUOrPIDs(t *testing.T) {
	reader := &recordingCgroupReader{files: validCgroupFiles("/resolved", 2500, 4096, 3)}
	sampler := &cgroupSampler{root: "/resolved", reader: reader, config: testCgroupSamplerConfig()}
	accounting, err := takeCompletionAccountingSnapshot(context.Background(), sampler)
	if err != nil {
		t.Fatal(err)
	}
	reader.files = validCgroupFiles("/resolved", 99000, 8192, 9)
	if accounting.Final.CPUUsageMicros != 2500 || accounting.Final.PIDsPeak != 3 || len(reader.paths) != 3 {
		t.Fatalf("retention revised endpoint: %#v paths=%v", accounting.Final, reader.paths)
	}
}

func TestCompletionProtocolPreservesExactlyOne250msRetention(t *testing.T) {
	skipUnlessBubblewrap(t)
	parent := t.TempDir()
	armRoot := filepath.Join(parent, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupCompletionEndpoint(endpoint)
	spec, err := newConfinementSpec("/bin/true", armRoot, testControllerBudgetLimits())
	if err != nil {
		t.Fatal(err)
	}
	command, err := completionProtocolCommand(spec, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	script := strings.Join(command, "\n")
	if strings.Count(script, "/bin/sleep 0.25") != 1 || strings.Index(script, `"$move_bin" --`) > strings.Index(script, "/bin/sleep 0.25") {
		t.Fatalf("retention command changed: %q", script)
	}
}

func TestLifetimeBurstUsesExactDiscardedSamplerHashesAsItsBaseline(t *testing.T) {
	if reconstructedSamplerMainSHA256 != "1e45eb75c0bbdf8bfe427957565b92bb6036ec1346da42d210874a5bbe2f366a" || reconstructedSamplerTestSHA256 != "baf439870b25a09becd2ec3e9093f6745eb72f7c036ee666ae468b60ed818082" {
		t.Fatal("lifetime burst baseline binding changed")
	}
}

func TestPathnameMarkerIsRejectedWhenHostileSameUIDIsInScope(t *testing.T) {
	if err := validateCompletionThreatScope(true); err == nil {
		t.Fatal("pathname marker claimed hostile same-UID protection")
	}
	if err := validateCompletionThreatScope(false); err != nil {
		t.Fatal(err)
	}
}

func TestAccountingEndpointUsesFirstBoundedDirectReadAfterAuthenticatedTargetExit(t *testing.T) {
	reader := &deadlineCheckingCgroupReader{}
	sampler := &cgroupSampler{root: "/resolved", reader: reader, config: testCgroupSamplerConfig()}
	accounting, err := takeCompletionAccountingSnapshot(context.Background(), sampler)
	if err != nil {
		t.Fatal(err)
	}
	if reader.calls != 3 || accounting.Diagnostics.SampleAttempts != 1 || accounting.Diagnostics.SuccessfulSamples != 1 || accounting.Diagnostics.FinalReadLatencyMicros < 0 {
		t.Fatalf("endpoint was not one bounded direct read: calls=%d diagnostics=%#v", reader.calls, accounting.Diagnostics)
	}
}

func TestWallEndsAtAuthenticatedMarkerAndExcludesLaterRetentionWork(t *testing.T) {
	started := time.Unix(100, 0)
	markerAt := started.Add(25 * time.Millisecond)
	joinedAt := markerAt.Add(250 * time.Millisecond)
	accounting := cgroupAccountingResult{Final: cgroupStats{CPUUsageMicros: 2500, MemoryPeak: 4096, PIDsPeak: 3}, Diagnostics: cgroupAccountingDiagnostics{SuccessfulCPUReads: 1}}
	actual, diagnostics, err := finalizeArmAccountingAtCompletion(started, markerAt, joinedAt, accounting, 1, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if actual.WallTimeMillis != 25 || diagnostics.MonitorJoinTailMillis != 250 {
		t.Fatalf("wall includes retention: budget=%#v diagnostics=%#v", actual, diagnostics)
	}
}

func TestCancellationStopsAndCleansNamedUnitWithoutLingeringUnitOrLateMarker(t *testing.T) {
	parent := t.TempDir()
	armRoot := filepath.Join(parent, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	handle := startCommandWait(func() error {
		<-release
		_ = publishCompletionMarker(endpoint, completionStatus{RC: 0})
		return context.Canceled
	}, time.Now)
	stops := 0
	if err := abortAndJoinCompletionPhased(endpoint, handle, time.Second, time.Second, func(context.Context) error {
		stops++
		close(release)
		return nil
	}, func(context.Context) error {
		t.Fatal("hard stop ran after normal join")
		return nil
	}, cleanupCompletionEndpoint); err != nil {
		t.Fatal(err)
	}
	if stops != 1 || !handle.Completed() {
		t.Fatalf("cancel stops=%d completed=%v", stops, handle.Completed())
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := os.Stat(endpoint.Path); !os.IsNotExist(err) {
		t.Fatalf("late marker survived cleanup: %v", err)
	}
}

func TestNoDirectReadFailureCanReresolveOrUsePeriodicOrSystemdFallback(t *testing.T) {
	resolver := &countingCgroupResolver{resolved: resolvedControlGroup{Root: "/resolved", Attempts: 1}}
	reader := &recordingCgroupReader{files: map[string]string{}}
	sampler, err := newCgroupSampler(context.Background(), "unit", resolver, reader, testCgroupSamplerConfig())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := takeCompletionAccountingSnapshot(context.Background(), sampler); err == nil {
		t.Fatal("direct read failure used a fallback")
	}
	if resolver.calls != 1 || len(reader.paths) != 1 {
		t.Fatalf("failure retried or fell back: resolves=%d paths=%v", resolver.calls, reader.paths)
	}
}

func TestReconstructedBaselineHashesAreExactly1e45AndBaf439(t *testing.T) {
	root := repoRoot(t)
	skipUnlessReleaseFixtures(t,
		filepath.Join(root, ".okra/content/sha256/1e45eb75c0bbdf8bfe427957565b92bb6036ec1346da42d210874a5bbe2f366a"),
		filepath.Join(root, ".okra/content/sha256/baf439870b25a09becd2ec3e9093f6745eb72f7c036ee666ae468b60ed818082"),
	)
	for hash, rel := range map[string]string{
		"1e45eb75c0bbdf8bfe427957565b92bb6036ec1346da42d210874a5bbe2f366a": ".okra/content/sha256/1e45eb75c0bbdf8bfe427957565b92bb6036ec1346da42d210874a5bbe2f366a",
		"baf439870b25a09becd2ec3e9093f6745eb72f7c036ee666ae468b60ed818082": ".okra/content/sha256/baf439870b25a09becd2ec3e9093f6745eb72f7c036ee666ae468b60ed818082",
	} {
		if got := sha(mustRead(t, filepath.Join(root, rel))); got != hash {
			t.Fatalf("baseline blob %s = %s", rel, got)
		}
	}
}

func TestAbortJoinTimeoutStillCleansMarkerAndDoesNotReturnWithLiveWait(t *testing.T) {
	newEndpoint := func(t *testing.T) completionEndpoint {
		t.Helper()
		parent := t.TempDir()
		armRoot := filepath.Join(parent, "arm")
		if err := os.Mkdir(armRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
		if err != nil {
			t.Fatal(err)
		}
		return endpoint
	}

	t.Run("stop_deadline_does_not_consume_join_deadline", func(t *testing.T) {
		endpoint := newEndpoint(t)
		release := make(chan struct{})
		handle := startCommandWait(func() error { <-release; return context.Canceled }, time.Now)
		hardStops := 0
		err := abortAndJoinCompletionPhased(endpoint, handle, 10*time.Millisecond, time.Second, func(ctx context.Context) error {
			<-ctx.Done()
			close(release)
			return ctx.Err()
		}, func(context.Context) error {
			hardStops++
			return nil
		}, cleanupCompletionEndpoint)
		if err == nil || errors.Is(err, errControllerFatalLiveWait) || !handle.Completed() || hardStops != 0 {
			t.Fatalf("independent join phase failed: completed=%v err=%v", handle.Completed(), err)
		}
		if _, statErr := os.Stat(endpoint.Dir); !os.IsNotExist(statErr) {
			t.Fatalf("marker directory survived phased abort: %v", statErr)
		}
	})

	t.Run("hard_stop_reaps_before_return", func(t *testing.T) {
		endpoint := newEndpoint(t)
		release := make(chan struct{})
		handle := startCommandWait(func() error { <-release; return context.Canceled }, time.Now)
		hardStops := 0
		err := abortAndJoinCompletionPhased(endpoint, handle, 5*time.Millisecond, 100*time.Millisecond, func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}, func(context.Context) error {
			hardStops++
			close(release)
			return nil
		}, cleanupCompletionEndpoint)
		if err == nil || errors.Is(err, errControllerFatalLiveWait) || !handle.Completed() || hardStops != 1 {
			t.Fatalf("hard reap did not join Wait: completed=%v hard_stops=%d err=%v", handle.Completed(), hardStops, err)
		}
		if _, statErr := os.Stat(endpoint.Dir); !os.IsNotExist(statErr) {
			t.Fatalf("marker directory survived hard reap: %v", statErr)
		}
	})

	t.Run("failed_hard_reap_is_controller_fatal", func(t *testing.T) {
		endpoint := newEndpoint(t)
		release := make(chan struct{})
		handle := startCommandWait(func() error { <-release; return context.Canceled }, time.Now)
		err := abortAndJoinCompletionPhased(endpoint, handle, 5*time.Millisecond, 5*time.Millisecond, func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}, func(context.Context) error { return errors.New("hard stop failed") }, cleanupCompletionEndpoint)
		if !errors.Is(err, errControllerFatalLiveWait) {
			t.Fatalf("failed hard reap was not controller-fatal: %v", err)
		}
		if _, statErr := os.Stat(endpoint.Dir); !os.IsNotExist(statErr) {
			t.Fatalf("marker directory survived controller-fatal reap: %v", statErr)
		}
		close(release)
		joinCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if _, joinErr := handle.Join(joinCtx); joinErr != nil {
			t.Fatal(joinErr)
		}
	})
}

func TestTransientUnitShowTransportErrorIsNotAcceptedAsAbsent(t *testing.T) {
	t.Run("transport_error", func(t *testing.T) {
		runner := &scriptedSystemctlRunner{responses: []scriptedSystemctlResponse{
			{}, {}, {err: errors.New("D-Bus transport unavailable")},
		}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := stopAndCleanTransientUnitWithRunner(ctx, "unit", runner)
		if err == nil || !strings.Contains(err.Error(), "D-Bus transport unavailable") {
			t.Fatalf("show transport error was accepted as absence: %v", err)
		}
	})

	t.Run("loaded_retries_until_successful_empty_absence", func(t *testing.T) {
		runner := &scriptedSystemctlRunner{responses: []scriptedSystemctlResponse{
			{}, {}, {output: "loaded\n"}, {output: "loaded\n"}, {output: "\n"},
		}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := stopAndCleanTransientUnitWithRunner(ctx, "unit", runner); err != nil {
			t.Fatal(err)
		}
		if runner.calls != 5 {
			t.Fatalf("systemctl calls=%d, want stop reset plus three show reads", runner.calls)
		}
	})

	t.Run("explicit_not_found_error_proves_absence", func(t *testing.T) {
		runner := &scriptedSystemctlRunner{responses: []scriptedSystemctlResponse{
			{}, {}, {output: "Unit unit.service could not be found.\n", err: errors.New("exit 1")},
		}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := stopAndCleanTransientUnitWithRunner(ctx, "unit", runner); err != nil {
			t.Fatal(err)
		}
	})
}

func TestExplicitExit130IsNotReclassifiedAsSignal2(t *testing.T) {
	err := exec.Command("/bin/sh", "-c", "exit 130").Run()
	status := completionStatusFromWaitError(err)
	if status.RC != 130 {
		t.Fatalf("explicit exit 130 became %#v", status)
	}
	encoded := encodeCompletionMarker(strings.Repeat("a", 64), status)
	if !strings.Contains(encoded, "rc=130\n") || strings.Contains(encoded, "signal") || strings.Contains(encoded, "kind=") {
		t.Fatalf("marker is not numeric-only: %q", encoded)
	}
}

func TestProductionShellPublisherRejectsTempWriteRenameAndLatePublishFailures(t *testing.T) {
	newEndpoint := func(t *testing.T) completionEndpoint {
		t.Helper()
		parent := t.TempDir()
		armRoot := filepath.Join(parent, "arm")
		if err := os.Mkdir(armRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
		if err != nil {
			t.Fatal(err)
		}
		return endpoint
	}

	t.Run("temp_write", func(t *testing.T) {
		endpoint := newEndpoint(t)
		if err := cleanupCompletionEndpoint(endpoint); err != nil {
			t.Fatal(err)
		}
		err := runProductionCompletionShell(endpoint, defaultCompletionShellTools(), "/bin/true")
		if completionRCFromError(err) != 125 {
			t.Fatalf("temp-write failure rc=%d err=%v", completionRCFromError(err), err)
		}
		assertNoCompletionArtifacts(t, endpoint)
	})

	t.Run("rename", func(t *testing.T) {
		endpoint := newEndpoint(t)
		tools := defaultCompletionShellTools()
		tools.MovePath = "/bin/false"
		err := runProductionCompletionShell(endpoint, tools, "/bin/true")
		if completionRCFromError(err) != 125 {
			t.Fatalf("rename failure rc=%d err=%v", completionRCFromError(err), err)
		}
		assertNoCompletionArtifacts(t, endpoint)
	})

	t.Run("late_publish", func(t *testing.T) {
		endpoint := newEndpoint(t)
		signalPath := filepath.Join(t.TempDir(), "started")
		releasePath := filepath.Join(t.TempDir(), "release")
		result := make(chan error, 1)
		go func() {
			result <- runProductionCompletionShell(endpoint, defaultCompletionShellTools(), "/bin/sh", "-c", `: >"$1"; while [ ! -e "$2" ]; do /bin/sleep 0.001; done`, "sh", signalPath, releasePath)
		}()
		deadline := time.Now().Add(time.Second)
		for {
			if _, err := os.Stat(signalPath); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("target did not reach late-publish barrier")
			}
			time.Sleep(time.Millisecond)
		}
		if err := cleanupCompletionEndpoint(endpoint); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(releasePath, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		err := <-result
		if completionRCFromError(err) != 125 {
			t.Fatalf("late publication rc=%d err=%v", completionRCFromError(err), err)
		}
		assertNoCompletionArtifacts(t, endpoint)
	})
}

func TestStatusMismatchRecordsCompletionCleanupFailure(t *testing.T) {
	cleanupFailure := errors.New("completion cleanup failed")
	diagnostics := cgroupAccountingDiagnostics{}
	err := reconcileAndCleanupCompletion(
		completionStatus{RC: 0},
		completionStatus{RC: 7},
		completionEndpoint{},
		func(completionEndpoint) error { return cleanupFailure },
		&diagnostics,
	)
	if err == nil || !errors.Is(err, cleanupFailure) || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("status and cleanup errors were not joined: %v", err)
	}
	if diagnostics.SampleErrorCount != 1 || len(diagnostics.Errors) != 1 || diagnostics.Errors[0].Stage != "completion_cleanup" || diagnostics.Errors[0].ErrorSHA256 != shaString(cleanupFailure.Error()) {
		t.Fatalf("cleanup failure is not durable: %#v", diagnostics)
	}
}

func TestCompletionEndpointCleanupRunsDuringPanicUnwind(t *testing.T) {
	parent := t.TempDir()
	armRoot := filepath.Join(parent, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	endpoint, err := newCompletionEndpoint(parent, armRoot, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	recovered := func() (recovered any) {
		defer func() { recovered = recover() }()
		defer guardCompletionEndpointCleanup(endpoint, cleanupCompletionEndpoint)()
		panic("panic after endpoint creation")
	}()
	if recovered == nil {
		t.Fatal("panic did not unwind through completion cleanup guard")
	}
	if _, statErr := os.Stat(endpoint.Dir); !os.IsNotExist(statErr) {
		t.Fatalf("completion endpoint survived panic unwind: %v", statErr)
	}
}

func writeTestCompletionMarker(t *testing.T, path, contents string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		t.Fatal(err)
	}
}

func indexString(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func validCgroupFiles(root string, cpu, memory int64, pids int) map[string]string {
	return map[string]string{
		filepath.Join(root, "cpu.stat"):    fmt.Sprintf("usage_usec %d\n", cpu),
		filepath.Join(root, "memory.peak"): fmt.Sprintf("%d\n", memory),
		filepath.Join(root, "pids.peak"):   fmt.Sprintf("%d\n", pids),
	}
}

type blockingCgroupReader struct{}

func (blockingCgroupReader) ReadFile(ctx context.Context, _ string) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

type scriptedSystemctlResponse struct {
	output string
	err    error
}

type scriptedSystemctlRunner struct {
	responses []scriptedSystemctlResponse
	calls     int
}

func (r *scriptedSystemctlRunner) Run(_ context.Context, _ ...string) ([]byte, error) {
	if r.calls >= len(r.responses) {
		return nil, errors.New("unexpected systemctl call")
	}
	response := r.responses[r.calls]
	r.calls++
	return []byte(response.output), response.err
}

func runProductionCompletionShell(endpoint completionEndpoint, tools completionShellTools, target ...string) error {
	script, err := completionShellScript(tools)
	if err != nil {
		return err
	}
	args := []string{"-c", script, "sh", endpoint.Dir, endpoint.Nonce, tools.MovePath, tools.RemovePath}
	args = append(args, target...)
	return exec.Command("/bin/sh", args...).Run()
}

func completionRCFromError(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func assertNoCompletionArtifacts(t *testing.T, endpoint completionEndpoint) {
	t.Helper()
	if _, err := os.Stat(endpoint.Path); !os.IsNotExist(err) {
		t.Fatalf("completion marker exists after failed publication: %v", err)
	}
	if entries, err := os.ReadDir(endpoint.Dir); err == nil && len(entries) != 0 {
		t.Fatalf("completion temp artifacts remain: %v", entries)
	} else if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
