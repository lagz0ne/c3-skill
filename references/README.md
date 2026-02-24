# C3 References Index

Shared reference documents used across skills, agents, and templates.

## Core References

| File | Purpose | Used By |
|------|---------|---------|
| `skill-harness.md` | Behavioral constraints, complexity rules, red flags | All skills |
| `layer-navigation.md` | How to traverse C3 docs (Context → Container → Component) | c3-query, c3-onboard skills |
| `audit-checks.md` | Full audit procedure with 10 phases | c3-audit skill |

## Content References

| File | Purpose | Used By |
|------|---------|---------|
| `component-categories.md` | Foundation vs Feature vs Ref categorization | c3-onboard, c3-ref, c3-audit |
| `onboard-ref-extraction.md` | Guidance for extracting refs during onboarding | c3-onboard skill |

## Workflow References

| File | Purpose | Used By |
|------|---------|---------|
| `adr-template.md` | ADR structure and lifecycle | Not bundled in skills (ADR format inlined in c3-change SKILL.md) |

## Loading References

Skills typically load references at startup:

```
## REQUIRED: Load References

Before proceeding, use Glob to find and Read these files:
1. `**/references/skill-harness.md` - Red flags and complexity rules
2. `**/references/layer-navigation.md` - How to traverse C3 docs
```

Additional references are loaded as needed for specific operations (ADR creation, auditing, etc.).
