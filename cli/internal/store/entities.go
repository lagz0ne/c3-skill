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
	_, err := s.exec.Exec(`
		INSERT INTO entities (id, type, title, slug, category, parent_id, goal,
			status, boundary, date, metadata, root_merkle, version)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Type, e.Title, e.Slug, e.Category, e.ParentID, e.Goal,
		e.Status, e.Boundary, e.Date, e.Metadata, e.RootMerkle, e.Version,
	)
	if err != nil {
		return fmt.Errorf("insert entity %s: %w", e.ID, err)
	}
	return nil
}

func (s *Store) GetEntity(id string) (*Entity, error) {
	row := s.exec.QueryRow(`SELECT `+entityColumns+` FROM entities WHERE id = ?`, id)
	return scanEntity(row)
}

// UpdateEntity writes an entity's body/metadata fields. It is EDIT-PROOF for status:
// the status column is NEVER written by this path — the existing stored status is
// preserved regardless of e.Status. The only DB path that may move the status column
// is SetEntityStatus (used by the status command, supersede, the auto-done latch, and
// migration). This makes the omission-demotion bug structurally impossible: no body
// write/import/repair can un-freeze a terminal doc (status is edit-proof).
func (s *Store) UpdateEntity(e *Entity) error {
	_, err := s.GetEntity(e.ID)
	if err != nil {
		return fmt.Errorf("update entity: get old: %w", err)
	}

	_, err = s.exec.Exec(`
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

	return nil
}

// SetEntityStatus is the privileged, dedicated status writer — the ONLY DB path that
// moves the status column. It is used by the status command, supersede, the auto-done
// latch, and migration. Because UpdateEntity is status-edit-proof, this is the sole
// way status changes, keeping terminal docs immutable from body paths.
func (s *Store) SetEntityStatus(id, status string) error {
	_, err := s.GetEntity(id)
	if err != nil {
		return fmt.Errorf("set status: get old: %w", err)
	}
	if _, err := s.exec.Exec(`
		UPDATE entities SET status = ?, updated_at = datetime('now') WHERE id = ?`,
		status, id,
	); err != nil {
		return fmt.Errorf("set status %s=%s: %w", id, status, err)
	}
	return nil
}

func (s *Store) DeleteEntity(id string) error {
	res, err := s.exec.Exec(`DELETE FROM entities WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete entity %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("entity %s not found", id)
	}
	return nil
}

func (s *Store) AllEntities() ([]*Entity, error) {
	return s.queryEntities(`SELECT ` + entityColumns + ` FROM entities ORDER BY type, id`)
}

func (s *Store) EntitiesByType(typ string) ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities WHERE type = ? ORDER BY id`, typ)
}

func (s *Store) Children(parentID string) ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities WHERE parent_id = ? ORDER BY id`, parentID)
}

func (s *Store) queryEntities(query string, args ...any) ([]*Entity, error) {
	rows, err := s.exec.Query(query, args...)
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
