package store

import "fmt"

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
	q := `SELECT e.id, e.type, e.title,
		snippet(entities_fts, -1, '>>>', '<<<', '...', 20),
		rank
		FROM entities_fts
		JOIN entities e ON entities_fts.rowid = e.rowid
		WHERE entities_fts MATCH ?`
	args := []any{query}
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

	rows, err := s.db.Query(q, query)
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
