---
id: adr-00000000-c3-adoption
type: adr
title: Adopt C3 Architecture Documentation
status: implemented
date: 2026-01-13
affects: [c3-0, c3-1, c3-2, c3-3, c3-4, c3-5]
---

# Adopt C3 Architecture Documentation

## Status

**Implemented** - 2026-01-13

## Problem

expansive.ly is a greenfield project with complex architecture spanning spatial canvas rendering, graph databases, real-time collaboration, and AI integration. Without structured documentation:

- New team members struggle to understand system boundaries
- AI assistants lack context for meaningful contributions
- Architectural decisions drift without recorded rationale
- Component responsibilities become unclear over time

## Decision

Adopt C3 (Context-Container-Component) architecture documentation methodology to:

1. Document system structure at three levels of abstraction
2. Capture component contracts and hand-offs explicitly
3. Maintain references for cross-cutting patterns
4. Record architectural decisions with ADRs

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| arc42 | Too heavy for early-stage project, overkill for team size |
| Informal wiki | No structure enforcement, drift inevitable |
| Code comments only | Lacks system-level view, hard to navigate |
| C4 diagrams only | Diagrams without contracts miss runtime behavior |

C3 provides:
- Right-sized documentation (not too much, not too little)
- AI-readable structure for assistant context
- Clear navigation from system to component level
- Living documentation that stays close to code

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Context | c3-0 | Created with container map |
| Container | c3-1 through c3-5 | Created with component inventories |
| Component | Various | Created for complex components |
| Refs | refs/ | Created for cross-cutting patterns |

## Verification

- [x] Context README exists with container diagram
- [x] All containers have README with component inventory
- [x] Complex components have dedicated docs
- [x] References created for recurring patterns
- [x] TOC provides navigation
- [x] Settings file configured

## Consequences

### Positive

| Consequence | Impact |
|-------------|--------|
| Faster onboarding | New team members can understand architecture in hours, not days |
| AI-assisted development | Claude/GPT can provide contextual help with system understanding |
| Decision traceability | Future team can understand why decisions were made |
| Reduced architecture drift | Documented contracts prevent silent divergence |
| Living documentation | Stays in sync with code through developer workflow |

### Negative

| Consequence | Mitigation |
|-------------|------------|
| Documentation maintenance overhead | Keep docs minimal, audit periodically |
| Learning curve for C3 methodology | Provide templates and examples |
| Risk of stale documentation | Include docs in code review, run audits |
| Additional files in repository | Use .c3/ directory to isolate |

## Audit Record

| Date | Auditor | Finding | Resolution |
|------|---------|---------|------------|
| | | | |
