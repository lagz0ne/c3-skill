# Change Reference

A change is a **work order with two parts**: the reasoning (the change-doc — usually an ADR) and the change material — patch files that mutate frozen facts, plus optional `.codemap.md` carriers that re-bind their code. The doc is yours to write freely. The facts are frozen — they move only when you `change apply` (which also lands the carriers, atomically, in the same transaction).

Flow: `ADR (work order) → Understand → Author patches → Accept → Apply → Audit`

Spawn parallel subagents via the Task tool for analysis and multi-file work.

## The one rule that reshapes everything

**A fact is frozen.** A fact = any entity whose canvas declares no `status:` — `system`, `container`, `component`, `ref`, `rule`, `recipe`, `pm-requirement`, `user-story`. The CLI **refuses** `write` / `set` / `delete` when the first argument is an existing fact:

```
error: <id> is a fact — facts are frozen and change only through a change-unit
hint: author patches in .c3/changes/<unit-id>/ then run 'c3x change apply <unit-id>'
```

This is not a bug to work around. It is the contract. There are exactly three moves:

| You want to… | Path | Guarded? |
|--------------|------|----------|
| **Edit an existing fact** | author patch file(s) in `.c3/changes/<unit-id>/`, then `c3 change apply` | this is the ONLY path |
| **Create a new entity** | `c3 add <type> <slug>` directly (or a no-base patch) | no — unguarded |
| **Write the reasoning (the ADR)** | `c3 add adr` / `c3 write` / `c3 set` on the doc | no — a change-doc is not a fact |

The change-doc (ADR, prd, atomic-design-change) declares `status:`, so it is **not** frozen. You author and revise its prose freely. The fact-mutations ride alongside it as patch files. **The ADR *is* the change-unit — same id.** The folder `.c3/changes/<adr-id>/` holds the patches; `c3 change apply <adr-id>` lands them.

---

## Progress Checklist

```
- [ ] Phase 1: `c3 schema adr` read; ADR body drafted to the canvas; `c3 add adr` created
- [ ] Phase 2: topology loaded, impact analyzed, ADR is a complete work order
- [ ] Phase 2b: provision gate (implement now, or design-only?)
- [ ] Phase 3: patches authored into .c3/changes/<adr-id>/ (creates direct; fact-edits as block patches)
- [ ] Phase 3a: contract-cascade gate satisfied by the authored patches (parent delta decided)
- [ ] Phase 3b: ref/rule-compliance gate satisfied by the authored patches
- [ ] Phase 4: `c3 change accept` → `c3 change apply` → `c3 check`; ADR latched to `done`
```

---

## Phase 1: ADR First — the work order (non-negotiable)

```bash
c3 schema adr
```

Read the canvas **before** writing the ADR — do not draft freehand and reconcile later. The schema output:
- LEADS with a `REJECT IF:` block — those bullets are the rejection contract; trip one and creation fails.
- Per-section `fill:` says what to write; `rejected when:` says what gets bounced.
- It is the earliest enforcement surface, not a late cleanup step.

The ADR is a change-doc, **not** a frozen fact, so you author and revise it with the normal commands — `c3 add adr`, `c3 write <adr-id>`, `c3 set <adr-id> <field>`. This is the bootstrap resolution: the reasoning surface is editable; only the facts it changes are frozen.

Draft `adr-body.md` to the canvas, then create it all at once:

```bash
c3 add adr <slug> --file adr-body.md
```

Slug = change intent (`add-rate-limiting`, `migrate-to-postgres`). The adr canvas is **laddered**: `c3 add adr` requires only the lean core (Goal, Context, Decision, Affected Topology, Verification) — enough for a small change; the work-order sections (Compliance Refs/Rules, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives, Risks) are optional, climbing in for weightier decisions. No thin ADR — any section you include must be substantive (thin sections fail; `N.A - <reason>` for inapplicable rows). Any section with tables, mermaid, or code fences MUST be authored via `--file`.

