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

func (s *Store) UpdateEntity(e *Entity) error {
	old, err := s.GetEntity(e.ID)
	if err != nil {
		return fmt.Errorf("update entity: get old: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE entities SET
			type = ?, title = ?, slug = ?, category = ?,
			parent_id = NULLIF(?, ''), goal = ?,
			status = ?, boundary = ?,
			date = ?, metadata = ?, root_merkle = ?, version = ?,
			updated_at = datetime('now')
		WHERE id = ?`,
		e.Type, e.Title, e.Slug, e.Category,
		e.ParentID, e.Goal,
		e.Status, e.Boundary,
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
	logFieldChange(s, e.ID, "status", old.Status, e.Status)
	logFieldChange(s, e.ID, "boundary", old.Boundary, e.Boundary)
	logFieldChange(s, e.ID, "date", old.Date, e.Date)
	logFieldChange(s, e.ID, "metadata", old.Metadata, e.Metadata)
	logFieldChange(s, e.ID, "root_merkle", old.RootMerkle, e.RootMerkle)

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
