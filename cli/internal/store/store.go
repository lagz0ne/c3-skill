package store

import (
	"database/sql"
	"fmt"
	"strings"

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
	// WAL mode allows concurrent readers + single writer across sessions.
	// busy_timeout(5000) waits up to 5s on lock contention instead of failing.
	dsn := dbPath + "?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1) // serialize access within a single process
	s := &Store{db: db, dbPath: dbPath}
	if err := s.createSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	if err := s.migrateSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate schema: %w", err)
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
	type        TEXT NOT NULL CHECK(type IN ('system','container','component','ref','adr','rule','recipe')),
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
	title, goal,
	content='entities',
	content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS entities_ai AFTER INSERT ON entities BEGIN
	INSERT INTO entities_fts(rowid, title, goal)
	VALUES (new.rowid, new.title, new.goal);
END;
CREATE TRIGGER IF NOT EXISTS entities_ad AFTER DELETE ON entities BEGIN
	INSERT INTO entities_fts(entities_fts, rowid, title, goal)
	VALUES ('delete', old.rowid, old.title, old.goal);
END;
CREATE TRIGGER IF NOT EXISTS entities_au AFTER UPDATE ON entities BEGIN
	INSERT INTO entities_fts(entities_fts, rowid, title, goal)
	VALUES ('delete', old.rowid, old.title, old.goal);
	INSERT INTO entities_fts(rowid, title, goal)
	VALUES (new.rowid, new.title, new.goal);
END;

-- Content node tree: every markdown element has identity + hash.
CREATE TABLE IF NOT EXISTS nodes (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	entity_id   TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
	parent_id   INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
	type        TEXT NOT NULL,
	level       INTEGER NOT NULL DEFAULT 0,
	seq         INTEGER NOT NULL,
	content     TEXT NOT NULL DEFAULT '',
	hash        TEXT NOT NULL DEFAULT ''
);
-- Composite index: covers entity lookups and ORDER BY parent_id, seq.
-- COALESCE handles NULL parent_id (SQLite treats NULL != NULL in UNIQUE).
CREATE UNIQUE INDEX IF NOT EXISTS idx_nodes_order
	ON nodes(entity_id, COALESCE(parent_id, 0), seq);

-- Content FTS: full-text search over node content.
CREATE VIRTUAL TABLE IF NOT EXISTS content_fts USING fts5(
	content,
	content='nodes',
	content_rowid='rowid'
);
CREATE TRIGGER IF NOT EXISTS content_fts_ai AFTER INSERT ON nodes BEGIN
	INSERT INTO content_fts(rowid, content) VALUES (new.rowid, new.content);
END;
CREATE TRIGGER IF NOT EXISTS content_fts_ad AFTER DELETE ON nodes BEGIN
	INSERT INTO content_fts(content_fts, rowid, content)
	VALUES ('delete', old.rowid, old.content);
END;
CREATE TRIGGER IF NOT EXISTS content_fts_au AFTER UPDATE ON nodes BEGIN
	INSERT INTO content_fts(content_fts, rowid, content)
	VALUES ('delete', old.rowid, old.content);
	INSERT INTO content_fts(rowid, content) VALUES (new.rowid, new.content);
END;

-- Version history: full content snapshots per entity.
CREATE TABLE IF NOT EXISTS versions (
	entity_id   TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
	version     INTEGER NOT NULL,
	content     TEXT NOT NULL DEFAULT '',
	root_merkle TEXT NOT NULL DEFAULT '',
	commit_hash TEXT NOT NULL DEFAULT '',
	created_at  TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (entity_id, version)
);

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

func (s *Store) migrateSchema() error {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('entities') WHERE name = 'root_merkle'`).Scan(&count)
	if count > 0 {
		return nil
	}
	migrations := []string{
		`ALTER TABLE entities ADD COLUMN root_merkle TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE entities ADD COLUMN version INTEGER NOT NULL DEFAULT 0`,
	}
	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
