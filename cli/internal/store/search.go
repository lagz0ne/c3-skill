package store

import (
	"fmt"
	"strings"
	"unicode"
)

// sanitizeFTS5 converts user input into a safe FTS5 query.
// Preserves AND/OR/NOT as boolean operators when between search terms.
// Converts | to OR. Strips all other FTS5 syntax characters.
// Returns empty string if input contains no searchable words.
func sanitizeFTS5(input string) string {
	// Pass 1: tokenize — extract words and pipe symbols.
	var tokens []string
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		word := strings.Trim(current.String(), "-")
		current.Reset()
		if word != "" {
			tokens = append(tokens, word)
		}
	}
	for _, r := range input {
		if r == '|' {
			flush()
			tokens = append(tokens, "|")
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	// Pass 2: classify tokens as operators or terms, enforce valid placement.
	// An operator is only valid when it has a term before and after it.
	type classified struct {
		text string
		isOp bool
	}
	var parts []classified
	for _, tok := range tokens {
		upper := strings.ToUpper(tok)
		if upper == "AND" || upper == "OR" || upper == "NOT" || upper == "|" {
			parts = append(parts, classified{text: "OR", isOp: upper == "OR" || upper == "|"})
			if upper == "AND" {
				parts[len(parts)-1].text = "AND"
			}
			if upper == "NOT" {
				parts[len(parts)-1].text = "NOT"
			}
			parts[len(parts)-1].isOp = true
		} else if upper == "NEAR" {
			// NEAR(n) syntax — too dangerous, skip entirely.
			continue
		} else {
			parts = append(parts, classified{text: tok, isOp: false})
		}
	}

	// Pass 3: build output — only emit operators that sit between terms.
	var out []string
	lastWasTerm := false
	for _, p := range parts {
		if p.isOp {
			if lastWasTerm {
				out = append(out, p.text)
				lastWasTerm = false
			}
			// else: dangling leading/consecutive operator — skip
		} else {
			if !lastWasTerm && len(out) > 0 {
				// Previous was an operator waiting for a right-hand term.
				// Check that the last item in out is actually an operator.
				last := out[len(out)-1]
				if last == "AND" || last == "OR" || last == "NOT" {
					// Good — operator has both sides now.
				}
			} else if !lastWasTerm && len(out) > 0 {
				// Remove trailing dangling operator.
				lastOut := out[len(out)-1]
				if lastOut == "AND" || lastOut == "OR" || lastOut == "NOT" {
					out = out[:len(out)-1]
				}
			}
			out = append(out, p.text)
			lastWasTerm = true
		}
	}
	// Strip trailing operator.
	if len(out) > 0 {
		last := out[len(out)-1]
		if last == "AND" || last == "OR" || last == "NOT" {
			out = out[:len(out)-1]
		}
	}
	return strings.Join(out, " ")
}

// EntitySummary is a lightweight entity reference for suggestions and samples.
type EntitySummary struct {
	ID    string
	Type  string
	Title string
}

// SuggestEntities finds entities whose title or ID partially matches the query.
// Uses LIKE for fuzzy matching — safe with any input. Returns up to limit results.
func (s *Store) SuggestEntities(query string, limit int, excludeTypes ...string) ([]EntitySummary, error) {
	if limit <= 0 {
		limit = 5
	}
	// Extract words from query for LIKE matching.
	safe := sanitizeFTS5(query)
	if safe == "" {
		return nil, nil
	}
	words := strings.Fields(safe)

	// Build LIKE clauses for each word against title, id, slug, and goal.
	// Uses substring (%word%) and short-prefix (first 2 chars%) for typo tolerance.
	var conditions []string
	var args []any
	for _, w := range words {
		substr := "%" + w + "%"
		col := "(LOWER(title) LIKE LOWER(?) OR LOWER(id) LIKE LOWER(?) OR LOWER(slug) LIKE LOWER(?) OR LOWER(goal) LIKE LOWER(?)"
		argSet := []any{substr, substr, substr, substr}

		// Prefix matching with first 2 characters catches transposition typos
		// (e.g., "auht" → "au%" matches "auth").
		if len(w) >= 3 {
			prefix := w[:2] + "%"
			col += " OR LOWER(title) LIKE LOWER(?) OR LOWER(id) LIKE LOWER(?) OR LOWER(goal) LIKE LOWER(?)"
			argSet = append(argSet, prefix, prefix, prefix)
		}
		col += ")"
		conditions = append(conditions, col)
		args = append(args, argSet...)
	}
	where := strings.Join(conditions, " AND ")
	if len(excludeTypes) > 0 {
		placeholders := make([]string, len(excludeTypes))
		for i, t := range excludeTypes {
			placeholders[i] = "?"
			args = append(args, t)
		}
		where += " AND type NOT IN (" + strings.Join(placeholders, ",") + ")"
	}
	q := fmt.Sprintf(
		`SELECT id, type, title FROM entities WHERE %s ORDER BY LENGTH(title) LIMIT %d`,
		where, limit,
	)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("suggest: %w", err)
	}
	defer rows.Close()

	var results []EntitySummary
	for rows.Next() {
		var e EntitySummary
		if err := rows.Scan(&e.ID, &e.Type, &e.Title); err != nil {
			return nil, fmt.Errorf("suggest scan: %w", err)
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

// SampleEntities returns a random sample of entities for discovery.
func (s *Store) SampleEntities(limit int) ([]EntitySummary, error) {
	if limit <= 0 {
		limit = 5
	}
	q := fmt.Sprintf(
		`SELECT id, type, title FROM entities WHERE type IN ('container','component','ref') ORDER BY RANDOM() LIMIT %d`,
		limit,
	)
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("sample: %w", err)
	}
	defer rows.Close()

	var results []EntitySummary
	for rows.Next() {
		var e EntitySummary
		if err := rows.Scan(&e.ID, &e.Type, &e.Title); err != nil {
			return nil, fmt.Errorf("sample scan: %w", err)
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

// SearchWithCount performs FTS search and also returns the total match count
// (not limited), so callers can show "showing N of M".
func (s *Store) SearchWithCount(query, entityType string, limit int) ([]SearchResult, int, error) {
	if limit <= 0 {
		limit = 20
	}
	safe := sanitizeFTS5(query)
	if safe == "" {
		return nil, 0, nil
	}

	// Count query.
	cq := `SELECT COUNT(*) FROM entities_fts JOIN entities e ON entities_fts.rowid = e.rowid WHERE entities_fts MATCH ?`
	cArgs := []any{safe}
	if entityType != "" {
		cq += " AND e.type = ?"
		cArgs = append(cArgs, entityType)
	}
	var total int
	if err := s.db.QueryRow(cq, cArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("search count: %w", err)
	}

	results, err := s.searchFTS(query, entityType, limit)
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// SearchResult represents a single full-text search hit.
type SearchResult struct {
	ID      string
	Type    string
	Title   string
	Snippet string
	Rank    float64
}

// Search performs a full-text search across all entity types.
func (s *Store) Search(query string) ([]SearchResult, error) {
	return s.searchFTS(query, "", 20)
}

// SearchWithFilter searches entities restricted to a specific type.
func (s *Store) SearchWithFilter(query, entityType string) ([]SearchResult, error) {
	return s.searchFTS(query, entityType, 20)
}

// SearchWithLimit searches with a type filter and custom result limit.
func (s *Store) SearchWithLimit(query, entityType string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.searchFTS(query, entityType, limit)
}

func (s *Store) searchFTS(query, entityType string, limit int) ([]SearchResult, error) {
	safe := sanitizeFTS5(query)
	if safe == "" {
		return nil, nil // no searchable terms
	}
	q := `SELECT e.id, e.type, e.title,
		snippet(entities_fts, -1, '>>>', '<<<', '...', 20),
		rank
		FROM entities_fts
		JOIN entities e ON entities_fts.rowid = e.rowid
		WHERE entities_fts MATCH ?`
	args := []any{safe}
	if entityType != "" {
		q += " AND e.type = ?"
		args = append(args, entityType)
	}
	q += fmt.Sprintf(" ORDER BY rank LIMIT %d", limit)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Type, &r.Title, &r.Snippet, &r.Rank); err != nil {
			return nil, fmt.Errorf("search scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SearchContent searches node content via content_fts, grouping results by entity.
// Returns one result per entity with the best-matching node's content as snippet.
func (s *Store) SearchContent(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	safe := sanitizeFTS5(query)
	if safe == "" {
		return nil, nil
	}
	// Over-fetch then dedup in Go — snippet() can't be used with GROUP BY.
	q := fmt.Sprintf(`
		SELECT e.id, e.type, e.title,
			snippet(content_fts, 0, '>>>', '<<<', '...', 20),
			content_fts.rank
		FROM content_fts
		JOIN nodes n ON content_fts.rowid = n.rowid
		JOIN entities e ON n.entity_id = e.id
		WHERE content_fts MATCH ?
		ORDER BY content_fts.rank
		LIMIT %d`, limit*3) // over-fetch to allow dedup

	rows, err := s.db.Query(q, safe)
	if err != nil {
		return nil, fmt.Errorf("search content: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Type, &r.Title, &r.Snippet, &r.Rank); err != nil {
			return nil, fmt.Errorf("search content scan: %w", err)
		}
		if seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		results = append(results, r)
		if len(results) >= limit {
			break
		}
	}
	return results, rows.Err()
}
