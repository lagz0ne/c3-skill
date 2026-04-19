# Change Reference

Flow: `ADR → Understand → Approve → Execute → Audit`

Spawn parallel subagents via Task tool for complex work.

## Progress Checklist

```
- [ ] Phase 1: complete ADR created (`cat complete-adr.md | c3x add adr <slug>`)
- [ ] Phase 2: topology loaded, impact analyzed, ADR body complete work order
- [ ] Phase 2b: provision gate (implement or design-only?)
- [ ] Phase 3: execute work breakdown
- [ ] Phase 3a: contract cascade gate (top-down parent delta)
- [ ] Phase 3b: ref compliance gate
- [ ] Phase 4: audit + ADR marked implemented
```

---

## Phase 1: ADR (FIRST — non-negotiable)

```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add adr <slug>## Goal
<what + why>
EOF
```

Create ADR immediately with at least Goal. Slug = change intent (e.g., `add-rate-limiting`, `migrate-to-postgres`). Body expanded in Phase 2 via `c3x write` after understanding impact.

## Phase 2: Understand + Fill ADR

```bash
bash <skill-dir>/bin/c3x.sh list
```

Clarify with user (ASSUMPTION_MODE: skip). Analyze:
- Affected containers, components, refs
- Per file mentioned/discovered: `c3x lookup <file>` — load constraint chain before reasoning
- Lookup returns no mapping → uncharted territory, flag coverage gap
- `c3x read` upward: component → container → context → cited refs
- Risks

ADR body must have enough detail for later agent to recover decision without chat history. Update `affects:` in frontmatter when entities known; otherwise replace full ADR once context loaded.

CLI = source of truth for ADR structure:

```bash
bash <skill-dir>/bin/c3x.sh schema adr
bash <skill-dir>/bin/c3x.sh read <adr-id> --full
```

Follow `c3x schema adr` sections and `help[]` hints. `c3x add adr` = all-or-nothing: no thin ADR + incremental fill. For validator/schema/migration/command/ref/rule/derived-material changes, ADR must preserve underlay C3 changes: exact commands, validators, tests, help/hints, schemas, templates, verification evidence.

**Visual Impact:** Run `c3x graph <primary-affected-container-or-component> --format mermaid` — include in approval presentation. Multiple containers → graph each separately.

Present for approval (ASSUMPTION_MODE: mark `[ASSUMED]`).

Complex changes: spawn parallel analyst + reviewer subagents, synthesize.

## Phase 2b: Provision Gate

Ask (ASSUMPTION_MODE: skip):
- **Implement now** → Phase 3
- **Design only** → create docs `status: provisioned`, no code-map, mark ADR `provisioned`, done

To implement provisioned later: invoke change, pick up ADR + docs, resume Phase 3.

## Phase 3: Execute

Scaffold: `add` templates in onboard.md §1.2-1.4 (body via stdin). Delete: `c3x delete <id> [--dry-run]`.

**File context gate (SKILL.md §File Context) — MANDATORY before touching any file.**

Change-specific additions:
- Parallel subagents: each runs gate on their files before code changes
- Per task: verify code + docs updated (code-map, Related Refs/Rules), no regressions

## Phase 3a: Contract Cascade Gate

Every component delta must have parent delta decision.

Mandatory table before audit:

| Layer | Question | Verdict | Evidence |
|-------|----------|---------|----------|
| Component | Did Goal, Dependencies, Related Refs/Rules, Code References, or Container Connection change? | YES/NO | changed section/file |
| Container | Did Components, Responsibilities, Goal Contribution, or boundary change? | YES/NO | updated section or no-delta reason |
| Context | Did project/container topology change? | YES/NO | updated section or no-delta reason |
| Refs/Rules | Did shared constraints change? | YES/NO | updated entity or no-delta reason |

Rules:
- New component → parent container `## Components` MUST update before component done.
- Component goal/dependency/interface/status change → parent `Goal Contribution` + `Responsibilities` MUST be reviewed.
- `NO` requires evidence. Blank parent delta = STOP.
- ADR must record `Parent Delta: updated` or `Parent Delta: none` with evidence.

## Phase 3b: Ref Compliance Gate

**Before audit, verify changes comply with applicable refs.**

Per file touched in Phase 3:
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```

Per returned ref, check compliance by comparison mode:

| Ref Section | Mode | Check |
|-------------|------|-------|
| `## How` (code examples) | Structural | Code matches golden pattern structure? |
| `## How` (prose) | Semantic | Implementation follows described approach? |
| `## Choice` only | Negative | Code contradicts stated choice? |
| `## Not This` | Anti-pattern | Code resembles rejected alternative? |

**Rule Compliance (strict):** Per `rule-*` from lookup, compare code against `## Golden Example` + `## Not This` loaded in Phase 3 step 2. No re-read — constraints already in context. Strict enforcement: must match golden pattern. Flag deviation as violation.

**ADVERSARIAL FRAMING: Look for violations — never confirm compliance.**

Mandatory output:

```
| Ref | Section Checked | Verdict | Evidence |
|-----|-----------------|---------|----------|
| ref-X | How | COMPLIANT | Matches pattern structure |
| ref-Y | Not This | VIOLATION | Uses rejected approach Z |
```

Rules:
- Scope to YOUR CHANGES — no full codebase audit
- Ref wins — code disagrees with ref → ref is right. Override needs ADR.
- Override via `## Override` — follow ref's documented process
- Conflicts — scope specificity wins (component ref > container ref > context ref)

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

Status: `proposed → accepted → (provisioned | implemented)`. ADRs hidden by default; `--include-adr` to inspect.
