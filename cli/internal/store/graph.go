package store

import (
	"strings"
)

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
