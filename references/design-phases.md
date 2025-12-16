# C3 Design - Phases

## Phase 1: Surface Understanding

Read TOC, form hypothesis about what layers are affected.

**Output:** Initial layer impact hypothesis
**Gate:** Hypothesis formed

## Phase 2: Iterative Scoping

Loop: HYPOTHESIZE → EXPLORE → DISCOVER

1. **Hypothesize:** "This change affects [layers]"
2. **Explore:** Use layer skills to investigate
3. **Discover:** Find actual impacts, adjust hypothesis

**Output:** Stable scope (layers, docs, changes)
**Gate:** No new discoveries in last iteration

## Phase 3: ADR Creation

Create ADR in `.c3/adr/adr-YYYYMMDD-slug.md`:
- Status: proposed
- Changes Across Layers: what docs change
- Verification Checklist: how to verify implementation

**Output:** ADR file exists
**Gate:** File created with all sections

## Phase 4: Handoff

Execute `.c3/settings.yaml` handoff steps (or defaults).

**Output:** Handoff complete
**Gate:** All steps executed
