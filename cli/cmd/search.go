package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	mdparse "github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// SearchOptions configures the search command boundary.
type SearchOptions struct {
	Store      *store.Store
	Query      string
	Hybrid     bool
	JSON       bool
	Limit      int
	TypeFilter string
	Semantic   bool
	NoSemantic bool
	Pack       bool
	ProjectDir string
	C3Dir      string
	// StructuralProjection activates the reviewed structural-owner candidate
	// adapter. It is intentionally internal: no CLI flag or default path wires
	// this capability yet.
	StructuralProjection bool
	// CaptureProvenance asks the candidate envelope to expose typed route
	// witnesses. It is intentionally internal and has no CLI wiring.
	CaptureProvenance bool
}

// SearchOutput is the structured search response.
type SearchOutput struct {
	Query          string               `json:"query"`
	Results        []SearchResultRow    `json:"results"`
	RouteWitnesses []SearchRouteWitness `json:"route_witnesses,omitempty"`
}

// SearchRouteWitness is candidate-only route provenance. It contains typed
// identifiers and the route fields already attached to the corresponding
// result; it deliberately omits oracle labels and raw content.
type SearchRouteWitness struct {
	EntityID                string          `json:"entity_id"`
	EntityContentIDs        []string        `json:"entity_content_ids"`
	MatchSource             string          `json:"match_source"`
	GraphFromID             string          `json:"graph_from_id"`
	GraphRelType            string          `json:"graph_rel_type"`
	GraphToID               string          `json:"graph_to_id"`
	DirectFTSEntityHitIDs   []string        `json:"direct_fts_entity_hit_ids,omitempty"`
	DirectFTSContentHitIDs  []string        `json:"direct_fts_content_hit_ids,omitempty"`
	DirectFTSEntityMissID   string          `json:"direct_fts_entity_miss_id"`
	DirectFTSContentMissIDs []string        `json:"direct_fts_content_miss_ids"`
	RouteFieldValues        RouteEnrichment `json:"route_field_values"`
}

// SearchPackOutput is a compact set of distinct, source-backed search evidence.
type SearchPackOutput struct {
	Query    string               `json:"query"`
	Evidence []SearchPackEvidence `json:"evidence"`
}

// SearchPackEvidence exposes only classifications the stored record can prove.
// It does not claim whether the record matches runtime truth.
type SearchPackEvidence struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	RecordClass   string   `json:"class"`
	State         string   `json:"state,omitempty"`
	Citation      string   `json:"cite,omitempty"`
	Why           string   `json:"why,omitempty"`
	Lanes         []string `json:"lanes,omitempty"`
	SourceAnchors []string `json:"anchors,omitempty"`
	RecordClaims  []string `json:"record_claims,omitempty"`
	Evidence      string   `json:"evidence"`
}

type compactSearchPackEvidence struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Class    string `json:"class"`
	State    string `json:"state,omitempty"`
	Cite     string `json:"cite,omitempty"`
	Why      string `json:"why,omitempty"`
	Lanes    string `json:"lanes,omitempty"`
	Anchors  string `json:"anchors,omitempty"`
	Claims   string `json:"record_claims,omitempty"`
	Evidence string `json:"evidence"`
}

// SearchResultRow is one ranked result with search and graph provenance.
type SearchResultRow struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Title        string          `json:"title"`
	Snippet      string          `json:"snippet"`
	MatchSources []string        `json:"match_sources"`
	Context      SearchContext   `json:"context"`
	Route        RouteEnrichment `json:"route,omitempty"`
}

// SearchSafetyProof is internal controller provenance. It is not serialized
// into SearchResultRow and is valid only when a controller-owned positive
// certificate is present. Absence is unsafe/unproven, never neutral by
// inference.
type SearchSafetyProof struct {
	EntityID       string
	Classification string
	EvidenceKind   string
	EvidenceID     string
	SourceHash     string
	Conflict       bool
}

func (p SearchSafetyProof) Valid() bool {
	return p.EntityID != "" && p.Classification == "safe_context" &&
		p.EvidenceKind != "" && p.EvidenceID != "" && p.SourceHash != "" && !p.Conflict
}

// BoundRouteWitness binds a route claim to the exact relationship and direct
// probe that produced it. It remains internal provenance; route output remains
// the existing public RouteEnrichment shape.
type BoundRouteWitness struct {
	EntityID                 string
	EntityContentIDs         []string
	MatchSource              string
	GraphFromID              string
	GraphRelType             string
	GraphToID                string
	DirectFTSEntityHitIDs    []string
	DirectFTSContentHitIDs   []string
	DirectFTSEntityMissID    string
	DirectFTSContentMissIDs  []string
	ExpectedRouteFieldValues RouteEnrichment
}

func (w BoundRouteWitness) Bound() bool {
	return w.Validate() == nil
}

// Validate rejects graph/source mismatches and direct-hit laundering before a
// witness can be used for route credit.
func (w BoundRouteWitness) Validate() error {
	if w.EntityID == "" || w.GraphRelType == "" || w.GraphToID == "" || w.GraphFromID != w.EntityID {
		return fmt.Errorf("route witness endpoints are not bound; hint: discard the unbound graph row")
	}
	wantSource := "graph:" + w.GraphRelType + ":" + w.GraphToID
	if w.MatchSource != wantSource {
		return fmt.Errorf("route witness source is not bound to graph endpoints; hint: discard the unbound graph row")
	}
	if w.DirectFTSEntityMissID == "" || len(w.DirectFTSContentMissIDs) == 0 {
		return fmt.Errorf("route witness lacks direct entity/content miss proof; hint: discard the incomplete route claim")
	}
	if w.DirectFTSEntityMissID != "" && containsString(w.DirectFTSEntityHitIDs, w.DirectFTSEntityMissID) {
		return fmt.Errorf("route witness names a direct entity hit as a miss; hint: discard the route claim")
	}
	for _, miss := range w.DirectFTSContentMissIDs {
		if miss != "" && containsString(w.DirectFTSContentHitIDs, miss) {
			return fmt.Errorf("route witness names a direct content hit as a miss; hint: discard the route claim")
		}
	}
	return nil
}

// searchProvenance is carried beside rows through collection, graph
// expansion, fusion, and route construction. It is intentionally unexported
// so no oracle/fixture labels can leak into the public response.
type searchProvenance struct {
	Direct store.DirectFTSProbe
	// OriginalDirect identifies rows emitted by the unchanged direct FTS
	// collection. Graph-expanded and semantic-only rows deliberately do not
	// inherit this bit; the structural projection may only use it for the
	// one-hop containment witness lookup.
	OriginalDirect   bool
	Parent           store.ImmediateParentWitness
	HasParentWitness bool
	Safety           SearchSafetyProof
	Route            *BoundRouteWitness
}

