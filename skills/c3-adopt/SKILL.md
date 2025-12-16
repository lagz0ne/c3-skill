---
name: c3-adopt
description: Use when bootstrapping C3 documentation for any project - guides through Socratic discovery and delegates to layer skills for document creation
---

# C3 Adopt

## Overview

Bootstrap C3 architecture documentation through Socratic questioning and delegation. Works for existing codebases and fresh projects.

**Announce:** "I'm using the c3-adopt skill to initialize architecture documentation."

## When to Use

| Scenario | Path |
|----------|------|
| Existing codebase needs C3 docs | Explore → Question → Document |
| New project, no code yet | Question → Scaffold → Document |
| `.c3/` exists but needs rebuild | Ask: update, backup+recreate, or abort |

## Process

| Phase | Actions | Delegates To |
|-------|---------|--------------|
| 1. Detect | Check for `.c3/`, existing code | - |
| 2. Discover | Socratic questions per `references/discovery-questions.md` | - |
| 3. Context | Document containers, actors, protocols | c3-context-design |
| 4. Containers | For each container: components, tech, patterns | c3-container-design |
| 5. Components | For key components: flows, dependencies | c3-component-design |
| 6. Settings | Configure if missing | c3-config |

## Socratic Discovery

Use questions from `references/discovery-questions.md`:
- Start broad: "What problem does this system solve?"
- Identify containers: "What are the major deployable parts?"
- Map relationships: "How do containers communicate?"

**For existing codebases:** Explore first, then refine with questions.
**For fresh projects:** Questions only, no code exploration.

## Delegation Pattern

After discovery, delegate to layer skills in order:
1. c3-context-design → Creates c3-0 (Context)
2. c3-container-design → Creates c3-N for each container
3. c3-component-design → Creates c3-NNN for key components
4. c3-config → Creates settings.yaml if missing

## Scaffolding

Create structure before delegation:

```bash
mkdir -p .c3/adr
```

Create container folders: `.c3/c3-{N}-{slug}/`

Generate TOC after documentation is complete.

## Checklist

- [ ] Scenario detected (existing vs fresh)
- [ ] Discovery questions completed
- [ ] Context delegated and created
- [ ] Containers delegated and created
- [ ] Key components documented
- [ ] Settings configured
- [ ] TOC generated

## Platform Docs (Recommended)

Platform concerns (deployment, networking, secrets, CI/CD) reduce onboarding friction. Ask user if they want to document:

```
.c3/platform/
├── deployment.md      # c3-0-deployment
├── networking.md      # c3-0-networking
├── secrets.md         # c3-0-secrets
└── ci-cd.md           # c3-0-cicd
```

Use `c3-0-*` IDs for Context-level platform docs.

## Related

- `references/discovery-questions.md` - Socratic question templates
- `references/v3-structure.md` - File structure conventions
- `references/archetype-hints.md` - Container type patterns
- `references/derivation-guardrails.md` - Hierarchy rules
