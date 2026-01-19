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

## Verification

- [ ] [Check 1]
- [ ] [Check 2]
- [ ] [Check 3]
```

## Status Values

| Status | Meaning | Gate Behavior |
|--------|---------|---------------|
| `proposed` | Awaiting review | Files NOT editable |
| `accepted` | Ready for implementation | Files in `approved-files` editable |
| `implemented` | Changes applied, audit passed | Gate relaxed |
| `superseded` | Replaced by another ADR | Files NOT editable |

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
