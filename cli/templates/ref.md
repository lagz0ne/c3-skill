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
QUALITY RUBRIC — apply before marking ref complete:

1. COMPLIANCE QUESTION: Can you express the pattern as 1-3 YES/NO questions?
   "Does the code use RS256 for JWT signing?" — GOOD
   "Is the code well-structured?" — BAD (subjective)

2. MECHANISM OVER OUTCOME: Use greppable tokens, not adjectives.
   "Use slog.With() for structured context" — GOOD
   "Use good logging practices" — BAD

3. VIOLATION EXAMPLE: At least 1 concrete anti-example in Not This.

4. SCOPE GROUNDING: WHERE section must be descriptive (what exists today),
   not aspirational (what should exist someday).

5. BREVITY: How section should be 2-15 lines, identifiable in 10 seconds.

6. DEPENDENCY VISIBILITY: Name specific libraries/versions when relevant.

7. SINGLE COMPLIANCE PATH: One way to do it, not "X or Y".
   Split into separate refs if multiple valid approaches exist.

Delete this comment block after drafting.
-->

## Choice

<!-- What option did we choose? What is the context of this decision? -->

## Why

<!-- Why this over alternatives? What tradeoffs did we accept? -->

## How

<!--
Golden pattern — format-flexible. Use whichever communicates the pattern most clearly:
- Code blocks with // REQUIRED / // OPTIONAL annotations
- Paired do/don't examples
- Checklists with concrete criteria
- Tables mapping situation → required action

The test: can a reviewer check compliance in under 10 seconds?
-->

## Not This

<!--
Dual purpose:
1. Rejected alternatives — options we considered and chose against (with reasons)
2. Anti-examples — concrete code/patterns that violate this ref

| What | Why It's Wrong |
|------|---------------|
| ... | ... |
-->

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

