package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/lagz0ne/c3-design/cli/cmd"
)

func TestV3FixtureRejectsUnboundCorpusEntity(t *testing.T) {
	f := sampleV3OwnerFixture()
	f.Entities = append(f.Entities, V3Entity{ID: "orphan", Role: RoleOwner})
	if err := ValidateFixture(f); err == nil {
		t.Fatal("unbound entity was accepted")
	}
}

func TestV3FixtureRejectsOverlappingRolePartition(t *testing.T) {
	f := sampleV3OwnerFixture()
	f.Oracle.UnsupportedEntityIDs = append(f.Oracle.UnsupportedEntityIDs, "owner")
	if err := ValidateFixture(f); err == nil {
		t.Fatal("overlapping role partition was accepted")
	}
}

func TestV3FixtureRejectsFactOutsideOracleRolePartition(t *testing.T) {
	f := sampleV3OwnerFixture()
	f.Oracle.NeutralFactIDs = nil
	if err := ValidateFixture(f); err == nil {
		t.Fatal("fact outside oracle role partition was accepted")
	}
}

func TestV3OwnerApplicabilityIsDerived(t *testing.T) {
	f := sampleV3OwnerFixture()
	f.Oracle.OwnerMetricsApplicable = boolPtr(false)
	if err := ValidateFixture(f); err == nil {
		t.Fatal("caller-supplied owner applicability overrode derivation")
	}
	f = sampleV3NoTargetFixture()
	if got := OwnerMetricsApplicable(f); got {
		t.Fatal("empty owner facts were marked applicable")
	}
}

func TestV3EmptyOwnerWithoutCompleteNoTargetPolicyRejects(t *testing.T) {
	f := sampleV3NoTargetFixture()
	f.Oracle.NoTargetPolicy = nil
	if err := ValidateFixture(f); err == nil {
		t.Fatal("empty owner case without policy was accepted")
	}
	f = sampleV3NoTargetFixture()
	f.Oracle.NoTargetPolicy.PreserveNeutralContext = false
	if err := ValidateFixture(f); err == nil {
		t.Fatal("no-target policy that drops neutral context was accepted")
	}
	f = sampleV3NoTargetFixture()
	f.Oracle.NoTargetPolicy.V2Action = "flag"
	if err := ValidateFixture(f); err == nil {
		t.Fatal("non-canonical v2 no-target action was accepted")
	}
	f = sampleV3NoTargetFixture()
	f.Oracle.NoTargetPolicy.V3Actions = []string{"omit", "flagged", "other"}
	if err := ValidateFixture(f); err == nil {
		t.Fatal("extra v3 no-target action was accepted")
	}
}

func TestV3RouteJSONFixtureAndTypedResponseAgree(t *testing.T) {
	f := sampleV3RouteFixture()
	fixtureBytes, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	var loaded V3Fixture
	if err := json.Unmarshal(fixtureBytes, &loaded); err != nil {
		t.Fatal(err)
	}
	resp := validV3Response(f)
	if _, err := ScoreV3([]V3Fixture{loaded}, resp, nil); err != nil {
		t.Fatalf("typed controller route fields did not match JSON fixture oracle: %v", err)
	}
}

func TestV3OwnerCaseCannotReportNAOwnerMetrics(t *testing.T) {
	f := sampleV3OwnerFixture()
	resp := validV3Response(f)
	resp.Cases[0].OwnerRecallAt5 = nil
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("owner case reported N/A owner metric")
	}
}

func TestV3NoTargetCaseReportsNAOwnerMetrics(t *testing.T) {
	f := sampleV3NoTargetFixture()
	resp := validV3Response(f)
	resp.Cases[0].OwnerRecallAt5 = nil
	resp.Cases[0].OwnerMRR = nil
	resp.Cases[0].StructuralOwnerPrecision = nil
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err != nil {
		t.Fatalf("no-target N/A metrics rejected: %v", err)
	}
}

