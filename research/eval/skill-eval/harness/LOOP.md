# Blindbox Growth Loop — grow-warehouse-system

Fixed blind instrument: **codex (gpt-5.5)**, `--auth session`, `--run-timeout 900`,
in `bwrap` isolation. Skill packet rebuilt each run from `skills/c3/` (SKILL.md +
audit/canvas/change/onboard/query). CLI binary held fixed at the freshly-built
`c3x-11.1.0-linux-amd64` unless a round explicitly fixes a CLI defect (noted).

Scored against `harness/rubric.md` (G1–G10, /30). Pass bar: ≥24/30 with G1≥1,
G2≥2, G6≥3, G7≥3, and no false "verification passed" claim. Scored against the
**actual workspace `.c3/` + command transcript**, not the agent self-report.

## Score trajectory

| Round | Total | G1 | G2 | G3 | G4 | G5 | G6 | G7 | G8 | G9 | G10 | Pass | Top fix landed |
| --- | ---: | -- | -- | -- | -- | -- | -- | -- | -- | -- | --- | --- | --- |
| R0 baseline | **28/30** | 2 | 3 | 3 | 2 | 3 | 4 | 3 | 3 | 3 | 2 | ✅ | (none — baseline) |
| R1 c3-0 fix | _run truncated_ | — | — | — | — | — | — | — | — | — | — | n/a | CLI: c3-0 creation window (validated; run cut at 900s cap) |
| R2 climb-order | **29/30** | 2 | 3 | 3 | 2 | 3 | 4 | **4** | 3 | 3 | 2 | ✅ | seed reorder + skill: higher-rung sections last |
| R3 recipes | **30/30** | 2 | 3 | 3 | **3** | 3 | 4 | 4 | 3 | 3 | 2 | ✅ | skill: recipe discovery for cross-container flows |
| R4 ID-churn | **30/30** | 2 | 3 | 3 | 3 | 3 | 4 | 4 | 3 | 3 | 2 | ✅ | skill: parent-table from real ids (+1800s → **run completes**) |
| R5 correctness | **30/30** | 2 | 3 | 3 | 3 | 3 | 4 | 4 | 3 | 3 | 2 | ✅ | skill: c3-0 parent-table + cascade vocab (**cleanest: full 3-section climb, no c3-0 churn**) |

**Run hygiene (from R1):** the R1 codex agent hit the 900s cap because a parallel Codex
*review* was contending for the same backend. Fix: run blindbox **alone**, cap **1200s**.

## R0 baseline — evidence

Run: `runs/baseline-r0-codex.{md,stderr.txt}` + `.workspace/.c3/`. Ground-truth
`c3 check`: **in sync**, 4 containers / 8 components / 3 refs, only 3 WARNs (all c3-0).

- **G6 4/4 — the laddered climb is real.** Agent went lean rung-1 → raised the
  component canvas (made Foundational Flow / Business Flow / Change Safety required)
  → `change scaffold` → filled every component → `change apply` → check. Verified
  c3-201 carries genuinely filled climb tables (not empty templates). This is exactly
  the v11.1 flow, executed blind.
- **G8 3/3 — change discipline held.** Genesis-ADR change-unit for the flip + the
  climb; never hand-edited a frozen instance (tried `write c3-0`, got refused, pivoted
  to the change flow — did not bypass).
- **G9 3/3 — honest verification.** Ran `check`, reported the c3-0 warnings and the
  blocked-c3-0 caveat truthfully.

### Failures traced

1. **c3-0 (system context doc) is unfillable on fresh init — CLI defect + skill
   contradiction.** Cost ~1pt G7 (system doc left ragged: missing Goal / Containers /
   Abstract Constraints) and wasted agent effort.
   - `c3 write c3-0 --section Goal` → **refused** ("c3-0 is a fact — frozen"). c3-0 is
     born frozen at `c3 init`. → contradicts onboard.md §1.1/§1.3 which tell the agent
     to author c3-0 directly "before the flip".
   - `c3 read c3-0 --cite` → "c3-0 has no versioned content hash for citation" → can't
     anchor an `insert`.
   - `change scaffold` emits the c3-0 insert patch with empty `base: c3-0@v0:sha256:`
     → apply fails: `yaml: line 3: mapping values are not allowed` (malformed YAML).
   - `whole` no-base patch for c3-0 → **apply hangs** (rc=143 on a 25s timeout),
     pristine init. c3-0 stays empty.
   - Net: no working path to give c3-0 a body with the current binary.
2. **G4 2/3 — cross-container flows are embedded, not traced.** Concepts (3 refs) and
   per-component Business Flow are good, but no explicit `recipe` artifacts tracing a
   complex operation (reservation→pick, receiving→putaway) across containers.

