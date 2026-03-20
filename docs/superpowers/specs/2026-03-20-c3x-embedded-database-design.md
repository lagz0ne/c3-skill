# c3x Embedded Database Design

**Date:** 2026-03-20
**Status:** Draft
**Scope:** Replace file-based `.c3/` content with SQLite database managed exclusively by c3x

## Problem

Agents scrawl across `.c3/` markdown files. No search, no ranking, no structured queries. The in-memory graph is rebuilt from a full file walk on every command. c3x should be the single authority for all architectural knowledge — impact maps, relationships, queries.

## Decision

Move from file-based `.c3/` markdown to an embedded SQLite database at `.c3/c3.db`. All interaction goes through c3x. No direct file reading by agents.

## Architecture

```
@c3x/cli (npm package)
├── c3x binary (Go, cross-compiled, pure Go, no CGo)
│   ├── .c3/c3.db (SQLite via modernc.org/sqlite)
│   │   ├── entities table + FTS5 virtual table
│   │   ├── relationships table (graph edges)
│   │   ├── changelog table (mutation log)
│   │   ├── chunks table (chunk metadata)
│   │   ├── code_map + code_map_excludes
│   │   └── store_meta (config, schema version)
│   ├── All commands read/write DB exclusively
│   └── c3x query/index → shells out to @c3x/embed
│
└── @c3x/embed (npm dependency)
    ├── GGUF embedding model (~300MB, cached at ~/.cache/c3x/models/)
    ├── Vector index (sqlite-vec, at ~/.cache/c3x/<project>/vectors.db)
    └── stdio interface: embed, search, sync
```

Diagram: https://diashort.apps.quickable.co/d/47b04e01

## Directory Structure

```
.c3/
├── c3.db            # Single source of truth (committed to git)
├── config.yaml      # Project-level c3x settings (optional)
└── ...              # Future: exports, snapshots, etc.
```

The `.c3/` directory remains, but contains a DB file instead of scattered markdown. This leaves room for future exports, caches, or config alongside the DB.

## Schema (Proven)

Validated against this repo's 22 entities, 45 relationships, 37 code-map entries. DB size: 147KB. All queries below tested and working.

```sql
-- Core entities: components, containers, context, refs, ADRs, recipes
CREATE TABLE entities (
    id          TEXT PRIMARY KEY,   -- c3-0, c3-1, c3-101, ref-auth, adr-20260320-foo
    type        TEXT NOT NULL,      -- context, container, component, ref, adr, recipe
    title       TEXT NOT NULL,
    slug        TEXT,
    category    TEXT,               -- foundation, feature (components only)
    parent_id   TEXT,               -- container ID for components
    goal        TEXT,
    summary     TEXT,
    description TEXT,               -- used by recipes
    body        TEXT,               -- full markdown body
    status      TEXT DEFAULT 'active',
    boundary    TEXT,               -- containers only
    date        TEXT,               -- ISO date (ADRs: YYYYMMDD)
    metadata    TEXT,               -- JSON blob for Extra/inline fields
    created_at  TEXT DEFAULT (datetime('now')),
    updated_at  TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (parent_id) REFERENCES entities(id)
);

-- Graph edges: uses, affects, scope, via, sources
-- Note: parent is denormalized in entities.parent_id (not duplicated here)
CREATE TABLE relationships (
    from_id     TEXT NOT NULL,
    to_id       TEXT NOT NULL,
    rel_type    TEXT NOT NULL,      -- uses, affects, scope, via, sources
    PRIMARY KEY (from_id, to_id, rel_type),
    FOREIGN KEY (from_id) REFERENCES entities(id),
    FOREIGN KEY (to_id) REFERENCES entities(id)
);

CREATE INDEX idx_relationships_to ON relationships(to_id);

-- File-to-entity mappings
CREATE TABLE code_map (
    entity_id   TEXT NOT NULL,
    glob        TEXT NOT NULL,
    PRIMARY KEY (entity_id, glob),
    FOREIGN KEY (entity_id) REFERENCES entities(id)
);

CREATE TABLE code_map_excludes (
    glob TEXT PRIMARY KEY
);

-- Full-text search (external content table, synced via triggers)
CREATE VIRTUAL TABLE entities_fts USING fts5(
    title, goal, summary, body,
    content=entities, content_rowid=rowid,
    tokenize='porter unicode61'
);

-- Chunk metadata (vectors stored externally by @c3x/embed)
CREATE TABLE chunks (
    chunk_id    TEXT PRIMARY KEY,   -- entity_id:seq
    entity_id   TEXT NOT NULL,
    seq         INTEGER NOT NULL,
    content     TEXT NOT NULL,
    model       TEXT,
    embedded_at TEXT,
    FOREIGN KEY (entity_id) REFERENCES entities(id)
);

CREATE INDEX idx_chunks_entity ON chunks(entity_id);

-- Mutation log for git-friendly diffs
CREATE TABLE changelog (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id   TEXT NOT NULL,
    action      TEXT NOT NULL,      -- add, modify, delete
    field       TEXT,               -- NULL for add/delete, field name for modify
    old_value   TEXT,
    new_value   TEXT,
    timestamp   TEXT DEFAULT (datetime('now')),
    commit_hash TEXT                -- set by c3x diff --mark after rendering
);

-- DB metadata
CREATE TABLE store_meta (
    key   TEXT PRIMARY KEY,
    value TEXT
);
-- Initial: schema_version=1, created_at, project_name
```

