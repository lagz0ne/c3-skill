# ADR Template

Architecture Decision Records capture decisions at a tactical level.

## Contents

- [File Naming](#file-naming)
- [Template](#template)
- [Status Values](#status-values)
- [Lifecycle](#lifecycle)

## File Naming

```
.c3/adr/adr-YYYYMMDD-{slug}.md        # ADR (tactical)
```

## Template

```markdown
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
affects: [c3-0, c3-1]
base-commit:        # Captured when status becomes accepted
superseded-by:      # Link to replacement ADR (when this ADR is superseded)
---

# [Decision Title]

## Status

**Proposed** - YYYY-MM-DD

## Problem

[What triggered this decision. 2-3 sentences max.]

## Decision

[What we decided. Clear and direct.]

## Rationale

[Why this approach. Key tradeoffs considered.]

| Considered | Rejected Because |
|------------|------------------|
| Option A | [reason] |
| Option B | [reason] |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Context | c3-0 | [what changes] |
| Container | c3-1 | [what changes] |
| Component | c3-101 | [what changes] |

## Work Breakdown

Tasks to decompose from this ADR:

| Task | Components | Refs | Acceptance |
|------|-----------|------|------------|
| [task description] | [component IDs] | [ref IDs] | [criteria] |

## Approved Files

Expected file touch set (guidance for task creation, verified by audit):

```yaml
approved-files:
  - src/path/to/file.ts
  - src/path/to/other.ts
```

**Note:** The `approved-files` list defines the expected set of files this ADR will modify. Used for task scoping and audit verification.

## Pattern Overrides (if applicable)

**Required when:** The change breaks an established pattern in `.c3/refs/`.

| Ref | Override Reason | Impact |
|-----|-----------------|--------|
| ref-{slug} | [Why this change justifiably breaks the pattern] | [Components/areas affected by divergence] |

**Rules:**
- This section MUST be present if the change breaks ref alignment
- Each override requires explicit justification (not just "we need to")
- Impact must list what existing code may need updating later

## Verification

- [ ] [Check 1]
- [ ] [Check 2]
- [ ] [Check 3]

## Verification Results (populated on implementation)

```yaml
actual-files: []      # Files actually changed (from git diff)
verification:
  matched: []         # In approved-files AND touched
  unplanned: []       # Touched but not in approved-files
  untouched: []       # In approved-files but not touched
  verified-at:        # ISO timestamp
  verified-by: claude
```
```

## Status Values

| Status | Meaning | Gate Behavior | base-commit |
|--------|---------|---------------|-------------|
| `proposed` | Awaiting review | Files NOT editable | Not set |
| `accepted` | Ready for implementation | approved-files editable | Captured (HEAD at acceptance) |
| `implemented` | Changes applied, verified | Files NOT editable (change complete) | Used for verification |
| `superseded` | Replaced by another ADR | Files NOT editable | N/A |

## Lifecycle

```
proposed → accepted → (execute tasks) → implemented
```

When ADR is accepted:
1. Break down work using the Work Breakdown table
2. Execute tasks
3. Run audit
4. Mark ADR as `implemented`
