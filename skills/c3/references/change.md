# Change Reference — the change-unit saga (Act 2) + climbing & morphing (Act 3)

Frozen facts change **only** through a change-unit (SKILL.md §The shared contract — cite it, don't re-derive it). This reference owns the saga: how a change-unit is authored, what its carrier is, the four gates `apply` runs, and the operational steps for raising a rung.

A **change-unit = reasoning + change material**. The reasoning is a change-doc (an ADR — a `status:` doc, so *not* a frozen fact; author and revise it freely). The material lives in `.c3/changes/<adr-id>/` and is one kind of carrier:

- `*.patch.md` — one primitive that mutates one fact.

Code-conformance is **not** carried here. The fact→code binding lives in `.c3/eval/<fact>.yaml` (an ordinary editable file, never a change-carrier), and whether the code still matches the doc is checked by the wrapper's `eval` operation — a separate, off-the-switch run, never a gate (`references/eval.md`).

**The ADR *is* the change-unit — same id.** The wrapper's `change apply <adr-id>` operation lands every patch all-or-nothing. That apply is the only legal mutation of a fact.

Spawn parallel subagents (Task tool) for analysis and multi-file authoring; they author into the same `.c3/changes/<adr-id>/` folder.

## The end-to-end sequence

```bash
# 1. Draft the reasoning. `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" schema adr` LEADS with REJECT-IF; honor it.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" schema adr
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" add adr <slug> --file adr-body.md     # slug = intent (add-rate-limiting). Tables/mermaid/code ⇒ --file.
#   Author --file bodies OUTSIDE .c3/ (e.g. /tmp or repo root) — .c3/ is a managed tree that
#   regenerates and silently drops stray scratch files between commands.

# 2. Anchor each fact-edit: cite the block you will replace.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" read <id> --section <name> --cite     # → one handle per citable block: <id>#nNODE@vVER:sha256:HASH

# 3. Scaffold the unit folder (same id as the ADR), then author patches into it.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change new <adr-id>                    # → .c3/changes/<adr-id>/
#    Author <seq>-<slug>.patch.md per fact-edit (see "Patch carriers").

# 4. Preview. The "files changed" panel: per-patch drift + state.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change view <adr-id>
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" graph <id> --unit <adr-id>            # the post-change graph plus route facets (staged edges), via a rolled-back apply
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change status <adr-id>                # per-patch state: pending / applied / drifted / new

# 5. Record human judgment, then flip.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change accept <adr-id>                # status → accepted (the one stored bit)
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change apply <adr-id>                 # the switch: drift → canvas → morph → retire, atomic
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" check                                 # close; --fix latches accepted → done when After-cites resolve
```

The **file-context gate is MANDATORY before authoring any fact-edit patch**: run the wrapper's `lookup <file>` operation, load every `rule-*` and the parent chain, honor the refs/rules. `apply` will not launder a non-compliant edit — the body you author must already comply. Each parallel subagent runs this gate on its own files.

## Patch carriers

A `*.patch.md` is YAML frontmatter + body, named `<seq>-<slug>.patch.md` (e.g. `01-tighten-goal.patch.md`); they apply in filename order. Each patch is **one primitive**:

```
---
target: <entity-id>
scope: block | insert | whole | frontmatter | retire | canvas
base: <cite-handle>        # required for every scope except no-base whole and canvas; absent ⇒ create
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
| `frontmatter` | rename (`title`) / move (`parent`) / re-edge (`uses`) when the canvas does not own `uses` in a body edge-column / set `boundary`, `category`, `date` | entity handle | frontmatter deltas |
| `retire` | remove the fact + its edges | entity handle | — |
| `canvas` | **morph** a fact-TYPE's shape (target = the type, not a fact) — the evolve-unit, §Morphing the model | absent | the full new canvas definition |

**`block` EDITS; `insert` ADDS.** Change a section that exists → `block`. Give a fact a section it lacks (the rung-climb move, §Climbing a rung) → `insert`: it appends additively, every existing section stays frozen. The `insert` body must start with a `## heading` and may not duplicate a section the fact already has.

**Table rows.** Cite the specific row (`--cite` lists per-node handles). Edit a row → `block` patch whose body is *just that row* (`| a | b | c |`, normalized to the stored cells — don't re-supply the header). Delete a row → `block` with an empty body. Add a row → `insert` with the row to insert *after* as the base. (Both anchor by the cited block's hash, so they survive node renumbering.)

**Body-owned edges.** When `schema <type>` marks a table column with `edge: uses`, that body column is the canonical edge source. A `frontmatter` patch carrying `uses:` is rejected before any write; edit or insert the cited table row instead. The rejection names the owning section and column and prints the exact `read --section ... --cite` repair command. Frontmatter `uses:` remains legal for legacy or custom canvases without a body-owned `uses` column.

**Cite handles** (from `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" read <id> --cite`): a **block** anchor `entity#nNODE@vVER:sha256:HASH` pins one node by its hash (`block` scope); the **entity** anchor `entity@vVER:sha256:ROOTMERKLE` pins the whole fact (`insert` / `frontmatter` / `retire`).

**Membership rows are NOT yours — set `parent:`, the row appears** (SKILL.md §Membership). A parent's `Components`/`Containers` table is synthesized from children's `parent:` links on every parentage path. Never insert, re-cite, or hand-remove a membership row; a reparent/retire heals the parent it leaves. Author a parent patch only when its **Responsibilities** or a member's **Goal Contribution** *framing* changes — that is a second patch, authored together (the parent-delta decision: record `Parent Delta: updated` and name the patch, or `Parent Delta: none` with evidence).

**`whole` *with* a base is REJECTED** — full-replace of a live fact must be block-anchored. Don't author it.

**The `result:` landing check** (block only). When set, the applied block must seal to exactly that hash or apply rejects *before that node is written* — so what lands is exactly what was reviewed. Omit it and the edit lands on the first `apply` (drift + canvas still run); there is no read-back loop. To pin deterministically: seed `result: sha256:0`, apply, copy the real hash from the rejection (`landing mismatch — applied content seals to sha256:<HASH>`; the node is left untouched), paste, re-apply. Or compute it as the `sha256` of the body as authored (trailing newlines trimmed).

## What the switch proves — the down-V

The switch enforces **one** thing: that the *doc* the patch lands is exactly the doc that was reviewed. That proof is a single down-V — the cited `base` pins the block you're replacing, the optional `result:` hash pins what the edit seals to, and the change-doc's resolving *After*-cite pins that the fact actually landed. The block you author must already be canvas-correct and comply with its refs/rules (the file-context gate above); `apply` validates and seals it, it does not launder it.

**Code-conformance is not on this switch.** Whether the code still matches the doc is a separate question, answered by the wrapper's `eval` operation against the fact's `code:` binding in `.c3/eval/<fact>.yaml` — a one-off, CI-cadence verdict you run when you want proof, never a gate (`references/eval.md`). Route-enriched `graph` output can point to first inspection anchors, lanes, and drift labels, but it is also not proof. The binding file is an ordinary editable file: when work moves or renames a fact's code, re-aim its `code:` globs directly (it is never frozen, so no change-carrier is involved), then re-run `eval` through the wrapper.

## The apply gates

`C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change apply <adr-id>` runs a **preflight over ALL patches before any write**, then writes inside **one transaction** — the unit lands completely or not at all; you never inspect a half-applied state. Four gates, in order:

1. **Drift / conflict** — every cited anchor must be fresh. A `block` patch checks the cited node's **hash** (a sibling block flipping does not stale you); a `frontmatter`/`retire` patch checks the entity's root merkle. A patch whose anchor is **gone** is a *conflict* (the frozen block moved under you) → the rebase loop below.
2. **Canvas** — the merged body (edit) or new body (create) must stay valid for its canvas. When this unit also morphs the target's type (a `canvas` patch), the body is validated against the **new** shape — migrating an instance up to a reshaped canvas in the same unit is not rejected.
3. **Morph safety (evolve-unit gate)** — a `canvas` reshape lands only if, once this unit's migration patches apply, **every** instance of the type is valid against the new shape. A morph that would strand even one instance is refused — migrate them all in the same unit (§Morphing the model).
4. **Retire safety (destruction gate)** — a `retire` is refused if it would **orphan a live child** or **dangle a live citer**, *unless this same unit* also retires/reparents the child and drops the citer's citation. (The membership-row drop is automatic.) Resolve the consequences in the unit; the destruction lands all-or-nothing. (`sweep.md` predicts this before you author.)

A landing-hash mismatch on a later patch or two patches editing the same block rolls back **every** earlier write — node, wiring edge, membership, and seal — together. Fix the cause and re-run. `--dry-run` reports the writes without performing them.

**Conflict → rebase loop** when apply rejects with drift/conflict. Re-author the patch against the moved frozen state; apply re-runs every gate, so a stale resolution still can't land:

```bash
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change rebase <adr-id>                 # per conflict: BASE (anchored) + YOURS (your change) + the re-anchor
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" read <id> --section <name> --cite      # re-read the moved block → fresh handle (CURRENT)
#   re-author the patch's base: (+ body, + result:) onto the live block — your intent, merged
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change apply <adr-id>                   # retry
```

## Climbing a rung (Act 3, operational)

A canvas is a **rung** — a fact is always complete to its current rung; growth is climbing to a higher one, never relaxing completeness (SKILL.md §A fact is always complete to its rung; the *why* is in `canvas.md`). The climb is a change-unit like any other — an ADR records *why* the project moved up a level, and `insert` patches carry each fact across.

**Order the new sections LAST in the canvas.** `insert` *appends* each new section at the **end** of a fact's body, so a climb stays check-clean only if the newly-required sections sit after every already-present section in the canvas's order (higher-rung sections are deeper → last). A new required section placed *before* existing ones makes the appended order mismatch the canvas and `check` fails with `sections out of order`. The seed canvases already order higher-rung sections last; preserve that.

```bash
# 1. Raise the canvas (user-owned) — make an optional section required, or author a richer one,
#    keeping newly-required sections ordered LAST.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" canvas write <type>

# 2. The bar moves; every fact below it now fails its canvas.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" check                                   # lights up exactly the facts missing the new section(s)

# 3. Stage the climb (same id as the ADR for the climb).
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change scaffold <adr-id>                # one EMPTY insert patch per fact below the bar:
#    the heading + the table's column headers, no rows. The emptiness IS the gate.

# 4. Fill each template — author the real section content for every staged patch.
#    This is the migration: each fact climbs to the new contract, completely.

# 5. Land it — gated, atomic.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change apply <adr-id>                    # REFUSES an empty required section, so an unfilled
#    template blocks the whole unit. The climb cannot land until every fact carries it.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" check
```

`change scaffold` does not author content — it stakes out *where* each fact is short of the raised bar and hands you empty, apply-refusing templates so the climb is impossible to fake.

## Morphing the model (Act 3, the non-additive move)

A climb *adds* a required section; a **morph** *reshapes* the canvas — split a column, retype it to an enum, restructure or rename a section — and reshapes the guidance with it (a canvas carries its own `description`, section `purpose`, and `reject_if`, so morphing the canvas morphs the instructions too). This is the **evolve-unit**: a change-unit like the climb, with one extra carrier — a **`canvas`-scope patch** whose body is the whole new canvas definition (target = the type).

The rule that makes a morph safe is the climb's rule made general: **no fact straddles two shapes.** The morph gate lands the reshape only if, once this unit's migration patches apply, *every* instance of the type is valid against the new shape — it refuses a morph that strands even one. So the canvas patch and the instance migrations ride **together**, in one atomic flip; never reshape-now-fix-later.

```bash
# 1. Author the new shape as a canvas-scope patch — base-optional (it re-authors a TYPE, not a
#    frozen fact). <seq>-<slug>.patch.md:  target: <type>,  scope: canvas,  body = the full new def.

# 2. Migrate EVERY affected fact in the SAME unit — block / insert / (empty-body) block-delete
#    patches that bring each instance to the new shape. "Affected" includes facts that NAME the
#    old shape: a rule citing a column the morph removed is stranded too — re-author it here.

# 3. Land it — gated, atomic.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change apply <adr-id>    # the morph gate refuses unless every instance is valid against the new
#    shape once the unit's migrations apply; the canvas file and the facts flip in one transaction.
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" check
```

The canvas patch needs no `base`: it re-authors a fact-TYPE, and the instance-migration patches drift-protect the reshape (a concurrent morph would stale their anchors). A reshape with no instances to migrate is a free edit — use the wrapper's `canvas write` operation, not a unit.

## Anti-goals

- **Don't route creation through a change-unit.** A new fact is not frozen — use the wrapper's `add` operation (the unguarded create path), or use a no-base `whole` patch only when you deliberately want it sealed in the unit.
- **A free canvas edit goes through the wrapper's `canvas write` operation, not a unit** — canvases are user-owned markdown (`canvas.md`). But a **morph that must migrate existing instances** is the evolve-unit: a `canvas`-scope patch + the migrations in one gated, atomic unit (§Morphing the model). The test is whether facts must move *with* the shape.
- **Don't author `whole`-with-base patches** (full-replace of a live fact — rejected; a live section is a `block` patch, a new section is an `insert`).
- **Don't use `insert` to edit an existing section** — `insert` only appends a section the fact lacks.
- **Don't `write`/`set`/`delete` a fact directly** — refused; author a patch (SKILL.md §The shared contract).
- **Don't hand-author a membership row** — set `parent:`; the tool synthesizes it.
- **Don't expect a body edit to advance status** — status moves only through `accept`, the auto-done latch, or `supersede` (SKILL.md §ADR status set).
