package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// RouteEnrichment is a lightweight routing layer for existing C3 surfaces.
// It is not a new primitive: search, lookup, graph, check, and eval remain the
// authority surfaces. This only adds first-inspection clues to their output.
type RouteEnrichment struct {
	Facts     []string `json:"facts,omitempty"`
	Graph     []string `json:"graph,omitempty"`
	Anchors   []string `json:"anchors,omitempty"`
	Lanes     []string `json:"lanes,omitempty"`
	Drift     []string `json:"drift,omitempty"`
	Hash      string   `json:"hash,omitempty"`
	HashBasis string   `json:"hash_basis,omitempty"`
}

// RouteEnrichmentSnapshot captures the expected route fields before route
// enrichment is attached to a public row. It is internal provenance used to
// detect a later mismatch; the snapshot itself is never serialized.
type RouteEnrichmentSnapshot struct {
	EntityID string
	InputIDs []string
	Expected RouteEnrichment
}

func snapshotRouteBeforeEnrichment(s *store.Store, c3Dir, projectDir string, row SearchResultRow, query string) RouteEnrichmentSnapshot {
	// Build the expected route from a private copy so the public row remains
	// untouched. Using the same read-only enrichment inputs avoids inventing a
	// second route-selection policy (context precedence is part of the existing
	// controller behavior).
	prepared := row
	prepared.MatchSources = append([]string(nil), row.MatchSources...)
	prepared.Context = row.Context
	_ = enrichSearchRow(s, &prepared)
	ids := []string{prepared.ID}
	for _, ref := range []EntityRef{prepared.Context.Component, prepared.Context.Ref, prepared.Context.Rule} {
		if ref.ID != "" {
			ids = append(ids, ref.ID)
		}
	}
	return RouteEnrichmentSnapshot{
		EntityID: row.ID,
		InputIDs: append([]string(nil), ids...),
		Expected: buildRouteEnrichmentForIDs(s, c3Dir, projectDir, ids, query),
	}
}

// routeInputIDsBeforeEnrichment mirrors the read-only relationship inputs used
// by enrichSearchRow. It lets the snapshot be taken before the row is mutated
// while still binding the same owner/component/ref/rule route fields.
func routeInputIDsBeforeEnrichment(s *store.Store, entityID string) []string {
	ids := []string{entityID}
	seen := map[string]bool{entityID: true}
	appendID := func(id string) {
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if s == nil {
		return ids
	}
	rels, err := s.RelationshipsFrom(entityID)
	if err != nil {
		return ids
	}
	var components []string
	for _, rel := range rels {
		if rel == nil {
			continue
		}
		target, err := s.GetEntity(rel.ToID)
		if err != nil {
			continue
		}
		appendID(target.ID)
		if target.Type == "component" {
			components = append(components, target.ID)
		}
	}
	for _, componentID := range components {
		componentRels, err := s.RelationshipsFrom(componentID)
		if err != nil {
			continue
		}
		for _, rel := range componentRels {
			if rel != nil {
				appendID(rel.ToID)
			}
		}
	}
	return ids
}

func validateRouteSnapshot(snapshot RouteEnrichmentSnapshot, actual RouteEnrichment) error {
	if snapshot.EntityID == "" {
		return fmt.Errorf("route snapshot entity id is empty; hint: rerun search with a stored entity id")
	}
	expected, err := json.Marshal(snapshot.Expected)
	if err != nil {
		return fmt.Errorf("marshal expected route: %w", err)
	}
	got, err := json.Marshal(actual)
	if err != nil {
		return fmt.Errorf("marshal actual route: %w", err)
	}
	if string(expected) != string(got) {
		return fmt.Errorf("route fields differ from pre-enrichment snapshot (expected=%s got=%s); hint: discard the unbound route claim", expected, got)
	}
	return nil
}

func buildRouteEnrichment(s *store.Store, c3Dir, projectDir, entityID, query string) RouteEnrichment {
	return buildRouteEnrichmentForIDs(s, c3Dir, projectDir, []string{entityID}, query)
}

func buildRouteEnrichmentForIDs(s *store.Store, c3Dir, projectDir string, ids []string, query string) RouteEnrichment {
	if s == nil {
		return RouteEnrichment{}
	}
	facts := map[string]bool{}
	graph := map[string]bool{}
	var texts []string

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || facts[id] {
			continue
		}
		entity, err := s.GetEntity(id)
		if err != nil {
			continue
		}
		facts[id] = true
		texts = append(texts, entity.Title, entity.Goal, entity.Type)
		if entity.ParentID != "" {
			graph[entity.ParentID] = true
		}
		for _, rel := range routeRelationships(s, id) {
			otherID := routeOtherID(id, rel)
			if otherID == "" || otherID == id {
				continue
			}
			graph[otherID] = true
			if other, err := s.GetEntity(otherID); err == nil {
				texts = append(texts, other.Title, other.Goal, other.Type)
				if isRouteFactType(other.Type) {
					facts[otherID] = true
				}
			}
		}
	}

	route := RouteEnrichment{
		Facts: sortedKeys(facts),
		Graph: sortedKeys(graph),
		Lanes: inferRouteLanes(strings.Join(append([]string{query}, texts...), " ")),
	}
	route.Anchors, route.Drift = routeAnchors(c3Dir, projectDir, route.Facts)
	route.HashBasis = "entity ids, neighboring fact ids, eval code globs, inferred lanes, drift labels"
	route.Hash = routeHash(route)
	return route
}

