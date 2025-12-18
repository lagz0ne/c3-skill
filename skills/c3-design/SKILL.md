---
name: c3-design
description: Use when designing, updating, or exploring system architecture with C3 methodology - iterative scoping through hypothesis, exploration, and discovery across Context/Container/Component layers
---

# C3 Architecture Design

## Overview

Transform requirements into C3 documentation through iterative scoping. Also supports exploration-only mode.

**Core principle:** Hypothesis → Explore → Discover → Iterate until stable.

**Announce:** "I'm using the c3-design skill to guide architecture design."

**CRITICAL:** Read `references/design-guardrails.md` before starting. These guardrails are non-negotiable.

## Mode Detection

| Intent | Mode |
|--------|------|
| "What's the architecture?" | Exploration |
| "How does X work?" | Exploration |
| "I need to add/change..." | Design |
| "Why did we choose X?" | Exploration |

### Exploration Mode

See `references/design-exploration-mode.md` for full workflow.

Quick: Load TOC → Present overview → Navigate on demand.

### Design Mode

Full workflow with mandatory phases.

## Design Phases

**IMMEDIATELY create TodoWrite items:**
1. Phase 1: Surface Understanding
2. Phase 2: Iterative Scoping
3. Phase 3: ADR Creation (MANDATORY)
4. Phase 4: Handoff

See `references/design-phases.md` for detailed phase requirements.

**Rules:**
- Mark `in_progress` when starting
- Mark `completed` only when gate met
- Phase 3 gate: ADR file MUST exist
- Phase 4 gate: Handoff steps executed

### 📋 ADR Status: New ADRs Start as `proposed`

**When creating an ADR in Phase 3:**
- Set `status: proposed` in frontmatter
- ADR will NOT appear in TOC until `implemented`
- Status workflow: `proposed` → `accepted` → `implemented`

See `references/adr-template.md` for status values and workflow.

## Quick Reference

| Phase | Output | Gate |
|-------|--------|------|
| 1. Surface | Layer hypothesis | Hypothesis formed |
| 2. Scope | Stable scope | No new discoveries |
| 3. ADR | ADR file | File exists |
| 4. Handoff | Complete | Steps executed |

## Layer Skill Delegation

| Impact | Delegate To |
|--------|-------------|
| Container inventory changes | c3-context-design |
| Component organization | c3-container-design |
| Implementation details | c3-component-design |

---

## ⛔ SKILL DELEGATION ENFORCEMENT (MANDATORY)

**Rule:** When work requires a layer skill, INVOKE it. Never describe what it "would do."

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "c3-container-design would create..." | Hallucinating skill behavior | Use Skill tool → c3-container-design |
| "Following c3-component-design patterns..." | You don't have its patterns loaded | Invoke the skill first |
| "This is simple, I'll handle it directly" | Layer skills have guardrails you're bypassing | Always invoke for layer work |
| Summarizing a skill's output without invoking | Fabrication | Invoke, get real output |

### Red Flags

🚩 Using layer skill name as a noun ("the c3-container-design approach")
🚩 Describing layer skill output without a Skill tool call in the conversation
🚩 "I'll apply c3-X-design principles" without invoking it

### Self-Check

- [ ] Did I use the Skill tool for each layer I'm affecting?
- [ ] Am I quoting actual skill output, not imagined output?
- [ ] Is there a Skill tool invocation in my message for each delegation?

### Escape Hatch

None. Layer work = layer skill invocation. No exceptions.

---

## ⛔ PHASE GATE ENFORCEMENT (MANDATORY)

**Rule:** Each phase has a gate condition. Phase is NOT complete until gate is met.

### Phase Gates

| Phase | Gate Condition | Evidence Required |
|-------|----------------|-------------------|
| 1. Surface | Hypothesis formed | Written hypothesis statement |
| 2. Scope | No new discoveries in last iteration | Explicit "scope stable" declaration |
| 3. ADR | ADR file exists on disk | `ls .c3/adr/adr-YYYYMMDD-*.md` shows file |
| 4. Handoff | All handoff steps executed | Each step has completion evidence |

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "Phase 3 complete" without ADR file | Gate not met | Create the file first |
| Marking phase complete in TodoWrite prematurely | Misrepresents status | Only mark after evidence |
| Skipping Phase 3 "because changes are small" | ADR documents journey, not size | Always create ADR |
| "Handoff: informed user" without executing steps | Handoff is action, not notification | Execute each step |

### Red Flags

🚩 TodoWrite shows Phase 3 complete but no `adr-*.md` in conversation
🚩 Moving to Phase 4 without stable scope declaration
🚩 Ending conversation without Phase 4 handoff execution

### Self-Check

- [ ] For each phase I'm marking complete, what is my evidence?
- [ ] Can I point to the artifact/action that satisfies the gate?

### Escape Hatch

None. Gates exist because skipping them causes downstream failures.

---

## ⛔ OUTPUT VERIFICATION ENFORCEMENT (MANDATORY)

**Rule:** Claiming completion requires verification evidence in the conversation.

### Verification Requirements

| Claim | Required Evidence |
|-------|-------------------|
| "Created file X" | File write command + success, or `ls` showing file |
| "Updated document" | Edit/write command visible |
| "Structure is correct" | Validation checklist executed with results |
| "Delegated to skill X" | Skill tool invocation visible |
| "ADR created" | File path + frontmatter visible |

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "I've created the container docs" (no file ops visible) | No evidence of creation | Show the write commands |
| "Following V3 structure" (no validation) | Structure errors are common | Run validation checklist |
| "Delegation complete" (no skill invocation) | Hallucination | Show Skill tool usage |

### Red Flags

🚩 Completion claim without corresponding tool usage
🚩 "Done" without checklist execution
🚩 Describing artifacts that weren't created in this conversation

### Self-Check

- [ ] For each artifact I claim exists, is there evidence of its creation?
- [ ] Did I run the skill's validation checklist?
- [ ] Can a reviewer see proof in this conversation?

### Escape Hatch

None. Unverified completion = not complete.

---

## Checklist

- [ ] Mode detected (exploration vs design)
- [ ] If design: TodoWrite phases created
- [ ] Each phase completed with gate
- [ ] ADR created (design mode)
- [ ] Handoff executed

## Related

- `references/design-guardrails.md` - **READ FIRST:** Key principles, common mistakes, red flags
- `references/design-exploration-mode.md`
- `references/design-phases.md`
- `references/adr-template.md`
