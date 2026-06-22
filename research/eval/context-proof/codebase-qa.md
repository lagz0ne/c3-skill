# C3-design codebase Q&A (context-proof test set)

A traceable Question -> Answer set distilled from real, tool-verified observations of the
c3-design codebase. Every entry was verified against the live tool
(`C3X_MODE=agent bash skills/c3/bin/c3x.sh <cmd>`) from the repo root.

Two intended uses:
- **Forward** — ask a question, get the gold answer.
- **Reverse** — take an answer/doc fact and trace back to the exact code.

It also serves as the test set for a later "agent-with-docs vs agent-without-docs" comparison,
so every entry is REAL and VERIFIED, not invented. Gaps and the known trust bug stay in, flagged.

**Verification baseline (this session):** full `c3x eval` = 23 verdicts, 22 holds, 0 drift,
1 needs-judgement (`rule-dispatcher-error-hint`). All version surfaces = 11.2.0.

Schema per entry: `question`, `answer`, `doc_trace` (fact ids), `code_trace` (code paths),
`via` (exact c3x commands), `coverage` (clean | partial | gap).

---

## Navigation / architecture

### qa-01 — Where is the change-apply saga and what are its gates?
- **Answer:** The saga lives in the changeset package; the command-side gate orchestration is in change-cmds. `apply` runs four gates in order: **drift/conflict, canvas, morph, retire**.
- **doc_trace:** c3-104, c3-112
- **code_trace:** `cli/internal/changeset/**`, `cli/internal/changeset/apply.go`, `cli/cmd/change.go`
- **via:** `c3x read c3-104`; `c3x lookup cli/internal/changeset/apply.go`; `references/change.md` (gate stack)
- **coverage:** clean

### qa-02 — How does the store persist entities?
- **Answer:** One SQLite-backed store; a single package owns all schema and SQL behind a typed API for entities, relationships, nodes, versions, code-map, and semantic vectors. Frontmatter fields become columns.
- **doc_trace:** c3-102
- **code_trace:** `cli/internal/store/**`, `cli/internal/store/store.go`, `cli/internal/store/entities.go`
- **via:** `c3x read c3-102`; `c3x lookup cli/internal/store/**`
- **coverage:** clean

### qa-03 — How does search rank results?
- **Answer:** Search fuses content/entity FTS (keyword), graph expansion (edges), and a local ONNX semantic index. The read side dispatches `search`; the fusion and indexes live in the store.
- **doc_trace:** c3-110, c3-102
- **code_trace:** `cli/cmd/search.go`, `cli/cmd/semantic_index.go`, `cli/internal/store/search.go`, `cli/internal/store/semantic.go`
- **via:** `c3x read c3-110`; `c3x lookup cli/internal/store/search.go`; `c3x search <term>`
- **coverage:** clean

### qa-04 — What is the canvas model and where does it live?
- **Answer:** A canvas is the typed shape (sections, column types, edges) a fact must match. Built-in canvases are embedded in the schema package; project canvas definitions live in `.c3/canvases/`.
- **doc_trace:** c3-103
- **code_trace:** `cli/internal/schema/**`, `cli/internal/schema/builtin.go`, `cli/internal/schema/builtin/`, `.c3/canvases/`
- **via:** `c3x read c3-103`; `c3x canvas list`
- **coverage:** clean

### qa-05 — What governs the CLI's output formatting?
- **Answer:** All structured output routes through shared helpers in `cli/cmd/output.go` (never raw `fmt.Print` in `cmd/`), serializing TOON by default and JSON only via explicit `--json`.
- **doc_trace:** rule-output-via-helpers
- **code_trace:** `cli/cmd/output.go`
- **via:** `c3x read rule-output-via-helpers`; `c3x lookup cli/cmd/output.go`
- **coverage:** clean

### qa-06 — How does the CLI bootstrap and dispatch commands?
- **Answer:** `main.go` and runtime-support resolve the `.c3/` directory, dispatch the command, serialize output, and serialize concurrent mutations behind a coordinator.
- **doc_trace:** c3-107
- **code_trace:** `cli/main.go`, `cli/internal/config/**`, `cli/internal/coord/**`, `cli/internal/toon/**`
- **via:** `c3x read c3-107`; `c3x eval c3-107`
- **coverage:** clean

