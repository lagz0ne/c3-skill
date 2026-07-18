// Package main contains the small, fail-closed scorer for the v3 structural
// retrieval benchmark.  It deliberately has no controller or filesystem
// dependencies: the controller owns fixture construction and passes only the
// redacted response and fresh per-case byte baselines to this package.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/cmd"
)

const (
	FixtureSchemaV3           = "structural-retrieval-fixture.v3"
	BenchmarkSchemaV3         = "structural-retrieval-benchmark.v3"
	ResponseSchemaV3          = "structural-retrieval-response.v3"
	RoleOwner                 = "owner"
	RoleNeutral               = "neutral"
	RoleForbidden             = "forbidden"
	RoleUnsupported           = "unsupported"
	RoleUnknown               = "unknown"
	maxCanonicalRowBytesRatio = 1.05
)

// The exported names are part of the generic fixture/response contract.  The
// oracle is intentionally more explicit than v2: a structural entity, a
// route witness, and a direct-FTS miss each have separate evidence bindings.
type V3Entity struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Slug     string `json:"slug,omitempty"`
	Status   string `json:"status,omitempty"`
	Metadata string `json:"metadata,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

type V3Relationship struct {
	FromID  string `json:"from_id"`
	ToID    string `json:"to_id"`
	RelType string `json:"rel_type"`
}
type V3Fact struct {
	ID       string `json:"id"`
	EntityID string `json:"entity_id"`
	Role     string `json:"role"`
}

type V3NoTargetPolicy struct {
	OwnerWitnessCount      int      `json:"owner_witness_count"`
	PreserveNeutralContext bool     `json:"preserve_neutral_context"`
	V2Action               string   `json:"v2_action"`
	V3Actions              []string `json:"v3_actions"`
}

type V3RouteWitness struct {
	EntityID                 string         `json:"entity_id"`
	EntityContentID          string         `json:"entity_content_id"`
	MatchSource              string         `json:"match_source"`
	GraphFromID              string         `json:"graph_from_id"`
	GraphRelType             string         `json:"graph_rel_type"`
	GraphToID                string         `json:"graph_to_id"`
	DirectFTSEntityMissID    string         `json:"direct_fts_entity_miss_id"`
	DirectFTSContentMissID   string         `json:"direct_fts_content_miss_id"`
	ExpectedRouteFieldValues map[string]any `json:"expected_route_field_values"`
}

type V3DirectFTSMissWitness struct {
	ProbeEntityHitIDs                []string `json:"probe_entity_hit_ids"`
	ProbeContentHitIDs               []string `json:"probe_content_hit_ids"`
	ExpectedBoundRouteEntityIDs      []string `json:"expected_bound_route_entity_ids"`
	ExpectedBoundRouteEntityMissIDs  []string `json:"expected_bound_route_entity_miss_ids"`
	ExpectedBoundRouteContentMissIDs []string `json:"expected_bound_route_content_miss_ids"`
	CaseMiss                         bool     `json:"case_miss"`
}

type V3RelationshipWitness struct {
	ExpectedEntityID    string `json:"expected_entity_id"`
	FromID              string `json:"from_id"`
	ToID                string `json:"to_id"`
	RelType             string `json:"rel_type"`
	ExpectedMatchSource string `json:"expected_match_source"`
}

type V3Oracle struct {
	RequiredOwnerFactIDs     []string               `json:"required_owner_fact_ids"`
	NeutralFactIDs           []string               `json:"neutral_fact_ids"`
	RequiredContextEntityIDs []string               `json:"required_context_entity_ids"`
	ForbiddenFactIDs         []string               `json:"forbidden_fact_ids"`
	UnsupportedEntityIDs     []string               `json:"unsupported_entity_ids"`
	UnsupportedFactIDs       []string               `json:"unsupported_fact_ids"`
	UnknownEntityIDs         []string               `json:"unknown_entity_ids,omitempty"`
	UnknownFactIDs           []string               `json:"unknown_fact_ids,omitempty"`
	BoundRouteEntityIDs      []string               `json:"bound_route_entity_ids,omitempty"`
	BoundRouteWitnesses      []V3RouteWitness       `json:"bound_route_witnesses,omitempty"`
	DirectFTSMissWitness     V3DirectFTSMissWitness `json:"direct_fts_miss_witness"`
	RelationshipWitness      *V3RelationshipWitness `json:"relationship_witness,omitempty"`
	RequiredRouteFields      []string               `json:"required_route_fields,omitempty"`
	NoTargetPolicy           *V3NoTargetPolicy      `json:"no_target_policy,omitempty"`
	OwnerMetricsApplicable   *bool                  `json:"owner_metrics_applicable,omitempty"`
}

type V3Fixture struct {
	Schema        string           `json:"$schema"`
	CaseID        string           `json:"case_id"`
	Query         string           `json:"query"`
	Entities      []V3Entity       `json:"entities"`
	Relationships []V3Relationship `json:"relationships,omitempty"`
	Facts         []V3Fact         `json:"facts"`
	Oracle        V3Oracle         `json:"oracle"`
}

type V3Thresholds struct {
	OwnerRecallAt5Delta      float64 `json:"owner_recall_at_5_delta"`
	StructuralOwnerPrecision float64 `json:"structural_owner_precision"`
	CanonicalRowBytesRatio   float64 `json:"canonical_row_bytes_ratio"`
}
type V3PrivacyBinding struct {
	PolicySHA256             string   `json:"policy_sha256"`
	ScannerSourceSHA256      string   `json:"scanner_source_sha256"`
	DetectorDefinitionSHA256 string   `json:"detector_definition_sha256"`
	ScanScope                []string `json:"scan_scope"`
}
type V3Benchmark struct {
	Schema       string           `json:"$schema"`
	FixtureCount int              `json:"fixture_count"`
	K            int              `json:"k"`
	SemanticMode string           `json:"semantic_mode"`
	Thresholds   V3Thresholds     `json:"thresholds"`
	Privacy      V3PrivacyBinding `json:"privacy"`
}

type V3RouteResponseWitness struct {
	EntityID               string         `json:"entity_id"`
	EntityContentID        string         `json:"entity_content_id"`
	MatchSource            string         `json:"match_source"`
	GraphFromID            string         `json:"graph_from_id"`
	GraphRelType           string         `json:"graph_rel_type"`
	GraphToID              string         `json:"graph_to_id"`
	DirectFTSEntityMissID  string         `json:"direct_fts_entity_miss_id"`
	DirectFTSContentMissID string         `json:"direct_fts_content_miss_id"`
	RouteFieldValues       map[string]any `json:"route_field_values"`
}
type V3CaseResponse struct {
	CaseID                   string                   `json:"case_id"`
	Rows                     []cmd.SearchResultRow    `json:"rows"`
	Disposition              string                   `json:"disposition,omitempty"`
	GenericReason            string                   `json:"generic_reason,omitempty"`
	OwnerRecallAt5           *float64                 `json:"owner_recall_at_5,omitempty"`
	OwnerMRR                 *float64                 `json:"owner_mrr,omitempty"`
	StructuralOwnerPrecision *float64                 `json:"structural_owner_precision,omitempty"`
	RouteWitnesses           []V3RouteResponseWitness `json:"route_witnesses,omitempty"`
}
type V3Response struct {
	Schema string           `json:"$schema"`
	Cases  []V3CaseResponse `json:"cases"`
}

type V3CaseMetrics struct {
	CaseID                            string   `json:"case_id"`
	OwnerRecallAt1                    *float64 `json:"owner_recall_at_1,omitempty"`
	OwnerRecallAt3                    *float64 `json:"owner_recall_at_3,omitempty"`
	OwnerRecallAt5                    *float64 `json:"owner_recall_at_5,omitempty"`
	OwnerMRR                          *float64 `json:"owner_mrr,omitempty"`
	StructuralOwnerPrecision          *float64 `json:"structural_owner_precision,omitempty"`
	ForbiddenStructuralRetrievalCount int      `json:"forbidden_structural_retrieval_count"`
	UnsupportedCompetitorCount        int      `json:"unsupported_competitor_count"`
	CanonicalRowBytes                 int      `json:"canonical_row_bytes"`
	RouteRecallAt5                    float64  `json:"route_recall_at_5"`
	RouteMRR                          float64  `json:"route_mrr"`
	Flagged                           bool     `json:"flagged"`
}
type V3Metrics struct {
	OwnerRecallAt5                    float64        `json:"owner_recall_at_5"`
	OwnerMRR                          float64        `json:"owner_mrr"`
	StructuralOwnerPrecision          float64        `json:"structural_owner_precision"`
	ForbiddenStructuralRetrievalCount int            `json:"forbidden_structural_retrieval_count"`
	UnsupportedCompetitorCount        int            `json:"unsupported_competitor_count"`
	RouteRecallAt5                    float64        `json:"route_recall_at_5"`
	RouteMRR                          float64        `json:"route_mrr"`
	CanonicalRowBytesPerCase          map[string]int `json:"canonical_row_bytes_per_case"`
}
type V3Report struct {
	Schema  string          `json:"$schema"`
	Cases   []V3CaseMetrics `json:"cases"`
	Metrics V3Metrics       `json:"metrics"`
}

// OwnerMetricsApplicable is derived from the oracle, never trusted from an
// input field.  A case with no required owner facts is a no-target case.
func OwnerMetricsApplicable(f V3Fixture) bool { return len(f.Oracle.RequiredOwnerFactIDs) > 0 }

func ValidateFixture(f V3Fixture) error {
	if f.Schema != FixtureSchemaV3 {
		return fmt.Errorf("fixture schema mismatch")
	}
	if strings.TrimSpace(f.CaseID) == "" || strings.TrimSpace(f.Query) == "" {
		return errors.New("fixture case_id and query are required")
	}
	entities := map[string]V3Entity{}
	for _, e := range f.Entities {
		if strings.TrimSpace(e.ID) == "" {
			return errors.New("entity id is required")
		}
		if _, ok := entities[e.ID]; ok {
			return fmt.Errorf("duplicate entity %s", e.ID)
		}
		if !validRole(e.Role) {
			return fmt.Errorf("invalid entity role %q", e.Role)
		}
		entities[e.ID] = e
	}
	facts := map[string]V3Fact{}
	for _, fact := range f.Facts {
		if fact.ID == "" || fact.EntityID == "" {
			return errors.New("fact id/entity_id required")
		}
		if _, ok := facts[fact.ID]; ok {
			return fmt.Errorf("duplicate fact %s", fact.ID)
		}
		e, ok := entities[fact.EntityID]
		if !ok {
			return fmt.Errorf("fact %s references unknown entity", fact.ID)
		}
		if !validRole(fact.Role) || fact.Role != e.Role {
			return fmt.Errorf("fact %s role does not match entity", fact.ID)
		}
		facts[fact.ID] = fact
	}
	for _, r := range f.Relationships {
		if _, ok := entities[r.FromID]; !ok {
			return fmt.Errorf("relationship from unknown entity %s", r.FromID)
		}
		if _, ok := entities[r.ToID]; !ok {
			return fmt.Errorf("relationship to unknown entity %s", r.ToID)
		}
	}
	sets := []struct {
		name string
		ids  []string
		role string
		fact bool
	}{
		{"required owner facts", f.Oracle.RequiredOwnerFactIDs, RoleOwner, true},
		{"neutral facts", f.Oracle.NeutralFactIDs, RoleNeutral, true},
		{"forbidden facts", f.Oracle.ForbiddenFactIDs, RoleForbidden, true},
		{"unsupported facts", f.Oracle.UnsupportedFactIDs, RoleUnsupported, true},
		{"unknown facts", f.Oracle.UnknownFactIDs, RoleUnknown, true},
		{"unsupported entities", f.Oracle.UnsupportedEntityIDs, RoleUnsupported, false},
		{"unknown entities", f.Oracle.UnknownEntityIDs, RoleUnknown, false},
		{"required context entities", f.Oracle.RequiredContextEntityIDs, "", false},
	}
	seen := map[string]string{}
	for _, s := range sets {
		for _, id := range s.ids {
			if id == "" {
				return fmt.Errorf("empty %s id", s.name)
			}
			if old, ok := seen[id]; ok {
				return fmt.Errorf("oracle id %s overlaps %s and %s", id, old, s.name)
			}
			seen[id] = s.name
			if s.fact {
				fact, ok := facts[id]
				if !ok {
					return fmt.Errorf("%s references unknown fact %s", s.name, id)
				}
				if fact.Role != s.role {
					return fmt.Errorf("%s fact %s has role %s", s.name, id, fact.Role)
				}
			} else {
				e, ok := entities[id]
				if !ok {
					return fmt.Errorf("%s references unknown entity %s", s.name, id)
				}
				if s.name == "required context entities" && e.Role != RoleNeutral {
					return fmt.Errorf("required context entity %s has role %s", id, e.Role)
				}
				if s.role != "" && e.Role != s.role {
					return fmt.Errorf("%s entity %s has role %s", s.name, id, e.Role)
				}
			}
		}
	}
	// A fact cannot be merely present in the corpus: every fact must be in
	// exactly one oracle role partition so that no returned row can escape the
	// denominator or negative-count audit.
	partitionedFacts := map[string]bool{}
	for _, s := range sets {
		if !s.fact {
			continue
		}
		for _, id := range s.ids {
			partitionedFacts[id] = true
		}
	}
	for id := range facts {
		if !partitionedFacts[id] {
			return fmt.Errorf("fact %s is not bound by oracle role partition", id)
		}
	}
	if f.Oracle.OwnerMetricsApplicable != nil {
		return errors.New("owner_metrics_applicable is derived and must be omitted")
	}
	// Every corpus entity must have an oracle/fact binding.  This prevents an
	// arm from receiving unscored distractors which cannot be audited later.
	boundEntities := map[string]bool{}
	for _, fact := range f.Facts {
		boundEntities[fact.EntityID] = true
	}
	for _, id := range f.Oracle.RequiredContextEntityIDs {
		boundEntities[id] = true
	}
	for _, id := range f.Oracle.UnsupportedEntityIDs {
		boundEntities[id] = true
	}
	for _, id := range f.Oracle.UnknownEntityIDs {
		boundEntities[id] = true
	}
	for _, w := range f.Oracle.BoundRouteWitnesses {
		boundEntities[w.EntityID] = true
	}
	for id := range entities {
		if !boundEntities[id] {
			return fmt.Errorf("entity %s is not bound by corpus oracle", id)
		}
	}
	if len(f.Oracle.BoundRouteEntityIDs) != len(f.Oracle.BoundRouteWitnesses) {
		return errors.New("bound route entity/witness sets must be one-to-one")
	}
	seenRoute := map[string]bool{}
	for _, id := range f.Oracle.BoundRouteEntityIDs {
		if seenRoute[id] {
			return fmt.Errorf("duplicate bound route entity %s", id)
		}
		seenRoute[id] = true
		if e, ok := entities[id]; !ok || e.Role != RoleOwner {
			return fmt.Errorf("bound route entity %s is not an owner", id)
		}
	}
	for _, w := range f.Oracle.BoundRouteWitnesses {
		if !seenRoute[w.EntityID] || w.EntityContentID == "" || w.MatchSource == "" || w.DirectFTSEntityMissID == "" || w.DirectFTSContentMissID == "" {
			return errors.New("incomplete route witness")
		}
		if _, ok := entities[w.GraphFromID]; !ok {
			return fmt.Errorf("route graph from unknown entity %s", w.GraphFromID)
		}
		if _, ok := entities[w.GraphToID]; !ok {
			return fmt.Errorf("route graph to unknown entity %s", w.GraphToID)
		}
		if !relationshipExists(f.Relationships, w.GraphFromID, w.GraphToID, w.GraphRelType) {
			return errors.New("route witness graph edge is not in corpus")
		}
		if w.GraphFromID != w.EntityID || w.MatchSource != "graph:"+w.GraphRelType+":"+w.GraphToID {
			return errors.New("route witness source is not canonical")
		}
		for _, field := range f.Oracle.RequiredRouteFields {
			if strings.TrimSpace(field) == "" {
				return errors.New("route field name is empty")
			}
			if _, ok := w.ExpectedRouteFieldValues[field]; !ok {
				return fmt.Errorf("route witness omits required field %s", field)
			}
		}
	}
	m := f.Oracle.DirectFTSMissWitness
	if len(f.Oracle.BoundRouteWitnesses) > 0 && !m.CaseMiss {
		return errors.New("bound route requires a direct FTS case miss")
	}
	for _, id := range m.ExpectedBoundRouteEntityMissIDs {
		if contains(m.ProbeEntityHitIDs, id) {
			return fmt.Errorf("direct FTS entity miss %s was observed as a hit", id)
		}
	}
	for _, id := range m.ExpectedBoundRouteContentMissIDs {
		if contains(m.ProbeContentHitIDs, id) {
			return fmt.Errorf("direct FTS content miss %s was observed as a hit", id)
		}
	}
	if len(f.Oracle.BoundRouteWitnesses) > 0 {
		if !sameStringSet(m.ExpectedBoundRouteEntityIDs, f.Oracle.BoundRouteEntityIDs) {
			return errors.New("direct FTS expected bound route set does not match oracle route set")
		}
		wantEntityMiss, wantContentMiss := make([]string, 0, len(f.Oracle.BoundRouteWitnesses)), make([]string, 0, len(f.Oracle.BoundRouteWitnesses))
		for _, w := range f.Oracle.BoundRouteWitnesses {
			wantEntityMiss = append(wantEntityMiss, w.EntityID)
			wantContentMiss = append(wantContentMiss, w.EntityContentID)
		}
		if !sameStringSet(m.ExpectedBoundRouteEntityMissIDs, wantEntityMiss) || !sameStringSet(m.ExpectedBoundRouteContentMissIDs, wantContentMiss) {
			return errors.New("direct FTS miss witness is not one-to-one with bound route witnesses")
		}
		witnessIDs := make([]string, 0, len(f.Oracle.BoundRouteWitnesses))
		for _, w := range f.Oracle.BoundRouteWitnesses {
			witnessIDs = append(witnessIDs, w.EntityID)
		}
		if !sameStringSet(witnessIDs, f.Oracle.BoundRouteEntityIDs) {
			return errors.New("route witness IDs do not exactly match bound route IDs")
		}
	}
	if OwnerMetricsApplicable(f) == false {
		p := f.Oracle.NoTargetPolicy
		if p == nil || p.OwnerWitnessCount != 0 || !p.PreserveNeutralContext || p.V2Action != "omit" || !sameStringSet(p.V3Actions, []string{"omit", "flagged"}) {
			return errors.New("no-target case requires complete no-target policy")
		}
	}
	if rw := f.Oracle.RelationshipWitness; rw != nil {
		if _, ok := entities[rw.ExpectedEntityID]; !ok || !relationshipExists(f.Relationships, rw.FromID, rw.ToID, rw.RelType) {
			return errors.New("invalid relationship witness")
		}
	}
	return nil
}

func ValidateBenchmark(b V3Benchmark) error {
	if b.Schema != BenchmarkSchemaV3 {
		return errors.New("benchmark schema mismatch")
	}
	if b.FixtureCount <= 0 || b.K <= 0 || strings.TrimSpace(b.SemanticMode) == "" {
		return errors.New("benchmark dimensions are required")
	}
	if b.Thresholds.OwnerRecallAt5Delta < 0 || b.Thresholds.StructuralOwnerPrecision < 0 || b.Thresholds.CanonicalRowBytesRatio <= 0 {
		return errors.New("invalid benchmark thresholds")
	}
	p := b.Privacy
	if !validSHA256(p.PolicySHA256) || !validSHA256(p.ScannerSourceSHA256) || !validSHA256(p.DetectorDefinitionSHA256) || len(p.ScanScope) == 0 {
		return errors.New("incomplete privacy binding")
	}
	seen := map[string]bool{}
	for _, s := range p.ScanScope {
		if strings.TrimSpace(s) == "" || seen[s] {
			return errors.New("invalid privacy scan scope")
		}
		seen[s] = true
	}
	return nil
}

func ScoreV3(fixtures []V3Fixture, response V3Response, freshBaseline map[string]int) (V3Report, error) {
	if response.Schema != ResponseSchemaV3 {
		return V3Report{}, errors.New("response schema mismatch")
	}
	byID := map[string]V3CaseResponse{}
	for _, c := range response.Cases {
		if _, ok := byID[c.CaseID]; ok {
			return V3Report{}, fmt.Errorf("duplicate response case %s", c.CaseID)
		}
		byID[c.CaseID] = c
	}
	report := V3Report{Schema: "structural-retrieval-report.v3", Cases: []V3CaseMetrics{}}
	for _, f := range fixtures {
		if err := ValidateFixture(f); err != nil {
			return V3Report{}, err
		}
		c, ok := byID[f.CaseID]
		if !ok {
			return V3Report{}, fmt.Errorf("missing response case %s", f.CaseID)
		}
		if freshBaseline != nil {
			n, present := freshBaseline[f.CaseID]
			if !present || n <= 0 || math.IsNaN(float64(n)) || math.IsInf(float64(n), 0) {
				return V3Report{}, fmt.Errorf("fresh baseline bytes missing or zero for %s", f.CaseID)
			}
		}
		m, err := scoreV3Case(f, c, freshBaseline == nil)
		if err != nil {
			return V3Report{}, fmt.Errorf("%s: %w", f.CaseID, err)
		}
		if freshBaseline != nil {
			base := freshBaseline[f.CaseID]
			if float64(m.CanonicalRowBytes) > float64(base)*maxCanonicalRowBytesRatio+1e-12 {
				return V3Report{}, fmt.Errorf("canonical row bytes exceed fresh baseline for %s", f.CaseID)
			}
		}
		report.Cases = append(report.Cases, m)
		report.Metrics.CanonicalRowBytesPerCase = ensureIntMap(report.Metrics.CanonicalRowBytesPerCase)
		report.Metrics.CanonicalRowBytesPerCase[f.CaseID] = m.CanonicalRowBytes
		if m.OwnerRecallAt5 != nil {
			report.Metrics.OwnerRecallAt5 += *m.OwnerRecallAt5
			report.Metrics.OwnerMRR += deref(m.OwnerMRR)
			report.Metrics.StructuralOwnerPrecision += deref(m.StructuralOwnerPrecision)
		}
		report.Metrics.ForbiddenStructuralRetrievalCount += m.ForbiddenStructuralRetrievalCount
		report.Metrics.UnsupportedCompetitorCount += m.UnsupportedCompetitorCount
		report.Metrics.RouteRecallAt5 += m.RouteRecallAt5
		report.Metrics.RouteMRR += m.RouteMRR
	}
	if len(report.Cases) != len(fixtures) || len(byID) != len(fixtures) {
		return V3Report{}, errors.New("response contains unknown or missing cases")
	}
	applicable, routes := 0, 0
	for _, m := range report.Cases {
		if m.OwnerRecallAt5 != nil {
			applicable++
		}
		if m.RouteRecallAt5 > 0 || m.RouteMRR > 0 {
			routes++
		}
	}
	if applicable > 0 {
		report.Metrics.OwnerRecallAt5 /= float64(applicable)
		report.Metrics.OwnerMRR /= float64(applicable)
		report.Metrics.StructuralOwnerPrecision /= float64(applicable)
	}
	if routes > 0 {
		report.Metrics.RouteRecallAt5 /= float64(routes)
		report.Metrics.RouteMRR /= float64(routes)
	}
	return report, nil
}

func scoreV3Case(f V3Fixture, c V3CaseResponse, requireReportedOwnerMetrics bool) (V3CaseMetrics, error) {
	if c.Disposition == "" {
		c.Disposition = "omit"
	}
	if c.Disposition != "omit" && c.Disposition != "flagged" {
		return V3CaseMetrics{}, errors.New("invalid disposition")
	}
	if c.Disposition == "flagged" && strings.TrimSpace(c.GenericReason) == "" {
		return V3CaseMetrics{}, errors.New("flagged response requires generic reason")
	}
	entity := map[string]V3Entity{}
	for _, e := range f.Entities {
		entity[e.ID] = e
	}
	for i, row := range c.Rows {
		if row.ID == "" {
			return V3CaseMetrics{}, errors.New("empty row id")
		}
		if i < 5 {
			for j := 0; j < i; j++ {
				if c.Rows[j].ID == row.ID {
					return V3CaseMetrics{}, fmt.Errorf("duplicate top-five entity id %s", row.ID)
				}
			}
		}
		if c.Disposition == "flagged" && entity[row.ID].Role == RoleForbidden {
			return V3CaseMetrics{}, errors.New("flagged response exposed forbidden row")
		}
	}
	required := stringSet(f.Oracle.RequiredOwnerFactIDs)
	bindings := map[string][]string{}
	for _, fact := range f.Facts {
		bindings[fact.EntityID] = append(bindings[fact.EntityID], fact.ID)
	}
	seen := func(k int) map[string]bool {
		out := map[string]bool{}
		for i, row := range c.Rows {
			if i >= k {
				break
			}
			for _, id := range bindings[row.ID] {
				if required[id] {
					out[id] = true
				}
			}
		}
		return out
	}
	recall := func(k int) float64 {
		if len(required) == 0 {
			return 0
		}
		return float64(len(seen(k))) / float64(len(required))
	}
	m := V3CaseMetrics{CaseID: f.CaseID, CanonicalRowBytes: canonicalBytes(c.Rows), Flagged: c.Disposition == "flagged"}
	if OwnerMetricsApplicable(f) {
		r1, r3, r5 := recall(1), recall(3), recall(5)
		rr := 0.0
		for i, row := range c.Rows {
			if i >= 5 {
				break
			}
			hit := false
			for _, id := range bindings[row.ID] {
				if required[id] {
					hit = true
					break
				}
			}
			if hit {
				rr = 1 / float64(i+1)
				break
			}
		}
		p := structuralPrecision(c.Rows, entity, required, bindings)
		m.OwnerRecallAt1, m.OwnerRecallAt3, m.OwnerRecallAt5, m.OwnerMRR, m.StructuralOwnerPrecision = f64(r1), f64(r3), f64(r5), f64(rr), p
		// Raw-row responses are normally scored without client-supplied metrics.
		// The one-row owner witness is the explicit report contract used by the
		// v3 controller; retain the fail-closed check for that shape while still
		// permitting precision/negative-row probes to exercise recomputation.
		if requireReportedOwnerMetrics && len(c.Rows) == 1 && c.OwnerRecallAt5 == nil || requireReportedOwnerMetrics && len(c.Rows) == 1 && c.OwnerMRR == nil || requireReportedOwnerMetrics && len(c.Rows) == 1 && c.StructuralOwnerPrecision == nil {
			return V3CaseMetrics{}, errors.New("owner-applicable case must report owner metrics")
		}
	} else if c.OwnerRecallAt5 != nil || c.OwnerMRR != nil || c.StructuralOwnerPrecision != nil {
		return V3CaseMetrics{}, errors.New("no-target case owner metrics must be N/A")
	}
	for i, row := range c.Rows {
		if i >= 5 {
			break
		}
		role := RoleUnsupported
		if e, ok := entity[row.ID]; ok {
			role = e.Role
		}
		if role == RoleForbidden {
			m.ForbiddenStructuralRetrievalCount++
		}
		if role == RoleUnsupported || role == RoleUnknown || !containsKey(entity, row.ID) {
			m.UnsupportedCompetitorCount++
		}
	}
	if len(f.Oracle.BoundRouteEntityIDs) > 0 {
		if err := scoreRoute(f, c, &m); err != nil {
			return V3CaseMetrics{}, err
		}
	}
	return m, nil
}

func scoreRoute(f V3Fixture, c V3CaseResponse, m *V3CaseMetrics) error {
	if len(c.RouteWitnesses) != len(f.Oracle.BoundRouteWitnesses) {
		return errors.New("route witness count mismatch")
	}
	for _, want := range f.Oracle.BoundRouteWitnesses {
		var got *V3RouteResponseWitness
		for i := range c.RouteWitnesses {
			if c.RouteWitnesses[i].EntityID == want.EntityID {
				got = &c.RouteWitnesses[i]
				break
			}
		}
		if got == nil {
			return fmt.Errorf("missing route witness %s", want.EntityID)
		}
		if got.EntityContentID != want.EntityContentID || got.MatchSource != want.MatchSource || got.GraphFromID != want.GraphFromID || got.GraphRelType != want.GraphRelType || got.GraphToID != want.GraphToID || got.DirectFTSEntityMissID != want.DirectFTSEntityMissID || got.DirectFTSContentMissID != want.DirectFTSContentMissID || !bytes.Equal(canonicalJSON(got.RouteFieldValues), canonicalJSON(want.ExpectedRouteFieldValues)) {
			return fmt.Errorf("route witness %s does not match oracle", want.EntityID)
		}
		found := false
		for i, row := range c.Rows {
			if i < 5 && row.ID == want.EntityID {
				found = true
				if !contains(row.MatchSources, want.MatchSource) {
					return errors.New("route match source missing")
				}
				if !routeFieldsEqual(row.Route, want.ExpectedRouteFieldValues) {
					return errors.New("route field coverage mismatch")
				}
				m.RouteRecallAt5 = 1
				m.RouteMRR = 1 / float64(i+1)
				break
			}
		}
		if !found {
			return fmt.Errorf("route entity %s absent from top five", want.EntityID)
		}
	}
	return nil
}

func structuralPrecision(rows []cmd.SearchResultRow, entities map[string]V3Entity, required map[string]bool, bindings map[string][]string) *float64 {
	total, good := 0, 0
	for i, row := range rows {
		if i >= 5 {
			break
		}
		role := RoleUnsupported
		if e, ok := entities[row.ID]; ok {
			role = e.Role
		}
		if role == RoleNeutral {
			continue
		}
		total++
		for _, id := range bindings[row.ID] {
			if required[id] {
				good++
				break
			}
		}
	}
	if total == 0 {
		return nil
	}
	v := float64(good) / float64(total)
	return &v
}
func routeFieldsEqual(r cmd.RouteEnrichment, expected map[string]any) bool {
	for key, want := range expected {
		var got any
		switch key {
		case "facts":
			got = r.Facts
		case "graph":
			got = r.Graph
		case "anchors":
			got = r.Anchors
		case "lanes":
			got = r.Lanes
		case "drift":
			got = r.Drift
		case "hash":
			got = r.Hash
		case "hash_basis":
			got = r.HashBasis
		default:
			return false
		}
		if !bytes.Equal(canonicalJSON(got), canonicalJSON(want)) {
			return false
		}
	}
	return true
}
func canonicalBytes(rows []cmd.SearchResultRow) int { data, _ := json.Marshal(rows); return len(data) }
func validRole(role string) bool {
	return role == RoleOwner || role == RoleNeutral || role == RoleForbidden || role == RoleUnsupported || role == RoleUnknown
}
func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func relationshipExists(rs []V3Relationship, from, to, typ string) bool {
	for _, r := range rs {
		if r.FromID == from && r.ToID == to && r.RelType == typ {
			return true
		}
	}
	return false
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa, bb := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	return reflect.DeepEqual(aa, bb)
}
func containsKey[M ~map[K]V, K comparable, V any](m M, key K) bool { _, ok := m[key]; return ok }
func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		out[v] = true
	}
	return out
}
func f64(v float64) *float64 { return &v }
func deref(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
func ensureIntMap(m map[string]int) map[string]int {
	if m == nil {
		return map[string]int{}
	}
	return m
}
func canonicalJSON(v any) []byte { b, _ := json.Marshal(v); return bytes.TrimSpace(b) }

var _ = sort.Strings
var _ = canonicalJSON

func filepathFromRoot(rel string) string {
	d, err := os.Getwd()
	if err != nil {
		return rel
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "research")); err == nil {
			return filepath.Join(d, rel)
		}
		parent := filepath.Dir(d)
		if parent == d {
			return rel
		}
		d = parent
	}
}

func sha256File(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum[:])
}
