package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store wraps an embedded SQLite database holding C3 entities,
// relationships, code-map entries, and a mutation changelog.
type Store struct {
	db     *sql.DB
	dbPath string
}

// Open creates or opens a SQLite database at dbPath, runs schema
// migrations, and returns a ready-to-use Store.
func Open(dbPath string) (*Store, error) {
	// DELETE journal mode avoids -wal/-shm sidecar files.
	dsn := dbPath + "?_pragma=journal_mode(delete)&_pragma=foreign_keys(on)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	s := &Store{db: db, dbPath: dbPath}
	if err := s.createSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return s, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for advanced queries.
func (s *Store) DB() *sql.DB {
	return s.db
}

// createSchema idempotently creates all tables and indexes.
func (s *Store) createSchema() error {
	_, err := s.db.Exec(schemaSQL)
	return err
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS entities (
	id          TEXT PRIMARY KEY,
	type        TEXT NOT NULL CHECK(type IN ('system','container','component')),
	title       TEXT NOT NULL,
	slug        TEXT NOT NULL,
	category    TEXT NOT NULL DEFAULT '',
	parent_id   TEXT REFERENCES entities(id) ON DELETE SET NULL,
	goal        TEXT NOT NULL DEFAULT '',
	summary     TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	body        TEXT NOT NULL DEFAULT '',
	status      TEXT NOT NULL DEFAULT 'active',
	boundary    TEXT NOT NULL DEFAULT '',
	date        TEXT NOT NULL DEFAULT '',
	metadata    TEXT NOT NULL DEFAULT '{}',
	created_at  TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS relationships (
	from_id  TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
	to_id    TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
	rel_type TEXT NOT NULL,
	PRIMARY KEY (from_id, to_id, rel_type)
);
CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_id);

CREATE TABLE IF NOT EXISTS code_map (
	entity_id TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
	pattern   TEXT NOT NULL,
	PRIMARY KEY (entity_id, pattern)
);

CREATE TABLE IF NOT EXISTS code_map_excludes (
	pattern TEXT PRIMARY KEY
);

CREATE VIRTUAL TABLE IF NOT EXISTS entities_fts USING fts5(
	title, goal, summary, description, body,
	content='entities',
	content_rowid='rowid'
);

-- FTS triggers: keep the FTS index in sync via rowid.
CREATE TRIGGER IF NOT EXISTS entities_ai AFTER INSERT ON entities BEGIN
	INSERT INTO entities_fts(rowid, title, goal, summary, description, body)
	VALUES (new.rowid, new.title, new.goal, new.summary, new.description, new.body);
END;
CREATE TRIGGER IF NOT EXISTS entities_ad AFTER DELETE ON entities BEGIN
	INSERT INTO entities_fts(entities_fts, rowid, title, goal, summary, description, body)
	VALUES ('delete', old.rowid, old.title, old.goal, old.summary, old.description, old.body);
END;
CREATE TRIGGER IF NOT EXISTS entities_au AFTER UPDATE ON entities BEGIN
	INSERT INTO entities_fts(entities_fts, rowid, title, goal, summary, description, body)
	VALUES ('delete', old.rowid, old.title, old.goal, old.summary, old.description, old.body);
	INSERT INTO entities_fts(rowid, title, goal, summary, description, body)
	VALUES (new.rowid, new.title, new.goal, new.summary, new.description, new.body);
END;

CREATE TABLE IF NOT EXISTS chunks (
	id        INTEGER PRIMARY KEY AUTOINCREMENT,
	entity_id TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
	heading   TEXT NOT NULL DEFAULT '',
	body      TEXT NOT NULL DEFAULT '',
	seq       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_chunks_entity ON chunks(entity_id);

CREATE TABLE IF NOT EXISTS changelog (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	entity_id   TEXT NOT NULL,
	action      TEXT NOT NULL,
	field       TEXT NOT NULL DEFAULT '',
	old_value   TEXT NOT NULL DEFAULT '',
	new_value   TEXT NOT NULL DEFAULT '',
	timestamp   TEXT NOT NULL DEFAULT (datetime('now')),
	commit_hash TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS store_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);
`
