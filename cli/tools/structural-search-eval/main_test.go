package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFrozenFixtureShape(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	if got, want := len(fixtures), 12; got != want {
		t.Fatalf("fixture count = %d, want %d", got, want)
	}

	wrongLayer := 0
	routes := 0
	for _, fixture := range fixtures {
		if err := validateFixture(fixture); err != nil {
			t.Fatalf("%s: %v", fixture.CaseID, err)
		}
		switch fixture.Family {
		case familyWrongLayer:
			wrongLayer++
			distractors := 0
			nonRequired := 0
			required := stringSet(fixture.RequiredOwnerFactIDs)
			for _, document := range fixture.Documents {
				if document.SourceKind == "topical_prose" {
					distractors++
				}
				for _, fact := range document.Facts {
					if !required[fact.FactID] {
						nonRequired++
					}
				}
			}
			if distractors < 2 {
				t.Fatalf("%s has %d topical distractors, want at least 2", fixture.CaseID, distractors)
			}
			if nonRequired < 3 {
				t.Fatalf("%s has %d allowed non-required facts, want at least 3", fixture.CaseID, nonRequired)
			}
		case familyRoute:
			routes++
		default:
			t.Fatalf("%s has unknown family %q", fixture.CaseID, fixture.Family)
		}
	}
	if wrongLayer != 8 || routes != 4 {
		t.Fatalf("family split = %d wrong-layer + %d route, want 8 + 4", wrongLayer, routes)
	}
}

