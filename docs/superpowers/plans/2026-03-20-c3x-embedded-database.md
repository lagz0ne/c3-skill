# c3x Embedded Database Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace file-based `.c3/` content with SQLite database at `.c3/c3.db`, making c3x the single authority for all architectural queries.

**Architecture:** New `internal/store/` package owns all DB access via `modernc.org/sqlite` (pure Go, no CGo). Existing `walker/frontmatter/writer` packages become migration-only. All commands adapted to read/write through store. FTS5 for full-text search. Vector search deferred to `@c3x/embed` sidecar (future chunk).

**Tech Stack:** Go 1.25, `modernc.org/sqlite`, SQLite FTS5, existing `gopkg.in/yaml.v3`

**Spec:** `docs/superpowers/specs/2026-03-20-c3x-embedded-database-design.md`

---

## File Structure

### New files

| File | Responsibility |
|------|---------------|
| `cli/internal/store/store.go` | DB lifecycle: Open, Close, schema creation, migrations |
| `cli/internal/store/entities.go` | Entity CRUD: Insert, Get, Update, Delete, List, ByType |
| `cli/internal/store/relationships.go` | Relationship CRUD: Add, Remove, query by from/to/type |
| `cli/internal/store/graph.go` | Graph traversal: Children, Forward, Reverse, Transitive, Impact |
| `cli/internal/store/search.go` | FTS5 search: Match, SearchWithFilter, Snippet |
| `cli/internal/store/codemap.go` | Code map: Set, Lookup, Coverage, Excludes |
| `cli/internal/store/changelog.go` | Changelog: Log, Diff, Mark |
| `cli/internal/store/store_test.go` | Core DB tests: schema, open/close |
| `cli/internal/store/entities_test.go` | Entity CRUD tests |
| `cli/internal/store/relationships_test.go` | Relationship tests |
| `cli/internal/store/graph_test.go` | Graph traversal tests |
| `cli/internal/store/search_test.go` | FTS5 search tests |
| `cli/internal/store/codemap_test.go` | Code map tests |
| `cli/internal/store/changelog_test.go` | Changelog tests |
| `cli/internal/store/testhelper_test.go` | Shared test fixture: `createTestStore()` |
| `cli/cmd/migrate.go` | `c3x migrate` command |
| `cli/cmd/migrate_test.go` | Migration tests |
| `cli/cmd/query.go` | `c3x query` command (FTS5) |
| `cli/cmd/query_test.go` | Query tests |
| `cli/cmd/diff.go` | `c3x diff` command |
| `cli/cmd/diff_test.go` | Diff tests |
| `cli/cmd/impact.go` | `c3x impact` command |
| `cli/cmd/impact_test.go` | Impact tests |
| `cli/cmd/export.go` | `c3x export` command |
| `cli/cmd/export_test.go` | Export tests |

### Modified files

| File | Change |
|------|--------|
| `cli/go.mod` | Add `modernc.org/sqlite` dependency |
| `cli/main.go` | Replace walker-based dispatch with store-based dispatch |
| `cli/cmd/options.go` | Add new command names + flags (`--fts`, `--mark`, `--keep-originals`) |
| `cli/cmd/init.go` | Create `.c3/c3.db` instead of markdown files |
| `cli/cmd/list.go` | Read from store instead of graph |
| `cli/cmd/add.go` | Write to store instead of files |
| `cli/cmd/add_rich.go` | Write to store instead of files |
| `cli/cmd/set.go` | Update via store instead of writer |
| `cli/cmd/wire.go` | Add/remove relationships via store |
| `cli/cmd/delete.go` | Delete via store with cascading cleanup |
| `cli/cmd/lookup.go` | Query store code_map instead of codemap package |
| `cli/cmd/check_enhanced.go` | Validate via store queries |
| `cli/cmd/graph.go` | Generate from store graph queries |
| `cli/cmd/codemap.go` | Manage code_map table |
| `cli/cmd/coverage.go` | Query store code_map coverage |
| `cli/cmd/helpers.go` | Remove `findEntityFile`, `writeEntityFile`; add store helpers |
| `cli/cmd/testhelper_test.go` | Add `createDBFixture()` alongside existing `createFixture()` |

### Unchanged files

| File | Reason |
|------|--------|
| `cli/internal/frontmatter/` | Reused by migration command |
| `cli/internal/walker/` | Reused by migration command |
| `cli/internal/markdown/` | Still used for body section parsing in set/check commands |
| `cli/internal/schema/` | Static section definitions, unchanged |
| `cli/internal/numbering/` | Modified: replace `*walker.C3Graph` params with `[]*store.Entity` slices (see Task 12) |
| `cli/internal/templates/` | Template rendering unchanged (used by init, add) |
| `cli/internal/wiring/` | Container table updates still needed |
| `cli/cmd/schema.go` | Static, no data access |
| `cli/cmd/help.go` | Static, add new command entries |

---

## Chunk 1: Store Package Foundation

### Task 1: Add SQLite dependency + store skeleton

**Files:**
- Modify: `cli/go.mod`
- Create: `cli/internal/store/store.go`
- Create: `cli/internal/store/store_test.go`

- [ ] **Step 1: Add modernc.org/sqlite dependency**

```bash
cd cli && go get modernc.org/sqlite
```

- [ ] **Step 2: Write failing test — store opens and creates schema**

```go
// cli/internal/store/store_test.go
package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_CreatesSchema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "c3.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("db file should exist")
	}

	// Verify tables exist
	tables := []string{"entities", "relationships", "code_map", "code_map_excludes", "chunks", "changelog", "store_meta"}
	for _, tbl := range tables {
		var name string
		err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&name)
		if err != nil {
			t.Errorf("table %s should exist: %v", tbl, err)
		}
	}

	// Verify FTS5 virtual table
	var ftsName string
	err = s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='entities_fts'").Scan(&ftsName)
	if err != nil {
		t.Error("entities_fts virtual table should exist")
	}

	// Verify schema_version in store_meta
	var version string
	err = s.db.QueryRow("SELECT value FROM store_meta WHERE key='schema_version'").Scan(&version)
	if err != nil {
		t.Fatal("schema_version should be set")
	}
	if version != "1" {
		t.Errorf("schema_version=%s, want 1", version)
	}
}

func TestOpen_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "c3.db")

	s1, _ := Open(dbPath)
	s1.Close()

	// Opening again should not fail or corrupt
	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	defer s2.Close()
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd cli && go test ./internal/store/ -run TestOpen -v`
Expected: FAIL — package doesn't exist yet

- [ ] **Step 4: Implement store.go**

```go
// cli/internal/store/store.go
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store provides access to the C3 SQLite database.
type Store struct {
	db     *sql.DB
	dbPath string
}

// Open opens (or creates) the C3 database at the given path.
func Open(dbPath string) (*Store, error) {
	// Use DELETE journal mode (not WAL) — avoids -wal/-shm sidecar files
	// which would pollute git. WAL is for concurrent access; c3x is single-process.
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(delete)&_pragma=foreign_keys(on)")
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

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for advanced queries.
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) createSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS entities (
			id          TEXT PRIMARY KEY,
			type        TEXT NOT NULL,
			title       TEXT NOT NULL,
			slug        TEXT,
			category    TEXT,
			parent_id   TEXT,
			goal        TEXT,
			summary     TEXT,
			description TEXT,
			body        TEXT,
			status      TEXT DEFAULT 'active',
			boundary    TEXT,
			date        TEXT,
			metadata    TEXT,
			created_at  TEXT DEFAULT (datetime('now')),
			updated_at  TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (parent_id) REFERENCES entities(id)
		)`,
		`CREATE TABLE IF NOT EXISTS relationships (
			from_id   TEXT NOT NULL,
			to_id     TEXT NOT NULL,
			rel_type  TEXT NOT NULL,
			PRIMARY KEY (from_id, to_id, rel_type),
			FOREIGN KEY (from_id) REFERENCES entities(id),
			FOREIGN KEY (to_id) REFERENCES entities(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_id)`,
		`CREATE TABLE IF NOT EXISTS code_map (
			entity_id TEXT NOT NULL,
			glob      TEXT NOT NULL,
			PRIMARY KEY (entity_id, glob),
			FOREIGN KEY (entity_id) REFERENCES entities(id)
		)`,
		`CREATE TABLE IF NOT EXISTS code_map_excludes (
			glob TEXT PRIMARY KEY
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS entities_fts USING fts5(
			title, goal, summary, body,
			content=entities, content_rowid=rowid,
			tokenize='porter unicode61'
		)`,
		// FTS triggers
		`CREATE TRIGGER IF NOT EXISTS entities_fts_insert AFTER INSERT ON entities BEGIN
			INSERT INTO entities_fts(rowid, title, goal, summary, body)
			VALUES (new.rowid, new.title, new.goal, new.summary, new.body);
		END`,
		`CREATE TRIGGER IF NOT EXISTS entities_fts_update AFTER UPDATE ON entities BEGIN
			INSERT INTO entities_fts(entities_fts, rowid, title, goal, summary, body)
			VALUES ('delete', old.rowid, old.title, old.goal, old.summary, old.body);
			INSERT INTO entities_fts(rowid, title, goal, summary, body)
			VALUES (new.rowid, new.title, new.goal, new.summary, new.body);
		END`,
		`CREATE TRIGGER IF NOT EXISTS entities_fts_delete AFTER DELETE ON entities BEGIN
			INSERT INTO entities_fts(entities_fts, rowid, title, goal, summary, body)
			VALUES ('delete', old.rowid, old.title, old.goal, old.summary, old.body);
		END`,
		`CREATE TABLE IF NOT EXISTS chunks (
			chunk_id   TEXT PRIMARY KEY,
			entity_id  TEXT NOT NULL,
			seq        INTEGER NOT NULL,
			content    TEXT NOT NULL,
			model      TEXT,
			embedded_at TEXT,
			FOREIGN KEY (entity_id) REFERENCES entities(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_entity ON chunks(entity_id)`,
		`CREATE TABLE IF NOT EXISTS changelog (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id   TEXT NOT NULL,
			action      TEXT NOT NULL,
			field       TEXT,
			old_value   TEXT,
			new_value   TEXT,
			timestamp   TEXT DEFAULT (datetime('now')),
			commit_hash TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS store_meta (
			key   TEXT PRIMARY KEY,
			value TEXT
		)`,
		`INSERT OR IGNORE INTO store_meta (key, value) VALUES ('schema_version', '1')`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt[:40], err)
		}
	}
	return nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd cli && go test ./internal/store/ -run TestOpen -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cli/internal/store/store.go cli/internal/store/store_test.go cli/go.mod cli/go.sum
