I now have full ground truth. The dogfood `.c3/` already models C3-the-product as architecture (c3-1-go-cli container, c3-2-claude-skill container, refs, rules, recipes). The builtin canvas set proves C3 ships custom canvas types beyond pure architecture (`prd`, `user-story`, `pm-requirement`, `atomic-design-change`). This grounds the "C3-own-canvas move" precisely: define a `narrative` (or `story-act`) canvas type to hold the three-act design narrative as first-class facts, absorbing the prose that currently lives scattered across SKILL.md/README/onboard.

I have everything needed. Here is the synthesized rewrite plan.

---

# C3 DOCUMENTATION REWRITE PLAN

Verified against source: `help.go:387` (subcommands `new|view|accept|apply|status|rebase|scaffold|inspect`), `change.go:48-87` (gate order drift→canvas→retire→inspection), `destruction.go:11-50` (orphan/dangle refusal + "unless this same unit heals it"), `add.go:127-134` + `check_enhanced.go:252` (membership-by-construction on both paths), `freeze.go:74-104` (only `codemap` carve-out; refusal string), `init.go:54` (`adr-00000000-c3-adoption`, `c3-0`). All ten findings' load-bearing claims hold.

---

## 1. CANONICAL STRUCTURE — one owner per act/operation, zero overlap

