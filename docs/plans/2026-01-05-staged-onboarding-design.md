# Staged Onboarding with Component Progression

**Date:** 2026-01-05
**Status:** Proposed

---

## Critical: Always Use AskUserQuestionTool

**All Socratic questioning MUST use the `AskUserQuestion` tool.**

This applies across all C3 skills:
- **Onboard**: Clarifying architecture at each level
- **Alter**: Understanding intent, scope, and impact
- **Query**: Clarifying user intent before exploration

Why:
- Structured questions get structured answers
- Multiple-choice reduces ambiguity
- User can select rather than type
- Conversation stays focused

Never use plain text questions when AskUserQuestion can be used instead.

---

## Problem

The current C3 adoption process stops at the Container level and doesn't detail Components. Users expect a full architectural documentation set, but the onboarding creates:
- Context doc (complete)
- Container docs with component inventories (complete)
- Component docs (missing)

## Solution

Transform adoption into a **recursive learning loop** that descends through all architectural levels, with Socratic questioning to build understanding and bidirectional navigation to resolve conflicts.

## Design

### Core Loop

At each architectural level, the adoption follows:

```
┌─────────────────────────────────────────────────────────────┐
│  ANALYZE                                                    │
│  - Read code/configs at this level                          │
│  - Form hypotheses about structure & purpose                │
│  - Build internal question list for unknowns                │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  ASK (Socratic Questioning)                                 │
│  - Probe critical unknowns one at a time                    │
│  - Challenge assumptions                                    │
│  - Continue until NO OPEN QUESTIONS remain                  │
│  - Use AskUserQuestion tool for interaction                 │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  ASSUME & SYNTHESIZE                                        │
│  - Make reasonable assumptions for administrative decisions │
│  - Fill documentation templates                             │
│  - Update ADR-000 progress tracker                          │
│                                                             │
│  Tiered Assumptions:                                        │
│  • HIGH-IMPACT (ask first):                                 │
│    - New External Systems                                   │
│    - Container reassignments                                │
│    - Structural boundary changes                            │
│  • LOW-IMPACT (auto-proceed):                               │
│    - Linkage reasoning                                      │
│    - Naming standardization                                 │
│    - Adding discovered dependencies                         │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  REVIEW                                                     │
│  - Present synthesized documentation to user                │
│  - User reviews for CORRECTNESS only                        │
│  - Corrections feed back into understanding                 │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
                  [ Auto-descend to next level ]
```

### Level Progression (Fixed Order)

```
1. Context (c3-0)
      ↓
2. Containers (c3-1, c3-2, ...)
      ↓
3. Components: Auxiliary (conventions/patterns)
      ↓
4. Components: Foundation (primitives)
      ↓
5. Components: Feature (business logic)
```

**Rationale for fixed ordering:**
- Auxiliary conventions are independent (how we use external tools)
- Foundation depends on auxiliary conventions
- Features depend on both
- This builds a mental model of dependency layering

### Bidirectional Navigation

When deeper analysis reveals conflicts or gaps, **ascend** to the appropriate level:

```
Descend to Component X
     │
     ▼
┌─────────────────────────────────────┐
│  Conflict/Gap detected?             │
│  • Component doesn't fit container  │
│  • Linkage contradicts parent doc   │
│  • Missing dependency in parent     │
└──────────────┬──────────────────────┘
               │
        Yes    │    No
         ▼     │     ▼
    ┌─────────────────────┐
    │  ASCEND             │   Continue
    │  - Go to parent     │   descent
    │  - Apply tiered     │
    │    assumption rules │
    │  - Update parent    │
    │  - Re-descend       │
    └─────────────────────┘
```

**No circuit breaker:** Iterate until architecture is coherent or user interrupts. Complex architectures may require multiple ascent/descent cycles.

### Confidence Definition

**Confident = No open questions in internal list**

The AI maintains an internal question list during analysis. Questions are added when:
- Purpose of a component/linkage is unclear
- Naming suggests ambiguity
- Dependencies are implicit
- Patterns don't match conventions

When the list is empty, proceed to synthesis.

### Progress Tracking in ADR-000

The adoption ADR (`adr-00000000-c3-adoption.md`) includes a Progress section:

```markdown
## Adoption Progress

| Level | Category | Status | Documented | Remaining |
|-------|----------|--------|------------|-----------|
| Context | - | ✅ Complete | 1 | 0 |
| Containers | - | ✅ Complete | 3 | 0 |
| Components | Auxiliary | 🔄 In Progress | 5 | 3 |
| Components | Foundation | ⏳ Pending | 0 | 4 |
| Components | Feature | ⏳ Pending | 0 | 12 |

### Open Questions
- (none currently)

### Pending High-Impact Decisions
- (none currently)
```

### Interaction Mode

**Always interactive.** User is present throughout adoption. No async/batch mode.

The Socratic questioning approach requires active user engagement to:
- Validate assumptions about architecture intent
- Provide domain knowledge not visible in code
- Correct misunderstandings before they propagate

## Example Flow

