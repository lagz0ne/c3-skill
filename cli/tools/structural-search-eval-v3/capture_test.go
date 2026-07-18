package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestFreshBaselineUsesRealRunSearchAndReturnsGenericRows(t *testing.T) {
	f := sampleV3OwnerFixture()
	artifact, err := CaptureFreshBaseline([]V3Fixture{f})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Status != "fresh" || artifact.Controller != "unchanged-c3-cmd.RunSearch" || !artifact.GenericOnly || artifact.CandidateSource {
		t.Fatalf("unexpected baseline envelope: %#v", artifact)
	}
	if len(artifact.Cases) != 1 || artifact.Cases[0].CanonicalRowBytes <= 0 || len(artifact.Cases[0].RowIDs) == 0 {
		t.Fatalf("fresh baseline has no usable rows: %#v", artifact.Cases)
	}
}

func TestFreshBaselineRejectsDuplicateCaseIDs(t *testing.T) {
	f := sampleV3OwnerFixture()
	if _, err := CaptureFreshBaseline([]V3Fixture{f, f}); err == nil {
		t.Fatal("duplicate fixture IDs accepted")
	}
}

func TestFreshBaselineForcesJSONOutsideAgentOutputMode(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	if _, err := CaptureFreshBaseline([]V3Fixture{sampleV3OwnerFixture()}); err != nil {
		t.Fatalf("capture depended on ambient agent output mode: %v", err)
	}
}

func TestV3CLIRequiresReadOnlyCaptureInputs(t *testing.T) {
	if err := runV3CLI(nil); err == nil {
		t.Fatal("missing capture command accepted")
	}
}

func TestWriteV3BaselineUsesPrivateMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	artifact := V3BaselineArtifact{Schema: baselineSchemaV3, Status: "fresh", Controller: "unchanged-c3-cmd.RunSearch", FixtureSHA256: strings.Repeat("a", 64), BenchmarkSHA256: strings.Repeat("b", 64), ScorerSHA256: strings.Repeat("c", 64), FixtureCount: 1, GenericOnly: true, RawRowsRetained: false, CandidateSource: false, Privacy: V3PrivacyBinding{PolicySHA256: strings.Repeat("d", 64), ScannerSourceSHA256: strings.Repeat("e", 64), DetectorDefinitionSHA256: strings.Repeat("f", 64), ScanScope: []string{"generic"}}, Cases: []V3BaselineCase{{CaseID: "x", CanonicalRowBytes: 1, ResultSHA256: strings.Repeat("1", 64)}}}
	if err := WriteV3Baseline(path, artifact); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("baseline mode=%o, want 600", info.Mode().Perm())
	}
	if err := WriteV3Baseline(path, artifact); err == nil {
		t.Fatal("existing baseline output was overwritten")
	}
}

func TestInsertV3EntitiesReadsOptionalGenericParentMetadata(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "c3.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	entities := []V3Entity{
		{ID: "parent", Type: "container", Metadata: "{}"},
		{ID: "child", Type: "component", Metadata: `{"parent_id":"parent"}`},
	}
	if err := insertV3Entities(s, entities); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetEntity("child")
	if err != nil {
		t.Fatal(err)
	}
	if got.ParentID != "parent" {
		t.Fatalf("parent id = %q, want parent", got.ParentID)
	}
	if parentID(V3Entity{ID: "frozen", Metadata: "{}"}) != "" {
		t.Fatal("frozen empty metadata must not invent a parent")
	}
}

func TestCandidateCaptureIsExplicitAndMarksGenericSource(t *testing.T) {
	artifact, err := CaptureCandidateV3Artifact([]V3Fixture{sampleV3NoTargetFixture()})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Status != "candidate" || !artifact.CandidateSource || !artifact.GenericOnly || artifact.RawRowsRetained {
		t.Fatalf("unexpected candidate envelope: %#v", artifact)
	}
	if artifact.Response.Schema != ResponseSchemaV3 || len(artifact.Response.Cases) != 1 {
		t.Fatalf("candidate response missing: %#v", artifact.Response)
	}
}

func TestCandidateNoTargetProjectionIsTruthfullyFlagged(t *testing.T) {
	response, err := CaptureCandidateV3([]V3Fixture{sampleV3NoTargetFixture()})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Cases) != 1 || (response.Cases[0].Disposition != "omit" && response.Cases[0].Disposition != "flagged") || len(response.Cases[0].Rows) != 0 {
		t.Fatalf("no-target response was not generic omit/flagged: %#v", response.Cases)
	}
}

func TestCandidateRouteWitnessConversionIsTypedAndOneToOne(t *testing.T) {
	response, err := candidateCaseFromOutput("route-case", cmd.SearchOutput{
		Results:        []cmd.SearchResultRow{{ID: "route-owner", MatchSources: []string{"graph:uses:route-anchor"}, Route: cmd.RouteEnrichment{Facts: []string{"route-owner-fact"}, Graph: []string{"graph:uses:route-anchor"}, Lanes: []string{"behavioral-route"}, Hash: "route:route-owner:route-anchor"}}},
		RouteWitnesses: []cmd.SearchRouteWitness{{EntityID: "route-owner", EntityContentIDs: []string{"content:route-owner"}, MatchSource: "graph:uses:route-anchor", GraphFromID: "route-owner", GraphRelType: "uses", GraphToID: "route-anchor", DirectFTSEntityMissID: "route-owner", DirectFTSContentMissIDs: []string{"content:route-owner"}, RouteFieldValues: cmd.RouteEnrichment{Facts: []string{"route-owner-fact"}, Graph: []string{"graph:uses:route-anchor"}, Lanes: []string{"behavioral-route"}, Hash: "route:route-owner:route-anchor"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	witnesses := response.RouteWitnesses
	if len(witnesses) != 1 || witnesses[0].EntityID != "route-owner" || witnesses[0].DirectFTSContentMissID != "content:route-owner" {
		t.Fatalf("route witness was not converted: %#v", witnesses)
	}
	if _, err := candidateCaseFromOutput("duplicate", cmd.SearchOutput{Results: []cmd.SearchResultRow{{ID: "same"}, {ID: "same"}}}); err == nil {
		t.Fatal("duplicate candidate rows accepted")
	}
}

func TestCandidatePrivacyDropsRawMarkdownFields(t *testing.T) {
	response, err := candidateCaseFromOutput("privacy", cmd.SearchOutput{Results: []cmd.SearchResultRow{{ID: "entity", Title: "SECRET_MARKDOWN", Snippet: "SECRET_MARKDOWN"}}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "SECRET_MARKDOWN") {
		t.Fatalf("candidate envelope retained raw markdown: %s", encoded)
	}
}
