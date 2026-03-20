# migrate — LLM-Assisted Database Migration

## When

User's `.c3/` has markdown files but no `c3.db`. c3x commands will show:
```
error: .c3/ contains markdown files but no database (c3.db)
Use /c3 in Claude Code to run an LLM-assisted migration
```

## Why LLM-Assisted

`.c3/` markdown files can be malformed — broken YAML frontmatter, missing required fields, wrong entity types, dangling references. A raw `c3x migrate` would silently lose content. The LLM reads each broken doc, fixes it, and validates before import.

## Flow

### Phase 1: Assess

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check --json
```

Parse the JSON output. Categorize issues:
- **Parse errors** (severity=error, "failed to parse") → broken YAML, highest priority
- **Broken references** → dangling entity IDs
- **Missing sections** → incomplete docs (warnings, lower priority)

If zero errors → skip to Phase 3.

### Phase 2: Repair (LLM-assisted)

For each error:

1. **Parse errors**: Read the raw file (migration pre-import exception). Fix the YAML frontmatter:
   - Close unclosed quotes, fix indentation, ensure `---` delimiters
   - Ensure `id:` field exists and matches filename convention
   - Ensure `type:` field is correct for the entity kind
   - Preserve all content — don't drop fields, move unknown fields to body

2. **Broken references**:
   - Check if the target was renamed (fuzzy match on title/slug)
   - If match found → update the reference
   - If no match → remove the reference, add a comment in body noting the removal

3. **Missing required sections**:
   - Add empty section stubs (e.g., `## Goal\n\n(to be filled)`)
   - These are warnings — don't block migration

After each fix, write the corrected file back.

IMPORTANT: Use Edit tool to fix `.c3/` markdown files during migration repair ONLY. This is the one exception to the "CLI-only" rule — we're fixing files so the CLI can import them.

### Phase 3: Validate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check --json
```

Verify zero errors. If errors remain, repeat Phase 2.
Warnings are acceptable — they indicate incomplete docs, not broken ones.

### Phase 4: Migrate

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh migrate
```

This imports all docs into `.c3/c3.db` and removes the old files.

### Phase 5: Verify

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh list --json
```

Compare entity count with pre-migration count. Verify no content was lost.

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh query "goal"
```

Test that FTS5 search works on the new database.

### Phase 6: Report

Present to user:
- Entities migrated: N
- Relationships: N
- Errors fixed: N (list each fix)
- Warnings remaining: N
- Status: migration complete

## Anti-patterns

- Do NOT run `c3x migrate` without checking first — malformed docs will lose content
- Do NOT skip the post-migration verification
- Do NOT edit `.c3/c3.db` directly — all access through c3x commands
- Do NOT delete `.c3/` backup files until user confirms migration is correct
