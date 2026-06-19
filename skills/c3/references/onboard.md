# Onboard Reference

## What onboarding is (v11)

LLM-driven, one complete change-unit cycle. You **discuss** the architecture, **size the canvas to fit the project**, **author the whole architecture into the genesis ADR** as staged create-patches, then **flip** — `c3x change apply` materializes it atomically as frozen facts. `c3 init` already created the system `c3-0` and the genesis ADR `adr-00000000-c3-adoption`; the staged create-patches **persist on disk** (apply does not delete them — `check` exempts `.c3/changes/`), so authoring is interruptible and resumable. After the flip the ADR's prose is the durable narrative; remove `.c3/changes/adr-00000000-c3-adoption/` once applied if you want to clear the staging.

**The order matters: canvas first, then facts.** Every fact is validated against its canvas at apply, so the canvas must already be the right shape before you author facts.

### Laddering — integrity first, always complete

A canvas is a **rung**: a complete contract for one complexity *level*. A fresh `c3 init` seeds **lean rung-1** canvases — the component canvas requires 6 sections (Goal, Parent Fit, Purpose, Governance, Contract, Derived Materials); the deeper Foundational Flow, Business Flow, and Change Safety are optional, a higher rung. What grows over a project's life is the **complexity level, not the completeness**: a fact is *always* complete to its canvas, never thin or "filled in later."

Climbing a rung is deliberate: **raise the canvas** (make a section required, or author a richer one with `c3 canvas write`) and **migrate every affected fact up to it, completely**. Each rung is solid on its own and is **not** responsible for future rungs — don't pre-build sections a complex project would need later. Size to the project you have now; the lean seeds are the default. Trim or enrich a canvas only if this project warrants it. See `canvas.md` for the climb mechanism.

## Precondition

`c3 list` returns facts → already onboarded. `AskUserQuestion`: re-onboard or cancel (skip if ASSUMPTION_MODE). Cancel → suggest audit/query.

## Component Categories

| Can name concrete file? | Category |
|------------------------|---------|
| Yes | Foundation (01-09) or Feature (10+) |
| No (rules only) | **Ref** — code-map entry optional |

Foundation: infra others depend on. Feature: biz logic. Ref: conventions/shared utils. Rule: coding standards/constraints. Refs with concrete files (shared middleware, util libs) → code-map entries; pure-convention refs and rules → empty.

## The v11 flow

The genesis ADR is the spine. Stage 0/1/2 below are the discovery → detail → finalize scaffolding; what changes in v11 is that **creation routes through the genesis ADR's create-patches and one flip**, not only ad-hoc `c3 add`.

1. **Discuss** — the idea and the architectural separation: containers, where the seams fall. Conversation first.
2. **Size the canvas** — keep or shape the rung-1 canvases to match the project (the lean seeds are the default; trim or enrich via `c3 canvas write`). This happens **before** authoring facts, because facts are validated against the canvas.
3. **Author** — write the architecture into the genesis ADR `adr-00000000-c3-adoption`. The containers/components/refs/rules go in as **create-patches** in `.c3/changes/adr-00000000-c3-adoption/` (scope `whole`, no base, with `type:` and `parent:`); the ADR body carries the narrative. Interruptible — the staged patches persist.
4. **Flip** — `c3x change apply adr-00000000-c3-adoption` materializes the whole architecture atomically as frozen facts (canvas-validated, all-or-nothing). The genesis ADR's Affected Topology Evidence was authored as `N.A` (the facts didn't exist yet), so **after the flip refresh those cites with real handles** (`c3 read <id> --cite` per affected fact) — now the facts exist. Then `c3 change accept adr-00000000-c3-adoption` + `c3 check --fix` latches it `accepted → done` once those refreshed After-cites resolve. Onboarding ends having completed one full change-unit cycle.

Direct `c3 add` remains valid and unguarded for creating a fact (a new fact is not frozen). Frame the genesis ADR as the demonstration **and** the record of how this architecture was created (the ADR prose is the durable narrative; the create-patches stay staged on disk after the flip until you clear them).

## Progress Checklist

```
- [ ] Stage 0: inventory complete, genesis ADR tables filled
- [ ] Gate 0: proceed to Details
- [ ] Stage 1: rung-1 canvases sized; all container/component/ref create-patches authored in .c3/changes/adr-00000000-c3-adoption/
- [ ] Gate 1: no new items discovered
- [ ] Stage 2: flip applied (facts materialized), code-map filled, integrity + audit pass
- [ ] Gate 2: genesis ADR accepted, then latched to `done` via `c3 check --fix`
```