### qa-07 — Where is the conformance/eval engine?
- **Answer:** The eval engine is the five-op pipeline interpreter (gather, filter, transform, eval, loop) that reduces each fact's eval spec to a stamped verdict. Owned by component c3-108 (added this session).
- **doc_trace:** c3-108, ref-eval-determinism
- **code_trace:** `cli/internal/eval/**`, `cli/internal/eval/eval.go`
- **via:** `c3x read c3-108`; `c3x lookup cli/internal/eval/**`; `c3x eval c3-108`
- **coverage:** clean

### qa-08 — Where does evolve-unit (morphing the model) live, and how does it differ from a change-unit?
- **Answer:** A change-unit edits facts; an evolve-unit reshapes a fact-TYPE's canvas (a `canvas` scope patch) and migrates instances in the same atomic unit. The CLI surface is `cli/cmd/morph.go` (owned by change-cmds c3-112); the saga support is in the changeset package.
- **doc_trace:** c3-112
- **code_trace:** `cli/cmd/morph.go`, `cli/internal/changeset/patch.go`, `cli/internal/changeset/**`
- **via:** `c3x lookup cli/cmd/morph.go`; `c3x read c3-112`; `references/change.md`
- **coverage:** partial (doc coverage of evolve-unit is thinner than change-unit)

---

## Change-impact

### qa-09 — What governs cli/internal/changeset/apply.go?
- **Answer:** Component c3-104 (changeset) owns it; it cites rule-wrap-error-cause.
- **doc_trace:** c3-104, rule-wrap-error-cause
- **code_trace:** `cli/internal/changeset/apply.go`
- **via:** `c3x lookup cli/internal/changeset/apply.go`
- **coverage:** clean

### qa-10 — What governs cli/internal/store/?
- **Answer:** Component c3-102 (store) owns the entire `cli/internal/store/` tree.
- **doc_trace:** c3-102
- **code_trace:** `cli/internal/store/**`
- **via:** `c3x lookup cli/internal/store/**`
- **coverage:** clean

### qa-11 — What code in cli/ is unowned by any fact?
- **Answer:** `cli/cmd/{help,helpers,options}.go` (and their `_test.go`), `cli/tools/search-eval/**`, and `cli/tools/semantic-assets/**` resolve to no owning component. `cli/internal/eval/**` is now owned (c3-108), closing a prior gap.
- **doc_trace:** (none)
- **code_trace:** `cli/cmd/help.go`, `cli/cmd/helpers.go`, `cli/cmd/options.go`, `cli/tools/search-eval/**`, `cli/tools/semantic-assets/**`
- **via:** `c3x lookup cli/cmd/**`; `c3x lookup cli/tools/**` (file_map empty owner)
- **coverage:** gap (real coverage holes)

### qa-12 — Which components cite rule-wrap-error-cause?
- **Answer:** c3-104, c3-108, c3-111, c3-112, c3-113 cite it (c3-108 added this session). Caveat: the rule's own eval `code:` binding in `.c3/eval/rule-wrap-error-cause.yaml` points at `cli/cmd/eval.go`, a file no component owns.
- **doc_trace:** rule-wrap-error-cause
- **code_trace:** `.c3/eval/rule-wrap-error-cause.yaml`, `cli/cmd/eval.go`
- **via:** `c3x graph rule-wrap-error-cause`
- **coverage:** partial (stale/unowned eval bind)

### qa-13 — Does every component's declared code surface still resolve?
- **Answer:** Yes. A full eval reports 23 verdicts: 22 holds, 0 drift, 1 needs-judgement (`rule-dispatcher-error-hint`). Every `code:` binding resolves.
- **doc_trace:** c3-1, c3-108
- **code_trace:** `cli/internal/eval/**`
- **via:** `c3x eval`
- **coverage:** clean

---

## Trust / conformance

### qa-14 — Is rule-output-via-helpers actually enforced?
- **Answer:** Yes. The eval gathers `fmt.Print*` occurrences in `cli/cmd/` excluding `output.go` and asserts `count == 0`; verdict holds (`count 0 == 0`).
- **doc_trace:** rule-output-via-helpers
- **code_trace:** `cli/cmd/output.go`
- **via:** `c3x eval rule-output-via-helpers`
- **coverage:** clean

### qa-15 — Does the platform-triple claim still hold in the wrapper?
- **Answer:** Yes. `eval c3-203` (contains_all) confirms `linux/amd64`, `linux/arm64`, `darwin/arm64` all present in `c3x.sh` — verdict holds.
- **doc_trace:** c3-203
- **code_trace:** `skills/c3/bin/c3x.sh`
- **via:** `c3x eval c3-203`
- **coverage:** clean

