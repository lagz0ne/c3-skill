# migrate — Cache Rebuild and Legacy Upgrade

In v9, sealed `.c3/` markdown is source of truth. `c3.db` is disposable local cache. Git review and submission should ignore cache artifacts and submit canonical files only.

Use this reference for two cases:
1. **Legacy adoption**: old unsealed markdown needs import via `c3x migrate-legacy`
2. **Version upgrade**: existing cache needs content-node migration via `c3x migrate`

Normal drift or missing-cache cases are not migration work:
- Missing/stale `c3.db` -> `c3x verify`
- Broken seals or merge-conflict cleanup -> `c3x repair`

## Detect Migration Type

```bash
ls .c3/c3.db 2>/dev/null && echo "DB_EXISTS" || echo "NO_DB"
find .c3 -name '*.md' -not -path '*/_index/*' 2>/dev/null | head -1 | grep -q . && echo "HAS_MD" || echo "NO_MD"
```

| c3.db | .md files | Path |
|-------|-----------|------|
| NO | YES | If files are sealed/canonical: run `c3x verify`. If files are legacy/unsealed: **Phase A** — `c3x migrate-legacy` |
| YES | NO | **Phase B** — `c3x migrate` (body → node tree) only for explicit upgrade work |
| YES | YES | Usually v9 steady state. Ignore cache in Git. Run `c3x verify` unless doing explicit migration |
| NO | NO | Route to **onboard** |

Zero-byte `c3.db` = corrupted, delete it, then run `c3x verify`. `.md` only in `_index/` = no content, treat as NO_MD.

In v9, normal commands can rebuild cache from canonical files. Only legacy/unsealed inputs require explicit migration work before standard commands will succeed.

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
| `migrate` reports `BLOCKED: N component(s)` | Follow the printed repair plan exactly: inspect listed IDs, repair canonical content, remove cache files, `c3x import --force`, then rerun `c3x migrate` |
| `migrate` write failure | Fix the database/write error, remove cache files, `c3x import --force`, rerun `c3x migrate`; canonical export is intentionally stopped before partial cache state is submitted |
| `migrate` (v7→v8) strips content | Restore from B1 export via `c3x write <id>` |
| Broken canonical seals block read-only commands | Use repair mutations (`write --section`, `set --section`, `add adr`) or `c3x repair`; read-only commands intentionally stay gated until canonical state verifies |

`migrate-legacy` is NOT idempotent — delete `c3.db` to retry. Always use `--keep-originals`.

---

## Phase A: Markdown → SQLite (v6→v7)

### A1: Dry-run

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run
```

All output goes to **stdout**. In agent mode, structured command output is TOON; parse failure warnings appear as text lines before the TOON manifest. Scan the full output for `warning:` lines first, then read the manifest fields.

Record from the manifest: `total`, `with_gaps`, `clean`, `entities[]` (IDs, types, gaps), `code_map_issues[]`.
Record `warning:` lines: any `warning: skipping X` — these entities are invisible in the manifest and will be lost.
Record `broken_ref` gap count — these relationships will be dropped.

This is the **before-manifest**.

### A2: Repair

Fix in priority order. Parse failures first — they're invisible to everything else.

| Priority | Gap | Fix |
|----------|-----|-----|
| 1 | Parse failures (warning text before manifest) | Edit raw `.c3/` file — fix YAML frontmatter. Preserve all content. |
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
1. Zero `warning:` text lines in the output (these appear before the TOON manifest on stdout).
2. Manifest fields: `with_gaps == 0`, `code_map_issues` empty/absent.

Fail → return to A2.

### A4: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --keep-originals
```

Any `warning:` line → delete `c3.db`, fix, restart from A1.
Any error → delete `c3.db`, investigate, restart from A1.

### A5: Verify

Canonical markdown files are authority. Cache must match them exactly when rebuilt.

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

Expected success summary: `N migrated, M already have nodes (ok), K strict component docs`.

Expected strict-component blocker shape:
```text
0 migrated, N blocked

BLOCKED: N component(s) need repair before migration can finish.
Why: strict component docs are all-or-nothing; C3 made no migration writes.
Fix loop:
  1. Edit the listed component docs in the copied/sandbox tree.
  2. Remove .c3/c3.db* and .c3/.c3.import.tmp.db*.
  3. Run: c3x import --force
  4. Run: c3x migrate
  5. Run: c3x check --include-adr && c3x verify
```

When this appears, do not chain speculative commands. Repair the listed component content, clear disposable cache files, import, migrate, then check/verify. A blocked strict migration is all-or-nothing: no migration writes should occur before blockers are repaired.

Expected write-failure shape:
```text
BLOCKED: migration write failed at <id> after N successful write(s).
Why: C3 stopped before canonical export, so submitted .c3/ markdown is not rewritten from a partial cache.
```

Fix the write/database error, clear disposable cache files, import, migrate, then check/verify. Do not export canonical markdown from a partial cache.

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
