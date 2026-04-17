# Change Reference

Flow: `ADR → Understand → Approve → Execute → Audit`

Spawn parallel subagents via Task tool for complex work.

## Progress Checklist

```
- [ ] Phase 1: complete ADR created (`cat complete-adr.md | c3x add adr <slug>`)
- [ ] Phase 2: topology loaded, impact analyzed, ADR body reviewed or replaced as a complete work order
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
<what this change achieves and why>
EOF
```

Create the ADR immediately with at least a Goal. The slug should capture the change intent (e.g., `add-rate-limiting`, `migrate-to-postgres`). The body will be expanded in Phase 2 via `c3x write` after understanding impact.

## Phase 2: Understand + Fill ADR

```bash
bash <skill-dir>/bin/c3x.sh list```

Clarify with user (ASSUMPTION_MODE: skip). Analyze:
- Affected containers, components, refs
- For every file mentioned or discovered: `c3x lookup <file>` — load constraint chain before reasoning
- If lookup returns no mapping → file is uncharted territory, flag as coverage gap
- `c3x read` upward: component → container → context → cited refs
- Risks

Create the ADR body with enough detail that a later agent can recover the decision without chat history. Update `affects:` in frontmatter before the ADR is added when the affected entities are known; otherwise replace the full ADR once context is loaded.

The CLI is the source of truth for ADR structure and next steps:

```bash
bash <skill-dir>/bin/c3x.sh schema adr
bash <skill-dir>/bin/c3x.sh read <adr-id> --full
```

Follow the `c3x schema adr` sections and `help[]` hints. `c3x add adr` is all-or-nothing: do not create a thin ADR and fill it incrementally. For validator, schema, migration, command workflow, ref, rule, or derived-material changes, the ADR must preserve the underlay C3 changes: exact commands, validators, tests, help/hints, schemas, templates, and verification evidence that enforce the decision.

**Visual Impact:** Run `c3x graph <primary-affected-container-or-component> --format mermaid` and include in the approval presentation as a mermaid code block. If change spans multiple containers, graph each affected container separately.

Present for approval (ASSUMPTION_MODE: mark `[ASSUMED]`).

Complex changes: spawn parallel analyst + reviewer subagents, synthesize.

## Phase 2b: Provision Gate

Ask (ASSUMPTION_MODE: skip):
- **Implement now** → Phase 3
- **Design only** → create docs `status: provisioned`, no code-map entry, mark ADR `provisioned`, done

To implement provisioned later: invoke change, pick up ADR + docs, resume Phase 3.

## Phase 3: Execute

Scaffold / tear down (all `add` commands require body via stdin):
```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add container <slug>## Goal
...
## Components
...
## Responsibilities
...
EOF

cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N [--feature]## Goal
...
## Dependencies
...
EOF

cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add ref <slug>## Goal
...
## Choice
...
## Why
...
EOF

cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add rule <slug>## Goal
...
## Rule
...
## Golden Example
...
EOF

bash <skill-dir>/bin/c3x.sh delete <id> [--dry-run]
```

**REQUIRED before touching any file — the 3-step context gate:**

**1. Lookup:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```

**2. Load rules:** For every `rule-*` in the lookup response:
```bash
bash <skill-dir>/bin/c3x.sh read <rule-id>
```
Extract `## Rule`, `## Golden Example`, and `## Not This`. Hold these in working context while editing. Code MUST match the golden pattern — deviations require the rule's `## Override` process or a new ADR.

**3. Top-down parent context:** For each matched component, read the parent container before component detail:
```bash
bash <skill-dir>/bin/c3x.sh read <parent-container-id>
bash <skill-dir>/bin/c3x.sh read <component-id>
```
For a new component, this order is mandatory: container responsibility and membership first, component detail second.

**4. Graph blast radius:** Graph the parent container first, then each matched component:
```bash
bash <skill-dir>/bin/c3x.sh graph <parent-container-id> --depth 1
bash <skill-dir>/bin/c3x.sh graph <component-id> --depth 1
```
If changes affect the component's interface, check consumers before proceeding.

Returned refs + loaded rule content + parent container + graph = hard constraints. No exceptions.

Parallel subagents: decompose tasks, each runs the context gate on their files before touching code.

Per task: verify code correct, docs updated (code-map entries, Related Refs/Rules), no regressions.

## Phase 3a: Contract Cascade Gate

Every component delta must have a parent delta decision.

Mandatory table before audit:

| Layer | Question | Verdict | Evidence |
|-------|----------|---------|----------|
| Component | Did Goal, Dependencies, Related Refs/Rules, Code References, or Container Connection change? | YES/NO | changed section/file |
| Container | Did Components, Responsibilities, Goal Contribution, or boundary change? | YES/NO | updated section or no-delta reason |
| Context | Did project/container topology change? | YES/NO | updated section or no-delta reason |
| Refs/Rules | Did shared constraints change? | YES/NO | updated entity or no-delta reason |

Rules:
- New component → parent container `## Components` MUST update before component detail is considered done.
- Component goal/dependency/interface/status change → parent `Goal Contribution` and `Responsibilities` MUST be reviewed.
- `NO` requires evidence. Blank parent delta = STOP.
- ADR must record `Parent Delta: updated` or `Parent Delta: none` with evidence.

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

**Rule Compliance (strict):** For each `rule-*` entity returned by lookup, compare code against the `## Golden Example` and `## Not This` content loaded in Phase 3 step 2. Do NOT re-read — the constraints are already in context. Rules use strict enforcement — code must match the golden pattern. Flag any deviation as a violation.

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
- Coding standards → rule
- Standalone audit → audit
