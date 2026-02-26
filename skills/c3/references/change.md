# Change Reference

Orchestrate architectural changes through ADR-first workflow with parallel execution.

## How It Works

```
Understand -> ADR -> Provision Gate -> Execute -> Audit
```

Use Task tool to spawn anonymous subagents for parallel work (analysis, implementation, audit).

## Progress Checklist

```
Change Progress:
- [ ] Phase 1: Load topology, clarify request, analyze impact
- [ ] Phase 2: Create ADR, user approves
- [ ] Phase 2b: Provision gate (implement now or design only?)
- [ ] Phase 3: Execute work breakdown
- [ ] Phase 4: Audit (structural + semantic)
- [ ] ADR marked implemented
```

---

## Phase 1: Understand

1. Load topology:
```bash
bash <skill-dir>/bin/c3x.sh list --json
```

2. Clarify request with user (skip if ASSUMPTION_MODE)
3. Analyze impact:
   - Which containers, components, refs are affected?
   - What constraints apply? (read upward: component -> container -> context -> cited refs)
   - What are the risks?

**Subagent option:** For complex changes, spawn parallel analyst + reviewer subagents who debate impact. Synthesize findings.

## Phase 2: ADR

Create ADR:
```bash
bash <skill-dir>/bin/c3x.sh add adr <slug>
```

Fill ADR content:

```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
base-commit: [git hash]
affects: [c3-N, c3-NNN]
approved-files: []
---
```

ADR body includes:
- Problem statement
- Decision
- Work Breakdown (task list for Phase 3)
- Affected Layers table
- Risks and mitigations

Present to user for approval (skip if ASSUMPTION_MODE; mark `[ASSUMED]`).

## Phase 2b: Provision Gate

After ADR acceptance, ask user (skip if ASSUMPTION_MODE):

**"Implement now or design only?"**

- **Implement now** -> Continue to Phase 3
- **Design only (provision)** -> Create component docs with `status: provisioned`, no code-map entry (no code yet), mark ADR `provisioned`, done. Docs are visible to query and audit immediately.

To implement a provisioned design later: invoke change again. Pick up existing provisioned ADR + docs as starting context, resume from Phase 3.

## Phase 3: Execute

Create tasks from Work Breakdown in ADR.

**Scaffolding via CLI:**
```bash
# New container:
bash <skill-dir>/bin/c3x.sh add container <slug>

# New component:
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N --feature

# New ref:
bash <skill-dir>/bin/c3x.sh add ref <slug>
```

**Subagent option:** Decompose into independent tasks, spawn parallel implementer subagents. Each task points to C3 component docs and refs (mandatory reading for subagents).

For each completed task, verify:
- Code implemented correctly
- C3 docs updated (code-map.yaml, Related Refs)
- No regressions in other components

## Phase 4: Audit

1. Structural validation:
```bash
bash <skill-dir>/bin/c3x.sh check
```

2. Semantic review:
   - Docs match code changes
   - Related Refs updated
   - ADR Affected Layers accurate

3. CLAUDE.md propagation:
   - Update c3-generated blocks in relevant CLAUDE.md files
   - Format: `<!-- c3-generated: c3-NNN -->` ... `<!-- end-c3-generated -->`

4. Transition ADR to `implemented`

---

## Regression

Late discoveries during any phase:

| Discovery | Action |
|-----------|--------|
| Changes the problem | Back to Phase 1 |
| Changes the approach | Back to Phase 2 |
| Expands scope | Amend ADR |
| Implementation detail | Adjust tasks |

Surface discoveries to user. Confirm anything that affects the ADR.

---

## ADR Format

Always use YAML frontmatter (not markdown-style headers):

```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed | accepted | provisioned | implemented
date: YYYY-MM-DD
base-commit: [git hash]
affects: [c3-N, c3-NNN]
approved-files: []
---
```

Status lifecycle: proposed -> accepted -> (provisioned | implemented)

---

## Routing

During change, if needed:
- Impact assessment before starting -> sweep operation
- Architecture questions -> query operation
- Pattern management -> ref operation
- Standalone audit -> audit operation
