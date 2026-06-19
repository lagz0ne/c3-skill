# Change Reference — the change-unit saga (Act 2) + climbing a rung (Act 3)

Frozen facts change **only** through a change-unit (SKILL.md §The shared contract — cite it, don't re-derive it). This reference owns the saga: how a change-unit is authored, what its three carriers are, the four gates `apply` runs, and the operational steps for raising a rung.

A **change-unit = reasoning + change material**. The reasoning is a change-doc (an ADR — a `status:` doc, so *not* a frozen fact; author and revise it freely). The material lives in `.c3/changes/<adr-id>/` and is three kinds of carrier:

- `*.patch.md` — the **internal arm**: one primitive that mutates one fact.
- `*.codemap.md` — the **external arm**: declares a fact's code binding (globs), landed atomically with the patches.
- `*.inspect.md` — the **up-V arm**: attests the code deriving from a touched fact was inspected against the fact's post-change body.

**The ADR *is* the change-unit — same id.** `c3 change apply <adr-id>` lands every carrier all-or-nothing. That apply is the only legal mutation of a fact.

Spawn parallel subagents (Task tool) for analysis and multi-file authoring; they author into the same `.c3/changes/<adr-id>/` folder.

## The end-to-end sequence

```bash
# 1. Draft the reasoning. `c3 schema adr` LEADS with REJECT-IF; honor it.
c3 schema adr
c3 add adr <slug> --file adr-body.md     # slug = intent (add-rate-limiting). Tables/mermaid/code ⇒ --file.

# 2. Anchor each fact-edit: cite the block you will replace.
c3 read <id> --section <name> --cite     # → one handle per citable block: <id>#nNODE@vVER:sha256:HASH

# 3. Scaffold the unit folder (same id as the ADR), then author carriers into it.
c3 change new <adr-id>                    # → .c3/changes/<adr-id>/
#    Author <seq>-<slug>.patch.md per fact-edit (see "Patch carriers"),
#    a <seq>.codemap.md per deliberate code-rebind (see "The code binding"),
#    and — for any touched fact with derivation obligations — its inspection (step 5).

# 4. Preview. The "files changed" panel: per-patch drift + state, plus carriers.
c3 change view <adr-id>
c3 graph <id> --unit <adr-id>            # the post-change graph (staged edges), via a rolled-back apply
c3 change status <adr-id>                # per-patch state: pending / applied / drifted / new

# 5. Inspect (the up-V). Surfaces what each touched fact obliges you to check.
c3 change inspect <adr-id>               # → obligations + code-map territory + material hashes per fact
#    Author <seq>.inspect.md attesting each obligation (see "The inspection arm").

# 6. Record human judgment, then flip.
c3 change accept <adr-id>                # status → accepted (the one stored bit)
c3 change apply <adr-id>                 # the switch: drift → canvas → retire → inspection, atomic
c3 check                                 # close; --fix latches accepted → done when After-cites resolve
```

The **file-context gate is MANDATORY before authoring any fact-edit patch**: `c3 lookup <file>`, load every `rule-*` and the parent chain, honor the refs/rules. `apply` will not launder a non-compliant edit — the body you author must already comply. Each parallel subagent runs this gate on its own files.

## Patch carriers — the internal arm

A `*.patch.md` is YAML frontmatter + body, named `<seq>-<slug>.patch.md` (e.g. `01-tighten-goal.patch.md`); they apply in filename order. Each patch is **one primitive**:

```
---
target: <entity-id>
scope: block | insert | whole | frontmatter | retire
base: <cite-handle>        # required for every scope except no-base whole; absent ⇒ create
result: sha256:<hash>      # optional landing check (block) — see below
# type / parent / title / uses / boundary / category / date — create + frontmatter metadata
---
<body>
```

| Scope | What it does | Base | Body |
|-------|--------------|------|------|
| `block` | replace **one** cited block (EDIT an existing section); **empty body deletes it** | block cite handle | the new block content |
| `insert` | **add** a section the fact lacks, or a new table **row** | section → entity handle; row → the block cite of the row to insert *after* | the new `## Section` (no duplicate) or the new row |
| `whole` (no base) | **create** a new fact, born sealed | absent | full body; `type:` required |
| `frontmatter` | rename (`title`) / move (`parent`) / re-edge (`uses`) / set `boundary`, `category`, `date` | entity handle | frontmatter deltas |
| `retire` | remove the fact + its edges | entity handle | — |

**`block` EDITS; `insert` ADDS.** Change a section that exists → `block`. Give a fact a section it lacks (the rung-climb move, §Climbing a rung) → `insert`: it appends additively, every existing section stays frozen. The `insert` body must start with a `## heading` and may not duplicate a section the fact already has.

**Table rows.** Cite the specific row (`--cite` lists per-node handles). Edit a row → `block` patch whose body is *just that row* (`| a | b | c |`, normalized to the stored cells — don't re-supply the header). Delete a row → `block` with an empty body. Add a row → `insert` with the row to insert *after* as the base. (Both anchor by the cited block's hash, so they survive node renumbering.)

**Cite handles** (from `c3 read <id> --cite`): a **block** anchor `entity#nNODE@vVER:sha256:HASH` pins one node by its hash (`block` scope); the **entity** anchor `entity@vVER:sha256:ROOTMERKLE` pins the whole fact (`insert` / `frontmatter` / `retire`).

**Membership rows are NOT yours — set `parent:`, the row appears** (SKILL.md §Membership). A parent's `Components`/`Containers` table is synthesized from children's `parent:` edges on every parentage path. Never insert, re-cite, or hand-remove a membership row; a reparent/retire heals the parent it leaves. Author a parent patch only when its **Responsibilities** or a member's **Goal Contribution** *framing* changes — that is a second patch, authored together (the parent-delta decision: record `Parent Delta: updated` and name the patch, or `Parent Delta: none` with evidence).

**`whole` *with* a base is REJECTED** — full-replace of a live fact must be block-anchored. Don't author it.

**The `result:` landing check** (block only). When set, the applied block must seal to exactly that hash or apply rejects *before that node is written* — so what lands is exactly what was reviewed. Omit it and the edit lands on the first `apply` (drift + canvas still run); there is no read-back loop. To pin deterministically: seed `result: sha256:0`, apply, copy the real hash from the rejection (`landing mismatch — applied content seals to sha256:<HASH>`; the node is left untouched), paste, re-apply. Or compute it as the `sha256` of the body as authored (trailing newlines trimmed).

## The code binding — the external arm

A fact's `codemap` globs are the one part of its footprint C3 cannot freeze (code moves on a cadence C3 doesn't control), so the binding is **declared and verified, never frozen**:

