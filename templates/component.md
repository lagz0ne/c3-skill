---
id: c3-${N}${NN}
c3-version: 3
title: ${COMPONENT_NAME}
type: component
category: ${CATEGORY}
parent: c3-${N}
goal: ${GOAL}
summary: ${SUMMARY}
---

# ${COMPONENT_NAME}

## Goal

${GOAL}

## Container Connection

<!-- How does this component advance the parent container's Goal? -->
<!-- What would break in the container without this component? -->

## Hand-offs

<!-- What data/control flows IN and OUT of this component -->

| Direction | What | From/To |
|-----------|------|---------|
| IN | | |
| OUT | | |

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting → stale quickly, maintenance burden
- Text walls → hard to review, hard to maintain
- Isolated content → can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal section
- Prefer: 2-4 focused sections
- Each section must serve the Goal - if not, delete
- If a section grows large, consider: diagram? split? ref-*?

REF HYGIENE (component level = component-specific concerns):
- Before writing: does a ref-* already cover this? Cite, don't duplicate.
- Each ref-* cited must directly serve the Goal - no tangential refs.
- If you're duplicating a ref, cite it instead.

Common sections (create whatever serves your Goal):
- Contract (Provides/Expects), Dependencies, Behavior, Edge Cases, Constraints

Delete this comment block after drafting.
-->

## Related Refs

<!-- ref-* documents that govern this component's behavior -->
| Ref | How It Serves Goal |
|-----|-------------------|

## Applicable atterns

<!-- Don't code, show patterns using design tools, code doesn't age well -->

## Criteria

Before finalizing, verify:
- [ ] Goal is specific and actionable
- [ ] Container Connection explains contribution to parent
- [ ] Each ref-* cited directly serves the Goal
- [ ] No duplicated content from refs (cite instead)
