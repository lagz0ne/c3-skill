# Design — the eval mechanism: a pipeline of `gather · filter · eval · loop`

> Supersedes both the op-centric cut and the fixed extract+transform+compare pipeline. The eval
> is a tiny **composable algebra** — four primitives that nest. Most conformance checks are a
> composition of them; "ops" (code/docs/artifact) and "compare modes" were both premature — they
> are just different compositions over different sources.

## The four primitives
Each is a stage in a pipeline; they nest (a `loop` body is itself a pipeline).

| Op | Signature | Does |
|---|---|---|
| **gather** | `source → []value` | Acquire data — **raw** (read content) or **mechanical** (run a command). The source + how to read it. |
| **filter** | `[]value → []value` | Keep items matching a predicate. Narrow the gathered set (drops items). |
| **transform** | `[]value → []value` | Reshape each value — extract a field, trim, normalize, project to the comparable form. Distinct from filter (filter drops; transform maps). |
| **eval** | `[]value → verdict\|value` | Assert / compute: equality, presence, comparison, or a **judgement** call. Emits `holds`/`drift`/`needs-judgement`, or a derived value to feed onward. |
| **loop** | `([]value, pipeline) → []verdict` | Fan a sub-pipeline over each item; collect per-item verdicts. The "for each component / file / fact". |

`loop` is the genuinely new one — iteration the flat pipeline couldn't express. **Practical note:**
in the *mechanical* gather, `gather` and `transform` fuse (`jq -r .version` reads *and* reshapes);
`transform` earns its own stage when you `gather: raw` and reshape separately.

## Why this is the whole mechanism
Four ops compose to a **conformance algebra**. "Does the external still hold the claim?" is, for
nearly everything: *gather the relevant material, filter to what the claim governs, eval the
assertion — looping when the claim spans a collection.* The cases that need a human are just an
`eval` whose mode is `judgement` instead of `mechanical`. Nothing else is needed — no per-kind
ops, no fixed compare shape.

```
# c3-design code-conformance sweep — all four
loop over each component fact (c3-1xx):
  gather  its codemap globs            → files          # mechanical
  filter  to files the claim names                       # predicate
  eval    every file exists / claim holds → verdict       # mechanical assert, per component

# c3-107 changelog drift — the degenerate inner case
gather  store files matching "changelog"  → []           # mechanical: grep
filter  (none)
eval    count == 0  vs claim "provides changelog storage"  → drift
```

## The eval-spec is a pipeline, not a schema
```yaml
# .c3/eval/<fact-id>.yaml  — a composition, read top-to-bottom
- loop: { over: { gather: { facts: "c3-1*" } } }    # optional outer iteration
  do:
    - gather: { source: "{codemap}", as: mechanical } # raw | mechanical(+command)
    - filter: { where: "names a claimed symbol" }
    - eval:   { assert: exists, mode: mechanical }     # mechanical | judgement
```
A bare existence check (the codemap) is just `gather(glob) → eval(exists)`. A judgement check is
the same shape with `eval.mode: judgement`. Generality is composition, not configuration.

## What survives unchanged (from [draft-eval-spec.md](draft-eval-spec.md))
- **Verdict + one-off stamp** `{fact, fact_root_merkle, eval_spec_hash, external_state, verdict,
  evidence}`; `external_state` = hash of what `gather` read. Still true only for that pair.
- **Solidity = frozenness of both sides**; living-code gather → snapshot re-run on CI.
- **Coverage**: HOLE = source material that no fact's `gather` reached.
- **c3x is a helper, never a gate**: it runs the pipeline, surfaces `needs-judgement`; no apply
  gate, no `--strict`.

## What to build — a tiny interpreter
The engine is a **pipeline interpreter** over four ops, not an op registry. ~one evaluator:
`gather` (read file / exec command), `filter` (predicate), `eval` (assert → verdict), `loop`
(map a sub-pipeline). The MVP runs the code-conformance sweep above over c3-design's own 15
fresh codemaps; `c3 eval` emits the verdict array; the up-V + `--strict-codemap` come off apply.

## Open
- **Predicate / assert language** — how `filter.where` and `eval.assert` are expressed. A tiny
  built-in vocabulary (exists, count, equals, contains)? An expression DSL? Or do `gather`'s
  `command` + a shell predicate cover it, keeping the spec declarative-thin?
- **`command` execution** — sandbox, working dir, `{source}`/`{codemap}` substitution, timeout.
- **How a pipeline references the claim** — `eval.assert` mentions "the claim"; is that the
  frozen body, a named projection of it, or a value the fact froze for this purpose?
- **`loop` scope** — only over gathered facts/files, or arbitrary collections?
- Naming: the four ops; `.c3/eval/`; `c3 eval`.
