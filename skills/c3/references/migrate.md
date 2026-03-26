# migrate — LLM-Assisted Database Migration

## When

User's `.c3/` needs migration. Two distinct paths exist — detect which one applies.

## Version Detection Gate

**HARD RULE: Detect the migration type FIRST. Never guess.**

```bash
# Check 1: Does c3.db exist?
ls .c3/c3.db 2>/dev/null && echo "DB_EXISTS" || echo "NO_DB"

# Check 2: Are there markdown files?
find .c3 -name '*.md' -not -path '*/_index/*' 2>/dev/null | head -1 | grep -q . && echo "HAS_MD" || echo "NO_MD"
```

| c3.db | .md files | Migration Type | Action |
|-------|-----------|---------------|--------|
| NO | YES | **v6→v7**: File-based → SQLite | Phase A |
| YES | NO | **v7→v8**: Legacy body → node tree | Phase B |
| YES | YES | **Interrupted v6→v7** | Recovery: delete `c3.db`, then Phase A |
| NO | NO | No `.c3/` content | Route to **onboard** |

**Edge cases**: If `c3.db` exists but is zero bytes, treat as corrupted — delete it and re-detect. If the only `.md` files are inside `_index/`, those are index artifacts — treat as NO_MD.

**CLI command mapping:**
- **Phase A** uses `c3x migrate-legacy` (file→DB). This is the ONLY command that works without a database.
- **Phase B** uses `c3x migrate` (body→nodes). This REQUIRES an existing database.
- `c3x check` REQUIRES a database — NOT available in Phase A until after migration completes.

**If detection is unclear, ask the user.** Don't proceed with assumptions on migration type.

---

## Why LLM-Assisted

Both migration paths have content transformations that can lose data:

- **v6→v7**: YAML frontmatter parsing can fail — `WalkC3DocsWithWarnings()` skips unparseable files with a warning. `addRelSafe()` prints a warning and continues when a relationship target doesn't exist (the relationship is dropped but NOT silently — a `warning:` line appears in output). Unknown entity types are skipped with a warning.
- **v7→v8**: `hasStaleFrontmatter()` detects content that looks like YAML at the start of a body. Specifically: if the body starts with `---\n`, OR if any of the first 5 lines has a `key: value` pattern (no spaces before the colon) before a `#` heading appears. Detected content is auto-stripped during node tree creation. This heuristic can eat real content that happens to look like YAML.

The LLM reads each broken doc, fixes it, validates, and proves content fidelity before and after.

---

## HARD RULE: Warnings Are Errors

Throughout this entire migration process:

- **Every `warning:` or `WARNING:` line in ANY c3x command output MUST be investigated and resolved before proceeding.**
- A `warning: skipping X (failed to parse frontmatter)` = entity will NOT be migrated (data loss).
- A `warning: skipping X (unknown type)` = entity will NOT be migrated (data loss).
- A `warning: relationship X->Y (relType): error` = relationship dropped (printed by `addRelSafe`).
- A `warning: code-map entity X not found in store, skipping` = code-map entry lost.
- A `warning: failed to parse code-map.yaml: error` = entire code-map lost.
- A `WARNING: N entities had stale frontmatter in body` = content will be auto-stripped.

**Do NOT proceed to the next phase while warnings exist.** Fix each one, re-run, confirm zero warnings + zero errors.

---

## Recovery: What To Do When Things Fail

Migration can fail at any point. Here's how to recover:

| Failure | State | Recovery |
|---------|-------|----------|
| `migrate-legacy --dry-run` crashes | No changes made | Fix the crash cause (usually a malformed file), re-run |
| `migrate-legacy` errors mid-run | Partial `c3.db` + original `.md` files intact | Delete `c3.db`, fix the issue, restart Phase A from A1 |
| `migrate-legacy` completes with warnings | `c3.db` exists but has data loss | Delete `c3.db`, fix the issue, restart Phase A from A1 |
| `migrate` (v7→v8) errors mid-run | Some entities have nodes, others don't | Fix the issue, re-run `c3x migrate` — it skips entities that already have nodes |
| `migrate` (v7→v8) completes with warnings | Nodes exist but content was stripped | Use `c3x write <id>` to restore content from B1 export (`/tmp/c3-before-v8`) |
| Post-migration `c3x check` has issues | DB exists, content may be incomplete | Fix via `c3x set` / `c3x write`, re-run `c3x check` |

