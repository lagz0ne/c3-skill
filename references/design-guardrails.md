# C3 Design Guardrails

Critical operational guidance for the c3-design skill. These principles, mistakes, and red flags ensure effective architecture design through hypothesis-driven exploration.

## Key Principles

| Principle | Application |
|-----------|-------------|
| **Hypothesis first** | Form from TOC, don't ask directly for location |
| **Explore to validate** | Investigate before confirming |
| **Socratic during exploration** | Questions confirm understanding, not discover location |
| **ID-based navigation** | Use document/heading IDs, not keyword search |
| **Higher = bigger impact** | Upstream/higher-level discoveries trigger revisit |
| **ADR as stream** | Capture journey, not just final answer |
| **Iterate freely** | Loop until stable, don't force forward |

## Common Mistakes

- Skipping the TOC/ID sweep and diving into files, which hides upstream impacts and duplicate IDs.
- Asking the user to point to files instead of forming a hypothesis from the TOC and existing docs.
- Drafting an ADR before the hypothesis → explore → discover loop stabilizes.
- Treating examples as throwaway and allowing duplicate IDs or missing TOC to persist.
- **Skipping Phase 3 (ADR+Plan creation)** and updating documents directly.
- **Creating ADR without Implementation Plan** - they are inseparable.
- **Orphan layer changes** - "Changes Across Layers" without corresponding Code Changes.
- **Orphan verifications** - Verification items without corresponding Acceptance Criteria.
- **Vague code locations** - "update auth" instead of `src/handlers/auth.ts:validateToken()`.
- **Skipping Phase 4 (Handoff)** and ending the session without executing settings.yaml steps.
- **Not creating TodoWrite items** for phase tracking.
- **Ignoring settings.yaml** handoff configuration.

## Red Flags & Counters

| Rationalization | Counter |
|-----------------|---------|
| "No time to refresh the TOC, I'll just skim files" | Stop and build/read the TOC first; C3 navigation depends on it. |
| "Examples can keep duplicate IDs, they're just sample data" | IDs must be unique or locate/anchor references break—fix collisions before scoping. |
| "I'll ask the user where to change docs instead of hypothesizing" | Hypothesis bounds exploration and prevents confirmation bias; form it before asking questions. |
| "The scope is clear, I can skip the ADR" | **NO.** ADR is mandatory. It documents the journey and enables review. |
| "I'll just update the docs and mention what I did" | **NO.** ADR first, then doc updates. This is non-negotiable. |
| "I'll add the Implementation Plan later" | **NO.** ADR and Plan are created together. Plan is part of Phase 3 gate. |
| "The code changes are obvious, no need to list them" | **NO.** Explicit mapping enables audit verification. List them. |
| "Handoff is just cleanup, I can skip it" | **NO.** Handoff ensures tasks are created and team is informed. Execute it. |
| "No settings.yaml means no handoff needed" | **NO.** Use default handoff steps. Always confirm completion with user. |

## Red Flags That Mean Pause

These signals indicate you should stop and address the issue:

- `.c3/TOC.md` missing or obviously stale.
- Component IDs reused across containers or layers.
- ADR being drafted without notes from hypothesis, exploration, and discovery.
- **Updating C3 documents without an ADR file existing.**
- **ADR without Implementation Plan section.**
- **"Changes Across Layers" count ≠ "Code Changes" count.**
- **"Verification" count ≠ "Acceptance Criteria" count.**
- **Ending the session without executing handoff.**
- **No TodoWrite items for the 4 phases.**

## Enforcement

These guardrails are non-negotiable. They exist because:

1. **Hypothesis-driven exploration** prevents confirmation bias and missed upstream impacts
2. **Mandatory ADR+Plan** enables review, audit, and team coordination
3. **Phase tracking** prevents skipping critical steps
4. **Handoff execution** ensures work transitions properly to implementation

When you feel tempted to skip a guardrail, that's exactly when you need it most.
