package store

import (
	"database/sql"
	"fmt"
)

type Entity struct {
	ID         string
	Type       string
	Title      string
	Slug       string
	Category   string
	ParentID   string
	Goal       string
	Status     string
	Boundary   string
	Date       string
	Metadata   string // JSON blob
	RootMerkle string
	Version    int
	CreatedAt  string
	UpdatedAt  string
}

const entityColumns = `id, type, title, slug, category, parent_id, goal,
	status, boundary, date, metadata, root_merkle, version, created_at, updated_at`

func (s *Store) InsertEntity(e *Entity) error {
	_, err := s.db.Exec(`
		INSERT INTO entities (id, type, title, slug, category, parent_id, goal,
			status, boundary, date, metadata, root_merkle, version)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Type, e.Title, e.Slug, e.Category, e.ParentID, e.Goal,
		e.Status, e.Boundary, e.Date, e.Metadata, e.RootMerkle, e.Version,
	)
	if err != nil {
		return fmt.Errorf("insert entity %s: %w", e.ID, err)
	}
	s.logChange(e.ID, "add", "", "", "")
	return nil
}

func (s *Store) GetEntity(id string) (*Entity, error) {
	row := s.db.QueryRow(`SELECT `+entityColumns+` FROM entities WHERE id = ?`, id)
	return scanEntity(row)
}

// UpdateEntity writes an entity's body/metadata fields. It is EDIT-PROOF for status:
// the status column is NEVER written by this path — the existing stored status is
// preserved regardless of e.Status. The only DB path that may move the status column
// is SetEntityStatus (used by the status command, supersede, the auto-done latch, and
// migration). This makes the omission-demotion bug structurally impossible: no body
// write/import/repair can un-freeze a terminal doc (status is edit-proof).
func (s *Store) UpdateEntity(e *Entity) error {
	old, err := s.GetEntity(e.ID)
	if err != nil {
		return fmt.Errorf("update entity: get old: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE entities SET
			type = ?, title = ?, slug = ?, category = ?,
			parent_id = NULLIF(?, ''), goal = ?,
			boundary = ?,
			date = ?, metadata = ?, root_merkle = ?, version = ?,
			updated_at = datetime('now')
		WHERE id = ?`,
		e.Type, e.Title, e.Slug, e.Category,
		e.ParentID, e.Goal,
		e.Boundary,
		e.Date, e.Metadata, e.RootMerkle, e.Version,
		e.ID,
	)
	if err != nil {
		return fmt.Errorf("update entity %s: %w", e.ID, err)
	}

	logFieldChange(s, e.ID, "title", old.Title, e.Title)
	logFieldChange(s, e.ID, "type", old.Type, e.Type)
	logFieldChange(s, e.ID, "slug", old.Slug, e.Slug)
	logFieldChange(s, e.ID, "category", old.Category, e.Category)
	logFieldChange(s, e.ID, "parent_id", old.ParentID, e.ParentID)
	logFieldChange(s, e.ID, "goal", old.Goal, e.Goal)
	logFieldChange(s, e.ID, "boundary", old.Boundary, e.Boundary)
	logFieldChange(s, e.ID, "date", old.Date, e.Date)
	logFieldChange(s, e.ID, "metadata", old.Metadata, e.Metadata)
	logFieldChange(s, e.ID, "root_merkle", old.RootMerkle, e.RootMerkle)

	return nil
}

// SetEntityStatus is the privileged, dedicated status writer — the ONLY DB path that
// moves the status column. It is used by the status command, supersede, the auto-done
// latch, and migration. Because UpdateEntity is status-edit-proof, this is the sole
// way status changes, keeping terminal docs immutable from body paths.
func (s *Store) SetEntityStatus(id, status string) error {
	old, err := s.GetEntity(id)
	if err != nil {
		return fmt.Errorf("set status: get old: %w", err)
	}
	if _, err := s.db.Exec(`
		UPDATE entities SET status = ?, updated_at = datetime('now') WHERE id = ?`,
		status, id,
	); err != nil {
		return fmt.Errorf("set status %s=%s: %w", id, status, err)
	}
	logFieldChange(s, id, "status", old.Status, status)
	return nil
}

func logFieldChange(s *Store, entityID, field, oldVal, newVal string) {
	if oldVal != newVal {
		s.logChange(entityID, "update", field, oldVal, newVal)
	}
}

func (s *Store) DeleteEntity(id string) error {
	res, err := s.db.Exec(`DELETE FROM entities WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete entity %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("entity %s not found", id)
	}
	s.logChange(id, "delete", "", "", "")
	return nil
}

func (s *Store) AllEntities() ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities ORDER BY type, id`)
}

func (s *Store) EntitiesByType(typ string) ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities WHERE type = ? ORDER BY id`, typ)
}

func (s *Store) Children(parentID string) ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities WHERE parent_id = ? ORDER BY id`, parentID)
}

func (s *Store) queryEntities(query string, args ...any) ([]*Entity, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Entity
	for rows.Next() {
		e, err := scanEntityFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func scanEntity(row *sql.Row) (*Entity, error) {
	var e Entity
	var parentID sql.NullString
	err := row.Scan(
		&e.ID, &e.Type, &e.Title, &e.Slug, &e.Category,
		&parentID, &e.Goal,
		&e.Status, &e.Boundary,
		&e.Date, &e.Metadata, &e.RootMerkle, &e.Version,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.ParentID = parentID.String
	return &e, nil
}

// LegacyBody reads the body column from a pre-v8 database via raw SQL.
// Returns empty string if the column doesn't exist.
func (s *Store) LegacyBody(entityID string) string {
	var body string
	s.db.QueryRow(`SELECT body FROM entities WHERE id = ?`, entityID).Scan(&body)
	return body
}

func scanEntityFromRows(rows *sql.Rows) (*Entity, error) {
	var e Entity
	var parentID sql.NullString
	err := rows.Scan(
		&e.ID, &e.Type, &e.Title, &e.Slug, &e.Category,
		&parentID, &e.Goal,
		&e.Status, &e.Boundary,
		&e.Date, &e.Metadata, &e.RootMerkle, &e.Version,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.ParentID = parentID.String
	return &e, nil
}
