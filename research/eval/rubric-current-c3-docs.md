# C3 Eval Rubric — Grounded in Current `.c3/` Docs

Ground truth derived from the live `.c3/` docs of `/home/lagz0ne/dev/c3-design`,
read via the **local** CLI only:

```bash
alias c3local='C3X_MODE=agent bash skills/c3/bin/c3x.sh'
```

Survey date: 2026-06-01. `c3local check` reports `total: 73`, **`issues:` empty (clean)**.

---

## Survey snapshot (the architecture the rubric is grounded on)

**System:** `c3-0` — *c3-design*. Goal: build & distribute the c3 Claude Code plugin (CLI-driven architecture docs). Two containers.

**Containers:**
- `c3-1` — *Go CLI* (boundary `process`). Owns all `.c3/` read/write, numbering, codemap matching, structural validation, binary compile.
- `c3-2` — *Claude Skill* (boundary `app`). Intent routing + workflow orchestration over `.c3/` docs via c3x.

**Components of `c3-1` (20):** c3-101 frontmatter, c3-102 walker, c3-103 templates, c3-104 wiring, c3-105 codemap-lib, c3-106 content-lib, c3-107 store-lib, c3-108 runtime-support, c3-109 npm-cli-wrapper, c3-110 init-cmd, c3-111 add-cmd, c3-112 list-cmd, c3-113 check-cmd, c3-114 lookup-cmd, c3-115 codemap-cmd, c3-116 coverage-cmd, c3-117 docs-state-cmds, c3-118 analysis-cmds, c3-119 sync-lifecycle-cmds, c3-120 history-marketplace-cmds.

**Components of `c3-2` (10):** c3-201 skill-router, c3-210 operation-workflow-index, c3-211 onboard-operation, c3-212 query-operation, c3-213 audit-operation, c3-214 change-operation, c3-215 migrate-operation, c3-216 ref-operation, c3-217 rule-operation, c3-218 sweep-operation.

**Refs (3):** ref-frontmatter-docs, ref-cross-compiled-binary, ref-embedded-templates. **Rules:** none in repo. **Recipe (1):** recipe-validation-system (sources: c3-101, c3-102, c3-103, c3-104, c3-113).

**`uses` edges (from frontmatter / graph):** c3-102→c3-101; c3-105→c3-102; c3-114→{c3-101,c3-102,c3-105}; c3-115→{c3-101,c3-102,c3-105}; c3-116→{c3-102,c3-105,c3-113}; c3-111→{c3-101,c3-102,c3-103,c3-104}; c3-113→{c3-101,c3-102,c3-104}; c3-212→{c3-201,c3-210}.

**Refs are wired by codemap only** (see `.c3/code-map.yaml`), NOT cited in component Governance tables. Every component's Governance table cites its **parent container** as a `policy` (e.g. c3-101/c3-103 Governance → `c3-1` / `policy`). This is a grounding caveat — see notes on items GOV-1/GOV-2.

---

## How to run / score

- Agent must use `c3local <cmd>` only. Any answer derived from bare `c3x` or a global skill is an automatic **FAIL** for that item.
- Each item below states the **exact** ground-truth string and a **mechanical** pass rule. Scoring is deterministic: normalize whitespace/case where noted, then apply the match rule. No partial credit unless stated.

---

## Category A — Navigation / file → entity (`c3local lookup`)

