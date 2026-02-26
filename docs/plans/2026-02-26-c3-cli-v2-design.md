# C3 CLI v2: Structured Document Engine

**Date:** 2026-02-26
**Status:** Draft

## Problem

The C3 CLI today is a scaffolder + reader: `init`, `list`, `check`, `add`. All content mutations — filling goals, wiring relationships, updating tables — are done by the LLM skill via raw Read/Edit calls. This burns ~40+ tool calls per onboard, ~25+ per change operation, each consuming context tokens on deterministic markdown manipulation.

The skill should decide *what* to write. The CLI should handle *how* to write it.

## Design: Three Layers

Each layer builds on the previous. Ship incrementally.

---

### Layer 1: Controlled Frontmatter + Wiring

**Goal:** CLI owns all structured metadata mutations. No more LLM-edited frontmatter or manual cross-doc table updates.

#### `c3x set <id> <field> <value>`

Update a frontmatter field on any entity:

```bash
c3x set c3-101 goal "Handle JWT authentication"
c3x set c3-101 status provisioned
c3x set adr-20260226-auth status accepted
c3x set ref-auth scope '["c3-1","c3-2"]'    # JSON for array fields
```

Implementation: read file → parse frontmatter → update field → write back (preserving body).

Fields: `goal`, `summary`, `status`, `boundary`, `category`, `scope`, `affects`, `date`, `title`.

#### `c3x wire <source> cite <target>`

Manage ref citation relationships (all three sides updated atomically):

```bash
c3x wire c3-101 cite ref-auth
# → adds ref-auth to c3-101's frontmatter refs[]
# → adds row to c3-101's Related Refs table
# → adds row to ref-auth's Cited By table
```

Unwiring:

```bash
c3x unwire c3-101 cite ref-auth
# → removes ref-auth from c3-101's refs[]
# → removes row from c3-101's Related Refs table
# → removes c3-101 from ref-auth's Cited By table
```

**v1 scope:** Only `cite` relation. Dependencies are informational — the skill fills them via `set section`, not wired bidirectionally. `depend` may come in v2 if a clear invariant emerges.

**Note:** `wire` requires the markdown table parser (`internal/markdown`), so `internal/markdown` is a prerequisite for Layer 1, not just Layer 2.

#### Richer `add`

Accept content at creation time:

```bash
c3x add component auth-provider \
  --container c3-1 \
  --goal "Handle JWT authentication" \
  --summary "Validates tokens, manages sessions"
```

All `--goal`, `--summary`, `--boundary`, `--status` flags map to frontmatter fields filled at write time. Eliminates the scaffold-then-edit-5-times pattern.

#### `c3x set <id> section <name> [--stdin]`

Write to a specific decorated section. Accepts content as argument or via stdin for multiline/JSON:

```bash
# Text section (inline)
c3x set c3-101 section "Container Connection" "Provides auth tokens"

# Table section via stdin (recommended for JSON)
echo '[{"File":"src/auth/jwt.ts","Purpose":"JWT validation"}]' | c3x set c3-101 section "Code References" --stdin

# Append a single row via stdin
echo '{"Direction":"IN","What":"user store","From/To":"c3-102"}' | c3x set c3-101 section "Dependencies" --stdin --append
```

#### Internal changes

- **New: `internal/markdown`** — section parser, table parser/writer (prerequisite for both Layer 1 wire and Layer 2 schema)
- **New: `internal/writer`** — frontmatter serializer (read YAML header → modify → write back, preserving markdown body)
- **Extend: `wiring.go`** — generalize beyond `AddComponentToContainerTable`. Support: Cited By ↔ Related Refs, Container ↔ Components table

---

### Layer 2: Decorated Content (Known Sections + Tables)

**Goal:** CLI knows the schema of each entity type's markdown body. Can parse, write, and validate structured sections.

#### Section registry

Each entity type declares its known sections:

| Entity Type | Section | Content Type | Required |
|------------|---------|-------------|----------|
| component | Goal | text | yes |
| component | Dependencies | table: direction, what, from_to | yes |
| component | Code References | table: file, purpose | yes |
| component | Related Refs | table: ref, how_it_serves_goal | no |
| component | Container Connection | text | no |
| container | Goal | text | yes |
| container | Components | table: id, name, category, status, goal_contribution | yes |
| container | Responsibilities | text | yes |
| container | Complexity Assessment | text | no |
| context | Goal | text | yes |
| context | Abstract Constraints | table: constraint, rationale, affected_containers | yes |
| context | Containers | table: id, name, boundary, status, responsibilities, goal_contribution | yes |
| ref | Goal | text | yes |
| ref | Choice | text | yes |
| ref | Why | text | yes |
| ref | How | text | no |
| ref | Cited By | table: component, usage | no |
| adr | Goal | text | yes |

#### `c3x schema <type> [--json]`

Expose the registry so the skill knows what to provide:

```bash
c3x schema component --json
```
```json
{
  "type": "component",
  "sections": [
    {"name": "Goal", "content_type": "text", "required": true},
    {"name": "Dependencies", "content_type": "table", "columns": ["Direction","What","From/To"], "required": true},
    {"name": "Code References", "content_type": "table", "columns": ["File","Purpose"], "required": true},
    {"name": "Related Refs", "content_type": "table", "columns": ["Ref","How It Serves Goal"], "required": false}
  ]
}
```

