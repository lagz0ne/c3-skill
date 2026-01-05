---
description: Initialize C3 architecture documentation for this project
argument-hint: [--force]
allowed-tools: Bash(rm:*), Bash(test:*), Bash(PROJECT=*), Read, Glob, Grep, Write, Edit, AskUserQuestion
---

# C3 Onboarding

## Current State

- .c3 directory exists: !`test -d "$CLAUDE_PROJECT_DIR/.c3" && echo "yes" || echo "no"`
- Arguments: $ARGUMENTS

## Instructions

You are initializing C3 architecture documentation using a **staged recursive learning loop**.

### Pre-flight Check

**If .c3 already exists AND --force was NOT passed:**

Stop and inform:
```
C3 architecture documentation already exists.
To start fresh: /onboard --force
```

**If .c3 already exists AND --force WAS passed:**
```bash
rm -rf "$CLAUDE_PROJECT_DIR/.c3"
```

### Critical: Always Use AskUserQuestionTool

All Socratic questioning MUST use the `AskUserQuestion` tool - never plain text questions.
- Structured questions get structured answers
- Multiple-choice reduces ambiguity
- User can select rather than type

### The Learning Loop

At each architectural level, follow this pattern:

```
ANALYZE → ASK (Socratic) → ASSUME/SYNTHESIZE → REVIEW → DESCEND
              │                    │
              └── until no ────────┘
                  open questions

On conflict at deeper level:
ASCEND → fix parent → re-descend
```

### Level Progression (fixed order)

```
1. Context (c3-0)
2. Containers (c3-1, c3-2, ...)
3. Components: Auxiliary
4. Components: Foundation
5. Components: Feature
```

---

## Stage 1: Context Level

### 1.1 Analyze

Read the codebase structure. Form hypotheses about:
- What is this system?
- Who/what interacts with it (actors)?
- What are the major deployable units (containers)?
- What external systems does it depend on?

Build an **open questions list** for anything unclear.

### 1.2 Ask (Socratic)

Use AskUserQuestion for each open question until the list is empty.

Example questions:
- "I see directories api/, web/, worker/. Are these separate deployable containers or modules within one app?"
- "The code connects to PostgreSQL and Redis. Are there other data stores?"
- "I see webhook handlers. Who sends these webhooks?"

**Continue until you have NO OPEN QUESTIONS about Context-level architecture.**

### 1.3 Gather Basics

Use AskUserQuestion to collect:
- **PROJECT**: Project name
- **Containers**: Confirm the containers you identified

### 1.4 Initialize Structure

Run the init script:
```bash
PROJECT="<name>" C1="<container1>" C2="<container2>" ... "${CLAUDE_PLUGIN_ROOT}/scripts/c3-init.sh"
```

### 1.5 Synthesize Context

Fill `.c3/README.md` with:
1. Mermaid diagram (actors → containers → external systems)
2. Actors table with IDs (A1, A2...)
3. Containers table with IDs (c3-1, c3-2...)
4. External Systems table with IDs (E1, E2...)
5. Linkages with REASONING (why they connect)

### 1.6 Review

Present the Context documentation. User reviews for correctness.
Corrections feed back into understanding.

### 1.7 Update Progress

Edit ADR-000 Progress section:
```
| Context | - | ✅ Complete | 1 | 0 |
```

---

## Stage 2: Container Level

**For each container (c3-1, c3-2, ...):**

### 2.1 Analyze

Read the container's code. Form hypotheses about:
- What is this container's responsibility?
- What components exist inside?
- How do they categorize? (Foundation/Auxiliary/Feature)
- How does this container fulfill Context-level linkages?

Build open questions list.

### 2.2 Ask (Socratic)

Use AskUserQuestion until no open questions.

Example questions:
- "I see AuthMiddleware used everywhere in api/. Is this a Foundation component others depend on?"
- "There's a consistent error format across all handlers. Is this an explicit convention or emergent?"
- "ProductService and OrderService both call PaymentGateway. Is PaymentGateway internal or external?"

### 2.3 Handle Conflicts (Ascent)

If you discover something that conflicts with Context:
- **High-impact** (new External System, container boundary change): Ask user to confirm, then update Context
- **Low-impact** (add linkage reasoning, minor name fix): Update Context automatically, note it in review

Example:
```
"While analyzing api/, I found it connects to Stripe (not in Context externals).
Adding Stripe as External System E3. Is this correct?"
```

### 2.4 Synthesize Container

Fill `.c3/c3-N-<slug>/README.md` with:
1. Container mermaid diagram
2. Components inventory by category:
   - Foundation (primitives others build on)
   - Auxiliary (conventions/patterns)
   - Feature (domain-specific)
3. Fulfillment section (which components handle Context links)
4. Linkages with REASONING

### 2.5 Review

