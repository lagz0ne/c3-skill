---
id: rule-${SLUG}
type: rule
c3-version: 4
title: ${TITLE}
goal: ${GOAL}
origin: []
scope: []
---

# ${TITLE}

## Goal

${GOAL}

<!--
WHAT RULES DOCUMENT:
Rules capture enforceable coding standards for THIS codebase.
NOT generic best practices (link to official docs instead).

THE SEPARATION TEST:
"Remove the Why section. Does the doc become useless?"
- Yes → Belongs in ref (the value is in the rationale)
- No → Belongs in rule (the value is in enforcement)
- Neither → Belongs in component (business/domain logic)

KEY SECTIONS (use what serves your Goal):
- Rule: One-line statement of what must be true
- Golden Example: Canonical code showing the correct pattern
- Not This: Anti-patterns with why they're wrong here
- Scope: Where it applies and where it doesn't
- Override: How to deviate from this rule when justified

ANTI-GOALS:
- Don't duplicate official documentation (link instead)
- Don't include business/domain logic (that goes in components)
- Don't create overly broad rules (split by concern if needed)
- Don't add a `## Code References` section (rules have no code counterpart)

GUARDRAILS:
- Must have: Goal, Rule, Golden Example sections
- Prefer: 3-5 focused sections
- Each section must serve the Goal - if not, delete

Delete this comment block after drafting.
-->

## Rule

<!-- One-line statement of what must be true in this codebase -->

## Golden Example

<!-- Canonical code showing the correct pattern. Multiple code blocks OK for multi-context rules. -->

## Not This

<!-- Anti-patterns — what's wrong and what to do instead -->

| Anti-Pattern | Correct | Why Wrong Here |
|-------------|---------|----------------|
| ... | ... | ... |

## Scope

**Applies to:**
- <!-- containers/components where this rule governs behavior -->

**Does NOT apply to:**
- <!-- explicit exclusions -->

## Override

To override this rule:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this rule and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

<!-- Updated when components cite this rule -->
- c3-{N}{NN} ({component name})