## Round log

### R1 — CLI fix: c3-0 creation window (binary rebuilt; skill unchanged)

**Change (one, for clean attribution):** `cmd/freeze.go GuardFactMutation` — a frozen-TYPE
fact with no authored content yet (`RootMerkle == ""`, the inverse of the born-sealed
invariant) is in its **creation window**: the first direct authoring is allowed; once it
carries content the freeze engages. This makes onboard.md's documented `c3 write c3-0`
path *actually work* — previously c3-0 was born frozen-and-empty with no authorable path
at all. Regression test `TestGuardFactMutation_CreationWindow` + realistic refuse-test
(seed c3-101 a body). Full `go test ./...` green. Binary rebuilt to
`c3x-11.1.0-linux-amd64`; this is the instrument for R1+.

Verified on a clean init: `write c3-0 --file sys.md` → "Updated c3-0"; `check` in sync,
the 3 c3-0 "missing required section" warnings gone; `set c3-0 goal` refused after content
(window closed). No skill edit needed — the CLI now backs the existing skill.

**Run outcome:** truncated at the 900s cap (parallel Codex review contended for the
codex backend → agent paced slower). c3-0 fill validated regardless (ground truth: c3-0
carries a full goal + body, the 3 missing-section warnings are gone). NOT fairly scorable
as a completed run; the c3-0 *fix* carries into R2's binary.

Run: `r1-c30fix-codex`. Live-validated mid-run: transcript shows
`c3 write c3-0 --section Goal` then `--file system.md` → **"Updated c3-0 (system)"** — the
documented path now works blind, no flailing. Score: _pending run completion_.

**Codex cross-review hardened the fix (folded into R1, like addressing review comments
before merge).** The first cut keyed the window on `RootMerkle == ""` and exempted
write/set/wire/delete — Codex flagged 3 valid holes: (1) empty RootMerkle ≠ never-authored
(legacy migration / frontmatter-only import can have authored frontmatter+edges with an
empty body root); (2) set/wire don't add body content so the window never closes for them;
(3) `WriteEntity(id,"")` can re-empty a fact. Tightened to: **`write` only**, gated on
**zero body nodes AND `Version == 0`** (set/wire/delete stay frozen even while bodyless).
New `TestGuardCanonicalMutation_CreationWindow` asserts the matrix; full `go test ./...`
green. c3-0 authoring is identical under both cuts (c3-0 is bodyless + v0), so R1's run
stands; the hardened binary is the instrument from R2 on.

### R2 — climb section-ordering (seed canvas reorder + skill note; binary rebuilt)

**Failure surfaced by R1's run (real, not truncation):** the climb left **34 check errors** —
every migrated component had `sections out of order: expected Business Flow before
Governance/Contract` and `Change Safety before Derived Materials`. Root cause is a
**seed-canvas defect**, not agent error: the default `component` canvas ordered its three
higher-rung *optional* sections **interleaved** (Foundational Flow / Business Flow at
positions 4–5, Change Safety at 8) among the required ones — but `change scaffold`→`insert`
**appends** new sections at the body's **end**. So climbing the seed as-authored always
yields out-of-order sections. (R0 only passed because *that* agent happened to rewrite the
canvas with new sections last — luck, not skill.)

**Fix:**
- `cli/internal/schema/builtin/canvases/component.md` — reordered so required sections come
  first (Goal, Parent Fit, Purpose, Governance, Contract, Derived Materials) and the three
  higher-rung optional sections come **last** (Foundational Flow, Business Flow, Change
  Safety). The append-only climb now produces canvas-valid order *by construction*. Seal
  regenerated (`547cca9e…`); `TestEmbeddedCanvases_ReSealClean` + full `go test ./...` green.
- `references/change.md` §Climbing a rung — added the rule: `insert` appends, so order the
  newly-required sections **last** when raising a canvas (guards an agent that rewrites the
  canvas from re-interleaving).

Run: `r2-climborder-codex` (alone, 1200s). Expected: climb migration check-clean (no
out-of-order), c3-0 filled, run completes. Score: _pending_.

**Ordering defect hit R0 too (not just R1).** R0's transcript (stderr line ~32328) shows
the *baseline* also produced `sections out of order` errors mid-run — it just spent extra
gardening to fix them before finishing. So the seed-canvas reorder removes wasted effort on
**every** run, not only the ones that fail. Other seed canvases audited: container / system /
ref / rule / recipe / pm-requirement / user-story all already order optional sections last —
`component` was the sole offender.

