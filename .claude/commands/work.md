---
description: Deep work on c3-design with structured thinking loop
allowed-tools: Bash, Read, Glob, Grep, Task, AskUserQuestion, Skill, WebFetch, TodoWrite
---

# /work Command

## Arguments

$ARGUMENTS

## Intent Detection

Analyze $ARGUMENTS to determine intent:

| Intent | Patterns | Focus |
|--------|----------|-------|
| **Change** | "improve", "add", "build", "work on", "implement", "create", "enhance" | What should exist |
| **Troubleshoot** | "off", "broken", "wrong", "failing", "not working", "weird", "bug" | What's going wrong |

## Phase 1: Context Gathering (inline)

Load project understanding before launching subagents:

```
CLAUDE.md + references/skill-harness.md (principles)
skills/**/*.md (all skills, their connections)
git status + eval/results/ (current state)
```

Summarize: What is the current state relevant to $ARGUMENTS?

## Phase 2: Brainstorming (subagent)

Launch subagent with `superpowers:brainstorming`:

**Input:** Context summary + $ARGUMENTS + detected intent
**Goal:** Pin-point the actual question/goal through socratic dialogue
**Output:** Clear problem statement + proposed approach

If later phases find complexity, this phase may be revisited.

## Phase 3: Testing Strategy (subagent)

Launch subagent to discover how to verify the work:

**Input:** Brainstorm output (goal + approach)
**Method:** Socratic questioning + educated guesses based on existing eval/

**Questions to answer:**
- How would you know it works?
- What's the simplest check?
- What breaks if this is wrong?
- What existing tests/patterns can we reuse?

**Output criteria - test approach must be:**
- Fast (seconds, not minutes)
- Good coverage (key paths)
- Cheap to maintain (minimal fixtures)
- Human readable (clear pass/fail)

**Loop back to brainstorming if:** approach is too complex to test practically

## Phase 4: Writing Plans (subagent)

Launch subagent with `superpowers:writing-plans`:

**Input:** Brainstorm output + testing strategy
**Output:** Two separate plans in `docs/plans/`:
1. `YYYY-MM-DD-<topic>-implementation.md` - Implementation steps
2. `YYYY-MM-DD-<topic>-test.md` - Test plan with verification steps

**Loop back if:**
- Scope issue discovered → return to brainstorming
- Test gap found → return to testing strategy

## Phase 5: Plan Review (subagent)

Launch review subagent to validate completeness:

**Completeness checks:**
- [ ] All steps defined (no "TBD" or vague steps)
- [ ] No ambiguity (each step is actionable)
- [ ] Clear success criteria (how to know step is done)
- [ ] Test plan covers implementation plan

**Use diashort to visualize:**
- Implementation flow diagram
- Test coverage map
- Dependency graph (what depends on what)

Present diagrams to user. Visual gaps are easier to spot than text gaps.

**Loop back if:**
- Gaps visible in diagrams → return to writing-plans
- Fundamental approach issue → return to brainstorming
- Proceed when plans are visually clear and complete

## Phase 6: Execution (subagent)

Launch subagent with `ralph-loop:ralph-loop` + `superpowers:executing-plans`:

**Input:** Both plans (implementation + test)
**Exit criteria:** All tests from test plan pass

The ralph-loop develops and refines until verification passes:
- Execute implementation steps
- Run test plan verification
- If tests fail → refine and retry
- If tests pass → done

## Flow Diagram

https://diashort.apps.quickable.co/d/e0f80dd4

## Notes

- Workflow is iterative, not waterfall
- Later phases can loop back to earlier phases when complexity is discovered
- Each subagent receives context from previous phases
- diashort diagrams make review visual and gaps obvious