**Key facts:**
- `c3x migrate-legacy` is NOT idempotent — it errors if `c3.db` exists. Delete the DB to retry.
- `c3x migrate` (v7→v8) IS idempotent — it skips entities that already have nodes. Safe to re-run.
- Phase A ALWAYS uses `--keep-originals` to preserve source files until verification passes.
- Phase B B1 exports to `/tmp/c3-before-v8` — this is the content backup for recovery.

---

## Phase A: File-Based → SQLite (v6→v7)

**Commands available (no database):** `c3x migrate-legacy`, `c3x migrate-legacy --dry-run`
**Commands NOT available until A5 completes:** `c3x check`, `c3x list`, `c3x read`, `c3x query`, `c3x export`

### A1: Snapshot Before State

Capture evidence of what exists before migration. This is the source of truth for post-migration verification.

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run --json 2>&1
```

**Capture BOTH the JSON output AND stderr.** They contain different information. The `2>&1` merges stderr into stdout so both are visible.

**From JSON** (the `MigrateDryRunResult` structure):
- `total` — entity count (files successfully parsed)
- `with_gaps` — entities with quality issues
- `clean` — entities ready as-is
- `total_gaps` — total issues across all entities
- `entities[]` — per-entity detail: ID, file, type, gaps
- `code_map_issues[]` — stale/broken code-map entries (may be absent if none)

**From stderr** (NOT in JSON):
- `warning: skipping X (failed to parse frontmatter)` — files that FAILED to parse entirely. These entities are INVISIBLE in the JSON — they will be LOST.

**Relationship baseline:** The dry-run JSON does NOT include relationship counts — only entity gaps and code-map issues. The actual relationship count is only known after A5 migration runs. For now, record the number of `broken_ref` gaps from the dry-run — these are relationships that WILL be dropped. The A5 output summary (`migrated N entities, N relationships`) is the authoritative relationship count.

Save ALL of this as the **before-manifest**.

### A2: Assess Gaps

From the dry-run output, prioritize fixes:

**Priority 1 — Parse failures** (stderr warnings, not in JSON):
- These entities don't exist to the migration — they will be LOST
- Fix: Edit the raw `.c3/` file (migration pre-import exception — the ONE case where Edit on `.c3/` files is allowed)

**Priority 2 — Broken references** (`broken_ref` gaps in JSON):
- These will cause `addRelSafe()` to print warnings and drop the relationship
- Fix: Update or remove the reference in frontmatter

**Priority 3 — Quality gaps** (other gap types in JSON):
- `empty_frontmatter_rich_body` — goal in body but not frontmatter → extract and add to frontmatter
- `missing_section` — schema-required section absent → add stub with `(to be filled)`
- `empty_section` — section exists but empty → add placeholder content
- `empty_table` — table headers but no data → add at least one row or note as intentionally empty

**Priority 4 — Code-map issues** (in `code_map_issues`):
- `unknown_entity` — code-map references entity that doesn't exist → remove entry or fix ID
- `no_file_matches` — glob matches zero files → update pattern or remove

**Fix in priority order.** Parse failures first — they block everything else.

### A3: Repair (LLM-assisted)

For each issue:

1. **Parse errors**: Read the raw `.c3/` file. Fix YAML frontmatter:
   - Close unclosed quotes, fix indentation, ensure `---` delimiters
   - Ensure `id:` field exists and matches filename convention
   - Ensure `type:` field is correct for the entity kind
   - Preserve ALL content — don't drop fields, move unknown fields to body

2. **Broken references**: Check if target was renamed (fuzzy match on title/slug).
   - If match → update the reference
   - If no match → remove the reference, add a body comment noting the removal
   - **Log every change** — these are evidence for the user

3. **Missing/empty sections**: Add stubs with `(to be filled)`. These prevent post-migration check warnings.

4. **Empty frontmatter with rich body**: Extract from body section, add to frontmatter.

5. **Code-map issues**: Remove stale entries or fix entity IDs/glob patterns.

**After fixing each priority level**, re-run:
```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run --json 2>&1
```

### A4: Validate — Zero Issues Gate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --dry-run --json 2>&1
```