func TestV3ScorerRejectsDuplicateTop5EntityIDs(t *testing.T) {
	f := sampleV3OwnerFixture()
	resp := validV3Response(f)
	resp.Cases[0].Rows = append(resp.Cases[0].Rows, resp.Cases[0].Rows[0])
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("duplicate top-five IDs were accepted")
	}
}

func TestV3PrecisionCountsUnknownForbiddenAndUnsupportedRows(t *testing.T) {
	f := sampleV3OwnerFixture()
	f.Entities = append(f.Entities,
		V3Entity{ID: "forbidden", Role: RoleForbidden},
		V3Entity{ID: "unsupported", Role: RoleUnsupported},
	)
	f.Facts = append(f.Facts,
		V3Fact{ID: "forbidden-fact", EntityID: "forbidden", Role: RoleForbidden},
		V3Fact{ID: "unsupported-fact", EntityID: "unsupported", Role: RoleUnsupported},
	)
	f.Oracle.ForbiddenFactIDs = []string{"forbidden-fact"}
	f.Oracle.UnsupportedEntityIDs = []string{"unsupported"}
	f.Oracle.UnsupportedFactIDs = []string{"unsupported-fact"}
	if err := ValidateFixture(f); err != nil {
		t.Fatal(err)
	}
	resp := V3Response{Schema: ResponseSchemaV3, Cases: []V3CaseResponse{{CaseID: f.CaseID, Rows: []cmd.SearchResultRow{{ID: "owner"}, {ID: "forbidden"}, {ID: "unsupported"}, {ID: "unknown"}}}}}
	report, err := ScoreV3([]V3Fixture{f}, resp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Cases[0].StructuralOwnerPrecision == nil || *report.Cases[0].StructuralOwnerPrecision != 0.25 {
		t.Fatalf("precision denominator did not include non-neutral rows: %#v", report.Cases[0].StructuralOwnerPrecision)
	}
	if report.Cases[0].UnsupportedCompetitorCount != 2 {
		t.Fatalf("unknown/unsupported count=%d", report.Cases[0].UnsupportedCompetitorCount)
	}
}

func TestV3FlaggedDispositionMetadataOmitsForbiddenRows(t *testing.T) {
	f := sampleV3NoTargetFixture()
	resp := validV3Response(f)
	resp.Cases[0].Disposition = "flagged"
	resp.Cases[0].GenericReason = "generic forbidden evidence omitted"
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err != nil {
		t.Fatalf("valid flagged metadata rejected: %v", err)
	}
	resp.Cases[0].Rows = append(resp.Cases[0].Rows, cmd.SearchResultRow{ID: "forbidden"})
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("flagged response exposed forbidden row")
	}
}

func TestV3RouteBoundSetAndWitnessAreOneToOne(t *testing.T) {
	f := sampleV3RouteFixture()
	if err := ValidateFixture(f); err != nil {
		t.Fatal(err)
	}
	f.Oracle.BoundRouteEntityIDs = []string{"route-owner", "other"}
	if err := ValidateFixture(f); err == nil {
		t.Fatal("multiple bound routes accepted for singleton witness")
	}
}

func TestV3RouteRejectsWrongEntityContentID(t *testing.T) {
	f := sampleV3RouteFixture()
	resp := validV3Response(f)
	resp.Cases[0].RouteWitnesses[0].EntityContentID = "content:wrong"
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("wrong route content ID earned credit")
	}
}

func TestV3RouteRejectsWrongGraphEdge(t *testing.T) {
	f := sampleV3RouteFixture()
	resp := validV3Response(f)
	resp.Cases[0].RouteWitnesses[0].GraphToID = "wrong"
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("wrong route graph edge earned credit")
	}
}

func TestV3RouteRejectsWrongDirectFTSMissContentID(t *testing.T) {
	f := sampleV3RouteFixture()
	resp := validV3Response(f)
	resp.Cases[0].RouteWitnesses[0].DirectFTSContentMissID = "content:wrong"
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("wrong direct FTS miss content ID earned credit")
	}
}

