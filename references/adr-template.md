# ADR Template

Architecture Decision Records capture the journey from problem to solution, documenting the exploration process.

## ADR + Plan: Mutual Reference

Every ADR **must** include an Implementation Plan. They are two sides of the same coin:

```
ADR (medium abstraction)              Plan (low abstraction)
├── Problem/Requirement               ├── ADR Reference ←────────┐
├── Exploration Journey               ├── Code Changes           │
├── Solution                          │   (maps from Changes     │
├── Changes Across Layers ───────────→│    Across Layers)        │
├── Verification ────────────────────→├── Acceptance Criteria    │
└── Implementation Plan ──────────────┴── (maps from Verification)
         │                                         │
         └─────────────────────────────────────────┘
                    Mutual Reference
```

**Key principle:** No ADR without a Plan. No Plan without an ADR. They verify each other.

## File Naming

```
.c3/adr/adr-YYYYMMDD-{slug}.md
```

Example: `.c3/adr/adr-20251127-add-auth-service.md`

## Determining ADR Number

```bash
last_adr=$(find .c3/adr -name "adr-*.md" | sed 's/.*adr-\([0-9-]*\).*/\1/' | sort | tail -1)
today=$(date +%Y%m%d)
```

## Template

```markdown
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
summary: >
  [Why read this - what decision, what it affects]
status: proposed
date: YYYY-MM-DD
---

# [Decision Title]

## Status {#adr-yyyymmdd-status}
**Proposed** - YYYY-MM-DD

## Problem/Requirement {#adr-yyyymmdd-problem}
<!--
Starting point - what user asked for, why change is needed.
-->

[What triggered this decision]

## Exploration Journey {#adr-yyyymmdd-exploration}
<!--
How understanding developed through scoping.
-->

**Initial hypothesis:** [What we first thought]

**Explored:**
- Isolated: [What we found at the element]
- Upstream: [Dependencies discovered]
- Adjacent: [Related elements at same level]
- Downstream: [Consumers/dependents affected]

**Discovered:** [Key insights that shaped the solution]

**Confirmed:** [What Socratic questions validated]

## Solution {#adr-yyyymmdd-solution}
<!--
Formed through exploration above.
-->

[The approach and why it fits]

## Changes Across Layers {#adr-yyyymmdd-changes}
<!--
Specific changes to each affected document.
-->

### Context Level
- [c3-0]: [What changes, why]

### Container Level
- [c3-N-slug]: [What changes, why]

### Component Level
- [c3-NNN-slug]: [What changes, why]

## Verification {#adr-yyyymmdd-verification}
<!--
Checklist derived from scoping - what to inspect when implementing.
Maps to Acceptance Criteria in Implementation Plan.
-->

- [ ] Is [X] at the right abstraction level?
- [ ] Does [Y] upstream dependency still hold?
- [ ] Are [Z] downstream consumers updated?
- [ ] [Specific checks from exploration]

## Implementation Plan {#adr-yyyymmdd-plan}
<!--
MANDATORY. Code-level details for implementing this ADR.
Each item in "Changes Across Layers" must have corresponding code work here.
-->

### Code Changes
<!--
Specific files/functions to modify or create.
Organized to match "Changes Across Layers" above.
-->

| Layer Change | Code Location | Action | Details |
|--------------|---------------|--------|---------|
| [c3-0 change] | `path/to/file` | create/modify/delete | [what specifically] |
| [c3-N change] | `path/to/file` | create/modify/delete | [what specifically] |
| [c3-NNN change] | `path/to/file` | create/modify/delete | [what specifically] |

### Dependencies
<!--
Order of operations. What must happen before what.
-->

```
1. [First thing] - no dependencies
2. [Second thing] - depends on #1
3. [Third thing] - depends on #1, #2
```

### Acceptance Criteria
<!--
Maps from Verification section. How to know code is correct.
Each verification item should have a testable criterion here.
-->

| Verification Item | Acceptance Criterion | How to Test |
|-------------------|---------------------|-------------|
| [X] at right abstraction | [Code pattern to verify] | [Test/command] |
| [Y] dependency holds | [What to check in code] | [Test/command] |
| [Z] consumers updated | [Expected behavior] | [Test/command] |

## Related {#adr-yyyymmdd-related}
- [Links to affected documents]
```

## Status Values

| Status | Meaning |
|--------|---------|
| `proposed` | Decision documented, awaiting review |
| `accepted` | Approved, ready for implementation |
| `implemented` | Changes made to C3 docs and code |
| `superseded` | Replaced by another ADR (link to it) |
| `deprecated` | No longer relevant |

## Key Sections Explained

### Problem/Requirement
The trigger for this decision. Keep it factual - what was asked, not what you interpreted.

### Exploration Journey
**This is the heart of the ADR.** Document how understanding developed:
- What was the initial hypothesis?
- What did exploration reveal in each direction?
- What key discoveries shaped the solution?

### Solution
Emerges naturally from exploration. Should feel inevitable given what was discovered.

### Changes Across Layers
Concrete list of what changes where. Organized by C3 layer for easy review.

### Verification
Checklist for implementation. Derived from exploration - what to check to confirm the change is correct.

### Implementation Plan
**This section is MANDATORY.** It bridges ADR (design) to code (implementation):

- **Code Changes** - Maps each "Changes Across Layers" item to specific code locations. If a doc changes, there's likely code work.
- **Dependencies** - Ordering matters. Some changes must happen before others.
- **Acceptance Criteria** - Maps each Verification item to testable code behavior. How do you know the code is right?

**Mutual verification:**
- Every item in "Changes Across Layers" should have a corresponding Code Change
- Every item in "Verification" should have a corresponding Acceptance Criterion
- Audit will verify this mapping

## Anti-Patterns

| Don't | Do Instead |
|-------|------------|
| Write ADR before exploring | Complete HYPOTHESIZE → EXPLORE → DISCOVER loop first |
| Skip exploration journey | Document the journey - it's the reasoning |
| List changes without why | Each change should have rationale |
| Empty verification section | Derive checks from what exploration revealed |
| Skip Implementation Plan | Plan is MANDATORY - no ADR is complete without it |
| Orphan Code Changes | Every Code Change must trace to a "Changes Across Layers" item |
| Orphan Acceptance Criteria | Every Acceptance Criterion must trace to a Verification item |
| Vague code locations | Be specific: `src/handlers/auth.ts:handleLogin()` not "auth code" |
