# Subagent Exploration for Audit/Adopt

**Status:** In Progress (brainstorming)
**Date:** 2025-12-29

## Problem

The c3 agent's audit and adopt modes do heavy sequential exploration that:
1. Exhausts context (reading many files)
2. Takes too long (no parallelization)

## Goals

- **Speed:** Parallelize exploration with subagents
- **Context efficiency:** Keep main agent lean, subagents return summaries

## Scope

Two separate designs needed (operations don't happen together):
1. **Audit** - Health checks on existing .c3/ docs
2. **Adopt** - Initial codebase discovery for new C3 setup

## Current State

### Audit Phases (from audit-checks.md)

| Phase | Work | Parallelizable? |
|-------|------|-----------------|
| 1. Gather | Read context, list containers, read each container | Yes - per container |
| 2. Validate Structure | Check frontmatter, ID patterns, parents | Yes - per doc |
| 3. Cross-Reference | Inventory vs actual code | Yes - per container |
| 4. Validate Sections | Required sections per layer | Yes - per doc |
| 5. ADR Lifecycle | Check orphan ADRs | Sequential (small) |
| 6. Code Sampling | Sample major directories for drift | Yes - per directory |

### Adopt Phases (from adopt-workflow.md)

| Phase | Work | Parallelizable? |
|-------|------|-----------------|
| 1. Discovery | Explore codebase for boundaries, tech, integrations | Yes - per aspect |
| 2-7. Creation | Create structure, context, containers | Sequential (writes) |

## Design Questions (To Resume)

- [ ] Which mode to design first? (Audit vs Adopt)
- [ ] What should subagents return? (structured findings vs prose summaries)
- [ ] How to coordinate parallel results?
- [ ] Error handling when subagent fails?

## Next Steps

Resume brainstorming - pick Audit or Adopt to design first.