func TestV3RouteFieldCoverageRejectsArbitraryNonEmptyValues(t *testing.T) {
	f := sampleV3RouteFixture()
	resp := validV3Response(f)
	resp.Cases[0].Rows[0].Route.Facts = []string{"arbitrary"}
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("arbitrary non-empty route field earned credit")
	}
}

func TestV3RouteRequiresControllerObservedDirectFTSMiss(t *testing.T) {
	f := sampleV3RouteFixture()
	f.Oracle.DirectFTSMissWitness.ProbeEntityHitIDs = []string{"route-owner"}
	if err := ValidateFixture(f); err == nil {
		t.Fatal("route owner direct FTS hit was accepted as a miss")
	}
}

func TestV3RowBytesUseSameCaseFreshBaseline(t *testing.T) {
	f := sampleV3OwnerFixture()
	resp := validV3Response(f)
	if _, err := ScoreV3([]V3Fixture{f}, resp, map[string]int{"owner-case": 0}); err == nil {
		t.Fatal("zero fresh baseline bytes were accepted")
	}
	if _, err := ScoreV3([]V3Fixture{f}, resp, map[string]int{"owner-case": canonicalBytes(resp.Cases[0].Rows)}); err != nil {
		t.Fatalf("valid same-case bytes rejected: %v", err)
	}
}

func TestV3PrivacyBindingAndGenericOnlyScan(t *testing.T) {
	b := sampleV3Benchmark()
	if err := ValidateBenchmark(b); err != nil {
		t.Fatal(err)
	}
	b.Privacy.ScanScope = nil
	if err := ValidateBenchmark(b); err == nil {
		t.Fatal("incomplete privacy scan scope accepted")
	}
	b = sampleV3Benchmark()
	b.Privacy.PolicySHA256 = "not-a-sha256"
	if err := ValidateBenchmark(b); err == nil {
		t.Fatal("malformed privacy hash accepted")
	}
}

func TestV3RouteFixtureRequiresCanonicalSourceAndFields(t *testing.T) {
	f := sampleV3RouteFixture()
	f.Oracle.BoundRouteWitnesses[0].MatchSource = "arbitrary"
	if err := ValidateFixture(f); err == nil {
		t.Fatal("non-canonical route source was accepted")
	}
	f = sampleV3RouteFixture()
	f.Oracle.RequiredRouteFields = []string{"facts", "missing"}
	if err := ValidateFixture(f); err == nil {
		t.Fatal("route field without oracle value was accepted")
	}
	f = sampleV3RouteFixture()
	f.Oracle.DirectFTSMissWitness.CaseMiss = false
	if err := ValidateFixture(f); err == nil {
		t.Fatal("route without a direct-FTS case miss was accepted")
	}
}

func TestV3FlaggedDispositionRejectsForbiddenRowsOutsideTopFive(t *testing.T) {
	f := sampleV3NoTargetFixture()
	resp := validV3Response(f)
	resp.Cases[0].Disposition = "flagged"
	resp.Cases[0].GenericReason = "generic review required"
	resp.Cases[0].Rows = []cmd.SearchResultRow{{ID: "neutral"}, {ID: "context-1"}, {ID: "context-2"}, {ID: "context-3"}, {ID: "context-4"}, {ID: "forbidden"}}
	if _, err := ScoreV3([]V3Fixture{f}, resp, nil); err == nil {
		t.Fatal("flagged response exposed forbidden row after top five")
	}
}

func TestFrozenV2ReplayHashesUnchanged(t *testing.T) {
	if got := sha256File(filepathFromRoot("research/eval/structural-retrieval/benchmark.v2.json")); got != "b960525cc42216e6598452946da5fb68735bbf989f311f170cedcfdbe92bf0d5" {
		t.Fatalf("v2 benchmark changed: %s", got)
	}
}

