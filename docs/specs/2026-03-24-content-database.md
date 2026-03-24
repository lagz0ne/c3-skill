# Content Database — Element-Level Node Trees

**Date**: 2026-03-24
**Status**: Draft
**Breaking**: Yes (one-off migration, no backward compat)

## Motivation

c3 currently stores document content as a markdown blob in `entities.body`. Every read/write/check/wire/export operation parses this blob on-the-fly via `markdown.ParseSections()`. This works but provides no element-level identity, no content hashing, and no version history.

This spec replaces the body blob with an element-level node tree where every heading, paragraph, list item, table row, and code block has its own identity and content hash — enabling change tracking, external reference by ID, and version history.

## Core Concepts

### Node Tree

Every entity's content is stored as a tree of nodes:

```
Entity c3-101 (API Gateway)
├── heading[2] "Goal"                    → node 1, hash: a7f3...
│   └── paragraph "Validate and route…"  → node 2, hash: b2c4...
├── heading[2] "Dependencies"            → node 3, hash: c5d6...
│   └── table                            → node 4, hash: d7e8...
│       ├── table_header "Direction|…"   → node 5, hash: e9f0...
│       ├── table_row "IN|auth-svc|…"    → node 6, hash: f1a2...
│       └── table_row "OUT|db|…"         → node 7, hash: a3b4...
└── heading[2] "Related Refs"            → node 8, hash: c5d6...
    └── checklist                        → node 9, hash: e7f8...
        ├── checklist_item "[x] ref-…"   → node 10, hash: a9b0...
        └── checklist_item "[ ] ref-…"   → node 11, hash: c1d2...
```

### Hashing

- **Per-node**: `hash = SHA256(content)`. Detects individual node changes.
- **Per-entity**: `root_merkle = SHA256(sorted concat of all node hashes)`. Detects any change in entity content with O(1) comparison.

No merkle cascade through the tree. Subtree change detection is O(children), which is acceptable for 10-50 nodes per entity.

### Versioning

Full content snapshots. On each write, the complete rendered content is stored as a version entry. No delta computation, no replay, no corruption chain.

Retention: keep all by default. Prune on-demand via `DELETE FROM versions WHERE version < ?`.

## Schema

### Modified: `entities` Table

**Dropped columns**: `body`, `description`, `summary`
**Added columns**: `root_merkle`, `version`

```sql
CREATE TABLE entities (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL CHECK(type IN ('system','container','component','ref','adr','rule','recipe')),
    title       TEXT NOT NULL,
    slug        TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT '',
    parent_id   TEXT REFERENCES entities(id) ON DELETE SET NULL,
    goal        TEXT NOT NULL DEFAULT '',     -- denormalized from Goal heading node
    status      TEXT NOT NULL DEFAULT 'active',
    boundary    TEXT NOT NULL DEFAULT '',
    date        TEXT NOT NULL DEFAULT '',
    metadata    TEXT NOT NULL DEFAULT '{}',
    root_merkle TEXT NOT NULL DEFAULT '',     -- SHA256(sorted node hashes)
    version     INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
```

`goal` is denormalized — updated whenever the Goal heading's child content changes. Enables fast listing without node joins.

### New: `nodes` Table

```sql
CREATE TABLE nodes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id   TEXT NOT NULL,
    parent_id   INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    level       INTEGER NOT NULL DEFAULT 0,
    seq         INTEGER NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    hash        TEXT NOT NULL,
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);
CREATE INDEX idx_nodes_entity ON nodes(entity_id);
```

**Node types**: `heading`, `paragraph`, `list`, `ordered_list`, `checklist`, `table`, `code_block`, `blockquote`, `list_item`, `checklist_item`, `table_row`, `table_header`

**ID**: Auto-increment integer. No encoding of entity ID or position. Stable across reorders.

**level**: Heading depth (1-6) for heading nodes. 0 for all others. Nesting depth is implicit in `parent_id` chain.

**seq**: Ordering within parent. Integer with no gap guarantees — reordering may resequence siblings.

**content**: Leaf text. For container nodes (list, table, etc.), empty string. For checklist items, includes checked state: `[x] item` or `[ ] item`.

**hash**: `SHA256(content)`. For container nodes, `SHA256(type)` (distinguishes container types).

### New: `versions` Table

```sql
CREATE TABLE versions (
    entity_id   TEXT NOT NULL,
    version     INTEGER NOT NULL,
    content     TEXT NOT NULL,            -- full rendered markdown at this version
    root_merkle TEXT NOT NULL,
    commit_hash TEXT NOT NULL DEFAULT '',
    created_at  TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (entity_id, version)
);
```

Full snapshots. Typical entity = 1-5KB. 50 entities × 100 versions = 5-25MB. Trivial for SQLite.

### New: `content_fts` (Replaces `entities_fts`)

```sql
CREATE VIRTUAL TABLE content_fts USING fts5(
    content,
    content='nodes',
    content_rowid='rowid'
);

-- Sync triggers
CREATE TRIGGER content_fts_ai AFTER INSERT ON nodes BEGIN
    INSERT INTO content_fts(rowid, content) VALUES (new.rowid, new.content);
END;
CREATE TRIGGER content_fts_ad AFTER DELETE ON nodes BEGIN
    INSERT INTO content_fts(content_fts, rowid, content)
    VALUES ('delete', old.rowid, old.content);
END;
CREATE TRIGGER content_fts_au AFTER UPDATE ON nodes BEGIN
    INSERT INTO content_fts(content_fts, rowid, content)
    VALUES ('delete', old.rowid, old.content);
    INSERT INTO content_fts(rowid, content) VALUES (new.rowid, new.content);
END;
```

