package store

import (
	"database/sql"
	"fmt"
)

// Version represents a point-in-time snapshot of an entity's content.
type Version struct {
	EntityID   string
	Version    int
	Content    string // full rendered content at this version
	RootMerkle string
	CommitHash string
	CreatedAt  string
}

// CreateVersion inserts a new version snapshot for an entity.
func (s *Store) CreateVersion(entityID string, content, rootMerkle string) (*Version, error) {
	var nextVersion int
	err := s.db.QueryRow(`
		SELECT COALESCE(MAX(version), 0) + 1 FROM versions WHERE entity_id = ?`,
		entityID,
	).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("next version: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO versions (entity_id, version, content, root_merkle)
		VALUES (?, ?, ?, ?)`,
		entityID, nextVersion, content, rootMerkle,
	)
	if err != nil {
		return nil, fmt.Errorf("insert version: %w", err)
	}

	return &Version{
		EntityID:   entityID,
		Version:    nextVersion,
		RootMerkle: rootMerkle,
		Content:    content,
	}, nil
}

// GetVersion retrieves a specific version of an entity.
func (s *Store) GetVersion(entityID string, version int) (*Version, error) {
	var v Version
	err := s.db.QueryRow(`
		SELECT entity_id, version, content, root_merkle, commit_hash, created_at
		FROM versions WHERE entity_id = ? AND version = ?`,
		entityID, version,
	).Scan(&v.EntityID, &v.Version, &v.Content, &v.RootMerkle, &v.CommitHash, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ListVersions returns all versions for an entity, newest first.
// Content is omitted (empty string) to avoid loading large blobs.
func (s *Store) ListVersions(entityID string) ([]*Version, error) {
	rows, err := s.db.Query(`
		SELECT entity_id, version, '', root_merkle, commit_hash, created_at
		FROM versions WHERE entity_id = ?
		ORDER BY version DESC`,
		entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	defer rows.Close()

	var result []*Version
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.EntityID, &v.Version, &v.Content, &v.RootMerkle, &v.CommitHash, &v.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &v)
	}
	return result, rows.Err()
}

// LatestVersion returns the most recent version number for an entity, or 0 if none.
func (s *Store) LatestVersion(entityID string) (int, error) {
	var v int
	err := s.db.QueryRow(`
		SELECT COALESCE(MAX(version), 0) FROM versions WHERE entity_id = ?`,
		entityID,
	).Scan(&v)
	return v, err
}

// PruneVersions deletes versions older than keepLast for an entity.
// Versions with a non-empty commit_hash are always kept.
func (s *Store) PruneVersions(entityID string, keepLast int) (int64, error) {
	res, err := s.db.Exec(`
		DELETE FROM versions
		WHERE entity_id = ?
		  AND commit_hash = ''
		  AND version <= (
			SELECT COALESCE(MAX(version), 0) - ? FROM versions WHERE entity_id = ?
		  )`,
		entityID, keepLast, entityID,
	)
	if err != nil {
		return 0, fmt.Errorf("prune versions: %w", err)
	}
	return res.RowsAffected()
}

// MarkVersion stamps a version with a git commit hash.
func (s *Store) MarkVersion(entityID string, version int, commitHash string) error {
	res, err := s.db.Exec(`
		UPDATE versions SET commit_hash = ? WHERE entity_id = ? AND version = ?`,
		commitHash, entityID, version,
	)
	if err != nil {
		return fmt.Errorf("mark version: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
