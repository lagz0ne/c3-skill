# C3 skill design

This is a repository containing a Claude code skill called c3. c3 is a trimmed down concept from c4, focusing on rationalizing architectural and architectural change for large codebase

# Required skills, always load them before hand
- load /superpowers:brainstorming, always do
- load /superpowers-developing-for-claude-code:developing-claude-code-plugins, always do
- use AskUserQuestionTool where possible, that'll give better answer

# Workflow
- Starts with brainstorming to understand clearly the intention
- Once it's all understood, use writing-plan and implement in parallel using subagent
- Delegate to /release command once things is done, confirm with user as needed. Assume to patch by default

---

# Evaluation Philosophy

## Core Principle

**Test outcomes, not structure.** Meaningful evaluation tests whether artifacts achieve their purpose, not whether they match a template.

## Evaluation Quality Hierarchy

| Level | Approach | What It Tests |
|-------|----------|---------------|
| Best | Outcome-based | Does artifact enable its purpose? |
| Mid | Behavior-based | Does artifact work correctly? |
| Worst | Structure-based | Does artifact match expected shape? |

Always start from outcomes and work backward to find minimal structure requirements.

## Question-Based Evaluation Pattern

The gold standard for evaluating generated artifacts:

```
Codebase Analysis → Question Generation → Fresh-Context Answering → Verification
       ↓                    ↓                       ↓                    ↓
  Extract truth      Simulate real            No cheating         Ground truth
  from source        developer needs          (isolated LLM)      comparison
```

**Why this works:**
- Grounded in THIS system, not generic templates
- Simulates real use with fresh context (prevents information leakage)
- Binary correctness - either the answer matches reality or it doesn't
- Actionable failures - know exactly which question failed and why

**Implementation:** `eval/lib/question-eval.ts`

## Anti-Patterns to Avoid

| Anti-Pattern | Why It Fails | Alternative |
|--------------|--------------|-------------|
| Golden example comparison | Measures similarity, not utility | Question-based testing |
| Structural checklists | Creates template compliance, not quality | Outcome verification |
| Generic dimensions | Same scores for different codebases | Codebase-specific questions |
| Single-context evaluation | LLM can cheat with prior knowledge | Fresh context isolation |
| Bidirectional links | Maintenance burden, stale links | Unidirectional hierarchy |
| Monolithic docs | Hard to maintain and navigate | Separate files per concept |

---

# Skill Development Philosophy

## The Structure vs Outcome Balance

Shift from prescriptive structure to goal-oriented guidance:

| Structural (Avoid) | Outcome-Oriented (Prefer) |
|--------------------|---------------------------|
| "Create separate file per component" | "Each component should be findable and answerable" |
| "Follow this directory structure" | "Structure should enable navigation" |
| "Add these specific sections" | "Enable developer to understand X" |

## Skill Development Loop

```
Goals → Minimal Skill → Meaningful Eval → Run → Trace Failures → Update Skill → Repeat
```

1. **Start with Goals** - What should the skill enable? What questions should users answer?
2. **Write Minimal Skill** - Focus on outcomes, not structure
3. **Build Meaningful Eval** - Question-based, grounded in actual use
4. **Run Eval** - Identify failing questions
5. **Trace Failures** - Map failures back to skill gaps
6. **Update Skill** - Add targeted guidance, remove unhelpful rules
7. **Repeat** - Improving scores should improve real utility

## Skill Writing Principles

1. **Fewer meaningful sections > many shallow sections** - Don't prescribe structure for structure's sake
2. **Outcome-focused instructions** - "Docs should enable X" not "Docs should contain Y"
3. **Anti-goals are important** - Explicitly state what NOT to do (came from eval failures)
4. **Progressive disclosure** - Core requirements first, details only when needed

## When to Use Prescriptive vs Goal-Oriented

| Use Prescriptive | Use Goal-Oriented |
|------------------|-------------------|
| Hard technical constraints (file names, IDs) | Content quality |
| Integration points (CI expects certain paths) | Organization decisions |
| Non-negotiable conventions | Style and depth |

---

# Running Evaluations

```bash
# Expectations-based (structural)
bun eval/run.ts eval/cases/onboard-simple.yaml

# Question-based (meaningful)
bun eval/test-question-eval.ts <codebase-path> <docs-path>

# Example
bun eval/test-question-eval.ts eval/fixtures/simple-express-app /tmp/output/.c3
```

---

# Plugin Troubleshooting

## When to Trigger

Run validation when **plugin behavior doesn't match expectations** - components not loading, invocations failing, or after structural changes.

## How to Validate

1. Load `plugin-dev:plugin-structure` skill for structure reference
2. Run `plugin-dev:plugin-validator` agent on plugin root
3. Fix issues reported by validator
4. Restart Claude Code session (components load at startup)