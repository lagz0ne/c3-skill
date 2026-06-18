# CLI surface audit — concept solidity + reduction

Four parallel read-only audits (authoring / change-lifecycle / read-query / definitions-peripheral)
against the clarified model: canvas-defined types, frozen facts edited only via change-units, a
column marked `edge: <rel>` IS the citation, `graph --unit` overlay, agent-mode TOON. Every claim
below is cited to source by the auditors; this is the synthesis + disposition.

## Verdict per command

| Command | Verdict | Note |
|---|---|---|
| `add` | **keep** (solid) | only create path; already syncs edge columns |
| `write` | **keep** (solid for change-docs/canvas) | still carries legacy frontmatter `syncRelationships` (deferred #7) |
| `set` | **keep** (solid) | `codemap`/`status` carve-outs are load-bearing |
| `wire` | **REMOVE** (legacy/dead) | refused on every frozen fact; one reachable path (`wire <adr> <ref>`) writes an **unmanaged orphan `uses` edge** into a no-`edge:` ADR table no flow uses; superseded by the edge column. `wire_test.go` is guard-dead (calls `RunWire` directly, bypassing the CLI freeze). |
| `delete` | **keep** (solid) | minor: row-cleanup still hardcodes legacy section names |
| `change new` | keep | 8-line mkdir + next-step |
| `change view` | **keep as the one surface** | |
| `change status` | **merge → `view`** | emits byte-identical struct in agent/TOON mode; differs only in human prose |
| `change rebase` | **merge → `view`** | thin read-only drift diagnostic overlapping `view`; **has no test**; weakest subcommand |
| `change accept` | keep | only writer of `accepted` |
| `change apply` | keep (the heart) | overlay preflight thinner than apply preflight (deferred Codex #1) |
| `change scaffold` | keep (solid) | rung-climb gate is real + enforced |
| `supersede` | keep (solid) | only path to `superseded`; routes through privileged status writer; no freeze bypass |
| `migrate` | keep-then-retire (legacy) | still has live work (42 legacy-status ADRs on disk); retire after repo swept |
| `list` | keep (solid) | only global inventory/coverage |
| `read` | keep (solid) | truncation applies only in JSON/agent branch (minor inconsistency) |
| `search` | keep; **trim flags** | dead `SearchOptions.Type` field; `--hybrid`/`--semantic` are documented no-ops |
| `lookup` | keep (solid) | the file-context gate |
| `graph` | keep (flagship) | only multi-hop view + the only overlay surface |
| `codemap` | already hidden; **retire** | scaffolds blank code-map rows; signal duplicated by `check`/`list` |
| `index` | **HIDE** from public surface | skill says "never a correctness step"; `search` self-heals the index on demand, yet `index` sits in public help inviting a forbidden "re-index first" detour |
| `schema` | keep | annotated authoring view; thinnest-justified dual with `canvas read` |
| `canvas` (list/read/add/write) | keep (core) | the only way to evolve a type's shape |
| `template` (cmd) + `--template` flag | **REMOVE** (100% dead) | pure retirement tombstone; no skill calls it; no CI coupling |
| `check` / `--fix` | keep (central) | |
| `repair` | keep | NOT redundant with `--fix`: force-rebuild + tree-wide reseal vs the gated `accepted→done` latch only |
| `git install` | keep (on-concept) | hook + gitattributes so canonical markdown is the committed artifact |
| `marketplace` (add/list/show/update/remove) | **carve-out candidate** (off-concept) | ~700 LOC + own `internal/` package; rule-pack package-manager grafted on; zero registered sources in practice |

## Fixed this pass (correctness, no decision needed)
- **`search` agent-mode TOON `%v` corruption** (`039b1ae`): the earlier nested-slice fix missed a
  plain nested *struct* field, so `search` dumped its `context` as `{{c3-210 …} { } { } }` on the
  hottest agent path. `marshalStructWithIndent` now recurses into struct fields (Stringer types stay scalar).

## The one concept gap that matters: `uses` is overloaded
`uses` carries two different relationships derived by two different mechanisms:
1. **Governance citation** — component → ref/rule, via the **edge column** (Governance/Reference,
   `edge: uses targets:[ref,rule]`, filtered by the cited entity's *actual* type). Clean, single source of truth.
2. **Structural dependency** — component → component, via **legacy frontmatter** `uses:` (`write`'s
   `syncRelationships`). e.g. c3-110 `uses: [c3-103]`.

The ≤1-writer rule means these can't both be `uses` in two columns. This is why **c3-design can't naively
adopt** the edge column: marking Governance/Reference `edge: uses targets:[ref,rule]` suppresses frontmatter
`uses:` and filters out the component→component deps — silently dropping c3-110→c3-103. Resolving this is
*the* "is the concept solid" decision (see the live decision thread / questions).

## Reduction tiers (by confidence)
- **Tier 1 — dead/safe:** remove `template`+`--template`; hide `index`; fix `change --help` (omits `scaffold`); trim `search` dead `Type` field + no-op flags.
- **Tier 2 — legacy, strong remove:** remove `wire`/`RunWire`/`RunUnwire`; retire hidden `codemap`.
- **Tier 3 — merge:** `change view`/`status`/`rebase` → one surface (7→5 subcommands).
- **Tier 4 — product decision:** carve out `marketplace`.

## Minor follow-ups (non-blocking)
- `change --help` omits `scaffold` (help.go:419-436) — doc drift.
- `delete` row-cleanup hardcodes legacy section names rather than resolving the canvas edge column.
- `read` truncation only in the JSON/agent branch.
- overlay is graph-only; `read`/`lookup`/`list`/`search` not yet `--unit` lens-aware (tracked Remaining).