From a diff, capture rationale into the draft first, then reshape to the canvas:
```bash
git diff <ref> > adr-notes.diff
```

## Phase 2: Understand + complete the work order

```bash
c3 list
```

Clarify with the user (ASSUMPTION_MODE: skip). Analyze:
- Affected containers, components, refs, rules.
- Per file mentioned or discovered: `c3 lookup <file>` — load the constraint chain *before* reasoning. No mapping → uncharted territory; flag the coverage gap.
- `c3 read` upward: component → container → context → cited refs/rules.
- Risks and consumers.

The ADR body must let a later agent recover the decision without chat history. Refine it freely (it is not frozen): `c3 write <adr-id> --section <name> --file <path>` for rich content, `c3 set <adr-id> <field> <value>` for frontmatter. Follow `c3 schema adr` sections and `help[]` literally; if a section exists to prevent a specific failure, fill it to prevent exactly that failure. For validator/schema/command/ref/rule changes, the ADR must preserve the C3 underlay: exact commands, validators, tests, help/hints, schemas, verification evidence.

**Visual impact:** `c3 graph <primary-affected-container-or-component> --format mermaid` — include in the approval presentation. Multiple containers → graph each.

Present for approval (ASSUMPTION_MODE: mark `[ASSUMED]`). Complex changes: spawn parallel analyst + reviewer subagents, synthesize.

## Phase 2b: Provision Gate

Ask (ASSUMPTION_MODE: skip):
- **Implement now** → Phase 3.
- **Design only** → author the new-entity docs the design needs (`add`, unguarded), leave the ADR open, stop. No fact-edit patches authored yet.

To implement later: reopen the ADR, author the patches, resume Phase 3.

## Phase 3: Author the change material

This phase is where the v11 model bites. **You do not mutate facts here — you author patches that will mutate them at apply.** Two kinds of work:

### 3.1 Create new entities — direct, unguarded

A new container/component/ref/rule does not yet exist, so it is not frozen. Create it the normal way (`add` patterns in `onboard.md` §1.2–1.4; body via `--file` for tables/mermaid/code, stdin for prose):

```bash
cat body.md | c3 add component <slug> --container <id>
```

### 3.2 Edit an existing fact — the change-unit sequence

This is the **only** way to change a fact. Every step:

```bash
# 1. Anchor: get the cite handle for the block you will replace.
c3 read <id> --section <name> --cite
#    → emits one handle per citable block in that section, shape:
#      <id>#nNODE@vVER:sha256:HASH "snippet"

# 2. Scaffold the change-unit folder (same id as the ADR).
c3 change new <adr-id>
#    → .c3/changes/<adr-id>/ ; drop <seq>-<slug>.patch.md files there

# 3. Author <seq>-<slug>.patch.md (see "Patch file shape" below).
#    scope: block · base: the cite handle from step 1 · result: the new block's seal hash

# 4. Preview — two-arm "files changed" panel: internal patches (drift verdict) +
#    external codemap carriers (applied? which globs resolve?).
c3 change view <adr-id>
#    Preview the GRAPH the unit would produce (staged edges included), without
#    landing it: `c3 graph <id> --unit <adr-id>` runs the real apply in a
#    rolled-back transaction — the contextual lens. cite/stage → graph --unit → apply.

# 5. Per-item state — patches (pending / applied / drifted / new) + carriers.
c3 change status <adr-id>

# 6. Human judgment — records the one stored bit.
c3 change accept <adr-id>

# 7. The only mutation. Two gates, atomic. --dry-run to preview the writes.
c3 change apply <adr-id>

# 8. Close.
c3 check
```

**File context gate (SKILL.md §File Context) is MANDATORY before authoring any fact-edit patch** — lookup the file, load every `rule-*` and the parent chain, honor refs/rules. The patches you author must already satisfy the gates below; `change apply` will not launder a non-compliant edit.

Parallel subagents: each runs the file-context gate on its files, authors its patches into the same `.c3/changes/<adr-id>/` folder (patches apply in filename order), and proves code + docs move together with no regressions.

