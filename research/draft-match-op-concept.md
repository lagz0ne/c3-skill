# Concept — `eval`: a fact + its uncontrolled context = a topic with a rubric

> Working name `eval`. Supersedes the codemap + the switch-bound up-V. The realisation
> (2026-06-20): this is the **skill-eval harness turned inward on the user's own facts** —
> same four-part shape every `topics/*/` already has. Code is one op, not the feature.

## The shape
C3 **contributes a fact** (a frozen claim) and **associates it with context it does not
control** — code, *or* a parallel C3 doc set, *or* an artifact. Fact + context are wrapped
as a **topic with a rubric** that **answers a question**: *does the external still hold the
claim?* `eval` runs that topic on a CI-ready cadence — not at authoring.

This is not a new machine. It is exactly the structure of `research/eval/skill-eval/harness/topics/*`:

| Topic part | In `topics/qa-coverage/` | In the `eval` op |
| --- | --- | --- |
| the **claim** | a `rubric-notes.md` invariant line | the frozen fact |
| **uncontrolled context** | the agent's `.c3/` workspace | the external lane (code / docs / artifact) |
| the **rubric** | the invariant + falsifier table | the fact's eval-spec |
| **mechanical evidence** | the reviewer runbook (`graph … reverse`, `read`) | the `c3x` helper |
| the **verdict** | the judge (K=3, anti-gaming) | CI / a human |
| the **question** | "is coverage complete? did edges form?" | "does external still ≡ claim?" |

The fact is frozen, so the *claim* side of the topic is stable; only the two lanes move, on
their own clocks. That stability is the value the dogfood proves: a frozen fact is the only
thing that makes two independently-moving lanes **checkable against each other**.

## An eval is a one-off — solidity = frozenness of both sides
An `eval` is **true/false at one instant**, for exactly the pair of states it measured. It
**guarantees nothing forward**: the moment either side moves, the verdict is stale. Its solidity
is bounded by the **less-frozen side** — a green eval over a frozen claim and *living code* is
only "true at this SHA", never a standing guarantee that code conforms. A verdict becomes
*durable* only when **both** sides are frozen (pinned SHA / released artifact / frozen parallel
doc), because then neither side can move. The codemap and the up-V sold this snapshot as a
standing guarantee — it never was; code moves the commit after the switch.

So **the op decides the regime.** Living code → the external can't be pinned, so a code-op eval
is *inherently* one-off; you accept the momentariness and re-run on CI. A released artifact or a
frozen parallel doc *is* pinned, so those ops can be durable. Momentariness is not a flaw to
fix — it is a property you read off the op.

**Dogfood, 2026-06-20.** Two findings on c3-design's own `.c3/`:

- *Mechanical sweep across every codemap'd fact* — **10 facts + 1 ref drifted**: their declared
  surface points at deleted/moved code (`c3-103` templates, `c3-111` numbering, `c3-113`
  index/check_legacy, `c3-116..120` removed/merged commands + marketplace, `c3-215` a removed
  doc). The codemap claims to be a *continuous* binding but nobody re-ran it, so it silently
  rotted across the CLI's evolution. Ten stale entries on `main` are the proof the "standing
  guarantee" was fiction — the one-off eval on CI cadence would have caught each.
- *Judgement, c3-107 store-lib* — frozen claim enumerates "entity, relationship, *changelog*,
  codemap, hash, node, version storage"; this session deleted `store/changelog.go`. One-off
  eval: **mechanical HOLDS** (glob `cli/internal/store/**` still resolves) yet **judgement
  DRIFT** (the changelog responsibility is gone, 0 refs). c3-107 isn't in the mechanical list —
  so true drift exceeds ten: mechanical catches surface-deleted, only judgement catches
  responsibility-deleted-while-surface-remains. The verdict is a snapshot of `(db80187…,
  this-tree)`, nothing more.

(Real residue: these frozen facts are stale and need change-units to re-state — the dogfood
found genuine bugs in our own docs, not just demonstrated the concept.)

**Loop closed on c3-107 (find → fix → re-eval), and it surfaced more.** Re-stated c3-107's Goal
via a real change-unit (`block` patch dropping "changelog"); the sealed `## Goal` now matches the
code (0 changelog refs) → judgement **HOLDS**. Three findings fell out of *doing* it:
1. **The up-V taxes a doc-correctness fix.** c3-107 is code-bound, so `apply`'s inspection gate
   refused until I authored a territory-grounded `*.inspect.md` — for *removing a stale word that
   reconciles the doc to code that already dropped it*. Exactly the ceremony the eval model moves
   off the apply switch and onto a one-off CI eval.
