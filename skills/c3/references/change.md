# Change Reference

Flow: `Understand → ADR → Provision Gate → Execute → Audit`

Spawn parallel subagents via Task tool for complex work.

## Progress Checklist

```
- [ ] Phase 1: topology loaded, request clarified, impact analyzed
- [ ] Phase 2: ADR created, user approves
- [ ] Phase 2b: provision gate (implement or design-only?)
- [ ] Phase 3: execute work breakdown
- [ ] Phase 4: audit + ADR marked implemented
```

---

## Phase 1: Understand

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Clarify with user (ASSUMPTION_MODE: skip). Analyze:
- Affected containers, components, refs
- For every file mentioned or discovered: `c3x lookup <file>` — load constraint chain before reasoning
- If lookup returns no mapping → file is uncharted territory, flag as coverage gap
- Read upward: component → container → context → cited refs
- Risks

Complex changes: spawn parallel analyst + reviewer subagents, synthesize.

## Phase 2: ADR

```bash
bash <skill-dir>/bin/c3x.sh add adr <slug>
```

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

Body: problem, decision, work breakdown, affected layers, risks.

Present for approval (ASSUMPTION_MODE: mark `[ASSUMED]`).

## Phase 2b: Provision Gate

Ask (ASSUMPTION_MODE: skip):
- **Implement now** → Phase 3
- **Design only** → create docs `status: provisioned`, no code-map entry, mark ADR `provisioned`, done

To implement provisioned later: invoke change, pick up ADR + docs, resume Phase 3.

## Phase 3: Execute

Scaffold:
```bash
bash <skill-dir>/bin/c3x.sh add container <slug>
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N [--feature]
bash <skill-dir>/bin/c3x.sh add ref <slug>
```

**REQUIRED before touching any file:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```
Returned refs = hard constraints. Every one must be honored. No exceptions.

Parallel subagents: decompose tasks, each reads component docs + refs before touching code.

Per task: verify code correct, docs updated (code-map.yaml, Related Refs), no regressions.

## Phase 4: Audit

```bash
bash <skill-dir>/bin/c3x.sh check
```

- Docs match code
- Related Refs updated
- CLAUDE.md blocks updated: `<!-- c3-generated: c3-NNN -->` ... `<!-- end-c3-generated -->`
- ADR → `implemented`

---

## Regression

| Discovery | Action |
|-----------|--------|
| Changes problem | Back to Phase 1 |
| Changes approach | Back to Phase 2 |
| Expands scope | Amend ADR |
| Implementation detail | Adjust tasks |

---

## ADR Format

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

Status: `proposed → accepted → (provisioned | implemented)`

---

## Routing

- Pre-change impact → sweep
- Architecture questions → query
- Pattern management → ref
- Standalone audit → audit