## Patch file shape

A patch file is YAML frontmatter + a body. File name: `<seq>-<slug>.patch.md` (e.g. `01-tighten-goal.patch.md`); they apply in filename order.

```
---
target: <entity-id>
scope: block | whole | frontmatter | retire
base: <cite-handle>        # required for every scope except no-base whole; absent ⇒ create
result: sha256:<hash>      # landing check (optional but recommended) — see below
# type / parent / title / uses / boundary / category / date — create + frontmatter metadata
---
<body>
```

The scopes you will actually use:

| Scope | What it does | Base | Body |
|-------|--------------|------|------|
| `block` | replace **one** cited block (EDIT an existing section); **empty body deletes it** | required (block cite handle) | the new block content |
| `insert` | **append a NEW section** to a frozen fact — additive, existing sections stay frozen | entity handle (`entity@vN:sha256:MERKLE`) | the new section; MUST start with a heading (`## Name`), MUST NOT duplicate an existing section |
| `whole` (no base) | **create** a new fact, born sealed | absent | the full body; `type:` required |
| `frontmatter` | rename (`title`) / move (`parent`) / re-edge (`uses`) / set `boundary`, `category`, `date` — parity with `set` | entity handle | frontmatter deltas |
| `retire` | remove the fact + its edges | entity handle | — |

**`block` EDITS an existing section; `insert` ADDS a new one.** When a section already exists and its content must change, replace it with a `block` patch. When the fact must *gain* a section it does not have — the move that lets a sealed fact grow as the rung rises (see §Climbing a rung) — use `insert`: it appends additively, leaving every existing section frozen, anchored to the entity handle from `c3 read <id> --cite`. The `insert` body must START WITH A SECTION HEADING and may not duplicate a section already on the fact.

**Editing one table row.** Cite the specific row (`c3 read <id> --section <name> --cite` lists per-node handles), then `block`-patch it: the body is **just that row** — paste it as natural markdown (`| a | b | c |`, outer pipes optional) and it's normalized to the stored cells; an **empty body deletes the row**. You do NOT re-supply the whole table or the header/separator. (To add a row, that's a body change to the table block — cite the table and replace it, or `insert` if it's a new section.)

One scope is deliberately closed and you must not author it:
- `whole` **with a base** (full-replace of a live fact) is **REJECTED** — an edit to a live fact must be block-anchored.

**Cite handles** (block anchors from `c3 read <id> --section <name> --cite`; the entity anchor from `c3 read <id> --cite`):
- block anchor: `entity#nNODE@vVER:sha256:HASH` — anchors one node by its hash (`block` scope).
- entity anchor (`insert` / frontmatter / retire): `entity@vVER:sha256:ROOTMERKLE` — anchors the whole fact by its root merkle.

**The `result:` landing check.** When `result:` is set, the applied block must seal to exactly that hash or apply rejects *before that node is written* — so what lands is exactly what was reviewed. When omitted, the check is skipped and the edit simply lands on the first `apply` (drift + canvas gates still run). There is no "apply then read it back" loop — a no-`result:` apply prints no hash and has already mutated the block. To pin the hash deterministically: seed `result: sha256:0`, run `c3 change apply`, and copy the real hash from the rejection it prints (`landing mismatch — applied content seals to sha256:<HASH>`; the node is left untouched), then paste it in and re-apply. Or compute it directly as the `sha256` of the patch body exactly as authored (trailing newlines trimmed). Drift already guarantees you are editing the block you anchored; `result:` adds the content-exactness lock.

## Climbing a rung

The canvas is a **rung** — a complete contract for one complexity *level*, not a form you fill in over time. A fact is *always* complete to its current rung; completeness is never relaxed. What grows is the complexity **level**, and you grow it **deliberately** by climbing a rung: raise the canvas, then migrate every fact up to the new contract, completely. Integrity is why migration exists — a fact may not straddle two rungs, so the moment the bar rises, every fact below it must rise with it. Rung-1 is meant to be lean; do not over-engineer it with sections a later rung carries. Climb when the architecture earns it.