#### `c3x set <id> section <name> <content>`

Write to a specific decorated section:

```bash
# Text section
c3x set c3-101 section "Container Connection" "Provides auth tokens to all API endpoints"

# Table section (JSON array of row objects)
c3x set c3-101 section "Code References" '[
  {"file": "src/auth/jwt.ts", "purpose": "JWT validation and signing"},
  {"file": "src/auth/middleware.ts", "purpose": "Express auth middleware"}
]'

# Append a single row
c3x set c3-101 section "Dependencies" --append '{"direction":"IN","what":"user store","from_to":"c3-102"}'
```

#### Enhanced `check`

Validate body content against schema:

```bash
c3x check
# Layer 1: ✗ c3-101: parent c3-99 not found
# Layer 2: ! c3-101: Code References table empty (required)
# Layer 2: ✗ c3-103: Dependencies table malformed (expected 3 columns, got 2)
# Layer 2: ! ref-auth: missing required section "Choice"
```

#### Internal changes

- **New: `internal/markdown`** — section parser (split markdown by `##` headers, parse tables into structured data, reconstruct markdown from structured data)
- **Extend: `walker.go`** — parse body content into structured sections during graph building
- **Extend: `check.go`** — validate sections against registered schemas

---

### Layer 3: Actionable Content Types

**Goal:** Certain cell values in decorated tables have types beyond plain text. The CLI knows what they mean and can verify them.

#### Type system

| Content Type | Where It Appears | What CLI Verifies |
|---|---|---|
| `filepath` | Code References `file` column | File exists on disk, detect renames/deletions |
| `entity_id` | Dependencies `from_to`, Cited By `component` | Entity exists in graph, bidirectional consistency |
| `ref_id` | Related Refs `ref` column | Ref exists, Cited By is consistent |
| `enum` | Dependencies `direction` (IN/OUT) | Value in allowed set |

#### Column type annotations in schema

```json
{
  "name": "Code References",
  "content_type": "table",
  "columns": [
    {"name": "File", "type": "filepath"},
    {"name": "Purpose", "type": "text"}
  ]
}
```

```json
{
  "name": "Dependencies",
  "content_type": "table",
  "columns": [
    {"name": "Direction", "type": "enum", "values": ["IN","OUT"]},
    {"name": "What", "type": "text"},
    {"name": "From/To", "type": "entity_id"}
  ]
}
```

#### Enhanced `check` (Layer 3)

```bash
c3x check
# Layer 3: ✗ c3-101: Code References: "src/auth/old.ts" does not exist
# Layer 3: ✗ c3-101: Dependencies from_to "c3-999" not found in graph
# Layer 3: ! ref-auth: Cited By lists c3-101 but c3-101 Related Refs omits ref-auth
# Layer 3: ✗ c3-103: Dependencies direction "INBOUND" not in [IN, OUT]
```

#### Agent benefit

When `check` reports typed issues, the skill gets structured, actionable findings. No markdown parsing needed — just `c3x set` to fix:

```
check reports: ✗ c3-101 Code References: "src/auth/old.ts" does not exist
skill runs:    c3x set c3-101 section "Code References" '[{"file":"src/auth/jwt.ts","purpose":"JWT validation"}]'
```

The feedback loop tightens: **check → structured issue → set → check passes**.

#### Internal changes

- **Extend: schema registry** — add column type annotations
- **Extend: `check.go`** — type-aware validation (file existence, graph lookup, enum matching)
- **New: bidirectional consistency checker** — verify cite↔cited-by (refs[] ↔ Cited By table ↔ Related Refs table)

---

## Migration: Skill Changes

With CLI v2, the skill reference files simplify:

| Current (LLM does) | Future (CLI does) |
|---|---|
| Read file → find Goal section → Edit | `c3x set <id> goal "..."` |
| Read file → find table → add row → Edit | `c3x set <id> section "Deps" --append '{...}'` |
| Read component → Edit Related Refs → Read ref → Edit Cited By | `c3x wire <id> cite <ref>` |
| Read container → find Components table → add row → Edit | Automatic via `c3x add component` with `--goal` |
| Multiple Edit calls to fill scaffolded template | Single `c3x add ... --goal --summary` |

The skill references would shift from "here's how to edit markdown" to "here's what content decisions to make, then call `c3x set`/`c3x wire`".

---

## Implementation Order

1. **Foundation** — `internal/markdown` (section parser, table parser/writer) — prerequisite for everything
2. **Layer 1** — `internal/writer` (frontmatter write-back), `set` (frontmatter), richer `add`, `wire`/`unwire` (cite only)
3. **Layer 2** — `schema` command, `set <id> section`, enhanced `check` (schema validation)
4. **Layer 3** — column type annotations, typed validation in `check` (filepath, entity_id, enum, bidirectional)

`internal/markdown` ships first because both `wire` (Layer 1) and `schema`/`check` (Layer 2) depend on it.

---

## Non-Goals

- **No interactive TUI** — the CLI is for agent consumption, not humans
- **No `apply` bulk orchestrator** — `set` and `wire` are sufficient; the skill can call them in sequence
- **No declarative target-state diffing** — imperative mutations are simpler and debuggable
- **No content generation** — the CLI never writes prose. That's the LLM's job.
