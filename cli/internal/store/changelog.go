package store

// ChangeEntry represents a single mutation log entry.
type ChangeEntry struct {
	ID         int
	EntityID   string
	Action     string
	Field      string
	OldValue   string
	NewValue   string
	Timestamp  string
	CommitHash string
}

// logChange records a mutation in the changelog table.
func (s *Store) logChange(entityID, action, field, oldVal, newVal string) {
	s.db.Exec(`
		INSERT INTO changelog (entity_id, action, field, old_value, new_value)
		VALUES (?, ?, ?, ?, ?)`,
		entityID, action, field, oldVal, newVal,
	)
}

// UnmarkedChanges returns all changelog entries with an empty commit_hash.
func (s *Store) UnmarkedChanges() ([]*ChangeEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, entity_id, action, field, old_value, new_value, timestamp, commit_hash
		FROM changelog
		WHERE commit_hash = ''
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*ChangeEntry
	for rows.Next() {
		var c ChangeEntry
		if err := rows.Scan(&c.ID, &c.EntityID, &c.Action, &c.Field,
			&c.OldValue, &c.NewValue, &c.Timestamp, &c.CommitHash); err != nil {
			return nil, err
		}
		result = append(result, &c)
	}
	return result, rows.Err()
}

// MarkChangelog stamps all unmarked changelog entries with a commit hash.
func (s *Store) MarkChangelog(commitHash string) error {
	_, err := s.db.Exec(`
		UPDATE changelog SET commit_hash = ? WHERE commit_hash = ''`,
		commitHash,
	)
	return err
}