**Two independent checks — BOTH must pass:**

1. **JSON check**: `with_gaps` must be `0`. `code_map_issues` must be empty or absent.
2. **Stderr check**: Zero lines matching `warning:` in the output. Parse failures appear ONLY on stderr, not in JSON — checking JSON alone is NOT sufficient.

If either check fails, return to A3. Do NOT proceed.

Present the validation result to user: entity count by type, total entities, clean count (should equal total).

### A5: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate-legacy --keep-originals 2>&1
```

**ALWAYS use `--keep-originals`** to preserve source `.md` files until A6 verification passes. If verification fails, the originals are needed for recovery (delete `c3.db`, fix, restart). Only delete originals manually after the Final Report shows PASS.

**Capture the full output (stdout + stderr).** Parse for:

- Lines containing `warning:` → **FAILURES**. Every warning is data loss.
- Lines starting with `  migrated` → count successful entities.
- Summary line: `migrated N entities, N relationships -> <absolute-path-to-c3.db>`

**If ANY `warning:` line appears:**
1. Stop immediately.
2. Delete `c3.db`: `rm .c3/c3.db`
3. Fix the issue that caused the warning.
4. Restart from A1.

**If the command itself errors** (non-zero exit, Go stack trace, etc.):
1. Delete `c3.db` if it was created: `rm -f .c3/c3.db`
2. Investigate the error.
3. Restart from A1.

### A6: Verify Content Fidelity

Now the database exists — all c3x commands are available.

```bash
# Post-migration entity list
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list --json

# Post-migration validation
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check --json

# FTS5 search test
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh query "goal"
```

**Evidence checklist — verify ALL, present ALL to user:**

| Check | How | Pass Criteria |
|-------|-----|--------------|
| Entity count | before-manifest `total` vs `list --json` count | Exact match |
| Entity IDs | before-manifest entity IDs vs `list --json` IDs | Every single ID present — enumerate missing ones if any |
| Relationships | A5 summary relationship count. If before-manifest had `broken_ref` gaps that were fixed, all should be imported. If any `warning: relationship` lines appeared, those were dropped. | Zero dropped relationships |
| Content (ALL entities) | `c3x read <id>` for EVERY entity | Body content preserved. Not a sample — check all. |
| Check clean | `c3x check --json` | Zero errors AND zero warnings |
| FTS5 search | `c3x query "goal"` | Returns results |

**If entity count doesn't match**: The missing entities were skipped. This should NOT happen if A4 gate passed — investigate why.

**If check has warnings**: Fix them now with `c3x set` or `c3x write`. Do not leave warnings for later.

### A7: Continue to v7→v8 (if needed)

v6→v7 creates the database but entities may not have node trees yet. Check:

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate --dry-run
```

- If output shows entities to migrate → proceed to **Phase B** (start at B1).
- If all entities already have nodes → skip to **Final Report**.

### Final Report (after ALL migration is complete)

Present to user with evidence:

```
Migration Complete
==================
Phase A (v6→v7): [PASS/FAIL]
  Entities: N before → N after (match: yes/no)
  Relationships: N
  Code-map entries: N
  Repairs made: N (list each)
  Warnings during migration: 0

Phase B (v7→v8): [PASS/FAIL or SKIPPED]
  Entities migrated to nodes: N
  Already had nodes: N
  Stale frontmatter stripped: N (list each)
  Empty entities: N (list IDs)

Post-migration check: 0 errors, 0 warnings
Status: PASS
```

If Phase B was performed, include both phases in one report. Do not produce two separate reports.

---

## Phase B: Legacy Body → Node Tree (v7→v8)

**All c3x commands are available** (database exists).

### B1: Snapshot Before State

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list --json
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh export /tmp/c3-before-v8
```

Record:
- **Total entity count**
- **Entity IDs** (complete list)
- **Content fingerprint**: for each entity, `c3x read <id>` — record content length + first heading

### B2: Dry Run

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate --dry-run
```

Parse the output. The actual format from the CLI:
- `  will migrate: <id> (<title>)` — entities that need body→node conversion
- Summary line: `dry-run: N to migrate`
- `, N already have nodes (ok)` — appended to summary if any
- `N entities have no content yet:` — followed by entity IDs, one per line

