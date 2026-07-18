package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckedInV3ArtifactsValidate(t *testing.T) {
	root := filepathFromRoot("research/eval/structural-retrieval-v3")
	fixtureBytes, err := os.ReadFile(filepath.Join(root, "fixtures.v3.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []V3Fixture
	if err := json.Unmarshal(fixtureBytes, &fixtures); err != nil {
		t.Fatalf("fixtures.v3.json: %v", err)
	}
	for _, fixture := range fixtures {
		if err := ValidateFixture(fixture); err != nil {
			t.Fatalf("%s: %v", fixture.CaseID, err)
		}
	}
	benchmarkBytes, err := os.ReadFile(filepath.Join(root, "benchmark.v3.json"))
	if err != nil {
		t.Fatal(err)
	}
	var benchmark V3Benchmark
	if err := json.Unmarshal(benchmarkBytes, &benchmark); err != nil {
		t.Fatalf("benchmark.v3.json: %v", err)
	}
	if err := ValidateBenchmark(benchmark); err != nil {
		t.Fatalf("benchmark.v3.json: %v", err)
	}
	if benchmark.FixtureCount != len(fixtures) {
		t.Fatalf("fixture_count=%d, fixtures=%d", benchmark.FixtureCount, len(fixtures))
	}
	var held struct {
		Status          string `json:"status"`
		FixtureCount    int    `json:"fixture_count"`
		GenericOnly     bool   `json:"generic_only"`
		RequiredCapture struct {
			IndependentAcceptance bool   `json:"independent_acceptance"`
			SameFixtureSHA256     string `json:"same_fixture_sha256"`
			SameBenchmarkSHA256   string `json:"same_benchmark_sha256"`
			ScorerSHA256          string `json:"scorer_sha256"`
		} `json:"required_capture"`
	}
	heldBytes, err := os.ReadFile(filepath.Join(root, "B-v3-baseline-held.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(heldBytes, &held); err != nil {
		t.Fatalf("B-v3-baseline-held.json: %v", err)
	}
	if held.Status != "prerequisite_held" || held.FixtureCount != len(fixtures) || !held.GenericOnly || !held.RequiredCapture.IndependentAcceptance {
		t.Fatalf("baseline-held artifact is not a truthful gated prerequisite: %+v", held)
	}
	if held.RequiredCapture.SameFixtureSHA256 != sha256File(filepath.Join(root, "fixtures.v3.json")) || held.RequiredCapture.SameBenchmarkSHA256 != sha256File(filepath.Join(root, "benchmark.v3.json")) || held.RequiredCapture.ScorerSHA256 != sha256File(filepathFromRoot("cli/tools/structural-search-eval-v3/main.go")) {
		t.Fatalf("baseline-held provenance does not bind current v3 artifacts: %+v", held.RequiredCapture)
	}
	baselineBytes, err := os.ReadFile(filepath.Join(root, "B-v3-baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	var baseline V3BaselineArtifact
	if err := json.Unmarshal(baselineBytes, &baseline); err != nil {
		t.Fatalf("B-v3-baseline.json: %v", err)
	}
	if baseline.Status != "fresh" || baseline.Controller != "unchanged-c3-cmd.RunSearch" || !baseline.GenericOnly || baseline.CandidateSource || baseline.FixtureCount != len(fixtures) || len(baseline.Cases) != len(fixtures) {
		t.Fatalf("fresh baseline artifact is not a truthful unchanged capture: %#v", baseline)
	}
	for _, c := range baseline.Cases {
		if c.CanonicalRowBytes <= 0 || len(c.RowIDs) == 0 || len(c.ResultSHA256) != 64 {
			t.Fatalf("invalid fresh baseline case: %#v", c)
		}
	}
	var envelope struct {
		CaseIDs []string `json:"case_ids"`
	}
	if err := json.Unmarshal(benchmarkBytes, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.CaseIDs) != len(fixtures) {
		t.Fatalf("case_ids=%d, fixtures=%d", len(envelope.CaseIDs), len(fixtures))
	}
	for i, fixture := range fixtures {
		if envelope.CaseIDs[i] != fixture.CaseID {
			t.Fatalf("case_ids[%d]=%q, fixture=%q", i, envelope.CaseIDs[i], fixture.CaseID)
		}
	}
}
