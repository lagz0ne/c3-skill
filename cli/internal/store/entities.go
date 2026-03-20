package store

import (
	"database/sql"
	"fmt"
)

// Entity represents a C3 architectural entity (system, container, or component).
type Entity struct {
	ID          string
	Type        string // system | container | component
	Title       string
	Slug        string
	Category    string
	ParentID    string // empty string = no parent
	Goal        string
	Summary     string
	Description string
	Body        string
	Status      string
	Boundary    string
	Date        string
	Metadata    string // JSON blob
	CreatedAt   string
	UpdatedAt   string
}

// Explicit column list — never use SELECT *.
const entityColumns = `id, type, title, slug, category, parent_id, goal, summary,
	description, body, status, boundary, date, metadata, created_at, updated_at`

// InsertEntity stores a new entity and logs the creation.
func (s *Store) InsertEntity(e *Entity) error {
	_, err := s.db.Exec(`
		INSERT INTO entities (id, type, title, slug, category, parent_id, goal,
			summary, description, body, status, boundary, date, metadata)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Type, e.Title, e.Slug, e.Category, e.ParentID, e.Goal,
		e.Summary, e.Description, e.Body, e.Status, e.Boundary, e.Date, e.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert entity %s: %w", e.ID, err)
	}
	s.logChange(e.ID, "add", "", "", "")
	return nil
}

// GetEntity retrieves an entity by ID. Returns sql.ErrNoRows if not found.
func (s *Store) GetEntity(id string) (*Entity, error) {
	row := s.db.QueryRow(`SELECT `+entityColumns+` FROM entities WHERE id = ?`, id)
	return scanEntity(row)
}

// UpdateEntity updates mutable fields and logs each changed field.
func (s *Store) UpdateEntity(e *Entity) error {
	old, err := s.GetEntity(e.ID)
	if err != nil {
		return fmt.Errorf("update entity: get old: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE entities SET
			type = ?, title = ?, slug = ?, category = ?,
			parent_id = NULLIF(?, ''), goal = ?, summary = ?,
			description = ?, body = ?, status = ?, boundary = ?,
			date = ?, metadata = ?,
			updated_at = datetime('now')
		WHERE id = ?`,
		e.Type, e.Title, e.Slug, e.Category,
		e.ParentID, e.Goal, e.Summary,
		e.Description, e.Body, e.Status, e.Boundary,
		e.Date, e.Metadata,
		e.ID,
	)
	if err != nil {
		return fmt.Errorf("update entity %s: %w", e.ID, err)
	}

	// Log each changed field.
	logFieldChange(s, e.ID, "title", old.Title, e.Title)
	logFieldChange(s, e.ID, "type", old.Type, e.Type)
	logFieldChange(s, e.ID, "slug", old.Slug, e.Slug)
	logFieldChange(s, e.ID, "category", old.Category, e.Category)
	logFieldChange(s, e.ID, "parent_id", old.ParentID, e.ParentID)
	logFieldChange(s, e.ID, "goal", old.Goal, e.Goal)
	logFieldChange(s, e.ID, "summary", old.Summary, e.Summary)
	logFieldChange(s, e.ID, "description", old.Description, e.Description)
	logFieldChange(s, e.ID, "body", old.Body, e.Body)
	logFieldChange(s, e.ID, "status", old.Status, e.Status)
	logFieldChange(s, e.ID, "boundary", old.Boundary, e.Boundary)
	logFieldChange(s, e.ID, "date", old.Date, e.Date)
	logFieldChange(s, e.ID, "metadata", old.Metadata, e.Metadata)

	return nil
}

// logFieldChange logs a single field change if values differ.
func logFieldChange(s *Store, entityID, field, oldVal, newVal string) {
	if oldVal != newVal {
		s.logChange(entityID, "update", field, oldVal, newVal)
	}
}

// DeleteEntity removes an entity. Returns an error if it does not exist.
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

// AllEntities returns every entity ordered by type then ID.
func (s *Store) AllEntities() ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities ORDER BY type, id`)
}

// EntitiesByType returns entities matching the given type.
func (s *Store) EntitiesByType(typ string) ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities WHERE type = ? ORDER BY id`, typ)
}

// Children returns direct children of the given parent entity.
func (s *Store) Children(parentID string) ([]*Entity, error) {
	return s.queryEntities(`SELECT `+entityColumns+` FROM entities WHERE parent_id = ? ORDER BY id`, parentID)
}

// queryEntities executes a query and scans all rows into Entity slices.
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

// scanEntity scans a single *sql.Row into an Entity.
func scanEntity(row *sql.Row) (*Entity, error) {
	var e Entity
	var parentID sql.NullString
	err := row.Scan(
		&e.ID, &e.Type, &e.Title, &e.Slug, &e.Category,
		&parentID, &e.Goal, &e.Summary,
		&e.Description, &e.Body, &e.Status, &e.Boundary,
		&e.Date, &e.Metadata, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.ParentID = parentID.String
	return &e, nil
}

// scanEntityFromRows scans the current row of *sql.Rows into an Entity.
func scanEntityFromRows(rows *sql.Rows) (*Entity, error) {
	var e Entity
	var parentID sql.NullString
	err := rows.Scan(
		&e.ID, &e.Type, &e.Title, &e.Slug, &e.Category,
		&parentID, &e.Goal, &e.Summary,
		&e.Description, &e.Body, &e.Status, &e.Boundary,
		&e.Date, &e.Metadata, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.ParentID = parentID.String
	return &e, nil
}