git commit -m "feat(store): add SQLite store package with schema creation"
```

---

### Task 2: Entity CRUD

**Files:**
- Create: `cli/internal/store/entities.go`
- Create: `cli/internal/store/entities_test.go`
- Create: `cli/internal/store/testhelper_test.go`

- [ ] **Step 1: Create shared test helper**

```go
// cli/internal/store/testhelper_test.go
package store

import "testing"

// createTestStore opens a temp in-memory store for tests.
func createTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("createTestStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// seedFixture populates store with standard test data matching cmd/testhelper_test.go fixture.
func seedFixture(t *testing.T, s *Store) {
	t.Helper()
	entities := []Entity{
		{ID: "c3-0", Type: "context", Title: "TestProject", Goal: "Test the system.", Body: "# TestProject\n\n## Goal\n\nTest the system."},
		{ID: "c3-1", Type: "container", Title: "api", ParentID: "c3-0", Goal: "Serve API requests", Boundary: "service", Body: "# api\n\n## Goal\n\nServe API requests"},
		{ID: "c3-2", Type: "container", Title: "web", ParentID: "c3-0", Boundary: "app", Body: "# web\n\n## Goal\n\nWeb frontend."},
		{ID: "c3-101", Type: "component", Title: "auth", ParentID: "c3-1", Category: "foundation", Goal: "Handle authentication.", Body: "# auth\n\n## Goal\n\nHandle authentication."},
		{ID: "c3-110", Type: "component", Title: "users", ParentID: "c3-1", Category: "feature", Goal: "Manage user accounts.", Body: "# users\n\n## Goal\n\nManage user accounts."},
		{ID: "ref-jwt", Type: "ref", Title: "JWT Authentication", Goal: "Standardize auth tokens", Body: "# JWT Authentication\n\n## Goal\n\nStandardize auth tokens."},
		{ID: "adr-20260226-use-go", Type: "adr", Title: "Use Go for CLI", Status: "proposed", Date: "20260226", Body: "# Use Go for CLI\n\n## Context\n\nNeed fast CLI."},
	}
	for _, e := range entities {
		if err := s.InsertEntity(&e); err != nil {
			t.Fatalf("seed %s: %v", e.ID, err)
		}
	}
	// Relationships
	rels := [][3]string{
		{"c3-101", "ref-jwt", "uses"},
		{"ref-jwt", "c3-1", "scope"},
		{"adr-20260226-use-go", "c3-0", "affects"},
	}
	for _, r := range rels {
		if err := s.AddRelationship(r[0], r[1], r[2]); err != nil {
			t.Fatalf("seed rel %v: %v", r, err)
		}
	}
}
```

- [ ] **Step 2: Write failing entity tests**

```go
// cli/internal/store/entities_test.go
package store

import "testing"

func TestInsertAndGetEntity(t *testing.T) {
	s := createTestStore(t)

	e := &Entity{
		ID:       "c3-1",
		Type:     "container",
		Title:    "api",
		Goal:     "Serve API requests",
		Boundary: "service",
		Body:     "# api",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := s.GetEntity("c3-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "api" {
		t.Errorf("title=%q, want api", got.Title)
	}
	if got.Boundary != "service" {
		t.Errorf("boundary=%q, want service", got.Boundary)
	}
}

func TestGetEntity_NotFound(t *testing.T) {
	s := createTestStore(t)
	_, err := s.GetEntity("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}

func TestUpdateEntity(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.UpdateEntity("c3-101", map[string]interface{}{
		"goal": "Updated goal",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := s.GetEntity("c3-101")
	if got.Goal != "Updated goal" {
		t.Errorf("goal=%q, want Updated goal", got.Goal)
	}
}

func TestDeleteEntity(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.DeleteEntity("c3-110")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = s.GetEntity("c3-110")
	if err == nil {
		t.Fatal("entity should be deleted")
	}
}

func TestDeleteEntity_CascadesRelationships(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// c3-101 uses ref-jwt
	err := s.DeleteEntity("c3-101")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	rels, _ := s.RelationshipsFrom("c3-101")
	if len(rels) != 0 {
		t.Errorf("expected 0 relationships, got %d", len(rels))
	}
}

func TestAllEntities(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	all, err := s.AllEntities()
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 7 {
		t.Errorf("got %d entities, want 7", len(all))
	}
}

func TestEntitiesByType(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	comps, _ := s.EntitiesByType("component")
	if len(comps) != 2 {
		t.Errorf("got %d components, want 2", len(comps))
	}

	containers, _ := s.EntitiesByType("container")
	if len(containers) != 2 {
		t.Errorf("got %d containers, want 2", len(containers))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd cli && go test ./internal/store/ -run TestInsert -v`
Expected: FAIL — Entity type not defined

- [ ] **Step 4: Implement entities.go**

```go
// cli/internal/store/entities.go
package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// Entity represents a C3 entity stored in the database.
type Entity struct {
	ID          string
	Type        string
	Title       string
	Slug        string
	Category    string
	ParentID    string
	Goal        string
	Summary     string
	Description string
	Body        string
	Status      string
	Boundary    string
	Date        string
	Metadata    string // JSON
	CreatedAt   string
	UpdatedAt   string
}

// InsertEntity adds a new entity to the database.
func (s *Store) InsertEntity(e *Entity) error {
	_, err := s.db.Exec(`INSERT INTO entities
		(id, type, title, slug, category, parent_id, goal, summary, description, body, status, boundary, date, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Type, e.Title, e.Slug,
		nullEmpty(e.Category), nullEmpty(e.ParentID),
		nullEmpty(e.Goal), nullEmpty(e.Summary), nullEmpty(e.Description),
		nullEmpty(e.Body),
		coalesce(e.Status, "active"),
		nullEmpty(e.Boundary), nullEmpty(e.Date), nullEmpty(e.Metadata),
	)
	return err
}

// GetEntity retrieves an entity by ID.
func (s *Store) GetEntity(id string) (*Entity, error) {
	e := &Entity{}
	err := s.db.QueryRow(`SELECT
		id, type, title, COALESCE(slug,''), COALESCE(category,''),
		COALESCE(parent_id,''), COALESCE(goal,''), COALESCE(summary,''),
		COALESCE(description,''), COALESCE(body,''), COALESCE(status,'active'),
		COALESCE(boundary,''), COALESCE(date,''), COALESCE(metadata,''),
		COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM entities WHERE id = ?`, id).Scan(
		&e.ID, &e.Type, &e.Title, &e.Slug, &e.Category,
		&e.ParentID, &e.Goal, &e.Summary, &e.Description,
		&e.Body, &e.Status, &e.Boundary, &e.Date, &e.Metadata,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entity %q not found", id)
	}
	return e, err
}

// UpdateEntity updates specified fields on an entity. Logs changes to changelog.
func (s *Store) UpdateEntity(id string, fields map[string]interface{}) error {
	allowed := map[string]bool{
		"title": true, "slug": true, "category": true, "parent_id": true,
		"goal": true, "summary": true, "description": true, "body": true,
		"status": true, "boundary": true, "date": true, "metadata": true,
	}

	// Get old values for changelog
	old, err := s.GetEntity(id)
	if err != nil {
		return err
	}

	var setClauses []string
	var args []interface{}
	for k, v := range fields {
		if !allowed[k] {
			return fmt.Errorf("field %q not updatable", k)
		}
		setClauses = append(setClauses, k+" = ?")
		args = append(args, v)
	}
	setClauses = append(setClauses, "updated_at = datetime('now')")
	args = append(args, id)

	_, err = s.db.Exec(
		"UPDATE entities SET "+strings.Join(setClauses, ", ")+" WHERE id = ?",
		args...,
	)
	if err != nil {
		return err
	}

	// Log each field change
	for k, v := range fields {
		oldVal := getEntityField(old, k)
		newVal := fmt.Sprintf("%v", v)
		if oldVal != newVal {
			s.logChange(id, "modify", k, oldVal, newVal)
		}
	}
	return nil
}

// DeleteEntity removes an entity and its relationships. Logs to changelog.
func (s *Store) DeleteEntity(id string) error {
	// Verify entity exists
	if _, err := s.GetEntity(id); err != nil {
		return fmt.Errorf("cannot delete: %w", err)
	}
	// Remove relationships
	s.db.Exec("DELETE FROM relationships WHERE from_id = ? OR to_id = ?", id, id)
	// Remove code map entries
	s.db.Exec("DELETE FROM code_map WHERE entity_id = ?", id)
	// Remove chunks
	s.db.Exec("DELETE FROM chunks WHERE entity_id = ?", id)
	// Remove entity
	result, err := s.db.Exec("DELETE FROM entities WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n > 0 {
		s.logChange(id, "delete", "", "", "")
	}
	return nil
}

// entityColumns is the explicit column list for entity queries — must match scan order.
const entityColumns = `id, type, title, slug, category, parent_id, goal, summary,
	description, body, status, boundary, date, metadata, created_at, updated_at`

// AllEntities returns all entities ordered by ID.
func (s *Store) AllEntities() ([]*Entity, error) {
	return s.queryEntities("SELECT "+entityColumns+" FROM entities ORDER BY id")
}

// EntitiesByType returns entities of a given type.
func (s *Store) EntitiesByType(entityType string) ([]*Entity, error) {
	return s.queryEntities("SELECT "+entityColumns+" FROM entities WHERE type = ? ORDER BY id", entityType)
}

// Children returns entities whose parent_id matches the given ID.
func (s *Store) Children(parentID string) ([]*Entity, error) {
	return s.queryEntities("SELECT "+entityColumns+" FROM entities WHERE parent_id = ? ORDER BY id", parentID)
}

func (s *Store) queryEntities(query string, args ...interface{}) ([]*Entity, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Entity
	for rows.Next() {
		e := &Entity{}
		var slug, category, parentID, goal, summary, desc, body, status, boundary, date, metadata, createdAt, updatedAt sql.NullString
		err := rows.Scan(
			&e.ID, &e.Type, &e.Title, &slug, &category,
			&parentID, &goal, &summary, &desc,
			&body, &status, &boundary, &date, &metadata,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}
		e.Slug = slug.String
		e.Category = category.String
		e.ParentID = parentID.String
		e.Goal = goal.String
		e.Summary = summary.String
		e.Description = desc.String
		e.Body = body.String
		e.Status = coalesce(status.String, "active")
		e.Boundary = boundary.String
		e.Date = date.String
		e.Metadata = metadata.String
		e.CreatedAt = createdAt.String
		e.UpdatedAt = updatedAt.String
		result = append(result, e)
	}
	return result, nil
}

func nullEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func coalesce(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func getEntityField(e *Entity, field string) string {
	switch field {
	case "title":
		return e.Title
	case "slug":
		return e.Slug
	case "category":
		return e.Category
	case "parent_id":
		return e.ParentID
	case "goal":
		return e.Goal
	case "summary":
		return e.Summary
	case "description":
		return e.Description
	case "body":
		return e.Body
	case "status":
		return e.Status
	case "boundary":
		return e.Boundary
	case "date":
		return e.Date
	case "metadata":
		return e.Metadata
	default:
		return ""
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd cli && go test ./internal/store/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cli/internal/store/entities.go cli/internal/store/entities_test.go cli/internal/store/testhelper_test.go
git commit -m "feat(store): entity CRUD with changelog logging"
```

---

### Task 3: Relationships

**Files:**
- Create: `cli/internal/store/relationships.go`
- Create: `cli/internal/store/relationships_test.go`

- [ ] **Step 1: Write failing relationship tests**

```go
// cli/internal/store/relationships_test.go
package store

import "testing"

func TestAddAndQueryRelationship(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	rels, err := s.RelationshipsFrom("c3-101")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rels) != 1 || rels[0].ToID != "ref-jwt" {
		t.Errorf("got %v, want [{c3-101 ref-jwt uses}]", rels)
	}
}

func TestRelationshipsTo(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	rels, err := s.RelationshipsTo("ref-jwt")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rels) != 1 || rels[0].FromID != "c3-101" {
		t.Errorf("got %v, want [{c3-101 ref-jwt uses}]", rels)
	}
}

func TestRemoveRelationship(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.RemoveRelationship("c3-101", "ref-jwt", "uses")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}

	rels, _ := s.RelationshipsFrom("c3-101")
	if len(rels) != 0 {
		t.Errorf("expected 0 rels after remove, got %d", len(rels))
	}
}

func TestAddRelationship_Idempotent(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Adding same relationship again should not error
	err := s.AddRelationship("c3-101", "ref-jwt", "uses")
	if err != nil {
		t.Fatalf("duplicate add should not error: %v", err)
	}

	rels, _ := s.RelationshipsFrom("c3-101")
	if len(rels) != 1 {
		t.Errorf("expected 1 rel (no dupe), got %d", len(rels))
	}
}

func TestRelationshipsByType(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	rels, err := s.RelationshipsByType("uses")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rels) != 1 {
		t.Errorf("expected 1 uses rel, got %d", len(rels))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./internal/store/ -run TestAdd.*Rel -v`
Expected: FAIL

- [ ] **Step 3: Implement relationships.go**

```go
// cli/internal/store/relationships.go
package store

// Relationship represents a directed edge in the C3 graph.
type Relationship struct {
	FromID  string
	ToID    string
	RelType string
}

// AddRelationship creates a relationship (idempotent).
func (s *Store) AddRelationship(fromID, toID, relType string) error {
	result, err := s.db.Exec(
		"INSERT OR IGNORE INTO relationships (from_id, to_id, rel_type) VALUES (?, ?, ?)",
		fromID, toID, relType,
	)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n > 0 {
		s.logChange(fromID, "modify", "rel:"+relType, "", toID)
	}
	return nil
}

// RemoveRelationship deletes a specific relationship.
func (s *Store) RemoveRelationship(fromID, toID, relType string) error {
	_, err := s.db.Exec(
		"DELETE FROM relationships WHERE from_id = ? AND to_id = ? AND rel_type = ?",
		fromID, toID, relType,
	)
	if err == nil {
		s.logChange(fromID, "modify", "rel:"+relType, toID, "")
	}
	return err
}

// RelationshipsFrom returns all relationships originating from the given entity.
func (s *Store) RelationshipsFrom(id string) ([]Relationship, error) {
	return s.queryRelationships("SELECT from_id, to_id, rel_type FROM relationships WHERE from_id = ?", id)
}

// RelationshipsTo returns all relationships pointing to the given entity.
func (s *Store) RelationshipsTo(id string) ([]Relationship, error) {
	return s.queryRelationships("SELECT from_id, to_id, rel_type FROM relationships WHERE to_id = ?", id)
}

// RelationshipsByType returns all relationships of a given type.
func (s *Store) RelationshipsByType(relType string) ([]Relationship, error) {
	return s.queryRelationships("SELECT from_id, to_id, rel_type FROM relationships WHERE rel_type = ?", relType)
}

func (s *Store) queryRelationships(query string, args ...interface{}) ([]Relationship, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.FromID, &r.ToID, &r.RelType); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./internal/store/ -run TestAdd.*Rel -v && go test ./internal/store/ -run TestRel -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/store/relationships.go cli/internal/store/relationships_test.go
git commit -m "feat(store): relationship CRUD with changelog"
```

---

### Task 4: Changelog

**Files:**
- Create: `cli/internal/store/changelog.go`
- Create: `cli/internal/store/changelog_test.go`

- [ ] **Step 1: Write failing changelog tests**

```go
// cli/internal/store/changelog_test.go
package store

import "testing"

func TestChangelog_LogAndDiff(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s) // seedFixture inserts entities which log "add" actions

	entries, err := s.UnmarkedChanges()
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	// seedFixture inserts 7 entities + 3 relationships = at least 7 add entries
	if len(entries) < 7 {
		t.Errorf("expected >=7 changelog entries, got %d", len(entries))
	}

	// Verify first entry is an add
	found := false
	for _, e := range entries {
		if e.EntityID == "c3-0" && e.Action == "add" {
			found = true
		}
	}
	if !found {
		t.Error("expected changelog entry for c3-0 add")
	}
}

func TestChangelog_Mark(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.MarkChangelog("abc123")
	if err != nil {
		t.Fatalf("mark: %v", err)
	}

	entries, _ := s.UnmarkedChanges()
	if len(entries) != 0 {
		t.Errorf("expected 0 unmarked after mark, got %d", len(entries))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./internal/store/ -run TestChangelog -v`
Expected: FAIL

- [ ] **Step 3: Implement changelog.go**

```go
// cli/internal/store/changelog.go
package store

// ChangeEntry represents a single change in the changelog.
type ChangeEntry struct {
	ID         int
	EntityID   string
	Action     string
	Field      string
	OldValue   string
	NewValue   string
	Timestamp  string
	CommitHash string
}

// logChange appends a changelog entry. Called internally by mutations.
func (s *Store) logChange(entityID, action, field, oldValue, newValue string) {
	s.db.Exec(`INSERT INTO changelog (entity_id, action, field, old_value, new_value)
		VALUES (?, ?, ?, ?, ?)`, entityID, action, nullEmpty(field), nullEmpty(oldValue), nullEmpty(newValue))
}

// UnmarkedChanges returns changelog entries not yet stamped with a commit hash.
func (s *Store) UnmarkedChanges() ([]ChangeEntry, error) {
	rows, err := s.db.Query(`SELECT id, entity_id, action, COALESCE(field,''), COALESCE(old_value,''),
		COALESCE(new_value,''), timestamp, COALESCE(commit_hash,'')
		FROM changelog WHERE commit_hash IS NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ChangeEntry
	for rows.Next() {
		var e ChangeEntry
		if err := rows.Scan(&e.ID, &e.EntityID, &e.Action, &e.Field, &e.OldValue, &e.NewValue, &e.Timestamp, &e.CommitHash); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

// MarkChangelog stamps all unmarked entries with the given commit hash.
func (s *Store) MarkChangelog(commitHash string) error {
	_, err := s.db.Exec("UPDATE changelog SET commit_hash = ? WHERE commit_hash IS NULL", commitHash)
	return err
}
```

- [ ] **Step 4: Also add `logChange` call to `InsertEntity`**

In `entities.go`, at the end of `InsertEntity`, add:
```go
s.logChange(e.ID, "add", "", "", "")
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd cli && go test ./internal/store/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add cli/internal/store/changelog.go cli/internal/store/changelog_test.go cli/internal/store/entities.go
git commit -m "feat(store): changelog with log, diff, mark"
```

---

### Task 5: FTS5 Search

**Files:**
- Create: `cli/internal/store/search.go`
- Create: `cli/internal/store/search_test.go`

- [ ] **Step 1: Write failing search tests**

```go
// cli/internal/store/search_test.go
package store

import "testing"

func TestSearch_BasicMatch(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	results, err := s.Search("authentication")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 'authentication'")
	}
	// ref-jwt and c3-101 should match
	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.ID] = true
	}
	if !ids["ref-jwt"] {
		t.Error("ref-jwt should match 'authentication'")
	}
}

func TestSearch_WithTypeFilter(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	results, err := s.SearchWithFilter("authentication", "ref")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	for _, r := range results {
		if r.Type != "ref" {
			t.Errorf("expected type=ref, got %s for %s", r.Type, r.ID)
		}
	}
}

func TestSearch_NoResults(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	results, _ := s.Search("zzz_nonexistent_term_zzz")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_Snippet(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	results, _ := s.Search("authentication")
	for _, r := range results {
		if r.Snippet == "" {
			t.Errorf("expected snippet for %s", r.ID)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./internal/store/ -run TestSearch -v`
Expected: FAIL

- [ ] **Step 3: Implement search.go**

```go
// cli/internal/store/search.go
package store

// SearchResult holds an FTS5 search result.
type SearchResult struct {
	ID      string
	Type    string
	Title   string
	Snippet string
	Rank    float64
}

// Search performs a full-text search across all entities.
func (s *Store) Search(query string) ([]SearchResult, error) {
	return s.searchFTS(query, "", 20)
}

// SearchWithFilter performs FTS search filtered to a specific entity type.
func (s *Store) SearchWithFilter(query, entityType string) ([]SearchResult, error) {
	return s.searchFTS(query, entityType, 20)
}

// SearchWithLimit performs FTS search with a custom limit.
func (s *Store) SearchWithLimit(query, entityType string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.searchFTS(query, entityType, limit)
}

func (s *Store) searchFTS(query, entityType string, limit int) ([]SearchResult, error) {
	sql := `SELECT e.id, e.type, e.title,
		snippet(entities_fts, 3, '>>>', '<<<', '...', 20),
		rank
		FROM entities_fts
		JOIN entities e ON entities_fts.rowid = e.rowid
		WHERE entities_fts MATCH ?`
	args := []interface{}{query}

	if entityType != "" {
		sql += " AND e.type = ?"
		args = append(args, entityType)
	}
	sql += fmt.Sprintf(" ORDER BY rank LIMIT %d", limit)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Type, &r.Title, &r.Snippet, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./internal/store/ -run TestSearch -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/store/search.go cli/internal/store/search_test.go
git commit -m "feat(store): FTS5 full-text search with type filtering and snippets"
```

---

### Task 6: Graph Traversal

**Files:**
- Create: `cli/internal/store/graph.go`
- Create: `cli/internal/store/graph_test.go`

- [ ] **Step 1: Write failing graph tests**

```go
// cli/internal/store/graph_test.go
package store

import "testing"

func TestChildren(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	kids, err := s.Children("c3-1")
	if err != nil {
		t.Fatalf("children: %v", err)
	}
	if len(kids) != 2 { // c3-101, c3-110
		t.Errorf("expected 2 children, got %d", len(kids))
	}
}

func TestRefsFor(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	refs, err := s.RefsFor("c3-101")
	if err != nil {
		t.Fatalf("refs: %v", err)
	}
	if len(refs) != 1 || refs[0].ID != "ref-jwt" {
		t.Errorf("expected [ref-jwt], got %v", refs)
	}
}

func TestCitedBy(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	citers, err := s.CitedBy("ref-jwt")
	if err != nil {
		t.Fatalf("cited: %v", err)
	}
	if len(citers) != 1 || citers[0].ID != "c3-101" {
		t.Errorf("expected [c3-101], got %v", citers)
	}
}

func TestImpact(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Add more relationships for impact testing
	s.AddRelationship("c3-110", "c3-101", "uses") // users depends on auth

	impact, err := s.Impact("c3-101", 3)
	if err != nil {
		t.Fatalf("impact: %v", err)
	}
	// c3-110 uses c3-101, so changing c3-101 impacts c3-110
	found := false
	for _, r := range impact {
		if r.ID == "c3-110" {
			found = true
		}
	}
	if !found {
		t.Error("c3-110 should be in impact of c3-101")
	}
}

func TestTransitive(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.AddRelationship("c3-110", "c3-101", "uses")

	reachable, err := s.Transitive("c3-101", 3)
	if err != nil {
		t.Fatalf("transitive: %v", err)
	}
	if len(reachable) == 0 {
		t.Error("expected reachable entities from c3-101")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./internal/store/ -run TestChildren -v`
Expected: FAIL

- [ ] **Step 3: Implement graph.go**

```go
// cli/internal/store/graph.go
package store

// ImpactResult represents an entity affected by a change, with traversal depth.
type ImpactResult struct {
	ID    string
	Title string
	Type  string
	Depth int
}

// RefsFor returns ref entities that the given entity cites (via 'uses' relationships).
func (s *Store) RefsFor(entityID string) ([]*Entity, error) {
	cols := prefixColumns("e", entityColumns)
	return s.queryEntities(`SELECT `+cols+` FROM entities e
		JOIN relationships r ON e.id = r.to_id
		WHERE r.from_id = ? AND r.rel_type = 'uses' AND e.type = 'ref'
		ORDER BY e.id`, entityID)
}

// CitedBy returns entities that cite the given ref (via 'uses' relationships).
func (s *Store) CitedBy(refID string) ([]*Entity, error) {
	cols := prefixColumns("e", entityColumns)
	return s.queryEntities(`SELECT `+cols+` FROM entities e
		JOIN relationships r ON e.id = r.from_id
		WHERE r.to_id = ? AND r.rel_type = 'uses'
		ORDER BY e.id`, refID)
}

// prefixColumns adds a table alias prefix to each column in a comma-separated list.
func prefixColumns(alias, columns string) string {
	parts := strings.Split(columns, ",")
	for i, p := range parts {
		parts[i] = alias + "." + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}

// Impact performs transitive impact analysis: who depends on this entity?
// Traverses reverse 'uses' + forward 'affects' up to maxDepth.
func (s *Store) Impact(entityID string, maxDepth int) ([]ImpactResult, error) {
	rows, err := s.db.Query(`
		WITH RECURSIVE impact AS (
			SELECT from_id as entity_id, 1 as depth
			FROM relationships WHERE to_id = ? AND rel_type = 'uses'
			UNION
			SELECT from_id as entity_id, 1 as depth
			FROM relationships WHERE from_id = ? AND rel_type = 'affects'
			UNION
			SELECT r.from_id, i.depth + 1
			FROM relationships r
			JOIN impact i ON r.to_id = i.entity_id
			WHERE r.rel_type = 'uses' AND i.depth < ?
			UNION
			SELECT r.to_id, i.depth + 1
			FROM relationships r
			JOIN impact i ON r.from_id = i.entity_id
			WHERE r.rel_type = 'affects' AND i.depth < ?
		)
		SELECT DISTINCT i.entity_id, e.title, e.type, MIN(i.depth) as depth
		FROM impact i
		JOIN entities e ON i.entity_id = e.id
		WHERE i.entity_id != ?
		GROUP BY i.entity_id
		ORDER BY depth, i.entity_id
	`, entityID, entityID, maxDepth, maxDepth, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ImpactResult
	for rows.Next() {
		var r ImpactResult
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Depth); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Transitive returns all entities reachable from the given ID via any relationship, up to maxDepth.
func (s *Store) Transitive(entityID string, maxDepth int) ([]ImpactResult, error) {
	rows, err := s.db.Query(`
		WITH RECURSIVE reach AS (
			SELECT to_id as entity_id, 1 as depth
			FROM relationships WHERE from_id = ?
			UNION
			SELECT r.to_id, re.depth + 1
			FROM relationships r
			JOIN reach re ON r.from_id = re.entity_id
			WHERE re.depth < ?
		)
		SELECT DISTINCT r.entity_id, e.title, e.type, MIN(r.depth)
		FROM reach r
		JOIN entities e ON r.entity_id = e.id
		WHERE r.entity_id != ?
		GROUP BY r.entity_id
		ORDER BY MIN(r.depth), r.entity_id
	`, entityID, maxDepth, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ImpactResult
	for rows.Next() {
		var r ImpactResult
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Depth); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./internal/store/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/store/graph.go cli/internal/store/graph_test.go
git commit -m "feat(store): graph traversal with impact analysis and transitive reachability"
```

---

### Task 7: Code Map

**Files:**
- Create: `cli/internal/store/codemap.go`
- Create: `cli/internal/store/codemap_test.go`

- [ ] **Step 1: Write failing code map tests**

```go
// cli/internal/store/codemap_test.go
package store

import "testing"

func TestCodeMap_SetAndLookup(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.SetCodeMap("c3-101", []string{"cli/internal/frontmatter/**", "cli/internal/markdown/**"})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	globs, err := s.CodeMapFor("c3-101")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(globs) != 2 {
		t.Errorf("expected 2 globs, got %d", len(globs))
	}
}

func TestCodeMap_LookupByFile(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.SetCodeMap("c3-101", []string{"cli/internal/frontmatter/**"})
	s.SetCodeMap("c3-110", []string{"cli/cmd/users/**"})

	matches, err := s.LookupByFile("cli/internal/frontmatter/parse.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(matches) != 1 || matches[0] != "c3-101" {
		t.Errorf("expected [c3-101], got %v", matches)
	}
}

func TestCodeMap_Excludes(t *testing.T) {
	s := createTestStore(t)

	err := s.AddExclude("**/*_test.go")
	if err != nil {
		t.Fatalf("add exclude: %v", err)
	}

	excludes, _ := s.Excludes()
	if len(excludes) != 1 || excludes[0] != "**/*_test.go" {
		t.Errorf("expected [**/*_test.go], got %v", excludes)
	}
}
```

- [ ] **Step 2: Run tests, verify fail**

Run: `cd cli && go test ./internal/store/ -run TestCodeMap -v`
Expected: FAIL

- [ ] **Step 3: Implement codemap.go**

```go
// cli/internal/store/codemap.go
package store

import (
	"github.com/bmatcuk/doublestar/v4"
)

// SetCodeMap replaces all glob patterns for an entity.
func (s *Store) SetCodeMap(entityID string, globs []string) error {
	s.db.Exec("DELETE FROM code_map WHERE entity_id = ?", entityID)
	for _, g := range globs {
		if g == "" {
			continue
		}
		if _, err := s.db.Exec("INSERT INTO code_map (entity_id, glob) VALUES (?, ?)", entityID, g); err != nil {
			return err
		}
	}
	return nil
}

// CodeMapFor returns glob patterns for a given entity.
func (s *Store) CodeMapFor(entityID string) ([]string, error) {
	rows, err := s.db.Query("SELECT glob FROM code_map WHERE entity_id = ? ORDER BY glob", entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var globs []string
	for rows.Next() {
		var g string
		rows.Scan(&g)
		globs = append(globs, g)
	}
	return globs, nil
}

// LookupByFile returns entity IDs whose code_map globs match the given file path.
func (s *Store) LookupByFile(filePath string) ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT entity_id, glob FROM code_map ORDER BY entity_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []string
	seen := make(map[string]bool)
	for rows.Next() {
		var entityID, glob string
		rows.Scan(&entityID, &glob)
		if seen[entityID] {
			continue
		}
		matched, _ := doublestar.Match(glob, filePath)
		if matched {
			matches = append(matches, entityID)
			seen[entityID] = true
		}
	}
	return matches, nil
}

// AddExclude adds a glob to the exclusion list.
func (s *Store) AddExclude(glob string) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO code_map_excludes (glob) VALUES (?)", glob)
	return err
}

// Excludes returns all exclusion glob patterns.
func (s *Store) Excludes() ([]string, error) {
	rows, err := s.db.Query("SELECT glob FROM code_map_excludes ORDER BY glob")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var globs []string
	for rows.Next() {
		var g string
		rows.Scan(&g)
		globs = append(globs, g)
	}
	return globs, nil
}

// AllCodeMap returns full code map as entity_id -> []glob.
func (s *Store) AllCodeMap() (map[string][]string, error) {
	rows, err := s.db.Query("SELECT entity_id, glob FROM code_map ORDER BY entity_id, glob")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var entityID, glob string
		rows.Scan(&entityID, &glob)
		result[entityID] = append(result[entityID], glob)
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./internal/store/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/store/codemap.go cli/internal/store/codemap_test.go
git commit -m "feat(store): code map with glob matching and exclusions"
```

---

## Chunk 2: Migration + Init + main.go Rewire

### Task 8: Migration command

**Files:**
- Create: `cli/cmd/migrate.go`
- Create: `cli/cmd/migrate_test.go`

- [ ] **Step 1: Write failing migration test**

```go
// cli/cmd/migrate_test.go
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunMigrate(t *testing.T) {
	// Create file-based fixture
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	err := RunMigrate(c3Dir, false, &buf)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// DB should exist
	dbPath := filepath.Join(c3Dir, "c3.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("c3.db should exist after migration")
	}

	// Open and verify
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer s.Close()

	// Should have all entities
	all, _ := s.AllEntities()
	if len(all) < 6 { // context + 2 containers + 2 components + 1 ref + 1 adr = 7
		t.Errorf("expected >=6 entities, got %d", len(all))
	}

	// c3-101 should have uses relationship to ref-jwt
	rels, _ := s.RelationshipsFrom("c3-101")
	foundUses := false
	for _, r := range rels {
		if r.ToID == "ref-jwt" && r.RelType == "uses" {
			foundUses = true
		}
	}
	if !foundUses {
		t.Error("c3-101 should use ref-jwt")
	}

	// ref-jwt should have scope relationship to c3-1
	rels, _ = s.RelationshipsFrom("ref-jwt")
	foundScope := false
	for _, r := range rels {
		if r.ToID == "c3-1" && r.RelType == "scope" {
			foundScope = true
		}
	}
	if !foundScope {
		t.Error("ref-jwt should scope c3-1")
	}

	// Old .md files should still exist (keep-originals=false but files are NOT removed in fixture test)
	// In real migration, files would be removed unless --keep-originals
}

func TestRunMigrate_KeepOriginals(t *testing.T) {
	c3Dir := createFixture(t)
	var buf bytes.Buffer

	err := RunMigrate(c3Dir, true, &buf)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Old files should still exist
	if _, err := os.Stat(filepath.Join(c3Dir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md should still exist with --keep-originals")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./cmd/ -run TestRunMigrate -v`
Expected: FAIL

- [ ] **Step 3: Implement migrate.go**

```go
// cli/cmd/migrate.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// RunMigrate imports .c3/ markdown files into a SQLite database.
func RunMigrate(c3Dir string, keepOriginals bool, w io.Writer) error {
	dbPath := filepath.Join(c3Dir, "c3.db")

	// Refuse if DB already exists
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("database already exists at %s", dbPath)
	}

	// Walk existing files
	fmt.Fprintf(w, "Scanning %s ...\n", c3Dir)
	walkResult, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return fmt.Errorf("walk .c3/: %w", err)
	}
	docs := walkResult.Docs
	fmt.Fprintf(w, "Found %d entities\n", len(docs))

	if len(docs) == 0 {
		return fmt.Errorf("no entities found in %s", c3Dir)
	}

	// Open database
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("create db: %w", err)
	}
	defer s.Close()

	// Import entities
	fmt.Fprintf(w, "Importing entities ")
	for _, doc := range docs {
		fm := doc.Frontmatter
		docType := frontmatter.ClassifyDoc(fm)

		e := &store.Entity{
			ID:          fm.ID,
			Type:        docType.String(),
			Title:       fm.Title,
			Slug:        walker.SlugFromPath(doc.Path),
			Category:    fm.Category,
			ParentID:    fm.Parent,
			Goal:        fm.Goal,
			Summary:     fm.Summary,
			Description: fm.Description,
			Body:        doc.Body,
			Status:      fm.Status,
			Boundary:    fm.Boundary,
			Date:        fm.Date,
		}

		// Serialize Extra fields as metadata JSON if present
		if len(fm.Extra) > 0 {
			if metaJSON, err := json.Marshal(fm.Extra); err == nil {
				e.Metadata = string(metaJSON)
			}
		}

		if err := s.InsertEntity(e); err != nil {
			return fmt.Errorf("insert %s: %w", fm.ID, err)
		}
		fmt.Fprint(w, ".")
	}
	fmt.Fprintln(w, " done")

	// Import relationships
	relCount := 0
	for _, doc := range docs {
		fm := doc.Frontmatter
		for _, ref := range fm.Refs {
			s.AddRelationship(fm.ID, ref, "uses")
			relCount++
		}
		for _, a := range fm.Affects {
			s.AddRelationship(fm.ID, a, "affects")
			relCount++
		}
		for _, sc := range fm.Scope {
			s.AddRelationship(fm.ID, frontmatter.StripAnchor(sc), "scope")
			relCount++
		}
		for _, src := range fm.Sources {
			s.AddRelationship(fm.ID, frontmatter.StripAnchor(src), "sources")
			relCount++
		}
		// Handle 'via' from Extra fields (not a top-level frontmatter field)
		if viaVal, ok := fm.Extra["via"]; ok {
			switch v := viaVal.(type) {
			case string:
				s.AddRelationship(fm.ID, v, "via")
				relCount++
			case []interface{}:
				for _, item := range v {
					if sv, ok := item.(string); ok {
						s.AddRelationship(fm.ID, sv, "via")
						relCount++
					}
				}
			}
		}
	}
	fmt.Fprintf(w, "Imported relationships: %d edges\n", relCount)

	// Import code-map
	projectDir := filepath.Dir(c3Dir)
	cmPath := filepath.Join(c3Dir, "code-map.yaml")
	cm, _ := codemap.ParseCodeMap(cmPath)
	cmCount := 0
	for id, globs := range cm {
		if id == "_exclude" {
			for _, g := range globs {
				s.AddExclude(g)
			}
			continue
		}
		nonEmpty := make([]string, 0)
		for _, g := range globs {
			if g != "" {
				nonEmpty = append(nonEmpty, g)
			}
		}
		if len(nonEmpty) > 0 {
			s.SetCodeMap(id, nonEmpty)
			cmCount += len(nonEmpty)
		}
	}
	_ = projectDir
	fmt.Fprintf(w, "Imported code-map: %d globs\n", cmCount)

	// Remove old files unless --keep-originals
	if !keepOriginals {
		removeOldFiles(c3Dir, dbPath)
		fmt.Fprintln(w, "Removed original .md files")
	}

	fmt.Fprintf(w, "Migration complete. Database: %s\n", dbPath)
	return nil
}

func removeOldFiles(c3Dir, dbPath string) {
	filepath.Walk(c3Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip the database file itself
		if path == dbPath {
			return nil
		}
		// Skip _index directory
		if info.IsDir() && info.Name() == "_index" {
			os.RemoveAll(path)
			return filepath.SkipDir
		}
		// Remove .md files
		if strings.HasSuffix(path, ".md") {
			os.Remove(path)
		}
		// Remove code-map.yaml
		if info.Name() == "code-map.yaml" {
			os.Remove(path)
		}
		return nil
	})

	// Clean up empty directories
	filepath.Walk(c3Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == c3Dir {
			return nil
		}
		entries, _ := os.ReadDir(path)
		if len(entries) == 0 {
			os.Remove(path)
		}
		return nil
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./cmd/ -run TestRunMigrate -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/migrate.go cli/cmd/migrate_test.go
git commit -m "feat(cmd): add c3x migrate command to import .c3/ files into SQLite"
```

---

### Task 9: Init command — create DB instead of files

**Files:**
- Modify: `cli/cmd/init.go`
- Modify: `cli/cmd/init_test.go`

- [ ] **Step 1: Read current init.go and init_test.go**

Read both files to understand the current implementation.

- [ ] **Step 2: Write new init test alongside existing**

Add a test for DB-based init:

```go
func TestRunInit_CreatesDB(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	var buf bytes.Buffer

	err := RunInitDB(c3Dir, "TestProject", &buf)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// DB should exist
	dbPath := filepath.Join(c3Dir, "c3.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("c3.db should exist")
	}

	// Should have context entity
	s, _ := store.Open(dbPath)
	defer s.Close()

	ctx, err := s.GetEntity("c3-0")
	if err != nil {
		t.Fatal("c3-0 should exist")
	}
	if ctx.Title != "TestProject" {
		t.Errorf("title=%q, want TestProject", ctx.Title)
	}

	// Should have adoption ADR
	adrs, _ := s.EntitiesByType("adr")
	if len(adrs) != 1 {
		t.Errorf("expected 1 adr, got %d", len(adrs))
	}
}
```

- [ ] **Step 3: Implement RunInitDB in init.go**

Add `RunInitDB` function that:
1. Creates `.c3/` directory
2. Opens `.c3/c3.db` via `store.Open()`
3. Inserts context entity (c3-0) from template
4. Inserts adoption ADR from template
5. Writes `config.yaml` (minimal)

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./cmd/ -run TestRunInit -v`
Expected: PASS (both old and new tests)

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/init.go cli/cmd/init_test.go
git commit -m "feat(cmd): init creates .c3/c3.db instead of markdown files"
```

---

### Task 10: main.go — store-based dispatch

**Files:**
- Modify: `cli/main.go`
- Modify: `cli/cmd/options.go`

- [ ] **Step 1: Read current main.go fully**

Understand the complete dispatch flow.

- [ ] **Step 2: Add new commands to options.go**

Add `migrate`, `query`, `diff`, `impact`, `export` to the command parser. Add flags: `--fts`, `--vec`, `--mark`, `--keep-originals`, `--query`.

- [ ] **Step 3: Modify main.go dispatch**

Replace the walker-based flow with store-based:

```go
// Before (current):
// c3Dir := config.ResolveC3Dir(...)
// walkResult := walker.WalkC3DocsWithWarnings(c3Dir)
// docs := walkResult.Docs
// graph := walker.BuildGraph(docs)

// After (new):
// c3Dir := config.ResolveC3Dir(...)
// dbPath := filepath.Join(c3Dir, "c3.db")
// s, err := store.Open(dbPath)
// defer s.Close()
// ... dispatch commands with s instead of graph
```

Key changes:
- `init` → calls `RunInitDB` (creates DB)
- `migrate` → calls `RunMigrate`
- All other commands receive `*store.Store` instead of `*walker.C3Graph`
- Remove post-mutation structural index rebuild (DB handles this)
- `capabilities` and `version` and `help` remain unchanged

- [ ] **Step 4: Update command signatures**

Each `Run*` function changes from `(opts Options, graph *walker.C3Graph, ...)` to `(opts Options, s *store.Store, ...)`. This is a large but mechanical change — update signatures one command at a time.

**Important:** Do this incrementally. First add `*store.Store` parameter alongside `*walker.C3Graph`, then migrate command internals, then remove graph parameter.

- [ ] **Step 5: Run all tests**

Run: `cd cli && go test ./... -v`
Expected: Compilation may fail for commands not yet adapted. Fix one at a time.

- [ ] **Step 6: Commit**

```bash
git add cli/main.go cli/cmd/options.go
git commit -m "refactor(main): store-based command dispatch replacing walker"
```

---

## Chunk 3: Adapt Existing Commands

> **PREREQUISITE:** Task 21 (DB test fixture) MUST be completed before any task in this chunk. Tasks 11-16 are independent and can be parallelized.

Each task below adapts one command to use `*store.Store` instead of `*walker.C3Graph`.

### Task 11: list command

**Files:**
- Modify: `cli/cmd/list.go`

- [ ] **Step 1: Read current list.go**

- [ ] **Step 2: Adapt listJSON to use store**

Replace `graph.All()` and `graph.ByType()` with `s.AllEntities()` and `s.EntitiesByType()`. Replace frontmatter field access with `Entity` struct fields. Replace code-map lookup with `s.CodeMapFor()`.

- [ ] **Step 3: Adapt listFlat and listTopology similarly**

- [ ] **Step 4: Update list_test.go to use DB fixture**

- [ ] **Step 5: Run tests**

Run: `cd cli && go test ./cmd/ -run TestList -v`

- [ ] **Step 6: Commit**

```bash
git commit -am "refactor(cmd): adapt list command to store"
```

---

### Task 12: add command

**Files:**
- Modify: `cli/cmd/add.go`
- Modify: `cli/cmd/add_rich.go`

- [ ] **Step 1: Adapt `numbering` package to accept entity slices**

The `numbering` package is coupled to `*walker.C3Graph`. Change signatures:

```go
// Before:
func NextContainerId(graph *walker.C3Graph) int
func NextComponentId(graph *walker.C3Graph, containerNum int, feature bool) (string, error)

// After — accept generic entity slices:
type EntityLike struct {
	ID string
}
func NextContainerId(entities []EntityLike) int
func NextComponentId(entities []EntityLike, containerNum int, feature bool) (string, error)
```

The caller converts `[]*store.Entity` → `[]EntityLike` at the call site. The internal logic (parse numeric suffix from ID) is unchanged.

Update `numbering_test.go` accordingly.

- [ ] **Step 2: Adapt RunAdd**

Replace file creation with `s.InsertEntity()` + `s.AddRelationship()`. ID generation:
- `containers, _ := s.EntitiesByType("container")` → convert to `[]EntityLike` → `numbering.NextContainerId()`
- Same for `NextComponentId`

- [ ] **Step 3: Adapt RunAddRich similarly**

- [ ] **Step 4: Update add_test.go**

- [ ] **Step 5: Run tests**

Run: `cd cli && go test ./cmd/ -run TestRunAdd -v && go test ./internal/numbering/ -v`

- [ ] **Step 6: Commit**

```bash
git commit -am "refactor(cmd): adapt add/add_rich to store, decouple numbering from walker"
```

---

### Task 13: set command

**Files:**
- Modify: `cli/cmd/set.go`

- [ ] **Step 1: Adapt RunSet**

Replace `writer.SetField()` with `s.UpdateEntity(id, fields)`. For section updates (body manipulation), read body from `s.GetEntity()`, apply markdown section change via `markdown.ReplaceSection()`, then `s.UpdateEntity(id, {"body": newBody})`.

- [ ] **Step 2: Update set_test.go**

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestSet -v`

- [ ] **Step 4: Commit**

```bash
git commit -am "refactor(cmd): adapt set command to store"
```

---

### Task 14: wire/unwire commands

**Files:**
- Modify: `cli/cmd/wire.go`

- [ ] **Step 1: Adapt RunWire**

Replace `writer.AddToArrayField()` with `s.AddRelationship(sourceID, targetID, "uses")`. Body table update (Related Refs section) can still use `markdown` package on the entity body, then save via `s.UpdateEntity()`.

- [ ] **Step 2: Adapt RunUnwire**

Replace `writer.RemoveFromArrayField()` with `s.RemoveRelationship()`.

- [ ] **Step 3: Update wire_test.go**

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./cmd/ -run TestWire -v`

- [ ] **Step 5: Commit**

```bash
git commit -am "refactor(cmd): adapt wire/unwire to store"
```

---

### Task 15: delete command

**Files:**
- Modify: `cli/cmd/delete.go`

- [ ] **Step 1: Adapt RunDelete**

Replace file-based cleanup with:
- `s.DeleteEntity(id)` — cascades relationships, code_map, chunks
- For reverse-dependent cleanup: query `s.RelationshipsTo(id)` to find citers, update their bodies
- Container child check: `s.Children(id)` to refuse delete if non-empty

- [ ] **Step 2: Update delete_test.go**

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestDelete -v`

- [ ] **Step 4: Commit**

```bash
git commit -am "refactor(cmd): adapt delete to store"
```

---

### Task 16: lookup, check, graph, coverage, codemap commands

**Files:**
- Modify: `cli/cmd/lookup.go`, `cli/cmd/check_enhanced.go`, `cli/cmd/graph.go`, `cli/cmd/coverage.go`, `cli/cmd/codemap.go`

- [ ] **Step 1: Adapt lookup**

Replace `codemap.Match()` with `s.LookupByFile()`. Enrich results with `s.GetEntity()` and `s.RefsFor()`.

- [ ] **Step 2: Adapt check_enhanced**

Replace graph-based validation with store queries:
- Schema validation: `s.GetEntity()` + `schema.ForType()`
- Broken ref detection: `s.RelationshipsFrom()` + verify target exists via `s.GetEntity()`

- [ ] **Step 3: Adapt graph**

Replace `graph.Transitive()` with `s.Transitive()`. Build mermaid/d2 from store relationships.

- [ ] **Step 4: Adapt coverage**

Replace `codemap.GlobFiles()` with `s.AllCodeMap()` + file system glob.

- [ ] **Step 5: Adapt codemap**

Replace `codemap.ParseCodeMap()` with `s.AllCodeMap()`. Scaffold stubs via `s.SetCodeMap()`.

- [ ] **Step 6: Run all command tests**

Run: `cd cli && go test ./cmd/ -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git commit -am "refactor(cmd): adapt lookup, check, graph, coverage, codemap to store"
```

---

## Chunk 4: New Commands + Skill Layer

### Task 17: query command (FTS5 search)

**Files:**
- Create: `cli/cmd/query.go`
- Create: `cli/cmd/query_test.go`

- [ ] **Step 1: Write failing query test**

```go
// cli/cmd/query_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestRunQuery_FTS(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := Options{
		Command: "query",
		Args:    []string{"authentication"},
		JSON:    true,
	}
	err := RunQuery(opts, s, &buf)
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref-jwt") {
		t.Errorf("should find ref-jwt for 'authentication': %s", output)
	}
}

func TestRunQuery_WithTypeFilter(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := Options{
		Command: "query",
		Args:    []string{"authentication"},
		JSON:    true,
		// TypeFilter: "ref",  // add to Options
	}
	err := RunQuery(opts, s, &buf)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
}
```

- [ ] **Step 2: Implement RunQuery**

```go
// cli/cmd/query.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func RunQuery(opts Options, s *store.Store, w io.Writer) error {
	if len(opts.Args) == 0 {
		return fmt.Errorf("usage: c3x query <search-terms>")
	}
	query := opts.Args[0]

	var results []store.SearchResult
	var err error

	if opts.TypeFilter != "" {
		results, err = s.SearchWithFilter(query, opts.TypeFilter)
	} else {
		results, err = s.Search(query)
	}
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if opts.JSON {
		return writeJSON(w, results)
	}

	if len(results) == 0 {
		fmt.Fprintln(w, "No results.")
		return nil
	}

	for i, r := range results {
		fmt.Fprintf(w, "%d. [%s] %s — %s\n", i+1, r.Type, r.ID, r.Title)
		if r.Snippet != "" {
			fmt.Fprintf(w, "   %s\n", r.Snippet)
		}
	}
	return nil
}
```

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestRunQuery -v`

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/query.go cli/cmd/query_test.go
git commit -m "feat(cmd): add c3x query command with FTS5 search"
```

---

### Task 18: diff command

**Files:**
- Create: `cli/cmd/diff.go`
- Create: `cli/cmd/diff_test.go`

- [ ] **Step 1: Write failing diff test**

```go
func TestRunDiff(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunDiff(s, false, &buf)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}

	output := buf.String()
	// seedDBFixture inserts entities, so there should be add entries
	if !strings.Contains(output, "+ ADDED") {
		t.Errorf("expected add entries in diff: %s", output)
	}
}

func TestRunDiff_Mark(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	RunDiff(s, false, &buf) // render

	err := RunDiffMark(s, "abc123", &buf)
	if err != nil {
		t.Fatalf("mark: %v", err)
	}

	buf.Reset()
	RunDiff(s, false, &buf)
	if strings.Contains(buf.String(), "+ ADDED") {
		t.Error("should have no unmarked entries after mark")
	}
}
```

- [ ] **Step 2: Implement RunDiff and RunDiffMark**

Read `s.UnmarkedChanges()`, group by entity_id, render as:
- `+ ADDED` for action="add"
- `~ MODIFIED` for action="modify" (show field, old→new)
- `- DELETED` for action="delete"

`RunDiffMark` calls `s.MarkChangelog(commitHash)`.

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestRunDiff -v`

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/diff.go cli/cmd/diff_test.go
git commit -m "feat(cmd): add c3x diff command with changelog rendering"
```

---

### Task 19: impact command

**Files:**
- Create: `cli/cmd/impact.go`
- Create: `cli/cmd/impact_test.go`

- [ ] **Step 1: Write failing impact test**

Test that `RunImpact("c3-101", ...)` finds entities that depend on c3-101.

- [ ] **Step 2: Implement RunImpact**

Calls `s.Impact(entityID, depth)`, renders as tree or JSON.

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestRunImpact -v`

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/impact.go cli/cmd/impact_test.go
git commit -m "feat(cmd): add c3x impact command with transitive analysis"
```

---

### Task 20: export command

**Files:**
- Create: `cli/cmd/export.go`
- Create: `cli/cmd/export_test.go`

- [ ] **Step 1: Write failing export test**

Test that `RunExport` creates markdown files from DB entities that match the original format.

- [ ] **Step 2: Implement RunExport**

For each entity:
1. Build YAML frontmatter from entity fields + relationships
2. Combine with body
3. Write to appropriate path (context→README.md, container→c3-N-slug/README.md, etc.)
4. Export code_map to code-map.yaml

This is the inverse of migration — the escape hatch.

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestRunExport -v`

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/export.go cli/cmd/export_test.go
git commit -m "feat(cmd): add c3x export command (DB to markdown)"
```

---

### Task 21: DB test fixture for cmd tests

**Files:**
- Modify: `cli/cmd/testhelper_test.go`

- [ ] **Step 1: Add createDBFixture helper**

```go
func createDBFixture(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("createDBFixture: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	// Same entities as createFixture but inserted via store
	entities := []store.Entity{
		{ID: "c3-0", Type: "context", Title: "TestProject", Goal: "Test the system."},
		{ID: "c3-1", Type: "container", Title: "api", ParentID: "c3-0", Goal: "Serve API requests", Boundary: "service"},
		{ID: "c3-2", Type: "container", Title: "web", ParentID: "c3-0", Boundary: "app"},
		{ID: "c3-101", Type: "component", Title: "auth", ParentID: "c3-1", Category: "foundation", Goal: "Handle authentication."},
		{ID: "c3-110", Type: "component", Title: "users", ParentID: "c3-1", Category: "feature", Goal: "Manage user accounts."},
		{ID: "ref-jwt", Type: "ref", Title: "JWT Authentication", Goal: "Standardize auth tokens"},
		{ID: "adr-20260226-use-go", Type: "adr", Title: "Use Go for CLI", Status: "proposed", Date: "20260226"},
	}
	for _, e := range entities {
		s.InsertEntity(&e)
	}
	s.AddRelationship("c3-101", "ref-jwt", "uses")
	s.AddRelationship("ref-jwt", "c3-1", "scope")
	s.AddRelationship("adr-20260226-use-go", "c3-0", "affects")

	return s
}
```

- [ ] **Step 2: Commit**

```bash
git add cli/cmd/testhelper_test.go
git commit -m "test: add createDBFixture helper for store-based command tests"
```

---

### Task 22: Update help command + options

**Files:**
- Modify: `cli/cmd/help.go`
- Modify: `cli/cmd/options.go`

- [ ] **Step 1: Add new commands to help output**

Add entries for: `migrate`, `query`, `diff`, `impact`, `export`

- [ ] **Step 2: Add new flags to options parser**

Add to `Options` struct:
```go
TypeFilter    string // --type <type> for query filtering
Mark          bool   // --mark for diff
KeepOriginals bool   // --keep-originals for migrate
Limit         int    // --limit <n> for query (default 20)
```

Add to `ParseArgs` switch:
```go
case "--type":
	if i+1 < len(args) { i++; opts.TypeFilter = args[i] }
case "--mark":
	opts.Mark = true
case "--keep-originals":
	opts.KeepOriginals = true
case "--limit":
	if i+1 < len(args) { i++; opts.Limit, _ = strconv.Atoi(args[i]) }
```

Default `Limit` to 20 if not set.

- [ ] **Step 3: Run help_test.go**

Run: `cd cli && go test ./cmd/ -run TestHelp -v`

- [ ] **Step 4: Commit**

```bash
git commit -am "feat(cmd): add new commands to help and options parser"
```

---

### Task 23: Update skill layer

**Files:**
- Modify: `skills/c3/SKILL.md`
- Modify: `skills/c3/references/query.md`
- Modify: `skills/c3/references/audit.md`
- Modify: `skills/c3/references/sweep.md`
- Modify: `skills/c3/references/change.md`

- [ ] **Step 1: Update SKILL.md**

Update the skill to emphasize "all access through c3x commands". Add `c3x query`, `c3x impact`, `c3x diff` to the available commands. Remove references to reading `.c3/` files directly.

- [ ] **Step 2: Update query.md**

Replace "read files + c3x list" with `c3x query "search terms"` as primary navigation. `c3x list --json` for topology.

- [ ] **Step 3: Update sweep.md**

Replace manual graph traversal with `c3x impact <id>`.

- [ ] **Step 4: Update audit.md and change.md**

Reference `c3x check` (unchanged) and `c3x query` for semantic audits. `c3x diff` for reviewing changes before commit.

- [ ] **Step 5: Commit**

```bash
git add skills/c3/
git commit -m "docs: update skill references for DB-backed c3x"
```

---

### Task 24: Full integration test

- [ ] **Step 1: Run all Go tests**

Run: `cd cli && go test ./... -v`
Expected: ALL PASS

- [ ] **Step 2: Build binary**

Run: `bash scripts/build.sh`
Expected: Successful cross-compilation for all 4 targets

- [ ] **Step 3: Manual smoke test**

```bash
# In a temp directory
mkdir /tmp/c3-smoke && cd /tmp/c3-smoke
c3x init --name SmokeTest
c3x add container api
c3x add component auth --container c3-1
c3x add ref error-handling
c3x wire c3-101 cite ref-error-handling
c3x list --json
c3x query "error"
c3x impact c3-101
c3x diff
c3x check
```

- [ ] **Step 4: Test migration on this repo**

```bash
cd /home/lagz0ne/dev/c3-design
# Backup first
cp -r .c3 .c3.backup
c3x migrate
c3x list --json
c3x query "frontmatter"
c3x impact c3-101
```

- [ ] **Step 5: Commit build artifacts**

```bash
git add -A
git commit -m "feat: c3x embedded database — complete migration from file-based to SQLite"
```

---

## Dependency Order

```
Task 1 (store skeleton)
  → Task 2 (entities) → Task 3 (relationships) → Task 4 (changelog)
  → Task 5 (search) → Task 6 (graph) → Task 7 (codemap)
    → Task 8 (migrate) → Task 9 (init)
    → Task 21 (DB fixture) ← MUST be done before Tasks 11-20
    → Task 10 (main.go rewire) + Tasks 11-16 (adapt commands) [atomic boundary]
      → Tasks 17-20 (new commands) [parallelizable]
      → Task 22 (help/options)
      → Task 23 (skill layer)
        → Task 24 (integration test)
```

**Execution notes:**
- Tasks 11-16 (adapting existing commands) are independent and can be parallelized via subagents
- Tasks 17-20 (new commands) are independent and can be parallelized
- Task 21 (DB fixture) MUST be done before Tasks 11-20
- Task 10 and Tasks 11-16 form an atomic boundary — `main.go` dispatch changes break compilation of all commands simultaneously, so all must be adapted together before tests pass
- `c3x index` command (embedding pipeline) is explicitly deferred to a future chunk — requires `@c3x/embed` npm package to exist first