### FTS5 Sync Triggers

Using external content table (`content=entities`), triggers use rowid-based operations for safe delete/update:

```sql
CREATE TRIGGER entities_fts_insert AFTER INSERT ON entities BEGIN
    INSERT INTO entities_fts(rowid, title, goal, summary, body)
    VALUES (new.rowid, new.title, new.goal, new.summary, new.body);
END;

CREATE TRIGGER entities_fts_update AFTER UPDATE ON entities BEGIN
    INSERT INTO entities_fts(entities_fts, rowid, title, goal, summary, body)
    VALUES ('delete', old.rowid, old.title, old.goal, old.summary, old.body);
    INSERT INTO entities_fts(rowid, title, goal, summary, body)
    VALUES (new.rowid, new.title, new.goal, new.summary, new.body);
END;

CREATE TRIGGER entities_fts_delete AFTER DELETE ON entities BEGIN
    INSERT INTO entities_fts(entities_fts, rowid, title, goal, summary, body)
    VALUES ('delete', old.rowid, old.title, old.goal, old.summary, old.body);
END;
```

## Query Capabilities

### 1. Full-text search (FTS5)

```sql
-- Search across all entity content with BM25 ranking
SELECT e.id, e.title, snippet(entities_fts, 3, '>>>', '<<<', '...', 20)
FROM entities_fts
JOIN entities e ON entities_fts.rowid = e.rowid
WHERE entities_fts MATCH 'frontmatter'
ORDER BY rank;

-- FTS + type filter
SELECT e.id, e.title
FROM entities_fts
JOIN entities e ON entities_fts.rowid = e.rowid
WHERE entities_fts MATCH 'error' AND e.type = 'ref';
```

**Proven:** "frontmatter" returns 5 ranked hits with highlighted snippets across titles, goals, bodies.

### 2. Graph queries

```sql
-- Children of a container
SELECT id, title FROM entities WHERE parent_id = 'c3-1';

-- What does component X depend on?
SELECT r.to_id, e.title FROM relationships r
JOIN entities e ON r.to_id = e.id
WHERE r.from_id = 'c3-111' AND r.rel_type = 'uses';

-- Who depends on component X? (reverse)
SELECT r.from_id, e.title FROM relationships r
JOIN entities e ON r.from_id = e.id
WHERE r.to_id = 'c3-102' AND r.rel_type = 'uses';

-- Transitive impact (recursive CTE, up to depth 3)
WITH RECURSIVE impact AS (
    SELECT from_id as entity_id, 1 as depth
    FROM relationships WHERE to_id = 'c3-101' AND rel_type = 'uses'
    UNION
    SELECT r.from_id, i.depth + 1
    FROM relationships r JOIN impact i ON r.to_id = i.entity_id
    WHERE r.rel_type = 'uses' AND i.depth < 3
)
SELECT i.entity_id, e.title, i.depth
FROM impact i JOIN entities e ON i.entity_id = e.id;
```

**Proven:** walker (c3-102) has 7 reverse dependents. Recursive CTEs traverse the full graph.

### 3. Vector search (via @c3x/embed sidecar)

Vector search is delegated to `@c3x/embed` which owns the vector index. c3x shells out:

```
c3x query --vec "how does authentication work?"
  → @c3x/embed search "how does authentication work?" --limit 10
  ← [{"entity_id": "c3-112", "chunk_seq": 0, "score": 0.87}, ...]
```

c3x then enriches results with entity metadata from its own DB.

### 4. Hybrid search (RRF fusion)

c3x implements Reciprocal Rank Fusion in Go:

1. Run FTS5 query → ranked list A
2. Run vector query → ranked list B
3. For each result: `score = Σ(weight / (k + rank + 1))` where k=60
4. Return merged, re-ranked results

This follows qmd's proven pattern but without the LLM re-ranking step (unnecessary for ~100 entity corpus).