2. **The mutation machinery itself drifts (real bug).** The `block` patch updated the sealed
   `## Goal` section but **not** the denormalized frontmatter `goal:` — `apply.go` never calls
   `syncGoalFromNodes` the way `bridge.go` `WriteEntity` does. `check` stays green (seal covers
   nodes, not the frontmatter copy), so `lookup`/`read` show the stale goal invisibly. The one
   legal mutation path is the one that desyncs derived state — a "divergence-by-construction"
   violation the project's own principle forbids.
3. **`change accept` errors on an ADR-less unit** ("no rows in result set") — a `change new` unit
   with no backing entity has no status row; `apply` still flips. Rough edge.

## "Op" = an external-kind, not a verb
A topic is bound to *one kind* of uncontrolled context. Each binding is an **op**:
- **code op** — external is code. The rubric asks "does the code implement the claim?"
  (This is everything the codemap did — one op, never a general feature.)
- **docs op** — external is a parallel C3 doc set. Rubric: "do the two lanes agree?"
- **artifact op** — external is a build output / spec target. Rubric: "does the artifact match?"

Code stops being privileged. The first op to *ship* may be the code op (it replaces the
codemap), but the machine is general from the start, exactly as the harness is general over
whatever workspace a topic points at.

## The rubric has two registers (both already in the harness)
- **Mechanical** — a deterministic `c3x` helper: the external the fact names resolves/exists,
  structural signals line up. This is the reviewer runbook's `graph`/`read`/`canvas read`
  rows, computed instead of eyeballed. Equality (`external ≡ claim`) is its simplest line.
- **Judgement** — agent/human: does the external's behaviour match the claim's prose? This is
  the harness judge — the questions (`INV-EDGE-TARGETED`-style) no equality check can pose.

**`c3x` is a helper, never a gate.** It *computes* the mechanical lines and *surfaces* what to
judge; the conformance **decision** lives in CI / a human / the eval verdict — never in a
`c3x` exit code. No `--strict` flip, no apply-time inspection gate. (This is not a new rule to
enforce; the harness already splits runbook-evidence from judge-verdict this way.)

## What this removes (conscious reversal of recent work)
- **The codemap** (`code_map` table, `c3 set codemap`, `codemap_introspect`) — it was the code
  op hard-wired as a feature. Replaced by the neutral eval-spec, which expresses the same
  fact→code binding *and* fact→docs / fact→artifact.
- **The switch-bound up-V** (`inspectionGate`, the 5th apply gate, 11.2.0) — the up-V stops
  gating `change apply`. The apply switch keeps only the **down-V** (doc valid + sealed).
- **`check --strict-codemap`** gating — `check` stops being an enforcement point for code↔doc.

## What survives — as helpers/format inside `eval`, none as gates
- glob/locator resolution (was `codemapGlobResolves`) → a mechanical rubric line.
- the attestation format + anti-rubber-stamp floor (was `InspectCarrier` + `evidenceCitesTerritory`)
  → the judgement rubric's evidence shape (the harness's "must-have evidence"), now a CI
  artifact, not a flip gate.
- the **falsifiability stamp**: an `eval` verdict records `{fact RootMerkle, external state hash
  e.g. git-SHA, mode, verdict}` — a green check is reproducible and you know *which* claim was
  matched against *which* external state. The fact stays pure (stamp lives in the verdict).

## Coverage (the rubric's scope), general
- **Inclusion** = the external surface the eval-spec claims to govern (the *declared* set, not
  the whole repo — avoids the false-hole flood the deleted coverage walker had).
- **Exclusion** = per-fact carve-outs (a fact may exempt part of its own declared surface).
- **Two failure modes, reported separately:** DRIFT (the fact's external is gone/changed) vs
  HOLE (declared surface matched by no fact). Neither is a CLI gate; both are CI signals —
  the harness's "orphan left standing" failure mode, generalised.

## Open / to design
1. **The eval-spec on a fact** — the schema that replaces `codemap:` and *is* the per-fact
   rubric: external kind (code|docs|artifact|…), the locator, inclusion/exclusion, and the
   mechanical/judgement lines. General over external kinds.
2. **The first ops to ship** — code op (replaces codemap) first; docs op + artifact op define
   the generality. Which two prove "general from the start"?
3. **The name** — `eval` vs `match`/`prove`/`conform`.
4. **The migration** — c3-design's own facts move off `codemap:` onto the eval-spec; the apply
   gate stack drops to four; `check` drops the strict-codemap arm. (Reverses 11.2.0 surface.)
5. **What "CI-ready" emits** — a verdict report/artifact CI consumes and a human reviews, since
   `c3x` itself does not gate. Likely the harness's own run-record shape.
