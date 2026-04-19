# migrate — Cache Rebuild and Legacy Upgrade

v9: sealed `.c3/` markdown = source of truth. `c3.db` = disposable cache. Git should ignore cache artifacts, submit canonical files only.

Two cases:
1. **Legacy adoption**: unsealed markdown needs import via `c3x migrate-legacy`
2. **Version upgrade**: cache needs content-node migration via `c3x migrate`

Not migration work:
- Missing/stale `c3.db` → `c3x verify`
- Broken seals / merge-conflict cleanup → `c3x repair`

## Detect Migration Type

```bash
ls .c3/c3.db 2>/dev/null && echo "DB_EXISTS" || echo "NO_DB"
find .c3 -name '*.md' -not -path '*/_index/*' 2>/dev/null | head -1 | grep -q . && echo "HAS_MD" || echo "NO_MD"
```

| c3.db | .md files | Path |
|-------|-----------|------|
| NO | YES | Sealed/canonical: `c3x verify`. Legacy/unsealed: **Phase A** — `c3x migrate-legacy` |
| YES | NO | **Phase B** — `c3x migrate` (body → node tree), explicit upgrade only |
| YES | YES | v9 steady state. Ignore cache in Git. `c3x verify` unless explicit migration |
| NO | NO | Route to **onboard** |

Zero-byte `c3.db` = corrupted → delete, run `c3x verify`. `.md` only in `_index/` = no content, treat as NO_MD.

Normal commands rebuild cache from canonical files. Only legacy/unsealed inputs need explicit migration.

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
| `migrate-legacy` errors/warns | Delete `c3.db`, fix, restart Phase A |
| `migrate` reports `BLOCKED: N component(s)` | Follow printed repair plan: inspect listed IDs, repair canonical content, remove cache, `c3x import --force`, rerun `c3x migrate` |
| `migrate` write failure | Fix database/write error, remove cache, `c3x import --force`, rerun `c3x migrate`; canonical export stopped before partial cache submitted |
| `migrate` (v7→v8) strips content | Restore from B1 export via `c3x write <id>` |
| Broken canonical seals block read-only commands | Use repair mutations (`write --section`, `set --section`, `add adr`) or `c3x repair`; read-only commands stay gated until canonical state verifies |

`migrate-legacy` NOT idempotent — delete `c3.db` to retry. Always use `--keep-originals`.

---

## Phase A: Markdown → SQLite (v6→v7)

### A1: Dry-run

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run
```

All output → stdout. Agent mode: structured output is TOON; parse failure warnings appear as text lines before TOON manifest. Scan full output for `warning:` lines first, then read manifest fields.

Record from manifest: `total`, `with_gaps`, `clean`, `entities[]` (IDs, types, gaps), `code_map_issues[]`.
Record `warning:` lines: any `warning: skipping X` — entities invisible in manifest, will be lost.
Record `broken_ref` gap count — relationships will be dropped.

= **before-manifest**.

### A2: Repair

Fix in priority order. Parse failures first — invisible to everything else.

| Priority | Gap | Fix |
|----------|-----|-----|
| 1 | Parse failures (warning text before manifest) | Edit raw `.c3/` file — fix YAML frontmatter. Preserve all content. |
| 2 | `broken_ref` | Update reference if target renamed, remove if gone. Log removals. |
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
1. Zero `warning:` text lines in output (before TOON manifest on stdout).
2. Manifest: `with_gaps == 0`, `code_map_issues` empty/absent.

Fail → return to A2.

### A4: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --keep-originals
```

Any `warning:` → delete `c3.db`, fix, restart from A1.
Any error → delete `c3.db`, investigate, restart from A1.

### A5: Verify

Canonical markdown = authority. Cache must match exactly when rebuilt.

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

Record entity count, IDs, `c3x read <id>` content length per entity. Verify export: file count in `/tmp/c3-before-v8` should match entity count from `list`.

### B2: Dry-run

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate --dry-run
```

Output: `  will migrate: <id> (<title>)`, summary `dry-run: N to migrate, N already have nodes (ok)`, then `N entities have no content yet:` with IDs.

If `WARNING: N entities had stale frontmatter` → review each flagged entity before proceeding.

### B3: Review flagged entities

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh read <id>
```

`hasStaleFrontmatter()` flags bodies starting with `---\n` or having `key: value` (no spaces in key) in first 5 lines before `#` heading or blank line. Flagged content stripped during migration.

If content is real (not stale YAML), move under heading before migrating:
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

Expected success: `N migrated, M already have nodes (ok), K strict component docs`.

Expected strict-component blocker:
```text
0 migrated, N blocked

BLOCKED: N component(s) need repair before migration can finish.
Why: strict component docs are all-or-nothing; C3 made no migration writes.
writesMade: false
Fix loop:
  1. Inspect the blockers below.
  2. Run scoped repairs: c3x migrate repair <id> --section <section> < content.md
  3. Run: c3x cache clear
  4. Run: c3x import --force
  5. Run: c3x migrate --continue
  6. Run: c3x check --include-adr && c3x verify
```

Machine-readable inspection:

```bash
c3x migrate --dry-run --json
c3x migrate repair-plan
```

On blocker: no speculative commands. Repair listed component sections, `c3x cache clear`, import, continue, check/verify. Blocked strict migration = all-or-nothing: no writes before blockers repaired.

Expected write-failure:
```text
BLOCKED: migration write failed at <id> after N successful write(s).
Why: C3 stopped before canonical export, so submitted .c3/ markdown is not rewritten from a partial cache.
```

Fix write/database error, `c3x cache clear`, import, continue, check/verify. Never export canonical markdown from partial cache.

### B5: Verify

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh export /tmp/c3-after-v8
```

| Check | Pass |
|-------|------|
| Entity count: B1 vs now | Exact match |
| `diff -r /tmp/c3-before-v8 /tmp/c3-after-v8` | No content loss (whitespace OK) |
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
