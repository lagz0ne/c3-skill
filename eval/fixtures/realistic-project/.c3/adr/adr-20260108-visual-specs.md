---
id: adr-20260108-visual-specs
title: Add Visual Specs Component for Screen Generation Consistency
status: implemented
date: 2026-01-08
affects: [c3-1, c3-114, c3-133]
---

# Add Visual Specs Component for Screen Generation Consistency

## Status

**Proposed** - 2026-01-08

## Problem

Frontend design review identified that existing C3 documentation (c3-131 IA, c3-132 Flows, c3-133 Patterns) provides strong behavioral consistency but weak visual consistency. Missing: typography scale, spacing tokens, and visual wireframes. This prevents consistent generation of new screens.

## Decision

Add c3-115 Visual Specs as an Auxiliary component documenting:
1. Typography scale (headings, body, labels, captions)
2. Spacing tokens (4px-64px scale)
3. ASCII wireframes for key patterns (PTN-01, PTN-03)
4. Status badge color mapping
5. Component gallery references

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| Expand c3-114 Design System | Would bloat existing doc; visual specs are consumption-focused, design system is definition-focused |
| Add to c3-133 UI Patterns | Patterns are behavioral; visual specs are spatial/typographic |
| External design tool | Breaks C3 self-contained documentation principle |

c3-115 complements existing docs:
- c3-114 defines **what** (tokens, variants)
- c3-115 documents **how to apply** (scales, wireframes)
- c3-133 documents **when** (behavioral triggers)

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-1 README | Add c3-115 to Auxiliary table, update mermaid diagram |
| Component | c3-115 (new) | Create Visual Specs document |
| Component | c3-114 | Add reference to c3-115 in References section |
| Component | c3-133 | Add reference to c3-115 in Uses table |

## Verification

- [ ] c3-115 created with proper C3 format (frontmatter, Uses, Conventions, Testing, References)
- [ ] c3-1 README updated with c3-115 in Auxiliary table
- [ ] c3-1 mermaid diagram includes c3-115
- [ ] c3-114 references c3-115
- [ ] c3-133 Uses table includes c3-115
- [ ] `/c3 audit` passes
