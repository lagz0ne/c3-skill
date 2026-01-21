---
name: c3-audit-adr
description: |
  Audit an ADR for C3 principle violations before approval.
  Use when reviewing ADRs or when c3-alter needs pre-approval validation.
  Checks abstraction boundaries, composition rules, context alignment, and ref compliance.
---

# C3 Audit ADR

Validate that an ADR respects C3 architectural principles before approval.

## Usage

```
/c3 audit-adr <path-to-adr>
```

Or invoke via Task tool with `c3-skill:c3-adr-auditor` agent.

## When to Use

1. **Before ADR approval** - Gate the proposed → accepted transition
2. **PR review** - Verify ADR in a pull request
3. **Self-check** - Validate your own ADR before submitting

## What Gets Checked

### 1. Abstraction Boundaries

Components can only change what they own per their documented responsibilities.

**Example violation:**
```
ADR affects: c3-103 (task-routes)
Decision: "Add OAuth verification"
But c3-103 doc says: "Does NOT Own: Authentication"
→ FAIL: auth belongs to c3-102
```

### 2. Composition Rules

Components hand-off; they don't orchestrate. Container owns coordination.

**Example violation:**
```
ADR decision: "Task routes will coordinate requests between routes"
Container doc: "Components hand-off to each other; they do not orchestrate"
→ FAIL: orchestration is container's job
```

### 3. Context Alignment

Changes cannot contradict context-level Key Decisions without override.

**Example violation:**
```
Context Key Decisions: "JWT-based authentication"
ADR decision: "Replace JWT with sessions"
No Pattern Overrides section
→ FAIL: contradicts context without justification
```

### 4. Ref Compliance

If touching a pattern domain, the relevant ref must be addressed.

**Example:**
```
ADR affects auth component
.c3/refs/ref-auth-patterns.md exists
ADR doesn't cite ref or explain override
→ FAIL: must address established pattern
```

## Output

The auditor returns:
- **PASS** - ADR is ready for approval
- **FAIL** - ADR has violations that must be fixed

Each violation includes:
- Which principle was violated
- Quote from ADR
- Quote from C3 doc that contradicts
- Specific fix needed

## Dispatch

```yaml
Task:
  subagent_type: c3-skill:c3-adr-auditor
  prompt: "ADR Path: .c3/adr/adr-20260121-my-change.md"
```
