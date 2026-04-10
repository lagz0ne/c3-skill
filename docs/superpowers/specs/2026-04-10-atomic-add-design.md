# Atomic `c3x add` — All-or-None Entity Creation

**Date:** 2026-04-10
**Status:** Draft

## Problem

`c3x add` creates a bare DB record with no body content. The skill references then instruct the LLM to fill content via a separate `c3x write` call. If the LLM gets interrupted, context-switches, or errors between the two calls, a placeholder entity lingers in the DB — incomplete, failing `c3x check`, and polluting the topology.

Since documents are always created by an LLM that already has the content in mind, there's no reason for the two-step dance. Creation should be atomic: all content or nothing.

## Design

### `add` requires body via stdin

`c3x add <type> <slug>` now **requires body content piped via stdin**. No stdin = error.

```bash
cat <<'EOF' | c3x add component auth-service
## Goal
Handles authentication and session management.

## Dependencies
| Target | Why |
|--------|-----|
| c3-101-user-store | Token persistence |
EOF
```

### Atomic transaction

Internally, `add` performs these steps in a single DB transaction:

1. Validate slug format
2. Read stdin, validate non-empty
3. `InsertEntity` (bare record)
4. `content.WriteEntity` (node tree + versioning)
5. `validateBodyContent` against `schema.ForType`
6. If any step fails → **rollback**, nothing persisted
7. On success → **commit**, output result

### Error output

```
Error: c3x add requires body content via stdin
Hint: cat body.md | c3x add component my-slug
Hint: run `c3x schema component` to see required sections

Error: missing required section "Dependencies" for component
Hint: run `c3x schema component` to see required sections
```

### JSON output

`--json` flag (consistent with other commands):

```json
{"id": "c3-301", "type": "component", "slug": "auth-service"}
```

### Flags retained

- `--container <id>` — parent container
- `--feature` — feature component
- `--json` — JSON output

### Flags removed

- `--goal` and `--boundary` from `add_rich` — Goal is extracted from `## Goal` section automatically via existing `syncGoalFromNodes`. No need for separate flags when body is always present.

## Impact on existing commands

### `add_rich` eliminated

`RunAddRich` and its `AddOptions` become dead code. All paths go through the new `RunAdd` which always requires body.

### `write` unchanged

`c3x write` still exists for **updating** existing entities. Its role doesn't change — it replaces body content on an already-created entity.

### `set` unchanged

Still patches individual frontmatter fields or sections.

## Skill reference changes

### `onboard.md`

Replace two-step creation with single atomic call:

```bash
# Before
c3x add container api-gateway
c3x write c3-1 < body.md

# After
cat <<'EOF' | c3x add container api-gateway --json
## Goal
...
## Components
...
## Responsibilities
...
EOF
```

### `change.md`

Same pattern for all entity creation in change phases, including ADRs. The LLM always knows the intent at creation time — even ADRs should have at least `## Goal` from birth.

### `SKILL.md`

Update the command table to reflect that `add` requires stdin body.

## Testing

1. **Happy path**: pipe valid body → entity created with full content, `c3x check` passes
2. **No stdin**: `add` without pipe → error, no entity created
3. **Invalid body** (missing required sections): → error, no entity created (rollback)
4. **Bad slug**: → error before DB interaction
5. **Existing entity**: → conflict error, same as today
6. **All entity types**: component, container, ref, rule, adr, recipe — each with their required sections

## Migration

Breaking change. No transition period. The skill references are the only consumer of `add` in LLM context, and they'll be updated simultaneously. Any external scripts using bare `add` will need to pipe body content.
