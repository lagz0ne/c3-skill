# Onboard Reference

Onboarding is **Act 1**: shape the model, then freeze the facts. You run **one complete change-unit cycle** — shape the canvas to fit the project, discover the topology, author the *whole* architecture into the genesis ADR as create-patches, then **flip once**. The flip materializes every fact atomically, canvas-validated, all-or-nothing — and from that moment the architecture is **frozen shared truth**, changed only through a future change-unit.

`c3 init` already did the setup: it created the system `c3-0` and the genesis ADR `adr-00000000-c3-adoption`. You author *into* those.

**Precondition.** `c3 list` returns facts → already onboarded; offer re-onboard or redirect to audit/query.

---

## The cycle

### 1. Shape the canvas to fit *now*

A canvas is a **rung** — a complete contract for one complexity level. `c3 init` seeds **lean rung-1** canvases; that is the default. Keep them, or trim/enrich with `c3 canvas write <type> --file` **only if this project warrants it** — do not pre-build deeper sections a complex project would need later. Read what each requires with `c3 schema <type>`.

Do this **before** authoring any fact: every fact is validated against its canvas at the flip, so the shape must be right first. The rung model and when to climb live in `canvas.md`.

### 2. Discover the topology

Conversation first — discuss the idea and where the seams fall. Then inventory, in this order:

| Layer | Is | Numbering |
|-------|----|-----------|
| **Container** | deployment/runtime boundary | `c3-1`, `c3-2`, … |
| **Component** | a unit inside a container | Foundation `c3-N01`–`c3-N09` (others depend on it) / Feature `c3-N10`+ (business logic) |
| **Ref** | a rationale-bearing convention — "would this change if we swapped the underlying tech?" | `ref-<slug>` |
| **Rule** | an enforceable standard — "a coding standard or constraint, not a pattern choice?" | `rule-<slug>` |
| **Recipe** | an end-to-end flow **no single component owns** — it crosses containers | `recipe-<slug>` |

A component that can name a concrete file is Foundation or Feature; a pure convention with no file is a ref. Trace the two or three cross-container operations the system is *for* (e.g. receiving→putaway, reservation→pick→pack→ship) as recipes — that is what makes behavior legible across container boundaries. Use `AskUserQuestion` for gaps.

### 3. Author the whole architecture as create-patches

Author **into the genesis ADR**: one `<seq>-<slug>.patch.md` per fact, staged in `.c3/changes/adr-00000000-c3-adoption/`. Each is a **create-patch** — scope `whole`, **no base**, with `type:` and `parent:` in the frontmatter and the canvas-correct body. Author each body to its canvas (`c3 schema component`, `c3 schema ref`, `c3 schema rule`, …, never a remembered section list); any table, mermaid, or code fence **must** go through `--file` — inline strings corrupt quoting.

**You pick the ids — a create-patch's `target:` *is* the entity id** (no auto-numbering on this path). Follow the convention above; avoid slug ids like `web`/`api`, which break the `c3-N` reference scheme and mangle filenames.

Nothing materializes yet. The work is interruptible and resumable — the staged patches persist on disk (`check` exempts `.c3/changes/`); the ADR body carries the narrative.

The system `c3-0` already exists from init — author its body directly **before** the flip (it is a bodyless fact in its creation window; editing closes once it has a body).

> **Membership is by construction — never author a membership row.** Leave every parent's table a **header only** (its canvas columns from `c3 schema`, no data rows): `c3-0`'s `## Containers`, each container's `## Components`. Set each child's `parent:` and the flip synthesizes every row from those edges, in the same pass. No placeholders, no "create then read", no membership patch. Mechanics → `change.md`.

> **`c3 add` is the unguarded create exception**, not the primary path here. It auto-numbers and materializes one fact immediately — fine for a one-off, but the genesis ADR is the demonstration *and* the durable record of how this architecture was built. Author through the create-patches.

### 4. Flip — freeze the facts

```bash
c3 change view adr-00000000-c3-adoption    # preview every staged create-patch
c3 change apply adr-00000000-c3-adoption   # materialize all-or-nothing; facts are now frozen
```

One atomic, canvas-validated transaction: every fact validates or nothing lands. After the flip the containers/components/refs/rules exist as **frozen facts** — editing any of them now rides a change-unit (`change.md`), never `c3 add`/`c3 write`/`c3 set`/`c3 delete`.

**One field survives the freeze:** `c3 set <id> codemap '<glob>'`. Code churns independently of the design, so its mapping is verified live, not frozen — set it per component/ref/rule after the flip and spot-check with `c3 lookup 'src/**'`.

### 5. Close the change-unit

The genesis ADR's Affected Topology cites were authored as `N.A` — the facts didn't exist yet. Now they do:

```bash
c3 read <id> --cite                        # refresh each After-cite with the real handle
c3 change inspect adr-00000000-c3-adoption # surface derivation obligations
# author <seq>.inspect.md per the inspect output
c3 change accept adr-00000000-c3-adoption  # the one stored human judgment → accepted
c3 check --fix                             # latches accepted → done when After-cites resolve fresh
```

`done` is **earned, never typed** — the latch actualizes `accepted → done` only once the refreshed After-cites resolve, proof the architecture actually landed. The gate stack `apply` runs (drift + canvas + retire + inspection) and the ADR status set are defined in SKILL.md and taught in `change.md`; cite them, don't re-derive them here. Onboarding ends having completed one full change-unit cycle.

---

## The one gate list

```
- [ ] Canvas sized to the project (lean rung-1 default; deeper sections only if warranted)
- [ ] Topology discovered: containers, components (with category), refs, rules, recipes
- [ ] Every fact a create-patch in .c3/changes/adr-00000000-c3-adoption/ (parent: set, membership headers only)
- [ ] Flip applied — facts materialized and frozen (change apply)
- [ ] Code-map set per component/ref/rule; c3 lookup 'src/**' resolves
- [ ] c3 check passes; coverage acceptable (or exclusions documented)
- [ ] Audit passes (audit.md)
- [ ] Genesis ADR: After-cites refreshed → inspected → accepted → latched done (check --fix)
```

A failed gate sends you back to canvas-shape (1) or topology (2), not forward.

---

## Post-onboard: inject CLAUDE.md

```markdown
# Architecture
This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits, file context -> use the C3 skill.
Operations: query, audit, change, ref, rule, canvas, sweep.
File lookup: `c3local lookup <file-or-glob>` maps files/directories to components + refs.
```

Then point them at `c3 --help` and `c3 <command> --help` for the full command set; SKILL.md routes intent to each operation.
