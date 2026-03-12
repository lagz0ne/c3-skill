---
id: c3-0
c3-version: 4
title: ${PROJECT}
goal: ${GOAL}
summary: ${SUMMARY}
---

# ${PROJECT}

## Goal

${GOAL}

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting -> stale quickly, maintenance burden
- Text walls -> hard to review, hard to maintain
- Isolated content -> can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal + Abstract Constraints + Containers table
- Prefer: 3-5 focused sections
- This is the entry point - navigable, not exhaustive

REF HYGIENE (context level = system-wide concerns):
- Cite refs that govern cross-container behavior
  (system-wide error strategy, auth patterns, inter-container data flow)
- Container-specific patterns belong in container docs
- If a ref only applies within one container, cite it there instead

Common sections (create whatever serves your Goal):
- Overview (diagram), Actors, Abstract Constraints, Containers, External Systems, Linkages

Delete this comment block after drafting.
-->

## Abstract Constraints

<!-- System-level constraints that containers must satisfy -->
<!-- These are non-negotiable requirements that shape how containers allocate responsibilities -->

| Constraint | Rationale | Affected Containers |
|------------|-----------|---------------------|
| | | |

## Containers

<!-- ID = entity ID assigned by c3x add container (e.g., c3-1, c3-2). NOT a name. -->

| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |
|----|------|----------|--------|------------------|-------------------|
| c3-N | | service | active | | |
<!-- Boundary: service | app | library | worker -->
<!-- Responsibilities: What responsibilities this container owns to satisfy abstract constraints -->
<!-- Goal Contribution: How this container advances the system Goal -->