| ID | Task | Ground truth | Source | Scoring (mechanical) |
|----|------|--------------|--------|----------------------|
| NAV-1 | Which component owns `cli/internal/codemap/codemap.go`? | `c3-105` (codemap-lib) | `lookup cli/internal/codemap/codemap.go` → matches c3-105 | PASS iff answer contains id `c3-105` (exact token). |
| NAV-2 | Which component owns `cli/cmd/lookup.go`? | `c3-114` (lookup-cmd) | `lookup cli/cmd/lookup.go` → c3-114 | PASS iff id `c3-114`. |
| NAV-3 | Which component owns `cli/internal/walker/walker.go`? | `c3-102` (walker) | `lookup` → c3-102 | PASS iff id `c3-102`. |
| NAV-4 | Which component owns `cli/cmd/add.go`? | `c3-111` (add-cmd) | `lookup` → c3-111 | PASS iff id `c3-111`. |
| NAV-5 | Which component owns `cli/cmd/check_enhanced.go`? | `c3-113` (check-cmd) | `lookup` → c3-113 | PASS iff id `c3-113`. |
| NAV-6 | Which component owns `cli/internal/store/store.go`? | `c3-107` (store-lib) | `lookup` → c3-107 | PASS iff id `c3-107`. |
| NAV-7 | Which component owns `skills/c3/references/query.md`? | `c3-212` (query-operation) | `lookup` → c3-212 | PASS iff id `c3-212`. |
| NAV-8 | What entity governs `scripts/build.sh`? (note: a **ref**, not a component) | `ref-cross-compiled-binary` | `lookup scripts/build.sh` → matches ref-cross-compiled-binary, no component | PASS iff answer names `ref-cross-compiled-binary` AND does not assert a component owner. |
| NAV-9 | Does `cli/internal/wiring/wiring.go` map to any component? | **No match** — the path is uncharted; c3-104 (wiring) is codemapped to `cli/cmd/wire.go`, not `cli/internal/wiring/**`. | `lookup cli/internal/wiring/wiring.go` → empty matches + coverage-gap hint | PASS iff answer states no component matches / it is uncovered. FAIL if it claims c3-104 owns it. |

---

## Category B — Ownership / topology

| ID | Task | Ground truth | Source | Scoring |
|----|------|--------------|--------|---------|
| OWN-1 | What container is `c3-114` (lookup-cmd) in? | `c3-1` (Go CLI) | `read c3-114` → `parent: c3-1` | PASS iff id `c3-1`. |
| OWN-2 | What container is `c3-201` (skill-router) in? | `c3-2` (Claude Skill) | `read c3-201` → `parent: c3-2` | PASS iff id `c3-2`. |
| OWN-3 | How many components does container `c3-1` own, and name 3? | **20** components (c3-101…c3-120). | `read c3-1` Components table / `list` | PASS iff count = 20 AND ≥3 named ids are members of the c3-1 set. |
| OWN-4 | What is the responsibility/goal of `c3-113`? | "Validate structural integrity of `.c3/` docs, ref and rule compliance — required fields, numbering, wiring, scope cross-checks, origin validation." | `read c3-113` goal / `list` | PASS iff answer conveys validation/structural-integrity ownership of `.c3/` docs (must mention validation AND `.c3/` docs). |
| OWN-5 | Which component owns embedded markdown templates? | `c3-103` (templates) | `list` / `read c3-103` | PASS iff id `c3-103`. |
| OWN-6 | What is the boundary of container `c3-1` vs `c3-2`? | c3-1 = `process`; c3-2 = `app` (skill = "Claude Code session"). | `read c3-1`/`read c3-2` `boundary:` field | PASS iff c3-1→process AND c3-2→app (both correct). |

---

## Category C — Relationships / governance (`uses` edges, refs, recipe)

