# C3 References Index

Shared reference documents used across skills, agents, and templates.

## Core References

| File | Purpose | Used By |
|------|---------|---------|
| `skill-harness.md` | Behavioral constraints, complexity rules, red flags | All skills |
| `layer-navigation.md` | How to traverse C3 docs (Context → Container → Component) | Query, Alter skills |
| `content-separation.md` | Canonical definition for component vs ref content separation | Onboard, Audit, Classifier |
| `audit-checks.md` | Full audit procedure with 9 phases | c3 skill (Audit mode) |

## Workflow References

| File | Purpose | Used By |
|------|---------|---------|
| `adr-template.md` | ADR structure and lifecycle | Alter skill, Orchestrator agent |
| `plan-template.md` | Implementation plan structure | Alter skill |
| `implementation-guide.md` | Code implementation patterns | Orchestrator agent |

## Type References

| File | Purpose | Used By |
|------|---------|---------|
| `component-types.md` | Foundation vs Feature categorization, external types | Onboard skill |
| `container-patterns.md` | Container component patterns (internal, linkage, adapter) | Onboard skill |

## Structure Reference

| File | Purpose | Used By |
|------|---------|---------|
| `v3-structure.md` | C3 v3 directory structure specification | Layer navigation |

## Loading References

Skills typically load references at startup:

```
## REQUIRED: Load References

Before proceeding, use Glob to find and Read these files:
1. `**/references/skill-harness.md` - Red flags and complexity rules
2. `**/references/layer-navigation.md` - How to traverse C3 docs
```

Additional references are loaded as needed for specific operations (ADR creation, auditing, etc.).
