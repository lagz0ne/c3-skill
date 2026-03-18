# Change Reference

Flow: `ADR → Understand → Approve → Execute → Audit`

Spawn parallel subagents via Task tool for complex work.

## Progress Checklist

```
- [ ] Phase 1: ADR created (`c3x add adr <slug>`)
- [ ] Phase 2: topology loaded, impact analyzed, ADR body filled
- [ ] Phase 2b: provision gate (implement or design-only?)
- [ ] Phase 3: execute work breakdown
- [ ] Phase 3b: ref compliance gate
- [ ] Phase 4: audit + ADR marked implemented
```

---

## Phase 1: ADR (FIRST — non-negotiable)

```bash
bash <skill-dir>/bin/c3x.sh add adr <slug>
```

Create the ADR immediately. The slug should capture the change intent (e.g., `add-rate-limiting`, `migrate-to-postgres`).

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

The body will be filled in Phase 2 after understanding impact.

## Phase 2: Understand + Fill ADR

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Clarify with user (ASSUMPTION_MODE: skip). Analyze:
- Affected containers, components, refs
- For every file mentioned or discovered: `c3x lookup <file>` — load constraint chain before reasoning
- If lookup returns no mapping → file is uncharted territory, flag as coverage gap
- Read upward: component → container → context → cited refs
- Risks

Fill the ADR body: Goal, Work Breakdown, Risks. Update `affects:` in frontmatter.

Present for approval (ASSUMPTION_MODE: mark `[ASSUMED]`).

Complex changes: spawn parallel analyst + reviewer subagents, synthesize.

## Phase 2b: Provision Gate

Ask (ASSUMPTION_MODE: skip):
- **Implement now** → Phase 3
- **Design only** → create docs `status: provisioned`, no code-map entry, mark ADR `provisioned`, done

To implement provisioned later: invoke change, pick up ADR + docs, resume Phase 3.

## Phase 3: Execute

Scaffold / tear down:
```bash
bash <skill-dir>/bin/c3x.sh add container <slug>
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N [--feature]
bash <skill-dir>/bin/c3x.sh add ref <slug>
bash <skill-dir>/bin/c3x.sh delete <id> [--dry-run]
```

**REQUIRED before touching any file:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```
Returned refs = hard constraints. Every one must be honored. No exceptions.

Parallel subagents: decompose tasks, each reads component docs + refs before touching code.

Per task: verify code correct, docs updated (code-map.yaml, Related Refs), no regressions.

## Phase 3b: Ref Compliance Gate

**Before moving to audit, verify changes comply with applicable refs.**

For each file touched in Phase 3:
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```

For each returned ref, check compliance using comparison mode:

| Ref Section | Comparison Mode | What To Check |
|-------------|-----------------|---------------|
| `## How` (code examples) | Structural | Does code match the golden pattern structure? |
| `## How` (prose) | Semantic | Does implementation follow the described approach? |
| `## Choice` only | Negative | Does code contradict the stated choice? |
| `## Not This` | Anti-pattern | Does code resemble any rejected alternative? |

**ADVERSARIAL FRAMING: Look for violations — do not confirm compliance.**

Mandatory output:

```
| Ref | Section Checked | Verdict | Evidence |
|-----|-----------------|---------|----------|
| ref-X | How | COMPLIANT | Matches pattern structure |
| ref-Y | Not This | VIOLATION | Uses rejected approach Z |
```

Rules:
- **Scope to YOUR CHANGES** — don't audit the entire codebase
- **Ref wins** — if your code disagrees with a ref, the ref is right. Create an ADR if override needed.
- **Override via `## Override`** — follow the ref's documented override process
- **Conflicts** — when multiple refs apply, scope specificity wins (component ref > container ref > context ref)

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

## ADR Lifecycle

ADRs are **ephemeral work orders**. They drive changes then become hidden.

Status: `proposed → accepted → (provisioned | implemented)`

`c3x list` and `c3x check` exclude ADRs by default. Use `--include-adr` to inspect.

---

## Routing

- Pre-change impact → sweep
- Architecture questions → query
- Pattern management → ref
- Standalone audit → audit
