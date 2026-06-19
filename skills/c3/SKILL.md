---
name: c3
description: >
  Triggers on /c3 or architecture questions in projects with .c3/ directory.
  Phrases: "adopt C3", "onboard", "where is X", "audit architecture", "check docs",
  "add component", "implement feature", "what breaks if I change X", "add ref",
  "coding standard", "edit the canvas", "add prd/user-story".
  Ops: onboard, query, audit, change, ref, rule, canvas, sweep. The CANVAS is your
  architecture's own vocabulary; onboarded facts FREEZE; work advances by change-units.
  Classifies intent, loads the reference, runs the CLI.
---

# C3 — one model, three acts

C3 is your architecture's own vocabulary, frozen into shared truth, that work edits only through reviewed change-units.

1. **Shape the model, freeze the facts.** Descend the domain top-down — draft the **facts** the work needs and wire them as you go; the lean **canvas** (sections + typed columns each entity carries *here*) is already in place, taking a custom fact-type only where the domain needs one. Then flip the gate and the facts **freeze**: shared truth, never hand-edited again.
2. **Change-units drive progress.** Work advances by **change-units** — an ADR plus its patch folder, one atomic saga: declare intent, the tool keeps the result integral (membership, citations, gates), flip it all-or-nothing. Frozen facts change *only* through this merge.
3. **The canvas grows with the need.** When work outgrows the model, **raise the canvas** (climb a rung) and migrate every fact up — completeness is never relaxed.

Build the model → freeze the facts → change-units drive the work → the canvas grows with the need.

## CLI handle

Packaged with this skill at `<skill-dir>/bin/c3x.sh`. Create a session-local handle once, then use it for every command:

```bash
c3() { C3X_MODE=agent bash <skill-dir>/bin/c3x.sh "$@"; }
```

`C3X_MODE=agent` → TOON output (~40% fewer tokens) and `help[]` hints appended to each result — follow them. The packaged CLI is the single source; this skill is the **router**: classify the intent, load the reference, run `c3`, follow the output. It names every gate and teaches no procedure — procedure lives in the references.

## Intent Classification

| Keywords | Op | Reference |
|----------|----|-----------|
| adopt, init, bootstrap, onboard, "create .c3" | **onboard** | `references/onboard.md` |
| where, explain, how, diagram, trace, "show me", "what is", "what handles" | **query** | `references/query.md` |
| audit, validate, "check docs", "is the doc intact" | **audit** | `references/audit.md` |
| add, change, fix, implement, refactor, remove, design | **change** | `references/change.md` |
| pattern, convention, "create ref", "update ref", standardize | **ref** | `references/ref.md` |
| "coding rule", "coding standard", "split ref into rule" | **rule** | `references/rule.md` |
| "edit the canvas", "change the shape", "what sections does X have", "add a doc type", "raise the bar", "add prd/user-story" | **canvas** | `references/canvas.md` |
| impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |
| recipe, "trace end-to-end", "cross-cutting flow" | **query**(read) / **change**(create) | `references/query.md` / `references/change.md` |

## Dispatch

1. Classify the op (ambiguous → `AskUserQuestion` with options).
2. Load `references/<op>.md`.
3. Run the CLI (Task tool for parallelism), follow `help[]`.

**Precondition — read-only fast path.** For conceptual discovery ("where is X", paraphrases) start with `c3 search "<question>"`; for known files/globs use `c3 lookup <file>`; for known ids/sections use `c3 read <id> --section <name>`. Reach for `c3 list` / `c3 check` only after a search miss, suspected drift, a topology-wide inventory, or an explicit audit. **Never Read/Glob/Edit `.c3/` instance files** — they are CLI-only; raw access bypasses the seal and goes stale. (Canvas *definitions* at `.c3/canvases/<type>.md` are the exception — user-owned markdown.) Missing `.c3/` → **onboard**.

## The shared contract

These rules are stated once here; every reference cites them.