---

## Stage 0: Inventory

### 0.1 Scaffold

`c3 init` already scaffolded `.c3/` (config, README, canvases/, system `c3-0`, genesis ADR `adr-00000000-c3-adoption`). The genesis ADR starts **bodyless**, so author the **whole body once** with `c3 write adr-00000000-c3-adoption --file <path>` first (`--section` can't create a section that doesn't exist yet — it edits existing ones). After the body exists, revise per section via `c3 write adr-00000000-c3-adoption --section <name> --file <path>` (for tables, mermaid, code blocks) or `echo "..." | c3 write adr-00000000-c3-adoption --section <name>` (short text); frontmatter fields via `c3 set adr-00000000-c3-adoption <field> <value>`. The ADR is a change-doc, not a frozen fact, so you author and revise it freely.

### 0.1b Size the canvas first

Before authoring any fact, **size the canvas to the project**. Read what each rung-1 canvas requires (`c3 schema component`, `c3 canvas list`) and keep, trim, or enrich it with `c3 canvas write <type> --file` — but only if this project warrants it; the lean seeds are the default. Do this now: facts are validated against the canvas at apply, so the shape must be right before facts exist. Don't pre-build higher-rung sections — climb later, deliberately (see `canvas.md`).

### 0.2 Context Discovery

Capture in the genesis ADR:

| Arg | Value |
|-----|-------|
| PROJECT | System name |
| GOAL | Why it exists |
| SUMMARY | One sentence |

Find **Abstract Constraints** — system-level non-negotiables.

`AskUserQuestion` for gaps (ASSUMPTION_MODE: assume, mark `[ASSUMED]`).

### 0.3 Container Discovery

Container = deployment/runtime boundary.

| N | CONTAINER_NAME | BOUNDARY | GOAL | SUMMARY |
|---|----------------|----------|------|---------|

### 0.4 Component Discovery

| N | NN | COMPONENT_NAME | CATEGORY | GOAL | SUMMARY |
|---|----|----|----------|------|---------|

Foundation (01-09): others depend on. Feature (10+): biz logic.

### 0.5 Ref Discovery

Patterns repeating across components:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common: error handling, form patterns, data fetching, design system. Each ref requires Choice + Why minimum.

### 0.6 Rule Discovery

Project-wide coding standards/constraints:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common: naming conventions, forbidden patterns, lint rules, security constraints. Look for repeated review feedback, linter configs, "always/never" statements.

### 0.6b Recipe Discovery — cross-container flows

A **recipe** captures an end-to-end operation that **no single component owns** — it
crosses containers and hands off between components. A component's own `Business Flow`
is its *local* slice; the recipe is the *whole* path. Whenever a feature spans the
frontend→backend→integration→database seam, trace it as a recipe.

| SLUG | TITLE | GOAL (the path it traces) |
|------|-------|---------------------------|

```bash
# body: ## Goal naming the ordered hand-offs (which component/container owns each step)
c3 add recipe <slug> --file body.md
```

Look for the multi-step operations the system is *for*: e.g. reservation→pick→pack→ship,
receiving→putaway, return-processing, cycle-count→adjustment. At least the two or three
most important cross-container operations earn a recipe — it is the artifact that makes
the architecture's behavior legible across the container boundaries.

### 0.7 Overview Diagram

Per container:
```bash
c3 graph <container-id> --format mermaid
```
Include each as mermaid code block.

### Gate 0

- [ ] Context args filled (PROJECT, GOAL, SUMMARY)
- [ ] Abstract Constraints identified
- [ ] All containers with args (including BOUNDARY)
- [ ] All components (brief) with category
- [ ] Cross-cutting refs (Choice + Why minimum)
- [ ] Coding standards as rules
- [ ] Overview diagram

---

## Stage 1: Details

**Route creation through the genesis ADR.** The container/component/ref/rule bodies below are authored as **create-patches** staged in `.c3/changes/adr-00000000-c3-adoption/` — one `<seq>-<slug>.patch.md` per fact: scope `whole`, **no base**, with `type:` and `parent:` in the frontmatter, and the body in the shapes shown below. Nothing materializes yet; the whole architecture lands in one flip at Stage 2. This keeps onboarding interruptible (the staged patches persist on disk) and makes the genesis ADR the record of how the architecture was built. (The `--file` body shapes below are the canvas-correct content for each patch body; direct `c3 add` is still valid for one-off facts, but the genesis ADR is the demonstration and the ledger.)

