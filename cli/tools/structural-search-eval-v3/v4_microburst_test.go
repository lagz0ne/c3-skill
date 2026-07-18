package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// TestV4PairedMicroburstArtifact is opt-in so ordinary evaluator tests never
// activate the candidate path. It replays the generic v4 containment pair and
// checks the retained aggregate without retaining raw rows.
func TestV4PairedMicroburstArtifact(t *testing.T) {
	if os.Getenv("RUN_V4_MICROBURST") != "1" {
		t.Skip("set RUN_V4_MICROBURST=1 to run the explicit candidate pair")
	}
	root := filepathFromRoot("research/eval/structural-retrieval-v4")
	fixtures, _, err := LoadV3FixtureFile(filepath.Join(root, "fixtures.v4.json"))
	if err != nil {
		t.Fatal(err)
	}
	var baselineArtifact V3BaselineArtifact
	data, err := os.ReadFile(filepath.Join(root, "B-v4-baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &baselineArtifact); err != nil {
		t.Fatal(err)
	}
	base := map[string]int{}
	for _, c := range baselineArtifact.Cases {
		base[c.CaseID] = c.CanonicalRowBytes
	}
	baselineResponse := V3Response{Schema: ResponseSchemaV3}
	for _, fixture := range fixtures {
		capture, err := captureV3Case(fixture)
		if err != nil {
			t.Fatal(err)
		}
		baselineResponse.Cases = append(baselineResponse.Cases, V3CaseResponse{CaseID: fixture.CaseID, Rows: capture.Rows, Disposition: "omit"})
	}
	baselineReport, err := ScoreV3(fixtures, baselineResponse, base)
	if err != nil {
		t.Fatal(err)
	}
	candidateResponse, err := CaptureCandidateV3(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	candidateReport, err := ScoreV3(fixtures, candidateResponse, base)
	if err != nil {
		t.Fatal(err)
	}
	var retained struct {
		Aggregate struct {
			BaselineRecall  float64 `json:"baseline_owner_recall_at_5"`
			CandidateRecall float64 `json:"candidate_owner_recall_at_5"`
			Delta           float64 `json:"owner_recall_at_5_delta"`
			BaselineMRR     float64 `json:"baseline_owner_mrr"`
			CandidateMRR    float64 `json:"candidate_owner_mrr"`
		} `json:"aggregate"`
	}
	data, err = os.ReadFile(filepath.Join(root, "paired-microburst.v4.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &retained); err != nil {
		t.Fatal(err)
	}
	if !closeEnough(retained.Aggregate.BaselineRecall, baselineReport.Metrics.OwnerRecallAt5) ||
		!closeEnough(retained.Aggregate.CandidateRecall, candidateReport.Metrics.OwnerRecallAt5) ||
		!closeEnough(retained.Aggregate.Delta, candidateReport.Metrics.OwnerRecallAt5-baselineReport.Metrics.OwnerRecallAt5) ||
		!closeEnough(retained.Aggregate.BaselineMRR, baselineReport.Metrics.OwnerMRR) ||
		!closeEnough(retained.Aggregate.CandidateMRR, candidateReport.Metrics.OwnerMRR) {
		t.Fatalf("retained aggregate does not replay: retained=%+v baseline=%+v candidate=%+v", retained.Aggregate, baselineReport.Metrics, candidateReport.Metrics)
	}
}

func closeEnough(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
