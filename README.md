# C3: Architecture That Agents Can Read — and Change Safely

C3 gives your codebase an architecture model an LLM can navigate, query, evaluate, and change **without breaking it**. The model is a sealed `.c3/` tree of markdown — reviewable in Git, validated by machine. (`c3.db` is just a local cache the CLI rebuilds at any time; never merge it.)

One model, three acts:

1. **Shape the model, freeze the facts.** You build the **canvas** — your architecture's own vocabulary, the sections and columns each entity carries *here* — then onboard the facts the work needs and flip the gate. From that moment the facts are **frozen**: shared truth, never hand-edited again.
2. **Change-units drive the work.** A frozen fact changes through exactly one path: a **change-unit** — a decision record plus a folder of patches, applied as one atomic, all-or-nothing saga. You declare intent; the tool keeps the result whole.
3. **The canvas grows with the need.** When the work outgrows the model, you *raise the canvas* and migrate every fact up — completeness is never relaxed.

The point of the freeze is Act 2: because facts only move through change-units, the tool can **guarantee integrity by construction** rather than hope docs stay current.

- **Membership writes itself.** Set a child's `parent:` and the parent's membership table grows the row — synthesized, never hand-authored, always in sync.
- **The destruction gate refuses strays.** A retire that would orphan a live child or dangle a live citation is **refused** — unless the same change-unit also reparents or retires it. The graph never strands.
- **Eval checks claims against reality.** `.c3/eval/*.yaml` binds facts to the code, files, commands, or ast-grep structural outlines they govern. `c3x eval` produces one-off `holds` / `drift` / `needs_judgement` verdicts without turning a single LLM answer into truth.
- **Search and lookup stay small.** `c3x search` finds concepts by semantic, keyword, and graph signal; `c3x lookup` maps files through eval bindings to owners, refs, and rules. Agent-mode output is TOON and tuned to keep the useful proof while dropping noise.
- **Rebase resolves conflict.** When a staged patch's cited block moves, `c3x change rebase` emits a drift bundle to re-author against the fresh anchor.
- **One edge has one source.** When a canvas derives `uses` from a body table, change apply rejects a competing frontmatter re-edge before any write and points to the exact body section to patch.

## Install / Run

**Claude plugin (no binary, installer-friendly):**

```bash
claude plugin install lagz0ne/c3-skill
```

Then: `/c3 onboard this project`

The source repository, `main` branch, and platform-neutral skill ZIP carry the skill, Claude plugin metadata, and wrapper only — no committed `c3x-*` binaries. On first real C3 command the wrapper delegates to the pinned `@c3x/cli` runtime manager, which downloads verified release assets into a versioned local cache.

**Fat skill ZIPs (self-contained):**

Use a per-platform release asset when the skill must run in a sandboxed or offline environment:

- `c3-skill-<os>-<arch>-v<version>.zip` is the full fat build. It carries the `c3x` binary, the release-pinned ast-grep binary for structural eval gathers, and embedded semantic model/native ONNX runtime. It includes `.gitattributes` so Git preserves bundled binaries as binary content.
- `c3-skill-linux-<arch>-portable-v<version>.zip` is the portable Linux fat build. It carries a bundled pure-Go `c3x` binary plus the release-pinned ast-grep binary for that Linux target; semantic ONNX search is unavailable in that build, so search falls back to keyword/graph behavior.

Fat ZIPs are GitHub Release artifacts, not files committed back to `main`.

**`npx` CLI (thin, fetched on demand):**

```bash
npx @c3x/cli check
npx @c3x/cli search "how do users sign in and get permissions"
npx @c3x/cli runtime versions
npx @c3x/cli runtime use 11.5.0
```

The npm package downloads the matching `c3x` binary, semantic model, and, for outline-capable runtimes, the pinned ast-grep binary from the GitHub Release into a versioned local cache on first use. `npx @c3x/cli runtime use <version>` writes `.c3/runtime.json` with only the selected runtime version; it never stores a binary path or URL.

## What You Get

