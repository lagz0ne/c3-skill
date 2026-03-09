# Change Reference

Flow: `ADR → Understand → Approve → Execute → Validate`

Spawn parallel subagents via Task tool for complex work.

## Progress

```
- [ ] Phase 1: ADR created (`c3x add adr <slug>`)
- [ ] Phase 2: Topology loaded, impact analyzed, ADR body filled
- [ ] Phase 2b: Provision gate (implement or design-only?)
- [ ] Phase 3: Execute work breakdown
- [ ] Phase 4: Validate + ADR marked implemented
```

---

## Phase 1: Create ADR

```bash
bash <skill-dir>/bin/c3x.sh add adr <slug>
```

This is the first action — before reading code or exploring files. The slug should capture the change intent (e.g., `add-rate-limiting`, `migrate-to-postgres`).

Edit the ADR frontmatter:
```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
affects: []
---
```

## Phase 2: Understand + Fill ADR

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Analyze affected entities. Use `c3x lookup <file>` on files you plan to modify — the returned refs are constraints that must be honored.

Read upward through entity docs: component → container → context → cited refs.

Fill the ADR body: Goal, Work Breakdown, Risks. Update `affects:` in frontmatter.

Present for approval (ASSUMPTION_MODE: mark `[ASSUMED]`).

## Phase 2b: Provision Gate

Ask (ASSUMPTION_MODE: auto-decide):
- **Implement now** → Phase 3
- **Design only** → create docs with `status: provisioned`, skip code-map entries, mark ADR `provisioned`, done

To implement provisioned work later: invoke change, pick up ADR + docs, resume Phase 3.

## Phase 3: Execute

Scaffold new entities:
```bash
bash <skill-dir>/bin/c3x.sh add container <slug>
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N [--feature]
bash <skill-dir>/bin/c3x.sh add ref <slug>
```

Use `c3x lookup <file>` before modifying files — returned refs are hard constraints.

**Code navigation: LSP first.** When exploring affected code, use LSP tools (go-to-definition, find-references) to precisely understand call chains and dependencies before making changes. Only fall back to Grep/Glob when LSP is unavailable.

For complex changes, decompose into parallel subagent tasks. Each task should read the relevant component docs and refs before touching code.

Per task: verify code, update docs (code-map.yaml, Related Refs), check for regressions.

## Phase 4: Validate

```bash
bash <skill-dir>/bin/c3x.sh check
```

- Docs match code
- Related Refs updated
- CLAUDE.md blocks updated if applicable
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

## ADR Lifecycle

ADRs are ephemeral work orders — they drive changes then become hidden.

Status: `proposed → accepted → (provisioned | implemented)`

`c3x list` and `c3x check` exclude ADRs by default. Use `--include-adr` to inspect.

---

## Routing

- Pre-change impact → sweep
- Architecture questions → query
- Pattern management → ref
- Standalone audit → audit