Search returns node hits, grouped by entity_id:

```sql
SELECT n.entity_id, e.title, n.id as node_id, n.type,
       snippet(content_fts, 0, '>>>', '<<<', '...', 20) as snippet
FROM content_fts f
JOIN nodes n ON n.rowid = f.rowid
JOIN entities e ON e.id = n.entity_id
WHERE content_fts MATCH ?
ORDER BY rank
```

Entity metadata search (title, goal) uses a trimmed `entities_fts`:

```sql
CREATE VIRTUAL TABLE entities_fts USING fts5(
    title, goal,
    content='entities',
    content_rowid='rowid'
);
```

### Dropped

- `entities.body` — replaced by node tree
- `entities.description` — unused or merged into nodes
- `entities.summary` — unused or merged into nodes
- `chunks` table — dead schema, never used
- Old `entities_fts` — replaced by content_fts + trimmed entities_fts

## Data Flow

### Write Path

```
markdown (stdin) → goldmark AST → walk AST → nodes
                                            ↓
                                  compute hashes
                                            ↓
                                  update root_merkle + version on entity
                                            ↓
                                  render nodes → markdown → INSERT versions
                                            ↓
                                  sync denormalized goal
```

1. Parse markdown input via goldmark (Go markdown AST parser)
2. Walk AST depth-first, creating node rows with parent_id links
3. Compute `SHA256(content)` per node
4. Compute `root_merkle = SHA256(sorted concat of all node hashes)`
5. Render node tree back to markdown string
6. INSERT into versions table
7. If Goal heading content changed, update `entities.goal`

### Read Path

```
SELECT nodes WHERE entity_id = ? ORDER BY parent_id, seq
    ↓
Walk tree → reconstruct output
    ↓
JSON (default) or markdown (--md)
```

1. Load all nodes for entity, ordered for tree reconstruction
2. Build in-memory tree
3. Output as JSON tree (default) or rendered markdown (`--md` flag)

### Search Path

```
content_fts MATCH ? → node hits → GROUP BY entity_id → ranked results
UNION
entities_fts MATCH ? → entity metadata hits → ranked results
```

### Check/Validation Path

Schema validation operates on the node tree directly:
- Required sections = heading nodes with matching content
- Table validation = table_header + table_row nodes under the section
- Column types validated from table_row content (pipe-separated)

## CLI Commands

### Modified Commands

| Command | Change |
|---------|--------|
| `read` | Reconstruct from nodes. Default JSON, `--md` for markdown. `--section` finds heading node + subtree |
| `write` | Parse markdown → goldmark → nodes. Full entity rewrite. |
| `set` | `--field goal` updates entity.goal. `--section X` updates heading node subtree. `--node <id>` updates specific node. |
| `check` | Validate node tree against schema registry (heading names, table structure) |
| `wire` | Find table node under target section, add table_row node |
| `unwire` | Find and remove table_row node |
| `export` | Render all entities from node trees to markdown files |
| `query` | Search content_fts + entities_fts, merge results |
| `diff` | Show changelog + node-level change entries |
| `init` | Parse template markdown → nodes |
| `add` | Parse template → nodes |

### New Commands

| Command | Purpose |
|---------|---------|
| `nodes <entity-id>` | List all nodes as tree with IDs + hashes |
| `node <node-id>` | Get single node content + metadata |
| `hash <entity-id>` | Get root_merkle hash |
| `changed <entity-id> --since <hash>` | List nodes with different hashes since the given root_merkle |
| `versions <entity-id>` | List version history (version, merkle, timestamp, commit) |
| `version <entity-id> <n>` | Get rendered content at version N |
| `prune <entity-id> --keep <n>` | Delete versions older than the last N |

### Unchanged Commands

`list`, `delete`, `lookup`, `impact`, `graph`, `wire`/`unwire` (relationship part), `coverage`, `schema`, `marketplace`

## Migration: `c3x migrate-v2`

One-off, non-reversible migration.

```
1. CREATE TABLE nodes, versions, content_fts (new schema)
2. For each entity:
   a. Parse entity.body via goldmark → AST
   b. Walk AST → INSERT nodes
   c. Compute hashes
   d. Render nodes → content string → INSERT versions (version 1)
   e. Update entity: root_merkle, version = 1
3. Rebuild FTS indexes
4. Recreate entities table without body/description/summary
   (SQLite: CREATE new table → INSERT SELECT → DROP old → ALTER RENAME)
5. DROP chunks table
6. DROP old entities_fts, recreate trimmed version
```

**Point of no return.** Users should `cp .c3/c3.db .c3/c3.db.bak` before running.

## New Dependency

**goldmark** (`github.com/yuin/goldmark`) — Go markdown parser that produces a typed AST. Required for markdown → node tree decomposition and for rendering nodes back to markdown.

## Implementation Order

1. **Store layer**: New tables, node CRUD, hash computation, version snapshots
2. **Markdown parser**: goldmark integration, AST → nodes, nodes → markdown
3. **Write path**: write, set, wire commands through nodes
4. **Read path**: read, export, check from node tree
5. **Search**: New FTS tables and query command
6. **New commands**: nodes, node, hash, changed, versions, version, prune
7. **Migration**: migrate-v2 command
8. **Cleanup**: Drop dead packages (writer, wiring, index), drop chunks
