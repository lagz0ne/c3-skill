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
| **gather** | acquire data — `raw` (read a file) or `mechanical` (run a `command`); also `files` (glob), `facts` (id-glob), `code` (a fact's declared globs), `each` (several), `literal` |
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
