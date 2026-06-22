package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpen_CreatesSchema(t *testing.T) {
	s := createTestStore(t)

	// Verify core tables exist by querying sqlite_master.
	tables := []string{"entities", "relationships",
		"nodes", "versions", "store_meta"}
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

func TestOpen_MigratesEntityTypeCheckForCanvasTypes(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "c3.db")
	raw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`
		CREATE TABLE entities (
			id          TEXT PRIMARY KEY,
			type        TEXT NOT NULL CHECK(type IN ('system','container','component','ref','adr','rule')),
			title       TEXT NOT NULL,
			slug        TEXT NOT NULL,
			category    TEXT NOT NULL DEFAULT '',
			parent_id   TEXT REFERENCES entities(id) ON DELETE SET NULL,
			goal        TEXT NOT NULL DEFAULT '',
			status      TEXT NOT NULL DEFAULT 'active',
			boundary    TEXT NOT NULL DEFAULT '',
			date        TEXT NOT NULL DEFAULT '',
			metadata    TEXT NOT NULL DEFAULT '{}',
			root_merkle TEXT NOT NULL DEFAULT '',
			version     INTEGER NOT NULL DEFAULT 0,
			created_at  TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
		);
		INSERT INTO entities(id, type, title, slug, status, metadata)
		VALUES ('c3-0', 'system', 'System', 'system', 'active', '{}');
	`); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	s, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var tableSQL string
	if err := s.DB().QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='entities'`).Scan(&tableSQL); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(tableSQL, "CHECK(type IN") {
		t.Fatalf("entities type CHECK was not migrated: %s", tableSQL)
	}

	if err := s.InsertEntity(&Entity{
		ID: "research-note-api-latency", Type: "research-note", Title: "API Latency", Slug: "api-latency",
		Goal: "Investigate API latency.", Status: "active", Metadata: "{}",
	}); err != nil {
		t.Fatalf("custom canvas entity insert after migration: %v", err)
	}
}
