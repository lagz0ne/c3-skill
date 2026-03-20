package store

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
)

// SetCodeMap replaces all code-map globs for an entity.
func (s *Store) SetCodeMap(entityID string, globs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("set code map begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM code_map WHERE entity_id = ?`, entityID); err != nil {
		return fmt.Errorf("set code map clear: %w", err)
	}
	for _, g := range globs {
		if _, err := tx.Exec(`INSERT INTO code_map (entity_id, pattern) VALUES (?, ?)`, entityID, g); err != nil {
			return fmt.Errorf("set code map insert: %w", err)
		}
	}
	return tx.Commit()
}

// CodeMapFor returns the glob patterns associated with an entity.
func (s *Store) CodeMapFor(entityID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT pattern FROM code_map WHERE entity_id = ? ORDER BY pattern`, entityID)
	if err != nil {
		return nil, fmt.Errorf("code map for: %w", err)
	}
	defer rows.Close()

	var patterns []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

// LookupByFile matches a file path against all code-map globs and returns
// the entity IDs that own matching patterns. Excluded patterns are filtered out.
func (s *Store) LookupByFile(filePath string) ([]string, error) {
	// Normalize the path to forward slashes for consistent matching.
	filePath = filepath.ToSlash(filePath)

	excludes, err := s.Excludes()
	if err != nil {
		return nil, err
	}
	for _, ex := range excludes {
		matched, _ := doublestar.Match(ex, filePath)
		if matched {
			return nil, nil
		}
	}

	allMap, err := s.AllCodeMap()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var result []string
	for entityID, patterns := range allMap {
		for _, p := range patterns {
			matched, _ := doublestar.Match(p, filePath)
			if matched && !seen[entityID] {
				seen[entityID] = true
				result = append(result, entityID)
			}
		}
	}

	// Sort for deterministic output.
	sortStrings(result)
	return result, nil
}

// AddExclude adds a global exclusion pattern.
func (s *Store) AddExclude(glob string) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO code_map_excludes (pattern) VALUES (?)`, glob)
	if err != nil {
		return fmt.Errorf("add exclude: %w", err)
	}
	return nil
}

// Excludes returns all global exclusion patterns.
func (s *Store) Excludes() ([]string, error) {
	rows, err := s.db.Query(`SELECT pattern FROM code_map_excludes ORDER BY pattern`)
	if err != nil {
		return nil, fmt.Errorf("excludes: %w", err)
	}
	defer rows.Close()

	var patterns []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

// AllCodeMap returns the full mapping of entity IDs to their glob patterns.
func (s *Store) AllCodeMap() (map[string][]string, error) {
	rows, err := s.db.Query(`SELECT entity_id, pattern FROM code_map ORDER BY entity_id, pattern`)
	if err != nil {
		return nil, fmt.Errorf("all code map: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var entityID, pattern string
		if err := rows.Scan(&entityID, &pattern); err != nil {
			return nil, err
		}
		result[entityID] = append(result[entityID], pattern)
	}
	return result, rows.Err()
}

func sortStrings(s []string) {
	sort.Strings(s)
}