### 5. Structured filters

```sql
-- All feature components in container c3-1 using ref-error-handling
SELECT e.id, e.title FROM entities e
JOIN relationships r ON e.id = r.from_id
WHERE e.parent_id = 'c3-1'
  AND e.category = 'feature'
  AND r.to_id = 'ref-error-handling'
  AND r.rel_type = 'uses';
```

### 6. Orphan detection

```sql
SELECT e.id, e.title, e.type FROM entities e
LEFT JOIN relationships r ON e.id = r.from_id OR e.id = r.to_id
WHERE r.from_id IS NULL;
```

**Proven:** Catches `ref-cross-compiled-binary` (no wiring).

## Commands

### Existing (adapted to DB)

| Command | Change |
|---------|--------|
| `c3x init` | Creates `.c3/c3.db` instead of markdown files |
| `c3x list` | Reads from `entities` table (no file walk) |
| `c3x add` | INSERTs into `entities` + `relationships` + changelog |
| `c3x set` | UPDATEs entity fields + changelog entry |
| `c3x wire` | INSERTs into `relationships` (uses) + changelog |
| `c3x unwire` | DELETEs from `relationships` + changelog |
| `c3x delete` | DELETEs entity + cascading relationship cleanup + changelog |
| `c3x lookup` | Queries `code_map` table (no glob walk) |
| `c3x check` | Validates DB integrity (FK, orphans, schema sections) |
| `c3x graph` | Generates mermaid/d2 from `relationships` table |
| `c3x coverage` | Queries `code_map` completeness |
| `c3x schema` | Unchanged (static section definitions) |
| `c3x codemap` | UPDATEs `code_map` table |
| `c3x capabilities` | Unchanged |
| `c3x version` | Unchanged |
| `c3x help` | Unchanged |

### New commands

| Command | Purpose |
|---------|---------|
| `c3x migrate` | Import `.c3/` markdown files into DB, remove old files |
| `c3x query "..."` | Hybrid search (FTS5 + vector + RRF). Flags: `--fts`, `--vec`, `--type`, `--limit` |
| `c3x diff` | Render changelog since last commit as human-readable text |
| `c3x diff --mark` | Stamp current changelog entries with commit hash (post-commit hook) |
| `c3x index` | Generate/refresh embeddings via `@c3x/embed` |
| `c3x impact <id>` | Transitive impact analysis (recursive CTE on reverse `uses` + forward `affects`) |
| `c3x export` | Dump DB to markdown files (escape hatch, debugging) |

## Diff Mechanism

Every mutation (add, set, wire, unwire, delete) appends to the `changelog` table:

```
| id | entity_id | action | field   | old_value        | new_value         | timestamp           | commit_hash |
|----|-----------|--------|---------|------------------|-------------------|---------------------|-------------|
| 1  | c3-115    | add    | NULL    | NULL             | NULL              | 2026-03-20 14:00:00 | NULL        |
| 2  | c3-102    | modify | summary | "Walk .c3/ tree" | "Walk c3.db"      | 2026-03-20 14:01:00 | NULL        |
| 3  | ref-xbin  | delete | NULL    | NULL             | NULL              | 2026-03-20 14:02:00 | NULL        |
```

`c3x diff` renders uncommitted changes (where `commit_hash IS NULL`):

```
$ c3x diff
Since last commit:

+ ADDED c3-115 [component] rate-limiter
    parent: c3-1 (cli)
    goal: "Rate limit API calls"
    uses: [ref-error-handling]

~ MODIFIED c3-102 [component] walker
    summary: "Walk .c3/ tree" → "Walk c3.db"

- DELETED ref-cross-compiled-binary
```

After committing, `c3x diff --mark` stamps entries with the commit hash so they don't show again.

## Embedding Pipeline

```
c3x index
  ├── Find entities with no chunks (or stale chunks)
  ├── For each entity: chunk body (~900 tokens, markdown-aware breaks)
  ├── Shell out to @c3x/embed:
  │     stdin:  {"texts": ["chunk1", "chunk2", ...]}
  │     stdout: {"vectors": [[0.1, 0.2, ...], [0.3, 0.4, ...]]}
  ├── Store chunks + vectors in DB
  └── Report: "Indexed 22 entities, 47 chunks"
```

**Model management:** `@c3x/embed` (npm package) handles model download to `~/.cache/c3x/models/`. The Go binary never touches GGUF files directly.

**Degradation:** If `@c3x/embed` is unavailable, `c3x query` falls back to FTS5-only. Vector search is additive, not required.

## Migration

Big bang via `c3x migrate`:

1. Scan `.c3/` for all `.md` files
2. Parse YAML frontmatter + body from each (reuses existing `frontmatter.ParseFrontmatter`)
3. Classify entity type (context/container/component/ref/adr/recipe)
4. INSERT into `entities`, `relationships` (uses, affects, scope, via, sources), `code_map`
5. Build FTS5 index (via triggers on insert)
6. Parse `code-map.yaml` → `code_map` rows + `code_map_excludes` rows (skip empty stubs)
7. Remove old `.md` files and `code-map.yaml` (use `--keep-originals` to skip)
8. Write `.c3/c3.db`
9. Trigger `c3x index` to generate embeddings (if `@c3x/embed` available)

No coexistence mode. No dual-path. Clean break.

### Git Merge Strategy

`.c3/c3.db` is a binary file — git cannot merge it. Strategy:

1. **`.gitattributes`:** `*.db binary` (prevents line-ending corruption)
2. **On conflict:** last-writer-wins. The losing side re-runs `c3x migrate` from their branch's export, or applies their changes via `c3x add/set/wire` commands on top of the winner
3. **Escape hatch:** `c3x export` dumps the entire DB to markdown files for manual inspection/diffing
4. **CI integration:** `c3x diff` runs in PR checks to show human-readable changes in PR comments

## Go Dependencies

| Package | Purpose | Notes |
|---------|---------|-------|
| `modernc.org/sqlite` | Pure Go SQLite | No CGo, cross-compiles cleanly |
| (existing) | All current packages | frontmatter parsing reused for migration only |

### Vector Search Architecture

`sqlite-vec` requires CGo, which conflicts with `modernc.org/sqlite` (pure Go) and breaks cross-compilation. Instead, vector storage and search are delegated to `@c3x/embed`:

```
c3x query "semantic question"
  ├── FTS5 search runs in Go (via modernc.org/sqlite) → ranked list A
  ├── Shells out to @c3x/embed for vector search:
  │     stdin:  {"action": "search", "query": "semantic question", "limit": 10}
  │     stdout: {"results": [{"entity_id": "c3-102", "score": 0.87}, ...]}
  ├── RRF fusion in Go merges both lists
  └── Returns combined results
```

`@c3x/embed` owns both the embedding model AND the vector index (stored in its own SQLite DB with sqlite-vec at `~/.cache/c3x/<project>/vectors.db`). The Go binary stores chunk metadata in `chunks` table but never touches vectors directly. This keeps c3x cross-compilable while giving full vector search capability.

**Degradation:** If `@c3x/embed` is unavailable, vector search is skipped. FTS5 + graph queries still work.

Post-migration, `internal/walker/`, `internal/frontmatter/`, `internal/writer/` become migration-only code. New `internal/store/` package owns all DB access.

## Skill Layer Changes

The Claude Code skill (`SKILL.md` + references) changes from "read files, run c3x" to "only run c3x":

- **query operation:** `c3x query "search terms"` replaces file exploration
- **audit operation:** `c3x check` (unchanged) + `c3x query` for semantic audits
- **change operation:** `c3x add/set/wire/delete` (unchanged interface)
- **sweep operation:** `c3x impact <id>` replaces manual graph traversal

The agent never reads `.c3/c3.db` directly. All access through c3x commands.

## Future Path

- `.c3/c3.db` can later move outside the repo (e.g., `~/.cache/c3x/<project>/c3.db`) with a pointer in `.c3/config.yaml`
- MCP server mode (`c3x mcp`) for richer tool integration
- Multi-project federation (query across repos)

## qmd Reference

Key patterns adopted from [tobi/qmd](https://github.com/tobi/qmd):

| Pattern | qmd Implementation | c3x Adaptation |
|---------|-------------------|----------------|
| Embedded DB | SQLite with content-addressable storage | SQLite with entity-based storage |
| Full-text search | FTS5 with porter unicode61, BM25 scoring | Same |
| Vector search | sqlite-vec with cosine distance | Same, delegated to @c3x/embed sidecar (Node.js) |
| Result fusion | Reciprocal Rank Fusion (k=60) with position-aware blending | RRF without LLM re-ranking (corpus too small) |
| Chunking | Smart markdown-aware breaks (~900 tokens, heading/fence scoring) | Same approach |
| Model management | GGUF models auto-downloaded to ~/.cache/qmd/models/ | Via @c3x/embed npm package to ~/.cache/c3x/models/ |
| Graceful degradation | Strong-signal bypass skips LLM expansion | FTS5-only fallback when embeddings unavailable |

Not adopted: LLM query expansion (overkill for ~100 entities), LLM re-ranking (same), MCP server (future).
