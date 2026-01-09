---
id: c3-0
c3-version: 3
title: ${PROJECT}
summary: ${SUMMARY}
---

# ${PROJECT}

## Goal

{Why does this system exist? What business problem does it solve?}

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
- Must have: Goal + Containers table
- Prefer: 3-5 focused sections
- This is the entry point - navigable, not exhaustive

REF HYGIENE (context level = system-wide concerns):
- Cite refs that govern cross-container behavior
  (system-wide error strategy, auth patterns, inter-container data flow)
- Container-specific patterns belong in container docs
- If a ref only applies within one container, cite it there instead

Common sections (create whatever serves your Goal):
- Overview (diagram), Actors, Containers, External Systems, Linkages

Delete this comment block after drafting.
-->

## Containers

| ID | Name | Type | Status | Purpose |
|----|------|------|--------|---------|
<!-- Type: service | app | library | external -->