**WARNING lines are critical:**
- `WARNING: N entities had stale frontmatter in body (auto-stripped during migration).` — content WILL be removed. Followed by `c3x read <id>` / `c3x write <id>` hints for each flagged entity.

**If the command errors** (non-zero exit): investigate, fix, re-run.

### B3: Review Stale Frontmatter Entities

For each entity flagged with stale frontmatter:

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh read <id>
```

Read the legacy body. The `hasStaleFrontmatter()` heuristic flags content when:
- The body starts with `---\n` (YAML frontmatter delimiter), OR
- Within the first 5 lines, before any `#` heading or blank line, a line matches `key: value` pattern (no spaces in the key portion before the colon)

Identified content will be stripped during node tree creation. **Review each flagged entity manually** to determine if the detected content is:
- **Actually stale YAML** (leftover frontmatter) → safe to let the migration strip it
- **Real content** that happens to look like YAML → rewrite BEFORE migrating

To preserve real content, move it under a proper heading:
```bash
printf '## Context\n\n<content that would be stripped>\n\n<rest of body>' | C3X_MODE=agent bash <skill-dir>/bin/c3x.sh write <id>
```

Note: shell-quote carefully when content contains special characters. Use heredoc if needed:
```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh write <id> <<'BODY'
## Context

<content that would be stripped>

<rest of body>
BODY
```

### B4: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate 2>&1
```

Parse output for:
- `  migrated: <id>` — success
- `  FAILED <id>: <error>` — **BLOCKER**
- Summary: `N migrated`, `, N already have nodes (ok)`, `, N failed`

**If any entity FAILED:**
1. Investigate the error for that entity.
2. Fix via `c3x write <id>` (rewrite the body to something parseable).
3. Re-run `c3x migrate` — it is idempotent (skips entities that already have nodes). Safe to re-run.

**`N failed` in summary MUST be 0.** Re-run until all entities succeed.

### B5: Verify Content Fidelity

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list --json
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check --json
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh export /tmp/c3-after-v8
```

**Evidence checklist — verify ALL:**

| Check | How | Pass Criteria |
|-------|-----|--------------|
| Entity count | B1 count vs `list --json` count | Exact match (v7→v8 doesn't add/remove entities) |
| Content preserved | `diff -r /tmp/c3-before-v8 /tmp/c3-after-v8` | Review every difference. Acceptable: whitespace/formatting normalization. Unacceptable: missing headings, dropped paragraphs, lost table rows. |
| Flagged entities | `c3x read <id>` for each stale-frontmatter entity | Content intact, stale YAML properly handled (stripped if actually stale, preserved if real) |
| Check clean | `c3x check --json` | Zero errors AND zero warnings |
| FTS5 search | `c3x query "goal"` | Returns results |

### B6: Cleanup

Remove temporary export directories:
```bash
rm -rf /tmp/c3-before-v8 /tmp/c3-after-v8
```

Then proceed to **Final Report** (defined at end of Phase A section — same format applies).

---

## Anti-patterns

- Do NOT run `c3x migrate-legacy` if `c3.db` exists — it errors. Delete the DB first to retry.
- Do NOT run `c3x migrate` without a database — use `c3x migrate-legacy` for file→DB.
- Do NOT use `c3x check` before database exists — use `c3x migrate-legacy --dry-run` instead.
- Do NOT run any migration without dry-run first.
- Do NOT skip warnings — every `warning:` line represents data loss or corruption.
- Do NOT check only JSON output at gates — parse failures appear on stderr only.
- Do NOT trust entity counts alone — verify IDs match one-to-one.
- Do NOT verify content on a "sample" — check every entity.
- Do NOT edit `.c3/c3.db` directly — all access through c3x commands.
- Do NOT run `migrate-legacy` without `--keep-originals` — originals are needed for recovery if verification fails.
- Do NOT trust `hasStaleFrontmatter()` — it's a heuristic that checks only the first 5 lines. Review flagged entities manually.
- Do NOT assume a failed `migrate-legacy` left a clean state — delete `c3.db` and restart.
- Do NOT assume `migrate` (v7→v8) failures are unrecoverable — it's idempotent, fix and re-run.
