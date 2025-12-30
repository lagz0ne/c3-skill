# Subagent Exploration for Audit/Adopt

**Status:** In Progress (brainstorming)
**Date:** 2025-12-29
**Priority:** Adopt first

## Problem

The c3 agent's audit and adopt modes do heavy sequential exploration that:
1. Exhausts context (reading many files)
2. Takes too long (no parallelization)

## Goals

- **Speed:** Parallelize exploration with subagents
- **Context efficiency:** Keep main agent lean, subagents return summaries
- **User efficiency:** Use AskUserQuestion for structured input (save typing)

---

## Adopt Design (Priority)

### Two-Stage Flow

```
Stage 1: Auto-Discovery
├── Dispatch subagents to explore codebase in parallel
├── Each returns structured findings (not raw files)
└── Main agent synthesizes into draft inventory

Stage 2: Socratic Clarification
├── Present findings via AskUserQuestion (multiple choice)
├── Confirm/correct container boundaries
├── Confirm/correct component inventory per container
└── Capture domain-specific naming/purpose
```

### Stage 1: Auto-Discovery (Parallel Subagents)

| Subagent | Explores | Returns |
|----------|----------|---------|
| **entry-points** | package.json scripts, main files, Dockerfiles | List of runnable entry points |
| **boundaries** | Top-level dirs, package structure, workspaces | Candidate container boundaries |
| **tech-stack** | Dependencies, configs, framework patterns | Tech per boundary |
| **external-integrations** | API calls, DB connections, env vars | External actors/services |

**Output format:** Structured JSON/YAML that main agent can process without reading raw files.

### Stage 2: Socratic via AskUserQuestion

**Inventory-first approach:** Build comprehensive questions from discovery.

**Question flow:**

1. **Container confirmation:**
   ```
   Q: "Based on discovery, these look like separate containers. Confirm?"
   Options:
   - [x] backend/ → Backend API
   - [x] frontend/ → Web App
   - [ ] scripts/ → (exclude, not a container)
   - Other: [specify]
   ```

2. **Per-container component inventory:**
   ```
   Q: "For Backend API, which are significant components?"
   Options (multi-select):
   - [x] src/auth/ → Authentication
   - [x] src/api/ → API Routes
   - [x] src/db/ → Database Layer
   - [ ] src/utils/ → (exclude, utility)
   ```

3. **External actors:**
   ```
   Q: "Which external systems does this interact with?"
   Options:
   - [x] PostgreSQL (detected from pg dependency)
   - [x] Stripe API (detected from stripe calls)
   - [ ] Redis (not detected, add if needed)
   ```

4. **Naming/purpose refinement:**
   ```
   Q: "What's the primary purpose of this system?"
   Options:
   - E-commerce platform
   - SaaS application
   - Internal tool
   - Other: [specify]
   ```

### Key Principles

1. **Discovery informs questions** - Don't ask what can be detected
2. **Confirmation over description** - "Is this right?" not "Describe your system"
3. **Multi-select for inventory** - Faster than individual yes/no
4. **Escape hatch** - Always allow "Other" for corrections

---

## Audit Design (Later)

To be designed after Adopt is implemented.

### Audit Phases (from audit-checks.md)

| Phase | Work | Parallelizable? |
|-------|------|-----------------|
| 1. Gather | Read context, list containers, read each container | Yes - per container |
| 2. Validate Structure | Check frontmatter, ID patterns, parents | Yes - per doc |
| 3. Cross-Reference | Inventory vs actual code | Yes - per container |
| 4. Validate Sections | Required sections per layer | Yes - per doc |
| 5. ADR Lifecycle | Check orphan ADRs | Sequential (small) |
| 6. Code Sampling | Sample major directories for drift | Yes - per directory |

---

## Open Questions

- [ ] What model for discovery subagents? (haiku for speed?)
- [ ] How to handle monorepo vs single-app detection?
- [ ] Max questions before user fatigue?
- [ ] Fallback if discovery finds nothing?

## Next Steps

1. Finalize Adopt Stage 1 subagent specs
2. Define structured output format for each subagent
3. Design Stage 2 question flow in detail
4. Implement and test