The climb is a change-unit like any other — an ADR records *why* the project moved up a level, and `insert` patches carry each fact across. `change scaffold` stages those patches for you:

**Order the new sections LAST in the canvas.** `insert` (the climb's mechanism)
*appends* each new section at the **end** of a fact's body. So a climb only stays
check-clean if the newly-required sections sit **after** every already-present
section in the canvas's order — higher-rung sections are deeper, so they belong last.
If you raise the canvas with a new required section placed *before* existing ones,
the appended body order won't match the canvas order and `check` fails with
`sections out of order`. The seed canvases already order their higher-rung sections
last; preserve that when you author a richer canvas.

```bash
# 1. Raise the canvas — make an optional section required, or author a richer one.
#    Keep the newly-required sections ordered LAST (insert appends; see above).
c3 canvas write <type>            # the deliberate decision to climb (user-owned)

# 2. The bar moves; every fact below it now fails its canvas.
c3 check                          # lights up exactly the facts missing the new required section(s)

# 3. Stage the climb into the change-unit (same id as the ADR for the climb).
c3 change scaffold <adr-id>
#    → scans every fact, finds those below the canvas's current required bar,
#      and writes one `insert` patch per fact with an EMPTY required-section
#      template: the heading + the table's column headers, no rows.
#    The emptiness is the gate — see step 5.

# 4. Fill each template — author the real section content for every staged patch.
#    This is the actual migration work; each fact climbs to the new contract, completely.

# 5. Land it — gated, atomic.
c3 change apply <adr-id>
#    `apply` REFUSES to land an empty required section, so an unfilled template
#    blocks the whole unit ("won't apply on error"). The climb cannot land until
#    every fact genuinely carries the new section. All-or-nothing as always.

c3 check                          # confirm the new rung holds across all facts
```

`change scaffold` does not author content — it stakes out *where* each fact is short of the raised bar and hands you empty, apply-refusing templates so the climb is impossible to fake. Fill them, then apply lands the whole rung at once.

## Phase 3a: Contract Cascade Gate — satisfied *by the patches*

Every fact delta must have a parent-delta decision **before apply**. A component edit that needs a container update means **two patches** (or one create + one block edit), authored together.

| Layer | Question | Verdict | Evidence |
|-------|----------|---------|----------|
| Component | Did Goal, Parent Fit, Governance, Contract, or Derived Materials change? | YES/NO | the patch / block |
| Container | Did the Components table, Responsibilities, or a member's Goal Contribution change? | YES/NO | the patch, or no-delta reason |
| Context | Did project/container topology change? | YES/NO | the patch, or no-delta reason |
| Refs/Rules | Did shared constraints change? | YES/NO | the patch, or no-delta reason |

Rules:
- New component (a create patch / `add`) → the parent container's `## Components` block needs its own **patch** before the change-unit is complete.
- Component goal/dependency/interface change → parent `Goal Contribution` + `Responsibilities` MUST be reviewed.
- `NO` requires evidence. A blank parent delta = STOP.
- The ADR records `Parent Delta: updated` (name the patch) or `Parent Delta: none` (with evidence).

## Phase 3b: Ref/Rule Compliance Gate — satisfied *by the patches*

**Before `change accept`, verify the body each patch will produce complies with applicable refs and rules.** The canvas gate at apply checks shape, not compliance — compliance is on you.

Per file the change touches:
```bash
c3 lookup <file-path>
```

Per returned ref, check the body the patch produces by comparison mode:

| Ref Section | Mode | Check |
|-------------|------|-------|
| `## How` (code) | Structural | Matches the golden pattern structure? |
| `## How` (prose) | Semantic | Follows the described approach? |
| `## Choice` only | Negative | Contradicts the stated choice? |
| `## Not This` | Anti-pattern | Resembles the rejected alternative? |

**Rule Compliance (strict):** per `rule-*` from lookup, compare against `## Golden Example` + `## Not This` already loaded in Phase 3. Strict — must match the golden pattern; flag any deviation as a violation.

**ADVERSARIAL FRAMING: look for violations — never confirm compliance.**

```
| Ref/Rule | Section Checked | Verdict | Evidence |
|----------|-----------------|---------|----------|
| ref-X    | How             | COMPLIANT | matches pattern structure |
| rule-Y   | Golden Example  | VIOLATION | uses rejected approach Z |
```

Rules:
- Scope to YOUR CHANGES — no full-codebase audit.
- Ref wins. Code disagrees with ref → ref is right; an override needs the ADR's documented `## Override` process.
- Conflicts → specificity wins (component ref > container ref > context ref).

## The two apply gates (and how to recover)

`c3 change apply <adr-id>` runs a **preflight over ALL patches before any write**. Two mechanical gates:

1. **Drift** — every cited anchor must still be fresh. A block patch checks the cited node's **hash** (not the entity version — a sibling block flipping does not stale you); a frontmatter/retire patch checks the entity's root merkle.
2. **Canvas** — the merged body (block edit) or new body (create) must stay valid for its canvas.

**Apply is all-or-nothing, fully transactional.** The preflight rejects on any gate failure before a single write; and the write phase itself runs inside one transaction. So even a failure that only surfaces mid-write — a landing-hash mismatch on a later patch, two patches editing the same block, or a codemap carrier whose target is missing — rolls back every earlier patch's node, edge, and seal **and** every codemap write together. The unit lands completely or not at all; you never inspect a half-applied state. Fix the cause and re-run.

**Drift → rebase loop** when apply rejects with drift:
```bash
c3 change rebase <adr-id>      # emits, per drifted patch: the anchor it expected vs. live
c3 read <id> --section <name> --cite   # re-read the moved block → fresh handle
#   re-author the patch's base: (and result:) with the fresh handle/hash
c3 change apply <adr-id>       # retry
```

## The external arm — code bindings you declare and match

Patches are the **internal** arm: they edit frozen facts, and the apply gates + the
auto-done latch *match* what the unit declared (a cited base, a `result:` hash, a
resolving *After* cite). That is the left-and-right of a "V" for everything C3 owns.

But a fact's **code binding** — its `codemap` globs, the files it maps to — is the one
part of the footprint C3 cannot freeze: code moves, gets renamed, gets deleted, on a
cadence C3 doesn't control. So the binding is **declared and verified, never frozen**:

- **Maintain it live.** `c3 set <id> codemap '<glob>'` is the *one* `set` that is **not**
  refused on a frozen fact — code churns, so keeping the map current is routine, not a
  change-unit event. Use it for drift-maintenance any time.
- **Declare it in the unit.** When a change deliberately re-binds a fact (the work moved
  or renamed its code), carry that as a `.codemap.md` file in `.c3/changes/<adr-id>/` so
  the re-binding is part of the work order and lands atomically with the patches:

  ```
  ---
  target: <entity-id>
  base:                       # optional: the globs before this change (for the view)
    - <old-glob>
  ---
  <new-glob>                  # body = the full declared post-state, one glob per line
  <new-glob>
  ```

  `c3 change apply` applies carriers in the **same transaction** as the patches (a
  carrier whose target is missing rejects the whole unit). A carrier-only unit is valid.
- **Match it at check.** `c3 check` runs an introspection over an `accepted` unit's
  Affected-Topology entities: a declared glob that matches **no files** WARNs — the
  binding you declared doesn't resolve (the code isn't where the map says). It is
  **WARN-only and never gates `done`** — code churn is expected, not a release blocker.
  `c3 check --strict-codemap` promotes it to an error for callers that want the gate.