Present Container documentation. User reviews for correctness.

### 2.6 Update Progress

Edit ADR-000:
```
| Containers | - | ✅ Complete | N | 0 |
```

---

## Stage 3: Component Level - Auxiliary

**For each container, document Auxiliary components first.**

Auxiliary = "How we use X" conventions (API patterns, error handling, state management)

### 3.1 Analyze

From Container inventory, identify Auxiliary components.
For each, understand:
- What convention does this establish?
- What are the do/don't rules?
- Which other components follow this?

### 3.2 Ask (Socratic)

Example questions:
- "I see all API calls use retry with exponential backoff. Is this a documented convention or should I document it now?"
- "Error responses follow a code/message/details format. Are there exceptions to this pattern?"

### 3.3 Handle Conflicts (Ascent)

If Auxiliary pattern reveals Container-level issues:
- Missing component in inventory → Add to Container
- Pattern applies to multiple containers → Consider promoting to Context

### 3.4 Synthesize

Create `.c3/c3-N-<slug>/c3-N01-<component>.md` files using Auxiliary template:
- Conventions table (Rule | Why)
- Applies To section
- Testing guidance

Use sequential IDs: c3-101, c3-102... for container c3-1.

### 3.5 Review & Progress

Present Auxiliary docs. User reviews.
Update ADR-000:
```
| Components | Auxiliary | ✅ Complete | X | 0 |
```

---

## Stage 4: Component Level - Foundation

**For each container, document Foundation components.**

Foundation = Primitives others build on (Router, AuthProvider, BaseComponent)

### 4.1 Analyze

From Container inventory, identify Foundation components.
For each, understand:
- What does this provide?
- What does it expect?
- What happens when it fails?

### 4.2 Ask (Socratic)

Example questions:
- "AuthProvider handles JWT refresh. What happens when refresh fails?"
- "Router supports nested routes. Are there depth limits or conventions?"

### 4.3 Handle Conflicts (Ascent)

Foundation discoveries may reveal:
- Missing Auxiliary conventions → Ascend to Stage 3
- External dependencies → Ascend to Context

### 4.4 Synthesize

Create component files using Foundation template:
- Contract table (Provides | Expects)
- Edge Cases table (Scenario | Behavior)
- Testing scenarios

### 4.5 Review & Progress

Present Foundation docs. User reviews.
Update ADR-000:
```
| Components | Foundation | ✅ Complete | X | 0 |
```

---

## Stage 5: Component Level - Feature

**For each container, document Feature components.**

Feature = Domain-specific business logic (CheckoutFlow, ProductCatalog)

### 5.1 Analyze

From Container inventory, identify Feature components.
For each, understand:
- What does this do for users?
- What Foundation/Auxiliary does it use?
- What triggers it and what's the result?

### 5.2 Ask (Socratic)

Example questions:
- "CheckoutFlow handles payment. What happens on payment failure?"
- "ProductSearch uses ElasticSearch. Is that the E2 External System from Context?"

### 5.3 Handle Conflicts (Ascent)

Feature analysis often reveals the most conflicts:
- Uses Foundation not documented → Ascend to Stage 4
- Follows convention not documented → Ascend to Stage 3
- Calls external not in Context → Ascend to Context

### 5.4 Synthesize

Create component files using Feature template:
- Uses table (Category | Component | For)
- Behavior table (Trigger | Result)
- Testing scenarios

### 5.5 Review & Progress

Present Feature docs. User reviews.
Update ADR-000:
```
| Components | Feature | ✅ Complete | X | 0 |
```

---

## Completion

When all stages complete:

1. **Final ADR-000 update**: Mark adoption complete
2. **Summary**: List what was created
3. **Next steps**:
   - Run `/c3 audit` to verify consistency
   - Use `/c3` for ongoing architecture work
   - Use `/c3:alter` for changes

---

## Tiered Assumption Rules

### High-Impact (ask first)
- New External System discovered
- Container boundary change
- Component reassignment between containers
- Structural changes to existing docs

### Low-Impact (auto-proceed, note in review)
- Adding linkage reasoning
- Naming standardization
- Adding discovered component to inventory
- Filling empty template sections

---

## Bidirectional Navigation

When conflict detected during descent:

1. Identify which level owns the conflict
2. **High-impact**: Ask user before ascending
3. **Low-impact**: Ascend and fix, note in current level's review
4. Re-descend and continue

**No circuit breaker**: Iterate until architecture is coherent or user interrupts.

---

## Reference Files

Before each stage, ensure you understand:
- Container patterns: `${CLAUDE_PLUGIN_ROOT}/references/container-patterns.md`
- Component templates: `${CLAUDE_PLUGIN_ROOT}/templates/component-*.md`
- Implementation guide: `${CLAUDE_PLUGIN_ROOT}/references/implementation-guide.md`