The system `c3-0` already exists from `c3 init` — its body is the context doc you author directly (it is a fact created at init; author it before the flip, then it is frozen).

### 1.1 Context Doc

Short text fields: `echo "<goal>" | c3 write c3-0 --section Goal`. Whole body rewrite: `c3 write c3-0 --file content.md`.

**c3-0's `Containers` table is tool-maintained — author it as a header, not rows.**
c3-0 is born bodyless, so it stays directly writable until you author its body (the creation
window); editing it after closes. Author c3-0's body with the `## Containers` **header row only**
(its canvas columns from `c3 schema system`; no data rows) — the flip fills it from each
container's `parent: c3-0`. Same for each container's `## Components` table. No placeholders, no
ordering churn, no re-patch: declare the parent edges and every membership table materializes
itself at apply.

### 1.2 Container Docs

**Create container** (body in a file — tables and mermaid require `--file`):
```bash
# body.md contains: ## Goal / ## Components (table) / ## Responsibilities
c3 add container <slug> --file body.md
```

**Create components** (body in a file):
```bash
# Foundation (01-09):
c3 add component <slug> --container c3-N --file body.md

# Feature (10+): add --feature flag
c3 add component <slug> --container c3-N --feature --file body.md
```
Author each component body to the **component canvas** (`c3 schema component` — its
required sections, not a remembered list). Any content with markdown tables, mermaid, or
code fences MUST go through `--file <path>` — inline strings corrupt quoting.

**In the genesis flow YOU choose the ids — a create-patch's `target:` IS the entity
id.** There is no CLI auto-numbering here (that's `c3 add`'s behavior, a separate path):
the flip materializes each create-patch as a fact with exactly the id you wrote. So pick the
ids up front, following the convention — container `c3-1`, `c3-2`, …; that container's
components `c3-101+` (foundation) / `c3-110+` (feature); refs/rules by slug (`ref-…`, `rule-…`).
Set each fact's `parent:` and leave its parent's membership table a **header only** — the flip
synthesizes every `Components` / `Containers` row from the `parent:` edges, in the same pass.
**No placeholders, no "create then read", no membership patch.** (Prefer the numbered convention
over slug ids like `web`/`api`: slug-as-id breaks the `c3-N` reference scheme and mangles on-disk
filenames.)

Code-map patterns: `c3 set <id> codemap <pattern>`. Bracket paths (`[id]`, `[...slug]`) work automatically.

**Extract Refs:** "Would this change if we swapped underlying tech?" Yes → ref.

**Extract Rules:** "Coding standard or constraint, not pattern choice?" Yes → rule.

| Signal | Action |
|--------|--------|
| "We use X with..." | ref-X |
| "Our convention is..." | new/existing ref |
| Same pattern in 2+ components | create ref, cite both |
| "We always/never do X" | rule |
| Lint rule, naming, security | rule |

### 1.3 Ref Docs

```bash
# body.md: ## Goal / ## Choice / ## Why (+ optional How, Scope, Not This, Override)
c3 add ref <slug> --file body.md
```

### 1.4 Rule Docs

```bash
# body.md: ## Goal / ## Rule / ## Golden Example (code fence) (+ optional Not This, Scope, Override)
c3 add rule <slug> --file body.md
```
Golden Example contains code fences -> `--file` is mandatory.

### Gate 1

- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] All rules documented
- [ ] No new items (else update the genesis ADR, return Stage 0)

---

## Stage 2: Finalize

### 2.0 Flip — materialize the architecture

Land the whole architecture in one atomic, canvas-validated transaction:

```bash
c3x change view adr-00000000-c3-adoption    # preview every staged create-patch
c3x change apply adr-00000000-c3-adoption   # all-or-nothing: the facts are now frozen
```

Apply is all-or-nothing — every fact validates against its canvas, or nothing lands. After the flip the containers/components/refs/rules exist as frozen facts; from here, editing any of them goes through a change-unit (see `change.md`), not `c3 add`/`c3 write`.

### 2.1 Code-Map