| ID | Task | Ground truth | Source | Scoring |
|----|------|--------------|--------|---------|
| REL-1 | Which components does `c3-114` (lookup-cmd) use? | `c3-101`, `c3-102`, `c3-105` | `read c3-114` → `uses: c3-101,c3-102,c3-105` | PASS iff answer set == {c3-101,c3-102,c3-105} exactly. |
| REL-2 | What does `c3-105` (codemap-lib) depend on (`uses`)? | `c3-102` | `read c3-105` → `uses: c3-102` | PASS iff set == {c3-102}. |
| REL-3 | Which components does `c3-111` (add-cmd) use? | `c3-101,c3-102,c3-103,c3-104` | `graph c3-111` / `read c3-111` `uses` | PASS iff set == {c3-101,c3-102,c3-103,c3-104}. |
| REL-4 | What does ref `ref-frontmatter-docs` standardize/scope? | Every `.c3/` doc uses YAML frontmatter (machine-readable metadata) + Markdown body; structural metadata must not live in the body. Scoped to `cli/internal/frontmatter/**` and `cli/internal/markdown/**`. | `read ref-frontmatter-docs` (Goal/Choice/Not This); `code-map.yaml` | PASS iff answer states the frontmatter+markdown-body separation pattern. |
| REL-5 | What does ref `ref-cross-compiled-binary` require, and which 4 targets? | Pre-built binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`; no Go toolchain needed; selected via `c3x.sh`. | `read ref-cross-compiled-binary` | PASS iff all 4 target tuples present (os/arch). |
| REL-6 | Which entities does `recipe-validation-system` point to as sources? | `c3-101, c3-102, c3-103, c3-104, c3-113` | `list` (recipe sources) / `read recipe-validation-system` + ADR | PASS iff set == {c3-101,c3-102,c3-103,c3-104,c3-113}. |
| GOV-1 | What does the Governance table of component `c3-103` (templates) cite, and at what precedence? | Cites parent `c3-1` as type `policy`; precedence "Explicit cited governance beats uncited local prose." | `read c3-103` Governance section (`.c3/c3-1-go-cli/c3-103-templates.md` L52-54) | PASS iff answer names `c3-1` as the governing reference of type policy. **Caveat:** governance is parent-policy, not a ref. |
| GOV-2 | Which refs are cited in component **Governance** tables? | **None** — refs are wired only via `code-map.yaml`; no component Governance row cites a `ref-*`. | grep of `.c3/` Governance tables; `code-map.yaml` ref entries | PASS iff answer states no ref appears in any component Governance table (refs are codemap-wired only). **AMBIGUOUS for naive agents** — flag. |

---

## Category D — Graph / change-impact (`c3local graph`)

| ID | Task | Ground truth | Source | Scoring |
|----|------|--------------|--------|---------|
| IMP-1 | If `c3-105` (codemap-lib) changes, which components directly depend on it (consumers)? | `c3-114`, `c3-115`, `c3-116` (all `uses` c3-105). | `graph c3-105` nodes: c3-114/c3-115/c3-116 have c3-105 in their uses list | PASS iff answer set == {c3-114,c3-115,c3-116}. |
| IMP-2 | What does `c3-105` itself depend on (so what could break it)? | `c3-102` (walker); transitively c3-101. | `graph c3-105` → c3-105 uses c3-102 | PASS iff answer includes c3-102 (transitive c3-101 acceptable, not required). |
| IMP-3 | Which ADRs are linked to `c3-114`'s subgraph? | `adr-20260320-add-rule-entity-type` and `adr-20260415-cascade-help-hints`. | `graph c3-114` nodes include both ADRs | PASS iff both ADR ids named. |
| IMP-4 | If `ref-frontmatter-docs` changes, what code paths are in scope? | `cli/internal/frontmatter/**`, `cli/internal/markdown/**`. | `graph ref-frontmatter-docs` sources / `code-map.yaml` | PASS iff both globbed paths named. |
| IMP-5 | What is the parent chain for `c3-114`? | `c3-114` → `c3-1` → `c3-0`. | `graph c3-114` / parent fields | PASS iff chain c3-114→c3-1→c3-0 stated in order. |

---

## Category E — Validation (`c3local check` + schema conformance)

| ID | Task | Ground truth | Source | Scoring |
|----|------|--------------|--------|---------|
| VAL-1 | Does the current repo pass `c3x check`? How many entities, how many issues? | PASS — `total: 73`, `issues:` empty. | `check` | PASS iff answer says check passes with 0 issues (count 73 is a bonus, not required). |
| VAL-2 | Does `c3-105`'s doc satisfy the `component` definition (all required sections present)? | Yes — c3-105 has all required component sections (Goal, Parent Fit, Purpose, Foundational Flow, Business Flow, Governance, Contract, Change Safety, Derived Materials); check is clean. | `schema component` (required sections) vs `read c3-105` | PASS iff answer = yes/satisfies AND references the required-section set. |
| VAL-3 (mutation, do NOT apply) | If the `parent:` field of `c3-114` were changed from `c3-1` to `c3-9` (a non-existent container) in its description/frontmatter, what does `check` report? | `check` FAILS: c3-114 references a parent that does not exist (dangling parent / unresolved reference). Repo is currently clean, so this surfaces a new issue. | inferred from `check` semantics (validates numbering, wiring, parent resolution) + current clean baseline | PASS iff answer predicts `check` FAILS due to the unknown/dangling parent `c3-9`. FAIL if it predicts pass. |
| VAL-4 (mutation, do NOT apply) | If the `## Governance` section were deleted from `c3-105`'s doc, what does `check` report? | `check` FAILS: required component section missing (Governance is a required section per `schema component`). | `schema component` marks Governance required (`true`) | PASS iff answer predicts FAIL citing missing required section. |
| VAL-5 (mutation, do NOT apply) | If a `uses: c3-999` edge (nonexistent target) were added to `c3-114`, what does `check` report? | `check` FAILS: wiring/citation points to an entity that does not exist (orphan/dangling citation). | wiring component c3-104 validates citations; check covers wiring | PASS iff answer predicts FAIL due to dangling `uses` target. |

---

## Category F — Definition awareness (canvas / `c3local schema`)

| ID | Task | Ground truth | Source | Scoring |
|----|------|--------------|--------|---------|
| DEF-1 | What sections does the `component` canvas definition require? | Required (`true`): **Goal, Parent Fit, Purpose, Foundational Flow, Business Flow, Governance, Contract, Change Safety, Derived Materials** (all 9 are required). | `schema component` | PASS iff answer lists all 9 required sections (set match; order-independent). |
| DEF-2 | The `Governance` table in the `component` definition has which columns? | `Reference`, `Type`, `Governs`, `Precedence`, `Notes` — where `Type` is an enum: `ref, rule, adr, spec, policy, N.A - <reason>`. | `schema component` (Governance table spec) | PASS iff all 5 column names present (Type enum values are bonus). |
| DEF-3 | What columns does the `container` definition's `Components` table require, and which sections are required? | Components table columns: `ID, Name, Category, Status, Goal Contribution`. Required sections: `Goal, Components, Responsibilities` (Complexity Assessment optional). | `schema container` | PASS iff Components columns set correct AND the 3 required sections named. |
| DEF-4 | What sections does the `ref` definition require, and what is its rejection rule about `Why`? | Required: `Goal, Choice, Why` (`How` optional). Reject if `Why` restates `Choice` (then it's a rule, not a ref). | `schema ref` (sections + reject_if) | PASS iff required set {Goal,Choice,Why} named AND the Why-restates-Choice rejection is described. |
| DEF-5 | What columns does the component `Contract` table require? | `Surface`, `Direction` (enum IN/OUT/IN/OUT/N.A), `Contract`, `Boundary`, `Evidence`. | `schema component` (Contract table spec) | PASS iff all 5 column names present. |

---

## Ambiguity / grounding-gap notes

- **GOV-2 is the main ambiguity**: refs (`ref-*`) are linked to code *only* through `code-map.yaml`; they are **not** cited inside any component's `## Governance` table. A naive "which refs govern component X?" question has no per-component answer in the Governance tables — the truthful answer is "none cited there; refs map to file globs." Treat GOV-2 as a discriminating item, not a gotcha to penalize good agents on; flagged AMBIGUOUS.
- **GOV-1**: every component's Governance row currently cites the parent container as `policy` with boilerplate ("Migrated from legacy component form; refine during next component touch."). So Governance answers are uniform across components — low discriminating power. Item is grounded but shallow.
- **No rules exist** in this repo, so the prompt's "which refs/rules govern" category is covered only by refs + the parent-policy Governance pattern. There is no rule-scoping item to ground.
- **`impact` is not a subcommand** in this build (agent mode `--help` lists no `impact`). Change-impact is grounded on `graph` instead (Category D). Any rubric harness should not expect a `c3x impact` command.
- **Lookup coverage gaps are intentional ground truth** (NAV-9): `cli/internal/wiring/**` is not in `code-map.yaml`; c3-104 is mapped to `cli/cmd/wire.go`. This is a real coverage gap and is itself a valid eval signal.
- Mutation items (VAL-3/4/5) are **predicted** failures derived from `check` semantics + the clean baseline; the repo was NOT mutated. A harness wanting hard ground truth could apply each mutation in a throwaway copy and capture real `check` output.