**Friction backlog (surfaced by R0/R1 transcripts; candidate rounds):**
- **Genesis ID-churn** — agent writes predicted component IDs (c3-101…) into container
  `Components` tables, but the CLI allocates *feature* IDs (c3-110, c3-210…), forcing it to
  re-patch every parent table ("replace stale placeholder IDs with actual C3 IDs"). Skill
  fix candidate: create components first, then author parent tables from real `c3 list` IDs.
- **Layer disconnect** — components created but missing from the parent's `Components` table
  → top-down amend via a change-doc.
- **Ungrounded evidence** — bare `c3 check` in an `evidence` column; needs a grounded
  ref/command-with-target/entity id.
- **Placeholder language** detected in filled sections.
- (Deeper, deferred) a position-aware `insert` would remove the canvas-ordering constraint
  for custom canvases entirely.

### R3 — recipe discovery for cross-container flows (skill-only)

**Gap (G4 = 2/3 in R0 and R2):** agents documented flows *inside* per-component `Business
Flow` sections but created no standalone cross-container artifact tracing a whole operation.

**Fix:** `references/onboard.md` §0.6b "Recipe Discovery" — a `recipe` captures an end-to-end
operation that *no single component owns* (it crosses containers and hands off between
components); a component's `Business Flow` is its local slice. Guides creating a recipe for
the 2–3 most important cross-container operations (reservation→pick→pack→ship, etc.).

Run: `r3-recipes-codex`. Result: **30/30** (G4 → 3/3). Created **4 recipes**
(reservation-pick-pack-ship, receiving-putaway, cycle-count-adjustment,
returns-quarantine-reporting), each naming the cross-container path. Climb done + check
**100% in sync** (30 entities). Report prose truncated at the 1200s cap, but all work
landed (ground-truth verified) — bumping to 1800s for R4+.

### Reckoning: the rubric is saturated

R3 hit **30/30** on a clean run — the G1–G10 blindbox rubric no longer has score headroom,
so "improve the result" can't mean the number past here. Two *real* weaknesses persist that
a single clean-run score does not capture, and they are the honest targets for R4–R5:
1. **Runs truncate** (R1, R3) — the task is long; the agent does avoidable extra work.
2. **Genesis ID-churn** — agents predict component ids in parent `Components` tables, the CLI
   allocates others (feature 10+), forcing a re-patch of frozen containers (wasted work that
   *lengthens* runs → feeds problem 1). Also the related **layer-disconnect** (components
   missing from parent tables until patched).

R4–R5 target reliability/efficiency (fewer self-inflicted re-patches, runs that complete in
budget), tracked via secondary signals (run completion, count of mid-run gardening of
self-inflicted errors), not the capped G-score.

### R4 — genesis ID-churn / parent-table integrity (skill-only; 1800s cap)

**Fix:** `references/onboard.md` §1.2 — parent `Components` tables must list the ids the CLI
*allocates* (foundation `c3-<N>01+`, features `c3-<N>10+`, creation order), not guesses;
preferred path is create components → `c3 list` → fill the table from real ids (one pass, no
reconciliation). Also corrected a stale "`## Dependencies` table" instruction (not a current
component-canvas section).

Run: `r4-idchurn-codex` (alone, 1800s). Target: run completes; **no** "unknown entity
reference" / stale-id re-patching mid-run; check clean. Score: _pending_.

**R5 candidate found while auditing for drift:** SKILL references an OLD component-canvas
vocabulary that no longer exists — `## Related Refs` / `## Related Rules` (ref.md, rule.md,
query.md) and `Dependencies` / `Code References` / `Container Connection` (change.md:232
cascade gate). The current component canvas cites refs/rules in the **Governance** table,
not those sections. In-packet instance = change.md:232; the bulk is in ref.md/rule.md
(out of the growth packet, but real drift). Candidate R5 = align the cascade-gate + citation
vocabulary to the live canvas.