| Surface | Owns (and ONLY this) | Acts | Routes/defers to |
|---|---|---|---|
| **SKILL.md** | The spine in prose (3-4 sentences) + Intent Classification table + Dispatch + ONE command table + ONE frozen-facts rule. The **router**. Names every gate; teaches no procedure. | frames 1+2+3 | every reference for steps |
| **references/canvas.md** | **What a canvas IS** (your architecture's vocabulary as a contract) + **why you RAISE it** (the rung). The **model** surface. | 1 (def) + 3 (why climb) | change.md for climb *steps*; `c3 schema` for column grammar |
| **references/onboard.md** | The **one Act-1 cycle**: shape canvas → discover topology → author whole architecture into the genesis ADR as create-patches → **flip → freeze**. | 1 | SKILL.md (commands), change.md (the unit), canvas.md (rungs) |
| **references/change.md** | The **change-unit saga** (Act 2) + the **climb steps** (Act 3 operational). Owns: the 4-gate apply stack, `inspect`/`*.inspect.md`, `rebase`, the destruction gate, membership-by-construction mechanics, the 3 patch carriers, climb sequence. The **operational home**. | 2 + 3 (steps) | SKILL.md for the frozen-fact *contract*, rung *definition*, status set (cite, don't re-teach) |
| **references/query.md** | **Reading frozen facts** — the read tool selector + ONE evidence rule + ONE causal-chain rule. The **read** surface. | 1 (read beat) | sweep.md (impact), SKILL.md+change.md (ADR status) |
| **references/audit.md** | **Integrity of the frozen facts**: seal/repair, structural (run `c3 check`), thin semantic layer. The **integrity** surface. Ends in PASS/WARN/FAIL. | 1 + 2 | ref.md/rule.md (compliance specifics), canvas.md (section contract), change.md (the only fix loop) |
| **references/ref.md** | **What makes something a ref** (rationale-bearing `## Why`) + the **Ref-vs-Rule Separation Test**. Nothing else. | 1 (a fact-type) | change.md (citing/editing/adoption — all of Act 2) |
| **references/rule.md** | **What makes something a rule** (enforceable standard: Goal/Rule/Golden Example) + create-whole + the one citation note. | 1 (a fact-type) | ref.md (Separation Test), change.md (the saga, adoption ADR, status latch) |
| **references/sweep.md** | **Pre-flight for a change-unit**: reverse-graph blast radius + **will the destruction gate refuse it**. The **impact** surface. Ends in Verification table + destruction-gate row. | 2 (pre-flight) | query.md (discovery/causal chain), change.md (the saga it feeds) |
| **README.md** | **The pitch**: spine in user terms + Install/Run + say-this/C3-does-this table + `c3x --help` pointer. ~60-80 lines. | frames 1+2 | SKILL.md, `--help`, CONTRIBUTING |
| **CLAUDE.md** | **Pure repo-dev contract**: Local-Source guardrail + Workflow + Build/CI/Release/Versioning. Tells NO act. | none | SKILL.md (one pointer for "what C3 does") |

**The clean rule:** SKILL.md *names the gate and routes*; the reference *teaches the steps*. Procedure lives in exactly one reference. The contract (freeze, rung-definition, status-set) lives once in SKILL.md and is *cited*, never re-stated.

---

## 2. REDUNDANCIES TO COLLAPSE — duplication → single home

| Duplicated content | Currently in | Single home |
|---|---|---|
| ADR-is-the-change-unit narrative (full saga) | SKILL.md ×4 (145, 77, 157, 244), onboard, ref, rule, audit, README | **change.md** (one line in SKILL.md) |
| Edit-a-fact procedure (`read --cite`→`change new`→patch→`apply`) | SKILL.md ×4, ref.md, rule.md ×4, audit | **change.md** |
| Frozen-fact rule + 3-moves table + refusal string | SKILL.md, change.md (9-26), ref, rule, canvas, README | **SKILL.md** (the contract); change.md cites it |
| "A fact is always complete to its rung" | SKILL.md:120, canvas.md:17, change.md:202, onboard | **SKILL.md HARD RULE** (canvas cites it) |
| Rung-climb scaffold→fill→gated-apply steps | canvas.md:28-39, SKILL.md:56, onboard | **change.md §Climbing a rung** |
| "Definitions are user-owned markdown" | SKILL.md ×3 (25, 122, 256), canvas, onboard, README | **canvas.md** (one-line Act-1 framing in SKILL.md) |
| ADR status set + `proposed`=`open` synonym + auto-done latch | SKILL.md, change.md ×2 (362, 416), ref, rule ×3, query, audit | **SKILL.md §change-doc** + **change.md** (one compact pass each, mutually cited) |
| Ref-vs-Rule Separation Test | ref.md, rule.md ×3 | **ref.md** (rule.md links it) |
| Cite-from-citer machinery (Governance row, edge column) | ref.md:76-92, rule.md (Step 8) | **change.md** (Phase 3.2); ref/rule each add one type-row note |
| Causal-chain expansion + direct-vs-transitive | query.md (Step 0a++ ×2 internally), sweep.md (×3) | **query.md** (one chain rule); sweep cites it |
| Discovery: search-first / never-Glob-`.c3/` / selector | SKILL.md Precondition, query, sweep, ref, rule | **SKILL.md Precondition** + **query.md** (others cite) |
| Command catalog / flag cheat-sheets | SKILL.md table, README (73-152), onboard (Capabilities Reveal), canvas (Commands table) | **SKILL.md table** + `c3x --help` (everyone else points) |
| Codemap WARN-at-check / `--strict-codemap` | SKILL.md, change.md:352, audit, onboard | **SKILL.md** `check` row (change.md keeps only the `.codemap.md` carrier-lands-atomically bit) |

---

## 3. CUTS — delete outright (off-story or stale)

**Stale (teaches the OLD model — highest priority to kill):**
- SKILL.md: `change` subcommand string missing `inspect`; apply-gate stack missing inspection; `rebase` never explained; destruction gate absent → **all corrected, not cut**.
- change.md: "Two gates, atomic" (line 138) and "THREE mechanical gates" (302-309) — **wrong gate count**; the entire passive "external arm" framing (322-359, "never gates done") contradicts the active inspection gate.
- README: change-as-ADR-four-step (line 64); `c3x delete ref-obsolete` example (134-140, teaches the *forbidden* path — delete is refused on a fact); "every write is validated" / enforce-on-write language (33,35,52,56,71,156); "drift" as an audit phase (68).
- audit.md: entire **"Drift Resolution" section** (190-199, "Create ADR / direct fix" = hand-edit-then-reconcile); "After-cites resolved at `c3 check --fix`" → resolution is `c3 change apply`; `uses:`/`via:`/`re-import`/`[provisioning]` Final-Checks table.
- onboard.md: **Final Checks `uses:`/`via:`/re-import/`[provisioning]` table** (311-314); the `c3 add container/ref/rule` bash recipes that fight the create-patch headers.
- ref.md/rule.md: scaffold-then-fill create flow (cannot work — a fact is frozen the moment it has a body); `c3 set origin` (refused — only `codemap` is exempt); `c3 list → relationships` citer-discovery (pre-graph); legacy `uses:`/Governance-fallback branch.
- CLAUDE.md: Repository Layout `cli/templates/` (**wrong path** — it's `cli/internal/templates/`); `Operations: query, audit, change, ref, sweep` (missing `rule`, `canvas`); references/ list missing `canvas.md`+`rule.md`; `bash scripts/build.sh` line that contradicts the CI-owns-build rule.

**Off-story (correct but wrong altitude — delete):**
- SKILL.md: §CoT Harness (97-104), §ASSUMPTION_MODE + AskUserQuestion denial (105-111), the MANDATORY File-Context block (163-192), §Graph Output playbook (198-211), embeddings self-heal / `c3 index` plumbing (139), change-doc lifecycle trivia (150-154).
- onboard.md: Laddering block (9-13), Size-the-canvas-first (58-60), **Complexity Guide table** (349-360), **Capabilities Reveal** (333-347), Recipe/Ref/Rule worked example catalogs, the three overlapping checklist systems → ONE linear gate list.
- canvas.md: standalone Commands table (44-55), embedded-seed-upgrade bullet (72-74 → onboard), **column-primitive enumeration** (75-77 — violates its own anti-goal; leave to `c3 schema`), two of three worked Outcomes (keep only `## Threat Model`), instances-are-CLI-only anti-goals (SKILL owns).
- change.md: **Provision Gate** (88-94 → one-liner or zero), re-index paragraph (388-389), diff-capture aside (64-68), the **duplicate ADR Lifecycle pass** (416-425 — collapse with 362-377).
- query.md: **Cross-Cutting Flow Expansion** + **Diverse Property Expansion** (48-98, domain-specific NATS/JetStream recipes), **Answer Depth Contract** 8-bar rubric (194-237), hardcoded example ids (`c3-211`, `sync.user.<email>`, `slackChannel`).
- audit.md: the **11-phase scaffold** (collapse to outcome-oriented), Phase 3 (Component Categorization), Phase 8 (abstraction-smell grep), Phase 10 (CLAUDE.md block example), Audit Scope table.
- ref.md/rule.md: Discover sub-mode matrices, Quality Gates, List/Usage output templates, Anti-Patterns tables (reduce to the single ref-vs-rule line), the long-form ADR-adoption blocks, the **Migrate saga** (rule.md 153-243 → 3-line pointer).
- sweep.md: Progress checklist (9-15), discovery preamble + property-expansion (19-49), Impact Topology (51-57), separate Constraint Chain phase (83-87).
- README: CLI cheat-sheet (73-152), hybrid-search internals (114-132), `.c3/` tree + daily-workflow (160-190), Distribution/Development (192-218), merkle/c3.db-as-selling-point bullets.
- CLAUDE.md: **entire "Skill Development Philosophy" section** (47-86) → relocate to a dedicated SKILL-DEV doc if wanted; "C3 skill design"/c4-derivation paragraph (30-32).

**Research/design docs now ABSORBED (their content becomes facts or one-liners):**
- The scattered "switch-gated double-V" explanation (memory: `switch-gated-double-v-model.md`) → **canonicalized in change.md** as the inspection-gate section; SKILL.md names it.
- "Membership-by-construction" design (v11.2.0 release note) → **stated once in canvas.md** (as shape) + **mechanics in change.md**; cut from every other surface.
- "Laddering / rung" research → **canvas.md** (the model) + **change.md** (the climb); the onboard Complexity Guide and SKILL HARD-RULE restatements are absorbed.
- README's hybrid-search write-up + SKILL's embeddings plumbing → **absorbed into `c3x --help`** (no longer prose anywhere).

---

## 4. PER-SURFACE DIRECTIVES (ordered by leverage — most-redundant first)

### 1. SKILL.md `(rewrite, ~260→~90 lines)` — DO FIRST, everything routes through it
**Says:** Open with the spine in 3-4 prose sentences (Act 1 build canvas → onboard facts → draft → **freeze**; Act 2 change-units, tool keeps it integral; Act 3 raise the canvas, migrate up). Then: Intent Classification table → Dispatch → **ONE** command table → **ONE** frozen-facts rule. Fix the stale facts: subcommands = `new|view|accept|apply|status|rebase|scaffold|inspect`; apply gate = **drift + canvas + retire + inspection**; name the **destruction gate** and `rebase` in one line each; state membership-by-construction as the load-bearing Act-2 guarantee (tool synthesizes parent rows from `parent:` edges).
**Drops:** CoT Harness, ASSUMPTION_MODE, File-Context playbook, Graph Output playbook, embeddings plumbing, all change-doc lifecycle trivia, every full procedure.
**Defers:** all step-by-step to references via one-line pointers to `{onboard,query,audit,change,ref,rule,canvas,sweep}.md`.

### 2. change.md `(rewrite, 425→~200)` — owns Act 2; the most-duplicated-INTO surface
**Says:** Change-unit = atomic saga (ADR=unit, patches in `.c3/changes/<id>/`, all-or-nothing). CURRENT end-to-end sequence: cite → `change new` → author patch(es) [+`.codemap.md`] → `view/status` → **`change inspect` → author `<seq>.inspect.md`** → `change accept` → `change apply` → `check`. New first-class **Inspection section**: `inspect` surfaces obligations + resolved territory + material hashes; the carrier pins hashes (`covers`) and attests each row (verdict `matches|updated`, evidence must name a file *in territory* — the anti-rubber-stamp floor). State the **FOUR** apply gates correctly with scoping: frontmatter-re-edge/retire **skip** inspection; zero-territory facts **defer** (docs-ahead-of-code). Keep membership-by-construction + Phase 3a + Climbing-a-rung's operational sequence (sections-ordered-LAST).
**Drops:** Provision Gate, re-index paragraph, diff aside, the duplicate status pass, the passive "external arm/never-gates-done" framing.
**Defers:** frozen-fact contract, rung *definition*, ADR section-lists, status *set* → cite SKILL.md once.

### 3. README.md `(rewrite, 222→~70)` — the pitch; teaches OLD model loudest
**Says:** Spine in user terms (3 acts). Make the v11.2.0 news the proof points of Act 2: membership-by-construction, the **inspection gate** (`change inspect`/`*.inspect.md`), the **destruction gate** (refuses orphan/dangle unless the unit heals it), **`change rebase`** conflict resolution. Keep: one-paragraph "what it is" (sealed `.c3/` markdown is truth, `c3.db` is rebuildable cache), Install/Run, say-this/C3-does-this table (fix the **change** row to the saga), single `c3x --help` pointer.
**Drops:** ADR-four-step, the `delete ref` example, enforce-on-every-write language, "drift" audit phase, CLI cheat-sheet, search internals, `.c3/` tree, Distribution/Development, the `build.sh` line.
**Defers:** commands → `--help`+SKILL.md; build → CONTRIBUTING/CLAUDE.md.

### 4. onboard.md `(rewrite, 360→~120)` — owns Act 1; tangled with create-patch contradiction
**Says:** Onboarding = ONE complete change-unit cycle. `c3 init` already made `c3-0` + genesis ADR `adr-00000000-c3-adoption`. You: (a) shape the canvas to fit NOW (lean rung-1 default; link canvas.md for the rung model), (b) discover topology (containers → Foundation 01-09 / Feature 10+ components → refs → rules → recipes), (c) author the **whole** architecture into the genesis ADR as **create-patches** (scope `whole`, no base, `type:`+`parent:`; you pick ids = the patch target), (d) **FLIP once**: `change apply adr-00000000-c3-adoption` materializes every fact atomically, canvas-validated — **now FROZEN**. State membership-by-construction once: leave every parent's membership table **header-only**, set `parent:` — the flip synthesizes rows. Pick **create-patches as the one primary path**; present `c3 add` only as the unguarded create exception. Close for v11.2.0: refresh After-cites → pass inspection → `accept` → `check --fix` latches `accepted→done`. One sentence on why `set <id> codemap` survives the freeze (verified external binding, not a frozen fact). 4-line CLAUDE.md injection with the **current op set** (query, audit, change, ref, rule, canvas, sweep) + `c3local` convention.
**Drops:** Laddering block, Size-the-canvas-first, Complexity Guide, Capabilities Reveal, Final Checks + the `uses:`/`via:`/`[provisioning]` table, the `c3 add container/ref/rule` recipes, three checklists → ONE gate list.
**Defers:** commands→SKILL.md, the unit→change.md, rungs→canvas.md.

### 5. rule.md `(rewrite, 324→~70)` — ~90% clone of ref.md; built on forbidden scaffold-then-fill
**Says:** A rule is a frozen FACT — enforceable standard, shape Goal/Rule/Golden Example (defer full contract to `c3 schema rule`). Ref-vs-rule boundary = one line → ref.md's test. **Two ops:** (1) CREATE — author the full body into a file, `c3 add rule <slug> --file body.md`; **no scaffold-then-fill**, no `write`/`set` afterward (`set <id> codemap` is the only exemption — so `origin:`/`scope:` go in the body file, never `set`). (2) CHANGE — a change-unit (cite→`change new`→patch→`apply`); defer mechanics to change.md. One citation note: citer carries a `Governance` row `Type: rule`. Migrate → 3-line pointer (classify via ref.md test, author whole, rewire+retire in one change-unit; the destruction gate prevents orphans — drop the manual "no orphan refs" check). Lead discovery with `c3 search` / `c3 graph --direction reverse`.
**Drops:** multi-mode taxonomy, scaffold flow, `set origin`, Migrate saga, long-form ADR-adoption blocks, Discover/Quality-Gate/Anti-Patterns, `c3 list` dumps.

### 6. ref.md `(rewrite, 195→~50)` — clones rule.md + re-teaches Act 2
**Says:** A ref is a frozen-fact type whose identity is **rationale** (`## Why` is load-bearing). KEEP the **Ref-vs-Rule Separation Test table** (the irreplaceable content) + the "can't name a concrete file → ref" identity. Everything else = one-liner: authoring → `c3 schema ref` (leads with REJECT IF); citing/adoption/auto-done → change.md §Phase 3.2 + §Status; editing → "refused; ride as a patch via `change apply`"; List/Usage → `c3 graph ref-<slug> --direction reverse [--format mermaid]`.
**Drops:** Discover matrix, Quality Gate, output templates, Anti-Patterns, the `relationships`-field reads, the duplicated cite-machinery and latch mechanics.

### 7. audit.md `(rewrite, 209→~80)` — keyed off the abandoned doc-rot "drift" model
**Says:** "Is the sealed truth intact and consistent?" Three layers: (1) **SEAL** — `c3 check` validates canonical seal + cache; on seal drift (branch switch / selective merge / conflict) → `c3 repair` rebuilds + reseals (currently absent, now the center of gravity). (2) **STRUCTURAL** — RUN `c3 check`, read its output (broken links, orphans, dup ids, missing parents, empty required sections per canvas, code-refs-on-disk, cite consistency, coverage signal `mapped/(total-excluded)`) — don't hand-walk tables↔dirs. (3) **SEMANTIC** — orphan refs/rules (cited by ≥1), ref `## How`/rule `## Golden Example` spot-checks → defer specifics to ref.md/rule.md/canvas.md. The ONLY fix loop = author a change-unit + `c3 change apply` (point to change.md). Phase 6 → one line: stuck unit = `status: accepted` long-unapplied; surface it, don't hand-close; canonical set `[open, accepted, done, superseded]`; terminal docs check-exempt. Acknowledge destruction gate + membership-by-construction so nothing teaches hand-authored parent tables. End in PASS/WARN/FAIL.
**Drops:** Drift Resolution section, 11-phase scaffold, Phases 3/8/10, Audit Scope table, `After-cites`/`check --fix` resolution language.

### 8. canvas.md `(rewrite, 88→~50)` — the model surface; missing membership-by-construction
**Says:** **(Act 1) what a canvas IS** — your architecture's own vocabulary made a contract (sections + typed columns each entity carries HERE). Lean rung-1 seeds materialize to user-owned `.c3/canvases/<type>.md`; `c3 check` validates against THAT file, so editing the definition changes what's enforced. State the **membership consequence once** (it's part of the shape): a parent's membership column is **synthesized from children's `parent:` edges** — shape the tool fills, never an author-edited row → point to change.md for mechanics. **(Act 3) raising the canvas** — a canvas is a RUNG (a complete contract for one level); integrity = a fact is always complete to its rung; raise the bar → every fact below migrates up completely (that's WHY the climb exists). State the principle, hand off to change.md §Climbing a rung for the scaffold→fill→gated-apply steps. Keep the `## Threat Model` worked example. Use bare `c3` uniformly.
**Drops:** Commands table, embedded-seed-upgrade bullet, column-primitive enumeration, instances-are-CLI-only anti-goals. **Keeps** only its two genuine anti-goals (don't enumerate a fixed type/section set in prose; don't treat `adr` as special — it's the `adr` canvas).

### 9. query.md `(rewrite, 254→~90)` — owns Act-1 read beat but never states facts are frozen
**Says:** Name its place: the topology you read is **FROZEN shared truth** — onboarded once, changed only via change-units — so a `read`/`lookup`/`search`/`graph` result is **canonical**; answer from it with confidence. Keep: discovery-tool selector (search=concepts, lookup=files/globs, read=known ids), read-only fast path (search before list/check), never-Glob-`.c3/`, recipe check, layered Context→Container→Component nav. Keep exactly ONE evidence rule (claim bound to a read you ran, cite the id) + ONE causal-chain rule (owner→mutation→mechanism→dependent→emergent→failure-boundary, each arrow = which contract carries it). Merge the internal duplicate into that one rule; make it abstract.
**Drops:** Cross-Cutting/Diverse Property Expansion, Answer Depth Contract, hardcoded example ids, the ADR Handling block.
**Defers:** ADR status/visibility → SKILL.md+change.md; impact+verification → sweep.md.

### 10. sweep.md `(rewrite, 130→~70)` — owns pre-flight but ignores the destruction gate (its whole point)
**Says:** Sweep = **pre-flight for a change-unit**. One job: given a proposed change to a frozen fact, predict the blast radius AND **whether the tool will let it land**. Make `c3 graph <id> --direction reverse` the spine, tied to the destruction gate: if it's a removal/retire, name which dependents would be **ORPHANED** (live children) or **DANGLED** (live citers) and state plainly that `apply`/`retire` will **REFUSE** it unless the same unit reparents/retires the child and drops the citation — so the deliverable is the list of consequences the unit must also handle. Never list "update parent membership" (synthesized). Bridge to the saga: end with `graph --unit <adr-id>` to preview staged patches' post-change graph before `apply`; frame "File Changes Required" as patches in `.c3/changes/<unit-id>/` landing all-or-nothing. Keep the **Verification table** + add a **destruction-gate-prediction row** (does retire/apply succeed or refuse?).
**Drops:** Progress checklist, discovery preamble, property-expansion, Impact Topology, separate Constraint Chain phase, direct-vs-transitive paragraph.
**Defers:** discovery/causal-chain → query.md (one line).

### 11. CLAUDE.md `(rewrite, 167→~110)` — repo-dev contract; stale paths, teaches no act (correct) but lists must be fixed
**Says:** Pure dev instructions. KEEP+tighten: (1) Local C3 Source Rule (load-bearing guardrail), (2) dev Workflow (brainstorm→plan/subagents→`/noslop`+`c3local check`→`/release`), (3) Plugin Structure/Build/CI/Release/Versioning. **FIX:** Repository Layout `cli/templates/` → `cli/internal/templates/` (and show core libs under `cli/internal/`); references/ list → all **8** files (add `canvas.md`, `rule.md`); reconcile the bin/ "gitignored" comment with the ~40 committed binaries actually present. Replace the inline op list + c4-framing with ONE pointer: "Product behavior, the op set, and the change-unit/freeze/canvas model are owned by `skills/c3/SKILL.md` — read it there, don't duplicate."
**Drops:** entire Skill Development Philosophy section (47-86), the c4-derivation paragraph, the stale `Operations:` line, the `bash scripts/build.sh` cross-compile line.

---

## 5. THE C3-OWN-CANVAS MOVE — dogfood Act 1 by modeling C3's own design narrative

**The principle being dogfooded:** Act 1 says *build the model that fits your project*. C3's repo has a need its default architecture canvases don't serve — **the three-act design narrative itself** (the "why" of the product's shape, the rationale for each act, the gate philosophy). Today that narrative is smeared across SKILL.md prose, README, onboard, and `memory/*.md`. It is unowned, duplicated, and drifts. The fix is the move C3 tells everyone else to make: **define a custom canvas type and store the narrative as first-class frozen facts.**

This is already proven viable in-tree: the builtin set ships `prd`, `user-story`, `pm-requirement`, `atomic-design-change` (cli/internal/schema/builtin/canvases/) — C3 already supports non-architecture canvas types. So the move is sanctioned by construction.

**The new canvas type — `design-act` (a narrative fact-type):**

Define `.c3/canvases/design-act.md` with a custom schema. One instance per act (three facts: `act-1-freeze`, `act-2-change-units`, `act-3-raise-canvas`), plus optionally a parent `design-spine` system-level fact. Sections (typed):

| Section | type | required | purpose |
|---|---|---|---|
| **Thesis** | text | yes | The act in one sentence (the spine beat) |
| **The Move** | text | yes | What the user does in this act |
| **Tool Guarantee** | table | yes | What the TOOL keeps integral, and how (cols: Guarantee · Mechanism · Source) — e.g. *membership · synthesized from `parent:` edges · add.go/check_enhanced.go* |
| **Gates** | table | act-2/3 | The gates this act enforces (cols: Gate · When · Refusal) |
| **Why This Shape** | text | yes | The rationale — *why this and not the alternative* (absorbs the design-rationale prose) |
| **Surfaces** | table | yes | Which doc surface teaches this act (cols: Surface · Owns) — makes the canonical-structure map a fact |

Because these are facts, `c3 check` validates that every act carries its Thesis/Move/Guarantee/Why — the narrative can no longer go thin or drift, and a change to the model (e.g. a new gate) **must** ride a change-unit, exactly like architecture facts.

**What it absorbs (these stop being free-floating prose):**
- The three-act spine prose now scattered in **SKILL.md lede, README lede, onboard intro** → becomes the `Thesis`+`The Move` of each `design-act` fact. The surfaces then *cite* it (`c3 read act-2-change-units`) instead of re-narrating.
- The v11.2.0 "tool keeps it integral" proof points (membership, inspection gate, destruction gate, rebase) → the `Tool Guarantee` + `Gates` tables of `act-2-change-units` — **one canonical home**, source-cited, instead of being re-listed in README + SKILL + change.md.
- The "why switch-gated, not enforce-on-write" and "why freeze" rationale (currently in `memory/switch-gated-double-v-model.md`, `eliminate-risk-by-construction.md`) → the `Why This Shape` sections. The memory notes become the *seed*; the canvas fact becomes the *owned, checked* version.
- The canonical-structure map from §1 of this plan → the `Surfaces` tables, so "who owns which act" is itself a frozen fact the audit can verify.

**Net dogfood payoff:** the docs that *teach* C3 stop *containing* the design narrative — they read it from C3. The product proves its own thesis: when the work (explaining C3) outgrew the architecture model, C3 didn't bend the facts — it **raised a new canvas** (`design-act`) and migrated the narrative into it as checked, change-unit-governed truth. SKILL.md's lede can then legitimately say "the spine is itself a C3 fact — `c3 read design-spine`," which is the strongest possible Act-1 demonstration.

---

**Execution order (highest leverage → lowest):** SKILL.md (1) → change.md (2) → README.md (3) → onboard.md (4) → rule.md (5) → ref.md (6) → audit.md (7) → canvas.md (8) → query.md (9) → sweep.md (10) → CLAUDE.md (11). The C3-own-canvas move (`design-act` definition + 3 facts authored through a real change-unit) lands **after SKILL.md+change.md** are correct, since it dogfoods exactly the flow those two surfaces must already describe accurately.

Key source anchors for the implementer (all absolute):
- `/home/lagz0ne/dev/c3-design/cli/cmd/help.go:387,403-407` — subcommands + apply gate + inspect copy
- `/home/lagz0ne/dev/c3-design/cli/cmd/change.go:48-94` — the four-gate preflight order
- `/home/lagz0ne/dev/c3-design/cli/cmd/destruction.go:11-50` — orphan/dangle refusal + heal-in-same-unit escape
- `/home/lagz0ne/dev/c3-design/cli/cmd/inspect.go:128-146` — anti-rubber-stamp territory floor, defer-on-zero-territory
- `/home/lagz0ne/dev/c3-design/cli/cmd/freeze.go:74-104` — only `codemap` carve-out + canonical refusal string
- `/home/lagz0ne/dev/c3-design/cli/cmd/add.go:127-134` + `check_enhanced.go:252` — membership-by-construction (both paths)
- `/home/lagz0ne/dev/c3-design/cli/cmd/init.go:54` — `adr-00000000-c3-adoption`, `c3-0`
- `/home/lagz0ne/dev/c3-design/cli/internal/schema/builtin/canvases/` — proof custom non-arch canvas types ship (`prd`, `user-story`, `system`); template for authoring `design-act.md`