func boolPtr(v bool) *bool { return &v }

func validV3Response(f V3Fixture) V3Response {
	rows := []cmd.SearchResultRow{{ID: "owner"}}
	if f.Oracle.BoundRouteEntityIDs != nil {
		rows = []cmd.SearchResultRow{{ID: "route-owner", MatchSources: []string{"graph:uses:route-anchor"}, Route: cmd.RouteEnrichment{Facts: []string{"route-owner-fact"}, Graph: []string{"graph:uses:route-anchor"}, Lanes: []string{"behavioral-route"}, Hash: "route:route-owner:route-anchor"}}}
	}
	r := V3CaseResponse{CaseID: f.CaseID, Rows: rows, Disposition: "omit"}
	if OwnerMetricsApplicable(f) {
		one := 1.0
		r.OwnerRecallAt5, r.OwnerMRR, r.StructuralOwnerPrecision = &one, &one, &one
	}
	if f.Oracle.BoundRouteEntityIDs != nil {
		r.RouteWitnesses = []V3RouteResponseWitness{{EntityID: "route-owner", EntityContentID: "content:route-owner", MatchSource: "graph:uses:route-anchor", GraphFromID: "route-owner", GraphRelType: "uses", GraphToID: "route-anchor", DirectFTSEntityMissID: "route-owner", DirectFTSContentMissID: "content:route-owner", RouteFieldValues: map[string]any{"facts": []string{"route-owner-fact"}, "graph": []string{"graph:uses:route-anchor"}, "lanes": []string{"behavioral-route"}, "hash": "route:route-owner:route-anchor"}}}
	}
	return V3Response{Schema: ResponseSchemaV3, Cases: []V3CaseResponse{r}}
}

func sampleV3OwnerFixture() V3Fixture {
	return V3Fixture{Schema: FixtureSchemaV3, CaseID: "owner-case", Query: "owner evidence", Entities: []V3Entity{{ID: "owner", Role: RoleOwner, Type: "component", Title: "Owner", Slug: "owner", Status: "active", Metadata: "{}", Markdown: "# Owner\n\nowner evidence"}, {ID: "neutral", Role: RoleNeutral, Type: "component", Title: "Neutral", Slug: "neutral", Status: "active", Metadata: "{}", Markdown: "# Neutral\n\nneutral context"}}, Facts: []V3Fact{{ID: "owner-fact", EntityID: "owner", Role: RoleOwner}, {ID: "neutral-fact", EntityID: "neutral", Role: RoleNeutral}}, Oracle: V3Oracle{RequiredOwnerFactIDs: []string{"owner-fact"}, NeutralFactIDs: []string{"neutral-fact"}, RequiredContextEntityIDs: []string{"neutral"}}}
}

func sampleV3NoTargetFixture() V3Fixture {
	f := sampleV3OwnerFixture()
	f.CaseID = "no-target"
	f.Oracle.RequiredOwnerFactIDs = nil
	f.Entities = []V3Entity{f.Entities[1]}
	f.Facts = []V3Fact{f.Facts[1]}
	f.Entities = append(f.Entities, V3Entity{ID: "forbidden", Role: RoleForbidden, Type: "component", Title: "Forbidden", Slug: "forbidden", Status: "active", Markdown: "# Forbidden"})
	f.Facts = append(f.Facts, V3Fact{ID: "forbidden-fact", EntityID: "forbidden", Role: RoleForbidden})
	f.Oracle.ForbiddenFactIDs = []string{"forbidden-fact"}
	f.Oracle.NoTargetPolicy = &V3NoTargetPolicy{OwnerWitnessCount: 0, PreserveNeutralContext: true, V2Action: "omit", V3Actions: []string{"omit", "flagged"}}
	return f
}