Set glob patterns per component/ref/rule (`c3 set <id> codemap` is the one `set` allowed on a frozen fact — code churns, so the map is maintained live):
```bash
c3 set <id> codemap '<glob>'
c3 lookup 'src/**'   # spot-check mapping
```

### 2.2 Validate

```bash
c3 check
c3 list              # coverage + counts
```

### 2.3 Semantic

| Check | Verify |
|-------|--------|
| Context ↔ Container | genesis ADR containers match the materialized facts |
| Container ↔ Component | Each component in container README has doc |
| * ↔ Refs | Citations match the component reference column (Governance) |

### 2.4 Audit + close the cycle

Run audit. Pass → close the genesis ADR through the canonical lifecycle (`[open, accepted, done, superseded]`):

```bash
c3 change accept adr-00000000-c3-adoption   # human judgment: → accepted
c3 check --fix                              # auto-done latch: accepted → done when After-cites resolve
```

`done` is **earned, never typed**: the auto-done latch actualizes `accepted → done` at `c3 check --fix` once the genesis ADR's After-cites resolve fresh — proof the architecture actually landed. (There is no `implemented` status in v11; the canonical terminal state is `done`.)

### Gate 2

- [ ] Flip applied — facts materialized (`change apply`)
- [ ] Code-map scaffolded + patterns filled
- [ ] Coverage % acceptable (or exclusions documented)
- [ ] Integrity checks pass
- [ ] Audit passes
- [ ] Genesis ADR accepted, then latched to `done` via `c3 check --fix`

Issues → Inventory (Gate 0) or Detail (Gate 1).

---

## Final Checks

```bash
c3 list
c3 check
c3 lookup <any-mapped-file>
c3 lookup 'src/**'
```

**Fix before completing:**

| Signal | Problem | Fix |
|--------|---------|-----|
| No system goal | Missing `goal:` in README.md | `c3 set <id> goal "<text>"` |
| No `files:` | Missing code-map pattern | `c3 set <id> codemap '<glob>'` |
| No `uses:` | Ref not wired | Add the ref to the component's reference column (`c3 schema component` — today `Governance`) — at create, or via a change-unit patch if the component already exists — then re-import (edges build at import) |
| Ref has no `via:` | Uncited ref | Cite it from a component (reference column, as above) or delete the ref |
| `[provisioning]` | Design-only | Expected or implement |
| `lookup` returns nothing | Bad/missing codemap | Fix patterns via `c3 set <id> codemap '<glob>'`; re-check with `lookup 'src/**'` |
| Low coverage % | Many unmapped files | `_exclude` for tests/configs, map rest |

---

## Post-Onboard

### CLAUDE.md Injection

```markdown
# Architecture
This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits, file context -> `/c3`.
Operations: query, audit, change, ref, rule, sweep.
File lookup: `c3 lookup <file-or-glob>` maps files/directories to components + refs.
```

### Capabilities Reveal

```
## Your C3 toolkit is ready

**Typical flow:**

1. Understand what exists: `c3 list` → topology + coverage, then `c3 lookup <file>` → which component owns it
2. Make changes: `c3 add` to create NEW entities — connect them by authoring the edge-marked column in the body (citations wire at import), use `--file <path>` for bodies with tables, mermaid, or code fences. Once an entity exists as a frozen fact, editing it goes through a change-unit, not `write`/`set`/`delete` (`c3 set <id> codemap` stays direct).
3. Validate: `c3 check` catches broken links, schema gaps, orphans
4. Visualize: `c3 graph <container-or-component> --format mermaid` renders architecture as mermaid diagrams

For architecture questions, changes, audits → just say `/c3` + what you want.

Run `c3 --help` to see all available commands.
Run `c3 <command> --help` for detailed usage.
```

## Complexity Guide

Use this to choose the **rung** — how much canvas the project warrants now. Higher complexity earns the optional, deeper sections (Foundational Flow, Business Flow, Change Safety); lower complexity stays on the lean rung-1 seed. Size to where you are, and climb deliberately later (see laddering above, and `canvas.md`) — don't pre-author a higher rung's sections.

| Level | Signals | Rung |
|-------|---------|------|
| trivial/simple | Single purpose | Lean rung-1 seed, deeper sections skipped |
| moderate | Multiple concerns | Rung-1 + 2-3 deeper sections where they earn it |
| complex | Orchestration | Raise the canvas; full discovery + code-map |
| critical | Distributed/compliance | + rationale each, deeper sections required |

Discover the right depth from code, never assume from templates.
