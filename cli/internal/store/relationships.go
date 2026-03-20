package store

import "fmt"

// Relationship represents a directed edge between two entities.
type Relationship struct {
	FromID  string
	ToID    string
	RelType string
}

// AddRelationship inserts a relationship, ignoring duplicates.
// Only logs when a new row is actually inserted.
func (s *Store) AddRelationship(r *Relationship) error {
	res, err := s.db.Exec(`
		INSERT OR IGNORE INTO relationships (from_id, to_id, rel_type)
		VALUES (?, ?, ?)`,
		r.FromID, r.ToID, r.RelType,
	)
	if err != nil {
		return fmt.Errorf("add relationship: %w", err)
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		s.logChange(r.FromID, "add_rel", "", "", r.ToID+":"+r.RelType)
	}
	return nil
}

// RemoveRelationship deletes a specific relationship.
func (s *Store) RemoveRelationship(r *Relationship) error {
	_, err := s.db.Exec(`
		DELETE FROM relationships
		WHERE from_id = ? AND to_id = ? AND rel_type = ?`,
		r.FromID, r.ToID, r.RelType,
	)
	if err != nil {
		return fmt.Errorf("remove relationship: %w", err)
	}
	s.logChange(r.FromID, "remove_rel", "", r.ToID+":"+r.RelType, "")
	return nil
}

// RelationshipsFrom returns all outbound relationships from an entity.
func (s *Store) RelationshipsFrom(entityID string) ([]*Relationship, error) {
	return s.queryRelationships(
		`SELECT from_id, to_id, rel_type FROM relationships WHERE from_id = ? ORDER BY to_id`,
		entityID,
	)
}

// RelationshipsTo returns all inbound relationships to an entity.
func (s *Store) RelationshipsTo(entityID string) ([]*Relationship, error) {
	return s.queryRelationships(
		`SELECT from_id, to_id, rel_type FROM relationships WHERE to_id = ? ORDER BY from_id`,
		entityID,
	)
}

// RelationshipsByType returns all relationships of a given type.
func (s *Store) RelationshipsByType(relType string) ([]*Relationship, error) {
	return s.queryRelationships(
		`SELECT from_id, to_id, rel_type FROM relationships WHERE rel_type = ? ORDER BY from_id, to_id`,
		relType,
	)
}

// queryRelationships executes a query and scans rows into Relationship slices.
func (s *Store) queryRelationships(query string, args ...any) ([]*Relationship, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.FromID, &r.ToID, &r.RelType); err != nil {
			return nil, err
		}
		result = append(result, &r)
	}
	return result, rows.Err()
}
