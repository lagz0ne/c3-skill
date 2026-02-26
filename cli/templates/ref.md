---
id: ref-${SLUG}
c3-version: 4
title: ${TITLE}
goal: ${GOAL}
scope: []
---

# ${TITLE}

## Goal

${GOAL}

<!--
WHAT REFS DOCUMENT:
Refs capture chosen options - "we chose X over Y for Z reason" specific to THIS codebase.
NOT generic technology documentation (link to official docs instead).

THE SEPARATION TEST:
"Is this a technology/pattern choice we made?"
- Yes → Belongs in ref (chosen option with rationale)
- No → Belongs in component (business/domain logic)

KEY SECTIONS (use what serves your Goal):
- Choice: What we chose and the context of that choice
- Why: Why this over the alternatives we considered
- How: Detailed implementation guidance for this codebase
- Not This: Alternatives rejected and why
- Scope: Where it applies and where it doesn't
- Override: How to deviate from this choice

CODE EXAMPLES AS GOLDEN REFERENCES:
- Refs MAY include code examples showing the canonical pattern
- These are prescriptive: "code should look like THIS"
- They serve as review standards, not implementation pointers
- Keep examples minimal — show the pattern, not a full implementation

ANTI-GOALS:
- Don't duplicate official documentation (link instead)
- Don't include business/domain logic (that goes in components)
- Don't document technology generically (document YOUR chosen option)
- Don't create overly broad refs (split by concern if needed)
- Don't add a `## Code References` section (refs have no code counterpart — that section marks a component as implemented)

GUARDRAILS:
- Must have: Goal section
- Prefer: 2-4 focused sections
- Each section must serve the Goal - if not, delete

Delete this comment block after drafting.
-->

## Choice

<!-- What option did we choose? What is the context of this decision? -->

## Why

<!-- Why this over alternatives? What tradeoffs did we accept? -->

## How

<!-- Detailed implementation guidance for THIS codebase. Be concrete. -->

| Guideline | Example |
|-----------|---------|
| ... | ... |

## Not This

<!-- Alternatives we rejected and why -->

| Alternative | Rejected Because |
|-------------|------------------|
| ... | ... |

## Scope

<!-- Where does this choice apply? Be explicit about exclusions. -->

**Applies to:**
- <!-- containers/components where this ref governs behavior -->

**Does NOT apply to:**
- <!-- explicit exclusions -->

## Override

<!-- How to deviate from this choice when justified -->

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