// SearchContext captures the most useful graph/code-map context for a hit.
type SearchContext struct {
	Component EntityRef `json:"component"`
	Ref       EntityRef `json:"ref"`
	Rule      EntityRef `json:"rule"`
	Path      string    `json:"path"`
}

type compactSearchResultRow struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Why     string `json:"why"`
	Ctx     string `json:"ctx"`
	Route   string `json:"route,omitempty"`
	Snippet string `json:"s"`
}

const (
	defaultSearchLimit      = 20
	defaultAgentSearchLimit = 5
	compactSearchSnippetMax = 118
	compactBehaviorClaimMax = 240
)

// EntityRef is a lightweight entity reference in search context.
type EntityRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// RunSearch performs semantic, content/entity FTS, and graph-context search.
func RunSearch(opts SearchOptions, w io.Writer) error {
	if opts.Store == nil {
		return fmt.Errorf("error: search store is required\nhint: run c3x check to rebuild the local cache, then rerun c3x search")
	}
	if strings.TrimSpace(opts.Query) == "" {
		return fmt.Errorf("error: search requires a <query> argument\nhint: c3x search \"pool wait\"")
	}
	limit := resolveSearchLimit(opts.Limit)
	entityType := opts.TypeFilter

	semanticEnabled := !opts.NoSemantic
	if semanticEnabled {
		if _, err := opts.Store.EnsureSemanticIndexWithOptions(context.Background(), store.SemanticIndexOptions{AllowDownload: true}); err != nil {
			if errors.Is(err, store.ErrSemanticUnavailable) {
				semanticEnabled = false
			} else {
				return fmt.Errorf("ensure semantic index: %w", err)
			}
		}
	}

	if semanticEnabled {
		if count, err := opts.Store.SemanticIndexCount(); err != nil {
			return err
		} else if count == 0 {
			semanticEnabled = false
		}
	}

	// Over-fetch a candidate pool so reciprocal-rank fusion can lift a doc that is
	// corroborated across signals even when it sits beyond the caller's limit on any
	// single signal. The one hard truncation to limit happens after fusion.
	pool := candidatePoolLimit(limit)
	rows, provenance, err := collectSearchRowsWithProvenance(opts.Store, opts.Query, entityType, pool)
	if err != nil {
		return err
	}
	rows, provenance, err = expandHybridRowsWithProvenance(opts.Store, rows, provenance, opts.Query, entityType, pool)
	if err != nil {
		return err
	}
	var semanticRows []store.SearchResult
	if semanticEnabled {
		semanticRows, err = opts.Store.SearchSemanticWithOptions(context.Background(), opts.Query, entityType, pool, store.SemanticSearchOptions{
			AllowDownload: true,
		})
		if err != nil {
			if errors.Is(err, store.ErrSemanticUnavailable) {
				semanticRows = nil
			} else {
				return fmt.Errorf("search semantic: %w", err)
			}
		}
	}
	rows, provenance = fuseSemanticRowsWithProvenance(rows, provenance, semanticRows, limit)
	if opts.StructuralProjection {
		rows, provenance, err = projectStructuralOwnerWitnessRows(opts.Store, opts.Query, entityType, rows, provenance)
		if err != nil {
			return err
		}
	}
	for i := range rows {
		snapshot := snapshotRouteBeforeEnrichment(opts.Store, opts.C3Dir, opts.ProjectDir, rows[i], opts.Query)
		if err := enrichSearchRow(opts.Store, &rows[i]); err != nil {
			return err
		}
		rows[i].Route = buildSearchRoute(opts.Store, opts.C3Dir, opts.ProjectDir, rows[i], opts.Query)
		if err := validateRouteSnapshot(snapshot, rows[i].Route); err != nil {
			return fmt.Errorf("search route provenance %s: %w", rows[i].ID, err)
		}
		if p := provenance[rows[i].ID]; p != nil {
			if p.Route != nil {
				p.Route.ExpectedRouteFieldValues = snapshot.Expected
			}
		}
	}

	if opts.Pack {
		return writeSearchPack(opts, rows, w)
	}

	format := ResolveFormat(opts.JSON, isAgentMode())
	out := SearchOutput{
		Query:   opts.Query,
		Results: rows,
	}
	if opts.CaptureProvenance {
		out.RouteWitnesses = searchRouteWitnesses(rows, provenance)
	}
	if format == FormatJSON {
		return WriteObjectOutput(w, out, format, searchHelpHints())
	}
	return WriteTableOutput(w, "results", compactSearchRows(rows), []string{"id", "title", "why", "ctx", "route", "s"}, format, nil)
}

func writeSearchPack(opts SearchOptions, rows []SearchResultRow, w io.Writer) error {
	evidence := make([]SearchPackEvidence, 0, len(rows))
	for _, row := range selectDiverseSearchRows(rows, opts.Limit) {
		entity, err := opts.Store.GetEntity(row.ID)
		if err != nil {
			return fmt.Errorf("search pack entity %s: %w", row.ID, err)
		}
		citation, _ := entityCitation(entity)
		if citation == "" {
			// An entity id is still a replayable record locator when legacy cache
			// data has no versioned hash. Do not imply hash-backed freshness.
			citation = entity.ID
		}
		recordClaims, err := currentBehaviorClaims(opts.Store, row.ID, opts.Query, 2)
		if err != nil {
			return fmt.Errorf("search pack current behavior %s: %w", row.ID, err)
		}
		evidence = append(evidence, SearchPackEvidence{
			ID:            row.ID,
			Type:          row.Type,
			RecordClass:   searchRecordClass(row.Type),
			State:         searchRecordState(row.Type, entity.Status),
			Citation:      citation,
			Why:           compactMatchSources(row.MatchSources),
			Lanes:         row.Route.Lanes,
			SourceAnchors: row.Route.Anchors,
			RecordClaims:  recordClaims,
			Evidence:      compactSearchSnippet(row.Snippet),
		})
	}
	format := ResolveFormat(opts.JSON, isAgentMode())
	out := SearchPackOutput{Query: opts.Query, Evidence: evidence}
	if format == FormatJSON {
		return WriteObjectOutput(w, out, format, nil)
	}
	compact := make([]compactSearchPackEvidence, 0, len(evidence))
	for _, row := range evidence {
		compact = append(compact, compactSearchPackEvidence{
			ID: row.ID, Type: row.Type, Class: row.RecordClass, State: row.State,
			Cite: row.Citation, Why: row.Why,
			Lanes: compactRouteList(row.Lanes, 4), Anchors: compactRouteList(row.SourceAnchors, 3),
			Claims:   strings.Join(row.RecordClaims, " || "),
			Evidence: row.Evidence,
		})
	}
	return WriteTableOutput(w, "evidence", compact,
		[]string{"id", "type", "class", "state", "cite", "why", "lanes", "anchors", "record_claims", "evidence"}, format, nil)
}