func routeRelationships(s *store.Store, id string) []*store.Relationship {
	var out []*store.Relationship
	if rels, err := s.RelationshipsFrom(id); err == nil {
		out = append(out, rels...)
	}
	if rels, err := s.RelationshipsTo(id); err == nil {
		out = append(out, rels...)
	}
	return out
}

func routeOtherID(id string, rel *store.Relationship) string {
	if rel == nil {
		return ""
	}
	if rel.FromID == id {
		return rel.ToID
	}
	if rel.ToID == id {
		return rel.FromID
	}
	return ""
}

func isRouteFactType(entityType string) bool {
	switch entityType {
	case "system", "container", "component", "ref", "rule":
		return true
	default:
		return false
	}
}

func routeAnchors(c3Dir, projectDir string, facts []string) ([]string, []string) {
	if c3Dir == "" || len(facts) == 0 {
		return nil, nil
	}
	specs, err := LoadEvalSpecs(c3Dir)
	if err != nil {
		return nil, []string{"eval-specs-unreadable"}
	}
	bindings := EvalBindings(specs)
	anchorSet := map[string]bool{}
	var drift []string
	for _, fact := range facts {
		for _, glob := range bindings[fact] {
			glob = filepath.ToSlash(strings.TrimSpace(glob))
			if glob == "" {
				continue
			}
			anchorSet[glob] = true
			if projectDir == "" {
				continue
			}
			matches, err := codemap.GlobFiles(os.DirFS(projectDir), glob)
			if err != nil {
				drift = append(drift, "anchor_error:"+glob)
				continue
			}
			if len(matches) == 0 {
				drift = append(drift, "missing_anchor:"+glob)
			}
		}
	}
	anchors := sortedKeys(anchorSet)
	sort.Strings(drift)
	return anchors, compactUniqueStrings(drift)
}

func inferRouteLanes(text string) []string {
	lower := strings.ToLower(text)
	rules := []struct {
		lane  string
		terms []string
	}{
		{"frontend/backend", []string{"frontend", "backend"}},
		{"auth", []string{"auth", "login", "guard", "jwt"}},
		{"invoice", []string{"invoice", "billing"}},
		{"e2e", []string{"e2e", "test", "tests", "playwright", "lightpanda"}},
		{"theming", []string{"theme", "theming", "variant", "component", "components"}},
		{"lifecycle", []string{"lifecycle", "state", "flow"}},
		{"realtime-cycle", []string{"sync", "nats", "real-time", "realtime", "cycle"}},
		{"ownership", []string{"owner", "ownership"}},
		{"time", []string{"time", "timeline", "cadence"}},
	}
	var lanes []string
	for _, rule := range rules {
		for _, term := range rule.terms {
			if strings.Contains(lower, term) {
				lanes = append(lanes, rule.lane)
				break
			}
		}
	}
	return lanes
}

func routeHash(route RouteEnrichment) string {
	if len(route.Facts) == 0 && len(route.Graph) == 0 && len(route.Anchors) == 0 && len(route.Lanes) == 0 && len(route.Drift) == 0 {
		return ""
	}
	payload := struct {
		Facts   []string `json:"facts,omitempty"`
		Graph   []string `json:"graph,omitempty"`
		Anchors []string `json:"anchors,omitempty"`
		Lanes   []string `json:"lanes,omitempty"`
		Drift   []string `json:"drift,omitempty"`
	}{
		Facts: route.Facts, Graph: route.Graph, Anchors: route.Anchors, Lanes: route.Lanes, Drift: route.Drift,
	}
	b, _ := json.Marshal(payload)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:16]
}

func compactRoute(route RouteEnrichment) string {
	var parts []string
	if len(route.Facts) > 0 {
		parts = append(parts, "facts="+compactRouteList(route.Facts, 5))
	}
	if len(route.Graph) > 0 {
		parts = append(parts, "graph="+compactRouteList(route.Graph, 4))
	}
	if len(route.Anchors) > 0 {
		parts = append(parts, "anchors="+compactRouteList(route.Anchors, 4))
	}
	if len(route.Lanes) > 0 {
		parts = append(parts, "lanes="+compactRouteList(route.Lanes, 6))
	}
	if len(route.Drift) > 0 {
		parts = append(parts, "drift="+compactRouteList(route.Drift, 3))
	}
	if route.Hash != "" {
		parts = append(parts, "hash="+route.Hash)
	}
	return strings.Join(parts, " ")
}

func compactRouteList(values []string, limit int) string {
	if limit <= 0 || len(values) <= limit {
		return strings.Join(values, ",")
	}
	return strings.Join(values[:limit], ",") + ",+" + strconv.Itoa(len(values)-limit)
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func compactUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := values[:0]
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
