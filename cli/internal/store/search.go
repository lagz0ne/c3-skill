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
	args := []interface{}{query}
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