func currentBehaviorClaims(s *store.Store, entityID, query string, limit int) ([]string, error) {
	nodes, err := s.NodesForEntity(entityID)
	if err != nil {
		return nil, err
	}
	body := content.RenderMarkdown(nodes)
	var table *mdparse.Table
	for _, section := range mdparse.ParseSections(body) {
		if strings.EqualFold(strings.TrimSpace(section.Name), "Current Behavior") {
			table, err = mdparse.ParseTable(section.Content)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if table == nil || limit <= 0 {
		return nil, nil
	}
	type candidate struct {
		text  string
		score int
		order int
	}
	queryTerms := claimTermSet(query)
	candidates := make([]candidate, 0, len(table.Rows))
	for i, row := range table.Rows {
		claim := firstTableValue(row, "Current claim", "Claim")
		if claim == "" {
			continue
		}
		mechanism := firstTableValue(row, "Mechanism")
		class := firstTableValue(row, "Class", "Behavior class")
		source := firstTableValue(row, "Source", "Source anchor")
		failure := firstTableValue(row, "Failure or absence", "Failure/absence")
		text := compactBehaviorClaim(mechanism, class, claim, source, failure)
		score := termOverlapScore(queryTerms, claimTermSet(strings.Join([]string{mechanism, class, claim, failure}, " ")))
		candidates = append(candidates, candidate{text: text, score: score, order: i})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].order < candidates[j].order
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, candidate.text)
	}
	return out, nil
}