### qa-16 — Are all version surfaces in sync?
- **Answer:** Yes. `eval ref-fat-thin-distribution` and `eval c3-301` both confirm all version files read **11.2.0** (VERSION, plugin.json, marketplace.json, package.json, version.ts, lockfile) — verdicts hold ("all equal: 11.2.0").
- **doc_trace:** ref-fat-thin-distribution, c3-301
- **code_trace:** `skills/c3/bin/VERSION`, `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, `packages/cli/package.json`, `packages/cli/src/version.ts`
- **via:** `c3x eval ref-fat-thin-distribution`; `c3x eval c3-301`
- **coverage:** clean

### qa-17 — Can a passing eval verdict be silently wrong?
- **Answer:** Yes — a known latent trust bug. The `assert` op treats `Exists` on an empty frame as **drift** ("nothing to check exists", eval.go ~L236) but a `count == 0` rule on an empty frame returns **holds** (`countOK`, eval.go ~L272). So a renamed/missing gather path can read as holds. Not yet fixed.
- **doc_trace:** (none)
- **code_trace:** `cli/internal/eval/eval.go` (assert / countOK asymmetry)
- **via:** read `cli/internal/eval/eval.go`
- **coverage:** gap (KNOWN BUG)

---

## Model / why

### qa-18 — What is C3, in one paragraph?
- **Answer:** C3 holds a codebase's architecture as frozen, verifiable facts in the architecture's own vocabulary, mutated only through reviewed atomic change-units, and ships three ways: a Go CLI engine, a Claude skill, and an npm installer.
- **doc_trace:** c3-0, c3-201
- **code_trace:** `skills/c3/SKILL.md`
- **via:** `c3x read c3-0`; `c3x read c3-201`; read `skills/c3/SKILL.md`
- **coverage:** clean

### qa-19 — What is the change-unit / freeze / canvas three-act story?
- **Answer:** Act 1: shape the model and freeze the facts (shared truth, never hand-edited). Act 2: change-units drive progress — frozen facts mutate only through an atomic gated saga. Act 3: the canvas grows (raise a rung, additive) or evolves (morph a mis-modeled type, non-additive).
- **doc_trace:** c3-201
- **code_trace:** `skills/c3/SKILL.md`
- **via:** read `skills/c3/SKILL.md`; `c3x read c3-201`
- **coverage:** clean

### qa-20 — Why is eval a one-off check, not a gate?
- **Answer:** Eval checks a frozen fact's claim against an uncontrolled external (live code) the change-unit cannot govern; blocking on it would couple the frozen saga to drift it can't control, so eval reports on CI cadence and never gates apply. A verdict is determinism-stamped to the exact (claim, external-state) pair measured.
- **doc_trace:** ref-eval-determinism, c3-108
- **code_trace:** `cli/internal/eval/eval.go`
- **via:** `c3x read ref-eval-determinism`; `c3x eval --help`
- **coverage:** clean

### qa-21 — What's the difference between a ref and a rule?
- **Answer:** A ref captures rationale — a "why this over alternatives" reference document whose value is the why. A rule is an enforceable coding standard with a literal golden example.
- **doc_trace:** ref-fat-thin-distribution, rule-output-via-helpers
- **code_trace:** `.c3/canvases/ref.md`, `.c3/canvases/rule.md`
- **via:** `c3x canvas list`
- **coverage:** clean

### qa-22 — Why does apply enforce integrity on the tool instead of trusting the author?
- **Answer:** A silently-drifting shared contract is worse than none, so divergence is made impossible by construction — the atomic saga and derived membership (parent tables synthesized from children's `parent:`) keep the result integral, not author discipline.
- **doc_trace:** c3-0, c3-104
- **code_trace:** `cli/internal/changeset/**`
- **via:** `c3x read c3-104`; read `references/change.md`
- **coverage:** clean

### qa-23 — What's the relationship between the skill, the CLI, and the npm client?
- **Answer:** The Go CLI binary is the single source of behavior. The Claude skill (cli-wrapper c3-203) detects the platform and execs the version-pinned binary; the npm `@c3x/cli` client (binary-downloader c3-301) downloads and checksum-verifies the binary then forwards args. Both are thin wrappers over the same engine.
- **doc_trace:** c3-0, c3-2, c3-203, c3-301, ref-fat-thin-distribution
- **code_trace:** `skills/c3/bin/c3x.sh`, `packages/cli/src/version.ts`
- **via:** `c3x read c3-203`; `c3x read c3-301`; `c3x read c3-2`
- **coverage:** clean