func TestBaselineReplayMatchesFrozenMetrics(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	manifest := loadRepoManifest(t)
	report, err := runCandidate(fixtures, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := compareFrozenBaseline(manifest.FrozenBaseline, report.Metrics); err != nil {
		t.Fatal(err)
	}
	if report.Metrics.FalseStructuralClaimCount != 0 {
		t.Fatalf("false structural claims = %d, want 0", report.Metrics.FalseStructuralClaimCount)
	}
	if len(report.Metrics.RetrievedContextBytesPerCase) != 12 {
		t.Fatalf("per-case byte reads = %d, want 12", len(report.Metrics.RetrievedContextBytesPerCase))
	}
}

func TestEntityDumpNegativeControlFailsPrecision(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	baselineManifest := loadRepoManifest(t)
	baseline, err := runCandidate(fixtures, baselineManifest)
	if err != nil {
		t.Fatal(err)
	}

	candidate := derivedManifest(t, baselineManifest, "negative-entity-dump", "entity_dump")
	candidateReport, err := runCandidate(fixtures, candidate)
	if err != nil {
		t.Fatal(err)
	}
	verdict := evaluateGate(baseline, candidateReport, baselineManifest)
	if verdict.Keep {
		t.Fatal("entity dump unexpectedly passed the keep gate")
	}
	if !contains(verdict.FailedWalls, wallStructuralOwnerPrecision) {
		t.Fatalf("failed walls = %v, want %s", verdict.FailedWalls, wallStructuralOwnerPrecision)
	}
	if candidateReport.Metrics.StructuralOwnerPrecision >= 0.80 {
		t.Fatalf("entity-dump precision = %.3f, want below 0.80", candidateReport.Metrics.StructuralOwnerPrecision)
	}
	t.Logf("negative control: recall@5=%.3f precision=%.3f verdict=%s failed=%v", candidateReport.Metrics.OwnerRecallAt5, candidateReport.Metrics.StructuralOwnerPrecision, verdict.Verdict, verdict.FailedWalls)
}

func TestUnknownAssertionCountsAsFalseStructure(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	manifest := loadRepoManifest(t)
	manifest.CandidateID = "unknown-assertion"
	manifest.AssertedFactIDs = []string{"fact-that-is-not-registered"}
	report, err := runCandidate(fixtures, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := report.Metrics.FalseStructuralClaimCount, len(fixtures); got != want {
		t.Fatalf("false structural claim count = %d, want %d", got, want)
	}
}

func TestForbiddenPositiveTripleCountsAsFalseStructure(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	manifest := loadRepoManifest(t)
	manifest.CandidateID = "forbidden-triple"
	manifest.AssertedTriples = []structuralTriple{{Subject: "Panel", Relation: "owns", Object: "Ledger"}}
	report, err := runCandidate(fixtures, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.FalseStructuralClaimCount == 0 {
		t.Fatal("forbidden positive triple was not counted")
	}
}

func TestUnregisteredTripleCountsAsFalseStructureWithoutAbsenceInference(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	manifest := loadRepoManifest(t)
	manifest.CandidateID = "unregistered-triple"
	manifest.AssertedTriples = []structuralTriple{{Subject: "UnknownWriter", Relation: "writes", Object: "UnknownRow"}}
	report, err := runCandidate(fixtures, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := report.Metrics.FalseStructuralClaimCount, len(fixtures); got != want {
		t.Fatalf("unregistered triple count = %d, want %d", got, want)
	}
	// The scorer evaluates only explicit positive assertions. No asserted facts
	// means no claim is manufactured from a missing document or edge.
	manifest.AssertedTriples = nil
	report, err = runCandidate(fixtures, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Metrics.FalseStructuralClaimCount != 0 {
		t.Fatalf("absence produced %d false claims, want 0", report.Metrics.FalseStructuralClaimCount)
	}
}

func TestFixtureRejectsAllowlistDriftAndNegativeFacts(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	allowlistDrift := cloneFixture(t, fixtures[0])
	allowlistDrift.AllowedStructuralFactIDs = allowlistDrift.AllowedStructuralFactIDs[1:]
	if err := validateFixture(allowlistDrift); err == nil {
		t.Fatal("fixture accepted an anchored fact outside the exact allowlist")
	}
	negative := cloneFixture(t, fixtures[0])
	negative.Documents[0].Facts[0].Polarity = "negative"
	if err := validateFixture(negative); err == nil {
		t.Fatal("fixture accepted a negative/absence-style structural fact")
	}
}

func TestDuplicateRankedDocumentDoesNotDoubleCountContextOrFacts(t *testing.T) {
	item := loadRepoFixtures(t)[0]
	manifest := loadRepoManifest(t)
	ranked := rankDocuments(item, manifest.Ranking.Mode)
	ranked[1] = ranked[0]
	result := scoreCase(item, ranked, manifest)
	if got, want := result.ContextBytesAt5, 132; got != want {
		t.Fatalf("deduplicated context bytes = %d, want %d", got, want)
	}
	if got, want := len(result.RetrievedFactIDsAt5), 3; got != want {
		t.Fatalf("deduplicated retrieved facts = %d, want %d", got, want)
	}
}

func TestPerCaseContextLimitCannotBeHiddenByAverage(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	baselineManifest := loadRepoManifest(t)
	baseline, err := runCandidate(fixtures, baselineManifest)
	if err != nil {
		t.Fatal(err)
	}
	candidateManifest := derivedManifest(t, baselineManifest, "one-bloated-case", "structural_first")
	candidate, err := runCandidate(fixtures, candidateManifest)
	if err != nil {
		t.Fatal(err)
	}
	for caseID, baselineBytes := range baseline.Metrics.RetrievedContextBytesPerCase {
		candidate.Metrics.RetrievedContextBytesPerCase[caseID] = baselineBytes
	}
	firstCase := fixtures[0].CaseID
	candidate.Metrics.RetrievedContextBytesPerCase[firstCase] = baseline.Metrics.RetrievedContextBytesPerCase[firstCase]*105/100 + 1

	verdict := evaluateGate(baseline, candidate, baselineManifest)
	if !contains(verdict.FailedWalls, wallContextBytesPerCase) {
		t.Fatalf("failed walls = %v, want %s", verdict.FailedWalls, wallContextBytesPerCase)
	}
}

func TestGateRejectsMoreThanOneChangedVariable(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	baselineManifest := loadRepoManifest(t)
	baseline, err := runCandidate(fixtures, baselineManifest)
	if err != nil {
		t.Fatal(err)
	}
	candidateManifest := derivedManifest(t, baselineManifest, "two-variables", "structural_first")
	candidateManifest.Ranking.TieBreak = "document_id_desc"
	candidate := baseline
	candidate.Candidate = candidateManifest
	verdict := evaluateGate(baseline, candidate, baselineManifest)
	if !contains(verdict.FailedWalls, wallOneVariable) {
		t.Fatalf("failed walls = %v, want %s", verdict.FailedWalls, wallOneVariable)
	}
}

func TestGateRejectsRouteAndWrongLayerMRRRegressions(t *testing.T) {
	fixtures := loadRepoFixtures(t)
	baselineManifest := loadRepoManifest(t)
	baseline, err := runCandidate(fixtures, baselineManifest)
	if err != nil {
		t.Fatal(err)
	}
	candidateManifest := derivedManifest(t, baselineManifest, "mrr-regressions", "structural_first")
	candidate := baseline
	candidate.Candidate = candidateManifest
	candidate.Metrics.BehavioralRouteMRR = baseline.Metrics.BehavioralRouteMRR - 0.01
	candidate.Metrics.WrongLayerMRR = baseline.Metrics.WrongLayerMRR - 0.01
	verdict := evaluateGate(baseline, candidate, baselineManifest)
	if !contains(verdict.FailedWalls, wallBlastRouteMRR) || !contains(verdict.FailedWalls, wallWrongLayerMRR) {
		t.Fatalf("failed walls = %v, want route and wrong-layer MRR walls", verdict.FailedWalls)
	}
}

func TestRunRejectsFixtureAndScorerHashDrift(t *testing.T) {
	fixturesPath := repoPath(t, "research/eval/structural-retrieval/fixtures.v1.jsonl")
	baselinePath := repoPath(t, "research/eval/structural-retrieval/candidate-baseline.v1.json")
	fixtureData, err := os.ReadFile(fixturesPath)
	if err != nil {
		t.Fatal(err)
	}
	driftedFixture := filepath.Join(t.TempDir(), "fixtures.jsonl")
	if err := os.WriteFile(driftedFixture, append(fixtureData, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := execute([]string{"run", "--fixtures", driftedFixture, "--candidate", baselinePath, "--out", filepath.Join(t.TempDir(), "report.json")}, &bytes.Buffer{}, &bytes.Buffer{}); code == 0 {
		t.Fatal("run accepted fixture hash drift")
	}

	manifest := loadRepoManifest(t)
	manifest.ScorerSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
	driftedManifest := filepath.Join(t.TempDir(), "candidate.json")
	writeJSON(t, driftedManifest, manifest)
	if code := execute([]string{"run", "--fixtures", fixturesPath, "--candidate", driftedManifest, "--out", filepath.Join(t.TempDir(), "report.json")}, &bytes.Buffer{}, &bytes.Buffer{}); code == 0 {
		t.Fatal("run accepted scorer hash drift")
	}
}

func TestBaselineReplayIsByteIdentical(t *testing.T) {
	fixturesPath := repoPath(t, "research/eval/structural-retrieval/fixtures.v1.jsonl")
	baselinePath := repoPath(t, "research/eval/structural-retrieval/candidate-baseline.v1.json")
	one := filepath.Join(t.TempDir(), "one.json")
	two := filepath.Join(t.TempDir(), "two.json")
	for _, output := range []string{one, two} {
		var stderr bytes.Buffer
		if code := execute([]string{"run", "--fixtures", fixturesPath, "--candidate", baselinePath, "--out", output}, &bytes.Buffer{}, &stderr); code != 0 {
			t.Fatalf("run exit %d: %s", code, stderr.String())
		}
	}
	oneBytes, err := os.ReadFile(one)
	if err != nil {
		t.Fatal(err)
	}
	twoBytes, err := os.ReadFile(two)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(oneBytes, twoBytes) {
		t.Fatal("identical frozen inputs did not produce byte-identical reports")
	}
}

func TestEntityDumpGateCommandReturnsNonZero(t *testing.T) {
	fixturesPath := repoPath(t, "research/eval/structural-retrieval/fixtures.v1.jsonl")
	baselinePath := repoPath(t, "research/eval/structural-retrieval/candidate-baseline.v1.json")
	baselineManifest := loadRepoManifest(t)
	candidate := derivedManifest(t, baselineManifest, "negative-entity-dump-cli", "entity_dump")
	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeJSON(t, candidatePath, candidate)

	baselineReport := filepath.Join(t.TempDir(), "baseline-report.json")
	candidateReport := filepath.Join(t.TempDir(), "candidate-report.json")
	var stderr bytes.Buffer
	if code := execute([]string{"run", "--fixtures", fixturesPath, "--candidate", baselinePath, "--out", baselineReport}, &bytes.Buffer{}, &stderr); code != 0 {
		t.Fatalf("baseline run exit = %d: %s", code, stderr.String())
	}
	stderr.Reset()
	if code := execute([]string{"run", "--fixtures", fixturesPath, "--candidate", candidatePath, "--out", candidateReport}, &bytes.Buffer{}, &stderr); code != 0 {
		t.Fatalf("candidate run exit = %d: %s", code, stderr.String())
	}
	stderr.Reset()
	var stdout bytes.Buffer
	code := execute([]string{"gate", "--baseline", baselineReport, "--candidate", candidateReport, "--freeze", baselinePath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("negative control gate exit = 0; report: %s", stdout.String())
	}
	var verdict gateVerdict
	if err := json.Unmarshal(stdout.Bytes(), &verdict); err != nil {
		t.Fatalf("decode gate report: %v (%s)", err, stdout.String())
	}
	if !contains(verdict.FailedWalls, wallStructuralOwnerPrecision) {
		t.Fatalf("failed walls = %v, want %s", verdict.FailedWalls, wallStructuralOwnerPrecision)
	}
}

func loadRepoFixtures(t *testing.T) []fixture {
	t.Helper()
	fixtures, err := loadFixtures(repoPath(t, "research/eval/structural-retrieval/fixtures.v1.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func loadRepoManifest(t *testing.T) candidateManifest {
	t.Helper()
	manifest, err := loadCandidate(repoPath(t, "research/eval/structural-retrieval/candidate-baseline.v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func derivedManifest(t *testing.T, baseline candidateManifest, id, mode string) candidateManifest {
	t.Helper()
	baselinePath := repoPath(t, "research/eval/structural-retrieval/candidate-baseline.v1.json")
	baselineHash, err := hashFile(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	candidate := baseline
	candidate.CandidateID = id
	candidate.ParentCandidateID = baseline.CandidateID
	candidate.BaselineManifestSHA256 = baselineHash
	candidate.ChangedVariable = &changedVariable{Key: "ranking.mode", BaselineValue: baseline.Ranking.Mode, CandidateValue: mode}
	candidate.Ranking.Mode = mode
	candidate.FrozenBaseline = metrics{}
	return candidate
}

func repoPath(t *testing.T, relative string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, relative)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func cloneFixture(t *testing.T, value fixture) fixture {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var cloned fixture
	if err := json.Unmarshal(data, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}
