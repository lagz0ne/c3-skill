# Eval Reference

Eval answers one question, per fact, on demand: **does this frozen claim still hold against the
uncontrolled external it governs?** A fact is frozen; the code/doc/artifact it describes is not.
`c3 eval` checks the two against each other — a **one-off** verdict at the instant it runs, never
a gate. You run it when you want proof (CI cadence), the way you'd run a test.

It is **not** a standing guarantee. A verdict is solid only for the exact `(claim, external-state)`
pair it stamped; living code drifts the next commit, so you re-run. Freeze both sides (a pinned
SHA, a released artifact) and the verdict becomes durable.

## The mechanism — a pipeline of five ops
An eval-spec is a small composition; most checks need two or three steps:

| Op | Does |
|----|------|
| **gather** | acquire data — `raw` (read a file), `mechanical` (run a `command`), or structural (`outline` via ast-grep); also `files` (glob), `facts` (id-glob), `code` (a fact's declared globs), `each` (several), `literal` |
| **filter** | keep values matching a predicate (`contains`, `matches`) |
| **transform** | reshape each value (`trim`, `first`, `lines`) |
| **eval** | assert → verdict: `exists`, `equals`, `all_equal`, `contains_all`, `contains`, `count`, or `judgement` (surfaces, never scores) |
| **loop** | fan a sub-pipeline over a collection, binding `$item` |

`gather` + `transform` fuse in the mechanical case (`jq -r .version` reads *and* reshapes).
`eval: judgement` is the escape hatch — when equality can't decide, it emits `needs-judgement`
with the gathered evidence for a human/agent to rule on.

## The spec — `.c3/eval/<fact-id>.yaml`
```yaml
fact: c3-102
claim: "store-lib is implemented at internal/store"
code:                              # the fact→code binding (also what `lookup` reads)
  - cli/internal/store/**
# no pipeline ⇒ default: every declared `code:` glob must resolve to ≥1 file
```
A richer, behavioural check writes an explicit pipeline:
```yaml
fact: c3-203
claim: "cli-wrapper gates linux/amd64, linux/arm64, darwin/arm64"
code: [ skills/c3/bin/c3x.sh, skills/c3/bin/VERSION ]
pipeline:
  - gather: { file: skills/c3/bin/c3x.sh }
  - eval:   { contains_all: ["linux/amd64", "linux/arm64", "darwin/arm64"] }
```
The `code:` field is the binding: it is what `c3 lookup <file>` resolves against, and (with no
pipeline) the default per-glob resolve check. The frozen fact body is the claim; the spec is the
**mutable lens** — re-aim it freely (code moved dirs), it is never frozen.

For broad code-shape checks, use ast-grep outline instead of shelling out to grep/awk. Outline is a
syntax-aware gather that emits compact structural units: top-level items and direct members, with
names, signatures, symbol types, import/export/public flags, and AST kinds. It intentionally does
not stamp whole function bodies or source ranges, so body-only edits do not churn the matched state.
```yaml
fact: c3-108
claim: "eval exposes deterministic gather types and direct eval helpers"
pipeline:
  - gather:
      outline:
        paths: [cli/internal/eval]
        lang: go
        view: digest          # default; item signatures + compact member names
        type: struct,function,method
  - eval:
      contains_all: ["type Gather struct", "func (e *Engine) gather", "func (e *Engine) assert"]
```
`outline` calls `ast-grep outline --json=stream` through C3's resolved ast-grep executable. Fat and
npm-managed C3 installs use the release-pinned bundled binary; source/dev `eval` runs try to fetch
that pinned binary when npm is available, and may set `C3_AST_GREP` explicitly otherwise. On Linux,
C3 deliberately does not fall back to `sg`, because that name may be util-linux `sg`, not ast-grep.

Supported languages follow ast-grep's built-in language list. C3's default outline extractor is most
useful today for claim-bearing code units in Go, Java, JavaScript/TypeScript/TSX, Kotlin, Python,
Rust, and Swift. Other ast-grep-supported languages can still be parsed or searched with custom
outline rules: Bash, C, C++, C#, CSS, Elixir, Haskell, HCL, HTML, JSON, Lua, Nix, PHP, Ruby, Scala,
Solidity, and YAML. Anything outside ast-grep's built-in list is unsupported by `gather.outline`
unless the project supplies compatible ast-grep custom language/rule support.

## Loop — one verdict over many facts
When a claim spans a group, `loop` fans a sub-pipeline over each item (binding `$item`) and rolls
the per-item verdicts into one — **holds iff all hold**, and the evidence names each item's
verdict, so a single drift says *which* member fell. A container asserting its components all
resolve:
```yaml
fact: c3-1
claim: "the Go CLI container's components each resolve to real code (roll-up)"
pipeline:
  - loop:
      over: { facts: "c3-1[0-9][0-9]" }   # each component id (over accepts any gather:
        # facts id-glob, files glob, literal list)
      do:
        - gather: { code: "$item" }        # the item fact's declared globs → files
        - eval:   { exists: true }
```
Use a per-fact spec for the detail (and the lookup binding); a `loop` spec on the parent for a
single roll-up line.

## Run it
```bash
c3 eval                 # every spec → verdict array (holds / drift / needs-judgement, stamped)
c3 eval c3-203          # one fact's spec
c3 eval --json          # machine output for CI
```
A `drift` row names what moved; a `needs-judgement` row carries the evidence to judge. `c3 eval`
**exits success regardless** — the verdict is the signal; CI decides what to do with it.

## When to use
- **Code conformance** — a component's claim vs its `code:` (replaces the old code-map check).
- **Cross-surface invariants** — versions in sync, a generated artifact matches its spec.
- **Cross-lane agreement** — this `.c3/` fact vs a parallel doc set.

Author a spec when a fact makes a claim worth re-checking against something it doesn't control.
`c3 lookup <file>` reads the same `code:` bindings to map a file back to its owning fact.
