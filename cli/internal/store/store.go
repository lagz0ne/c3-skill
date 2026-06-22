package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// dbExecutor is the read/write surface shared by *sql.DB and *sql.Tx. Data-layer
// store methods issue every statement through this seam so the SAME method runs
// either autocommitted (against the pool) or enlisted in a transaction (against a
// *sql.Tx) with no signature change. It deliberately omits Begin: a missed
// inner-Begin site fails to compile rather than deadlocking under MaxOpenConns(1).
type dbExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Store wraps an embedded SQLite database holding C3 entities and
// relationships.
type Store struct {
	db     *sql.DB
	exec   dbExecutor // the pool by default; a *sql.Tx inside WithTx
	dbPath string
}

// WithTx runs fn inside a single transaction: every store method fn calls on the
// supplied *Store enlists in it, so the whole closure commits or rolls back as one
// unit. Calls nest — if s is already transactional, fn reuses the open tx and the
// outer WithTx owns the commit (SQLite has no real nested transactions). This is
// the only mutation path that spans multiple writes atomically (change apply).
func (s *Store) WithTx(fn func(*Store) error) error {
	if _, already := s.exec.(*sql.Tx); already {
		return fn(s) // reuse the open tx; the outermost WithTx commits
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	txStore := &Store{db: s.db, exec: tx, dbPath: s.dbPath}
	// Roll back on panic before re-raising: otherwise a panicking closure leaks the
	// open transaction and, under MaxOpenConns(1), permanently blocks the one pooled
	// connection (every later query would hang on busy_timeout).
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(txStore); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// WithPreviewTx runs fn inside a transaction that is ALWAYS rolled back — a
// read-only overlay. Writes are visible to fn via read-your-writes (so applying a
// change-unit's staged patches inside it produces a previewed graph) but are never
// committed. A nested WithTx reuses this tx, so the real apply path runs inside.
func (s *Store) WithPreviewTx(fn func(*Store) error) error {
	if _, already := s.exec.(*sql.Tx); already {
		// Under MaxOpenConns(1) a nested Begin would wait forever on the only
		// connection. A preview must roll back, so it can't reuse an outer tx that
		// will commit — nesting is a programming error, refused loudly.
		return fmt.Errorf("preview tx cannot run inside an open transaction")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin preview tx: %w", err)
	}
	txStore := &Store{db: s.db, exec: tx, dbPath: s.dbPath}
	defer func() {
		_ = tx.Rollback() // a preview never commits, even on success
		if p := recover(); p != nil {
			panic(p)
		}
	}()
	return fn(txStore)
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
	s := &Store{db: db, exec: db, dbPath: dbPath}
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
	type        TEXT NOT NULL,
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

CREATE TABLE IF NOT EXISTS entity_embeddings (
	entity_id TEXT PRIMARY KEY REFERENCES entities(id) ON DELETE CASCADE,
	model     TEXT NOT NULL,
	dims      INTEGER NOT NULL,
	text_hash TEXT NOT NULL,
	vector    BLOB NOT NULL,
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_entity_embeddings_model
	ON entity_embeddings(model, dims);

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

CREATE TABLE IF NOT EXISTS store_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);
`

func (s *Store) migrateSchema() error {
	if err := s.ensureEntityMerkleColumns(); err != nil {
		return err
	}
	if err := s.ensureEntityTypeIsCanvasDefined(); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureEntityMerkleColumns() error {
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

func (s *Store) ensureEntityTypeIsCanvasDefined() error {
	var tableSQL string
	err := s.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'entities'`).Scan(&tableSQL)
	if err != nil {
		return fmt.Errorf("inspect entities schema: %w", err)
	}
	if !strings.Contains(tableSQL, "CHECK(type IN") {
		return nil
	}

	if _, err := s.db.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
		return fmt.Errorf("disable foreign keys for entity type migration: %w", err)
	}
	defer s.db.Exec(`PRAGMA foreign_keys=ON`)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin entity type migration: %w", err)
	}
	defer tx.Rollback()

	migrations := []string{
		`DROP TRIGGER IF EXISTS entities_ai`,
		`DROP TRIGGER IF EXISTS entities_ad`,
		`DROP TRIGGER IF EXISTS entities_au`,
		`DROP TABLE IF EXISTS entities_new`,
		`CREATE TABLE entities_new (
			id          TEXT PRIMARY KEY,
			type        TEXT NOT NULL,
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
		)`,
		`INSERT INTO entities_new(rowid, id, type, title, slug, category, parent_id, goal, status, boundary, date, metadata, root_merkle, version, created_at, updated_at)
		 SELECT rowid, id, type, title, slug, category, parent_id, goal, status, boundary, date, metadata, root_merkle, version, created_at, updated_at FROM entities`,
		`DROP TABLE entities`,
		`ALTER TABLE entities_new RENAME TO entities`,
	}
	for _, migration := range migrations {
		if _, err := tx.Exec(migration); err != nil {
			return fmt.Errorf("migrate entity type constraint: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit entity type migration: %w", err)
	}
	if err := s.createSchema(); err != nil {
		return fmt.Errorf("recreate entity triggers: %w", err)
	}
	if _, err := s.db.Exec(`INSERT INTO entities_fts(entities_fts) VALUES ('rebuild')`); err != nil {
		return fmt.Errorf("rebuild entity search index: %w", err)
	}
	return nil
}