```
=== CONTEXT LEVEL ===
AI: Analyzing codebase structure...
AI: I see 3 main directories: api/, web/, worker/.
    Are these separate deployable containers or modules within one app?
User: Separate containers, deployed independently.
AI: The api/ connects to a PostgreSQL database. Is this the only data store?
User: Also Redis for caching.
AI: [Synthesizes Context doc with 3 containers + 2 external systems]
AI: Here's the Context documentation. [shows output]
User: Looks correct.

=== CONTAINER LEVEL (api/) ===
AI: Analyzing api/ container...
AI: I see authentication middleware. Does this use JWT, sessions, or OAuth?
User: JWT with refresh tokens.
AI: [Synthesizes Container doc with auth as Foundation component]
...

=== COMPONENT LEVEL (Auxiliary) ===
AI: Documenting Auxiliary conventions...
AI: I see consistent patterns in error handling across api/.
    Errors are wrapped in a standard format with code/message/details.
    [AUTO-PROCEEDING: adding as Auxiliary convention "Error Handling"]
AI: The API client in web/ uses a retry pattern with exponential backoff.
    This appears to be a convention. Documenting as "API Client Patterns".
...

=== ASCENT EXAMPLE ===
AI: Analyzing Foundation component "AuthProvider"...
AI: This depends on an OAuth provider (Google) not listed in Context.
    [HIGH-IMPACT: Ascending to Context level]
AI: Adding "Google OAuth" as External System in Context.
    AuthProvider → Google OAuth linkage added.
    Is this correct?
User: Yes.
AI: [Updates Context, re-descends to complete AuthProvider doc]
```

## Implementation Changes

### Files to Modify

1. **commands/onboard.md** - Rewrite to implement staged loop
2. **templates/adr-000.md** - Add Progress section template
3. **references/implementation-guide.md** - Update with staged adoption guidance

### New Behaviors

1. **Question tracking:** Maintain internal list, surface via Socratic dialogue
2. **Tiered assumptions:** Classify assumptions, ask for high-impact only
3. **Progress updates:** Update ADR-000 after each synthesis
4. **Bidirectional state:** Track current level, enable ascent/descent

## Success Criteria

- [ ] Adoption produces complete documentation (Context + Containers + Components)
- [ ] User only answers correctness questions, not administrative decisions
- [ ] Conflicts discovered at lower levels trigger automatic parent updates
- [ ] ADR-000 shows clear progress at any interruption point
- [ ] Component docs follow Auxiliary → Foundation → Feature order

---

## Alignment: Alter Command

The same recursive learning loop applies to the alter workflow. Every stage uses:

```
ANALYZE → ASK (until confident) → SYNTHESIZE → REVIEW
```

### Alter Stages with Loop Pattern

| Stage | Analyze | Ask Until Confident | Synthesize | Review |
|-------|---------|---------------------|------------|--------|
| **Intent** | What is user asking? | "Is this a new feature or a fix?", "What's the goal?" | Clear intent statement | User confirms |
| **Understand** | Read affected C3 layers | "This component uses X - is that still true?", "Any recent changes not in docs?" | Current state summary | User confirms |
| **Scope** | Map impact to layers | "Will this affect the API contract?", "What depends on this?" | Impact assessment | User confirms scope |
| **ADR** | - | - | Create ADR document | User accepts/rejects |
| **Plan** | Order of operations | "Should we update tests first?", "Any deployment considerations?" | Execution plan | User approves |
| **Execute** | Apply changes per layer | If conflict → ascend, ask, fix, re-descend | Code + doc changes | Implicit (next stage) |
| **Verify** | Run audit | - | Audit results | Pass/fail → loop if fail |

### Bidirectional Navigation in Alter

During execution, discoveries may require ascending:

```
Executing Component change...
  → Discovers linkage not in Container doc
  → ASCEND to Scope stage
  → "This change also affects Container linkages. Adding to scope."
  → Update ADR if scope grew significantly
  → Re-descend and continue
```

### Tiered Assumptions in Alter

| High-Impact (ask first) | Low-Impact (auto-proceed) |
|-------------------------|---------------------------|
| Scope expansion | Adding linkage reasoning |
| New affected layer | Minor doc wording fix |
| Breaking change detected | Updating diagram to match |
| ADR revision needed | Fixing ID inconsistency |

### Shared Principles (Onboard + Alter)

1. **Confidence = No open questions** - Don't proceed until confident
2. **Tiered assumptions** - High-impact asks, low-impact proceeds
3. **Bidirectional** - Ascend when conflicts found, no circuit breaker
4. **User validates correctness** - AI handles administrative decisions
5. **Progress visible** - ADR tracks what's done/pending

---

## Alignment: Query Command

Query is read-only exploration, so uses a **lighter pattern**:

```
Query → [Clarify Intent (Socratic)] → Navigate → Extract → Explore → Present
             │                  │
             └── if ambiguous ──┘
```

### When to Clarify

| Trigger | Action |
|---------|--------|
| Vague query ("how does X work?") | Ask which aspect |
| Multiple interpretations | Ask which one |
| Unclear scope | Ask for specificity |
| C3 ID provided | Skip - direct lookup |
| Specific query | Skip - proceed |

### Difference from Onboard/Alter

- **No synthesis/review loop** - Query presents findings, doesn't create docs
- **Clarification only at start** - Not at each navigation step
- **No bidirectional ascent** - Linear exploration, present what's found
