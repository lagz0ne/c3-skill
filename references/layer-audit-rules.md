# Layer Audit Rules

Quick reference for auditing C3 documentation layer compliance.

## Context Layer (c3-0) - Bird's-eye view

**MUST INCLUDE:**
- Container responsibilities (WHY each exists)
- Container relationships (how they connect)
- Connecting points (APIs, protocols, events)
- External actors (who/what interacts)
- System boundary (inside vs outside)

**MUST EXCLUDE:**
- Component lists (push to Container)
- How containers work internally (push to Container)
- Implementation details (push to Component)

**LITMUS:** "Is this about WHY containers exist and HOW they relate?"

**DIAGRAMS:** System Context, Container Overview
**AVOID:** Sequence with methods, class diagrams, flowcharts with logic

---

## Container Layer (c3-N) - Inside view

**MUST INCLUDE:**
- Component responsibilities (WHAT each does)
- Component relationships (how they interact)
- Data flows (how data moves)
- Business flows (workflows spanning components)
- Inner patterns (logging, config, errors)

**MUST EXCLUDE:**
- WHY this container exists (push to Context)
- Container-to-container details (push to Context)
- HOW components work (push to Component)

**LITMUS:** "Is this about WHAT components do and HOW they relate?"

**DIAGRAMS:** Component Relationships, Data Flow
**AVOID:** System context, actor diagrams

---

## Component Layer (c3-NNN) - Close-up view

**MUST INCLUDE:**
- Flows (step-by-step processing)
- Dependencies (what it calls)
- Decision logic (branching points)
- Edge cases (non-obvious scenarios)
- Error handling (what can go wrong)

**MUST EXCLUDE:**
- WHAT this component does (already in Container)
- Component relationships (push to Container)
- Container relationships (push to Context)

**LITMUS:** "Is this about HOW this component implements its contract?"

**DIAGRAMS:** Flowcharts, Sequence (to dependencies), State charts
**AVOID:** System context, container overview