- **Maintain it live.** `c3 set <id> codemap '<glob>'` is the *one* `set` not refused on a frozen fact (SKILL.md §The shared contract) — routine drift-maintenance, not a change-unit event.
- **Declare a deliberate rebind in the unit.** When work moves or renames a fact's code, carry a `<seq>.codemap.md`: frontmatter `target:` (+ optional `base:` globs for the view), body = the full post-state, one glob per line. It lands in the **same transaction** as the patches; a carrier whose target neither exists nor is created by this unit rejects the whole unit, and two carriers for one target is refused (one carrier per target). A carrier-only unit is valid.
- **Match it at check.** `c3 check` WARNs when a declared glob matches no files; `--strict-codemap` promotes it to an error (SKILL.md §Command table).

## The inspection arm — the up-V the switch forces

A touched fact may **derive** code (a `## Derived Materials` row with a real Material) or carry a **risk** that demands proof (a `## Change Safety` row with a Required Verification). When you change such a fact's body, `apply` forces an attestation that the deriving code was actually inspected against the new contract. This is the up-V of the switch-gated double-V: the down-V (cited base, `result:` hash, resolving *After* cite) proves the doc; the up-V proves the code matches it.

```bash
c3 change inspect <adr-id>
```

`inspect` surfaces, per touched fact: its **obligations** (the Derived-Materials / Change-Safety duties), the resolved **code-map territory** (the real files those globs match — the only place evidence may point), and the **material hashes** to stamp into `covers`. Author `<seq>.inspect.md`:

```
---
target: <entity-id>
covers:                         # the change-material this was inspected against, with hashes
  - source: 01-tighten-goal.patch.md
    hash: sha256:<from `c3 change inspect`>
---
| Obligation | Territory | Verdict | Evidence |
|------------|-----------|---------|----------|
| <the duty> | <file in territory> | matches | <command / path / entity id naming a file in territory> |
```

