# ADR Template

Architecture Decision Records capture decisions at a tactical level. Detailed implementation goes in a separate Plan file.

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

## Verification

- [ ] [Check 1]
- [ ] [Check 2]
- [ ] [Check 3]
```

## Status Values

| Status | Meaning |
|--------|---------|
| `proposed` | Awaiting review |
| `accepted` | Ready for plan + implementation |
| `implemented` | Changes applied, audit passed |
| `superseded` | Replaced by another ADR |

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
