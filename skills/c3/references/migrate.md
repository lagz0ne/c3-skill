# migrate — Database Migration

Before migration, the `.c3/` markdown files are the source of truth. Everything in the database is derived from them. Migration transfers that content faithfully — if the database doesn't match the markdown, the database is wrong. After migration, the database is authoritative.

## Detect Migration Type

```bash
ls .c3/c3.db 2>/dev/null && echo "DB_EXISTS" || echo "NO_DB"
find .c3 -name '*.md' -not -path '*/_index/*' 2>/dev/null | head -1 | grep -q . && echo "HAS_MD" || echo "NO_MD"
```

| c3.db | .md files | Path |
|-------|-----------|------|
| NO | YES | **Phase A** — `c3x migrate-legacy` (markdown → SQLite) |
| YES | NO | **Phase B** — `c3x migrate` (body → node tree) |
| YES | YES | Interrupted — delete `c3.db`, then Phase A |
| NO | NO | Route to **onboard** |

Zero-byte `c3.db` = corrupted, delete it. `.md` only in `_index/` = no content, treat as NO_MD.

`c3x check`, `c3x list`, `c3x read` all require a database — unavailable in Phase A until after migration.

## Warnings Are Errors

Any `warning:` line in any c3x output = data loss. Stop, fix, re-run. No exceptions.

| Warning | Meaning |
|---------|---------|
| `warning: skipping X (failed to parse frontmatter)` | Entity lost |
| `warning: skipping X (unknown type)` | Entity lost |
| `warning: relationship X->Y (relType): error` | Relationship dropped |
| `warning: code-map entity X not found in store, skipping` | Code-map entry lost |
| `warning: failed to parse code-map.yaml: error` | Entire code-map lost |
| `WARNING: N entities had stale frontmatter in body` | Content auto-stripped |

## Recovery

| Failure | Recovery |
|---------|----------|
| `migrate-legacy --dry-run` crashes | Fix cause, re-run |
| `migrate-legacy` errors or warns | Delete `c3.db`, fix, restart Phase A |
| `migrate` (v7→v8) errors | Fix, re-run (`migrate` is idempotent — skips done entities) |
| `migrate` (v7→v8) strips content | Restore from B1 export via `c3x write <id>` |

`migrate-legacy` is NOT idempotent — delete `c3.db` to retry. Always use `--keep-originals`.

---

## Phase A: Markdown → SQLite (v6→v7)

### A1: Dry-run

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run
```

All output (warnings + JSON) goes to **stdout**. In agent mode, parse failure warnings appear as text lines BEFORE the JSON blob — the combined output is not valid JSON if warnings exist. Scan the full output for `warning:` lines first, then parse the JSON portion.

Record from **JSON**: `total`, `with_gaps`, `clean`, `entities[]` (IDs, types, gaps), `code_map_issues[]`.
Record `warning:` lines: any `warning: skipping X` — these entities are invisible in the JSON and will be lost.
Record `broken_ref` gap count — these relationships will be dropped.

This is the **before-manifest**.

### A2: Repair

Fix in priority order. Parse failures first — they're invisible to everything else.

| Priority | Gap | Fix |
|----------|-----|-----|
| 1 | Parse failures (text before JSON) | Edit raw `.c3/` file — fix YAML frontmatter. Preserve all content. |
| 2 | `broken_ref` | Update reference if target was renamed, remove if gone. Log removals. |
| 3 | `empty_frontmatter_rich_body` | Extract from body, add to frontmatter. |
| 3 | `missing_section` / `empty_section` | Add stub: `(to be filled)` |
| 3 | `empty_table` | Add row or mark intentionally empty. |
| 4 | `unknown_entity` / `no_file_matches` | Fix entity ID or glob, or remove stale entry. |

After each priority level, re-run:
```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run
```

### A3: Gate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run
```

Pass requires BOTH:
1. Zero `warning:` text lines in the output (these appear before the JSON blob on stdout).
2. JSON: `with_gaps == 0`, `code_map_issues` empty/absent.

Fail → return to A2.

### A4: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --keep-originals
```

Any `warning:` line → delete `c3.db`, fix, restart from A1.
Any error → delete `c3.db`, investigate, restart from A1.

### A5: Verify

The markdown files are the authority. The database must match them exactly.

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh query "goal"
```

| Check | Pass |
|-------|------|
| Entity count: before-manifest `total` vs `list` | Exact match |
| Entity IDs: every before-manifest ID in `list` | All present |
| `c3x read <id>` for every entity | Content matches original markdown |
| `c3x check` | 0 errors, 0 warnings |
| `c3x query "goal"` | Returns results |

### A6: Continue to v7→v8

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate --dry-run
```

Entities to migrate → Phase B. All have nodes → Final Report.

---

## Phase B: Body → Node Tree (v7→v8)

### B1: Snapshot

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh export /tmp/c3-before-v8
```

Record entity count, IDs, and `c3x read <id>` content length per entity. Verify export succeeded: file count in `/tmp/c3-before-v8` should match entity count from `list`.

### B2: Dry-run

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate --dry-run
```

Output format: `  will migrate: <id> (<title>)`, summary `dry-run: N to migrate, N already have nodes (ok)`, then `N entities have no content yet:` with IDs.

If `WARNING: N entities had stale frontmatter` appears — review each flagged entity before proceeding.

### B3: Review flagged entities

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh read <id>
```

`hasStaleFrontmatter()` flags bodies starting with `---\n` or having `key: value` (no spaces in key) in the first 5 lines before a `#` heading or blank line. Flagged content is stripped during migration.

If the content is real (not stale YAML), move it under a heading before migrating:
```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh write <id> <<'BODY'
## Context

<content that would be stripped>

<rest of body>
BODY
```

### B4: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate
```

`FAILED <id>: <error>` → fix via `c3x write <id>`, re-run. `migrate` is idempotent.
`N failed` must be 0.

### B5: Verify

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh export /tmp/c3-after-v8
```

| Check | Pass |
|-------|------|
| Entity count: B1 vs now | Exact match |
| `diff -r /tmp/c3-before-v8 /tmp/c3-after-v8` | No content loss (whitespace changes OK) |
| `c3x read <id>` for flagged entities | Content intact |
| `c3x check` | 0 errors, 0 warnings |
| `c3x query "goal"` | Returns results |

Clean up: `rm -rf /tmp/c3-before-v8 /tmp/c3-after-v8`

---

## Final Report

```
Migration Complete
==================
Phase A (v6→v7): [PASS/FAIL]
  Entities: N before → N after
  Relationships: N
  Repairs: N (list each)

Phase B (v7→v8): [PASS/FAIL or SKIPPED]
  Migrated: N, Already done: N
  Stale frontmatter: N (list each)
  Empty: N (list IDs)

Post-check: 0 errors, 0 warnings
Status: PASS
```