**R4 result:** `r4-idchurn-codex` — **30/30, and the run COMPLETED** (1800s cap fixed the
truncation that hit R1/R3). check clean (24 entities); climb done (mid-run "missing
Foundational/Business/Change Safety" were transient during the raise). Component-level
ID-churn dropped sharply (report describes no component re-patching). **But the churn is
multi-level:** the agent authored **c3-0's `Containers` table with placeholders**
(`N.A - initial`) before containers existed, then patched to real ids — the same chicken-
and-egg one level up. R4's §1.2 fix covered container→component; the system→container level
remains. So R5 also generalizes the parent-table rule to c3-0.

### R5 — skill correctness pass (generalize parent-table rule + cascade-gate vocabulary)

Two skill-only correctness fixes:
1. **Generalize the parent-table rule to every level.** `c3-0`'s `Containers` table churns
   like a container's `Components` table. Since the creation-window keeps c3-0 writable until
   it has a body, the clean flow is to author c3-0's body **after** the containers exist, with
   real ids — onboard.md §1.1 / Stage-2 note.
2. **Fix the cascade-gate vocabulary drift** (change.md:232): the Contract Cascade Gate asks
   whether `Goal, Dependencies, Related Refs/Rules, Code References, or Container Connection`
   changed — all dead section names. The live component canvas is Goal / Parent Fit / Purpose /
   Governance / Contract / Derived Materials. Align it.

Run: `r5-correctness-codex` (alone, 1800s). **Result: 30/30 — the cleanest run.** Completed,
check 100% in sync (23 entities). The agent added all 4 containers *then* wrote c3-0's body
(transcript: containers at lines 1948–2346, `write c3-0 --file system.md` at 8636) — **zero
placeholder churn** (the only `N.A - initial` hit is my guidance text echoed in the packet).
Climb carried **all three** higher-rung sections (`Foundational Flow`, `Business Flow`,
`Change Safety`) across 9 components, correctly ordered, applied cleanly under a dedicated
climb ADR.

---

## Loop synthesis

**Trajectory: 28 → 29 → 30 → 30 → 30** (R0 baseline; R1 c3-0 fix validated but run truncated;
R2 ordering 29; R3 recipes 30; R4 ID-churn 30 + run-completes; R5 correctness 30, cleanest).

**Two genuine C3 product defects fixed (CLI), both surfaced by the eval and unfixable by prose:**
1. **The system context doc `c3-0` was unfillable on every fresh `c3 init`** — born frozen-and-
   empty, with no authorable path (write/set refused, no citable anchor, scaffold emits malformed
   YAML, whole-no-base apply hangs). Fix: a fact with no authored content yet is in its *creation
   window* — first `write` allowed, freeze engages once it has a body. Codex-reviewed and tightened
   (write-only, `Version==0` + zero nodes) so it can't become a freeze bypass. `cmd/freeze.go` +
   `TestGuardCanonicalMutation_CreationWindow`.
2. **The laddered climb produced check-failing docs by construction** — the seed `component`
   canvas interleaved its higher-rung optional sections, but `change scaffold`→`insert` only
   *appends*, so climbing always went out-of-order. (Cost R0 gardening effort; broke R1.) Fix:
   reorder the seed so higher-rung sections come last (`builtin/canvases/component.md`, seal
   regenerated). The other 7 seed canvases were audited clean.

**Five skill improvements** (all from observed gaps, none test-gaming): climb-ordering note;
recipe discovery for cross-container flows; parent-`Components`-table from real ids; c3-0
`Containers`-table from real ids (author after containers); cascade-gate vocabulary realigned to
the live canvas (Goal/Parent Fit/Governance/Contract/Derived Materials).

**Run hygiene:** run blindbox *alone* (a parallel Codex review starved R1's agent) and cap at
1800s (the task is long; R1/R3 truncated at 900/1200s). With both, R4 and R5 completed cleanly.

**The rubric saturated at R3 (30/30).** R4–R5 therefore targeted real weaknesses the single-run
score can't show — run reliability and self-inflicted ID-churn — verified via secondary signals
(run completion; zero placeholder re-patching in R5).

**Still weak / follow-ups (honest):**
- **ref.md / rule.md teach a dead vocabulary** — `## Related Refs` / `## Related Rules` body
  sections the current component canvas lacks (citations ride the `Governance` table). Out of the
  growth packet so the eval didn't score it, but an agent following them authors invalid sections.
  Needs a dedicated drift pass + a verify of the live edge-extraction mechanism.
- **Multi-level ID-churn is mitigated, not eliminated** — inherent to authoring a parent body that
  tables its not-yet-created children; the guidance defers parent-table authoring, but a
  create-order-aware tooling affordance would remove it.
- **Single run per config is noisy** — a strong baseline swings several points on G6/G7 by luck
  (R0 worked around the ordering bug; R1 didn't). Multi-run averaging would sharpen attribution.
- **Position-aware `insert`** would remove the climb-ordering constraint for *custom* canvases
  (the seed reorder only fixes the default).

**Release-readiness:** full `go test ./...` green; c3-design's own `c3 check` in sync (84 entities,
0 issues — the seed reorder only affects fresh inits). The two CLI fixes warrant a patch release.

**Logged for a separate follow-up (out of growth-packet scope, larger change):** ref.md and
rule.md teach authoring `## Related Refs` / `## Related Rules` body sections that the current
component canvas does not have (citations now ride the `Governance` table) — an agent
following them would author invalid sections. Worth a dedicated drift-cleanup pass + a verify
of the live edge-extraction mechanism.
