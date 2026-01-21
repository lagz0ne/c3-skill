# ADR Template

Architecture Decision Records capture decisions at a tactical level. Detailed implementation goes in a separate Plan file.

## Contents

- [File Naming](#file-naming)
- [ADR → Plan Flow](#adr--plan-flow)
- [Template](#template)
- [Status Values](#status-values)
- [Lifecycle](#lifecycle)

## File Naming

```
.c3/adr/adr-YYYYMMDD-{slug}.md        # ADR (tactical)
.c3/adr/adr-YYYYMMDD-{slug}.plan.md   # Plan (detailed)
```

## ADR → Plan Flow

```
ADR (tactical)              Plan (detailed)
├── Problem                 ├── References ADR
├── Decision                ├── Exact file changes
├── Rationale               ├── Step-by-step edits
├── Affected Layers         └── Verification steps
└── Verification
         ↓
    User accepts
         ↓
    Plan generated
```

## Template

```markdown
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
affects: [c3-0, c3-1]
approved-files: []
base-commit:        # Captured when status becomes accepted
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

## Approved Files

Files that may be modified under this ADR (checked by c3-gate):

```yaml
approved-files:
  - src/path/to/file.ts
  - src/path/to/other.ts
```

**Note:** The `approved-files` list in frontmatter gates Edit/Write operations. Only files listed can be changed when this ADR is `status: accepted`.

## Pattern Overrides (if applicable)

**Required when:** The change breaks an established pattern in `.c3/refs/`.

| Ref | Override Reason | Impact |
|-----|-----------------|--------|
| ref-{slug} | [Why this change justifiably breaks the pattern] | [Components/areas affected by divergence] |

**Rules:**
- This section MUST be present if `c3-patterns` analysis returned `breaks`
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
| `accepted` | Ready for implementation | All code editable | Captured (HEAD at acceptance) |
| `implemented` | Changes applied, verified | Gate relaxed | Used for verification |
| `superseded` | Replaced by another ADR | Files NOT editable | N/A |

## Lifecycle

```
proposed → accepted → implemented
              ↓
        Generate Plan
              ↓
        Execute Plan
              ↓
        Run Audit
```

When ADR is accepted:
1. Generate `.plan.md` file (use `superpowers:writing-plans` if available)
2. Execute plan
3. Run audit
4. Mark ADR as `implemented`