`c3 change view` / `c3 change status` show **both arms**: internal patches (drift/state)
and external codemap carriers (applied? which globs are unresolved?). That two-arm view
is the unit's full footprint — what it *derives* (declares) and how well it *matches*.

## Status — derived, never typed

**Per-patch state** is computed from seal hashes (`c3 change status`), never stored:

| State | Meaning |
|-------|---------|
| `pending` | anchor fresh, not yet applied |
| `applied` | the live block already seals to the patch's `result` |
| `drifted` | the anchor moved to something unexpected → rebase |
| `new` | a create patch whose target does not exist yet |

**Change-doc status** is the canonical set `[open, accepted, done, superseded]`. (`c3 add adr` stamps `proposed` at creation — the unmigrated synonym for `open`; `c3 migrate` folds it, and `change accept` works from either.)
- `c3 change accept <adr-id>` records the one stored bit (`→ accepted`) — human judgment.
- `accepted` **auto-latches to `done`**, one-way, when the doc's per-row *After* cites all resolve fresh. The latch is observed and actualized at `c3 check --fix` (not during apply) — proof the change actually landed.
- `superseded` is reachable only via the `supersede` command.

**Edit-proof:** a body write **never** moves `status` — only the privileged status writer does. So revising the ADR's prose with `c3 write` does **not** advance it; do not expect a body edit to change status.

