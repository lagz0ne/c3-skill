# Query — reading the frozen facts

The **read** beat of Act 1. The topology you read here is **frozen shared truth** — onboarded once, changed only through change-units (`change.md`). So a `read` / `lookup` / `search` / `graph` result is **canonical**: answer from it with confidence, no hedging about whether the doc is current. Full context = these facts + the code they bind.

Reads are free and side-effect-free — favor them. Reach for `c3 list` / `c3 check` only on a search miss, suspected drift, or a topology-wide inventory (the Precondition in `SKILL.md`). **Never Read/Glob/Edit `.c3/` instance files** — they are CLI-only and raw access goes stale.

## Pick the discovery tool by what you were handed

| You have | Run | You get |
|----------|-----|---------|
| A concept, capability, paraphrase ("where is X", "what handles Y") | `c3 search "<question>"` | Ranked entities by semantic + keyword + graph signal, plus route clues when available |
| A known file or glob | `c3 lookup <file-or-glob>` | Owning component(s) + governing refs/rules |
| A known id (± section) | `c3 read <id> --section <name>` | Entity body |

For natural-language questions reach for `search` **before** `list`-and-title-matching — search ranks by meaning. `match_sources` tells you *why* a hit ranked: `semantic` = meaning matched despite different wording; `content_fts` / `entity_fts` = keyword; `graph:*` = relationship context. Use `c3 list` only for inventory/coverage **after** candidates are known.

## Navigate the layers

Start from the best candidate, then move through Context → Container → Component as the answer demands ownership, boundaries, or implementation:

1. `c3 read <id>` (`--full` for the whole body) when the snippet isn't enough.
2. `c3 lookup <file>` on **every** file path before you open source — it returns the owning component plus the refs/rules governing that file (those are its constraints). Directory-level: `c3 lookup 'src/auth/**'`.
3. `c3 graph <id> --depth 1` for relationships and route enrichment. Read `route.facts`, `route.graph`, `route.anchors`, `route.lanes`, and `route.hash` before opening broad source search.
4. `c3 graph <id> --format mermaid` when a relationship diagram helps — include the mermaid as a code block. Root it on the matched container/component, never `c3-0`.
5. Then explore code: start with the paths/symbols surfaced by `lookup` and graph `route.anchors`.

## Use route-enriched graph output

`c3 search` and `c3 graph` can include a `route:` block. Treat it as a context pack:

| Field | Use |
|-------|-----|
| `facts` | The fact/ref/rule ids that carry authority for the route. Read these before making claims. |
| `graph` | Neighbor ids that explain the path or blast radius. Read a neighbor before calling it affected. |
| `anchors` | First files, globs, docs, or tests to inspect. These are clues, not proof. |
| `lanes` | Lifecycle/ownership lanes such as auth, invoice, realtime-cycle, theming, or time. Use them to avoid a one-file answer to a cross-cutting question. |
| `hash` / `hash_basis` | Change signal for the route shape. It is not a correctness verdict. |

Route enrichment is not a new command, not a code-correction layer, not an eval replacement, and not an apply gate. The answer still cites the C3 facts it read and the code/tests/eval/runtime evidence it used.

## Two rules every answer obeys

**Evidence.** Every material claim is bound to a read you actually ran, and the answer names it (entity + section) by its exact c3 id — never a truncated id or a path-derived name. No read behind a claim → run the read or drop the claim. Listing commands up front grounds nothing.

**Causal chain.** When a question crosses mechanisms ("trace end-to-end", "what informs users", "what breaks if X changes"), deliver a **chain**, not a flat entity list:

```
owner of the action → owner of the state mutation → the mechanism that propagates it
  → the dependent/observer → the emergent property → the failure boundary
```

Each arrow states *which contract carries the hop* — the ref, subject, permission, or edge that makes the next entity follow. Use graph `route:` fields to choose candidate hops and first code anchors, then confirm each hop with `c3 read`, `lookup`, code, tests, or eval. A reverse-graph neighbor is a candidate, not a conclusion: read it before assigning behavior, and label it **direct** (cites/consumes the thing) or **transitive** (reached through another). Copy concrete names (queues, subjects, channels) verbatim from the docs — don't flatten them to "the notification system". If the docs don't say how the path degrades, report that gap explicitly; never guess. State negatives from evidence too: "no rules apply" means "no `rule-*` found in the output", not an invented `rule-auth`.

## Boundaries

- **ADRs** — status meaning, the `[open, accepted, done, superseded]` set, and default exclusion (`--include-adr` to surface) are the shared contract in `SKILL.md`; the saga that moves them lives in `change.md`. Cite an ADR only with its current/superseded/historical label, and verify against the live entity doc before acting on it.
- **Impact / "what breaks" / verification checks** — that's the pre-flight blast radius and destruction-gate prediction in `sweep.md`.
- Topic not in C3 → search code directly, suggest documenting it. Docs read stale → that's a seal/consistency question for `audit.md`.
