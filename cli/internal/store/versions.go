package store

import (
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
	err := s.exec.QueryRow(`
		SELECT COALESCE(MAX(version), 0) + 1 FROM versions WHERE entity_id = ?`,
		entityID,
	).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("next version: %w", err)
	}

	_, err = s.exec.Exec(`
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
	err := s.exec.QueryRow(`
		SELECT entity_id, version, content, root_merkle, commit_hash, created_at
		FROM versions WHERE entity_id = ? AND version = ?`,
		entityID, version,
	).Scan(&v.EntityID, &v.Version, &v.Content, &v.RootMerkle, &v.CommitHash, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