func firstTableValue(row map[string]string, names ...string) string {
	for _, name := range names {
		for key, value := range row {
			if strings.EqualFold(strings.TrimSpace(key), name) {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func compactBehaviorClaim(mechanism, class, claim, source, failure string) string {
	parts := make([]string, 0, 4)
	if mechanism != "" {
		parts = append(parts, mechanism)
	}
	if class != "" {
		parts = append(parts, "["+class+"]")
	}
	parts = append(parts, claim)
	if failure != "" && !strings.EqualFold(failure, "n/a") && !strings.EqualFold(failure, "none") {
		parts = append(parts, "boundary: "+failure)
	}
	if source != "" {
		parts = append(parts, "@ "+source)
	}
	return compactText(strings.Join(parts, " "), compactBehaviorClaimMax)
}

func compactText(value string, maxRunes int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if maxRunes <= 0 || len(runes) <= maxRunes {
		return value
	}
	cut := maxRunes
	for i := maxRunes; i >= maxRunes/2; i-- {
		if unicode.IsSpace(runes[i-1]) {
			cut = i - 1
			break
		}
	}
	return strings.TrimSpace(string(runes[:cut])) + "..."
}

func claimTermSet(value string) map[string]bool {
	terms := map[string]bool{}
	for _, field := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len(field) >= 3 {
			terms[field] = true
		}
	}
	return terms
}

func termOverlapScore(left, right map[string]bool) int {
	score := 0
	for term := range left {
		if right[term] {
			score++
		}
	}
	return score
}

func selectDiverseSearchRows(rows []SearchResultRow, limit int) []SearchResultRow {
	if limit <= 0 {
		limit = defaultAgentSearchLimit
	}
	if len(rows) <= 1 {
		return rows
	}
	selected := make([]SearchResultRow, 0, min(limit, len(rows)))
	remaining := append([]SearchResultRow(nil), rows...)
	selected = append(selected, remaining[0])
	remaining = remaining[1:]
	seen := searchDiversityKeys(selected[0])
	for len(selected) < limit && len(remaining) > 0 {
		best := 0
		bestGain := -1
		for i, row := range remaining {
			gain := 0
			for key := range searchDiversityKeys(row) {
				if !seen[key] {
					gain++
				}
			}
			if gain > bestGain {
				best, bestGain = i, gain
			}
		}
		row := remaining[best]
		selected = append(selected, row)
		for key := range searchDiversityKeys(row) {
			seen[key] = true
		}
		remaining = append(remaining[:best], remaining[best+1:]...)
	}
	return selected
}

func searchDiversityKeys(row SearchResultRow) map[string]bool {
	keys := map[string]bool{"type:" + row.Type: true}
	for _, value := range row.Route.Lanes {
		keys["lane:"+value] = true
	}
	for _, value := range row.Route.Facts {
		keys["fact:"+value] = true
	}
	for _, value := range row.Route.Anchors {
		keys["anchor:"+value] = true
	}
	return keys
}

func searchRecordClass(entityType string) string {
	if isRouteFactType(entityType) || entityType == "pm-requirement" || entityType == "user-story" {
		return "fact"
	}
	switch entityType {
	case "adr", "prd", "atomic-design-change":
		return "change-doc"
	default:
		return "document"
	}
}

func searchRecordState(entityType, status string) string {
	if searchRecordClass(entityType) == "fact" {
		return ""
	}
	return status
}

func searchHelpHints() []HelpHint {
	return []HelpHint{
		{Command: "c3x read <id>", Description: "inspect a matching fact before relying on the search snippet"},
		{Command: "c3x graph <id>", Description: "expand a match into parent, child, ref, rule, and affected topology"},
		{Command: "c3x lookup <file-or-glob>", Description: "map a known file path back to owning facts and governing refs"},
	}
}

// candidatePoolLimit sizes the over-fetched candidate pool that feeds fusion. Each
// retrieval signal returns up to this many rows so fusion sees docs corroborated
// across signals before the single post-fusion truncation to the caller's limit.
// Sized at ~2x the cutoff: deep enough to recover a corroborated doc that sits just
// past the cutoff on a single signal, but not so deep that low-signal docs dilute the
// pool and evict a true hit (search-eval peaks at this ratio; 4x measurably regressed).
func candidatePoolLimit(limit int) int {
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	n := limit * 2
	if n < 10 {
		return 10
	}
	if n > 100 {
		return 100
	}
	return n
}

func collectSearchRows(s *store.Store, query, entityType string, limit int) ([]SearchResultRow, error) {
	byID := make(map[string]*SearchResultRow)
	order := make(map[string]int)
	nextOrder := 0

	addResults := func(results []store.SearchResult, source string) {
		for _, hit := range results {
			if entityType != "" && hit.Type != entityType {
				continue
			}
			snippet := cleanSnippet(hit.Snippet)
			if source == "content_fts" {
				snippet = bestContentSnippet(s, hit.ID, query, snippet)
			}
			row, ok := byID[hit.ID]
			if !ok {
				nextOrder++
				order[hit.ID] = nextOrder
				row = &SearchResultRow{
					ID:      hit.ID,
					Type:    hit.Type,
					Title:   hit.Title,
					Snippet: snippet,
				}
				byID[hit.ID] = row
			}
			addMatchSource(row, source)
			if row.Snippet == "" || source == "content_fts" {
				row.Snippet = snippet
			}
		}
	}

	contentHits, err := s.SearchContent(query, limit)
	if err != nil {
		return nil, err
	}
	if len(contentHits) == 0 {
		contentHits, err = s.SearchContent(ftsDisjunction(query), limit)
		if err != nil {
			return nil, err
		}
	}
	addResults(contentHits, "content_fts")

	entityHits, err := s.SearchWithLimit(query, entityType, limit)
	if err != nil {
		return nil, err
	}
	if len(entityHits) == 0 {
		entityHits, err = s.SearchWithLimit(ftsDisjunction(query), entityType, limit)
		if err != nil {
			return nil, err
		}
	}
	addResults(entityHits, "entity_fts")

	rows := make([]SearchResultRow, 0, len(byID))
	for _, row := range byID {
		rows = append(rows, *row)
	}
	sortSearchRows(rows, order)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

func collectSearchRowsWithProvenance(s *store.Store, query, entityType string, limit int) ([]SearchResultRow, map[string]*searchProvenance, error) {
	probe, err := s.ProbeDirectFTS(query, entityType, limit)
	if err != nil {
		return nil, nil, err
	}
	rows, err := collectSearchRows(s, query, entityType, limit)
	if err != nil {
		return nil, nil, err
	}
	provenance := make(map[string]*searchProvenance, len(rows))
	for _, row := range rows {
		copyProbe := probe
		copyProbe.EntityHitIDs = append([]string(nil), probe.EntityHitIDs...)
		copyProbe.ContentHitIDs = append([]string(nil), probe.ContentHitIDs...)
		copyProbe.ContentEntityHitIDs = append([]string(nil), probe.ContentEntityHitIDs...)
		provenance[row.ID] = &searchProvenance{Direct: copyProbe, OriginalDirect: true}
	}
	return rows, provenance, nil
}

func expandHybridRows(s *store.Store, rows []SearchResultRow, query, entityType string, limit int) ([]SearchResultRow, error) {
	byID := make(map[string]SearchResultRow, len(rows))
	order := make(map[string]int, len(rows))
	for i, row := range rows {
		byID[row.ID] = row
		order[row.ID] = i + 1
	}

	for _, base := range rows {
		inbound, err := s.RelationshipsTo(base.ID)
		if err != nil {
			return nil, fmt.Errorf("search inbound graph %s: %w", base.ID, err)
		}
		for _, rel := range inbound {
			from, err := s.GetEntity(rel.FromID)
			if err != nil {
				return nil, fmt.Errorf("search inbound graph source %s: %w", rel.FromID, err)
			}
			if entityType != "" && from.Type != entityType {
				continue
			}
			row, ok := byID[from.ID]
			if !ok {
				row = SearchResultRow{
					ID:      from.ID,
					Type:    from.Type,
					Title:   from.Title,
					Snippet: bestContentSnippet(s, from.ID, query, from.Title),
				}
				order[from.ID] = len(order) + 1
			}
			addMatchSource(&row, "graph:"+rel.RelType+":"+base.ID)
			byID[from.ID] = row
		}
	}

	expanded := make([]SearchResultRow, 0, len(byID))
	for _, row := range byID {
		expanded = append(expanded, row)
	}
	sortSearchRows(expanded, order)
	if len(expanded) > limit {
		expanded = expanded[:limit]
	}
	return expanded, nil
}

func expandHybridRowsWithProvenance(s *store.Store, rows []SearchResultRow, provenance map[string]*searchProvenance, query, entityType string, limit int) ([]SearchResultRow, map[string]*searchProvenance, error) {
	expanded, err := expandHybridRows(s, rows, query, entityType, limit)
	if err != nil {
		return nil, nil, err
	}
	out := make(map[string]*searchProvenance, len(expanded))
	for _, row := range expanded {
		if p := provenance[row.ID]; p != nil {
			out[row.ID] = p
			continue
		}
		p := &searchProvenance{}
		// A graph expansion source is the only provenance that can establish a
		// bound relationship here. Enrichment sources are intentionally ignored.
		for _, source := range row.MatchSources {
			relType, anchor, ok := parseGraphSource(source)
			if !ok {
				continue
			}
			base := provenance[anchor]
			if base == nil {
				continue
			}
			p.Direct = base.Direct
			p.Direct.EntityHitIDs = append([]string(nil), base.Direct.EntityHitIDs...)
			p.Direct.ContentHitIDs = append([]string(nil), base.Direct.ContentHitIDs...)
			p.Direct.ContentEntityHitIDs = append([]string(nil), base.Direct.ContentEntityHitIDs...)
			p.Direct.EntityMissIDs = append([]string(nil), base.Direct.EntityMissIDs...)
			p.Direct.ContentMissIDs = append([]string(nil), base.Direct.ContentMissIDs...)
			if !p.Direct.HasEntityHit(row.ID) {
				p.Direct.RecordEntityMiss(row.ID)
			}
			contentIDs, contentErr := s.ContentIDsForEntity(row.ID)
			if contentErr != nil {
				return nil, nil, contentErr
			}
			for _, contentID := range contentIDs {
				if !p.Direct.HasContentHit(contentID) {
					p.Direct.RecordContentMiss(contentID)
				}
			}
			p.Route = &BoundRouteWitness{
				EntityID: row.ID, MatchSource: source, GraphFromID: row.ID,
				GraphRelType: relType, GraphToID: anchor,
				DirectFTSEntityHitIDs:  append([]string(nil), p.Direct.EntityHitIDs...),
				DirectFTSContentHitIDs: append([]string(nil), p.Direct.ContentHitIDs...),
				EntityContentIDs:       append([]string(nil), contentIDs...),
			}
			if len(p.Direct.EntityMissIDs) > 0 {
				p.Route.DirectFTSEntityMissID = p.Direct.EntityMissIDs[0]
			}
			p.Route.DirectFTSContentMissIDs = append([]string(nil), p.Direct.ContentMissIDs...)
			if containsString(p.Direct.ContentEntityHitIDs, row.ID) || p.Route.Validate() != nil {
				// Missing proof retires the internal witness. Existing public rows
				// remain unchanged; no route credit can be inferred from a partial
				// graph expansion.
				p.Route = nil
			}
			break
		}
		out[row.ID] = p
	}
	return expanded, out, nil
}

func fuseSemanticRows(rows []SearchResultRow, semanticHits []store.SearchResult, limit int) []SearchResultRow {
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if len(semanticHits) == 0 {
		if len(rows) > limit {
			return rows[:limit]
		}
		return rows
	}

	const rrfK = 60.0
	type scoredRow struct {
		row       SearchResultRow
		score     float64
		firstRank int
	}
	byID := make(map[string]*scoredRow, len(rows)+len(semanticHits))
	addScore := func(id string, rank int) {
		item := byID[id]
		if item == nil {
			return
		}
		item.score += 1 / (rrfK + float64(rank))
		if item.firstRank == 0 || rank < item.firstRank {
			item.firstRank = rank
		}
	}

	for i, row := range rows {
		rank := i + 1
		copyRow := row
		byID[row.ID] = &scoredRow{row: copyRow, firstRank: rank}
		addScore(row.ID, rank)
	}
	for i, hit := range semanticHits {
		rank := i + 1
		item, ok := byID[hit.ID]
		if !ok {
			row := SearchResultRow{
				ID:      hit.ID,
				Type:    hit.Type,
				Title:   hit.Title,
				Snippet: cleanSnippet(hit.Snippet),
			}
			item = &scoredRow{row: row, firstRank: len(rows) + rank}
			byID[hit.ID] = item
		}
		addMatchSource(&item.row, "semantic")
		if item.row.Snippet == "" {
			item.row.Snippet = cleanSnippet(hit.Snippet)
		}
		addScore(hit.ID, rank)
	}

	scored := make([]scoredRow, 0, len(byID))
	for _, item := range byID {
		scored = append(scored, *item)
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].firstRank != scored[j].firstRank {
			return scored[i].firstRank < scored[j].firstRank
		}
		return scored[i].row.ID < scored[j].row.ID
	})

	fused := make([]SearchResultRow, 0, min(limit, len(scored)))
	for i := 0; i < len(scored) && i < limit; i++ {
		fused = append(fused, scored[i].row)
	}
	return fused
}

func fuseSemanticRowsWithProvenance(rows []SearchResultRow, provenance map[string]*searchProvenance, semanticHits []store.SearchResult, limit int) ([]SearchResultRow, map[string]*searchProvenance) {
	fused := fuseSemanticRows(rows, semanticHits, limit)
	out := make(map[string]*searchProvenance, len(fused))
	for _, row := range fused {
		if p := provenance[row.ID]; p != nil {
			out[row.ID] = p
		} else {
			out[row.ID] = &searchProvenance{}
		}
	}
	return fused, out
}

// projectStructuralOwnerWitnessRows is the unactivated structural retrieval
// candidate. It is deliberately kept outside RunSearch: the default search
// contract, ranking, cutoff, enrichment, and route controls remain unchanged
// until a separately reviewed candidate adapter is promoted.
//
// The projection is one pass over the already-fused rows. Direct rows may be
// replaced by their exact immediate containment parent; relationship-expanded
// rows may survive only with an exact bound route witness. No fixture/oracle
// labels are consulted and no generated parent is fed back through the pass.
func projectStructuralOwnerWitnessRows(s *store.Store, query, requestedType string, rows []SearchResultRow, provenance map[string]*searchProvenance) ([]SearchResultRow, map[string]*searchProvenance, error) {
	if s == nil {
		return nil, nil, fmt.Errorf("structural owner projection requires a store; hint: provide the unchanged search store")
	}
	if provenance == nil && len(rows) > 0 {
		return nil, nil, fmt.Errorf("structural owner projection provenance is missing; hint: rerun the unchanged controller")
	}

	type parentTarget struct {
		parentID string
		witness  store.ImmediateParentWitness
	}
	parentTargets := make(map[string]parentTarget)
	resolvedProvenance := make(map[string]*searchProvenance, len(rows))
	protectedIDs := make(map[string]bool)
	seenInput := make(map[string]bool, len(rows))
	ownerTargetCount := 0

	// Resolve all one-hop parent witnesses before constructing output. This
	// makes the no-cascade rule explicit: a row that is itself another row's
	// parent is kept as that parent's exact row, never projected to a grandparent.
	for _, row := range rows {
		if row.ID == "" {
			return nil, nil, fmt.Errorf("structural owner projection row has empty id; hint: discard malformed search output")
		}
		if seenInput[row.ID] {
			return nil, nil, fmt.Errorf("structural owner projection received duplicate row %q; hint: deduplicate before projection", row.ID)
		}
		seenInput[row.ID] = true
		originalP := provenance[row.ID]
		if originalP == nil {
			return nil, nil, fmt.Errorf("structural owner projection provenance missing for %q; hint: rerun the unchanged controller", row.ID)
		}
		p := cloneSearchProvenance(originalP)
		resolvedProvenance[row.ID] = p
		if !p.OriginalDirect {
			continue
		}
		// An accepted relationship witness has higher target priority than a
		// containment lookup. Do not let an unrelated parent read obscure it.
		if p.Route != nil {
			continue
		}
		w := p.Parent
		if !p.HasParentWitness {
			var err error
			w, err = s.LookupImmediateParentWithProbe(row.ID, requestedType, p.Direct)
			if err != nil {
				return nil, nil, fmt.Errorf("structural owner projection parent %s: %w", row.ID, err)
			}
		}
		if err := validateImmediateParentWitness(row.ID, requestedType, w); err != nil {
			return nil, nil, err
		}
		p.Parent = w
		p.HasParentWitness = true
		if w.ParentRead != "found" || !w.Immediate || !w.CycleChecked || w.ParentID == "" {
			continue
		}
		if requestedType != "" && w.ParentType != requestedType {
			// A filtered lookup is exact. A mismatch is simply no target; it
			// must never widen to the parent's stored type.
			continue
		}
		parent, err := s.GetEntity(w.ParentID)
		if err != nil {
			return nil, nil, fmt.Errorf("structural owner projection parent %s: %w", w.ParentID, err)
		}
		if parent.ID != w.ParentID || parent.Type != w.ParentType {
			return nil, nil, fmt.Errorf("structural owner projection parent witness type mismatch for %q; hint: discard malformed provenance", row.ID)
		}
		parentTargets[row.ID] = parentTarget{parentID: w.ParentID, witness: w}
		protectedIDs[w.ParentID] = true
		ownerTargetCount++
	}

	type eligibleRow struct {
		row SearchResultRow
		p   *searchProvenance
	}
	eligible := make([]eligibleRow, 0, len(rows))
	existing := make(map[string]eligibleRow, len(rows))
	for _, row := range rows {
		p := resolvedProvenance[row.ID]
		if p == nil {
			p = provenance[row.ID]
		}
		if p == nil {
			return nil, nil, fmt.Errorf("structural owner projection provenance missing for %q; hint: rerun the unchanged controller", row.ID)
		}
		if p.Route != nil {
			if err := validateProjectionRouteWitness(row, *p.Route); err != nil {
				if p.OriginalDirect {
					return nil, nil, err
				}
				// A relationship-expanded row without a complete bound witness
				// is an unsupported decoy, not route credit.
				continue
			}
			ownerTargetCount++
		} else if !p.OriginalDirect && hasGraphMatchSource(row.MatchSources) {
			// Graph expansion is not a containment witness. Do not retain an
			// unbound graph owner merely because it has a graph-shaped source.
			continue
		}
		copyP := cloneSearchProvenance(p)
		e := eligibleRow{row: row, p: copyP}
		eligible = append(eligible, e)
		existing[row.ID] = e
	}

	if ownerTargetCount == 0 {
		for _, e := range eligible {
			if e.p.Safety.EntityID != e.row.ID || !e.p.Safety.Valid() {
				return nil, nil, fmt.Errorf("structural owner projection has no structurally witnessed owner target and unsafe retained row %q; hint: require a complete safe_context certificate", e.row.ID)
			}
		}
	}

	type projected struct {
		row          SearchResultRow
		p            *searchProvenance
		rank         int
		containChild string
	}
	byTarget := make(map[string]*projected, len(eligible))
	for rank, e := range eligible {
		targetID := e.row.ID
		containChild := ""
		if target, ok := parentTargets[e.row.ID]; ok && !protectedIDs[e.row.ID] {
			targetID = target.parentID
			containChild = e.row.ID
		}

		candidate := &projected{rank: rank, containChild: containChild}
		if existingTarget, ok := existing[targetID]; ok {
			candidate.row = existingTarget.row
			candidate.p = cloneSearchProvenance(existingTarget.p)
		} else {
			entity, err := s.GetEntity(targetID)
			if err != nil {
				return nil, nil, fmt.Errorf("structural owner projection parent %s: %w", targetID, err)
			}
			candidate.row = SearchResultRow{
				ID:           entity.ID,
				Type:         entity.Type,
				Title:        entity.Title,
				Snippet:      bestContentSnippet(s, entity.ID, query, entity.Title),
				MatchSources: []string{"containment_parent:" + e.row.ID},
			}
			candidate.p = &searchProvenance{}
		}
		if containChild != "" {
			addMatchSource(&candidate.row, "containment_parent:"+containChild)
		}
		if previous, ok := byTarget[targetID]; ok {
			if candidate.rank < previous.rank || (candidate.rank == previous.rank && targetID < previous.row.ID) {
				candidate.rank = previous.rank
			}
			// Existing content/provenance wins. Multiple children may point to
			// one target; only their deduplicated containment source is merged.
			for _, source := range candidate.row.MatchSources {
				addMatchSource(&previous.row, source)
			}
			continue
		}
		byTarget[targetID] = candidate
	}

	projectedRows := make([]*projected, 0, len(byTarget))
	for _, candidate := range byTarget {
		projectedRows = append(projectedRows, candidate)
	}
	sort.SliceStable(projectedRows, func(i, j int) bool {
		if projectedRows[i].rank != projectedRows[j].rank {
			return projectedRows[i].rank < projectedRows[j].rank
		}
		return projectedRows[i].row.ID < projectedRows[j].row.ID
	})

	outRows := make([]SearchResultRow, 0, len(projectedRows))
	outProvenance := make(map[string]*searchProvenance, len(projectedRows))
	for _, candidate := range projectedRows {
		outRows = append(outRows, candidate.row)
		outProvenance[candidate.row.ID] = candidate.p
	}
	return outRows, outProvenance, nil
}

func validateImmediateParentWitness(childID, requestedType string, w store.ImmediateParentWitness) error {
	if w.ChildID != childID {
		return fmt.Errorf("structural owner projection parent witness child mismatch for %q; hint: discard malformed provenance", childID)
	}
	wantMode := store.ParentFilterDefault
	if requestedType != "" {
		wantMode = store.ParentFilterFiltered
	}
	if w.FilterMode != wantMode || w.RequestedTypeFilter != requestedType {
		return fmt.Errorf("structural owner projection parent witness filter mismatch for %q; hint: discard malformed provenance", childID)
	}
	if w.ParentRead != "found" && w.ParentRead != "missing" && w.ParentRead != "error" {
		return fmt.Errorf("structural owner projection parent witness read state is invalid for %q; hint: discard malformed provenance", childID)
	}
	if w.ParentRead == "error" {
		return fmt.Errorf("structural owner projection parent lookup failed for %q; hint: repair the unchanged store and retry", childID)
	}
	if w.ParentRead == "found" && (w.ParentID == "" || w.ParentType == "") {
		return fmt.Errorf("structural owner projection parent witness is incomplete for %q; hint: discard malformed provenance", childID)
	}
	if w.ParentRead == "missing" && (w.Immediate || !w.CycleChecked) {
		return fmt.Errorf("structural owner projection missing-parent witness is inconsistent for %q; hint: discard malformed provenance", childID)
	}
	return nil
}

func validateProjectionRouteWitness(row SearchResultRow, w BoundRouteWitness) error {
	if err := w.Validate(); err != nil {
		return fmt.Errorf("structural owner projection route witness %s: %w", row.ID, err)
	}
	if w.EntityID != row.ID || w.GraphFromID != row.ID || !hasSource(row.MatchSources, w.MatchSource) {
		return fmt.Errorf("structural owner projection route witness %s is not bound to the row; hint: discard the unbound graph row", row.ID)
	}
	if w.DirectFTSEntityMissID != row.ID || containsString(w.DirectFTSEntityHitIDs, row.ID) {
		return fmt.Errorf("structural owner projection route witness %s lacks an exact entity miss; hint: discard the route claim", row.ID)
	}
	if len(w.EntityContentIDs) == 0 {
		return fmt.Errorf("structural owner projection route witness %s lacks entity content binding; hint: discard the route claim", row.ID)
	}
	for _, miss := range w.DirectFTSContentMissIDs {
		if !containsString(w.EntityContentIDs, miss) || containsString(w.DirectFTSContentHitIDs, miss) {
			return fmt.Errorf("structural owner projection route witness %s lacks an exact content miss; hint: discard the route claim", row.ID)
		}
	}
	if routeNonEmpty(w.ExpectedRouteFieldValues) {
		if !sameRouteEnrichment(w.ExpectedRouteFieldValues, row.Route) {
			return fmt.Errorf("structural owner projection route witness %s fields differ from the bound route; hint: discard the unbound graph row", row.ID)
		}
	}
	return nil
}

func hasGraphMatchSource(sources []string) bool {
	for _, source := range sources {
		if strings.HasPrefix(source, "graph:") {
			return true
		}
	}
	return false
}

func cloneSearchProvenance(p *searchProvenance) *searchProvenance {
	if p == nil {
		return nil
	}
	out := *p
	out.Direct.EntityHitIDs = append([]string(nil), p.Direct.EntityHitIDs...)
	out.Direct.ContentHitIDs = append([]string(nil), p.Direct.ContentHitIDs...)
	out.Direct.ContentEntityHitIDs = append([]string(nil), p.Direct.ContentEntityHitIDs...)
	out.Direct.EntityMissIDs = append([]string(nil), p.Direct.EntityMissIDs...)
	out.Direct.ContentMissIDs = append([]string(nil), p.Direct.ContentMissIDs...)
	if p.Route != nil {
		route := *p.Route
		route.EntityContentIDs = append([]string(nil), p.Route.EntityContentIDs...)
		route.DirectFTSEntityHitIDs = append([]string(nil), p.Route.DirectFTSEntityHitIDs...)
		route.DirectFTSContentHitIDs = append([]string(nil), p.Route.DirectFTSContentHitIDs...)
		route.DirectFTSContentMissIDs = append([]string(nil), p.Route.DirectFTSContentMissIDs...)
		route.ExpectedRouteFieldValues = cloneRouteEnrichment(p.Route.ExpectedRouteFieldValues)
		out.Route = &route
	}
	return &out
}

// searchRouteWitnesses converts only bound internal witnesses into the
// candidate response envelope. Invalid or absent internal witnesses are not
// promoted to public evidence.
func searchRouteWitnesses(rows []SearchResultRow, provenance map[string]*searchProvenance) []SearchRouteWitness {
	if len(rows) == 0 || len(provenance) == 0 {
		return nil
	}
	witnesses := make([]SearchRouteWitness, 0, len(rows))
	for _, row := range rows {
		p := provenance[row.ID]
		if p == nil || p.Route == nil || p.Route.Validate() != nil {
			continue
		}
		route := p.Route.ExpectedRouteFieldValues
		if !routeNonEmpty(route) {
			route = row.Route
		}
		witnesses = append(witnesses, SearchRouteWitness{
			EntityID:                p.Route.EntityID,
			EntityContentIDs:        append([]string(nil), p.Route.EntityContentIDs...),
			MatchSource:             p.Route.MatchSource,
			GraphFromID:             p.Route.GraphFromID,
			GraphRelType:            p.Route.GraphRelType,
			GraphToID:               p.Route.GraphToID,
			DirectFTSEntityHitIDs:   append([]string(nil), p.Route.DirectFTSEntityHitIDs...),
			DirectFTSContentHitIDs:  append([]string(nil), p.Route.DirectFTSContentHitIDs...),
			DirectFTSEntityMissID:   p.Route.DirectFTSEntityMissID,
			DirectFTSContentMissIDs: append([]string(nil), p.Route.DirectFTSContentMissIDs...),
			RouteFieldValues:        cloneRouteEnrichment(route),
		})
	}
	if len(witnesses) == 0 {
		return nil
	}
	return witnesses
}

func cloneRouteEnrichment(route RouteEnrichment) RouteEnrichment {
	route.Facts = append([]string(nil), route.Facts...)
	route.Graph = append([]string(nil), route.Graph...)
	route.Anchors = append([]string(nil), route.Anchors...)
	route.Lanes = append([]string(nil), route.Lanes...)
	route.Drift = append([]string(nil), route.Drift...)
	return route
}

func routeNonEmpty(route RouteEnrichment) bool {
	return len(route.Facts) > 0 || len(route.Graph) > 0 || len(route.Anchors) > 0 || len(route.Lanes) > 0 || len(route.Drift) > 0 || route.Hash != "" || route.HashBasis != ""
}

func sameRouteEnrichment(left, right RouteEnrichment) bool {
	left = cloneRouteEnrichment(left)
	right = cloneRouteEnrichment(right)
	return fmt.Sprintf("%#v", left) == fmt.Sprintf("%#v", right)
}

func parseGraphSource(source string) (relType, anchor string, ok bool) {
	if !strings.HasPrefix(source, "graph:") {
		return "", "", false
	}
	rest := strings.TrimPrefix(source, "graph:")
	idx := strings.IndexByte(rest, ':')
	if idx <= 0 || idx == len(rest)-1 {
		return "", "", false
	}
	return rest[:idx], rest[idx+1:], true
}

func sortSearchRows(rows []SearchResultRow, order map[string]int) {
	sort.SliceStable(rows, func(i, j int) bool {
		leftContent := hasSource(rows[i].MatchSources, "content_fts")
		rightContent := hasSource(rows[j].MatchSources, "content_fts")
		if leftContent != rightContent {
			return leftContent
		}
		leftGraphDoc := isGraphExpandedDocument(rows[i])
		rightGraphDoc := isGraphExpandedDocument(rows[j])
		if leftGraphDoc != rightGraphDoc {
			return leftGraphDoc
		}
		if len(rows[i].MatchSources) != len(rows[j].MatchSources) {
			return len(rows[i].MatchSources) > len(rows[j].MatchSources)
		}
		return order[rows[i].ID] < order[rows[j].ID]
	})
}

func isGraphExpandedDocument(row SearchResultRow) bool {
	if !isContentDocumentType(row.Type) {
		return false
	}
	for _, source := range row.MatchSources {
		if strings.HasPrefix(source, "graph:") {
			return true
		}
	}
	return false
}

func isContentDocumentType(entityType string) bool {
	switch entityType {
	case "system", "container", "component", "ref", "rule":
		return false
	default:
		return true
	}
}

func enrichSearchRow(s *store.Store, row *SearchResultRow) error {
	rels, err := s.RelationshipsFrom(row.ID)
	if err != nil {
		return fmt.Errorf("search graph %s: %w", row.ID, err)
	}
	for _, rel := range rels {
		target, err := s.GetEntity(rel.ToID)
		if err != nil {
			return fmt.Errorf("search graph target %s: %w", rel.ToID, err)
		}
		addMatchSource(row, "graph:"+rel.RelType+":"+rel.ToID)
		assignSearchContext(row, rel, target)
	}

	if row.Context.Component.ID != "" {
		if err := enrichFromComponent(s, row, row.Context.Component.ID); err != nil {
			return err
		}
	}
	sort.Strings(row.MatchSources)
	return nil
}

func enrichFromComponent(s *store.Store, row *SearchResultRow, componentID string) error {
	rels, err := s.RelationshipsFrom(componentID)
	if err != nil {
		return fmt.Errorf("search component graph %s: %w", componentID, err)
	}
	for _, rel := range rels {
		target, err := s.GetEntity(rel.ToID)
		if err != nil {
			return fmt.Errorf("search component graph target %s: %w", rel.ToID, err)
		}
		assignSearchContext(row, rel, target)
	}
	return nil
}

func buildSearchRoute(s *store.Store, c3Dir, projectDir string, row SearchResultRow, query string) RouteEnrichment {
	ids := []string{row.ID}
	for _, ref := range []EntityRef{row.Context.Component, row.Context.Ref, row.Context.Rule} {
		if ref.ID != "" {
			ids = append(ids, ref.ID)
		}
	}
	return buildRouteEnrichmentForIDs(s, c3Dir, projectDir, ids, query)
}

func assignSearchContext(row *SearchResultRow, rel *store.Relationship, target *store.Entity) {
	ref := EntityRef{ID: target.ID, Title: target.Title}
	switch target.Type {
	case "component":
		if rel.RelType == "affects" || rel.RelType == "sources" || row.Context.Component.ID == "" {
			row.Context.Component = ref
		}
	case "ref":
		if row.Context.Ref.ID == "" {
			row.Context.Ref = ref
		}
	case "rule":
		if row.Context.Rule.ID == "" {
			row.Context.Rule = ref
		}
	}
}

func cleanSnippet(snippet string) string {
	snippet = strings.ReplaceAll(snippet, ">>>", "")
	snippet = strings.ReplaceAll(snippet, "<<<", "")
	return strings.Join(strings.Fields(snippet), " ")
}

func bestContentSnippet(s *store.Store, entityID, query, fallback string) string {
	body, err := content.ReadEntity(s, entityID)
	if err != nil {
		return fallback
	}
	terms := searchTerms(query)
	best := fallback
	bestScore := 0
	for _, line := range strings.Split(body, "\n") {
		line = cleanSnippet(line)
		if line == "" {
			continue
		}
		score := termScore(line, terms)
		if score > bestScore {
			best = line
			bestScore = score
		}
	}
	return best
}

func termScore(text string, terms []string) int {
	text = strings.ToLower(text)
	score := 0
	for _, term := range terms {
		if strings.Contains(text, strings.ToLower(term)) {
			score++
		}
	}
	return score
}

func searchTerms(query string) []string {
	parts := strings.FieldsFunc(query, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-')
	})
	terms := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "-")
		if part == "" {
			continue
		}
		switch strings.ToUpper(part) {
		case "AND", "OR", "NOT", "NEAR":
			continue
		default:
			terms = append(terms, part)
		}
	}
	return terms
}

