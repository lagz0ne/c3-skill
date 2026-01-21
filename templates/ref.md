---
id: ref-${SLUG}
title: ${TITLE}
goal: ${GOAL}
---

# ${TITLE}

## Goal

${GOAL}

<!--
WHAT REFS DOCUMENT:
Refs capture "how WE use X HERE" - decisions and conventions specific to THIS codebase.
NOT generic technology documentation (link to official docs instead).

THE SEPARATION TEST: (see references/content-separation.md for full definition)
"Would this content change if we swapped the underlying technology?"
- Yes → Belongs in ref (integration/usage pattern)
- No → Belongs in component (business/domain logic)

KEY SECTIONS (use what serves your Goal):
- When: When this pattern applies
- Why: Why we chose this approach (decision rationale)
- Conventions: Specific rules for THIS codebase
- Where: Which layers/components use this pattern

ANTI-GOALS:
- Don't duplicate official documentation (link instead)
- Don't include business/domain logic (that goes in components)
- Don't document technology generically (document YOUR usage)
- Don't create overly broad refs (split by concern if needed)

GUARDRAILS:
- Must have: Goal section
- Prefer: 2-4 focused sections
- Each section must serve the Goal - if not, delete
- If a section grows large, consider: diagram? split?

Delete this comment block after drafting.
-->

## When

<!-- When does this pattern apply? What triggers its use? -->

## Why

<!-- Why did we choose this approach? What alternatives were considered? -->

## Conventions

<!-- Specific rules for THIS codebase. Be concrete. -->

| Rule | Example |
|------|---------|
| ... | ... |

## Cited By

<!-- Updated when components cite this ref -->
- c3-{N}{NN} ({component name})

## Criteria

Before finalizing, verify:
- [ ] Goal is specific and actionable
- [ ] Captures "how WE use X HERE" not generic documentation
- [ ] Each section directly serves the Goal
- [ ] No business/domain logic included (that belongs in components)