**Frozen facts.** A *fact* is any entity whose canvas declares no `status:` set — `system`, `container`, `component`, `ref`, `rule`, `recipe`, `pm-requirement`, `user-story`. The moment a fact carries a body it is frozen: `c3 write`, `c3 set`, and `c3 delete` on it are **refused** (the guard keys on the first arg; an unknown type is treated as frozen). The refusal names the only legal path: *"<id> is a fact — facts are frozen and change only through a change-unit."* A fact changes **only** by authoring patches in a change-unit and running `c3 change apply`.

*Exempt from the freeze:* `c3 add` (creating a new fact is unguarded), the first `write` that authors a never-bodied fact, editing a **change-doc** (`adr`/`prd`/`atomic-design-change` — they declare `status:`), editing a **canvas definition** (user-owned), and **`c3 set <id> codemap '<glob>'`** — the external code binding is verified, not frozen, so the map is live-maintained (deliberate re-bindings still ride a change-unit as a `.codemap.md` carrier).

**Membership is by construction.** A parent's membership table is synthesized by the tool from its children's `parent:` links, on every path that changes parentage (`c3 add`, `change apply`, `check --fix`). Leave the row to the tool — never hand-author or hand-edit it. (Set `parent:`; the row appears.) Parentage is **not** a graph edge: `parent:` sets the child's `parent_id` and the synthesized membership row — a separate axis from *wiring* edges (`uses`, citations, `edge<>` columns), which `c3 graph` shows. A reparent updates `parent_id` + both parents' membership; it does not touch wiring.

**A fact is always complete to its rung.** A canvas is a **rung** — a complete contract for one complexity *level*, not a target to fill in over time. A fresh init's canvas is deliberately lean (rung-1); the deeper sections a complex project needs are a *higher* rung, not a hole. Completeness is never relaxed. To grow, **climb a rung**: raise the canvas, then migrate every affected fact up to the new contract, completely — integrity forbids a fact straddling two rungs.

**ADR status set:** `[open, accepted, done, superseded]`. (`c3 add adr` stamps `proposed` — the legacy synonym for `open`; `accepted` auto-latches to `done` when its After-cites resolve fresh.) Terminal change-docs (`done`/`superseded`) are content-frozen historical records, **exempt from `c3 check`**. `list`/`check` exclude ADRs by default; `--include-adr` to include.

## Command table

The packaged CLI is the catalog — `c3 <cmd> --help` is authoritative. The change-unit gate stack (the load-bearing flow) lives in `references/change.md`.

| Command | Purpose |
|---------|---------|
| `list` | Topology with counts + coverage (`--flat`, `--compact`) |
| `check` | Validate facts against their canvas + code-map bindings + consistency (`--fix`, `--only`, `--include-adr`). Code-binding mismatch is a WARN; `--strict-codemap` gates it |
| `repair` | Rebuild the disposable cache from canonical `.c3/` and reseal (after a branch switch / selective merge) |
| `search <query>` | Concept → entities by semantic + keyword + graph signal |
| `lookup <file-or-glob>` | File/glob → component(s) + refs |
| `read <id>` | Entity content (`--full`; `--section <name> --cite` emits the patch base anchor) |
| `graph <id>` | Relationship graph (`--depth`, `--direction forward\|reverse`, `--format mermaid`, `--unit <adr-id>` previews staged patches) |
| `add <type> <slug>` | **Create** a fact (body via stdin or `--file`; `--container`, `--feature`). The unguarded create path |
| `canvas <list\|read\|add\|write>` | Manage canvas definitions (user-owned shape, at `.c3/canvases/`) |
| `schema <type>` | Render a canvas's sections/columns/REJECT-IF (leads with the rejection contract) |
| `write <id>` / `set <id> <field> <val>` / `delete <id>` | Direct edits — **refused on a frozen fact** (see the contract). For change-docs, canvas bodies, and `set codemap` only |
| `change <new\|view\|status\|accept\|apply\|rebase\|scaffold\|inspect>` | The change-unit saga — the **only** way to mutate a fact. `apply` runs the gate stack atomically; `rebase` emits the drift bundle for drifted patches; `scaffold` stages a rung-climb; `inspect` surfaces a touched fact's derivation obligations for its `.inspect.md`. See `references/change.md` |

Missing a packaged operation → STOP, tell the user. No file-tool workarounds.