func ftsDisjunction(query string) string {
	terms := searchTerms(query)
	return strings.Join(terms, " OR ")
}

func addMatchSource(row *SearchResultRow, source string) {
	if source == "" || hasSource(row.MatchSources, source) {
		return
	}
	row.MatchSources = append(row.MatchSources, source)
}

func hasSource(sources []string, source string) bool {
	for _, existing := range sources {
		if existing == source {
			return true
		}
	}
	return false
}

func resolveSearchLimit(limit int) int {
	if limit > 0 {
		return limit
	}
	if isAgentMode() {
		return defaultAgentSearchLimit
	}
	return defaultSearchLimit
}

func compactSearchRows(rows []SearchResultRow) []compactSearchResultRow {
	out := make([]compactSearchResultRow, 0, len(rows))
	for _, row := range rows {
		snippet := compactSearchSnippet(row.Snippet)
		if isTitleDuplicateSnippet(row.Title, snippet) {
			snippet = ""
		}
		out = append(out, compactSearchResultRow{
			ID:      row.ID,
			Title:   row.Title,
			Why:     compactMatchSources(row.MatchSources),
			Ctx:     compactSearchContext(row.Context),
			Route:   compactRoute(row.Route),
			Snippet: snippet,
		})
	}
	return out
}

func isTitleDuplicateSnippet(title, snippet string) bool {
	title = normalizeSearchDuplicateText(title)
	snippet = normalizeSearchDuplicateText(snippet)
	return title != "" && title == snippet
}

