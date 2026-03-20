package store

import "testing"

func TestOpen_CreatesSchema(t *testing.T) {
	s := createTestStore(t)

	// Verify core tables exist by querying sqlite_master.
	tables := []string{"entities", "relationships", "code_map", "code_map_excludes",
		"chunks", "changelog", "store_meta"}
	for _, table := range tables {
		var name string
		err := s.DB().QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}

	// Verify FTS virtual table.
	var ftsName string
	err := s.DB().QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='entities_fts'`,
	).Scan(&ftsName)
	if err != nil {
		t.Errorf("FTS table entities_fts not found: %v", err)
	}
}

func TestOpen_InvalidPath(t *testing.T) {
	// Try to open a database at a path that can't be created
	_, err := Open("/nonexistent/deep/path/test.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestOpen_Idempotent(t *testing.T) {
	s := createTestStore(t)

	// Running createSchema again should not error.
	if err := s.createSchema(); err != nil {
		t.Fatalf("second createSchema failed: %v", err)
	}

	// Insert an entity to ensure the DB is functional after double-init.
	e := &Entity{
		ID: "test-sys", Type: "system", Title: "Test", Slug: "test",
		Status: "active", Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert after double-init: %v", err)
	}
}