The gate (`apply`'s fourth gate) refuses unless, for every contract-touched fact with obligations: an inspection exists, its `covers` is **fresh** against the unit's current material, every row's verdict is `matches` or `updated`, every evidence is **grounded** (names a command, path, or entity id), and **at least one row's evidence names a file inside the fact's resolved territory** — the anti-rubber-stamp floor (a "matches" with no real file in scope does not pass). Self-attestation forces the inspection to *exist*, grounded and reviewable; the human judges its truth at `change accept`.

**Scoping — when the up-V does and doesn't fire:**

- **Contract change required.** Only `block`/`insert`/`whole` patches and codemap rebinds trigger it. A `frontmatter` re-edge (rename / `uses`) or a `retire` does **not** change the contract the code derives from → no inspection.
- **Docs-ahead-of-code defers.** A fact that declares obligations but whose code-map resolves to **zero** files (onboarding / docs-first) **defers** — there is nothing to inspect yet; the inspection fires later, when the map binds real files and the fact is next changed.
- A fact with no derivation obligations needs no inspection regardless of how it was edited.

## The apply gates

`c3 change apply <adr-id>` runs a **preflight over ALL carriers before any write**, then writes inside **one transaction** — the unit lands completely or not at all; you never inspect a half-applied state. Four gates, in order:

1. **Drift / conflict** — every cited anchor must be fresh. A `block` patch checks the cited node's **hash** (a sibling block flipping does not stale you); a `frontmatter`/`retire` patch checks the entity's root merkle. A patch whose anchor is **gone** is a *conflict* (the frozen block moved under you) → the rebase loop below.
2. **Canvas** — the merged body (edit) or new body (create) must stay valid for its canvas.
3. **Retire safety (destruction gate)** — a `retire` is refused if it would **orphan a live child** or **dangle a live citer**, *unless this same unit* also retires/reparents the child and drops the citer's citation. (The membership-row drop is automatic.) Resolve the consequences in the unit; the destruction lands all-or-nothing. (`sweep.md` predicts this before you author.)
4. **Inspection (up-V)** — every contract-touched fact with obligations needs a fresh, territory-grounded `*.inspect.md` (above). Computed from the unit's preview overlay, so it runs only once gates 1–3 pass.

A landing-hash mismatch on a later patch, two patches editing the same block, or a missing codemap target rolls back **every** earlier write — node, edge, seal, and codemap — together. Fix the cause and re-run. `--dry-run` reports the writes without performing them.

**Conflict → rebase loop** when apply rejects with drift/conflict. Re-author the patch against the moved frozen state; apply re-runs every gate, so a stale resolution still can't land:

```bash
c3 change rebase <adr-id>                 # per conflict: BASE (anchored) + YOURS (your change) + the re-anchor
c3 read <id> --section <name> --cite      # re-read the moved block → fresh handle (CURRENT)
#   re-author the patch's base: (+ body, + result:) onto the live block — your intent, merged
c3 change apply <adr-id>                   # retry
```

## Climbing a rung (Act 3, operational)

A canvas is a **rung** — a fact is always complete to its current rung; growth is climbing to a higher one, never relaxing completeness (SKILL.md §A fact is always complete to its rung; the *why* is in `canvas.md`). The climb is a change-unit like any other — an ADR records *why* the project moved up a level, and `insert` patches carry each fact across.

**Order the new sections LAST in the canvas.** `insert` *appends* each new section at the **end** of a fact's body, so a climb stays check-clean only if the newly-required sections sit after every already-present section in the canvas's order (higher-rung sections are deeper → last). A new required section placed *before* existing ones makes the appended order mismatch the canvas and `check` fails with `sections out of order`. The seed canvases already order higher-rung sections last; preserve that.

```bash
# 1. Raise the canvas (user-owned) — make an optional section required, or author a richer one,
#    keeping newly-required sections ordered LAST.
c3 canvas write <type>

# 2. The bar moves; every fact below it now fails its canvas.
c3 check                                   # lights up exactly the facts missing the new section(s)

# 3. Stage the climb (same id as the ADR for the climb).
c3 change scaffold <adr-id>                # one EMPTY insert patch per fact below the bar:
#    the heading + the table's column headers, no rows. The emptiness IS the gate.

# 4. Fill each template — author the real section content for every staged patch.
#    This is the migration: each fact climbs to the new contract, completely.

# 5. Land it — gated, atomic.
c3 change apply <adr-id>                    # REFUSES an empty required section, so an unfilled
#    template blocks the whole unit. The climb cannot land until every fact carries it.
c3 check
```

`change scaffold` does not author content — it stakes out *where* each fact is short of the raised bar and hands you empty, apply-refusing templates so the climb is impossible to fake.

## Anti-goals

- **Don't route creation through a change-unit.** A new fact is not frozen — `c3 add` it (the unguarded create path), or use a no-base `whole` patch only when you deliberately want it sealed in the unit.
- **Don't route canvas-definition edits through a change-unit.** Canvases are user-owned markdown — `c3 canvas write` or edit by hand (`canvas.md`).
- **Don't author `whole`-with-base patches** (full-replace of a live fact — rejected; a live section is a `block` patch, a new section is an `insert`).
- **Don't use `insert` to edit an existing section** — `insert` only appends a section the fact lacks.
- **Don't `write`/`set`/`delete` a fact directly** — refused; author a patch (SKILL.md §The shared contract).
- **Don't hand-author a membership row** — set `parent:`; the tool synthesizes it.
- **Don't expect a body edit to advance status** — status moves only through `accept`, the auto-done latch, or `supersede` (SKILL.md §ADR status set).