| Say this | C3 does this |
|----------|-------------|
| `/c3` adopt this project | **onboard** — shape the canvas, discover the topology, author it all into the genesis change-unit, flip once to freeze |
| `/c3` where is auth? | **query** — read the frozen truth: `search` by meaning, then `lookup`, `read`, `graph` |
| `/c3` add rate limiting | **change** — open a change-unit, author patches, `apply` lands them all-or-nothing past the gate stack |
| `/c3` create a ref for error handling | **ref** — a rationale-bearing fact other facts cite |
| `/c3` add a rule for structured logging | **rule** — an enforceable standard with a golden example |
| `/c3` edit the canvas for ADRs | **canvas** — reshape the definitions that govern your docs, or raise a rung |
| `/c3` does the auth doc still match code? | **eval** — run deterministic gather/transform/verdict specs over the external state a fact governs |
| `/c3` audit the docs | **audit** — is the sealed truth intact and consistent? ends in PASS / WARN / FAIL |
| `/c3` what breaks if I change payments? | **sweep** — reverse-graph blast radius + whether the destruction gate will even let the change land |

## Current CLI Shape

The common read path is intentionally narrow:

```bash
c3x search "billing retries"
c3x lookup 'cli/cmd/*.go'
c3x read c3-108 --section Goal
c3x graph c3-108 --format mermaid
```

Use `c3x check` for sealed-doc integrity and `c3x eval` for fact-vs-external conformance. They are separate on purpose: a frozen fact can be structurally valid while its governed code has drifted.

The npm entrypoint adds a namespaced runtime manager so cache operations do not collide with project commands:

```bash
npx @c3x/cli runtime versions
npx @c3x/cli runtime installed
npx @c3x/cli runtime install latest
npx @c3x/cli runtime use 11.5.0
npx @c3x/cli runtime prune
```

The full command catalog, flags, and gate details live in the skill: read `skills/c3/SKILL.md`, or run `c3x --help` (the packaged CLI is authoritative).

> **For agents:** the `/c3` skill invokes the CLI for you via `bash <skill-dir>/bin/c3x.sh`. Never run bare `c3x` — go through `/c3`.

## Evaluation status

The repository includes a generic, isolated retrieval evaluation for the
structural-owner use case: before changing a record, project a direct hit to
its immediate owner while preserving legitimate peer context.

The candidate path is deliberately opt-in. `RunSearch` is byte-compatible by
default; `StructuralProjection` and `CaptureProvenance` are internal evaluator
options, not public CLI flags. Missing provenance fails closed instead of
guessing an owner or route.

The accepted preliminary v4 containment result is:

| Metric | Unchanged C3 | Explicit candidate |
|---|---:|---:|
| Owner recall @5 | 0.667 | 1.000 |
| Owner MRR | 0.278 | 1.000 |
| Structural-owner precision | 0.667 | 1.000 |
| Forbidden rows in no-target case | 1 | 0 |

The owner-recall delta is **+0.333** across five repeatable replays. This is
controller-level benchmark evidence only. Agent turns, token spend, money,
and product impact are not measured by this microbenchmark. Route cases remain
held out because their direct-FTS miss witness is not reproducible from the
current generic loader.

Generic retained artifacts:

- [v4 paired microburst](research/eval/structural-retrieval-v4/paired-microburst.v4.json)
- [v4 repeatability](research/eval/structural-retrieval-v4/repeatability.v4.json)
- [v4 fixtures](research/eval/structural-retrieval-v4/fixtures.v4.json)
- [v4 benchmark](research/eval/structural-retrieval-v4/benchmark.v4.json)

## Release verification

Run the normal suite and the explicit evaluator checks before releasing:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh check
cd cli && go test ./...
cd cli && go test ./tools/structural-search-eval-v3 -count=1
cd cli && go vet ./...
cd cli && RUN_V4_MICROBURST=1 go test ./tools/structural-search-eval-v3 \
  -run TestV4PairedMicroburstArtifact -count=1 -v
```

The v3 baseline and benchmark are frozen inputs. If you replay a capture, use
the canonical output basename (`B-v3-baseline.json` or `B-v4-baseline.json`)
so the artifact self-reference remains correct. Do not treat the v4
microbenchmark as proof of product effectiveness.

## Contributing

Building, testing, and releasing C3 itself is covered in [CLAUDE.md](CLAUDE.md). In short: `cd cli && go test ./...` runs the suite; CI owns the cross-compile and release.

## License

MIT