func sampleV3RouteFixture() V3Fixture {
	return V3Fixture{
		Schema: FixtureSchemaV3, CaseID: "route-case", Query: "route anchor",
		Entities: []V3Entity{
			{ID: "route-owner", Role: RoleOwner, Type: "component", Title: "Route Owner", Slug: "route-owner", Status: "active", Markdown: "# Route Owner\n\nroute anchor"},
			{ID: "route-anchor", Role: RoleNeutral, Type: "component", Title: "Route Anchor", Slug: "route-anchor", Status: "active", Markdown: "# Route Anchor\n\nanchor"},
			{ID: "route-unbound", Role: RoleUnsupported, Type: "component", Title: "Route Unbound", Slug: "route-unbound", Status: "active", Markdown: "# Route Unbound\n\nnoise"},
		},
		Relationships: []V3Relationship{{FromID: "route-owner", ToID: "route-anchor", RelType: "uses"}},
		Facts: []V3Fact{
			{ID: "route-owner-fact", EntityID: "route-owner", Role: RoleOwner},
			{ID: "route-anchor-fact", EntityID: "route-anchor", Role: RoleNeutral},
			{ID: "route-unbound-fact", EntityID: "route-unbound", Role: RoleUnsupported},
		},
		Oracle: V3Oracle{
			RequiredOwnerFactIDs: []string{"route-owner-fact"}, NeutralFactIDs: []string{"route-anchor-fact"}, RequiredContextEntityIDs: []string{"route-anchor"},
			UnsupportedEntityIDs: []string{"route-unbound"}, UnsupportedFactIDs: []string{"route-unbound-fact"}, BoundRouteEntityIDs: []string{"route-owner"},
			BoundRouteWitnesses: []V3RouteWitness{{
				EntityID: "route-owner", EntityContentID: "content:route-owner", MatchSource: "graph:uses:route-anchor", GraphFromID: "route-owner", GraphRelType: "uses", GraphToID: "route-anchor", DirectFTSEntityMissID: "route-owner", DirectFTSContentMissID: "content:route-owner",
				ExpectedRouteFieldValues: map[string]any{"facts": []string{"route-owner-fact"}, "graph": []string{"graph:uses:route-anchor"}, "lanes": []string{"behavioral-route"}, "hash": "route:route-owner:route-anchor"},
			}},
			DirectFTSMissWitness: V3DirectFTSMissWitness{ProbeEntityHitIDs: []string{}, ProbeContentHitIDs: []string{}, ExpectedBoundRouteEntityIDs: []string{"route-owner"}, ExpectedBoundRouteEntityMissIDs: []string{"route-owner"}, ExpectedBoundRouteContentMissIDs: []string{"content:route-owner"}, CaseMiss: true},
			RelationshipWitness:  &V3RelationshipWitness{ExpectedEntityID: "route-owner", FromID: "route-owner", ToID: "route-anchor", RelType: "uses", ExpectedMatchSource: "graph:uses:route-anchor"},
			RequiredRouteFields:  []string{"facts", "graph", "lanes", "hash"},
		},
	}
}

func sampleV3Benchmark() V3Benchmark {
	return V3Benchmark{Schema: BenchmarkSchemaV3, FixtureCount: 7, K: 5, SemanticMode: "no-semantic", Thresholds: V3Thresholds{OwnerRecallAt5Delta: 0.2, StructuralOwnerPrecision: 0.8, CanonicalRowBytesRatio: 1.05}, Privacy: V3PrivacyBinding{PolicySHA256: "7f038b3eef110d2765aa1aa003003fc19e0b390a3623a6978712e813831d7055", ScannerSourceSHA256: "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e", DetectorDefinitionSHA256: "e9f28b8fe0dafd3effa283acba95e4625b3de38bc1d80615e8d771042edfa8e7", ScanScope: []string{"v3 fixture", "v3 benchmark", "v3 scorer", "fresh B-v3 artifacts", "review records", "retained generic output"}}}
}

var _ = json.Valid
var _ = reflect.DeepEqual
