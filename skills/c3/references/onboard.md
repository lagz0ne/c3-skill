# Onboard Reference

Onboarding is **Act 1**: walk the domain, draft the facts, freeze them. You don't configure a canvas and then fill it — `c3 init` already seeded the **lean rung-1** canvas and created the system `c3-0` plus the genesis ADR `adr-00000000-c3-adoption`. You **descend the abstraction** — system → containers → components → the refs, rules, and recipes that govern them — drafting each fact into the genesis ADR and **wiring it as you go**. The canvas is already in place; it grows a custom type only when the domain needs one. Then you **flip once**: every fact materializes atomically, canvas-validated, all-or-nothing — and the architecture is **frozen shared truth**, changed only through a future change-unit.

**Precondition.** `c3 list` returns facts → already onboarded; offer re-onboard or redirect to audit/query.

---

## The cycle

### 1. Descend the abstraction — draft each fact, wire as you go

Conversation first — discuss the idea and where the seams fall. Then walk **top-down**, drafting each fact and wiring it to the facts it already touches:

| Layer | Is | Numbering |
|-------|----|-----------|
| **System** `c3-0` | the whole — its goal and abstract constraints | exists from `init`; author its body |
| **Container** | deployment/runtime boundary | `c3-1`, `c3-2`, … |
| **Component** | a unit inside a container | Foundation `c3-N01`–`c3-N09` (others depend on it) / Feature `c3-N10`+ (business logic) |
| **Ref** | a rationale-bearing convention — "would this change if we swapped the underlying tech?" | `ref-<slug>` |
| **Rule** | an enforceable standard — "a coding standard or constraint, not a pattern choice?" | `rule-<slug>` |
| **Recipe** | an end-to-end flow **no single component owns** — it crosses containers | `recipe-<slug>` |

A component that can name a concrete file is Foundation or Feature; a pure convention with no file is a ref. Trace the two or three cross-container operations the system is *for* (e.g. receiving→putaway, reservation→pick→pack→ship) as recipes — that is what makes behavior legible across container boundaries. Use `AskUserQuestion` for gaps.

**Wire as you descend.** A component **uses** a ref; a rule **governs** a component; a recipe steps through components — each citation *is* a graph edge (edge columns → `canvas.md`; citing → `ref.md`/`rule.md`). The graph forms **with** the facts, not after them — by the time you reach the leaves the topology is already connected.

**The canvas grows to the domain — you never pre-configure it.** The lean seed carries system/container/component/ref/rule/recipe; keep it. When the domain surfaces a fact-type the seed doesn't carry — a QA, PM, or design doc-type — define its canvas **in the flow** (`c3 canvas write <type> --file`), then draft its facts like any other. Don't pre-build deeper sections a complex project would only need later (the rung model → `canvas.md`). Read what any type requires with `c3 schema <type>`.

### 2. Each fact is a create-patch

Author **into the genesis ADR**: one `<seq>-<slug>.patch.md` per fact in `.c3/changes/adr-00000000-c3-adoption/`. Each is a **create-patch** — scope `whole`, **no base**, with `type:` and `parent:` in the frontmatter and a canvas-correct body (author to `c3 schema <type>`, never a remembered section list; any table, mermaid, or code fence **must** go through `--file` — inline strings corrupt quoting).

**You pick the ids — a create-patch's `target:` *is* the entity id** (no auto-numbering here). Follow the convention above; avoid slug ids like `web`/`api`, which break the `c3-N` reference scheme and mangle filenames.

> **Membership is by construction — never author a membership row.** Leave every parent's table a **header only**: `c3-0`'s `## Containers`, each container's `## Components`. Set each child's `parent:` and the flip synthesizes every row from those `parent:` links, in the same pass. (Parentage is hierarchy — a separate axis from the wiring edges above.) Mechanics → `change.md`.

> **`c3 add` is the unguarded create exception**, not the primary path here. It auto-numbers and materializes one fact immediately — fine for a one-off, but the genesis ADR is the demonstration *and* the durable record of how this architecture was built.

Nothing materializes yet — the staged patches persist on disk (`check` exempts `.c3/changes/`), so the walk is interruptible and resumable; the ADR body carries the narrative. Author `c3-0`'s body **before** the flip (it is bodyless in its creation window; editing closes once it has a body).

### 3. Flip — freeze the facts

```bash
c3 change view adr-00000000-c3-adoption    # preview every staged create-patch
c3 change apply adr-00000000-c3-adoption   # materialize all-or-nothing; facts are now frozen
```

One atomic, canvas-validated transaction: every fact validates or nothing lands. After the flip the facts are **frozen** — editing any of them now rides a change-unit (`change.md`), never `c3 add`/`c3 write`/`c3 set`/`c3 delete`.

**One field survives the freeze:** `c3 set <id> codemap '<glob>'`. Code churns independently of the design, so its mapping is verified live, not frozen — set it per component/ref/rule after the flip and spot-check with `c3 lookup 'src/**'`.

### 4. Close the change-unit

The genesis ADR's Affected Topology cites were authored as `N.A` — the facts didn't exist yet. Now they do:

```bash
c3 read <id> --cite                        # refresh each After-cite with the real handle
c3 change inspect adr-00000000-c3-adoption # surface derivation obligations
# author <seq>.inspect.md per the inspect output
c3 change accept adr-00000000-c3-adoption  # the one stored human judgment → accepted
c3 check --fix                             # latches accepted → done when After-cites resolve fresh
```

`done` is **earned, never typed** — the latch actualizes `accepted → done` only once the refreshed After-cites resolve, proof the architecture actually landed. The gate stack `apply` runs (drift + canvas + retire + inspection) and the ADR status set live in SKILL.md and `change.md`; cite them, don't re-derive them here. Onboarding ends having completed one full change-unit cycle.

---

## The one gate list

```
- [ ] Topology walked top-down: system, containers, components (with category), refs, rules, recipes — wired as drafted
- [ ] Canvas left lean (seed kept; a custom type defined only where the domain needed one, not pre-built deep)
- [ ] Every fact a create-patch in .c3/changes/adr-00000000-c3-adoption/ (parent: set, membership headers only)
- [ ] Flip applied — facts materialized and frozen (change apply)
- [ ] Code-map set per component/ref/rule; c3 lookup 'src/**' resolves
- [ ] c3 check passes; coverage acceptable (or exclusions documented)
- [ ] Audit passes (audit.md)
- [ ] Genesis ADR: After-cites refreshed → inspected → accepted → latched done (check --fix)
```

A failed gate sends you back to the walk (1), not forward.

---

## Post-onboard: the natural further run

The canvas is now **yours** — shaped by your domain, your abstraction levels, your wiring. So the ongoing operations read straight off it, no extra ceremony:

- **Fill** more facts → a change-unit (`change.md`); the canvas says what each one carries.
- **Use** what's there → `query.md` (`read` / `search` / `graph`); the facts are frozen shared truth — answer from them.
- **Progress** the work → change-units drive it (Act 2); grow the canvas a rung only when real pressure demands it (Act 3, `canvas.md`).

Inject `CLAUDE.md` so the project points here:

```markdown
# Architecture
This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits, file context -> use the C3 skill.
Operations: query, audit, change, ref, rule, canvas, sweep.
File lookup: `c3local lookup <file-or-glob>` maps files/directories to components + refs.
```

Then point them at `c3 --help` and `c3 <command> --help`; SKILL.md routes intent to each operation.
