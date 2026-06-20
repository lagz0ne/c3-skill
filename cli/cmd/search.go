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
}

// SearchOutput is the structured search response.
type SearchOutput struct {
	Query   string            `json:"query"`
	Results []SearchResultRow `json:"results"`
}

// SearchResultRow is one ranked result with search and graph provenance.
type SearchResultRow struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Title        string        `json:"title"`
	Snippet      string        `json:"snippet"`
	MatchSources []string      `json:"match_sources"`
	Context      SearchContext `json:"context"`
}

// SearchContext captures the most useful graph/code-map context for a hit.
type SearchContext struct {
	Component EntityRef `json:"component"`
	Ref       EntityRef `json:"ref"`
	Rule      EntityRef `json:"rule"`
	Path      string    `json:"path"`
}

// EntityRef is a lightweight entity reference in search context.
type EntityRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// RunSearch performs semantic, content/entity FTS, and graph-context search.
func RunSearch(opts SearchOptions, w io.Writer) error {
	if opts.Store == nil {
		return fmt.Errorf("error: search store is required")
	}
	if strings.TrimSpace(opts.Query) == "" {
		return fmt.Errorf("error: search requires a <query> argument\nhint: c3x search \"pool wait\"")
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
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
	rows, err := collectSearchRows(opts.Store, opts.Query, entityType, pool)
	if err != nil {
		return err
	}
	rows, err = expandHybridRows(opts.Store, rows, opts.Query, entityType, pool)
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
	rows = fuseSemanticRows(rows, semanticRows, limit)
	for i := range rows {
		if err := enrichSearchRow(opts.Store, &rows[i]); err != nil {
			return err
		}
	}

	format := ResolveFormat(opts.JSON, isAgentMode())
	return WriteObjectOutput(w, SearchOutput{
		Query:   opts.Query,
		Results: rows,
	}, format, nil)
}

// candidatePoolLimit sizes the over-fetched candidate pool that feeds fusion. Each
// retrieval signal returns up to this many rows so fusion sees docs corroborated
// across signals before the single post-fusion truncation to the caller's limit.
// Sized at ~2x the cutoff: deep enough to recover a corroborated doc that sits just
// past the cutoff on a single signal, but not so deep that low-signal docs dilute the
// pool and evict a true hit (search-eval peaks at this ratio; 4x measurably regressed).
func candidatePoolLimit(limit int) int {
	if limit <= 0 {
		limit = 20
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

func fuseSemanticRows(rows []SearchResultRow, semanticHits []store.SearchResult, limit int) []SearchResultRow {
	if limit <= 0 {
		limit = 20
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
