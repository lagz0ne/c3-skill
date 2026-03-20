package store

import (
	"fmt"
	"strings"
)

// ImpactResult represents a node discovered during graph traversal.
type ImpactResult struct {
	ID    string
	Title string
	Type  string
	Depth int
}

// prefixColumns prepends a table alias to each column in a comma-separated list.
func prefixColumns(alias, columns string) string {
	parts := strings.Split(columns, ",")
	for i, p := range parts {
		parts[i] = alias + "." + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}

// RefsFor returns all entities that the given entity references via outbound relationships.
func (s *Store) RefsFor(entityID string) ([]*Entity, error) {
	q := `SELECT ` + prefixColumns("e", entityColumns) + `
		FROM entities e
		JOIN relationships r ON r.to_id = e.id
		WHERE r.from_id = ?
		ORDER BY e.id`
	return s.queryEntities(q, entityID)
}

// CitedBy returns all entities that reference the given entity via inbound relationships.
func (s *Store) CitedBy(entityID string) ([]*Entity, error) {
	q := `SELECT ` + prefixColumns("e", entityColumns) + `
		FROM entities e
		JOIN relationships r ON r.from_id = e.id
		WHERE r.to_id = ?
		ORDER BY e.id`
	return s.queryEntities(q, entityID)
}

// Impact performs reverse impact analysis: finds all entities affected by changes
// to the given entity. Traverses inbound 'uses' and outbound 'affects' relationships
// up to maxDepth levels.
func (s *Store) Impact(entityID string, maxDepth int) ([]ImpactResult, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	q := `
	WITH RECURSIVE impact(id, depth) AS (
		SELECT ?, 0
		UNION
		SELECT r.from_id, impact.depth + 1
		FROM relationships r
		JOIN impact ON r.to_id = impact.id
		WHERE r.rel_type = 'uses'
		AND impact.depth < ?
		UNION
		SELECT r.to_id, impact.depth + 1
		FROM relationships r
		JOIN impact ON r.from_id = impact.id
		WHERE r.rel_type = 'affects'
		AND impact.depth < ?
	)
	SELECT DISTINCT e.id, e.title, e.type, i.depth
	FROM impact i
	JOIN entities e ON e.id = i.id
	WHERE i.id != ?
	ORDER BY i.depth, e.id`

	rows, err := s.db.Query(q, entityID, maxDepth, maxDepth, entityID)
	if err != nil {
		return nil, fmt.Errorf("impact: %w", err)
	}
	defer rows.Close()

	return scanImpactResults(rows)
}

// Transitive performs forward graph traversal from the given entity through
// any relationship type, up to maxDepth levels.
func (s *Store) Transitive(entityID string, maxDepth int) ([]ImpactResult, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	q := `
	WITH RECURSIVE traverse(id, depth) AS (
		SELECT ?, 0
		UNION
		SELECT r.to_id, traverse.depth + 1
		FROM relationships r
		JOIN traverse ON r.from_id = traverse.id
		WHERE traverse.depth < ?
	)
	SELECT DISTINCT e.id, e.title, e.type, t.depth
	FROM traverse t
	JOIN entities e ON e.id = t.id
	WHERE t.id != ?
	ORDER BY t.depth, e.id`

	rows, err := s.db.Query(q, entityID, maxDepth, entityID)
	if err != nil {
		return nil, fmt.Errorf("transitive: %w", err)
	}
	defer rows.Close()

	return scanImpactResults(rows)
}

// scanImpactResults scans rows into a slice of ImpactResult.
func scanImpactResults(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}) ([]ImpactResult, error) {
	var results []ImpactResult
	for rows.Next() {
		var r ImpactResult
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Depth); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