## Phase 4: Audit + close

```bash
c3 check
```

- Docs match code; the component reference column (Governance) updated.
- CLAUDE.md blocks updated: `<!-- c3-generated: c3-NNN -->` … `<!-- end-c3-generated -->`.
- `c3 check --fix` actualizes the ADR `accepted → done` once all After-cites resolve fresh.

**No manual re-index.** `change apply` does not touch embeddings — the next `c3 search` reconciles incrementally (a fact's `text_hash` mismatch re-embeds only that fact). `c3 index` is a full rebuild for maintenance only (force-warm / model bump), never a correctness step after a change. Offline / no local model: lexical + graph still answer; only the semantic lane drops, silently.

---

## Anti-goals

- **Don't route new-entity creation through a change-unit.** A new fact is not frozen — `c3 add` it directly. Patches are for editing *existing* facts (and for creates only when you deliberately want them sealed in the unit).
- **Don't route canvas-definition edits through a change-unit.** Canvases (`.c3/canvases/<type>.md`) are user-owned markdown, not facts — edit them via `c3 canvas write` or by hand (see `canvas.md`).
- **Don't author `whole`-with-base patches.** Full-replace of a live fact is rejected — an edit to a live section is a `block` patch; *adding* a new section is an `insert` patch (see §Climbing a rung).
- **Don't use `insert` to edit an existing section** — `insert` only *appends* a section the fact lacks; changing a section that already exists is a `block` patch.
- **Don't try to `write` / `set` / `delete` a fact directly** — it is refused. Author a patch.
- **Don't expect a body edit to advance status.** Status moves only through `accept` / the auto-done latch / `supersede`.

---

## Regression

| Discovery | Action |
|-----------|--------|
| Changes problem | Back to Phase 1 (rewrite the ADR — it's editable) |
| Changes approach | Back to Phase 2 |
| Expands scope | Amend the ADR, add patches |
| A patch drifted | `change rebase` → re-cite → re-author → re-apply |
| Implementation detail | Adjust the affected patch |

---

## ADR Lifecycle

Canonical change-doc set: `open → accepted → (done | superseded)`. ADRs are hidden by default; `--include-adr` to inspect.

- `c3 change accept <adr-id>` is the path to `accepted` (records the stored bit).
- `accepted → done` is the one-way auto-done latch — earned when After-cites resolve fresh, actualized at `c3 check --fix`. You do not type it.
- `c3 check --include-adr` skips terminal change-docs (`done`, `superseded`) — historical, content frozen. `c3 check --only <adr-id>` forces re-validation of a specific one.
- `supersede` is the only path to `superseded`.

(A new ADR is born `proposed` — the unmigrated synonym for `open`; `c3 change accept` advances it. `c3 migrate` folds legacy `proposed`/`implemented`/`provisioned` onto the canonical set.)
