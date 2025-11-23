# Socratic Method for C3

Question-driven approach for building understanding during C3 documentation.

## Core Principle

**Ask questions to understand, don't scan to assume.**

| Scanning | Questioning |
|----------|-------------|
| Finds files | Finds purpose |
| Shows structure | Shows intent |
| Detects patterns | Discovers context |
| Misses relationships | Builds understanding |
| Can be wrong | User validates |

## Question Flow Pattern

```
Open Question → Listen → Reflect Back → Clarify → Confirm
```

**Example:**
1. "What containers exist?" (open)
2. Listen: "We have a Next.js app and a database"
3. Reflect: "So you have a fullstack Next.js container and PostgreSQL"
4. Clarify: "Is the database managed or self-hosted?"
5. Confirm: "Got it - Next.js frontend+API with managed PostgreSQL"

## When User Doesn't Know

If user is uncertain:

> "That's fine - let's mark this as TBD and move on. We can fill it in later."

Create placeholder:
```markdown
## [Section] {#id}
<!--
TODO: Needs clarification
-->

[To be documented]
```

## Question Categories by Level

### Context Level

**System Boundary:**
- "What is inside your system vs what is external?"
- "Who or what interacts with your system from outside?"

**Actors:**
- "What types of users interact with the system?"
- "Are there other systems that call into yours?"

**Containers:**
- "If you deployed this today, what separate processes would run?"
- "What data stores exist?"

**Protocols:**
- "How do these pieces talk to each other?"
- "What external services does your system call?"

### Container Level

**Identity:**
- "What is the single responsibility of this container?"
- "If this container disappeared, what would break?"

**Technology:**
- "What language and framework does this use?"
- "Why was this technology chosen?"

**Structure:**
- "How is code organized inside?"
- "What are the main entry points?"

**APIs:**
- "What endpoints does this container expose?"
- "What APIs does it consume?"

### Component Level

**Purpose:**
- "What specific problem does this component solve?"
- "What's the single most important thing this does?"

**Implementation:**
- "What library/framework does this use?"
- "What's the core algorithm or pattern?"

**Configuration:**
- "What environment variables does this need?"
- "What must change between dev and prod?"

**Error Handling:**
- "What can go wrong?"
- "What errors are retriable vs fatal?"

## During c3-design Exploration

Socratic questions **confirm understanding**, not **discover location**:

**Good:** "Based on c3-1-backend, the auth middleware handles tokens. Is that still accurate?"

**Bad:** "Where is authentication handled?" (should form hypothesis from TOC first)

## Building Models from Answers

After questions, construct a model before delegating:

```markdown
## Context Understanding

**System:** [Name]
**Purpose:** [One sentence]

**Actors:**
- [Actor 1]
- [Actor 2]

**Containers:**
| Name | Type | Purpose |
|------|------|---------|
| [Name] | [Backend/Frontend/DB] | [What it does] |
```

Then delegate to appropriate layer skill with this understanding.

## Anti-Patterns

| Anti-Pattern | Why Wrong | Do Instead |
|--------------|-----------|------------|
| Scanning codebase first | Misses intent, wastes context | Ask about purpose |
| Multiple questions at once | Overwhelming, unclear answers | One question at a time |
| Leading questions | Confirms bias | Open questions |
| Skipping reflection | Misunderstandings propagate | Always reflect back |