func normalizeSearchDuplicateText(s string) string {
	s = strings.TrimSpace(cleanSnippet(s))
	s = strings.TrimLeft(s, "# ")
	s = strings.NewReplacer("-", " ", "_", " ").Replace(s)
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func compactSearchSnippet(snippet string) string {
	snippet = cleanSnippet(snippet)
	if snippet == "" {
		return ""
	}
	runes := []rune(snippet)
	if len(runes) <= compactSearchSnippetMax {
		return snippet
	}
	cut := compactSearchSnippetMax
	for i := compactSearchSnippetMax; i >= compactSearchSnippetMax/2; i-- {
		if unicode.IsSpace(runes[i-1]) {
			cut = i - 1
			break
		}
	}
	return strings.TrimSpace(string(runes[:cut])) + "..."
}

func compactMatchSources(sources []string) string {
	if len(sources) == 0 {
		return ""
	}
	parts := make([]string, 0, len(sources))
	seen := make(map[string]bool, len(sources))
	for _, source := range sources {
		part := compactMatchSource(source)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		parts = append(parts, part)
	}
	return strings.Join(parts, "+")
}

func compactMatchSource(source string) string {
	switch source {
	case "content_fts":
		return "body"
	case "entity_fts":
		return "meta"
	case "semantic":
		return "sem"
	}
	if strings.HasPrefix(source, "graph:") {
		return ""
	}
	return source
}

func compactSearchContext(ctx SearchContext) string {
	parts := make([]string, 0, 4)
	if ctx.Component.ID != "" {
		parts = append(parts, "comp="+ctx.Component.ID)
	}
	if ctx.Ref.ID != "" {
		parts = append(parts, "ref="+ctx.Ref.ID)
	}
	if ctx.Rule.ID != "" {
		parts = append(parts, "rule="+ctx.Rule.ID)
	}
	if ctx.Path != "" {
		parts = append(parts, "path="+ctx.Path)
	}
	return strings.Join(parts, " ")
}
